SHELL = /bin/bash
OS = $(shell uname -s)

# Project variables
PACKAGE = github.com/eclipse-iofog/iofogctl
BINARY_NAME = iofogctl
IMAGE = iofog/cli

# Build variables
BUILD_DIR ?= bin
BUILD_PACKAGE = $(PACKAGE)/cmd/iofogctl
VERSION ?= $(shell git rev-parse --abbrev-ref HEAD)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE ?= $(shell date +%FT%T%z)
LDFLAGS += -X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildDate=$(BUILD_DATE)
export CGO_ENABLED ?= 0
ifeq ($(VERBOSE), 1)
	GOARGS += -v
endif

GOLANG_VERSION = 1.12

GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./client/*")

.PHONY: all
all: dep build install ## Get deps, build, and install binary

.PHONY: clean
clean: ## Clean the working area and the project
	rm -rf $(BUILD_DIR)/ vendor/

.PHONY: dep
dep: ## Install dependencies
	dep ensure -v -vendor-only

.PHONY: build
build: GOARGS += -tags "$(GOTAGS)" -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)
build: ## Build the binary
ifneq ($(IGNORE_GOLANG_VERSION_REQ), 1)
	@printf "$(GOLANG_VERSION)\n$$(go version | awk '{sub(/^go/, "", $$3);print $$3}')" | sort -t '.' -k 1,1 -k 2,2 -k 3,3 -g | head -1 | grep -q -E "^$(GOLANG_VERSION)$$" || (printf "Required Go version is $(GOLANG_VERSION)\nInstalled: `go version`" && exit 1)
endif
	go build $(GOARGS) $(BUILD_PACKAGE)

.PHONY: install
install: ## Install the ioFog binary to /usr/local/bin
	go install github.com/eclipse-iofog/iofogctl/cmd/iofogctl

.PHONY: fmt
fmt: ## Format the source
	@gofmt -s -w $(GOFILES_NOVENDOR)

.PHONY: test
test: ## Run unit tests
	set -o pipefail; go list ./... | xargs -n1 go test $(GOARGS) -v -parallel 1 2>&1 | tee test.txt

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