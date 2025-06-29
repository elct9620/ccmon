# Multi-stage build for production
FROM golang:1.24.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ccmon .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S ccmon && \
    adduser -u 1001 -S ccmon -G ccmon

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ccmon .

# Create data directory and set ownership
RUN mkdir -p /data && \
    chown -R ccmon:ccmon /app /data

# Create volume for database persistence
VOLUME ["/data"]

# Switch to non-root user
USER ccmon

# Expose gRPC port for OTLP receiver and query service
EXPOSE 4317

# Run as server by default with database in volume
ENTRYPOINT ["./ccmon"]
CMD ["--server", "--database-path", "/data/ccmon.db", "--server-address", "0.0.0.0:4317"]