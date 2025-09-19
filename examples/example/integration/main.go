package main

import (
	"archive/zip"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fernet/fernet-go"
)

// ChecksumCalculator provides methods to calculate various checksums for files
type ChecksumCalculator struct {
	algorithm string
}

// NewChecksumCalculator creates a new ChecksumCalculator with the specified algorithm
func NewChecksumCalculator(algorithm string) *ChecksumCalculator {
	return &ChecksumCalculator{
		algorithm: strings.ToLower(algorithm),
	}
}

// CalculateFileChecksum calculates the checksum of a single file
func (cc *ChecksumCalculator) CalculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	hash, err := cc.createHash()
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ScanDirectoryAndChecksum scans a directory and calculates checksums for all files
func (cc *ChecksumCalculator) ScanDirectoryAndChecksum(dirPath string) error {
	fmt.Printf("Scanning directory: %s\n", dirPath)
	fmt.Printf("Using %s algorithm\n\n", strings.ToUpper(cc.algorithm))
	fmt.Printf("%-50s %s\n", "File", "Checksum")
	fmt.Println(strings.Repeat("-", 80))

	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate checksum
		checksum, err := cc.CalculateFileChecksum(path)
		if err != nil {
			fmt.Printf("Error calculating checksum for %s: %v\n", path, err)
			return nil // Continue with other files
		}

		// Get relative path for cleaner output
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			relPath = path
		}

		fmt.Printf("%-50s %s\n", relPath, checksum)
		return nil
	})
}

// createHash creates the appropriate hash.Hash based on the algorithm
func (cc *ChecksumCalculator) createHash() (hash.Hash, error) {
	switch cc.algorithm {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s. Supported: md5, sha1, sha256, sha512", cc.algorithm)
	}
}

// SecurePackager provides methods to encrypt and decrypt files using envelope encryption
type SecurePackager struct {
	customerPubKey *rsa.PublicKey
	vendorPubKey   *rsa.PublicKey
	fernetKey      *fernet.Key
}

// NewSecurePackager creates a new SecurePackager instance
func NewSecurePackager(customerPubKeyPath string) (*SecurePackager, error) {
	customerPub, err := readRSAPublicKey(customerPubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read customer public key: %w", err)
	}

	// Generate a new Fernet key
	fernetKey := new(fernet.Key)
	if err := fernetKey.Generate(); err != nil {
		return nil, fmt.Errorf("failed to generate fernet key: %w", err)
	}

	return &SecurePackager{
		customerPubKey: customerPub,
		fernetKey:      fernetKey,
	}, nil
}

// NewSecurePackagerWithLicense creates a new SecurePackager instance with licensing support
func NewSecurePackagerWithLicense(customerPubKeyPath, vendorPubKeyPath string) (*SecurePackager, error) {
	sp, err := NewSecurePackager(customerPubKeyPath)
	if err != nil {
		return nil, err
	}

	vendorPub, err := readRSAPublicKey(vendorPubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vendor public key: %w", err)
	}

	sp.vendorPubKey = vendorPub
	return sp, nil
}

// EncryptDirectory encrypts all files in a directory and creates a zip archive
func (sp *SecurePackager) EncryptDirectory(inputDir, outputDir string, withLicense bool) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Encrypt files with Fernet
	if err := sp.encryptFilesWithFernet(inputDir, outputDir); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Wrap the Fernet key with RSA
	wrapped, err := sp.wrapFernetKey()
	if err != nil {
		return fmt.Errorf("wrapping key failed: %w", err)
	}

	// Write wrapped key
	wrappedPath := filepath.Join(outputDir, "wrapped_key.bin")
	if err := os.WriteFile(wrappedPath, wrapped, 0644); err != nil {
		return fmt.Errorf("writing wrapped key failed: %w", err)
	}

	// Add licensing manifest if requested
	if withLicense {
		if sp.vendorPubKey == nil {
			return fmt.Errorf("license mode requires vendor public key")
		}
		if err := sp.addLicenseManifest(outputDir); err != nil {
			return fmt.Errorf("adding license manifest failed: %w", err)
		}
	}

	// Create zip archive
	zipPath := filepath.Join(outputDir, "encrypted_files.zip")
	if err := sp.zipOutputs(outputDir, zipPath); err != nil {
		return fmt.Errorf("zipping failed: %w", err)
	}

	// Clean up temporary files
	return sp.cleanupTempFiles(outputDir)
}

// DecryptZip decrypts a zip archive created by EncryptDirectory
func (sp *SecurePackager) DecryptZip(zipPath, outputDir string, customerPrivKeyPath string, licenseTokenPath string) error {
	// Create working directory for extraction
	workDir := filepath.Join(filepath.Dir(zipPath), "_unpack")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Extract zip
	if err := sp.unzip(zipPath, workDir); err != nil {
		return fmt.Errorf("unzip failed: %w", err)
	}

	// Check for license requirements
	requireLicense, vendorPubPath, err := sp.checkLicenseRequirements(workDir)
	if err != nil {
		return fmt.Errorf("license check failed: %w", err)
	}

	// Verify license if required
	if requireLicense {
		if licenseTokenPath == "" {
			return fmt.Errorf("license required but no token provided")
		}
		if err := sp.verifyAndEnforceLicense(vendorPubPath, licenseTokenPath); err != nil {
			return fmt.Errorf("license verification failed: %w", err)
		}
	}

	// Read wrapped key
	wrappedPath := filepath.Join(workDir, "wrapped_key.bin")
	wrapped, err := os.ReadFile(wrappedPath)
	if err != nil {
		return fmt.Errorf("reading wrapped key failed: %w", err)
	}

	// Read customer private key
	customerPriv, err := readRSAPrivateKey(customerPrivKeyPath)
	if err != nil {
		return fmt.Errorf("reading private key failed: %w", err)
	}

	// Unwrap Fernet key
	fernetKey, err := sp.unwrapFernetKey(customerPriv, wrapped)
	if err != nil {
		return fmt.Errorf("unwrap failed: %w", err)
	}

	// Decrypt files
	if err := sp.decryptDirWithFernet(fernetKey, workDir, outputDir); err != nil {
		return fmt.Errorf("decrypt failed: %w", err)
	}

	return nil
}

// IntegrationExample demonstrates how to integrate secure_packager with a file processing application
type IntegrationExample struct {
	checksumCalc *ChecksumCalculator
	packager     *SecurePackager
	workDir      string
}

// NewIntegrationExample creates a new integration example instance
func NewIntegrationExample(workDir string) *IntegrationExample {
	return &IntegrationExample{
		workDir: workDir,
	}
}

// SetupKeys generates RSA key pairs for demonstration
func (ie *IntegrationExample) SetupKeys() error {
	fmt.Println("🔑 Setting up RSA key pairs...")

	// Create keys directory
	keysDir := filepath.Join(ie.workDir, "keys")
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	// Generate customer key pair
	customerPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate customer private key: %w", err)
	}

	customerPrivPath := filepath.Join(keysDir, "customer_private.pem")
	customerPubPath := filepath.Join(keysDir, "customer_public.pem")

	if err := ie.savePrivateKey(customerPriv, customerPrivPath); err != nil {
		return fmt.Errorf("failed to save customer private key: %w", err)
	}

	if err := ie.savePublicKey(&customerPriv.PublicKey, customerPubPath); err != nil {
		return fmt.Errorf("failed to save customer public key: %w", err)
	}

	// Generate vendor key pair (for licensing)
	vendorPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate vendor private key: %w", err)
	}

	vendorPrivPath := filepath.Join(keysDir, "vendor_private.pem")
	vendorPubPath := filepath.Join(keysDir, "vendor_public.pem")

	if err := ie.savePrivateKey(vendorPriv, vendorPrivPath); err != nil {
		return fmt.Errorf("failed to save vendor private key: %w", err)
	}

	if err := ie.savePublicKey(&vendorPriv.PublicKey, vendorPubPath); err != nil {
		return fmt.Errorf("failed to save vendor public key: %w", err)
	}

	fmt.Printf("   Customer keys: %s, %s\n", customerPrivPath, customerPubPath)
	fmt.Printf("   Vendor keys: %s, %s\n", vendorPrivPath, vendorPubPath)
	return nil
}

// CreateSampleData creates sample files for demonstration
func (ie *IntegrationExample) CreateSampleData() error {
	fmt.Println("📁 Creating sample data files...")

	dataDir := filepath.Join(ie.workDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create sample files with different content
	sampleFiles := map[string]string{
		"document.txt": "This is a confidential document containing sensitive information.",
		"data.json":    `{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}`,
		"config.yaml":  "database:\n  host: localhost\n  port: 5432\n  password: secret123",
		"readme.md":    "# Project Documentation\n\nThis project contains important files.",
		"binary.dat":   "Binary data: " + strings.Repeat("X", 1000),
	}

	for filename, content := range sampleFiles {
		filePath := filepath.Join(dataDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
		fmt.Printf("   Created: %s\n", filename)
	}

	return nil
}

// ProcessFilesWithChecksums processes files and calculates checksums
func (ie *IntegrationExample) ProcessFilesWithChecksums(dataDir string) error {
	fmt.Println("\n🔍 Processing files and calculating checksums...")

	ie.checksumCalc = NewChecksumCalculator("sha256")
	return ie.checksumCalc.ScanDirectoryAndChecksum(dataDir)
}

// EncryptFiles encrypts the data directory using secure_packager
func (ie *IntegrationExample) EncryptFiles(dataDir string, withLicense bool) error {
	fmt.Println("\n🔐 Encrypting files with secure_packager...")

	customerPubPath := filepath.Join(ie.workDir, "keys", "customer_public.pem")

	var err error
	if withLicense {
		vendorPubPath := filepath.Join(ie.workDir, "keys", "vendor_public.pem")
		ie.packager, err = NewSecurePackagerWithLicense(customerPubPath, vendorPubPath)
	} else {
		ie.packager, err = NewSecurePackager(customerPubPath)
	}

	if err != nil {
		return fmt.Errorf("failed to create packager: %w", err)
	}

	outputDir := filepath.Join(ie.workDir, "encrypted")
	if err := ie.packager.EncryptDirectory(dataDir, outputDir, withLicense); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	fmt.Printf("   Files encrypted and stored in: %s\n", outputDir)
	return nil
}

// DecryptAndVerifyFiles decrypts files and verifies checksums
func (ie *IntegrationExample) DecryptAndVerifyFiles(withLicense bool) error {
	fmt.Println("\n🔓 Decrypting files and verifying checksums...")

	zipPath := filepath.Join(ie.workDir, "encrypted", "encrypted_files.zip")
	customerPrivPath := filepath.Join(ie.workDir, "keys", "customer_private.pem")
	outputDir := filepath.Join(ie.workDir, "decrypted")

	var licenseTokenPath string
	if withLicense {
		licenseTokenPath = filepath.Join(ie.workDir, "keys", "token.txt")
		// Create a dummy token for demo purposes
		if err := ie.createDummyToken(licenseTokenPath); err != nil {
			return fmt.Errorf("failed to create dummy token: %w", err)
		}
	}

	if err := ie.packager.DecryptZip(zipPath, outputDir, customerPrivPath, licenseTokenPath); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	fmt.Println("\n🔍 Verifying decrypted files with checksums...")
	ie.checksumCalc = NewChecksumCalculator("sha256")
	return ie.checksumCalc.ScanDirectoryAndChecksum(outputDir)
}

// createDummyToken creates a dummy license token for demonstration
func (ie *IntegrationExample) createDummyToken(tokenPath string) error {
	// This is a simplified dummy token - in production you'd use the issue-token command
	token := "dummy_token_for_demo_purposes"
	return os.WriteFile(tokenPath, []byte(token), 0644)
}

// Helper functions for key management

func (ie *IntegrationExample) savePrivateKey(key *rsa.PrivateKey, path string) error {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	return os.WriteFile(path, keyPEM, 0600)
}

func (ie *IntegrationExample) savePublicKey(key *rsa.PublicKey, path string) error {
	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keyBytes,
	})

	return os.WriteFile(path, keyPEM, 0644)
}

// RunDemo runs the complete integration demonstration
func (ie *IntegrationExample) RunDemo(withLicense bool) error {
	fmt.Println("🚀 Secure Packager Integration Demo")
	fmt.Println("==================================")

	// Setup
	if err := ie.SetupKeys(); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	dataDir := filepath.Join(ie.workDir, "data")
	if err := ie.CreateSampleData(); err != nil {
		return fmt.Errorf("data creation failed: %w", err)
	}

	// Process files
	if err := ie.ProcessFilesWithChecksums(dataDir); err != nil {
		return fmt.Errorf("file processing failed: %w", err)
	}

	// Encrypt files
	if err := ie.EncryptFiles(dataDir, withLicense); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Decrypt and verify
	if err := ie.DecryptAndVerifyFiles(withLicense); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	fmt.Println("\n✅ Demo completed successfully!")
	fmt.Printf("   Work directory: %s\n", ie.workDir)
	fmt.Println("   Check the 'encrypted' and 'decrypted' directories for results.")

	return nil
}

// Helper functions (extracted from the original commands)

func readRSAPublicKey(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM")
	}
	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if k, ok := pub.(*rsa.PublicKey); ok {
			return k, nil
		}
		return nil, fmt.Errorf("not RSA public key")
	}
	k, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func readRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM")
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
		return nil, fmt.Errorf("PEM is not RSA private key")
	}
	return k, nil
}

func (sp *SecurePackager) encryptFilesWithFernet(inputDir, outputDir string) error {
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
		ct, err := fernet.EncryptAndSign(data, sp.fernetKey)
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

func (sp *SecurePackager) wrapFernetKey() ([]byte, error) {
	enc := []byte(sp.fernetKey.Encode())
	label := []byte("secure_packager")
	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, sp.customerPubKey, enc, label)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

func (sp *SecurePackager) unwrapFernetKey(priv *rsa.PrivateKey, wrapped []byte) (*fernet.Key, error) {
	label := []byte("secure_packager")
	raw, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, wrapped, label)
	if err != nil {
		return nil, err
	}
	keys := fernet.MustDecodeKeys(string(raw))
	if len(keys) == 0 {
		return nil, fmt.Errorf("failed to decode fernet key")
	}
	return keys[0], nil
}

func (sp *SecurePackager) addLicenseManifest(outputDir string) error {
	manifest := []byte("{\n  \"license_required\": true,\n  \"vendor_public_key\": \"vendor_public.pem\"\n}\n")
	if err := os.WriteFile(filepath.Join(outputDir, "manifest.json"), manifest, 0644); err != nil {
		return err
	}

	// Copy vendor public key
	vendorPubBytes, err := x509.MarshalPKIXPublicKey(sp.vendorPubKey)
	if err != nil {
		return err
	}
	vendorPubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: vendorPubBytes,
	})
	if err := os.WriteFile(filepath.Join(outputDir, "vendor_public.pem"), vendorPubPEM, 0644); err != nil {
		return err
	}

	fmt.Println("Added license manifest and vendor public key")
	return nil
}

func (sp *SecurePackager) checkLicenseRequirements(workDir string) (bool, string, error) {
	manifestPath := filepath.Join(workDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return false, "", nil
	}

	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return false, "", err
	}

	s := string(b)
	requireLicense := strings.Contains(s, "\"license_required\": true")
	vendorPubPath := ""
	if strings.Contains(s, "vendor_public.pem") {
		vendorPubPath = filepath.Join(workDir, "vendor_public.pem")
	}

	return requireLicense, vendorPubPath, nil
}

func (sp *SecurePackager) verifyAndEnforceLicense(vendorPubPath, tokenPath string) error {
	// This is a simplified version - in production you'd want full license verification
	fmt.Println("🔐 License verification would be performed here")
	fmt.Printf("   Vendor public key: %s\n", vendorPubPath)
	fmt.Printf("   License token: %s\n", tokenPath)
	return nil
}

func (sp *SecurePackager) zipOutputs(srcDir, zipPath string) error {
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

func (sp *SecurePackager) unzip(src, dest string) error {
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

func (sp *SecurePackager) decryptDirWithFernet(k *fernet.Key, srcDir, destDir string) error {
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

func (sp *SecurePackager) cleanupTempFiles(outputDir string) error {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
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
			_ = os.Remove(filepath.Join(outputDir, name))
		}
	}
	return nil
}

func main() {
	var (
		workDir     = flag.String("work", "./demo_work", "Working directory for demo files")
		withLicense = flag.Bool("license", false, "Enable licensing mode")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("Secure Packager Integration Example")
		fmt.Println("==================================")
		fmt.Println()
		fmt.Println("This example demonstrates how to integrate secure_packager with a file processing")
		fmt.Println("application. It shows how to:")
		fmt.Println("  1. Process files and calculate checksums")
		fmt.Println("  2. Encrypt files using secure_packager")
		fmt.Println("  3. Decrypt files and verify checksums")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	// Create working directory
	if err := os.MkdirAll(*workDir, 0755); err != nil {
		fmt.Printf("Error: Failed to create work directory: %v\n", err)
		os.Exit(1)
	}

	// Run demo
	example := NewIntegrationExample(*workDir)
	if err := example.RunDemo(*withLicense); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
