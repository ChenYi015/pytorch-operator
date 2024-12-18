.SILENT:

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

## Location to install binaries
LOCALBIN ?= $(shell pwd)/bin

## Versions
CONTROLLER_TOOLS_VERSION ?= v0.15.0
GOLANGCI_LINT_VERSION ?= v1.61.0
GEN_CRD_API_REFERENCE_DOCS_VERSION ?= v0.3.0

## Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
GEN_CRD_API_REFERENCE_DOCS ?= $(LOCALBIN)/gen-crd-api-reference-docs-$(GEN_CRD_API_REFERENCE_DOCS_VERSION)

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate CustomResourceDefinition, RBAC and WebhookConfiguration manifests.
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true,maxDescLen=400 rbac:roleName=pytorch-operator webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: go-clean
go-clean: ## Clean up caches and output.
	@echo "cleaning up caches and output"
	go clean -cache -testcache -r -x 2>&1 >/dev/null
	-rm -rf _output

.PHONY: go-fmt
go-fmt: ## Run go fmt against code.
	@echo "Running go fmt..."
	if [ -n "$(shell go fmt ./...)" ]; then \
		echo "Go code is not formatted, need to run \"make go-fmt\" and commit the changes."; \
		false; \
	else \
	    echo "Go code is formatted."; \
	fi

.PHONY: go-vet
go-vet: ## Run go vet against code.
	@echo "Running go vet..."
	go vet ./...

.PHONY: go-lint
go-lint: golangci-lint ## Run golangci-lint linter.
	@echo "Running golangci-lint run..."
	$(GOLANGCI_LINT) run

.PHONY: go-lint-fix
go-lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	@echo "Running golangci-lint run --fix..."
	$(GOLANGCI_LINT) run --fix

.PHONY: unit-test
unit-test: envtest ## Run unit tests.
	@echo "Running unit tests..."
	go test $(shell go list ./... | grep -v /e2e) -coverprofile cover.out

##@ Dependencies

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: $(GEN_CRD_API_REFERENCE_DOCS) ## Download gen-crd-api-reference-docs locally if necessary.
$(GEN_CRD_API_REFERENCE_DOCS): $(LOCALBIN)
	$(call go-install-tool,$(GEN_CRD_API_REFERENCE_DOCS),github.com/ahmetb/gen-crd-api-reference-docs,$(GEN_CRD_API_REFERENCE_DOCS_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef
