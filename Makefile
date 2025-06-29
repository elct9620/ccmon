BINARY_NAME=ccmon

.PHONY: build clean generate proto test fmt vet run-server run-monitor

# Default target
all: build

# Generate protobuf code
generate:
	go generate ./...

# Alternative target for generating protobuf code directly
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/query.proto

# Build the application
build: generate
	go build -o $(BINARY_NAME) .

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