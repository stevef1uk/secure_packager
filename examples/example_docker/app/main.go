package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo represents information about a processed file
type FileInfo struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Checksum  string    `json:"checksum"`
	Algorithm string    `json:"algorithm"`
	Modified  time.Time `json:"modified"`
}

// FileProcessor handles file processing operations
type FileProcessor struct {
	algorithm string
	baseDir   string
}

// NewFileProcessor creates a new file processor
func NewFileProcessor(algorithm, baseDir string) *FileProcessor {
	return &FileProcessor{
		algorithm: algorithm,
		baseDir:   baseDir,
	}
}

// ProcessFile calculates checksum for a single file
func (fp *FileProcessor) ProcessFile(filePath string) (*FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	// Calculate checksum
	hash, err := fp.createHash()
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Get relative path
	relPath, err := filepath.Rel(fp.baseDir, filePath)
	if err != nil {
		relPath = filePath
	}

	return &FileInfo{
		Path:      relPath,
		Name:      filepath.Base(filePath),
		Size:      info.Size(),
		Checksum:  fmt.Sprintf("%x", hash.Sum(nil)),
		Algorithm: strings.ToUpper(fp.algorithm),
		Modified:  info.ModTime(),
	}, nil
}

// ProcessDirectory processes all files in a directory
func (fp *FileProcessor) ProcessDirectory(dirPath string) ([]*FileInfo, error) {
	var results []*FileInfo

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Process file
		fileInfo, err := fp.ProcessFile(path)
		if err != nil {
			log.Printf("Error processing file %s: %v", path, err)
			return nil // Continue with other files
		}

		results = append(results, fileInfo)
		return nil
	})

	return results, err
}

// createHash creates the appropriate hash.Hash based on the algorithm
func (fp *FileProcessor) createHash() (hash.Hash, error) {
	switch strings.ToLower(fp.algorithm) {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s. Supported: md5, sha1, sha256, sha512", fp.algorithm)
	}
}

// WebServer provides HTTP API for file processing
type WebServer struct {
	processor *FileProcessor
	port      string
}

// NewWebServer creates a new web server
func NewWebServer(processor *FileProcessor, port string) *WebServer {
	return &WebServer{
		processor: processor,
		port:      port,
	}
}

// handleHealth provides health check endpoint
func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "file-processor",
	})
}

// handleProcessFiles processes files and returns results
func (ws *WebServer) handleProcessFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Directory string `json:"directory"`
		Algorithm string `json:"algorithm,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use provided algorithm or default
	algorithm := request.Algorithm
	if algorithm == "" {
		algorithm = ws.processor.algorithm
	}

	// Create processor with specified algorithm
	processor := NewFileProcessor(algorithm, request.Directory)

	// Process files
	results, err := processor.ProcessDirectory(request.Directory)
	if err != nil {
		http.Error(w, fmt.Sprintf("Processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files":     results,
		"count":     len(results),
		"algorithm": strings.ToUpper(algorithm),
		"directory": request.Directory,
	})
}

// handleListFiles lists files in a directory
func (ws *WebServer) handleListFiles(w http.ResponseWriter, r *http.Request) {
	directory := r.URL.Query().Get("directory")
	if directory == "" {
		http.Error(w, "directory parameter required", http.StatusBadRequest)
		return
	}

	var files []map[string]interface{}
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(directory, path)
			files = append(files, map[string]interface{}{
				"name":     info.Name(),
				"path":     relPath,
				"size":     info.Size(),
				"modified": info.ModTime().Format(time.RFC3339),
			})
		}
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list files: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files":     files,
		"count":     len(files),
		"directory": directory,
	})
}

// Start starts the web server
func (ws *WebServer) Start() error {
	http.HandleFunc("/health", ws.handleHealth)
	http.HandleFunc("/api/process", ws.handleProcessFiles)
	http.HandleFunc("/api/files", ws.handleListFiles)

	log.Printf("Starting file processor server on port %s", ws.port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health - Health check")
	log.Printf("  POST /api/process - Process files in directory")
	log.Printf("  GET  /api/files?directory=<path> - List files in directory")

	return http.ListenAndServe(":"+ws.port, nil)
}

func main() {
	var (
		port      = flag.String("port", "8080", "Port to listen on")
		algorithm = flag.String("algo", "sha256", "Default checksum algorithm")
		baseDir   = flag.String("dir", "/app/decrypted", "Base directory for file processing")
		help      = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("File Processor Application")
		fmt.Println("==========================")
		fmt.Println()
		fmt.Println("This application processes files and calculates checksums.")
		fmt.Println("It's designed to work with decrypted files from secure_packager.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	// Check if base directory exists
	if _, err := os.Stat(*baseDir); os.IsNotExist(err) {
		log.Printf("Warning: Base directory %s does not exist", *baseDir)
		log.Printf("This is expected if decryption hasn't been performed yet")
	}

	// Create processor
	processor := NewFileProcessor(*algorithm, *baseDir)

	// Create and start web server
	server := NewWebServer(processor, *port)

	log.Println("File Processor Application Started")
	log.Printf("Base directory: %s", *baseDir)
	log.Printf("Default algorithm: %s", strings.ToUpper(*algorithm))

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
