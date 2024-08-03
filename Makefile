# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=proglog
BINARY_UNIX=$(BINARY_NAME)_unix

MAIN_PACKAGE=./cmd/server
CERT_DIR=./certs
BUILD_DIR=./bin

.PHONY: all build run clean setup test gencert

all: build

setup:
	mkdir -p $(BUILD_DIR) $(CERT_DIR)

build: setup
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

run: build
	CONFIG_DIR=$(CERT_DIR) $(BUILD_DIR)/$(BINARY_NAME)

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(CERT_DIR)

compile:
	protoc api/v1/*.proto \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

test:
	CONFIG_DIR=../../$(CERT_DIR) $(GOTEST) -v ./...

gencert:
	cfssl gencert -initca test/ca-csr.json | cfssljson -bare ca
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=server test/server-csr.json | cfssljson -bare server
	mv *.pem *.csr ${CERT_DIR}

# Build and run in one command
build-and-run: build run

# Cross compilation
build-linux: setup
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_PACKAGE)