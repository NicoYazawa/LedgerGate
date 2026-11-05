GO ?= go
GOLANGCI_LINT ?= golangci-lint
POWERSHELL ?= powershell
COVERAGE_MIN ?= 90
ARTIFACT_STAGE ?= shared-value-primitives
EXEEXT := $(if $(filter windows,$(shell $(GO) env GOOS)),.exe,)

.PHONY: build test test-integration coverage lint fmt diagrams verify-artifacts
build:
	mkdir -p bin
	$(GO) build -o bin/gateway$(EXEEXT) ./cmd/gateway
test:
	$(GO) test ./...
test-integration:
	$(GO) test -tags=integration ./internal/testkit -count=1
coverage:
	$(GO) test -coverprofile coverage.out ./...
	$(GO) run ./cmd/coverage-gate -profile coverage.out -minimum $(COVERAGE_MIN) -json artifacts/coverage.json
lint:
	$(GOLANGCI_LINT) run
fmt:
	$(GO) fmt ./...
diagrams:
	$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/render-plantuml.ps1
verify-artifacts:
	$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/verify-stage-artifacts.ps1 -Stage $(ARTIFACT_STAGE)
