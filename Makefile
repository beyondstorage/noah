SHELL := /bin/bash
GO_BUILD_OPTION := -trimpath -tags netgo

.PHONY: all check format vet build clean test generate

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check      to format and vet"
	@echo "  build      to create bin directory and build noah"
	@echo "  clean      to clean build and test files"
	@echo "  test       to run test"

check: format vet

format:
	@echo "go fmt"
	@go fmt ./...
	@echo "ok"

vet:
	@echo "go vet"
	@go vet ./...
	@echo "ok"

setup:
	go install github.com/golang/protobuf/protoc-gen-go
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate:
	@echo "generate code..."
	@go generate ./...
	@echo "Done"

build: tidy setup generate check
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
