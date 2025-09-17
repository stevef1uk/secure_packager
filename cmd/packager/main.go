package main

import (
	"archive/zip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fernet/fernet-go"
)

func readRSAPublicKey(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if k, ok := pub.(*rsa.PublicKey); ok {
			return k, nil
		}
		return nil, errors.New("not RSA public key")
	}
	k, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func encryptFilesWithFernet(key *fernet.Key, inputDir, outputDir string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		inPath := filepath.Join(inputDir, e.Name())
		outPath := filepath.Join(outputDir, e.Name()+".enc")
		data, err := os.ReadFile(inPath)
		if err != nil {
			return err
		}
		ct, err := fernet.EncryptAndSign(data, key)
		if err != nil {
			return err
		}
		if err := os.WriteFile(outPath, ct, 0644); err != nil {
			return err
		}
		fmt.Printf("Encrypted %s -> %s\n", e.Name(), filepath.Base(outPath))
	}
	return nil
}

func wrapFernetKey(pub *rsa.PublicKey, key *fernet.Key) ([]byte, error) {
	// Encrypt the base64-encoded fernet key string bytes with RSA-OAEP
	enc := []byte(key.Encode())
	label := []byte("secure_packager")
	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, enc, label)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

func zipOutputs(srcDir, zipPath string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()

	addFile := func(path, name string) error {
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, in)
		return err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		p := filepath.Join(srcDir, e.Name())
		if err := addFile(p, e.Name()); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	inputDir := flag.String("in", "", "Input directory with files to encrypt")
	outDir := flag.String("out", "", "Output directory for encrypted payload")
	customerPub := flag.String("pub", "", "Path to customer's RSA public key (PEM)")
	makeZip := flag.Bool("zip", true, "Also create encrypted_files.zip in output directory")
	cleanup := flag.Bool("cleanup", true, "After zipping, remove generated .enc files and helper artifacts")
	licenseMode := flag.Bool("license", false, "If set, write manifest to require license check in unzip")
	vendorPubPath := flag.String("vendor-pub", "", "Vendor public key (PEM) to embed for license verification when -license is set")
	flag.Parse()

	if *inputDir == "" || *outDir == "" || *customerPub == "" {
		fmt.Println("Usage: packager -in <input_dir> -out <output_dir> -pub <customer_public.pem> [-zip=true]")
		os.Exit(1)
	}

	pub, err := readRSAPublicKey(*customerPub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read public key: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	k := new(fernet.Key)
	if err := k.Generate(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate fernet key: %v\n", err)
		os.Exit(1)
	}

	if err := encryptFilesWithFernet(k, *inputDir, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Encryption failed: %v\n", err)
		os.Exit(1)
	}

	wrapped, err := wrapFernetKey(pub, k)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Wrapping key failed: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(filepath.Join(*outDir, "wrapped_key.bin"), wrapped, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Writing wrapped key failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Wrote wrapped_key.bin")

	// Optional: include licensing manifest and vendor public key for verification at unpack time
	if *licenseMode {
		if vendorPubPath == nil || strings.TrimSpace(*vendorPubPath) == "" {
			fmt.Fprintln(os.Stderr, "-license requires -vendor-pub <vendor_public.pem>")
			os.Exit(1)
		}
		manifest := []byte("{\n  \"license_required\": true,\n  \"vendor_public_key\": \"vendor_public.pem\"\n}\n")
		if err := os.WriteFile(filepath.Join(*outDir, "manifest.json"), manifest, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Writing manifest failed: %v\n", err)
			os.Exit(1)
		}
		// Copy vendor public key alongside manifest so the unpacker can verify tokens without external files
		vp, err := os.ReadFile(*vendorPubPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Reading vendor public key failed: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(filepath.Join(*outDir, "vendor_public.pem"), vp, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Writing vendor public key failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Wrote manifest.json and vendor_public.pem for license enforcement")
	}

	if *makeZip {
		zipPath := filepath.Join(*outDir, "encrypted_files.zip")
		if err := zipOutputs(*outDir, zipPath); err != nil {
			fmt.Fprintf(os.Stderr, "Zipping failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created %s\n", zipPath)
		if *cleanup {
			// Remove generated artifacts, but keep the zip and any user-provided files
			entries, err := os.ReadDir(*outDir)
			if err == nil {
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					name := e.Name()
					// Keep the final zip
					if name == "encrypted_files.zip" {
						continue
					}
					// Remove our generated files: .enc, wrapped_key.bin, manifest.json, vendor_public.pem
					if strings.HasSuffix(name, ".enc") || name == "wrapped_key.bin" || name == "manifest.json" || name == "vendor_public.pem" {
						_ = os.Remove(filepath.Join(*outDir, name))
					}
				}
			}
		}
	}
}
