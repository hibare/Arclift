SHELL=/bin/bash

UI := $(shell id -u)
GID := $(shell id -g)
MAKEFLAGS += -s
DOCKER_COMPOSE_PREFIX = HOST_UID=${UID} HOST_GID=${GID} docker-compose -f docker-compose.dev.yml

# Bold
BCYAN=\033[1;36m
BBLUE=\033[1;34m

# No color (Reset)
NC=\033[0m

.DEFAULT_GOAL := help

.PHONY: init
init: ## Initialize the project
	$(MAKE) install-golangci-lint
	$(MAKE) install-pre-commit

.PHONY: install-golangci-lint
install-golangci-lint: ## Install golangci-lint
ifeq (, $(shell which golangci-lint))
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin
endif

.PHONY: install-pre-commit
install-pre-commit: ## Install pre-commit
	pre-commit install

.PHONY: dev
dev: ## Start development environment
	mkdir .ignored/test-backup
	mkdir .ignored/test-backup/test-dir
	touch .ignored/test-backup/test-dir/test-file
	touch .ignored/test-backup/test-dir/test-file2
	touch .ignored/test-backup/test-file3
	${DOCKER_COMPOSE_PREFIX} up minio create-buckets

.PHONY: clean
clean: ## Clean up
	${DOCKER_COMPOSE_PREFIX} down
	go mod tidy

.PHONY: test
test: ## Run tests
	go test ./... -cover

.PHONY: help
help: ## Display this help
		@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BCYAN)%-18s$(NC)%s\n", $$1, $$2}'
