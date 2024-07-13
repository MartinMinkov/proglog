# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
BINARY_NAME=proglog
BINARY_UNIX=$(BINARY_NAME)_unix

# Main package path
MAIN_PACKAGE=./cmd/server

# Build directory
BUILD_DIR=./bin

.PHONY: all build run clean setup

all: build

setup:
	mkdir -p $(BUILD_DIR)

build: setup
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

run: build
	$(BUILD_DIR)/$(BINARY_NAME)

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Build and run in one command
build-and-run: build run

# Cross compilation
build-linux: setup
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_PACKAGE)