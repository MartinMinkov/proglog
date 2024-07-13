#!/bin/bash

# Define the URL
BASE_URL=${1:-"http://localhost:8080"}

# Define the offset
OFFSET=${2:-0}

# Construct the full URL with the offset in the path
URL="${BASE_URL}/${OFFSET}"

# Make the curl GET request with verbose output
curl -X GET "$URL" \
     -H "Content-Type: application/json" \
     --max-time 5 \
     -v \
     2>&1 | sed 's/^/[DEBUG] /'

# Add a newline for better formatting of output
echo