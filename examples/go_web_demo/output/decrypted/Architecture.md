```mermaid
graph LR
  subgraph Producer
    A[Input files] --> B[Fernet encrypt -> *.enc]
    K[Generate Fernet key] --> B
    KP[Customer Public Key] --> W[Wrap key RSA-OAEP]
    B --> Z[encrypted_files.zip]
    W --> Z
    subgraph Licensing
      VP[Vendor Public Key] --> M[manifest + vendor_public.pem]
    end
    M --> Z
  end

  Z --> C[Deliver ZIP]

  subgraph Consumer
    C --> U[Unpack Tool]
    PR[Customer Private Key] --> U
    T[License Token optional] --> U
    U --> KF[Unwrap Fernet key]
    U --> L{License OK?}
    L -->|yes or not required| D[Decrypt to output]
    L -->|no| X[Blocked]
  end
```

Key flows:
- Files are encrypted symmetrically (Fernet), and the Fernet key is wrapped with the customer’s RSA public key (RSA-OAEP SHA-256).
- The zip contains only ciphertext (*.enc), the wrapped key, and optionally licensing manifest/vendor public key.
- At decrypt time, the customer’s private key unwraps the Fernet key; licensing (if present) is verified before decryption proceeds.



