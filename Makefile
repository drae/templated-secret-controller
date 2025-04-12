# Makefile for templated-secret-controller

.DEFAULT_GOAL := build

OUT_DIR ?= build/

# Image settings
REPOSITORY_OWNER ?= drae
REPOSITORY_NAME ?= templated-secret-controller

# Build settings
PLATFORMS ?= linux/amd64
LDFLAGS := -ldflags="-X 'main.Version=$(TAG)' -buildid="
BUILD_FLAGS := -trimpath -mod=vendor $(LDFLAGS)


# Find or download controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.2 ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif


# Compile binary
fmt:
	go fmt ./...

vet:
	go vet ./...

build: fmt vet
	go build $(BUILD_FLAGS) -o build/templated-secret-controller ./cmd/controller/...


# Generate client code for CRDs
client-gen:
	@echo "=== Generating client code for CRDs ==="
	@chmod +x hack/client-codegen.sh
	@./hack/client-codegen.sh

# Verify client code is up-to-date (for CI)
verify-client-gen: client-gen
	@echo "=== Verifying client code is up-to-date ==="
	@if [ -n "$$(git status --porcelain pkg/client)" ]; then \
		echo "Client code is not up-to-date, run 'make client-gen'"; \
		git status --porcelain pkg/client; \
		exit 1; \
	fi


# Code generation
generate: controller-gen client-gen
	$(CONTROLLER_GEN) object:headerFile="code-header-template.txt" paths="./pkg/apis/..."

# Run manifests generation
manifests: controller-gen
	$(CONTROLLER_GEN) crd paths="./pkg/apis/templatedsecret/v1alpha1" output:crd:artifacts:config=config/kustomize/base/crds

clean:
	@rm -rf build
	@go clean -cache


# Run tests
test:
	NAMESPACE=templated-secret-dev go test ./... -coverprofile cover.out

coverage: test
	NAMESPACE=templated-secret-dev go tool cover -func=cover.out

coverage-html: test
	NAMESPACE=templated-secret-dev go tool cover -html=cover.out


# Build the docker image
docker-build:
	REPOSITORY_OWNER=$(REPOSITORY_OWNER) REPOSITORY_NAME=$(REPOSITORY_NAME) goreleaser release --snapshot --clean --skip=publish,sign

# Phony
.PHONY : controller-gen fmt vet test coverage coverage-html build generate manifests docker-build client-gen verify-client-gen clean