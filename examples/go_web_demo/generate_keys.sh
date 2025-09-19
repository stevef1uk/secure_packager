#!/bin/bash
set -e

# Generate keys if they don't exist
if [ ! -f "keys/customer_private.pem" ] || [ ! -f "keys/customer_public.pem" ]; then
    echo "Generating customer keys..."
    docker run --rm -v "$(pwd)/keys:/output" secure-packager-keygen:latest customer 2048
fi

if [ ! -f "keys/vendor_private.pem" ] || [ ! -f "keys/vendor_public.pem" ]; then
    echo "Generating vendor keys..."
    docker run --rm -v "$(pwd)/keys:/output" secure-packager-keygen:latest vendor 2048
fi

echo "Keys generated successfully!"
