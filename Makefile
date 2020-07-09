SHELL = /bin/bash
OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')

# Build variables
BINARY_NAME = iofogctl
BUILD_DIR ?= bin
PACKAGE_DIR = cmd/iofogctl
MAJOR ?= $(shell cat version | grep MAJOR | sed 's/MAJOR=//g')
MINOR ?= $(shell cat version | grep MINOR | sed 's/MINOR=//g')
PATCH ?= $(shell cat version | grep PATCH | sed 's/PATCH=//g')
SUFFIX ?= $(shell cat version | grep SUFFIX | sed 's/SUFFIX=//g')
VERSION = $(MAJOR).$(MINOR).$(PATCH)$(SUFFIX)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null)
BUILD_DATE ?= $(shell date +%FT%T%z)
PREFIX = github.com/eclipse-iofog/iofogctl/v2/pkg/util
LDFLAGS += -X $(PREFIX).versionNumber=$(VERSION) -X $(PREFIX).commit=$(COMMIT) -X $(PREFIX).date=$(BUILD_DATE) -X $(PREFIX).platform=$(GOOS)/$(GOARCH)
LDFLAGS += -X $(PREFIX).portManagerTag=develop
LDFLAGS += -X $(PREFIX).kubeletTag=develop
LDFLAGS += -X $(PREFIX).operatorTag=develop
LDFLAGS += -X $(PREFIX).proxyTag=develop
LDFLAGS += -X $(PREFIX).routerTag=develop
LDFLAGS += -X $(PREFIX).controllerTag=develop
LDFLAGS += -X $(PREFIX).agentTag=develop
LDFLAGS += -X $(PREFIX).controllerVersion=$(VERSION)
LDFLAGS += -X $(PREFIX).agentVersion=$(VERSION)
LDFLAGS += -X $(PREFIX).repo=gcr.io/focal-freedom-236620
GO_SDK_MODULE = iofog-go-sdk/v2@develop
OPERATOR_MODULE = iofog-operator/v2@develop
REPORTS_DIR ?= reports
TEST_RESULTS ?= TEST-iofogctl.txt
TEST_REPORT ?= TEST-iofogctl.xml

# Go variables
export CGO_ENABLED ?= 0
export GOOS ?= $(OS)
export GOARCH ?= amd64
GOLANG_VERSION = 1.12
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./client/*")

.PHONY: all
all: init build install ## Build, and install binary

.PHONY: clean
clean: ## Clean the working area and the project
	rm -rf $(BUILD_DIR)/
	rm -rf $(REPORTS_DIR)

.PHONY: init
init: ## Init git repository
	@cp gitHooks/* .git/hooks/

.PHONY: modules
modules: get vendor ## Get modules and vendor them

.PHONY: get
get: ## Pull modules
	@for module in $(GO_SDK_MODULE) $(OPERATOR_MODULE); do \
		go get github.com/eclipse-iofog/$$module; \
	done
	@go get github.com/eclipse-iofog/iofogctl@v1.3

.PHONY: vendor
vendor: ## Vendor all modules
	@go mod vendor
	@for module in GeertJohan akavel jessevdk jstemmer nkovacs valyala; do \
		git checkout -- vendor/github.com/$$module; \
	done

.PHONY: build
build: GOARGS += -mod=vendor -tags "$(GOTAGS)" -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)
build: fmt ## Build the binary
ifneq ($(IGNORE_GOLANG_VERSION_REQ), 1)
	@printf "$(GOLANG_VERSION)\n$$(go version | awk '{sub(/^go/, "", $$3);print $$3}')" | sort -t '.' -k 1,1 -k 2,2 -k 3,3 -g | head -1 | grep -q -E "^$(GOLANG_VERSION)$$" || (printf "Required Go version is $(GOLANG_VERSION)\nInstalled: `go version`" && exit 1)
endif
	@cd pkg/util && rice embed-go
	@go build -v $(GOARGS) $(PACKAGE_DIR)/main.go

.PHONY: install
install: ## Install the iofogctl binary to /usr/local/bin
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin

.PHONY: fmt
fmt: ## Format the source
	@gofmt -s -w $(GOFILES_NOVENDOR)

.PHONY: test
test: ## Run unit tests
	mkdir -p $(REPORTS_DIR)
	rm -f $(REPORTS_DIR)/*
	set -o pipefail; find ./internal -name '*_test.go' -not -path vendor/ | sed -E "s|(/.*/).*_test.go|\1|g" | xargs -n1 go test -mod=vendor -ldflags "$(LDFLAGS)" -coverprofile=$(REPORTS_DIR)/coverage.txt -v -parallel 1 2>&1 | tee $(REPORTS_DIR)/$(TEST_RESULTS)
	cat $(REPORTS_DIR)/$(TEST_RESULTS) | go-junit-report -set-exit-code > $(REPORTS_DIR)/$(TEST_REPORT)

.PHONY: list
list: ## List all make targets
	@$(MAKE) -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help: ## Get help output
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Variable outputting/exporting rules
var-%: ; @echo $($*)
varexport-%: ; @echo $*=$($*)
