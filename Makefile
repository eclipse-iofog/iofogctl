SHELL = /bin/bash
OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Build variables
include versions.mk

# Build variables
FLAVOR ?= iofog
BINARY_NAME ?= iofogctl
BUILD_DIR ?= bin
PACKAGE_DIR ?= cmd/iofogctl
GOTAGS ?= containers_image_openpgp,exclude_graphdriver_btrfs
export CGO_ENABLED=1
LATEST_TAG = $(shell git for-each-ref refs/tags --sort=-taggerdate --format='%(refname)' | tail -n1 | sed "s|refs/tags/||")
MAJOR ?= $(shell echo "$(LATEST_TAG)" | tr -d "v" | sed "s|-.*||" | sed -E "s|(.)\..\..|\1|g")
MINOR ?= $(shell echo "$(LATEST_TAG)" | tr -d "v" | sed "s|-.*||" | sed -E "s|.\.(.)\..|\1|g")
PATCH ?= $(shell echo "$(LATEST_TAG)" | tr -d "v" | sed "s|-.*||" | sed -E "s|.\..\.(.)|\1|g")
TAG_SUFFIX = $(shell echo "$(LATEST_TAG)" | sed "s|v$(MAJOR)\.$(MINOR)\.$(PATCH)||")
DEV_SUFFIX = -dev
SUFFIX ?= $(shell [ -z "$(shell git tag --points-at HEAD)" ] && echo "$(DEV_SUFFIX)" || echo "$(TAG_SUFFIX)")
VERSION ?= $(MAJOR).$(MINOR).$(PATCH)$(SUFFIX)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PREFIX = github.com/eclipse-iofog/iofogctl/pkg/util

ifeq ($(FLAVOR),datasance)
  CLI_BINARY_NAME = potctl
  CLI_CRD_GROUP = datasance.com
  CLI_API_VERSION = datasance.com/v3
  IMAGE_REGISTRY = ghcr.io/datasance
  CLI_DOCS_URL = https://docs.datasance.com
  PACKAGE_REPO_BASE = downloads.datasance.com
  OCI_SOURCE_REPO = https://github.com/Datasance/potctl
else
  CLI_BINARY_NAME = iofogctl
  CLI_CRD_GROUP = iofog.org
  CLI_API_VERSION = iofog.org/v3
  IMAGE_REGISTRY = ghcr.io/eclipse-iofog
  CLI_DOCS_URL = https://iofog.org
  PACKAGE_REPO_BASE = https://packagecloud.io/iofog
  OCI_SOURCE_REPO = https://github.com/eclipse-iofog/iofogctl
endif

LDFLAGS += -X $(PREFIX).versionNumber=$(VERSION) -X $(PREFIX).commit=$(COMMIT) -X $(PREFIX).date=$(BUILD_DATE) -X $(PREFIX).platform=$(GOOS)/$(GOARCH)
LDFLAGS += -X $(PREFIX).cliBinaryName=$(CLI_BINARY_NAME)
LDFLAGS += -X $(PREFIX).cliCrdGroup=$(CLI_CRD_GROUP)
LDFLAGS += -X $(PREFIX).cliApiVersion=$(CLI_API_VERSION)
LDFLAGS += -X $(PREFIX).imageRegistry=$(IMAGE_REGISTRY)
LDFLAGS += -X $(PREFIX).cliDocsUrl=$(CLI_DOCS_URL)
LDFLAGS += -X $(PREFIX).packageRepoBase=$(PACKAGE_REPO_BASE)
LDFLAGS += -X $(PREFIX).ociSourceRepo=$(OCI_SOURCE_REPO)
LDFLAGS += -X $(PREFIX).operatorTag=$(OPERATOR_VERSION)
LDFLAGS += -X $(PREFIX).routerTag=$(ROUTER_VERSION)
LDFLAGS += -X $(PREFIX).controllerTag=$(CONTROLLER_VERSION)
LDFLAGS += -X $(PREFIX).natsTag=$(NATS_VERSION)
LDFLAGS += -X $(PREFIX).edgeletTag=$(EDGELET_VERSION)
LDFLAGS += -X $(PREFIX).controllerVersion=$(CONTROLLER_VERSION)
LDFLAGS += -X $(PREFIX).edgeletVersion=$(EDGELET_VERSION)
LDFLAGS += -X $(PREFIX).debuggerTag=latest

REPORTS_DIR ?= reports
TEST_RESULTS ?= TEST-iofogctl.txt
TEST_REPORT ?= TEST-iofogctl.xml

.PHONY: all
all: bootstrap build install ## Bootstrap env, build and install binary

.PHONY: bootstrap
bootstrap: ## Bootstrap environment
	@cp gitHooks/* .git/hooks/
	@script/bootstrap.sh

.PHONY: verify-gpgme
verify-gpgme:
	@if ! pkg-config --exists gpgme; then \
		echo "Missing gpgme development headers. Install via 'brew install gpgme' or 'sudo apt-get install libgpgme-dev'."; \
		exit 1; \
	fi

.PHONY: potctl
potctl: ## Build potctl binary
	@$(MAKE) FLAVOR=datasance BINARY_NAME=potctl PACKAGE_DIR=cmd/potctl build

.PHONY: iofogctl
iofogctl: ## Build iofogctl binary
	@$(MAKE) FLAVOR=iofog BINARY_NAME=iofogctl PACKAGE_DIR=cmd/iofogctl build

.PHONY: build
build: GOARGS += -tags "$(GOTAGS)" -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)
build: fmt ## Build the binary
	@go build -v $(GOARGS) $(PACKAGE_DIR)/main.go

.PHONY: install
install: ## Install the binary
	@GOBIN=$$(go env GOPATH)/bin go install -tags "$(GOTAGS)" -ldflags "$(LDFLAGS)" ./$(PACKAGE_DIR)/

.PHONY: lint
lint: golangci-lint fmt ## Lint the source
	@$(GOLANGCI_LINT) run --timeout 5m0s

golangci-lint: ## Install golangci
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.4 ;\
	}
GOLANGCI_LINT=$(GOBIN)/golangci-lint
else
GOLANGCI_LINT=$(shell which golangci-lint)
endif

.PHONY: fmt
fmt: ## Format the source
	@gofmt -s -w .

.PHONY: test
test: ## Run unit tests
	mkdir -p $(REPORTS_DIR)
	rm -f $(REPORTS_DIR)/*
	set -o pipefail; find ./internal ./pkg -name '*_test.go' -not -path vendor | sed -E "s|(/.*/).*_test.go|\1|g" | xargs -n1 go test -tags "$(GOTAGS)" -ldflags "$(LDFLAGS)" -coverprofile=$(REPORTS_DIR)/coverage.txt -v -parallel 1 2>&1 | tee $(REPORTS_DIR)/$(TEST_RESULTS)
	cat $(REPORTS_DIR)/$(TEST_RESULTS) | go-junit-report -set-exit-code > $(REPORTS_DIR)/$(TEST_REPORT)

.PHONY: list
list: ## List all make targets
	@$(MAKE) -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: clean
clean: ## Clean working env
	rm -rf $(BUILD_DIR)/
	rm -rf $(REPORTS_DIR)

.PHONY: help
.DEFAULT_GOAL := help
help: ## Get help output
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Variable outputting/exporting rules
var-%: ; @echo $($*)
varexport-%: ; @echo $*=$($*)
