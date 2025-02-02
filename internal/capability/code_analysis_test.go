package capability

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestCodeAnalysisCapability(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	config := &CodeAnalysisConfig{
		Rules: []string{
			"complexity",
			"documentation",
			"naming",
			"error-handling",
		},
		ExcludePatterns: []string{
			"*_test.go",
			"vendor/*",
		},
		MaxComplexity:        15,
		MinTestCoverage:      80.0,
		RequireDocumentation: true,
	}

	t.Run("basic functionality", func(t *testing.T) {
		cap := NewCodeAnalysisCapability(logger, config)
		assert.NotNil(t, cap)
		assert.Equal(t, "code_analysis", cap.Name())
		assert.Equal(t, "1.0.0", cap.Version())
	})

	t.Run("initialization", func(t *testing.T) {
		cap := NewCodeAnalysisCapability(logger, config)
		err := cap.Initialize(ctx)
		assert.NoError(t, err)
	})

	t.Run("analyze valid file", func(t *testing.T) {
		// Create temporary test file
		tmpDir, err := os.MkdirTemp("", "code_analysis_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		testFile := filepath.Join(tmpDir, "test.go")
		code := `package test

// TestFunc is a test function
func TestFunc() error {
	// Missing error check
	os.Open("test.txt")
	return nil
}

func test_invalid_name() {}
`
		err = os.WriteFile(testFile, []byte(code), 0644)
		assert.NoError(t, err)

		cap := NewCodeAnalysisCapability(logger, config)
		err = cap.Initialize(ctx)
		assert.NoError(t, err)

		result, err := cap.Execute(ctx, []string{testFile})
		assert.NoError(t, err)

		analysisResult, ok := result.(*AnalysisResult)
		assert.True(t, ok)
		assert.NotNil(t, analysisResult)

		// Should find at least one error handling issue and one naming issue
		assert.Greater(t, len(analysisResult.Issues), 1)
		
		var foundErrorIssue, foundNamingIssue bool
		for _, issue := range analysisResult.Issues {
			switch issue.Rule {
			case "unchecked-error":
				foundErrorIssue = true
			case "invalid-name":
				foundNamingIssue = true
			}
		}

		assert.True(t, foundErrorIssue, "should find unchecked error")
		assert.True(t, foundNamingIssue, "should find invalid naming")
	})

	t.Run("analyze complex function", func(t *testing.T) {
		// Create temporary test file
		tmpDir, err := os.MkdirTemp("", "code_analysis_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		testFile := filepath.Join(tmpDir, "complex.go")
		code := `package test

func ComplexFunction(x int) int {
	result := 0
	if x > 0 {
		if x > 10 {
			for i := 0; i < x; i++ {
				if i%2 == 0 {
					result += i
				} else {
					result -= i
				}
				switch {
				case i < 5:
					result *= 2
				case i < 10:
					result *= 3
				default:
					result *= 4
				}
			}
		} else {
			result = x * 2
		}
	}
	return result
}
`
		err = os.WriteFile(testFile, []byte(code), 0644)
		assert.NoError(t, err)

		cap := NewCodeAnalysisCapability(logger, config)
		err = cap.Initialize(ctx)
		assert.NoError(t, err)

		result, err := cap.Execute(ctx, []string{testFile})
		assert.NoError(t, err)

		analysisResult, ok := result.(*AnalysisResult)
		assert.True(t, ok)

		var foundComplexityIssue bool
		for _, issue := range analysisResult.Issues {
			if issue.Rule == "high-complexity" {
				foundComplexityIssue = true
				break
			}
		}

		assert.True(t, foundComplexityIssue, "should find high complexity issue")
	})

	t.Run("analyze missing documentation", func(t *testing.T) {
		// Create temporary test file
		tmpDir, err := os.MkdirTemp("", "code_analysis_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		testFile := filepath.Join(tmpDir, "nodoc.go")
		code := `package test

// Documented is a documented function
func Documented() {}

func Undocumented() {}
`
		err = os.WriteFile(testFile, []byte(code), 0644)
		assert.NoError(t, err)

		cap := NewCodeAnalysisCapability(logger, config)
		err = cap.Initialize(ctx)
		assert.NoError(t, err)

		result, err := cap.Execute(ctx, []string{testFile})
		assert.NoError(t, err)

		analysisResult, ok := result.(*AnalysisResult)
		assert.True(t, ok)

		var foundDocIssue bool
		for _, issue := range analysisResult.Issues {
			if issue.Rule == "missing-func-doc" {
				foundDocIssue = true
				break
			}
		}

		assert.True(t, foundDocIssue, "should find missing documentation issue")
	})

	t.Run("exclude patterns", func(t *testing.T) {
		// Create temporary test files
		tmpDir, err := os.MkdirTemp("", "code_analysis_test")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create a test file and a test file that should be excluded
		testFile := filepath.Join(tmpDir, "test.go")
		testTestFile := filepath.Join(tmpDir, "test_test.go")

		code := `package test
func Test() {}`

		err = os.WriteFile(testFile, []byte(code), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(testTestFile, []byte(code), 0644)
		assert.NoError(t, err)

		cap := NewCodeAnalysisCapability(logger, config)
		err = cap.Initialize(ctx)
		assert.NoError(t, err)

		result, err := cap.Execute(ctx, []string{testFile, testTestFile})
		assert.NoError(t, err)

		analysisResult, ok := result.(*AnalysisResult)
		assert.True(t, ok)

		// Should only analyze one file
		assert.Equal(t, 1, analysisResult.Stats.FilesAnalyzed)
	})

	t.Run("invalid input", func(t *testing.T) {
		cap := NewCodeAnalysisCapability(logger, config)
		err := cap.Initialize(ctx)
		assert.NoError(t, err)

		result, err := cap.Execute(ctx, "invalid input")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("cleanup", func(t *testing.T) {
		cap := NewCodeAnalysisCapability(logger, config)
		err := cap.Cleanup(ctx)
		assert.NoError(t, err)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("calculateComplexity", func(t *testing.T) {
		code := `
package test

func SimpleFunction() {
	x := 1
	if x > 0 {
		x++
	}
}
`
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
		assert.NoError(t, err)

		var fn *ast.FuncDecl
		for _, decl := range node.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				fn = f
				break
			}
		}

		complexity := calculateComplexity(fn)
		assert.Equal(t, 2, complexity) // Base complexity (1) + if statement (1)
	})

	t.Run("isValidFuncName", func(t *testing.T) {
		tests := []struct {
			name     string
			funcName string
			want     bool
		}{
			{
				name:     "valid camel case",
				funcName: "ValidFunction",
				want:     true,
			},
			{
				name:     "valid lower camel case",
				funcName: "validFunction",
				want:     true,
			},
			{
				name:     "invalid snake case",
				funcName: "invalid_function",
				want:     false,
			},
			{
				name:     "invalid all lowercase",
				funcName: "invalidfunction",
				want:     false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := isValidFuncName(tt.funcName)
				assert.Equal(t, tt.want, got)
			})
		}
	})
} 