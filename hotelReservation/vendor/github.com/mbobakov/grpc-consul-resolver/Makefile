.DEFAULT_GOAL := help

BIN_DIR = ${PWD}/bin

.PHONY: clean download tools lint test generate test-integration terraform-fmt

RED=$(shell tput -T xterm setaf 1)
GREEN=$(shell tput -T xterm setaf 2)
YELLOW=$(shell tput -T xterm setaf 3)
RESET=$(shell tput -T xterm sgr0)

DOCKER_IMAGE ?= connectivity_controller


export GOPRIVATE=github.com/reemote/*
export PATH := ${BIN_DIR}:$(PATH)
export GOBIN := ${BIN_DIR}

tools: ## Installing tools from tools.go
	echo Installing tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	go install github.com/matryer/moq@v0.3.1

.PHONY: clean
clean: ## run all cleanup tasks
	go clean ./...
	rm -rf $(BIN_DIR)
	rm -rf ./builds

test: generate ## Run unit tests
	go test -count=1 -v ./...
	@echo ""
	@echo "${GREEN} All tests passed ✅"
	@echo "${RESET}"

test-integration:  ## Integration test
test-integration:
	go test -timeout 300s -tags integration -v ./...
	@echo ""
	@echo "${GREEN} All tests passed ✅"
	@echo "${RESET}"

lint: tools ## Run linter
	${BIN_DIR}/golangci-lint --color=always run ./... -v --timeout 15m

generate: tools
# Because go generate doesn't support subschell we need to call it directly
	- ./bin/moq -pkg consul -out ./mocks_grpc_test.go $$(go list -f '{{.Dir}}' google.golang.org/grpc/resolver) ClientConn
	- go generate -x ./...

help: ## Display help screen
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
	{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
