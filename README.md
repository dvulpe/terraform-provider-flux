# Terraform Provider Flux

This repository contains a terraform provider for [Flux](https://github.com/fluxcd/flux).

It can be used to read Flux's generated public SSH key from a Kubernetes cluster as an 
approach to avoid storing secrets in the Terraform state.

An example of using it in action can be found in [dvulpe/terraform-eks](https://github.com/dvulpe/terraform-eks)

## Build

Run the following to build the provider:
```
make build
```
