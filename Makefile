.PHONY: all build run test test-cover test-verbose lint fmt vet clean install dev

APP_NAME := skill-man
GO := go
GOFLAGS :=
BUILD_DIR := bin

all: fmt vet test build

build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

run: build
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	$(GO) run ./cmd/$(APP_NAME)

test:
	$(GO) test ./...

test-verbose:
	$(GO) test -v ./...

test-cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

test-cover-html: test-cover
	$(GO) tool cover -html=coverage.out

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

install:
	$(GO) install ./cmd/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR) coverage.out
