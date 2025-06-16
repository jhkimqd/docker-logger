.PHONY: build clean run test

# Binary name
BINARY_NAME=docker-logger

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Main build target
build:
	$(GOBUILD) -o $(BINARY_NAME) cmd/logger/main.go

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run the application (requires network parameter)
run:
	@if [ -z "$(network)" ]; then \
		echo "Usage: make run network=<network-name>"; \
		exit 1; \
	fi
	$(GOCMD) run cmd/logger/main.go --network $(network)

# Run tests
test:
	$(GOTEST) -v ./...

# Install dependencies
deps:
	$(GOGET) -v -t -d ./...
	$(GOCMD) mod tidy

# Build for multiple platforms
build-all: clean
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux-amd64 cmd/logger/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-darwin-amd64 cmd/logger/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-windows-amd64.exe cmd/logger/main.go