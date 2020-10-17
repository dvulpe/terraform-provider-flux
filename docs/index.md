---
page_title: "Provider: Flux"
subcategory: ""
description: |-
  Terraform provider for interacting with FluxCD API.
---

# Flux Provider

The Flux provider is used to interact with FluxCD api.

## Example Usage

Do not keep your authentication password in HCL for production environments, use Terraform environment variables.

```terraform

provider "flux" {
  version                = "0.0.1"
  host                   = CLUSTER_HOST
  cluster_ca_certificate = CA_CERTIFICATE
  token                  = TOKEN
}

terraform {
  required_providers {
    flux = {
      source  = "dvulpe/flux"
      version = "x.x.x"
    }
  }
}

```

## Schema

### Optional
- **host** (String, Optional) Kubernetes cluster endpoint
- **cluster_ca_certificate** (String, Optional) Kubernetes cluster CA certificate
- **token** (String, Optional) Kubernetes Authentication Token - useful for EKS
- **load_config_file** (Bool, Optional) whether to load a kube config file instead
- **config_path** (String, Optional) Path to a kube config file
