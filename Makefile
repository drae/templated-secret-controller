# Makefile for templated-secret-controller

.DEFAULT_GOAL := build

OUT_DIR ?= build/

# Tool versions
CODE_GENERATOR_VERSION ?= v0.32.3
CONTROLLER_TOOLS_VERSION ?= v0.17.3

# Image settings
REPOSITORY_OWNER ?= drae
REPOSITORY_NAME ?= templated-secret-controller

# Pin controller version
CONTROLLER_GEN=$(GOBIN)/controller-gen

# Build settings
PLATFORMS ?= linux/amd64
LDFLAGS := -ldflags="-X 'main.Version=$(TAG)' -buildid="
BUILD_FLAGS := -trimpath -mod=vendor $(LDFLAGS)

# Cache directories with absolute path
CACHE_DIR := $(shell pwd)/.cache
CACHE_BIN := $(CACHE_DIR)/bin

# 
.PHONY: controller-gen ensure-codegen-tools fmt vet build client-gen verify-client-gen generate manifests verify-generated clean clean-cache test test-unit test-integration coverage coverage-html docker-build

# Find or download controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e; \
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d); \
	cd $$CONTROLLER_GEN_TMP_DIR; \
	go mod init tmp; \
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION); \
	GOBIN=$(GOBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION); \
	rm -rf $$CONTROLLER_GEN_TMP_DIR; \
	}
endif

# Ensure code generation tools are available
ensure-codegen-tools:
	@echo "=== Installing code generator tools ==="
	@mkdir -p $(CACHE_BIN)
	@GOBIN=$(CACHE_BIN) go install k8s.io/code-generator/cmd/client-gen@$(CODE_GENERATOR_VERSION)
	@GOBIN=$(CACHE_BIN) go install k8s.io/code-generator/cmd/lister-gen@$(CODE_GENERATOR_VERSION)
	@GOBIN=$(CACHE_BIN) go install k8s.io/code-generator/cmd/informer-gen@$(CODE_GENERATOR_VERSION)
	@GOBIN=$(CACHE_BIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(CODE_GENERATOR_VERSION)
	@echo "=== Code generator tools installed to $(CACHE_BIN) ==="

# Compile binary
fmt:
	go fmt ./...

vet:
	go vet ./...

build: verify-generated fmt vet
	go build $(BUILD_FLAGS) -o build/templated-secret-controller ./cmd/controller/...

# Generate client code for CRDs
client-gen: ensure-codegen-tools
	@echo "=== Generating client code for CRDs ==="
	@PATH=$(CACHE_BIN):$$PATH ./hack/client-codegen.sh

# Verify client code is up-to-date (for CI)
verify-client-gen: client-gen
	@echo "=== Verifying client code is up-to-date ==="
	@if [ -n "$$(git status --porcelain pkg/client)" ]; then \
		echo "ERROR: Client code is not up-to-date. Run 'make client-gen' and commit the changes."; \
		git status --porcelain pkg/client; \
		exit 1; \
	fi
	@echo "Client code is up-to-date."

# Code generation
generate: controller-gen client-gen
	$(CONTROLLER_GEN) object:headerFile="code-header-template.txt" paths="./pkg/apis/..."

# Run manifests generation
manifests: controller-gen
	$(CONTROLLER_GEN) crd paths="./pkg/apis/templatedsecret/v1alpha1" output:crd:artifacts:config=config/kustomize/base/crds

# Verify all generated code is up-to-date
verify-generated: verify-client-gen
	@echo "=== All generated code is up-to-date ==="

# Clean up
clean: clean-cache
	rm -rfv "$(OUT_DIR)"

clean-cache:
	@echo "=== Cleaning cached tools ==="
	@rm -rf $(CACHE_DIR)

# Run tests
test: test-unit

# Run unit tests only (excludes integration tests)
test-unit:
	go test ./... -coverprofile cover.txt

# Run integration tests (requires Kubernetes cluster)
test-integration:
	NAMESPACE=templated-secret-dev go test -tags=integration ./test/ci/... -timeout 60m -v

# Run all tests (unit and integration, requires Kubernetes cluster)
test-all: test-unit test-integration

coverage: test-unit
	go tool cover -func=cover.txt

coverage-html: test-unit
	go tool cover -html=cover.txt

# Build the docker image
docker-build:
	REPOSITORY_OWNER=$(REPOSITORY_OWNER) REPOSITORY_NAME=$(REPOSITORY_NAME) goreleaser release --snapshot --clean --skip=publish,sign