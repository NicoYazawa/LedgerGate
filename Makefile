GO ?= go
GOLANGCI_LINT ?= golangci-lint
POWERSHELL ?= powershell
COVERAGE_MIN ?= 90
EXEEXT := $(if $(filter windows,$(shell $(GO) env GOOS)),.exe,)

.PHONY: build test coverage lint fmt diagrams
build:
	mkdir -p bin
	$(GO) build -o bin/gateway$(EXEEXT) ./cmd/gateway
test:
	$(GO) test ./...
coverage:
	$(GO) test -coverprofile coverage.out ./...
	$(GO) run ./cmd/coverage-gate -profile coverage.out -minimum $(COVERAGE_MIN) -json artifacts/coverage.json
lint:
	$(GOLANGCI_LINT) run
fmt:
	$(GO) fmt ./...
diagrams:
	$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/render-plantuml.ps1
