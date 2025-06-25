# Build stage
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o computehive-agent ./cmd/agent

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' computehive

# Create work directory
RUN mkdir -p /var/lib/computehive && \
    chown -R computehive:computehive /var/lib/computehive

# Copy binary from builder
COPY --from=builder /build/computehive-agent /usr/local/bin/

# Switch to non-root user
USER computehive

# Set working directory
WORKDIR /var/lib/computehive

# Expose metrics port
EXPOSE 9090

# Set default environment variables
ENV COMPUTEHIVE_WORK_DIR=/var/lib/computehive/work \
    COMPUTEHIVE_LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# Run the agent
ENTRYPOINT ["computehive-agent"] 