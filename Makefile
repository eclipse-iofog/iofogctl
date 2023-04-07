SHELL = /bin/bash
OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')

# Build variables
BINARY_NAME = iofogctl
BUILD_DIR ?= bin
PACKAGE_DIR = cmd/iofogctl
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
PREFIX = github.com/eclipse-iofog/iofogctl/v3/pkg/util
LDFLAGS += -X $(PREFIX).versionNumber=$(VERSION) -X $(PREFIX).commit=$(COMMIT) -X $(PREFIX).date=$(BUILD_DATE) -X $(PREFIX).platform=$(GOOS)/$(GOARCH)
LDFLAGS += -X $(PREFIX).portManagerTag=3.0.0
LDFLAGS += -X $(PREFIX).kubeletTag=3.0.0-beta1
LDFLAGS += -X $(PREFIX).operatorTag=3.2.0
LDFLAGS += -X $(PREFIX).proxyTag=3.0.0
LDFLAGS += -X $(PREFIX).routerTag=3.0.0
LDFLAGS += -X $(PREFIX).controllerTag=3.0.1
LDFLAGS += -X $(PREFIX).agentTag=3.0.1
LDFLAGS += -X $(PREFIX).controllerVersion=3.0.2
LDFLAGS += -X $(PREFIX).agentVersion=3.0.1
LDFLAGS += -X $(PREFIX).repo=iofog
GO_SDK_MODULE = iofog-go-sdk/v3@v3.0.0
OPERATOR_MODULE = iofog-operator/v3@v3.0.1
REPORTS_DIR ?= reports
TEST_RESULTS ?= TEST-iofogctl.txt
TEST_REPORT ?= TEST-iofogctl.xml

.PHONY: all
all: bootstrap build install ## Bootstrap env, build and install binary

.PHONY: bootstrap
bootstrap: ## Bootstrap environment
	@cp gitHooks/* .git/hooks/
	@script/bootstrap.sh

.PHONY: build
build: GOARGS += -tags "$(GOTAGS)" -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)
build: fmt ## Build the binary
	@cd pkg/util && rice embed-go
	@go build -v $(GOARGS) $(PACKAGE_DIR)/main.go

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/iofogctl/

.PHONY: lint
lint: golangci-lint fmt ## Lint the source
	@$(GOLANGCI_LINT) run --timeout 5m0s

golangci-lint: ## Install golangci
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1 ;\
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
	set -o pipefail; find ./internal ./pkg -name '*_test.go' -not -path vendor | sed -E "s|(/.*/).*_test.go|\1|g" | xargs -n1 go test -ldflags "$(LDFLAGS)" -coverprofile=$(REPORTS_DIR)/coverage.txt -v -parallel 1 2>&1 | tee $(REPORTS_DIR)/$(TEST_RESULTS)
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
