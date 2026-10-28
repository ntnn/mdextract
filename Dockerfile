FROM golang:1.27@sha256:921f61120011c774a7b37fd80bd3c3d259636090d3cd002b6a344322a0328082 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOCACHE=/root/.cache/go-build \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    make build

FROM gcr.io/distroless/static:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99
WORKDIR /
COPY --from=builder /workspace/bin/mdextract .
USER 65532:65532

ENTRYPOINT ["/mdextract"]
