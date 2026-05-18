.PHONY: all build run test test-cover test-verbose lint fmt vet clean install dev demo

APP_NAME := skill-man
GO := go
GOFLAGS :=
BUILD_DIR := bin
GOBIN := $(shell $(GO) env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

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

demo: build
	@command -v vhs >/dev/null 2>&1 || { \
		echo "vhs not found. Install: go install github.com/charmbracelet/vhs@latest"; \
		echo "Also: brew install ffmpeg ttyd  (see docs/demo/README.md)"; \
		echo "Ensure $$(go env GOPATH)/bin is on your PATH."; \
		exit 1; \
	}
	vhs docs/demo/demo.tape

clean:
	rm -rf $(BUILD_DIR) coverage.out
