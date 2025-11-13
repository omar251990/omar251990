.PHONY: all build clean install test run help

# Application info
APP_NAME := protei-monitoring
VERSION := 1.0.0
BUILD_DIR := build
INSTALL_DIR := /usr/protei/Protei_Monitoring

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build parameters
BINARY_NAME := $(APP_NAME)
BINARY_UNIX := $(BINARY_NAME)
MAIN_PATH := ./cmd/$(APP_NAME)

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

all: clean deps build

help:
	@echo "Protei_Monitoring Build System"
	@echo "==============================="
	@echo "Available targets:"
	@echo "  make build      - Build the application"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make deps       - Download dependencies"
	@echo "  make test       - Run tests"
	@echo "  make install    - Install to $(INSTALL_DIR)"
	@echo "  make run        - Build and run locally"
	@echo "  make all        - Clean, deps, and build"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_UNIX)"

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

install: build
	@echo "Installing to $(INSTALL_DIR)..."
	@sudo mkdir -p $(INSTALL_DIR)/{bin,configs,logs,out,ingest,scripts,dictionaries}
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/bin/
	@sudo cp configs/*.yaml $(INSTALL_DIR)/configs/
	@sudo cp scripts/*.sh $(INSTALL_DIR)/scripts/
	@sudo chmod +x $(INSTALL_DIR)/scripts/*.sh
	@sudo chmod +x $(INSTALL_DIR)/bin/$(BINARY_NAME)
	@echo "Installation complete!"
	@echo ""
	@echo "To start: sudo $(INSTALL_DIR)/scripts/start.sh"
	@echo "To stop:  sudo $(INSTALL_DIR)/scripts/stop.sh"
	@echo "Status:   sudo $(INSTALL_DIR)/scripts/status.sh"

run: build
	@echo "Running $(APP_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME) -config=configs/config.yaml

docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/configs:/app/configs $(APP_NAME):$(VERSION)

# Development helpers
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .

vet:
	@echo "Running go vet..."
	@go vet ./...

lint:
	@echo "Running linters..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi
