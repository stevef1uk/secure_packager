package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	defaultAppPort          = "8080"
	defaultPrivateKeyPath   = "/app/keys/customer_private.pem"
	defaultTokenFilePath    = "/app/keys/token.txt"
	defaultEncryptedZipPath = "/app/data/encrypted_files.zip"
	defaultDecryptOutputDir = "/app/decrypted"
	maxWaitTime             = 60 * time.Second
	checkInterval           = 2 * time.Second
)

type Config struct {
	PrivateKeyPath   string
	TokenFilePath    string
	EncryptedZipPath string
	DecryptOutputDir string
	AppPort          string
}

func loadConfig() *Config {
	return &Config{
		PrivateKeyPath:   getEnvWithDefault("PRIVATE_KEY_PATH", defaultPrivateKeyPath),
		TokenFilePath:    getEnvWithDefault("TOKEN_FILE_PATH", defaultTokenFilePath),
		EncryptedZipPath: getEnvWithDefault("ENCRYPTED_ZIP_PATH", defaultEncryptedZipPath),
		DecryptOutputDir: getEnvWithDefault("DECRYPT_OUTPUT_DIR", defaultDecryptOutputDir),
		AppPort:          getEnvWithDefault("APP_PORT", defaultAppPort),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func checkMountPoint(path string) bool {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), path) {
			return true
		}
	}
	return false
}

func printLicenseHeader(config *Config) {
	showLicensedMessage := checkMountPoint("/app/keys") ||
		checkMountPoint("/app/data") ||
		fileExists(config.PrivateKeyPath)

	fmt.Println("====================================================")
	fmt.Println(" Secure Packager Integration Example")
	if showLicensedMessage {
		fmt.Println(" Licensed container with encrypted data")
	} else {
		fmt.Println(" Demo container - no encrypted data mounted")
	}
	fmt.Println(" This container demonstrates secure file processing")
	fmt.Println("====================================================")
}

func createDirectories(config *Config) error {
	return os.MkdirAll(config.DecryptOutputDir, 0755)
}

func autoSelectFiles(config *Config) {
	// Auto-pick first zip if specific zip not found
	if !fileExists(config.EncryptedZipPath) {
		dataDir := "/app/data"
		if dirExists(dataDir) {
			if zipFiles, err := filepath.Glob(filepath.Join(dataDir, "*.zip")); err == nil && len(zipFiles) > 0 {
				config.EncryptedZipPath = zipFiles[0]
				fmt.Printf("[entrypoint] Auto-selected encrypted zip: %s\n", config.EncryptedZipPath)
			}
		}
	}

	// Auto-pick first token if specific token not found
	if !fileExists(config.TokenFilePath) {
		keysDir := "/app/keys"
		if dirExists(keysDir) {
			if tokenFiles, err := filepath.Glob(filepath.Join(keysDir, "token.txt")); err == nil && len(tokenFiles) > 0 {
				config.TokenFilePath = tokenFiles[0]
				fmt.Printf("[entrypoint] Auto-selected token: %s\n", config.TokenFilePath)
			}
		}
	}
}

func runDecryption(config *Config) error {
	if !fileExists(config.EncryptedZipPath) || !fileExists(config.PrivateKeyPath) {
		fmt.Println("[entrypoint] Skipping decryption (zip or private key missing).")
		fmt.Println("[entrypoint] Container will start with empty decrypted directory.")
		return nil
	}

	fmt.Println("[entrypoint] Extracting files from the zip archive using secure_packager...")

	// Change to the /app/tmp directory where we have write permissions
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir("/app/tmp"); err != nil {
		return fmt.Errorf("failed to change to /app/tmp directory: %w", err)
	}

	// Ensure we change back to original directory
	defer os.Chdir(originalDir)

	// Build args for secure_packager unpack CLI
	args := []string{
		"/app/unpack",
		"-zip", config.EncryptedZipPath,
		"-priv", config.PrivateKeyPath,
		"-out", config.DecryptOutputDir,
	}
	// Pass token only if present; unpack auto-detects licensing
	if fileExists(config.TokenFilePath) {
		args = append(args, "-license-token", config.TokenFilePath)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	fmt.Println("[entrypoint] Decryption completed successfully")
	return nil
}

func waitForApp(config *Config) error {
	log.Println("[entrypoint] Waiting for application to be ready...")
	address := fmt.Sprintf("http://127.0.0.1:%s", config.AppPort)
	deadline := time.Now().Add(maxWaitTime)
	debugMode := os.Getenv("DEBUG") != ""

	for time.Now().Before(deadline) {
		resp, err := http.Get(address + "/health")
		if err != nil {
			if debugMode {
				log.Printf("[entrypoint] Health check failed with error: %v", err)
			}
		} else {
			if debugMode {
				log.Printf("[entrypoint] Health check response: status=%d", resp.StatusCode)
			}
			if resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				log.Println("[entrypoint] ✅ Application is ready")
				return nil
			}
			resp.Body.Close()
		}
		log.Printf("[entrypoint] Application not ready, retrying in %v...", checkInterval)
		time.Sleep(checkInterval)
	}

	return fmt.Errorf("application did not become ready within %s", maxWaitTime)
}

func startApp(config *Config) error {
	fmt.Println("[entrypoint] Launching file processor application...")

	// Set environment variables for the app
	os.Setenv("APP_PORT", config.AppPort)
	os.Setenv("DECRYPT_OUTPUT_DIR", config.DecryptOutputDir)

	// Replace current process with the application
	return syscall.Exec("/app/app", []string{"/app/app"}, os.Environ())
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	return !os.IsNotExist(err) && info.IsDir()
}

func main() {
	// Check if this is a health check request
	if len(os.Args) > 1 && os.Args[1] == "--health-check" {
		config := loadConfig()
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/health", config.AppPort))
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		resp.Body.Close()
		os.Exit(0)
	}

	config := loadConfig()

	printLicenseHeader(config)

	fmt.Println("[entrypoint] Starting decryption step...")
	if err := createDirectories(config); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	autoSelectFiles(config)

	if err := runDecryption(config); err != nil {
		log.Printf("Decryption failed: %v", err)
		log.Println("[entrypoint] Continuing without decrypted files...")
	}

	// Start the application in the background
	go func() {
		if err := startApp(config); err != nil {
			log.Fatalf("Failed to start application: %v", err)
		}
	}()

	// Wait for the application to be ready
	if err := waitForApp(config); err != nil {
		log.Printf("Warning: application health check failed: %v", err)
	}

	// Keep the entrypoint running
	fmt.Println("[entrypoint] Application started successfully")
	fmt.Println("[entrypoint] Entrypoint will continue running to maintain container health")

	// Simple keep-alive loop
	for {
		time.Sleep(30 * time.Second)
	}
}
