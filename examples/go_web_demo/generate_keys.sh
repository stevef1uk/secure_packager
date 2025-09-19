#!/bin/bash
set -e

# Get current user's UID and GID
USER_ID=$(id -u)
GROUP_ID=$(id -g)

# Generate keys if they don't exist
if [ ! -f "keys/customer_private.pem" ] || [ ! -f "keys/customer_public.pem" ]; then
    echo "Generating customer keys..."
    docker run --rm \
        --user "$USER_ID:$GROUP_ID" \
        -v "$(pwd)/keys:/output" \
        secure-packager-keygen:latest customer 2048
fi

if [ ! -f "keys/vendor_private.pem" ] || [ ! -f "keys/vendor_public.pem" ]; then
    echo "Generating vendor keys..."
    docker run --rm \
        --user "$USER_ID:$GROUP_ID" \
        -v "$(pwd)/keys:/output" \
        secure-packager-keygen:latest vendor 2048
fi

echo "Keys generated successfully!"
