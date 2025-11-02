# Makefile for symval

# Ignore built-in inference rules that determine eg how to build object files from C source code.
.SUFFIXES:

# Warn if any variables are undefined.
MAKEFLAGS += --warn-undefined-variables

# Fail if any command fails.
.SHELLFLAGS := -euc

# Show a nice table of Make targets.
.PHONY: help
help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Disable CGO for all builds.
# We want to make sure all our code works without CGO, for faster startup times in Lambda.
CGO_ENABLED ?= 0
export CGO_ENABLED

GOSRCS := $(shell find . -type f -name '*.go' -or -name 'go.mod' -or -name 'go.sum')

dist/symval: $(GOSRCS)
	go build -o dist/symval cmd/symval/main.go

dist/reattestbatch: $(GOSRCS)
	go build -o dist/reattestbatch cmd/reattestbatch/main.go

dist/httpapi: $(GOSRCS)
	go build -o dist/httpapi cmd/httpapi/main.go

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: all
all: dist/symval dist/reattestbatch dist/httpapi ## Build all binaries

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf dist
