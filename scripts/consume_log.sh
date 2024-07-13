
#!/bin/bash

# Define the URL
URL=${1:-"http://localhost:8080"}

# Define the offset
OFFSET=${2:-0}

# Construct the JSON payload
DATA="{\"offset\": $OFFSET}"

# Make the curl GET request
# Note: We use -G to append data to the URL for a GET request
curl -X GET "$URL" -G --data-urlencode "$(echo $DATA)"

# Add a newline for better formatting of output
echo