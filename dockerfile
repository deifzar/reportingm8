# Build
FROM golang:1.25.0-alpine3.22 AS builder
# Install only required build dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata \
    && adduser -D -g '' appuser
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (use .dockerignore to exclude sensitive files)
COPY . .

# Build with security flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o reportingm8 .

# Release
FROM alpine:3.20

# Security updates and minimal runtime dependencies
RUN apk --no-cache add ca-certificates tzdata \
    && apk --no-cache upgrade \
    && rm -rf /var/cache/apk/* \
    && adduser -D -g '' -s /bin/sh appuser

COPY --from=builder --chown=appuser:appuser /app/reportingm8 /usr/local/bin/

# Security: Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD reportingm8 --help || exit 1

# Metadata
LABEL maintainer="i@deifzar.me" \
    version="1.0" \
    description="ReportinM8 - Hardened Report Builder (Runtime)" \
    security.scan="required-non-root-privileges"

# Expose port (document the port used)
# EXPOSE 8000

# Use exec form for better signal handling
CMD ["reportingm8", "help"]