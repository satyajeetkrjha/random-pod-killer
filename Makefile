# Variables
BINARY_NAME=random-pod-killer
MAIN_PATH=./cmd/killer
BUILD_DIR=./bin
DOCKER_IMAGE=random-pod-killer
VERSION?=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-s -w"
CGO_ENABLED=0

.PHONY: all build clean test deps fmt vet lint run help docker-build docker-run

# Default target
all: clean deps fmt vet test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run golint (requires golint to be installed)
lint:
	@echo "Running golint..."
	@which golint > /dev/null || (echo "golint not installed, run: go install golang.org/x/lint/golint@latest" && exit 1)
	golint ./...

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# Run with custom namespace
run-namespace:
	@echo "Running $(BINARY_NAME) with custom namespace..."
	$(GOCMD) run $(MAIN_PATH) -namespace=kube-system

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it $(DOCKER_IMAGE):$(VERSION)

# Create a simple Dockerfile if it doesn't exist
dockerfile:
	@if [ ! -f Dockerfile ]; then \
		echo "Creating Dockerfile..."; \
		echo "FROM golang:1.24-alpine AS builder" > Dockerfile; \
		echo "WORKDIR /app" >> Dockerfile; \
		echo "COPY go.mod go.sum ./" >> Dockerfile; \
		echo "RUN go mod download" >> Dockerfile; \
		echo "COPY . ." >> Dockerfile; \
		echo "RUN CGO_ENABLED=0 GOOS=linux go build -ldflags \"-s -w\" -o $(BINARY_NAME) $(MAIN_PATH)" >> Dockerfile; \
		echo "" >> Dockerfile; \
		echo "FROM alpine:latest" >> Dockerfile; \
		echo "RUN apk --no-cache add ca-certificates" >> Dockerfile; \
		echo "WORKDIR /root/" >> Dockerfile; \
		echo "COPY --from=builder /app/$(BINARY_NAME) ." >> Dockerfile; \
		echo "CMD [\"./$(BINARY_NAME)\"]" >> Dockerfile; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Clean, deps, fmt, vet, test, and build"
	@echo "  build         - Build the binary"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-darwin  - Build for macOS"
	@echo "  build-windows - Build for Windows"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run golint"
	@echo "  run           - Run the application"
	@echo "  run-namespace - Run with kube-system namespace"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  dockerfile    - Create a Dockerfile"
	@echo "  help          - Show this help message"
