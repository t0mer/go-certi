# syntax=docker/dockerfile:1

# Stage 1: Build Go binary
# Uses golang:1.26-alpine to match the `go 1.26.0` directive in go.mod
# (auto-bumped when sqlc was added as a tool dependency)
FROM golang:1.26-alpine AS go-builder
ARG VERSION=dev
WORKDIR /src

# git is needed by `go mod download` for VCS-tagged dependencies
RUN apk add --no-cache git

# Prevent toolchain auto-download under QEMU multi-arch builds, where
# downloading a different Go toolchain is flaky. The base image already
# provides a Go version that satisfies go.mod.
ENV GOTOOLCHAIN=local

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/t0mer/go-certi/internal/version.Version=${VERSION}" \
    -o /go-certi \
    ./cmd/go-certi

# Stage 2: Build React frontend (only runs when web/package.json exists)
FROM node:20-alpine AS node-builder
WORKDIR /web
# If web/package.json doesn't exist this stage produces nothing useful;
# the Go stage already embedded web/dist from the COPY above.
COPY web/package*.json ./
RUN if [ -f package.json ]; then npm ci --silent; fi
COPY web/ ./
RUN if [ -f package.json ]; then npm run build; fi

# Stage 3: Final minimal image
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-builder /go-certi /go-certi
EXPOSE 8111
VOLUME ["/data"]
ENV GO_CERTI_CONF=/data
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/go-certi", "--version"]
ENTRYPOINT ["/go-certi"]
CMD ["--conf", "/data"]
