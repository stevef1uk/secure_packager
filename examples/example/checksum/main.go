package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// SupportedAlgorithms returns a list of supported checksum algorithms
func SupportedAlgorithms() []string {
	return []string{"md5", "sha1", "sha256", "sha512"}
}

func main() {
	var (
		dirPath   = flag.String("dir", ".", "Directory to scan for files")
		algorithm = flag.String("algo", "sha256", "Checksum algorithm (md5, sha1, sha256, sha512)")
		help      = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("Checksum Calculator Demo")
		fmt.Println("========================")
		fmt.Println()
		fmt.Println("This program scans a directory and calculates checksums for all files.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Printf("Supported algorithms: %s\n", strings.Join(SupportedAlgorithms(), ", "))
		return
	}

	// Validate algorithm
	supported := false
	for _, algo := range SupportedAlgorithms() {
		if *algorithm == algo {
			supported = true
			break
		}
	}
	if !supported {
		fmt.Printf("Error: Unsupported algorithm '%s'. Supported algorithms: %s\n",
			*algorithm, strings.Join(SupportedAlgorithms(), ", "))
		os.Exit(1)
	}

	// Check if directory exists
	if _, err := os.Stat(*dirPath); os.IsNotExist(err) {
		fmt.Printf("Error: Directory '%s' does not exist\n", *dirPath)
		os.Exit(1)
	}

	// Create calculator and scan directory
	calculator := NewChecksumCalculator(*algorithm)
	if err := calculator.ScanDirectoryAndChecksum(*dirPath); err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nChecksum calculation completed!")
}
