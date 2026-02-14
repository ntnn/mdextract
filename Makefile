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

.PHONY: lint
lint:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test -cover -race ./...
