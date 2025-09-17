## secure_packager
![secure_packager logo](https://raw.githubusercontent.com/stevef1uk/secure_packager/main/images/secure_packager.png)

### Why this exists

Modern teams need to ship valuable data/models to customers securely, without painful key management or custom builds on every machine. Plain zips leak keys, signatures don’t enforce access, and full-blown DRM is heavy and brittle. This project offers a pragmatic middle ground:

- **Problem**: You must distribute files inside containers or via zip, but only the intended recipient should be able to open them.
- **Solution**: Encrypt files with a symmetric key (Fernet), then wrap that key with the recipient’s RSA public key. Only their private key can unwrap and decrypt.
- **Optional licensing**: Add a vendor-signed token that’s verified at decrypt time for friendly messaging and basic enforcement (expiry, warnings, block within 24h).

### Key features

- **Confidentiality by default**: No plaintext Fernet key shipped.
- **Two modes**: with or without licensing enforcement (auto-detected from the zip).
- **Simple CLI or Docker**: Use locally or via container with volume mounts.
- **Portable**: Multi-arch container images (linux/amd64, linux/arm64).

Envelope encryption utilities for distributing data/models:
- Packager: Fernet-encrypts files; wraps the Fernet key with customer's RSA public key (RSA-OAEP SHA-256)
- Unpack: Requires customer's RSA private key to unwrap key and decrypt files
- Issue-token: Generates vendor-signed license tokens for messaging/enforcement

### Architecture

![secure_packager architecture](https://raw.githubusercontent.com/stevef1uk/secure_packager/main/images/architecture.png)

### Build

```
cd secure_packager
go build ./cmd/packager
go build ./cmd/unpack
go build ./cmd/issue-token
```

### Docker (multi-arch)

Build multi-arch image (requires buildx):
```
cd secure_packager
docker buildx build --platform linux/amd64,linux/arm64 -t yourorg/secure-packager:latest --push .
```

Run examples (volume mount input/output):
```
# Packager (no licensing):
docker run --rm -v $(pwd)/input:/in -v $(pwd)/out:/out \
  yourorg/secure-packager:latest packager -in /in -out /out -pub /out/customer_public.pem -zip=true

# Packager (with licensing):
docker run --rm -v $(pwd)/input:/in -v $(pwd)/out:/out -v $(pwd)/keys:/keys \
  yourorg/secure-packager:latest packager -in /in -out /out -pub /out/customer_public.pem -zip=true -license -vendor-pub /keys/vendor_public.pem

# Unpack (auto-detect licensing from zip):
docker run --rm -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  yourorg/secure-packager:latest unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec -license-token /keys/token.txt

# Issue token:
docker run --rm -v $(pwd)/keys:/keys yourorg/secure-packager:latest \
  issue-token -priv /keys/vendor_private.pem -expiry 2025-12-31 -company Acme -email ops@acme.com -out /keys/token.txt
```

### Docker quickstart (clean tmp workspace)

If you prefer to set everything up under a clean `tmp` directory and keep your project tree untouched:

1) Create a temporary workspace and directories
```
mkdir -p tmp && cd tmp
mkdir -p in out dec keys env
```

2) Create required keys
- Customer keypair (required in all modes; public key used by packager, private key used by unpacker)
```
openssl genrsa -out keys/customer_private.pem 2048
openssl rsa -in keys/customer_private.pem -pubout -out out/customer_public.pem
```
- Vendor keypair (only needed for licensing mode; signs license tokens)
```
openssl genrsa -out keys/vendor_private.pem 2048
openssl rsa -in keys/vendor_private.pem -pubout -out keys/vendor_public.pem
```

3) Add files to encrypt
```
echo "demo secret" > in/demo.txt
```

4) Package without licensing
```
docker run --rm \
  -v $(pwd)/in:/in -v $(pwd)/out:/out \
  stevef1uk/secure-packager:latest \
  packager -in /in -out /out -pub /out/customer_public.pem -zip=true
```

5) Package with licensing (optional)
```
docker run --rm \
  -v $(pwd)/in:/in -v $(pwd)/out:/out -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  packager -in /in -out /out -pub /out/customer_public.pem -zip=true -license -vendor-pub /keys/vendor_public.pem
```

6) Issue a token (licensing mode only)
```
docker run --rm -v $(pwd)/keys:/keys stevef1uk/secure-packager:latest \
  issue-token -priv /keys/vendor_private.pem -expiry 2025-12-31 -company Acme -email ops@acme.com -out /keys/token.txt
```

7) Unpack (auto-detects licensing from zip)
```
# no licensing
docker run --rm \
  -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec

# with licensing
docker run --rm \
  -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec -license-token /keys/token.txt
```

Tip: Simulate expiry by setting FAKE_NOW using an env file:
```
echo FAKE_NOW=2100-01-01 > env/.env
docker run --rm --env-file $(pwd)/env/.env \
  -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec -license-token /keys/token.txt
```

### Quick demo (end-to-end)

Script: `secure_packager/examples/quick_demo.sh`

What it does:
- Generates vendor and customer RSA keys (OpenSSL)
- Builds tools
- Packages with and without licensing
- Issues a long-lived vendor-signed token
- Unpacks both zips; license flow is auto-enforced for the licensed one

Run:
```
./examples/quick_demo.sh
```

Outputs (under `secure_packager/_demo`):
- `out_no_license/encrypted_files.zip` and decrypted `decrypted_no_license`
- `out_with_license/encrypted_files.zip` and decrypted `decrypted_with_license`
- `token.txt`, `vendor_public.pem`, `customer_public.pem`

### Modes

- Without licensing: default; zip contains encrypted files and `wrapped_key.bin` only
- With licensing: add manifest and vendor public key; unzip enforces license automatically

### Package (no licensing)

```
./packager -in ./input_dir -out ./out_dir -pub ./customer_public.pem -zip=true
```

### Package (license required)

```
./packager -in ./input_dir -out ./out_dir -pub ./customer_public.pem -zip=true \
  -license -vendor-pub ./vendor_public.pem
```

Outputs add:
- `manifest.json` with `{ "license_required": true, "vendor_public_key": "vendor_public.pem" }`
- `vendor_public.pem`

### Unpack (auto-detects licensing from zip)

```
# no licensing required in zip
./unpack -zip ./out_dir/encrypted_files.zip -priv ./customer_private.pem -out ./decrypted

# licensing required in zip (manifest present)
./unpack -zip ./out_dir/encrypted_files.zip -priv ./customer_private.pem -out ./decrypted \
  -license-token ./token.txt                # required
# -vendor-pub optional; defaults to vendor_public.pem inside zip when present
```

License token format (compatible with existing):
- `base64url(expiry:company:email:placeholder_key:signature_b64)`
- Signature: RSA-PSS over `expiry:company:email:placeholder_key`
- Behavior: prints Company/Email/Expiry, warns at <=7 days, blocks if expired or <=24h (supports `FAKE_NOW`)

### Issue a license token

```
./issue-token -priv ./vendor_private.pem -expiry 2025-12-31 -company "Acme" -email "ops@acme.com" -out ./token.txt
```

### Manual steps for first-time test

1) Generate keys (OpenSSL):
```
openssl genrsa -out vendor_private.pem 2048
openssl rsa -in vendor_private.pem -pubout -out vendor_public.pem
openssl genrsa -out customer_private.pem 2048
openssl rsa -in customer_private.pem -pubout -out customer_public.pem
```

2) Build tools:
```
cd secure_packager
go build ./cmd/packager && go build ./cmd/unpack && go build ./cmd/issue-token
```

3) Package without licensing:
```
./packager -in ./input_dir -out ./out_no_license -pub ./customer_public.pem -zip=true
./unpack -zip ./out_no_license/encrypted_files.zip -priv ./customer_private.pem -out ./dec_no_lic
```

4) Package with licensing and issue token:
```
./packager -in ./input_dir -out ./out_with_license -pub ./customer_public.pem -zip=true -license -vendor-pub ./vendor_public.pem
./issue-token -priv ./vendor_private.pem -expiry 2099-12-31 -company "Demo Co" -email "demo@example.com" -out ./token.txt
./unpack -zip ./out_with_license/encrypted_files.zip -priv ./customer_private.pem -out ./dec_with_lic -license-token ./token.txt
```

### Notes
- RSA key size >= 2048 recommended
- Only the private key holder can unwrap the Fernet key
- No Fernet key in plaintext is shipped


