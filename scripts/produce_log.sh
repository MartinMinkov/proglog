#!/bin/bash

# Default values
URL=${1:-"http://localhost:8080"}
VALUE=${2:-"TGV0J3MgR28gIzEK"}

# Construct the JSON payload
DATA="{\"record\": {\"value\": \"$VALUE\"}}"

# Make the curl POST request with verbose output
curl -X POST $URL -d "$DATA" \
     -H "Content-Type: application/json" \
     -v \
     2>&1 | sed 's/^/[DEBUG] /'

# Add a newline for better formatting of output
echo