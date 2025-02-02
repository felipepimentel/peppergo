package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestFileReaderTool(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "file_reader_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"test.txt":  "Hello, World!",
		"test.md":   "# Test Markdown",
		"test.json": `{"key": "value"}`,
		"test.bin":  "binary data",
	}

	for name, content := range testFiles {
		path := filepath.Join(tempDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)
	}

	config := &Config{
		BasePath: tempDir,
		AllowedExtensions: []string{
			".txt",
			".md",
			".json",
		},
		MaxFileSize: 1024,
	}

	t.Run("basic functionality", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		assert.NotNil(t, tool)
		assert.Equal(t, "file_reader", tool.Name())
		assert.Equal(t, "1.0.0", tool.Version())
	})

	t.Run("initialization", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Initialize(ctx)
		assert.NoError(t, err)
	})

	t.Run("initialization with invalid base path", func(t *testing.T) {
		invalidConfig := &Config{
			BasePath: "/nonexistent/path",
		}
		tool := NewFileReaderTool(logger, invalidConfig)
		err := tool.Initialize(ctx)
		assert.Error(t, err)
	})

	t.Run("read valid file", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Initialize(ctx)
		assert.NoError(t, err)

		result, err := tool.Execute(ctx, map[string]interface{}{
			"path": "test.txt",
		})
		assert.NoError(t, err)

		response := result.(map[string]interface{})
		assert.Equal(t, "Hello, World!", response["content"])
		assert.Equal(t, "test.txt", response["path"])
		assert.Equal(t, int64(13), response["size"])
	})

	t.Run("read file with disallowed extension", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Initialize(ctx)
		assert.NoError(t, err)

		result, err := tool.Execute(ctx, map[string]interface{}{
			"path": "test.bin",
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "path is not allowed")
	})

	t.Run("read nonexistent file", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Initialize(ctx)
		assert.NoError(t, err)

		result, err := tool.Execute(ctx, map[string]interface{}{
			"path": "nonexistent.txt",
		})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("path traversal attempt", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Initialize(ctx)
		assert.NoError(t, err)

		testCases := []string{
			"../test.txt",
			"../../test.txt",
			"/etc/passwd",
			"test.txt/../../../etc/passwd",
		}

		for _, path := range testCases {
			result, err := tool.Execute(ctx, map[string]interface{}{
				"path": path,
			})
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "path is not allowed")
		}
	})

	t.Run("missing path argument", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Initialize(ctx)
		assert.NoError(t, err)

		result, err := tool.Execute(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "path argument is required")
	})

	t.Run("invalid path argument type", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Initialize(ctx)
		assert.NoError(t, err)

		result, err := tool.Execute(ctx, map[string]interface{}{
			"path": 123,
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "path must be a string")
	})

	t.Run("schema validation", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		schema := tool.Schema()

		assert.NotNil(t, schema)
		assert.Equal(t, "object", schema.Type)
		assert.Contains(t, schema.Properties, "path")
		assert.Equal(t, "string", schema.Properties["path"].Type)
		assert.Contains(t, schema.Required, "path")
	})

	t.Run("cleanup", func(t *testing.T) {
		tool := NewFileReaderTool(logger, config)
		err := tool.Cleanup(ctx)
		assert.NoError(t, err)
	})
}

func TestFileReaderToolConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := &Config{
			BasePath:    "/tmp",
			MaxFileSize: 1024,
		}

		assert.Empty(t, config.AllowedExtensions)
	})

	t.Run("config with allowed extensions", func(t *testing.T) {
		config := &Config{
			BasePath: "/tmp",
			AllowedExtensions: []string{
				".txt",
				".md",
			},
			MaxFileSize: 1024,
		}

		assert.Len(t, config.AllowedExtensions, 2)
		assert.Contains(t, config.AllowedExtensions, ".txt")
		assert.Contains(t, config.AllowedExtensions, ".md")
	})
} 