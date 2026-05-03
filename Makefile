GO ?= go
CONTAINER_TOOL ?= docker

TOOLS_DIR := hack/tools
GOLANGCI_LINT_VER := 2.12.1
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint-$(GOLANGCI_LINT_VER)

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

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_FLAGS) ./...

.PHONY: lint-fix
lint-fix: override GOLANGCI_LINT_FLAGS := $(GOLANGCI_LINT_FLAGS) --fix
lint-fix: lint

.PHONY: test
test:
	$(GO) test -cover -race ./...

## tools
$(GOLANGCI_LINT):
	mkdir -p $(TOOLS_DIR)
	$(GO) tool github.com/ntnn/mindl download -tool golangci-lint -common -out $@ -version $(GOLANGCI_LINT_VER)
