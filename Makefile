GO ?= go

.PHONY: check
check: fmt lint test

.PHONY: build
build:
	@mkdir -p bin
	$(GO) build -o ./bin/mdextract ./cmd/mdextract

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: lint
lint:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test -v ./...
