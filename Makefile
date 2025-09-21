# Variables
BINARY_NAME=random-pod-killer
MAIN_PATH=./cmd/killer
BUILD_DIR=./bin

.PHONY: build run clean

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	@go run $(MAIN_PATH)

# Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
