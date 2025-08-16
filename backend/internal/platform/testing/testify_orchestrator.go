package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/config"
	"go.uber.org/zap"
)

type TestifyOrchestrator struct {
	config         *config.Config
	logger         *zap.Logger
	projectRoot    string
	goModPath      string
	mu             sync.RWMutex
}

type GoTestResult struct {
	Packages    []GoPackageResult `json:"packages"`
	TotalTests  int               `json:"totalTests"`
	PassedTests int               `json:"passedTests"`
	FailedTests int               `json:"failedTests"`
	SkippedTests int              `json:"skippedTests"`
	Duration    time.Duration     `json:"duration"`
	Coverage    *GoCoverage       `json:"coverage,omitempty"`
	Success     bool              `json:"success"`
	Output      string            `json:"output"`
}

type GoPackageResult struct {
	Package   string       `json:"package"`
	Tests     []GoTest     `json:"tests"`
	Duration  time.Duration `json:"duration"`
	Status    string       `json:"status"`
	Coverage  float64      `json:"coverage"`
	Output    string       `json:"output"`
}

type GoTest struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Output   string        `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
	File     string        `json:"file"`
	Line     int           `json:"line"`
}

type GoCoverage struct {
	TotalStatements   int                    `json:"totalStatements"`
	CoveredStatements int                    `json:"coveredStatements"`
	Percentage        float64                `json:"percentage"`
	PackageCoverage   map[string]float64     `json:"packageCoverage"`
	FileCoverage      map[string]FileCoverage `json:"fileCoverage"`
}

type FileCoverage struct {
	Filename    string     `json:"filename"`
	Statements  int        `json:"statements"`
	Covered     int        `json:"covered"`
	Percentage  float64    `json:"percentage"`
	Lines       []LineCoverage `json:"lines"`
}

type LineCoverage struct {
	Line    int  `json:"line"`
	Covered bool `json:"covered"`
	Count   int  `json:"count"`
}

type GoTestOptions struct {
	Package        string            `json:"package,omitempty"`
	TestPattern    string            `json:"testPattern,omitempty"`
	Coverage       bool              `json:"coverage"`
	CoverageMode   string            `json:"coverageMode,omitempty"`
	Verbose        bool              `json:"verbose"`
	Short          bool              `json:"short"`
	Race           bool              `json:"race"`
	Parallel       int               `json:"parallel,omitempty"`
	Timeout        time.Duration     `json:"timeout,omitempty"`
	Tags           string            `json:"tags,omitempty"`
	Environment    map[string]string `json:"environment,omitempty"`
	BenchmarkPattern string          `json:"benchmarkPattern,omitempty"`
	BenchmarkTime    time.Duration   `json:"benchmarkTime,omitempty"`
}

type TestifyTestTemplate struct {
	TestName       string            `json:"testName"`
	PackageName    string            `json:"packageName"`
	TestType       string            `json:"testType"` // unit, integration, benchmark
	FunctionUnderTest string         `json:"functionUnderTest"`
	TestCases      []TestCase        `json:"testCases"`
	Setup          string            `json:"setup"`
	Teardown       string            `json:"teardown"`
	Mocks          []MockDefinition  `json:"mocks"`
	Assertions     []string          `json:"assertions"`
	Complexity     TestComplexity    `json:"complexity"`
}

type TestCase struct {
	Name        string                 `json:"name"`
	Input       map[string]interface{} `json:"input"`
	Expected    map[string]interface{} `json:"expected"`
	ShouldError bool                   `json:"shouldError"`
	Setup       string                 `json:"setup,omitempty"`
}

type MockDefinition struct {
	InterfaceName string            `json:"interfaceName"`
	Methods       []MockMethod      `json:"methods"`
	Package       string            `json:"package"`
}

type MockMethod struct {
	Name       string                 `json:"name"`
	Parameters []Parameter            `json:"parameters"`
	Returns    []Parameter            `json:"returns"`
	Behavior   map[string]interface{} `json:"behavior"`
}

type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func NewTestifyOrchestrator(cfg *config.Config, logger *zap.Logger) *TestifyOrchestrator {
	projectRoot := findProjectRoot()
	
	return &TestifyOrchestrator{
		config:      cfg,
		logger:      logger,
		projectRoot: projectRoot,
		goModPath:   filepath.Join(projectRoot, "go.mod"),
	}
}

func (to *TestifyOrchestrator) ExecuteTests(ctx context.Context, options *GoTestOptions) (*GoTestResult, error) {
	to.mu.Lock()
	defer to.mu.Unlock()

	to.logger.Info("Executing Go tests with Testify", 
		zap.String("project_root", to.projectRoot),
		zap.Any("options", options))

	// Build go test command
	args := to.buildGoTestCommand(options)
	
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = to.projectRoot
	
	// Set environment variables
	env := append(os.Environ())
	if options != nil && options.Environment != nil {
		for key, value := range options.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	cmd.Env = env

	// Execute go test and capture output
	output, err := cmd.CombinedOutput()
	
	to.logger.Info("Go test execution completed",
		zap.Error(err),
		zap.String("output_preview", truncateString(string(output), 500)))

	// Parse go test output
	result, parseErr := to.parseGoTestOutput(output, options)
	if parseErr != nil {
		to.logger.Error("Failed to parse go test output", zap.Error(parseErr))
		return nil, fmt.Errorf("failed to parse go test output: %w", parseErr)
	}

	to.logger.Info("Go tests completed",
		zap.Int("total_tests", result.TotalTests),
		zap.Int("passed", result.PassedTests),
		zap.Int("failed", result.FailedTests),
		zap.Bool("success", result.Success))

	return result, err
}

func (to *TestifyOrchestrator) ExecutePackageTests(ctx context.Context, packagePath string, options *GoTestOptions) (*GoTestResult, error) {
	if options == nil {
		options = &GoTestOptions{}
	}
	options.Package = packagePath
	
	return to.ExecuteTests(ctx, options)
}

func (to *TestifyOrchestrator) ExecuteTestsWithCoverage(ctx context.Context, coverageThreshold float64) (*GoTestResult, error) {
	options := &GoTestOptions{
		Coverage:     true,
		CoverageMode: "atomic",
		Verbose:      true,
	}

	result, err := to.ExecuteTests(ctx, options)
	if err != nil {
		return nil, err
	}

	// Check coverage threshold
	if result.Coverage != nil && result.Coverage.Percentage < coverageThreshold {
		return result, fmt.Errorf("coverage %.2f%% below threshold %.2f%%", 
			result.Coverage.Percentage, coverageThreshold)
	}

	return result, nil
}

func (to *TestifyOrchestrator) ExecuteBenchmarks(ctx context.Context, benchmarkPattern string) (*GoTestResult, error) {
	options := &GoTestOptions{
		BenchmarkPattern: benchmarkPattern,
		BenchmarkTime:    time.Minute,
		Verbose:          true,
	}

	return to.ExecuteTests(ctx, options)
}

func (to *TestifyOrchestrator) DiscoverTestFiles() ([]string, error) {
	var testFiles []string
	
	err := filepath.Walk(to.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}
		
		return nil
	})
	
	return testFiles, err
}

func (to *TestifyOrchestrator) AnalyzeTestCoverage(ctx context.Context, packagePattern string) (*GoCoverage, error) {
	options := &GoTestOptions{
		Package:      packagePattern,
		Coverage:     true,
		CoverageMode: "atomic",
	}

	result, err := to.ExecuteTests(ctx, options)
	if err != nil {
		return nil, err
	}

	return result.Coverage, nil
}

func (to *TestifyOrchestrator) GenerateTestifyTests(ctx context.Context, sourceFile string, complexity TestComplexity) ([]*TestifyTestTemplate, error) {
	to.logger.Info("Generating Testify tests", 
		zap.String("source_file", sourceFile),
		zap.String("complexity", string(complexity)))

	// Parse source file
	functions, err := to.parseSourceFunctions(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source file: %w", err)
	}

	var templates []*TestifyTestTemplate
	for _, function := range functions {
		template := to.generateTestTemplate(function, complexity)
		templates = append(templates, template)
	}

	to.logger.Info("Generated Testify test templates",
		zap.String("source_file", sourceFile),
		zap.Int("template_count", len(templates)))

	return templates, nil
}

func (to *TestifyOrchestrator) CreateTestFromTemplate(template *TestifyTestTemplate) (string, error) {
	testCode := to.generateTestifyTestCode(template)
	
	// Determine test file path
	testFileName := fmt.Sprintf("%s_test.go", strings.ToLower(template.FunctionUnderTest))
	packageDir := filepath.Join(to.projectRoot, "backend", "internal", template.PackageName)
	testFilePath := filepath.Join(packageDir, testFileName)
	
	// Ensure directory exists
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write test file
	if err := os.WriteFile(testFilePath, []byte(testCode), 0644); err != nil {
		return "", fmt.Errorf("failed to write test file: %w", err)
	}
	
	to.logger.Info("Created Testify test",
		zap.String("test_file", testFilePath),
		zap.String("template_name", template.TestName))
	
	return testFilePath, nil
}

func (to *TestifyOrchestrator) buildGoTestCommand(options *GoTestOptions) []string {
	args := []string{"test"}

	if options == nil {
		args = append(args, "./...")
		return args
	}

	// Package selection
	if options.Package != "" {
		args = append(args, options.Package)
	} else {
		args = append(args, "./...")
	}

	// Test pattern
	if options.TestPattern != "" {
		args = append(args, "-run", options.TestPattern)
	}

	// Coverage
	if options.Coverage {
		args = append(args, "-cover")
		if options.CoverageMode != "" {
			args = append(args, "-covermode", options.CoverageMode)
		}
		args = append(args, "-coverprofile", "coverage.out")
	}

	// Verbose output
	if options.Verbose {
		args = append(args, "-v")
	}

	// Short tests
	if options.Short {
		args = append(args, "-short")
	}

	// Race detection
	if options.Race {
		args = append(args, "-race")
	}

	// Parallel execution
	if options.Parallel > 0 {
		args = append(args, "-parallel", fmt.Sprintf("%d", options.Parallel))
	}

	// Timeout
	if options.Timeout > 0 {
		args = append(args, "-timeout", options.Timeout.String())
	}

	// Build tags
	if options.Tags != "" {
		args = append(args, "-tags", options.Tags)
	}

	// Benchmarks
	if options.BenchmarkPattern != "" {
		args = append(args, "-bench", options.BenchmarkPattern)
		if options.BenchmarkTime > 0 {
			args = append(args, "-benchtime", options.BenchmarkTime.String())
		}
	}

	// JSON output for parsing
	args = append(args, "-json")

	return args
}

func (to *TestifyOrchestrator) parseGoTestOutput(output []byte, options *GoTestOptions) (*GoTestResult, error) {
	result := &GoTestResult{
		Packages: []GoPackageResult{},
		Success:  true,
		Output:   string(output),
	}

	lines := strings.Split(string(output), "\n")
	
	currentPackage := &GoPackageResult{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as JSON
		var testEvent map[string]interface{}
		if err := json.Unmarshal([]byte(line), &testEvent); err == nil {
			to.processTestEvent(testEvent, result, currentPackage)
		} else {
			// Parse text output patterns
			to.parseTextOutput(line, result)
		}
	}

	// Finalize current package if exists
	if currentPackage.Package != "" {
		result.Packages = append(result.Packages, *currentPackage)
	}

	// Calculate totals
	for _, pkg := range result.Packages {
		for _, test := range pkg.Tests {
			result.TotalTests++
			switch test.Status {
			case "PASS":
				result.PassedTests++
			case "FAIL":
				result.FailedTests++
				result.Success = false
			case "SKIP":
				result.SkippedTests++
			}
		}
	}

	// Parse coverage if available
	if options != nil && options.Coverage {
		coverage, err := to.parseCoverageProfile("coverage.out")
		if err == nil {
			result.Coverage = coverage
		}
	}

	return result, nil
}

func (to *TestifyOrchestrator) processTestEvent(event map[string]interface{}, result *GoTestResult, currentPackage *GoPackageResult) {
	action, _ := event["Action"].(string)
	packageName, _ := event["Package"].(string)
	testName, _ := event["Test"].(string)
	
	switch action {
	case "start":
		if testName == "" {
			// Package start
			currentPackage.Package = packageName
			currentPackage.Tests = []GoTest{}
		}
	case "pass", "fail", "skip":
		if testName != "" {
			// Test completion
			test := GoTest{
				Name:   testName,
				Status: strings.ToUpper(action),
			}
			
			if elapsed, ok := event["Elapsed"].(float64); ok {
				test.Duration = time.Duration(elapsed * float64(time.Second))
			}
			
			if output, ok := event["Output"].(string); ok {
				test.Output = output
			}
			
			currentPackage.Tests = append(currentPackage.Tests, test)
		}
	}
}

func (to *TestifyOrchestrator) parseTextOutput(line string, result *GoTestResult) {
	// Parse common go test output patterns
	if strings.Contains(line, "PASS") && strings.Contains(line, "coverage:") {
		// Extract coverage percentage
		re := regexp.MustCompile(`coverage: ([\d.]+)%`)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			var percentage float64
			fmt.Sscanf(matches[1], "%f", &percentage)
			if result.Coverage == nil {
				result.Coverage = &GoCoverage{}
			}
			result.Coverage.Percentage = percentage
		}
	}
}

func (to *TestifyOrchestrator) parseCoverageProfile(filename string) (*GoCoverage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	coverage := &GoCoverage{
		PackageCoverage: make(map[string]float64),
		FileCoverage:    make(map[string]FileCoverage),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		
		// Parse coverage line format: filename:startLine.startCol,endLine.endCol numStmt count
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			coverage.TotalStatements++
			if parts[2] != "0" {
				coverage.CoveredStatements++
			}
		}
	}

	if coverage.TotalStatements > 0 {
		coverage.Percentage = float64(coverage.CoveredStatements) / float64(coverage.TotalStatements) * 100
	}

	return coverage, nil
}

type SourceFunction struct {
	Name       string
	Package    string
	Parameters []string
	Returns    []string
	Comments   []string
	File       string
	Line       int
}

func (to *TestifyOrchestrator) parseSourceFunctions(filename string) ([]SourceFunction, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var functions []SourceFunction

	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.IsExported() {
				function := SourceFunction{
					Name:    fn.Name.Name,
					Package: file.Name.Name,
					File:    filename,
					Line:    fset.Position(fn.Pos()).Line,
				}
				
				// Extract parameters and returns
				if fn.Type.Params != nil {
					for _, field := range fn.Type.Params.List {
						function.Parameters = append(function.Parameters, fmt.Sprintf("%v", field.Type))
					}
				}
				
				if fn.Type.Results != nil {
					for _, field := range fn.Type.Results.List {
						function.Returns = append(function.Returns, fmt.Sprintf("%v", field.Type))
					}
				}
				
				functions = append(functions, function)
			}
		}
		return true
	})

	return functions, nil
}

func (to *TestifyOrchestrator) generateTestTemplate(function SourceFunction, complexity TestComplexity) *TestifyTestTemplate {
	template := &TestifyTestTemplate{
		TestName:          fmt.Sprintf("Test%s", function.Name),
		PackageName:       function.Package,
		TestType:          "unit",
		FunctionUnderTest: function.Name,
		Complexity:        complexity,
		TestCases:         []TestCase{},
		Assertions:        []string{},
	}

	// Generate basic test cases
	template.TestCases = append(template.TestCases, TestCase{
		Name:        "ValidInput",
		Input:       map[string]interface{}{"param1": "valid_value"},
		Expected:    map[string]interface{}{"result": "expected_value"},
		ShouldError: false,
	})

	if complexity == IntermediateTest || complexity == ComplexTest {
		template.TestCases = append(template.TestCases, TestCase{
			Name:        "InvalidInput",
			Input:       map[string]interface{}{"param1": nil},
			Expected:    map[string]interface{}{},
			ShouldError: true,
		})
	}

	if complexity == ComplexTest {
		template.TestCases = append(template.TestCases, TestCase{
			Name:        "EdgeCase",
			Input:       map[string]interface{}{"param1": "edge_case_value"},
			Expected:    map[string]interface{}{"result": "edge_case_result"},
			ShouldError: false,
		})
	}

	// Generate assertions
	template.Assertions = []string{
		"assert.NotNil(t, result)",
		"assert.NoError(t, err)",
	}

	return template
}

func (to *TestifyOrchestrator) generateTestifyTestCode(template *TestifyTestTemplate) string {
	return fmt.Sprintf(`package %s

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type %sTestSuite struct {
	suite.Suite
}

func (suite *%sTestSuite) SetupTest() {
	// Setup test environment
}

func (suite *%sTestSuite) TearDownTest() {
	// Cleanup after test
}

%s

func Test%sTestSuite(t *testing.T) {
	suite.Run(t, new(%sTestSuite))
}`,
		template.PackageName,
		template.FunctionUnderTest,
		template.FunctionUnderTest,
		template.FunctionUnderTest,
		to.generateTestMethods(template),
		template.FunctionUnderTest,
		template.FunctionUnderTest)
}

func (to *TestifyOrchestrator) generateTestMethods(template *TestifyTestTemplate) string {
	methods := []string{}

	for _, testCase := range template.TestCases {
		method := fmt.Sprintf(`func (suite *%sTestSuite) Test%s%s() {
	// Test case: %s
	// TODO: Implement test logic
	%s
}`,
			template.FunctionUnderTest,
			template.FunctionUnderTest,
			testCase.Name,
			testCase.Name,
			strings.Join(template.Assertions, "\n\t"))
		
		methods = append(methods, method)
	}

	return strings.Join(methods, "\n\n")
}