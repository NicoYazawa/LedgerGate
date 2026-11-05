GO ?= go
GOLANGCI_LINT ?= golangci-lint
EXEEXT := $(if $(filter windows,$(shell $(GO) env GOOS)),.exe,)

.PHONY: build test lint fmt
build:
	mkdir -p bin
	$(GO) build -o bin/gateway$(EXEEXT) ./cmd/gateway
test:
	$(GO) test ./...
lint:
	$(GOLANGCI_LINT) run
fmt:
	$(GO) fmt ./...
