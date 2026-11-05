GO ?= go
GOLANGCI_LINT ?= golangci-lint
POWERSHELL ?= powershell
EXEEXT := $(if $(filter windows,$(shell $(GO) env GOOS)),.exe,)

.PHONY: build test lint fmt diagrams
build:
	mkdir -p bin
	$(GO) build -o bin/gateway$(EXEEXT) ./cmd/gateway
test:
	$(GO) test ./...
lint:
	$(GOLANGCI_LINT) run
fmt:
	$(GO) fmt ./...
diagrams:
	$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/render-plantuml.ps1
