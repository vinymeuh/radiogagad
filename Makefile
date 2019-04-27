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

LDFLAGS = -ldflags "-w -s -X main.Version=${VERSION} -X main.Build=${BUILD}"


build: ## Build for current host
	go build ${LDFLAGS} -o ${BINARY}

buildarm6: ## Build for Pi Zero
	GOOS=linux GOARCH=arm GOARM=6 go build ${LDFLAGS} -o ${BINARY}

buildarm7: ## Build for Pi 3
	GOOS=linux GOARCH=arm GOARM=7 go build ${LDFLAGS} -o ${BINARY}

clean: ## Delete binary
	rm -f ${BINARY}

coverage: ## Show test coverage
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

help: ## Show Help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	go test -coverprofile=coverage.out ./...
