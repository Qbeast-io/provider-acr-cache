
# Provider Azure ACR Cache

<div style="text-align: center;">

![CI](https://github.com/Qbeast-io/provider-acr-cache/workflows/CI/badge.svg)
[![GitHub release](https://img.shields.io/github/release/Qbeast-io/provider-acr-cache/all.svg)](https://github.com/Qbeast-io/provider-acr-cache/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/Qbeast-io/provider-acr-cache)](https://goreportcard.com/report/github.com/Qbeast-io/provider-acr-cache)
[![Contributors](https://img.shields.io/github/contributors/Qbeast-io/provider-acr-cache)](https://github.com/Qbeast-io/provider-acr-cache/graphs/contributors)


</div>

`provider-acr-cache` is a [Crossplane](https://crossplane.io/) provider for managing ACR cache built using the Go template and related Pulumi providers.

Warning:
`This provider is not ready to be used for production and might be missing resources`

Most of the testing have been done on [Azure Databricks](https://azure.microsoft.com/en-ca/products/acr-cache)

## Getting Started

Install the provider by using the following command after changing the image tag
to the [latest release](https://marketplace.upbound.io/providers/spiegela/provider-acr-cache):
```
up ctp provider install spiegela/provider-acr-cache:v0.1.0
```

Alternatively, you can use declarative installation:
```
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-acr-cache
spec:
  package: Qbeast-io/provider-acr-cache:v0.1.0
EOF
```

Notice that in this example Provider resource is referencing ControllerConfig with debug enabled.

You can see the API reference [here](https://doc.crds.dev/github.com/Qbeast-io/provider-acr-cache).

## Exemples

A few examples are provided [here](./examples/) to show case how to configure the provider, and how to use some resources


## Developing

Run code-generation pipeline:
```console
make generate
```

Run against a Kubernetes cluster:

```console
make run
```

Build, push, and install:

```console
make all
```

Build binary:

```console
make build
```

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/Qbeast-io/provider-acr-cache/issues).