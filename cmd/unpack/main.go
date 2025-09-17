package main

import (
	"archive/zip"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fernet/fernet-go"
)

func readRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := keyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("PEM is not RSA private key")
	}
	return k, nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return err
		}
		rc.Close()
		outFile.Close()
	}
	return nil
}

func unwrapFernetKey(priv *rsa.PrivateKey, wrapped []byte) (*fernet.Key, error) {
	label := []byte("secure_packager")
	raw, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, wrapped, label)
	if err != nil {
		return nil, err
	}
	// raw holds the base64-url encoded fernet key string
	keys := fernet.MustDecodeKeys(string(raw))
	if len(keys) == 0 {
		return nil, fmt.Errorf("failed to decode fernet key")
	}
	return keys[0], nil
}

func decryptDirWithFernet(k *fernet.Key, srcDir, destDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".enc") {
			continue
		}
		inPath := filepath.Join(srcDir, e.Name())
		outPath := filepath.Join(destDir, strings.TrimSuffix(e.Name(), ".enc"))
		data, err := os.ReadFile(inPath)
		if err != nil {
			return err
		}
		pt := fernet.VerifyAndDecrypt(data, 0, []*fernet.Key{k})
		if pt == nil {
			return fmt.Errorf("failed to decrypt %s", e.Name())
		}
		if err := os.WriteFile(outPath, pt, 0644); err != nil {
			return err
		}
		fmt.Printf("Decrypted %s -> %s\n", e.Name(), filepath.Base(outPath))
	}
	return nil
}

func main() {
	zipPath := flag.String("zip", "", "Path to encrypted zip produced by packager")
	workDir := flag.String("work", "./_unpack", "Working directory to extract zip")
	outDir := flag.String("out", "./decrypted", "Output directory for decrypted files")
	privPath := flag.String("priv", "", "Path to RSA private key (PEM) to unwrap key")
	licenseToken := flag.String("license-token", "", "Optional path to vendor license token (no key) for messaging/enforcement; if omitted and zip contains manifest.json with license_required, unpack requires this flag")
	vendorPub := flag.String("vendor-pub", "", "Optional path to vendor RSA public key (PEM) to verify license token; if omitted, unpacker looks for vendor_public.pem in the zip")
	flag.Parse()

	if *zipPath == "" || *privPath == "" {
		fmt.Println("Usage: unpack -zip <encrypted_files.zip> -priv <private.pem> [-work ./_unpack] [-out ./decrypted]")
		os.Exit(1)
	}

	if err := os.MkdirAll(*workDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create work dir: %v\n", err)
		os.Exit(1)
	}
	if err := unzip(*zipPath, *workDir); err != nil {
		fmt.Fprintf(os.Stderr, "Unzip failed: %v\n", err)
		os.Exit(1)
	}

	// Detect manifest.json to determine if license enforcement is required
	requireLicense := false
	vendorPubPath := *vendorPub
	manifestPath := filepath.Join(*workDir, "manifest.json")
	if b, err := os.ReadFile(manifestPath); err == nil {
		// naive detection of flag and embedded public key name
		s := string(b)
		if strings.Contains(s, "\"license_required\": true") {
			requireLicense = true
		}
		if vendorPubPath == "" && strings.Contains(s, "vendor_public.pem") {
			vendorPubPath = filepath.Join(*workDir, "vendor_public.pem")
		}
	}

	// License verification & messaging if required or requested
	if requireLicense || *licenseToken != "" || vendorPubPath != "" {
		if *licenseToken == "" {
			fmt.Fprintln(os.Stderr, "license required: provide -license-token <path> (as per manifest)")
			os.Exit(1)
		}
		if vendorPubPath == "" {
			fmt.Fprintln(os.Stderr, "license required: vendor public key not found; provide -vendor-pub <path> or include vendor_public.pem in zip")
			os.Exit(1)
		}
		if err := verifyAndEnforceLicense(vendorPubPath, *licenseToken); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	wrappedPath := filepath.Join(*workDir, "wrapped_key.bin")
	wrapped, err := os.ReadFile(wrappedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Reading wrapped_key.bin failed: %v\n", err)
		os.Exit(1)
	}

	priv, err := readRSAPrivateKey(*privPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Reading private key failed: %v\n", err)
		os.Exit(1)
	}

	k, err := unwrapFernetKey(priv, wrapped)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unwrap failed: %v\n", err)
		os.Exit(1)
	}

	if err := decryptDirWithFernet(k, *workDir, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Decrypt failed: %v\n", err)
		os.Exit(1)
	}
}

// verifyAndEnforceLicense verifies the vendor token signature, prints license info,
// warns on nearing expiry, and blocks if expired or within 24 hours of expiry.
// The token format matches the existing system but WITHOUT the Fernet key in use here:
// base64url( expiry:company:email:placeholder_key:signature_b64 )
func verifyAndEnforceLicense(vendorPubPath, tokenPath string) error {
	pubBytes, err := os.ReadFile(vendorPubPath)
	if err != nil {
		return fmt.Errorf("error reading vendor public key: %w", err)
	}
	block, _ := pem.Decode(pubBytes)
	if block == nil {
		return fmt.Errorf("invalid vendor public key PEM")
	}
	var parsed any
	if k, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		parsed = k
	} else if k2, err2 := x509.ParsePKCS1PublicKey(block.Bytes); err2 == nil {
		parsed = k2
	} else {
		return fmt.Errorf("error parsing vendor public key: %v", err)
	}
	pub, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("vendor public key is not RSA")
	}

	tokenB64, err := os.ReadFile(tokenPath)
	if err != nil {
		return fmt.Errorf("error reading license token: %w", err)
	}
	decoded, err := base64.URLEncoding.DecodeString(strings.TrimSpace(string(tokenB64)))
	if err != nil {
		return fmt.Errorf("invalid token b64: %w", err)
	}
	parts := strings.SplitN(string(decoded), ":", 5)
	if len(parts) != 5 {
		return fmt.Errorf("invalid token format")
	}
	expiryStr, company, email, kB64, sigB64 := parts[0], parts[1], parts[2], parts[3], parts[4]
	sig, err := base64.URLEncoding.DecodeString(sigB64)
	if err != nil {
		return fmt.Errorf("invalid signature b64: %w", err)
	}
	payload := []byte(expiryStr + ":" + company + ":" + email + ":" + kB64)
	hashed := sha256.Sum256(payload)
	if err := rsa.VerifyPSS(pub, crypto.SHA256, hashed[:], sig, nil); err != nil {
		return fmt.Errorf("token signature invalid: %w", err)
	}
	expiry, err := time.Parse("2006-01-02", expiryStr)
	if err != nil {
		return fmt.Errorf("invalid expiry date: %w", err)
	}

	// Display license info and enforce as in existing solution
	fmt.Printf("\U0001F4C4 License Information:\n")
	fmt.Printf("   Company: %s\n", company)
	fmt.Printf("   Email: %s\n", email)
	fmt.Printf("   Expires: %s\n\n", expiry.Format("2006-01-02"))

	now := time.Now()
	if fakeNow := os.Getenv("FAKE_NOW"); fakeNow != "" {
		if parsed, err := time.Parse("2006-01-02", fakeNow); err == nil {
			now = parsed
		}
	}
	if now.After(expiry) {
		return fmt.Errorf("❌ Token expired (expiry: %s, now: %s)", expiry.Format("2006-01-02"), now.Format("2006-01-02"))
	}
	remaining := expiry.Sub(now).Hours() / 24
	if remaining < 0 {
		fmt.Printf("❌ Model access has expired %d days ago.\n", int(-remaining))
		fmt.Println("❌ Access denied. Please contact sales@sjfisher.com for license renewal.")
		os.Exit(1)
	} else if remaining <= 7 {
		fmt.Printf("⚠️ WARNING: Model access will expire in %d days (%s).\n", int(remaining), expiry.Format("2006-01-02"))
		fmt.Println("⚠️ Please contact sales@sjfisher.com for license renewal.")
	} else {
		fmt.Printf("✅ Model access valid for %d more days (expires %s).\n", int(remaining), expiry.Format("2006-01-02"))
	}
	if remaining <= 1 {
		return fmt.Errorf("❌ Model access blocked - license expires within 24 hours.")
	}
	return nil
}
