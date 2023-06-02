RELEASE_VERSION  =v0.1.7
SERVICE_NAME    ?=$(notdir $(shell pwd))
DOCKER_USERNAME ?=$(DOCKER_USER)
REMOTE_REPO_URL :=$(shell git config remote.origin.url)
PACKAGE_NAME    :=$(subst git@,,$(REMOTE_REPO_URL))
PACKAGE_NAME    :=$(subst .git,,$(PACKAGE_NAME))
PACKAGE_NAME    :=$(subst :,/,$(PACKAGE_NAME))

.PHONY: all
all: help

.PHONY: tidy
tidy: ## Updates the go modules and vendors all dependencies 
	go mod tidy
	go mod vendor

.PHONY: test
test: tidy ## Tests the entire project 
	go test -count=1 -race ./...

.PHONY: run
run: tidy ## Runs uncompiled code in Dapr
	dapr run \
		--app-id $(SERVICE_NAME) \
		--app-port 50001 \
		--app-protocol grpc \
		--dapr-http-port 3500 \
		--components-path ./config \
		go run main.go

.PHONY: build
build: tidy ## Builds local release binary
	CGO_ENABLED=0 go build -a -tags netgo -mod vendor -o bin/$(SERVICE_NAME) .

.PHONY: invoke
invoke: ## Invoke service with sample JSON content using Dapr API
	curl -d '{ "message": "hi" }' \
     -H "Content-type: application/json" \
     "http://localhost:3500/v1.0/invoke/$(SERVICE_NAME)/method/echo"

.PHONY: image
image: tidy ## Builds and publish docker image 
	docker build -t "$(DOCKER_USERNAME)/$(SERVICE_NAME):$(RELEASE_VERSION)" .
	docker push "$(DOCKER_USERNAME)/$(SERVICE_NAME):$(RELEASE_VERSION)"

.PHONY: lint
lint: ## Lints the entire project 
	golangci-lint run --timeout=3m

.PHONY: tag
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: clean
clean: ## Cleans up generated files 
	go clean
	rm -fr ./bin
	rm -fr ./vendor

.PHONY: init
init: clean ## Resets go modules 
	rm -f go.*
	go mod init $(PACKAGE_NAME)
	go mod tidy 
	go mod vendor 

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
