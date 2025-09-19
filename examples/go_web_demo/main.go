package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// DemoConfig holds the configuration for the demo
type DemoConfig struct {
	WorkDir   string
	DataDir   string
	OutputDir string
	KeysDir   string
	LogsDir   string
}

// KeyGenRequest represents a key generation request
type KeyGenRequest struct {
	KeySize int `json:"key_size"`
}

// FileCreateRequest represents a file creation request
type FileCreateRequest struct {
	Content string `json:"content"`
}

// PackageRequest represents a packaging request
type PackageRequest struct {
	UseLicensing bool `json:"use_licensing"`
}

// TokenRequest represents a token issuance request
type TokenRequest struct {
	Company    string `json:"company"`
	Email      string `json:"email"`
	ExpiryDays int    `json:"expiry_days"`
}

// UnpackRequest represents an unpacking request
type UnpackRequest struct {
	UseLicensing bool `json:"use_licensing"`
}

// FileReadRequest represents a file read request
type FileReadRequest struct {
	Filename  string `json:"filename"`
	Directory string `json:"directory"`
}

// Response represents a generic API response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// FileInfo represents file information
type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

// DemoService handles the demo operations
type DemoService struct {
	config DemoConfig
}

// NewDemoService creates a new demo service
func NewDemoService() *DemoService {
	workDir := "/app"
	return &DemoService{
		config: DemoConfig{
			WorkDir:   workDir,
			DataDir:   filepath.Join(workDir, "data"),
			OutputDir: filepath.Join(workDir, "output"),
			KeysDir:   filepath.Join(workDir, "keys"),
			LogsDir:   filepath.Join(workDir, "logs"),
		},
	}
}

// Setup creates necessary directories
func (ds *DemoService) Setup() error {
	dirs := []string{ds.config.DataDir, ds.config.OutputDir, ds.config.KeysDir, ds.config.LogsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// runDockerCommand executes a Docker command
func (ds *DemoService) runDockerCommand(image string, command []string, volumes map[string]string) (string, error) {
	args := []string{"run", "--rm"}

	// Add volume mounts
	for host, container := range volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", host, container))
	}

	// Add image and command
	args = append(args, image)
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// GenerateKeys generates RSA key pairs
func (ds *DemoService) GenerateKeys(keySize int) error {
	// Check if keys already exist
	customerPrivate := filepath.Join(ds.config.KeysDir, "customer_private.pem")
	customerPublic := filepath.Join(ds.config.KeysDir, "customer_public.pem")
	vendorPrivate := filepath.Join(ds.config.KeysDir, "vendor_private.pem")
	vendorPublic := filepath.Join(ds.config.KeysDir, "vendor_public.pem")

	if _, err1 := os.Stat(customerPrivate); err1 == nil {
		if _, err2 := os.Stat(customerPublic); err2 == nil {
			if _, err3 := os.Stat(vendorPrivate); err3 == nil {
				if _, err4 := os.Stat(vendorPublic); err4 == nil {
					return fmt.Errorf("keys already exist - please delete them first if you want to regenerate")
				}
			}
		}
	}

	// Since we're running inside a container, we can't run Docker commands
	// Instead, we'll return an error with instructions
	return fmt.Errorf("key generation from within the container is not supported. Keys should be pre-generated before starting the container. Please run './generate_keys.sh' from the host system to generate keys.")
}

// CreateSampleFiles creates sample files for encryption
func (ds *DemoService) CreateSampleFiles(content string) error {
	// Create sample.txt
	sampleFile := filepath.Join(ds.config.DataDir, "sample.txt")
	if err := os.WriteFile(sampleFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create sample.txt: %w", err)
	}

	// Create config.json
	config := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
			"name": "secure_db",
		},
		"api": map[string]interface{}{
			"version":  "1.0",
			"endpoint": "/api/v1",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configFile := filepath.Join(ds.config.DataDir, "config.json")
	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		return fmt.Errorf("failed to create config.json: %w", err)
	}

	return nil
}

// PackageFiles packages files using secure_packager
func (ds *DemoService) PackageFiles(useLicensing bool) error {
	// Copy customer public key to output directory
	srcKey := filepath.Join(ds.config.KeysDir, "customer_public.pem")
	dstKey := filepath.Join(ds.config.OutputDir, "customer_public.pem")
	if err := copyFile(srcKey, dstKey); err != nil {
		return fmt.Errorf("failed to copy public key: %w", err)
	}

	// Build command arguments
	args := []string{
		"-in", ds.config.DataDir,
		"-out", ds.config.OutputDir,
		"-pub", dstKey,
		"-zip=true",
	}

	if useLicensing {
		// Copy vendor public key to output directory for licensing
		vendorSrcKey := filepath.Join(ds.config.KeysDir, "vendor_public.pem")
		vendorDstKey := filepath.Join(ds.config.OutputDir, "vendor_public.pem")
		if err := copyFile(vendorSrcKey, vendorDstKey); err != nil {
			return fmt.Errorf("failed to copy vendor public key: %w", err)
		}

		// Add licensing arguments
		args = append(args, "-license", "-vendor-pub", vendorDstKey)
	}

	// Run the packager command
	cmd := exec.Command("packager", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("packaging failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// IssueToken issues a license token
func (ds *DemoService) IssueToken(company, email string, expiryDays int) error {
	expiryDate := time.Now().AddDate(0, 0, expiryDays).Format("2006-01-02")

	// Build command arguments
	args := []string{
		"-priv", filepath.Join(ds.config.KeysDir, "vendor_private.pem"),
		"-expiry", expiryDate,
		"-company", company,
		"-email", email,
		"-out", filepath.Join(ds.config.KeysDir, "token.txt"),
	}

	// Run the issue-token command
	cmd := exec.Command("issue-token", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("token issuance failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// UnpackFiles unpacks encrypted files
func (ds *DemoService) UnpackFiles(useLicensing bool) (string, error) {
	decryptedDir := filepath.Join(ds.config.OutputDir, "decrypted")
	if err := os.MkdirAll(decryptedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create decrypted directory: %w", err)
	}

	// Build command arguments
	args := []string{
		"-zip", filepath.Join(ds.config.OutputDir, "encrypted_files.zip"),
		"-priv", filepath.Join(ds.config.KeysDir, "customer_private.pem"),
		"-out", decryptedDir,
	}

	if useLicensing {
		// Add licensing arguments
		args = append(args, "-license-token", filepath.Join(ds.config.KeysDir, "token.txt"))

		// Check if vendor public key exists in output directory (from packaging)
		vendorPubKey := filepath.Join(ds.config.OutputDir, "vendor_public.pem")
		if _, err := os.Stat(vendorPubKey); err == nil {
			// Vendor public key is in output directory, add it to args
			args = append(args, "-vendor-pub", vendorPubKey)
		} else {
			// Fallback to keys directory
			vendorPubKey = filepath.Join(ds.config.KeysDir, "vendor_public.pem")
			args = append(args, "-vendor-pub", vendorPubKey)
		}
	}

	// Run the unpack command
	cmd := exec.Command("unpack", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("unpacking failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// UnpackUploadedFiles unpacks uploaded encrypted files with custom keys
func (ds *DemoService) UnpackUploadedFiles(zipPath, customerPrivatePath, vendorPublicPath, tokenPath string, useLicensing bool) (string, error) {
	decryptedDir := filepath.Join(ds.config.OutputDir, "decrypted")
	if err := os.MkdirAll(decryptedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create decrypted directory: %w", err)
	}

	// Build command arguments
	args := []string{
		"-zip", zipPath,
		"-priv", customerPrivatePath,
		"-out", decryptedDir,
	}

	if useLicensing {
		// Add licensing arguments
		args = append(args, "-license-token", tokenPath)
		args = append(args, "-vendor-pub", vendorPublicPath)
	}

	// Run the unpack command
	cmd := exec.Command("unpack", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("unpacking failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// ListFiles lists files in a directory
func (ds *DemoService) ListFiles(directory string) ([]FileInfo, error) {
	var dir string
	switch directory {
	case "data":
		dir = ds.config.DataDir
	case "output":
		dir = ds.config.OutputDir
	case "decrypted":
		dir = filepath.Join(ds.config.OutputDir, "decrypted")
	default:
		return nil, fmt.Errorf("invalid directory: %s", directory)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		}

		files = append(files, FileInfo{
			Name: entry.Name(),
			Size: info.Size(),
			Type: fileType,
		})
	}

	return files, nil
}

// ReadFile reads the content of a file
func (ds *DemoService) ReadFile(filename, directory string) (string, error) {
	var dir string
	switch directory {
	case "data":
		dir = ds.config.DataDir
	case "output":
		dir = ds.config.OutputDir
	case "decrypted":
		dir = filepath.Join(ds.config.OutputDir, "decrypted")
	default:
		return "", fmt.Errorf("invalid directory: %s", directory)
	}

	filePath := filepath.Join(dir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return string(content), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func main() {
	// Initialize demo service
	demo := NewDemoService()
	if err := demo.Setup(); err != nil {
		log.Fatalf("Failed to setup demo: %v", err)
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Main page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Secure Packager Demo",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Generate keys
		api.POST("/keys/generate", func(c *gin.Context) {
			var req KeyGenRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Invalid request: " + err.Error(),
				})
				return
			}

			if req.KeySize < 1024 || req.KeySize > 4096 {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Key size must be between 1024 and 4096 bits",
				})
				return
			}

			if err := demo.GenerateKeys(req.KeySize); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to generate keys: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: fmt.Sprintf("Keys generated successfully with %d bits", req.KeySize),
			})
		})

		// Create sample files
		api.POST("/files/create", func(c *gin.Context) {
			var req FileCreateRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Invalid request: " + err.Error(),
				})
				return
			}

			if err := demo.CreateSampleFiles(req.Content); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to create sample files: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: "Sample files created successfully",
			})
		})

		// Package files
		api.POST("/package", func(c *gin.Context) {
			var req PackageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Invalid request: " + err.Error(),
				})
				return
			}

			if err := demo.PackageFiles(req.UseLicensing); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to package files: " + err.Error(),
				})
				return
			}

			message := "Files packaged successfully"
			if req.UseLicensing {
				message += " (with licensing)"
			} else {
				message += " (no licensing)"
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: message,
			})
		})

		// Issue token
		api.POST("/token/issue", func(c *gin.Context) {
			var req TokenRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Invalid request: " + err.Error(),
				})
				return
			}

			if err := demo.IssueToken(req.Company, req.Email, req.ExpiryDays); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to issue token: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: fmt.Sprintf("License token issued successfully for %s (expires in %d days)", req.Company, req.ExpiryDays),
			})
		})

		// Unpack files
		api.POST("/unpack", func(c *gin.Context) {
			var req UnpackRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Invalid request: " + err.Error(),
				})
				return
			}

			output, err := demo.UnpackFiles(req.UseLicensing)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to unpack files: " + err.Error(),
				})
				return
			}

			message := "Files unpacked successfully"
			if req.UseLicensing {
				message += " (with licensing verification)"
			} else {
				message += " (no licensing)"
			}

			// Include the output in the response for license details
			if output != "" {
				message += "\n\n" + output
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: message,
			})
		})

		// Download file
		api.GET("/files/download/:filename", func(c *gin.Context) {
			filename := c.Param("filename")
			if filename == "" {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Filename is required",
				})
				return
			}

			// Check if dir parameter is provided (for decrypted files)
			dir := c.Query("dir")
			var filePath string
			if dir == "decrypted" {
				filePath = filepath.Join(demo.config.OutputDir, "decrypted", filename)
			} else {
				filePath = filepath.Join(demo.config.OutputDir, filename)
			}

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, Response{
					Success: false,
					Message: "File not found: " + filename,
				})
				return
			}

			// Set headers for file download
			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", "attachment; filename="+filename)
			c.Header("Content-Type", "application/octet-stream")

			// Serve the file
			c.File(filePath)
		})

		// Clear output directory
		api.POST("/files/clear-output", func(c *gin.Context) {
			// Remove all files in output directory
			files, err := os.ReadDir(demo.config.OutputDir)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to read output directory: " + err.Error(),
				})
				return
			}

			var removedFiles []string
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(demo.config.OutputDir, file.Name())
					if err := os.Remove(filePath); err != nil {
						c.JSON(http.StatusInternalServerError, Response{
							Success: false,
							Message: "Failed to remove file " + file.Name() + ": " + err.Error(),
						})
						return
					}
					removedFiles = append(removedFiles, file.Name())
				}
			}

			// Also clear the decrypted subdirectory
			decryptedDir := filepath.Join(demo.config.OutputDir, "decrypted")
			decryptedFiles, err := os.ReadDir(decryptedDir)
			if err == nil {
				for _, file := range decryptedFiles {
					if !file.IsDir() {
						filePath := filepath.Join(decryptedDir, file.Name())
						if err := os.Remove(filePath); err == nil {
							removedFiles = append(removedFiles, "decrypted/"+file.Name())
						}
					}
				}
			}

			message := fmt.Sprintf("Cleared output directory. Removed %d files: %s", len(removedFiles), strings.Join(removedFiles, ", "))
			if len(removedFiles) == 0 {
				message = "Output directory is already empty"
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: message,
			})
		})

		// Clear decrypted directory
		api.POST("/files/clear-decrypted", func(c *gin.Context) {
			// Remove all files in decrypted directory
			decryptedDir := filepath.Join(demo.config.OutputDir, "decrypted")
			files, err := os.ReadDir(decryptedDir)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to read decrypted directory: " + err.Error(),
				})
				return
			}

			var removedFiles []string
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(decryptedDir, file.Name())
					if err := os.Remove(filePath); err != nil {
						c.JSON(http.StatusInternalServerError, Response{
							Success: false,
							Message: "Failed to remove file " + file.Name() + ": " + err.Error(),
						})
						return
					}
					removedFiles = append(removedFiles, file.Name())
				}
			}

			message := fmt.Sprintf("Cleared decrypted directory. Removed %d files: %s", len(removedFiles), strings.Join(removedFiles, ", "))
			if len(removedFiles) == 0 {
				message = "Decrypted directory is already empty"
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: message,
			})
		})

		// Clear data directory
		api.POST("/files/clear-data", func(c *gin.Context) {
			// Remove all files in data directory
			files, err := os.ReadDir(demo.config.DataDir)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to read data directory: " + err.Error(),
				})
				return
			}

			var removedFiles []string
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(demo.config.DataDir, file.Name())
					if err := os.Remove(filePath); err != nil {
						c.JSON(http.StatusInternalServerError, Response{
							Success: false,
							Message: "Failed to remove file " + file.Name() + ": " + err.Error(),
						})
						return
					}
					removedFiles = append(removedFiles, file.Name())
				}
			}

			message := fmt.Sprintf("Cleared data directory. Removed %d files: %s", len(removedFiles), strings.Join(removedFiles, ", "))
			if len(removedFiles) == 0 {
				message = "Data directory is already empty"
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: message,
			})
		})

		// Upload and unpack encrypted files
		api.POST("/files/upload-unpack", func(c *gin.Context) {
			// Parse multipart form
			form, err := c.MultipartForm()
			if err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Failed to parse form: " + err.Error(),
				})
				return
			}

			// Get uploaded files
			encryptedZip := form.File["encryptedZip"]
			customerPrivate := form.File["customerPrivate"]
			vendorPublic := form.File["vendorPublic"]
			token := form.File["token"]
			useLicensing := form.Value["useLicensing"][0] == "true"

			// Validate required files
			if len(encryptedZip) == 0 {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Encrypted ZIP file is required",
				})
				return
			}
			if len(customerPrivate) == 0 {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Customer private key is required",
				})
				return
			}
			if useLicensing {
				if len(vendorPublic) == 0 {
					c.JSON(http.StatusBadRequest, Response{
						Success: false,
						Message: "Vendor public key is required for licensing",
					})
					return
				}
				if len(token) == 0 {
					c.JSON(http.StatusBadRequest, Response{
						Success: false,
						Message: "License token is required for licensing",
					})
					return
				}
			}

			// Save uploaded files to temporary locations
			zipPath := filepath.Join(demo.config.OutputDir, "uploaded_encrypted.zip")
			customerPrivatePath := filepath.Join(demo.config.KeysDir, "uploaded_customer_private.pem")
			vendorPublicPath := filepath.Join(demo.config.KeysDir, "uploaded_vendor_public.pem")
			tokenPath := filepath.Join(demo.config.KeysDir, "uploaded_token.txt")

			// Save encrypted zip
			if err := c.SaveUploadedFile(encryptedZip[0], zipPath); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to save encrypted zip: " + err.Error(),
				})
				return
			}

			// Save customer private key
			if err := c.SaveUploadedFile(customerPrivate[0], customerPrivatePath); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to save customer private key: " + err.Error(),
				})
				return
			}

			// Save vendor public key and token if licensing is enabled
			if useLicensing {
				if err := c.SaveUploadedFile(vendorPublic[0], vendorPublicPath); err != nil {
					c.JSON(http.StatusInternalServerError, Response{
						Success: false,
						Message: "Failed to save vendor public key: " + err.Error(),
					})
					return
				}
				if err := c.SaveUploadedFile(token[0], tokenPath); err != nil {
					c.JSON(http.StatusInternalServerError, Response{
						Success: false,
						Message: "Failed to save license token: " + err.Error(),
					})
					return
				}
			}

			// Unpack the uploaded files
			output, err := demo.UnpackUploadedFiles(zipPath, customerPrivatePath, vendorPublicPath, tokenPath, useLicensing)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to unpack files: " + err.Error(),
				})
				return
			}

			message := "Files uploaded and unpacked successfully"
			if useLicensing {
				message += " (with licensing verification)"
			}
			if output != "" {
				message += "\n\n" + output
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: message,
			})
		})

		// Upload files
		api.POST("/files/upload", func(c *gin.Context) {
			// Parse multipart form
			form, err := c.MultipartForm()
			if err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Failed to parse form: " + err.Error(),
				})
				return
			}

			files := form.File["files"]
			if len(files) == 0 {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "No files uploaded",
				})
				return
			}

			var uploadedFiles []string
			for _, file := range files {
				// Save file to data directory
				dst := filepath.Join(demo.config.DataDir, file.Filename)
				if err := c.SaveUploadedFile(file, dst); err != nil {
					c.JSON(http.StatusInternalServerError, Response{
						Success: false,
						Message: "Failed to save file " + file.Filename + ": " + err.Error(),
					})
					return
				}
				uploadedFiles = append(uploadedFiles, file.Filename)
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: fmt.Sprintf("Successfully uploaded %d files", len(uploadedFiles)),
				Data:    uploadedFiles,
			})
		})

		// List files
		api.GET("/files/:directory", func(c *gin.Context) {
			directory := c.Param("directory")
			files, err := demo.ListFiles(directory)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to list files: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: "Files listed successfully",
				Data:    files,
			})
		})

		// Read file
		api.POST("/files/read", func(c *gin.Context) {
			var req FileReadRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "Invalid request: " + err.Error(),
				})
				return
			}

			content, err := demo.ReadFile(req.Filename, req.Directory)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "Failed to read file: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: "File read successfully",
				Data:    content,
			})
		})

		// Run complete workflow
		api.POST("/workflow/complete", func(c *gin.Context) {
			var steps []string

			// Step 1: Verify keys exist (pre-generated)
			steps = append(steps, "🔑 Step 1: Verifying RSA key pairs...")
			customerPrivate := filepath.Join(demo.config.KeysDir, "customer_private.pem")
			customerPublic := filepath.Join(demo.config.KeysDir, "customer_public.pem")
			vendorPrivate := filepath.Join(demo.config.KeysDir, "vendor_private.pem")
			vendorPublic := filepath.Join(demo.config.KeysDir, "vendor_public.pem")

			if _, err := os.Stat(customerPrivate); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, "   ❌ Failed: customer private key not found - keys should be pre-generated"), "\n"),
				})
				return
			}
			if _, err := os.Stat(customerPublic); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, "   ❌ Failed: customer public key not found - keys should be pre-generated"), "\n"),
				})
				return
			}
			if _, err := os.Stat(vendorPrivate); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, "   ❌ Failed: vendor private key not found - keys should be pre-generated"), "\n"),
				})
				return
			}
			if _, err := os.Stat(vendorPublic); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, "   ❌ Failed: vendor public key not found - keys should be pre-generated"), "\n"),
				})
				return
			}
			steps = append(steps, "   ✅ Keys verified successfully")

			// Step 2: Create sample files
			steps = append(steps, "\n📄 Step 2: Creating sample files...")
			if err := demo.CreateSampleFiles("Complete workflow demo file content."); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, fmt.Sprintf("   ❌ Failed: %s", err.Error())), "\n"),
				})
				return
			}
			steps = append(steps, "   ✅ Sample files created")

			// Step 3: Package without licensing
			steps = append(steps, "\n📦 Step 3: Packaging files (no licensing)...")
			if err := demo.PackageFiles(false); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, fmt.Sprintf("   ❌ Failed: %s", err.Error())), "\n"),
				})
				return
			}
			steps = append(steps, "   ✅ Files packaged (no licensing)")

			// Step 4: Package with licensing
			steps = append(steps, "\n📦 Step 4: Packaging files (with licensing)...")
			if err := demo.PackageFiles(true); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, fmt.Sprintf("   ❌ Failed: %s", err.Error())), "\n"),
				})
				return
			}
			steps = append(steps, "   ✅ Files packaged (with licensing)")

			// Step 5: Issue token
			steps = append(steps, "\n🎫 Step 5: Issuing license token...")
			if err := demo.IssueToken("Demo Co", "demo@example.com", 365); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, fmt.Sprintf("   ❌ Failed: %s", err.Error())), "\n"),
				})
				return
			}
			steps = append(steps, "   ✅ License token issued")

			// Step 6: Unpack without licensing
			steps = append(steps, "\n📤 Step 6: Unpacking files (no licensing)...")
			output1, err := demo.UnpackFiles(false)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, fmt.Sprintf("   ❌ Failed: %s", err.Error())), "\n"),
				})
				return
			}
			steps = append(steps, "   ✅ Files unpacked (no licensing)")
			if output1 != "" {
				steps = append(steps, "   📄 Output: "+output1)
			}

			// Step 7: Unpack with licensing
			steps = append(steps, "\n📤 Step 7: Unpacking files (with licensing)...")
			output2, err := demo.UnpackFiles(true)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: strings.Join(append(steps, fmt.Sprintf("   ❌ Failed: %s", err.Error())), "\n"),
				})
				return
			}
			steps = append(steps, "   ✅ Files unpacked (with licensing)")
			if output2 != "" {
				steps = append(steps, "   📄 License Details: "+output2)
			}

			steps = append(steps, "\n✅ Complete workflow finished successfully!")
			steps = append(steps, "\n📁 Check the File Browser tab to view the results.")

			c.JSON(http.StatusOK, Response{
				Success: true,
				Message: strings.Join(steps, "\n"),
			})
		})
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(r.Run(":" + port))
}
