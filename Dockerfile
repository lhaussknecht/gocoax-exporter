# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o gocoax-exporter .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 gocoax && \
    adduser -D -u 1000 -G gocoax gocoax

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gocoax-exporter .

# Create config directory
RUN mkdir -p /etc/gocoax-exporter && \
    chown -R gocoax:gocoax /app /etc/gocoax-exporter

# Switch to non-root user
USER gocoax

# Expose port
EXPOSE 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# Set default config path
ENV CONFIG_PATH=/etc/gocoax-exporter/config.yaml

# Run the exporter
ENTRYPOINT ["/app/gocoax-exporter"]
CMD ["-config", "/etc/gocoax-exporter/config.yaml"]
