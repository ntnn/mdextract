GO ?= go
CONTAINER_TOOL ?= docker

.PHONY: check
check: fmt lint test

.PHONY: build
build:
	@mkdir -p bin
	$(GO) build -o ./bin/mdextract .

.PHONY: docker-build
docker-build:
	$(CONTAINER_TOOL) build -t mdextract:dev .

.PHONY: fmt
fmt:
	$(GO) fmt ./...

GOLANGCI_LINT = $(UGET_DIRECTORY)/golangci-lint-$(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: install-golangci-lint
	$(GOLANGCI_LINT) run ./...

.PHONY: lint-fix
lint-fix: install-golangci-lint
	$(GOLANGCI_LINT) run --fix ./...

.PHONY: test
test:
	$(GO) test -cover -race ./...

## tools
export UGET_DIRECTORY ?= hack/tools
export UGET_CHECKSUMS ?= hack/tools.checksums
export UGET_VERSIONED_BINARIES = true
GOLANGCI_LINT_VERSION ?= 2.9.0

.PHONY: install-golangci-lint
install-golangci-lint:
	@hack/uget.sh https://github.com/golangci/golangci-lint/releases/download/v{VERSION}/golangci-lint-{VERSION}-{GOOS}-{GOARCH}.tar.gz golangci-lint $(GOLANGCI_LINT_VERSION)
