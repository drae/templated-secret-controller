#!/bin/bash

set -euo pipefail

# This script assumes code generator tools have been installed by 'make ensure-codegen-tools'
# and are available in the PATH (which the Makefile ensures).

# Get the absolute path to the project root directory
PROJECT_DIR="$(pwd)"

# Get the correct Go module paths
MODULE_PATH="github.com/drae/templated-secret-controller"
API_PACKAGE="${MODULE_PATH}/pkg/apis/templatedsecret/v1alpha1"
HEADER_FILE="${PROJECT_DIR}/code-header-template.txt"
OUTPUT_PATH=$(pwd)

# Clean up any existing generated code
echo "=== Cleaning up existing client code ==="
rm -rf pkg/client
mkdir -p pkg/client

# Generate deepcopy methods
echo "Generating deepcopy methods..."
deepcopy-gen \
  --go-header-file ${HEADER_FILE} \
  --output-file zz_generated.deepcopy.go \
  --bounding-dirs ${API_PACKAGE} \
  ${API_PACKAGE}

# Generate client code - adjusted to point directly to v1alpha1
echo "Generating client code..."
client-gen \
  --go-header-file ${HEADER_FILE} \
  --clientset-name versioned \
  --input-base "" \
  --input ${MODULE_PATH}/pkg/apis/templatedsecret/v1alpha1 \
  --output-pkg ${MODULE_PATH}/pkg/client/clientset/ \
  --output-dir ${OUTPUT_PATH}/pkg/client/clientset

# Generate lister code
echo "Generating lister code..."
lister-gen \
  --go-header-file ${HEADER_FILE} \
  --output-pkg ${MODULE_PATH}/pkg/client/listers/ \
  --output-dir ${OUTPUT_PATH}/pkg/client/listers \
  ${API_PACKAGE}

# Generate informer code
echo "Generating informer code..."
informer-gen \
  --go-header-file ${HEADER_FILE} \
  --versioned-clientset-package ${MODULE_PATH}/pkg/client/clientset/versioned \
  --listers-package ${MODULE_PATH}/pkg/client/listers \
  --output-pkg ${MODULE_PATH}/pkg/client/informers/ \
  --output-dir ${OUTPUT_PATH}/pkg/client/informers/ \
  ${API_PACKAGE}

# Install vendor dependencies and cleanup
go mod vendor
go mod tidy

echo "=== Code generation complete ==="
echo "Generated files:"
find pkg/client -type f -name "*.go" | head -n 10
