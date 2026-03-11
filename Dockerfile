# =============================================================================
# Dockerfile for r8s — CLI Automation Engine for RKE2 & K3S Triage
# Resolves: https://github.com/Rancheroo/r8s/issues/112
#
# Multi-stage build:
#   Stage 1 (builder): Compiles a fully static Go binary
#   Stage 2 (runtime): Minimal alpine image with the r8s binary
#
# Build:
#   docker build -t r8s .
#
# Run:
#   docker run --rm -v $(pwd)/support-bundle:/bundle r8s analyze /bundle
#
# Multi-arch build (requires Docker buildx):
#   docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/rancheroo/r8s:latest .
# =============================================================================

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Install git so `go build` can embed VCS info via ldflags
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Cache dependency downloads separately from the source compile step
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build arguments forwarded from docker build / CI
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build a fully static binary:
#   CGO_ENABLED=0   — no C dependencies
#   -trimpath       — reproducible builds, removes local path prefixes
#   -ldflags        — inject version metadata + strip debug symbols (-s -w)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w \
              -X main.version=${VERSION} \
              -X main.commit=${COMMIT} \
              -X main.date=${DATE}" \
    -o /build/r8s \
    main.go

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates — needed for HTTPS requests (AI/API features)
# timezone data   — consistent timestamps in log output
RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

# Run as a non-root user for security best practices
# uid/gid 65532 mirrors the "nonroot" convention used by distroless images
RUN addgroup -g 65532 nonroot && \
    adduser -u 65532 -G nonroot -s /sbin/nologin -D nonroot

COPY --from=builder /build/r8s /usr/local/bin/r8s

# Ensure the binary is executable
RUN chmod +x /usr/local/bin/r8s

USER nonroot

# Default working directory where users can mount support bundles
WORKDIR /data

ENTRYPOINT ["/usr/local/bin/r8s"]
CMD ["--help"]
