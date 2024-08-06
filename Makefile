# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=proglog
BINARY_UNIX=$(BINARY_NAME)_unix

MAIN_PACKAGE=./cmd/server
TEST_DIR=./test
CERT_DIR=./certs
AUTH_DIR=./auth
BUILD_DIR=./bin

.PHONY: all build run clean setup test gencert

all: build

setup:
	mkdir -p $(BUILD_DIR) $(CERT_DIR)
	touch $(CERT_DIR)/.gitkeep

build: setup
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

run: build
	CERT_DIR=$(CERT_DIR) AUTH_DIR=$(AUTH_DIR) $(BUILD_DIR)/$(BINARY_NAME)

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

$(AUTH_DIR)/model.conf: $(TEST_DIR)/model.conf
	cp $< $@

$(AUTH_DIR)/policy.csv: $(TEST_DIR)/policy.csv
	cp $< $@

test: $(AUTH_DIR)/model.conf $(AUTH_DIR)/policy.csv
	CERT_DIR=../../$(CERT_DIR) AUTH_DIR=../../$(AUTH_DIR) $(GOTEST) -v -race -count=1 ./...

gencert:
	cfssl gencert -initca test/ca-csr.json | cfssljson -bare ca
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=server test/server-csr.json | cfssljson -bare server
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=client -cn="root" test/client-csr.json | cfssljson -bare root-client
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=test/ca-config.json -profile=client -cn="nobody" test/client-csr.json | cfssljson -bare nobody-client
	mv *.pem *.csr ${CERT_DIR}

# Build and run in one command
build-and-run: build run

# Cross compilation
build-linux: setup
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_PACKAGE)