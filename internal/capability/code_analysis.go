package capability

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/yourusername/peppergo/pkg/types"
)

// CodeAnalysisCapability provides code analysis functionality
type CodeAnalysisCapability struct {
	logger *zap.Logger
	config *CodeAnalysisConfig
}

// CodeAnalysisConfig represents the configuration for CodeAnalysisCapability
type CodeAnalysisConfig struct {
	// Rules defines which analysis rules to apply
	Rules []string `yaml:"rules"`

	// ExcludePatterns defines patterns to exclude from analysis
	ExcludePatterns []string `yaml:"exclude_patterns"`

	// MaxComplexity is the maximum allowed cyclomatic complexity
	MaxComplexity int `yaml:"max_complexity"`

	// MinTestCoverage is the minimum required test coverage percentage
	MinTestCoverage float64 `yaml:"min_test_coverage"`

	// RequireDocumentation indicates if documentation is required
	RequireDocumentation bool `yaml:"require_documentation"`
}

// AnalysisResult represents the result of code analysis
type AnalysisResult struct {
	Issues []Issue     `json:"issues"`
	Stats  Statistics  `json:"stats"`
}

// Issue represents a code issue found during analysis
type Issue struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Rule        string `json:"rule"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// Statistics represents code analysis statistics
type Statistics struct {
	FilesAnalyzed  int     `json:"files_analyzed"`
	IssuesFound    int     `json:"issues_found"`
	TestCoverage   float64 `json:"test_coverage"`
	AverageComplexity float64 `json:"average_complexity"`
}

// NewCodeAnalysisCapability creates a new CodeAnalysisCapability instance
func NewCodeAnalysisCapability(logger *zap.Logger, config *CodeAnalysisConfig) *CodeAnalysisCapability {
	return &CodeAnalysisCapability{
		logger: logger,
		config: config,
	}
}

// Name returns the capability's name
func (c *CodeAnalysisCapability) Name() string {
	return "code_analysis"
}

// Description returns the capability's description
func (c *CodeAnalysisCapability) Description() string {
	return "Analyzes Go code for best practices and potential issues"
}

// Initialize initializes the capability
func (c *CodeAnalysisCapability) Initialize(ctx context.Context) error {
	c.logger.Info("Initializing code analysis capability",
		zap.Strings("rules", c.config.Rules),
		zap.Int("max_complexity", c.config.MaxComplexity))
	return nil
}

// Execute runs the capability
func (c *CodeAnalysisCapability) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	files, ok := input.([]string)
	if !ok {
		return nil, fmt.Errorf("input must be a slice of file paths")
	}

	result := &AnalysisResult{
		Issues: make([]Issue, 0),
		Stats: Statistics{
			FilesAnalyzed: len(files),
		},
	}

	fset := token.NewFileSet()
	for _, file := range files {
		// Skip excluded files
		if c.shouldExclude(file) {
			continue
		}

		// Parse file
		node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", file, err)
		}

		// Analyze file
		issues := c.analyzeFile(fset, node)
		result.Issues = append(result.Issues, issues...)
		result.Stats.IssuesFound += len(issues)
	}

	return result, nil
}

// Cleanup performs cleanup
func (c *CodeAnalysisCapability) Cleanup(ctx context.Context) error {
	return nil
}

// Requirements returns capability requirements
func (c *CodeAnalysisCapability) Requirements() *types.Requirements {
	reqs := types.NewRequirements()
	reqs.AddTool("file_reader")
	return reqs
}

// Version returns the capability version
func (c *CodeAnalysisCapability) Version() string {
	return "1.0.0"
}

// shouldExclude checks if a file should be excluded from analysis
func (c *CodeAnalysisCapability) shouldExclude(file string) bool {
	for _, pattern := range c.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
	}
	return false
}

// analyzeFile performs analysis on a single file
func (c *CodeAnalysisCapability) analyzeFile(fset *token.FileSet, node *ast.File) []Issue {
	var issues []Issue

	// Check documentation if required
	if c.config.RequireDocumentation {
		issues = append(issues, c.checkDocumentation(fset, node)...)
	}

	// Check function complexity
	issues = append(issues, c.checkComplexity(fset, node)...)

	// Check naming conventions
	issues = append(issues, c.checkNaming(fset, node)...)

	// Check error handling
	issues = append(issues, c.checkErrorHandling(fset, node)...)

	return issues
}

// checkDocumentation checks for proper documentation
func (c *CodeAnalysisCapability) checkDocumentation(fset *token.FileSet, node *ast.File) []Issue {
	var issues []Issue

	// Check package documentation
	if node.Doc == nil {
		issues = append(issues, Issue{
			File:     fset.Position(node.Package).Filename,
			Line:     fset.Position(node.Package).Line,
			Rule:     "missing-package-doc",
			Severity: "warning",
			Message:  "package is missing documentation",
		})
	}

	// Check exported declarations
	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			if ast.IsExported(decl.Name.Name) && decl.Doc == nil {
				issues = append(issues, Issue{
					File:     fset.Position(decl.Pos()).Filename,
					Line:     fset.Position(decl.Pos()).Line,
					Rule:     "missing-func-doc",
					Severity: "warning",
					Message:  fmt.Sprintf("exported function %s is missing documentation", decl.Name.Name),
				})
			}
		}
		return true
	})

	return issues
}

// checkComplexity checks function complexity
func (c *CodeAnalysisCapability) checkComplexity(fset *token.FileSet, node *ast.File) []Issue {
	var issues []Issue

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			complexity := calculateComplexity(fn)
			if complexity > c.config.MaxComplexity {
				issues = append(issues, Issue{
					File:     fset.Position(fn.Pos()).Filename,
					Line:     fset.Position(fn.Pos()).Line,
					Rule:     "high-complexity",
					Severity: "warning",
					Message:  fmt.Sprintf("function %s has cyclomatic complexity of %d (max %d)", fn.Name.Name, complexity, c.config.MaxComplexity),
					Suggestion: "Consider breaking down the function into smaller functions",
				})
			}
		}
		return true
	})

	return issues
}

// checkNaming checks naming conventions
func (c *CodeAnalysisCapability) checkNaming(fset *token.FileSet, node *ast.File) []Issue {
	var issues []Issue

	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			if !isValidFuncName(decl.Name.Name) {
				issues = append(issues, Issue{
					File:     fset.Position(decl.Pos()).Filename,
					Line:     fset.Position(decl.Pos()).Line,
					Rule:     "invalid-name",
					Severity: "warning",
					Message:  fmt.Sprintf("function name %s doesn't follow Go naming conventions", decl.Name.Name),
					Suggestion: "Use MixedCaps or mixedCaps naming style",
				})
			}
		}
		return true
	})

	return issues
}

// checkErrorHandling checks error handling patterns
func (c *CodeAnalysisCapability) checkErrorHandling(fset *token.FileSet, node *ast.File) []Issue {
	var issues []Issue

	ast.Inspect(node, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, expr := range assign.Rhs {
				if call, ok := expr.(*ast.CallExpr); ok {
					if isErrorReturningFunc(call) {
						if !hasErrorCheck(assign) {
							issues = append(issues, Issue{
								File:     fset.Position(assign.Pos()).Filename,
								Line:     fset.Position(assign.Pos()).Line,
								Rule:     "unchecked-error",
								Severity: "error",
								Message:  "error is not checked",
								Suggestion: "Add error handling code",
							})
						}
					}
				}
			}
		}
		return true
	})

	return issues
}

// Helper functions

func calculateComplexity(fn *ast.FuncDecl) int {
	complexity := 1
	ast.Inspect(fn, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause, *ast.CommClause:
			complexity++
		}
		return true
	})
	return complexity
}

func isValidFuncName(name string) bool {
	return !strings.Contains(name, "_") && name != strings.ToLower(name)
}

func isErrorReturningFunc(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return strings.HasSuffix(ident.Name, "Error") || strings.HasPrefix(ident.Name, "Error")
	}
	return false
}

func hasErrorCheck(assign *ast.AssignStmt) bool {
	if len(assign.Lhs) != 2 {
		return false
	}
	if ident, ok := assign.Lhs[1].(*ast.Ident); ok {
		return ident.Name == "err"
	}
	return false
}

// Example YAML configuration:
/*
name: code_analysis
version: "1.0.0"
description: "Code analysis capability"

config:
  rules:
    - complexity
    - documentation
    - naming
    - error-handling
  exclude_patterns:
    - "*_test.go"
    - "vendor/*"
  max_complexity: 15
  min_test_coverage: 80.0
  require_documentation: true
*/ 