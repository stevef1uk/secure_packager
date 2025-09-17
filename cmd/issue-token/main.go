package main

import (
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
	"os"
	"time"
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

func main() {
	privPath := flag.String("priv", "", "Vendor RSA private key (PEM)")
	expiry := flag.String("expiry", "", "Expiry date YYYY-MM-DD")
	company := flag.String("company", "", "Company name")
	email := flag.String("email", "", "Email address")
	out := flag.String("out", "token.txt", "Output token path")
	flag.Parse()

	if *privPath == "" || *expiry == "" || *company == "" || *email == "" {
		fmt.Println("Usage: issue-token -priv vendor_private.pem -expiry YYYY-MM-DD -company NAME -email ADDRESS [-out token.txt]")
		os.Exit(1)
	}
	if _, err := time.Parse("2006-01-02", *expiry); err != nil {
		fmt.Fprintf(os.Stderr, "invalid expiry: %v\n", err)
		os.Exit(1)
	}

	priv, err := readRSAPrivateKey(*privPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading private key failed: %v\n", err)
		os.Exit(1)
	}

	// Keep placeholder for compatibility with existing format
	payload := fmt.Sprintf("%s:%s:%s:%s", *expiry, *company, *email, "NOFERNET")
	sum := sha256.Sum256([]byte(payload))
	sig, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, sum[:], nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign failed: %v\n", err)
		os.Exit(1)
	}
	sigB64 := base64.URLEncoding.EncodeToString(sig)
	token := fmt.Sprintf("%s:%s:%s:%s:%s", *expiry, *company, *email, "NOFERNET", sigB64)
	tokenB64 := base64.URLEncoding.EncodeToString([]byte(token))

	if err := os.WriteFile(*out, []byte(tokenB64), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write token failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Token issued -> %s\n", *out)
}
