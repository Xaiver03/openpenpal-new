// Package ai provides intelligent test generation capabilities using static code analysis and ML
package ai

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"openpenpal-backend/internal/platform/testing/core"
)

// GoCodeAnalyzer analyzes Go source code to extract testable units and patterns
type GoCodeAnalyzer struct {
	fileSet  *token.FileSet
	packages map[string]*ast.Package
	
	// Analysis results
	functions     []*FunctionInfo
	types         []*TypeInfo
	interfaces    []*InterfaceInfo
	dependencies  []*DependencyInfo
	patterns      []*PatternInfo
	riskAreas     []*RiskAreaInfo
	
	// Configuration
	config *AnalyzerConfig
}

// AnalyzerConfig configures the static analysis behavior
type AnalyzerConfig struct {
	MaxComplexity       int      `json:"max_complexity"`
	IgnorePatterns     []string `json:"ignore_patterns"`
	FocusPatterns      []string `json:"focus_patterns"`
	EnableDeepAnalysis bool     `json:"enable_deep_analysis"`
	AnalyzeTestFiles   bool     `json:"analyze_test_files"`
}

// FunctionInfo represents analyzed information about a function
type FunctionInfo struct {
	Name           string            `json:"name"`
	PackageName    string            `json:"package_name"`
	FilePath       string            `json:"file_path"`
	StartLine      int               `json:"start_line"`
	EndLine        int               `json:"end_line"`
	Parameters     []*ParameterInfo  `json:"parameters"`
	ReturnTypes    []string          `json:"return_types"`
	Complexity     int               `json:"complexity"`
	IsExported     bool              `json:"is_exported"`
	IsMethod       bool              `json:"is_method"`
	ReceiverType   string            `json:"receiver_type,omitempty"`
	CallsExternal  bool              `json:"calls_external"`
	HasErrorReturn bool              `json:"has_error_return"`
	Dependencies   []string          `json:"dependencies"`
	Patterns       []string          `json:"patterns"`
	RiskLevel      core.RiskLevel    `json:"risk_level"`
	TestPriority   core.TestPriority `json:"test_priority"`
}

// ParameterInfo represents function parameter information
type ParameterInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsPointer bool  `json:"is_pointer"`
	IsSlice   bool  `json:"is_slice"`
	IsMap     bool  `json:"is_map"`
}

// TypeInfo represents analyzed information about a type
type TypeInfo struct {
	Name        string   `json:"name"`
	PackageName string   `json:"package_name"`
	FilePath    string   `json:"file_path"`
	Kind        string   `json:"kind"` // struct, interface, alias, etc.
	Fields      []string `json:"fields"`
	Methods     []string `json:"methods"`
	IsExported  bool     `json:"is_exported"`
	Complexity  int      `json:"complexity"`
}

// InterfaceInfo represents analyzed information about an interface
type InterfaceInfo struct {
	Name        string   `json:"name"`
	PackageName string   `json:"package_name"`
	FilePath    string   `json:"file_path"`
	Methods     []string `json:"methods"`
	IsExported  bool     `json:"is_exported"`
	Embeddings  []string `json:"embeddings"`
}

// DependencyInfo represents package dependencies
type DependencyInfo struct {
	PackageName string   `json:"package_name"`
	ImportPath  string   `json:"import_path"`
	IsStandard  bool     `json:"is_standard"`
	IsLocal     bool     `json:"is_local"`
	UsageCount  int      `json:"usage_count"`
	Functions   []string `json:"functions"`
}

// PatternInfo represents identified code patterns
type PatternInfo struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Locations   []string `json:"locations"`
	Frequency   int      `json:"frequency"`
	Confidence  float64  `json:"confidence"`
	TestStrategy string  `json:"test_strategy"`
}

// RiskAreaInfo represents identified risk areas in code
type RiskAreaInfo struct {
	Type        string          `json:"type"`
	Location    string          `json:"location"`
	Severity    core.RiskLevel  `json:"severity"`
	Description string          `json:"description"`
	Patterns    []string        `json:"patterns"`
	Mitigation  string          `json:"mitigation"`
}

// NewGoCodeAnalyzer creates a new Go code analyzer
func NewGoCodeAnalyzer(config *AnalyzerConfig) *GoCodeAnalyzer {
	if config == nil {
		config = &AnalyzerConfig{
			MaxComplexity:       15,
			EnableDeepAnalysis:  true,
			AnalyzeTestFiles:    false,
			IgnorePatterns:     []string{"_test.go", "vendor/", ".git/"},
			FocusPatterns:      []string{"*.go"},
		}
	}

	return &GoCodeAnalyzer{
		fileSet:      token.NewFileSet(),
		packages:     make(map[string]*ast.Package),
		functions:    make([]*FunctionInfo, 0),
		types:        make([]*TypeInfo, 0),
		interfaces:   make([]*InterfaceInfo, 0),
		dependencies: make([]*DependencyInfo, 0),
		patterns:     make([]*PatternInfo, 0),
		riskAreas:    make([]*RiskAreaInfo, 0),
		config:       config,
	}
}

// AnalyzeCodebase analyzes a Go codebase and extracts testable units
func (a *GoCodeAnalyzer) AnalyzeCodebase(ctx context.Context, codebase *core.Codebase) (*core.CodeAnalysis, error) {
	log.Printf("🔍 Starting Go codebase analysis for: %s", codebase.Path)
	
	// Parse the codebase
	if err := a.parseCodebase(codebase.Path); err != nil {
		return nil, fmt.Errorf("failed to parse codebase: %w", err)
	}
	
	// Extract information from AST
	if err := a.extractInformation(); err != nil {
		return nil, fmt.Errorf("failed to extract information: %w", err)
	}
	
	// Identify patterns
	if err := a.identifyPatterns(); err != nil {
		return nil, fmt.Errorf("failed to identify patterns: %w", err)
	}
	
	// Assess risks
	if err := a.assessRisks(); err != nil {
		return nil, fmt.Errorf("failed to assess risks: %w", err)
	}
	
	// Convert to core format
	analysis := a.convertToCodeAnalysis(codebase)
	
	log.Printf("✅ Analysis completed: %d testable units, %d patterns, %d risk areas",
		len(analysis.TestableUnits), len(analysis.Patterns), len(analysis.RiskAreas))
	
	return analysis, nil
}

// parseCodebase parses all Go files in the codebase
func (a *GoCodeAnalyzer) parseCodebase(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip if not a Go file
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		// Skip test files if configured
		if !a.config.AnalyzeTestFiles && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		// Skip ignored patterns
		for _, pattern := range a.config.IgnorePatterns {
			if strings.Contains(path, pattern) {
				return nil
			}
		}
		
		// Parse the file
		src, err := os.ReadFile(path)
		if err != nil {
			log.Printf("⚠️  Failed to read file %s: %v", path, err)
			return nil // Continue with other files
		}
		
		file, err := parser.ParseFile(a.fileSet, path, src, parser.ParseComments)
		if err != nil {
			log.Printf("⚠️  Failed to parse file %s: %v", path, err)
			return nil // Continue with other files
		}
		
		// Add to packages
		pkgName := file.Name.Name
		if a.packages[pkgName] == nil {
			a.packages[pkgName] = &ast.Package{
				Name:  pkgName,
				Files: make(map[string]*ast.File),
			}
		}
		a.packages[pkgName].Files[path] = file
		
		return nil
	})
}

// extractInformation extracts functions, types, and other information from AST
func (a *GoCodeAnalyzer) extractInformation() error {
	for pkgName, pkg := range a.packages {
		for filePath, file := range pkg.Files {
			// Extract functions
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					funcInfo := a.analyzeFunctionDecl(d, pkgName, filePath)
					a.functions = append(a.functions, funcInfo)
					
				case *ast.GenDecl:
					// Extract types and interfaces
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							switch s.Type.(type) {
							case *ast.InterfaceType:
								interfaceInfo := a.analyzeInterfaceDecl(s, pkgName, filePath)
								a.interfaces = append(a.interfaces, interfaceInfo)
							case *ast.StructType:
								typeInfo := a.analyzeTypeDecl(s, pkgName, filePath)
								a.types = append(a.types, typeInfo)
							}
						}
					}
				}
			}
			
			// Extract dependencies
			for _, imp := range file.Imports {
				depInfo := a.analyzeDependency(imp, pkgName)
				if depInfo != nil {
					a.dependencies = append(a.dependencies, depInfo)
				}
			}
		}
	}
	
	return nil
}

// analyzeFunctionDecl analyzes a function declaration
func (a *GoCodeAnalyzer) analyzeFunctionDecl(decl *ast.FuncDecl, pkgName, filePath string) *FunctionInfo {
	info := &FunctionInfo{
		Name:         decl.Name.Name,
		PackageName:  pkgName,
		FilePath:     filePath,
		StartLine:    a.fileSet.Position(decl.Pos()).Line,
		EndLine:      a.fileSet.Position(decl.End()).Line,
		IsExported:   ast.IsExported(decl.Name.Name),
		Parameters:   make([]*ParameterInfo, 0),
		ReturnTypes:  make([]string, 0),
		Dependencies: make([]string, 0),
		Patterns:     make([]string, 0),
	}
	
	// Check if it's a method
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		info.IsMethod = true
		info.ReceiverType = a.extractTypeName(decl.Recv.List[0].Type)
	}
	
	// Analyze parameters
	if decl.Type.Params != nil {
		for _, param := range decl.Type.Params.List {
			paramInfo := &ParameterInfo{
				Type: a.extractTypeName(param.Type),
			}
			
			// Extract parameter names
			for _, name := range param.Names {
				paramInfo.Name = name.Name
				break // Take first name for simplicity
			}
			
			// Analyze type characteristics
			switch param.Type.(type) {
			case *ast.StarExpr:
				paramInfo.IsPointer = true
			case *ast.ArrayType, *ast.SliceExpr:
				paramInfo.IsSlice = true
			case *ast.MapType:
				paramInfo.IsMap = true
			}
			
			info.Parameters = append(info.Parameters, paramInfo)
		}
	}
	
	// Analyze return types
	if decl.Type.Results != nil {
		for _, result := range decl.Type.Results.List {
			returnType := a.extractTypeName(result.Type)
			info.ReturnTypes = append(info.ReturnTypes, returnType)
			
			// Check for error return
			if returnType == "error" {
				info.HasErrorReturn = true
			}
		}
	}
	
	// Calculate complexity and analyze body
	if decl.Body != nil {
		info.Complexity = a.calculateComplexity(decl.Body)
		info.CallsExternal = a.hasExternalCalls(decl.Body)
		info.Dependencies = a.extractDependencies(decl.Body)
		info.Patterns = a.identifyFunctionPatterns(decl)
	}
	
	// Determine risk level and test priority
	info.RiskLevel = a.calculateRiskLevel(info)
	info.TestPriority = a.calculateTestPriority(info)
	
	return info
}

// analyzeTypeDecl analyzes a type declaration
func (a *GoCodeAnalyzer) analyzeTypeDecl(spec *ast.TypeSpec, pkgName, filePath string) *TypeInfo {
	info := &TypeInfo{
		Name:        spec.Name.Name,
		PackageName: pkgName,
		FilePath:    filePath,
		IsExported:  ast.IsExported(spec.Name.Name),
		Fields:      make([]string, 0),
		Methods:     make([]string, 0),
	}
	
	// Analyze struct fields
	if structType, ok := spec.Type.(*ast.StructType); ok {
		info.Kind = "struct"
		if structType.Fields != nil {
			for _, field := range structType.Fields.List {
				fieldType := a.extractTypeName(field.Type)
				
				if len(field.Names) > 0 {
					for _, name := range field.Names {
						info.Fields = append(info.Fields, fmt.Sprintf("%s %s", name.Name, fieldType))
					}
				} else {
					// Embedded field
					info.Fields = append(info.Fields, fieldType)
				}
			}
		}
		info.Complexity = len(info.Fields)
	}
	
	return info
}

// analyzeInterfaceDecl analyzes an interface declaration
func (a *GoCodeAnalyzer) analyzeInterfaceDecl(spec *ast.TypeSpec, pkgName, filePath string) *InterfaceInfo {
	info := &InterfaceInfo{
		Name:        spec.Name.Name,
		PackageName: pkgName,
		FilePath:    filePath,
		IsExported:  ast.IsExported(spec.Name.Name),
		Methods:     make([]string, 0),
		Embeddings:  make([]string, 0),
	}
	
	if interfaceType, ok := spec.Type.(*ast.InterfaceType); ok {
		if interfaceType.Methods != nil {
			for _, method := range interfaceType.Methods.List {
				if len(method.Names) > 0 {
					// Regular method
					methodName := method.Names[0].Name
					methodType := a.extractTypeName(method.Type)
					info.Methods = append(info.Methods, fmt.Sprintf("%s%s", methodName, methodType))
				} else {
					// Embedded interface
					embeddedType := a.extractTypeName(method.Type)
					info.Embeddings = append(info.Embeddings, embeddedType)
				}
			}
		}
	}
	
	return info
}

// analyzeDependency analyzes an import declaration
func (a *GoCodeAnalyzer) analyzeDependency(imp *ast.ImportSpec, pkgName string) *DependencyInfo {
	if imp.Path == nil {
		return nil
	}
	
	importPath := strings.Trim(imp.Path.Value, "\"")
	
	info := &DependencyInfo{
		PackageName: pkgName,
		ImportPath:  importPath,
		IsStandard:  a.isStandardLibrary(importPath),
		IsLocal:     a.isLocalPackage(importPath),
		UsageCount:  1, // Will be calculated later
		Functions:   make([]string, 0),
	}
	
	return info
}

// Helper methods

func (a *GoCodeAnalyzer) extractTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + a.extractTypeName(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + a.extractTypeName(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", a.extractTypeName(t.Len), a.extractTypeName(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", a.extractTypeName(t.Key), a.extractTypeName(t.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", a.extractTypeName(t.X), t.Sel.Name)
	case *ast.FuncType:
		return "func" // Simplified
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	default:
		return "unknown"
	}
}

func (a *GoCodeAnalyzer) calculateComplexity(body *ast.BlockStmt) int {
	complexity := 1 // Base complexity
	
	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		}
		return true
	})
	
	return complexity
}

func (a *GoCodeAnalyzer) hasExternalCalls(body *ast.BlockStmt) bool {
	hasExternal := false
	
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			// Check if it's a function call with package selector
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok {
					// This is a potential external call (pkg.Function())
					// In a real implementation, we'd check if pkg is imported
					_ = ident.Name
					hasExternal = true
					return false
				}
			}
		}
		return true
	})
	
	return hasExternal
}

func (a *GoCodeAnalyzer) extractDependencies(body *ast.BlockStmt) []string {
	deps := make([]string, 0)
	seen := make(map[string]bool)
	
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok {
					dep := ident.Name
					if !seen[dep] {
						deps = append(deps, dep)
						seen[dep] = true
					}
				}
			}
		}
		return true
	})
	
	return deps
}

func (a *GoCodeAnalyzer) identifyFunctionPatterns(decl *ast.FuncDecl) []string {
	patterns := make([]string, 0)
	
	// Check common patterns
	if strings.HasPrefix(decl.Name.Name, "Test") {
		patterns = append(patterns, "test_function")
	}
	if strings.HasPrefix(decl.Name.Name, "Benchmark") {
		patterns = append(patterns, "benchmark_function")
	}
	if strings.HasPrefix(decl.Name.Name, "Get") || strings.HasPrefix(decl.Name.Name, "Fetch") {
		patterns = append(patterns, "getter_pattern")
	}
	if strings.HasPrefix(decl.Name.Name, "Set") || strings.HasPrefix(decl.Name.Name, "Update") {
		patterns = append(patterns, "setter_pattern")
	}
	if strings.HasPrefix(decl.Name.Name, "Create") || strings.HasPrefix(decl.Name.Name, "New") {
		patterns = append(patterns, "constructor_pattern")
	}
	if strings.HasPrefix(decl.Name.Name, "Validate") || strings.HasPrefix(decl.Name.Name, "Check") {
		patterns = append(patterns, "validation_pattern")
	}
	
	// Check for error handling patterns
	if decl.Type.Results != nil {
		for _, result := range decl.Type.Results.List {
			if a.extractTypeName(result.Type) == "error" {
				patterns = append(patterns, "error_handling")
				break
			}
		}
	}
	
	return patterns
}

func (a *GoCodeAnalyzer) calculateRiskLevel(info *FunctionInfo) core.RiskLevel {
	// High complexity = higher risk
	if info.Complexity > a.config.MaxComplexity {
		return core.RiskLevelHigh
	}
	
	// External calls = medium risk
	if info.CallsExternal {
		return core.RiskLevelMedium
	}
	
	// Error handling functions = medium risk
	if info.HasErrorReturn {
		return core.RiskLevelMedium
	}
	
	// Many parameters = medium risk
	if len(info.Parameters) > 5 {
		return core.RiskLevelMedium
	}
	
	return core.RiskLevelLow
}

func (a *GoCodeAnalyzer) calculateTestPriority(info *FunctionInfo) core.TestPriority {
	// Exported functions get higher priority
	if info.IsExported {
		// High complexity exported functions are critical
		if info.Complexity > a.config.MaxComplexity {
			return core.TestPriorityCritical
		}
		
		// Exported functions with error returns are high priority
		if info.HasErrorReturn {
			return core.TestPriorityHigh
		}
		
		return core.TestPriorityMedium
	}
	
	// Internal functions are lower priority unless complex
	if info.Complexity > a.config.MaxComplexity {
		return core.TestPriorityHigh
	}
	
	return core.TestPriorityLow
}

func (a *GoCodeAnalyzer) isStandardLibrary(importPath string) bool {
	// Standard library packages typically don't have dots or are well-known
	standardPrefixes := []string{
		"bufio", "bytes", "context", "crypto", "database", "encoding", "errors",
		"fmt", "go", "hash", "html", "image", "io", "log", "math", "net", "os",
		"path", "reflect", "regexp", "runtime", "sort", "strconv", "strings",
		"sync", "syscall", "testing", "text", "time", "unicode", "unsafe",
	}
	
	for _, prefix := range standardPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	
	return !strings.Contains(importPath, ".")
}

func (a *GoCodeAnalyzer) isLocalPackage(importPath string) bool {
	// Local packages typically start with the module name or are relative
	return !a.isStandardLibrary(importPath) && !strings.Contains(importPath, "github.com")
}

// identifyPatterns identifies common code patterns across the codebase
func (a *GoCodeAnalyzer) identifyPatterns() error {
	// Analyze function naming patterns
	a.analyzeNamingPatterns()
	
	// Analyze structural patterns
	a.analyzeStructuralPatterns()
	
	// Analyze dependency patterns
	a.analyzeDependencyPatterns()
	
	return nil
}

func (a *GoCodeAnalyzer) analyzeNamingPatterns() {
	// Count naming patterns
	patterns := make(map[string]int)
	
	for _, fn := range a.functions {
		for _, pattern := range fn.Patterns {
			patterns[pattern]++
		}
	}
	
	// Convert to PatternInfo
	for pattern, count := range patterns {
		patternInfo := &PatternInfo{
			Type:         "naming",
			Name:         pattern,
			Frequency:    count,
			Confidence:   float64(count) / float64(len(a.functions)),
			TestStrategy: a.getTestStrategyForPattern(pattern),
			Locations:    make([]string, 0),
		}
		
		// Collect locations
		for _, fn := range a.functions {
			for _, fnPattern := range fn.Patterns {
				if fnPattern == pattern {
					patternInfo.Locations = append(patternInfo.Locations, 
						fmt.Sprintf("%s:%s", fn.FilePath, fn.Name))
				}
			}
		}
		
		a.patterns = append(a.patterns, patternInfo)
	}
}

func (a *GoCodeAnalyzer) analyzeStructuralPatterns() {
	// Analyze interface implementation patterns
	interfaceCount := len(a.interfaces)
	if interfaceCount > 0 {
		patternInfo := &PatternInfo{
			Type:         "structural",
			Name:         "interface_usage",
			Frequency:    interfaceCount,
			Confidence:   0.9,
			TestStrategy: "mock_based_testing",
			Locations:    make([]string, 0),
		}
		
		for _, iface := range a.interfaces {
			patternInfo.Locations = append(patternInfo.Locations, 
				fmt.Sprintf("%s:%s", iface.FilePath, iface.Name))
		}
		
		a.patterns = append(a.patterns, patternInfo)
	}
}

func (a *GoCodeAnalyzer) analyzeDependencyPatterns() {
	// Group dependencies by type
	depTypes := make(map[string]int)
	
	for _, dep := range a.dependencies {
		if dep.IsStandard {
			depTypes["standard"]++
		} else if dep.IsLocal {
			depTypes["local"]++
		} else {
			depTypes["external"]++
		}
	}
	
	// Create patterns for dependency usage
	for depType, count := range depTypes {
		patternInfo := &PatternInfo{
			Type:         "dependency",
			Name:         fmt.Sprintf("%s_dependencies", depType),
			Frequency:    count,
			Confidence:   0.8,
			TestStrategy: a.getTestStrategyForDependency(depType),
			Locations:    make([]string, 0),
		}
		
		a.patterns = append(a.patterns, patternInfo)
	}
}

func (a *GoCodeAnalyzer) getTestStrategyForPattern(pattern string) string {
	strategies := map[string]string{
		"getter_pattern":     "property_based_testing",
		"setter_pattern":     "state_verification_testing",
		"constructor_pattern": "initialization_testing",
		"validation_pattern":  "boundary_value_testing",
		"error_handling":      "error_path_testing",
	}
	
	if strategy, exists := strategies[pattern]; exists {
		return strategy
	}
	
	return "unit_testing"
}

func (a *GoCodeAnalyzer) getTestStrategyForDependency(depType string) string {
	strategies := map[string]string{
		"standard": "integration_testing",
		"local":    "unit_testing",
		"external": "mock_based_testing",
	}
	
	return strategies[depType]
}

// assessRisks identifies potential risk areas in the codebase
func (a *GoCodeAnalyzer) assessRisks() error {
	// Assess complexity risks
	for _, fn := range a.functions {
		if fn.Complexity > a.config.MaxComplexity {
			riskInfo := &RiskAreaInfo{
				Type:        "complexity",
				Location:    fmt.Sprintf("%s:%s:%d", fn.FilePath, fn.Name, fn.StartLine),
				Severity:    core.RiskLevelHigh,
				Description: fmt.Sprintf("Function '%s' has high complexity (%d)", fn.Name, fn.Complexity),
				Patterns:    []string{"high_complexity"},
				Mitigation:  "Consider breaking down into smaller functions and add comprehensive unit tests",
			}
			a.riskAreas = append(a.riskAreas, riskInfo)
		}
	}
	
	// Assess external dependency risks
	externalDeps := 0
	for _, dep := range a.dependencies {
		if !dep.IsStandard && !dep.IsLocal {
			externalDeps++
		}
	}
	
	if externalDeps > 10 {
		riskInfo := &RiskAreaInfo{
			Type:        "dependency",
			Location:    "codebase",
			Severity:    core.RiskLevelMedium,
			Description: fmt.Sprintf("High number of external dependencies (%d)", externalDeps),
			Patterns:    []string{"external_dependencies"},
			Mitigation:  "Add integration tests and dependency isolation",
		}
		a.riskAreas = append(a.riskAreas, riskInfo)
	}
	
	// Assess missing error handling
	functionsWithoutErrorHandling := 0
	for _, fn := range a.functions {
		if fn.CallsExternal && !fn.HasErrorReturn {
			functionsWithoutErrorHandling++
		}
	}
	
	if functionsWithoutErrorHandling > 0 {
		riskInfo := &RiskAreaInfo{
			Type:        "error_handling",
			Location:    "multiple_functions",
			Severity:    core.RiskLevelMedium,
			Description: fmt.Sprintf("%d functions with external calls lack error handling", functionsWithoutErrorHandling),
			Patterns:    []string{"missing_error_handling"},
			Mitigation:  "Add error handling and test error scenarios",
		}
		a.riskAreas = append(a.riskAreas, riskInfo)
	}
	
	return nil
}

// convertToCodeAnalysis converts internal analysis to core format
func (a *GoCodeAnalyzer) convertToCodeAnalysis(codebase *core.Codebase) *core.CodeAnalysis {
	analysis := &core.CodeAnalysis{
		CodebaseID:       fmt.Sprintf("analysis_%d", time.Now().Unix()),
		TotalFiles:       len(a.functions), // Simplified
		TotalLines:       a.calculateTotalLines(),
		Complexity:       a.calculateTotalComplexity(),
		TestableUnits:    make([]*core.TestableUnit, 0),
		Dependencies:     make([]*core.Dependency, 0),
		RiskAreas:        make([]*core.RiskArea, 0),
		CoverageGaps:     make([]*core.CoverageGap, 0),
		Patterns:         make([]*core.CodePattern, 0),
		Metadata:         make(map[string]interface{}),
	}
	
	// Convert functions to testable units
	for _, fn := range a.functions {
		unit := &core.TestableUnit{
			ID:           fmt.Sprintf("unit_%s_%s", fn.PackageName, fn.Name),
			Type:         "function",
			Name:         fn.Name,
			Path:         fn.FilePath,
			Complexity:   fn.Complexity,
			Dependencies: fn.Dependencies,
			Parameters:   a.convertParameters(fn.Parameters),
			ReturnTypes:  fn.ReturnTypes,
			Examples:     make([]string, 0), // Will be generated later
			Priority:     fn.TestPriority,
		}
		analysis.TestableUnits = append(analysis.TestableUnits, unit)
	}
	
	// Convert dependencies
	for _, dep := range a.dependencies {
		dependency := &core.Dependency{
			Name:    dep.ImportPath,
			Type:    a.getDependencyType(dep),
			Version: "unknown", // Could be extracted from go.mod
			Path:    dep.ImportPath,
		}
		analysis.Dependencies = append(analysis.Dependencies, dependency)
	}
	
	// Convert risk areas
	for _, risk := range a.riskAreas {
		riskArea := &core.RiskArea{
			Name:        risk.Type,
			Type:        risk.Type,
			Severity:    risk.Severity,
			Description: risk.Description,
			Location:    risk.Location,
		}
		analysis.RiskAreas = append(analysis.RiskAreas, riskArea)
	}
	
	// Convert patterns
	for _, pattern := range a.patterns {
		codePattern := &core.CodePattern{
			Name:         pattern.Name,
			Type:         pattern.Type,
			Occurrences:  pattern.Frequency,
			Examples:     pattern.Locations,
			TestStrategy: pattern.TestStrategy,
		}
		analysis.Patterns = append(analysis.Patterns, codePattern)
	}
	
	// Add metadata
	analysis.Metadata["analyzer_version"] = "1.0.0"
	analysis.Metadata["analysis_time"] = time.Now().Format(time.RFC3339)
	analysis.Metadata["function_count"] = len(a.functions)
	analysis.Metadata["type_count"] = len(a.types)
	analysis.Metadata["interface_count"] = len(a.interfaces)
	
	return analysis
}

// Helper methods for conversion

func (a *GoCodeAnalyzer) calculateTotalLines() int {
	totalLines := 0
	for _, fn := range a.functions {
		totalLines += fn.EndLine - fn.StartLine + 1
	}
	return totalLines
}

func (a *GoCodeAnalyzer) calculateTotalComplexity() int {
	totalComplexity := 0
	for _, fn := range a.functions {
		totalComplexity += fn.Complexity
	}
	return totalComplexity
}

func (a *GoCodeAnalyzer) convertParameters(params []*ParameterInfo) []string {
	result := make([]string, len(params))
	for i, param := range params {
		result[i] = fmt.Sprintf("%s %s", param.Name, param.Type)
	}
	return result
}

func (a *GoCodeAnalyzer) getDependencyType(dep *DependencyInfo) string {
	if dep.IsStandard {
		return "standard"
	}
	if dep.IsLocal {
		return "local"
	}
	return "external"
}

// GetAnalysisReport returns a detailed analysis report
func (a *GoCodeAnalyzer) GetAnalysisReport() *AnalysisReport {
	return &AnalysisReport{
		Summary: &AnalysisSummary{
			TotalFunctions:   len(a.functions),
			TotalTypes:       len(a.types),
			TotalInterfaces:  len(a.interfaces),
			TotalDependencies: len(a.dependencies),
			TotalPatterns:    len(a.patterns),
			TotalRiskAreas:   len(a.riskAreas),
			AnalysisTime:     time.Now(),
		},
		Functions:    a.functions,
		Types:        a.types,
		Interfaces:   a.interfaces,
		Dependencies: a.dependencies,
		Patterns:     a.patterns,
		RiskAreas:    a.riskAreas,
	}
}

// AnalysisReport represents a comprehensive analysis report
type AnalysisReport struct {
	Summary      *AnalysisSummary   `json:"summary"`
	Functions    []*FunctionInfo    `json:"functions"`
	Types        []*TypeInfo        `json:"types"`
	Interfaces   []*InterfaceInfo   `json:"interfaces"`
	Dependencies []*DependencyInfo  `json:"dependencies"`
	Patterns     []*PatternInfo     `json:"patterns"`
	RiskAreas    []*RiskAreaInfo    `json:"risk_areas"`
}

// AnalysisSummary provides summary statistics
type AnalysisSummary struct {
	TotalFunctions     int       `json:"total_functions"`
	TotalTypes         int       `json:"total_types"`
	TotalInterfaces    int       `json:"total_interfaces"`
	TotalDependencies  int       `json:"total_dependencies"`
	TotalPatterns      int       `json:"total_patterns"`
	TotalRiskAreas     int       `json:"total_risk_areas"`
	AnalysisTime       time.Time `json:"analysis_time"`
}