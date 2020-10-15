MODULE := github.com/mongodb-forks/drone-helm3
CMD_NAME ?= drone-helm

RUN ?= .*
PKG ?= ./...
.PHONY: test
test: ## Run tests in local environment
	golangci-lint run --timeout=5m $(PKG)
	go test -cover -run=$(RUN) $(PKG)

.PHONY: docker
docker: ## Build local development docker image with cached go modules, builds, and tests
	@docker build -f build/Dockerfile-test -t $(CMD_NAME):latest .

.PHONY: docker-test
docker-test: ## Run tests using local development docker image
	@docker run -v $(shell pwd):/go/src/$(MODULE):delegated $(CMD_NAME) make test RUN=$(RUN) PKG=$(PKG)

.PHONY: docker-snyk
docker-snyk: ## Run local snyk scan, SNYK_TOKEN environment variable must be set
	@docker run --rm -e SNYK_TOKEN -w /go/src/$(MODULE) -v $(shell pwd):/go/src/$(MODULE):delegated snyk/snyk:golang

.PHONY: drone-sign
drone-sign:
	drone sign mongodb-forks/drone-helm3 --save

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'