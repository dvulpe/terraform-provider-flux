---
page_title: "Flux Identity Public Key - terraform-provider-flux"
subcategory: ""
description: |-
  The flux identity public key resource retrieves Flux's public SSH key.
---

# Data Source `flux_identity_public_key`

The flux identity public key resource retrieves Flux's public SSH key.

## Example Usage

```terraform
data "flux_identity_public_key" "flux_key" {
  namespace = "flux_namespace"
  pod_labels = {
    app = "flux"
  }
  port = 3030
  depends_on = [
    helm_release.flux,
  ]
}
```

## Attributes Reference

- `namespace` -  The kubernetes namespace where flux is deployed.
- `pod_labels` - Pod labels that can be used to identity the flux pod.
- `port` - The port the Flux API will listen to (defaults to 3030)

The following attributes are exported.

- `public_key` - Flux generated public key.
