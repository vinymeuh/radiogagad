BINARY=radiogagad

SHELL := $(shell which bash)
ENV = /usr/bin/env

.SHELLFLAGS = -c

.ONESHELL: ;
.NOTPARALLEL: ;
.EXPORT_ALL_VARIABLES:

.PHONY: all
.DEFAULT_GOAL := help

VERSION = `git describe --tags --always`
BUILD   = `date +%FT%T%z`

GOFLAGS = -trimpath
LDFLAGS = -ldflags "-w -s -X main.buildVersion=${VERSION} -X main.buildDate=${BUILD}"

buildarm6: ## Build for Pi Zero
	GOOS=linux GOARCH=arm GOARM=6 go build ${LDFLAGS} -o ${BINARY} ./cmd/${BINARY}
.PHONY: buildarm6

buildarm7: ## Build for Pi 3
	GOOS=linux GOARCH=arm GOARM=7 go build ${LDFLAGS} -o ${BINARY} ./cmd/${BINARY}
.PHONY: buildarm7

clean: ## Delete binary
	rm -f ${BINARY}
.PHONY: clean

coverage: ## Show test coverage
	go tool cover -func=coverage.txt
	go tool cover -html=coverage.txt
.PHONY: coverage

help: ## Show Help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
.PHONY: help

test: ## Run tests
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

test-radiogaga: buildarm7 ## Test on radiogaga (not persistent)
	scp radiogagad root@radiogaga:/tmp
	ssh root@radiogaga rc-service radiogagad stop
	ssh root@radiogaga cp /tmp/radiogagad /usr/local/bin
	ssh root@radiogaga rc-service radiogagad start
.PHONY: test-radiogaga
