SHELL := /bin/bash
GO_BUILD_OPTION := -trimpath -tags netgo

.PHONY: all check format vet lint build clean test generate

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check      to format, vet and lint "
	@echo "  build      to create bin directory and build noah"
	@echo "  clean      to clean build and test files"
	@echo "  test       to run test"

check: format vet lint

format:
	@echo "go fmt"
	@go fmt ./...
	@echo "ok"

vet:
	@echo "go vet"
	@go vet ./...
	@echo "ok"

lint:
	@echo "golint"
	@golint ./...
	@echo "ok"

generate:
	@echo "generate code..."
	@go generate ./...
	@echo "Done"

build: tidy generate check
	@echo "build noah"
	@mkdir -p ./bin
	@go build ${GO_BUILD_OPTION} -race ./...
	@echo "ok"

clean:
	@rm -rf ./bin
	@rm -rf ./release
	@rm -rf ./coverage

test:
	@echo "run test"
	@go test -race -coverprofile=coverage.txt -covermode=atomic -v ./...
	@go tool cover -html="coverage.txt" -o "coverage.html"
	@echo "ok"

tidy:
	@echo "Tidy and check the go mod files"
	@go mod tidy
	@go mod verify
	@echo "Done"
