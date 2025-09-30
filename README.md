# templated-secret-controller

[![codecov](https://codecov.io/gh/drae/templated-secret-controller/branch/main/graph/badge.svg?token=XCY5S8HZK1)](https://codecov.io/gh/drae/templated-secret-controller)

A Kubernetes controller for generating secrets from existing resources.

## Overview

templated-secret-controller provides a custom resource for generating and managing Kubernetes secrets by combining data from other resources:

- **SecretTemplate**: Generate secrets using data from existing Kubernetes resources, including other Secrets, ConfigMaps, Services, and more.

> **Note on naming:** While the controller is named "templated-secret-controller" (with a hyphen), the API group remains "templatedsecret.starstreak.dev" (without a hyphen) for compatibility with code generation tools.

## Key Features

- Generate secrets by combining data from multiple existing Kubernetes resources
- Template data using JSONPath expressions to extract specific values
- Continuously reconcile secrets when source resources change
- Support for various Kubernetes resource types as input sources
- Role-based access control for reading input resources
- Optional cross-namespace Secret inputs gated by feature flag + per-Secret export annotation
- Health and readiness endpoints with cache-sync aware readiness
- Metrics endpoint (Prometheus friendly) and optional ServiceMonitor
- SBOM generation and image signing (Cosign) in release pipeline

## Installation

### Using Helm

Deploy the controller using Helm:

```shell
# Clone the repository
git clone https://github.com/drae/templated-secret-controller.git
cd templated-secret-controller

# Install with default settings
helm install templated-secret-controller ./charts/templated-secret-controller

# Install with metrics disabled
helm install templated-secret-controller ./charts/templated-secret-controller --set metrics.enabled=false

# Install without CRDs (if you've already installed them)
helm install templated-secret-controller ./charts/templated-secret-controller --set crds.create=false
```

For more information on configuration options, see the [Helm chart README](./charts/templated-secret-controller/README.md).

### Using Kustomize

Deploy the controller directly with kustomize:

```shell
# Production deployment
kubectl apply -k https://github.com/drae/templated-secret-controller/config/kustomize/overlays/prod

# Development deployment
kubectl apply -k https://github.com/drae/templated-secret-controller/config/kustomize/overlays/dev
```

### Using pre-built manifests

Download and apply the latest release manifests:

```shell
kubectl apply -f https://github.com/drae/templated-secret-controller/releases/latest/download/templated-secret-controller.yaml
```

## Example

```yaml
apiVersion: templatedsecret.starstreak.dev/v1alpha1
kind: SecretTemplate
metadata:
  name: combined-secret
spec:
  inputResources:
    - name: secret1
      ref:
        apiVersion: v1
        kind: Secret
        name: secret1
    - name: secret2
      ref:
        apiVersion: v1
        kind: Secret
        name: secret2
  template:
    type: mysecrettype
    data:
      key1: $(.secret1.data.key1)
      key2: $(.secret1.data.key2)
      key3: $(.secret2.data.key3)
      key4: $(.secret2.data.key4)
```

See [the SecretTemplate documentation](docs/secret-template.md) for more detailed examples and explanations.

## Local Development

This project uses standard Go tools and Kubernetes controller patterns:

```shell
# Build
make build

# Run tests
make test

# Build container image
make docker-build

# Generate a local SBOM (requires syft) after build
make sbom

# Run controller locally with probes and metrics disabled
./build/templated-secret-controller \
  --metrics-bind-address=0 \
  --health-probe-bind-address=0 \
  --enable-cross-namespace-secret-inputs=false
```

## CI/CD

The project uses GitHub Actions for continuous integration and deployment:

- CI workflow runs on PRs and pushes to main
- Release workflow triggers on tags formatted as 'v*'
- Images are published to GitHub Container Registry
- Release pipeline generates a CycloneDX SBOM per archive and signs images & multi-arch manifests with Cosign (when not a snapshot)

### Probes & Readiness

The controller exposes (by default):

- Metrics: `:8080/metrics`
- Health: `:8081/healthz`
- Readiness: `:8081/readyz` (only returns 200 after informer caches sync and leader election, if enabled)

Disable by setting Helm values `metrics.enabled=false` and/or `probes.enabled=false` (or flags `--metrics-bind-address=0`, `--health-probe-bind-address=0`).

### Cross-Namespace Secret Inputs

Enable via `--enable-cross-namespace-secret-inputs` (Helm: `crossNamespace.enabled=true`). Source Secrets must include annotation:

```yaml
templatedsecret.starstreak.dev/export-to-namespaces: "ns-a,ns-b"   # or "*"
```

Warnings (condition `CrossNamespaceInputDegraded`) appear if a referenced source namespace is not watched (updates may not propagate).

### SBOM Retrieval

Release assets include `sbom_<project>_<version>.cdx.json`. Locally generate with `make sbom` (syft required in PATH).

### Image & Manifest Signing

All per-arch images and multi-arch manifests are signed with Cosign. To verify (example):

```shell
cosign verify ghcr.io/drae/templated-secret-controller:<version>
```

Additional policy tooling (e.g., Kyverno / Ratify) can enforce signature & SBOM presence.

## Code Coverage

![Code coverage graph](https://codecov.io/gh/drae/templated-secret-controller/graphs/tree.svg?token=XCY5S8HZK1)
