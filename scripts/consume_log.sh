#!/bin/bash

# Usage:
#   ./scripts/consume_log.sh [HOST:PORT] [OFFSET] [CA_CERT_PATH] [CERT_PATH] [KEY_PATH]
#
# Examples:
#   ./scripts/consume_log.sh                             # Use default values
#   ./scripts/consume_log.sh localhost:8080 5            # Custom host and offset
#   ./scripts/consume_log.sh localhost:8080 5 /path/to/ca.pem /path/to/cert.pem /path/to/key.pem
#
# Note: Run "make gencert" in the project root to generate TLS certificates before using this script with TLS.

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo "Error: grpcurl is not installed. Please install it first."
    echo "You can install it using: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

# Get the directory of the script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Get the root directory (one level up from scripts)
ROOT_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

# Default values
HOST=${1:-"localhost:8080"}
OFFSET=${2:-0}
CA_CERT_PATH=${3:-"$ROOT_DIR/certs/ca.pem"}
CERT_PATH=${4:-"$ROOT_DIR/certs/root-client.pem"}
KEY_PATH=${5:-"$ROOT_DIR/certs/root-client-key.pem"}

# Construct the JSON payload
DATA="{\"offset\":$OFFSET}"

# Prepare grpcurl options
GRPCURL_OPTS=()

# Add CA certificate, client certificate and key if they exist
if [ -f "$CA_CERT_PATH" ] && [ -f "$CERT_PATH" ] && [ -f "$KEY_PATH" ]; then
    GRPCURL_OPTS+=("-cacert" "$CA_CERT_PATH" "-cert" "$CERT_PATH" "-key" "$KEY_PATH")
else
    echo "Warning: CA certificate, client certificate or key not found. Running without TLS."
    echo "Run 'make gencert' in the project root to generate TLS certificates."
    GRPCURL_OPTS+=("-plaintext")
fi

# Make the grpcurl POST request with verbose output
grpcurl "${GRPCURL_OPTS[@]}" \
    -v \
    -d "$DATA" \
    -import-path "$ROOT_DIR" \
    -proto "$ROOT_DIR/api/v1/log.proto" \
    -H 'Content-Type: application/json' \
    "$HOST" \
    log.v1.Log/Consume 2>&1 | sed 's/^/[DEBUG] /'

echo