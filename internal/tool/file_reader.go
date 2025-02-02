package tool

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/yourusername/peppergo/pkg/types"
)

// FileReaderTool provides file reading functionality
type FileReaderTool struct {
	logger *zap.Logger
	config *Config
}

// Config represents the configuration for FileReaderTool
type Config struct {
	// BasePath is the base path for file operations
	BasePath string `yaml:"base_path"`

	// AllowedExtensions lists allowed file extensions
	AllowedExtensions []string `yaml:"allowed_extensions"`

	// MaxFileSize is the maximum file size in bytes
	MaxFileSize int64 `yaml:"max_file_size"`
}

// NewFileReaderTool creates a new FileReaderTool instance
func NewFileReaderTool(logger *zap.Logger, config *Config) *FileReaderTool {
	return &FileReaderTool{
		logger: logger,
		config: config,
	}
}

// Name returns the tool's name
func (t *FileReaderTool) Name() string {
	return "file_reader"
}

// Description returns the tool's description
func (t *FileReaderTool) Description() string {
	return "Reads file contents with safety checks"
}

// Initialize initializes the tool
func (t *FileReaderTool) Initialize(ctx context.Context) error {
	// Validate base path
	if t.config.BasePath == "" {
		return fmt.Errorf("base path is required")
	}

	// Ensure base path exists
	if _, err := os.Stat(t.config.BasePath); os.IsNotExist(err) {
		return fmt.Errorf("base path does not exist: %w", err)
	}

	t.logger.Info("Initializing file reader tool",
		zap.String("base_path", t.config.BasePath),
		zap.Strings("allowed_extensions", t.config.AllowedExtensions),
		zap.Int64("max_file_size", t.config.MaxFileSize))

	return nil
}

// Execute runs the tool
func (t *FileReaderTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Get path argument
	pathRaw, ok := args["path"]
	if !ok {
		return nil, fmt.Errorf("path argument is required")
	}

	path, ok := pathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	// Clean and validate path
	fullPath := filepath.Clean(filepath.Join(t.config.BasePath, path))
	if !t.isPathAllowed(fullPath) {
		return nil, fmt.Errorf("path is not allowed: %s", path)
	}

	// Check file size
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > t.config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	// Read file
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	t.logger.Debug("Read file",
		zap.String("path", path),
		zap.Int64("size", info.Size()))

	return map[string]interface{}{
		"content": string(data),
		"size":    info.Size(),
		"path":    path,
	}, nil
}

// Cleanup performs cleanup
func (t *FileReaderTool) Cleanup(ctx context.Context) error {
	return nil
}

// Schema returns the tool's schema
func (t *FileReaderTool) Schema() *types.ToolSchema {
	schema := types.NewToolSchema()
	schema.AddProperty("path", &types.PropertySchema{
		Type:        "string",
		Description: "Path to the file to read, relative to base path",
	})
	schema.AddRequired("path")
	return schema
}

// Version returns the tool version
func (t *FileReaderTool) Version() string {
	return "1.0.0"
}

// isPathAllowed checks if the path is allowed
func (t *FileReaderTool) isPathAllowed(path string) bool {
	// Check if path is under base path
	rel, err := filepath.Rel(t.config.BasePath, path)
	if err != nil || rel == ".." || filepath.IsAbs(rel) {
		return false
	}

	// Check extension if allowed extensions are specified
	if len(t.config.AllowedExtensions) > 0 {
		ext := filepath.Ext(path)
		allowed := false
		for _, allowedExt := range t.config.AllowedExtensions {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	return true
}

// Example YAML configuration:
/*
name: file_reader
version: "1.0.0"
description: "File reader tool"

config:
  base_path: "/path/to/files"
  allowed_extensions:
    - ".txt"
    - ".md"
    - ".json"
  max_file_size: 1048576  # 1MB
*/ 