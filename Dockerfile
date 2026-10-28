FROM golang:1.27@sha256:b76020929b709ddb6202afc05410a0faefae20ae05e27ad548b3870092688a69 AS builder

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
