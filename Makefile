BINARY_NAME=ccmon

.PHONY: build clean generate proto test fmt vet run-server run-monitor

# Default target
all: build

# Generate protobuf code
generate: check-protoc
	go generate ./...

# Alternative target for generating protobuf code directly
proto: check-protoc
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/query.proto

# Build the application with version info
build: generate
	@VERSION=$$(git describe --tags --exact-match 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	DATE=$$(date -u '+%Y-%m-%d_%H:%M:%S'); \
	go build -ldflags "-X main.version=$$VERSION -X main.commit=$$COMMIT -X main.date=$$DATE" -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f proto/*.pb.go

# Format code
fmt:
	gofmt -w .

# Vet code
vet:
	go vet ./...

# Run tests
test:
	go test ./...

# Run server mode
run-server: build
	./$(BINARY_NAME) -s

# Run monitor mode
run-monitor: build
	./$(BINARY_NAME)

# Development: rebuild and run server
dev-server: clean build
	./$(BINARY_NAME) -s

# Development: rebuild and run monitor
dev-monitor: clean build
	./$(BINARY_NAME)

# Install dependencies (if needed)
deps:
	go mod tidy
	go mod download

# Check protoc and Go plugin versions for consistency
check-protoc:
	@echo "Checking Protocol Buffers toolchain..."
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "Error: protoc is not installed. Please install Protocol Buffers compiler v30.2+"; \
		echo "Visit: https://github.com/protocolbuffers/protobuf/releases"; \
		exit 1; \
	fi
	@PROTOC_VERSION=$$(protoc --version | sed 's/libprotoc //'); \
	MAJOR_VERSION=$$(echo $$PROTOC_VERSION | cut -d. -f1); \
	if [ "$$MAJOR_VERSION" -lt 30 ] 2>/dev/null; then \
		echo "Warning: protoc version $$PROTOC_VERSION detected. Recommended: v30.2+"; \
		echo "This may cause version inconsistencies in generated files."; \
		echo "Consider upgrading: https://github.com/protocolbuffers/protobuf/releases"; \
	else \
		echo "✓ protoc version $$PROTOC_VERSION is compatible"; \
	fi
	@if ! command -v protoc-gen-go >/dev/null 2>&1; then \
		echo "Error: protoc-gen-go is not installed. Install with:"; \
		echo "go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1"; \
		exit 1; \
	fi
	@PROTOC_GEN_GO_VERSION=$$(protoc-gen-go --version | sed 's/protoc-gen-go v//'); \
	echo "✓ protoc-gen-go version $$PROTOC_GEN_GO_VERSION detected"; \
	if [ "$$PROTOC_GEN_GO_VERSION" != "1.28.1" ]; then \
		echo "Warning: protoc-gen-go version $$PROTOC_GEN_GO_VERSION differs from pinned v1.28.1"; \
		echo "For consistency, install: go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1"; \
	fi
	@if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then \
		echo "Error: protoc-gen-go-grpc is not installed. Install with:"; \
		echo "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0"; \
		exit 1; \
	fi
	@PROTOC_GEN_GRPC_VERSION=$$(protoc-gen-go-grpc --version | sed 's/protoc-gen-go-grpc //'); \
	echo "✓ protoc-gen-go-grpc version $$PROTOC_GEN_GRPC_VERSION detected"; \
	if [ "$$PROTOC_GEN_GRPC_VERSION" != "1.2.0" ]; then \
		echo "Warning: protoc-gen-go-grpc version $$PROTOC_GEN_GRPC_VERSION differs from pinned v1.2.0"; \
		echo "For consistency, install: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0"; \
	fi