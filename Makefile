.POSIX:
SHELL := /bin/bash
GO := $(shell which go)
BIN_NAME := httpbin
DOCKER_TAG := impossible98/$(BIN_NAME)
# Build the application
.PHONY: build
build: fmt
	@echo -e "\033[32mBuilding the application...\033[0m"
	$(GO) build -ldflags "-s -w" -o ./dist/$(BIN_NAME) ./cmd/$(BIN_NAME)/main.go
	@echo -e "\033[32mBuild finished.\033[0m"
# Local development
dev:
	$(GO) run ./cmd/$(BIN_NAME)/main.go
# Build docker image
docker:
	docker buildx create --name $(BIN_NAME)
	docker buildx use $(BIN_NAME)
	docker buildx build -f ./build/docker/Dockerfile --push --platform linux/amd64,linux/arm64 -t $(DOCKER_TAG) .
	docker buildx rm $(BIN_NAME)
# Format the code
fmt:
	@echo -e "\033[32mFormatting the code...\033[0m"
	$(GO) fmt ./...
	@echo -e "\033[32mFormatting finished.\033[0m"
# Install dependencies
install:
	@echo -e "\033[32mInstalling dependencies...\033[0m"
	$(GO) mod download
	@echo -e "\033[32mDependencies installed.\033[0m"
# Lint the code
lint:
	@echo -e "\033[32mLinting the code...\033[0m"
	$(GO) vet ./...
	@echo -e "\033[32mLinting finished.\033[0m"
# Run tests
test:
	$(GO) test -race ./...
