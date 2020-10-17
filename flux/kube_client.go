package flux

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"log"
	"net/http"
)

type KubeClient interface {
	PortForward(req PortForwardRequest) (localPort int, err error)
}

type PortForwardRequest struct {
	RemotePort int
	Namespace  string
	PodLabels  map[string]string
	StopCh     <-chan struct{}
}

type kubeClient struct {
	config *restclient.Config
}

func (k *kubeClient) PortForward(req PortForwardRequest) (localPort int, err error) {
	kc, err := kubernetes.NewForConfig(k.config)
	if err != nil {
		return -1, err
	}
	if err := k.awaitPodAvailable(kc, req.Namespace, req.PodLabels); err != nil {
		return -1, fmt.Errorf("could not find pod in the cluster: %w", err)
	}
	podList, err := kc.CoreV1().Pods(req.Namespace).List(v1.ListOptions{
		LabelSelector: labels.Set(req.PodLabels).String(),
	})
	if err != nil {
		return -1, err
	}
	if len(podList.Items) == 0 {
		return -1, fmt.Errorf("did not find a pod in namespace: %s matching labels: %v", req.Namespace, req.PodLabels)
	}

	transport, upgrader, err := spdy.RoundTripperFor(k.config)
	if err != nil {
		return 0, err
	}
	fwdURL := kc.CoreV1().RESTClient().
		Post().Resource("pods").
		Namespace(req.Namespace).Name(podList.Items[0].Name).
		SubResource("portforward").URL()
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, fwdURL)
	readyCh := make(chan struct{})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", req.RemotePort)}, req.StopCh, readyCh, ioutil.Discard, ioutil.Discard)
	if err != nil {
		return 0, err
	}
	go fw.ForwardPorts()
	select {
	case <-readyCh:
		break
	}
	ports, err := fw.GetPorts()
	if err != nil {
		return 0, err
	}
	return int(ports[0].Local), nil
}

func (k *kubeClient) awaitPodAvailable(kc *kubernetes.Clientset, namespace string, podLabels map[string]string) error {
	pods, err := kc.CoreV1().Pods(namespace).List(v1.ListOptions{LabelSelector: labels.Set(podLabels).String()})
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("found 0 pods matching selector")
	}
	for _, c := range pods.Items[0].Status.ContainerStatuses {
		if c.Ready == false {
			return fmt.Errorf("container %s not ready", c.Name)
		}
	}
	return nil
}

func initializeConfiguration(d *schema.ResourceData) (*restclient.Config, error) {
	overrides := &clientcmd.ConfigOverrides{}
	loader := &clientcmd.ClientConfigLoadingRules{}

	if d.Get("load_config_file").(bool) {
		path, err := homedir.Expand(d.Get("config_path").(string))
		if err != nil {
			return nil, err
		}
		loader.ExplicitPath = path
	}

	if v, ok := d.GetOk("cluster_ca_certificate"); ok {
		overrides.ClusterInfo.CertificateAuthorityData = bytes.NewBufferString(v.(string)).Bytes()
	}
	if v, ok := d.GetOk("host"); ok {
		host, _, err := restclient.DefaultServerURL(v.(string), "", apimachineryschema.GroupVersion{}, true)
		if err != nil {
			return nil, fmt.Errorf("failed to parse host: %s", err)
		}
		overrides.ClusterInfo.Server = host.String()
	}
	if v, ok := d.GetOk("token"); ok {
		overrides.AuthInfo.Token = v.(string)
	}

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	cfg, err := cc.ClientConfig()
	if err != nil {
		// return default config here - the provider will be initialised once the k8s cluster is up and config values can be resolved
		// side-effect: when provider configuration is invalid it might fail with "could not connect to localhost"
		return &restclient.Config{}, nil
	}

	log.Printf("[INFO] Successfully initialized config")
	return cfg, nil
}
