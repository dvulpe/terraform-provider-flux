package flux

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"time"
)

func datasourceFluxIdentityPublicKey() *schema.Resource {
	return &schema.Resource{
		Read: datasourceFluxIdentityPublicKeyRead,
		Schema: map[string]*schema.Schema{
			"public_key": {
				Type:        schema.TypeString,
				Description: "Flux generated SSH public key.",
				Computed:    true,
			},
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The namespace where flux is deployed",
			},
			"pod_labels": {
				Type:        schema.TypeMap,
				Description: "Flux pod labels",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Flux listening port",
				Default:     "3030",
			},
		},
	}
}

type FluxIdentity struct {
	Key string `json:"key"`
}

func datasourceFluxIdentityPublicKeyRead(d *schema.ResourceData, meta interface{}) error {
	stopCh := make(chan struct{})
	localPort, err := meta.(KubeClient).PortForward(PortForwardRequest{
		RemotePort:  d.Get("port").(int),
		Namespace:   d.Get("namespace").(string),
		PodLabels:   asMap(d, "pod_labels"),
		StopCh:      stopCh,
	})
	defer close(stopCh)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d/api/flux/v6/identity.pub", localPort), nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var fluxIdentity FluxIdentity

	if err := json.NewDecoder(res.Body).Decode(&fluxIdentity); err != nil {
		return err
	}

	hash := sha256.New()
	hash.Write([]byte(fluxIdentity.Key))
	d.SetId(base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	return d.Set("public_key", fluxIdentity.Key)
}

func asMap(d *schema.ResourceData, key string) map[string]string {
	src := d.Get(key).(map[string]interface{})
	res := make(map[string]string)
	for k, v := range src {
		res[k] = v.(string)
	}
	return res
}
