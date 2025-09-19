#!/bin/bash
set -e

# Get current user's UID and GID
USER_ID=$(id -u)
GROUP_ID=$(id -g)

# Generate keys if they don't exist
if [ ! -f "keys/customer_private.pem" ] || [ ! -f "keys/customer_public.pem" ]; then
    echo "Generating customer keys..."
    docker run --rm -v "$(pwd)/keys:/output" secure-packager-keygen:latest customer 2048
    
    # Fix ownership to current user
    echo "Fixing ownership of customer keys..."
    sudo chown -R $USER_ID:$GROUP_ID keys/customer_private.pem keys/customer_public.pem 2>/dev/null || {
        echo "Warning: Could not change ownership of customer keys. You may need to run:"
        echo "  sudo chown -R $USER_ID:$GROUP_ID keys/"
    }
fi

if [ ! -f "keys/vendor_private.pem" ] || [ ! -f "keys/vendor_public.pem" ]; then
    echo "Generating vendor keys..."
    docker run --rm -v "$(pwd)/keys:/output" secure-packager-keygen:latest vendor 2048
    
    # Fix ownership to current user
    echo "Fixing ownership of vendor keys..."
    sudo chown -R $USER_ID:$GROUP_ID keys/vendor_private.pem keys/vendor_public.pem 2>/dev/null || {
        echo "Warning: Could not change ownership of vendor keys. You may need to run:"
        echo "  sudo chown -R $USER_ID:$GROUP_ID keys/"
    }
fi

echo "Keys generated successfully!"
