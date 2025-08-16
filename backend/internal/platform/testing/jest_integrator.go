package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/config"
	"go.uber.org/zap"
)

type JestIntegrator struct {
	config         *config.Config
	logger         *zap.Logger
	projectRoot    string
	jestConfigPath string
	mu             sync.RWMutex
}

type JestTestResult struct {
	TestResults []JestTestSuite `json:"testResults"`
	NumTotalTests     int       `json:"numTotalTests"`
	NumPassedTests    int       `json:"numPassedTests"`
	NumFailedTests    int       `json:"numFailedTests"`
	NumPendingTests   int       `json:"numPendingTests"`
	StartTime         int64     `json:"startTime"`
	Success           bool      `json:"success"`
	CoverageMap       map[string]CoverageInfo `json:"coverageMap,omitempty"`
}

type JestTestSuite struct {
	TestFilePath string        `json:"testFilePath"`
	TestResults  []JestTest    `json:"testResults"`
	StartTime    int64         `json:"startTime"`
	EndTime      int64         `json:"endTime"`
	Status       string        `json:"status"`
	Message      string        `json:"message"`
}

type JestTest struct {
	Title         string   `json:"title"`
	Status        string   `json:"status"`
	Duration      int      `json:"duration"`
	FailureMessages []string `json:"failureMessages"`
	Location      *TestLocation `json:"location,omitempty"`
}

type TestLocation struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	File   string `json:"file"`
}

type CoverageInfo struct {
	Lines     map[string]int `json:"lines"`
	Functions map[string]int `json:"functions"`
	Branches  map[string]int `json:"branches"`
	Statements map[string]int `json:"statements"`
}

type JestExecutionOptions struct {
	TestPathPattern string   `json:"testPathPattern,omitempty"`
	TestNamePattern string   `json:"testNamePattern,omitempty"`
	Coverage        bool     `json:"coverage"`
	UpdateSnapshot  bool     `json:"updateSnapshot"`
	Verbose         bool     `json:"verbose"`
	Silent          bool     `json:"silent"`
	WatchMode       bool     `json:"watchMode"`
	MaxWorkers      int      `json:"maxWorkers,omitempty"`
	Environment     string   `json:"environment,omitempty"`
	SetupFiles      []string `json:"setupFiles,omitempty"`
}

func NewJestIntegrator(cfg *config.Config, logger *zap.Logger) *JestIntegrator {
	projectRoot := findProjectRoot()
	
	return &JestIntegrator{
		config:         cfg,
		logger:         logger,
		projectRoot:    projectRoot,
		jestConfigPath: filepath.Join(projectRoot, "jest.config.js"),
	}
}

func (ji *JestIntegrator) ExecuteTests(ctx context.Context, options *JestExecutionOptions) (*JestTestResult, error) {
	ji.mu.Lock()
	defer ji.mu.Unlock()

	ji.logger.Info("Executing Jest tests", 
		zap.String("project_root", ji.projectRoot),
		zap.Any("options", options))

	// Build Jest command
	args := ji.buildJestCommand(options)
	
	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Dir = ji.projectRoot
	
	// Set environment variables
	cmd.Env = append(os.Environ(), 
		"NODE_ENV=test",
		"CI=true",
	)

	// Execute Jest and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		ji.logger.Error("Jest execution failed", 
			zap.Error(err),
			zap.String("output", string(output)))
	}

	// Parse Jest output
	result, parseErr := ji.parseJestOutput(output)
	if parseErr != nil {
		ji.logger.Error("Failed to parse Jest output", zap.Error(parseErr))
		return nil, fmt.Errorf("failed to parse Jest output: %w", parseErr)
	}

	ji.logger.Info("Jest tests completed",
		zap.Int("total_tests", result.NumTotalTests),
		zap.Int("passed", result.NumPassedTests),
		zap.Int("failed", result.NumFailedTests),
		zap.Bool("success", result.Success))

	return result, err
}

func (ji *JestIntegrator) ExecuteTestFile(ctx context.Context, testFile string, options *JestExecutionOptions) (*JestTestResult, error) {
	if options == nil {
		options = &JestExecutionOptions{}
	}
	options.TestPathPattern = testFile
	
	return ji.ExecuteTests(ctx, options)
}

func (ji *JestIntegrator) ExecuteTestsWithCoverage(ctx context.Context, threshold *CoverageThreshold) (*JestTestResult, error) {
	options := &JestExecutionOptions{
		Coverage: true,
		Silent:   false,
		Verbose:  true,
	}

	result, err := ji.ExecuteTests(ctx, options)
	if err != nil {
		return nil, err
	}

	// Validate coverage against threshold
	if threshold != nil && result.CoverageMap != nil {
		if !ji.validateCoverage(result.CoverageMap, threshold) {
			return result, fmt.Errorf("coverage threshold not met")
		}
	}

	return result, nil
}

func (ji *JestIntegrator) GenerateAIEnhancedTests(ctx context.Context, component string, complexity TestComplexity) ([]*GeneratedTest, error) {
	ji.logger.Info("Generating AI-enhanced Jest tests", 
		zap.String("component", component),
		zap.String("complexity", string(complexity)))

	// Analyze existing component
	componentPath := ji.findComponentPath(component)
	if componentPath == "" {
		return nil, fmt.Errorf("component not found: %s", component)
	}

	// Read component source
	source, err := os.ReadFile(componentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component: %w", err)
	}

	// Generate Jest test patterns
	testPatterns := ji.generateJestTestPatterns(string(source), complexity)
	
	var generatedTests []*GeneratedTest
	for _, pattern := range testPatterns {
		test := &GeneratedTest{
			Name:           pattern.Name,
			TestFramework:  "jest",
			Code:          pattern.Code,
			Dependencies:  pattern.Dependencies,
			Complexity:    complexity,
			GeneratedAt:   time.Now(),
		}
		generatedTests = append(generatedTests, test)
	}

	ji.logger.Info("Generated AI-enhanced Jest tests",
		zap.String("component", component),
		zap.Int("test_count", len(generatedTests)))

	return generatedTests, nil
}

func (ji *JestIntegrator) buildJestCommand(options *JestExecutionOptions) []string {
	args := []string{"test"}

	if options.TestPathPattern != "" {
		args = append(args, options.TestPathPattern)
	}

	if options.TestNamePattern != "" {
		args = append(args, "--testNamePattern", options.TestNamePattern)
	}

	if options.Coverage {
		args = append(args, "--coverage")
	}

	if options.UpdateSnapshot {
		args = append(args, "--updateSnapshot")
	}

	if options.Verbose {
		args = append(args, "--verbose")
	}

	if options.Silent {
		args = append(args, "--silent")
	}

	if options.WatchMode {
		args = append(args, "--watch")
	}

	if options.MaxWorkers > 0 {
		args = append(args, "--maxWorkers", fmt.Sprintf("%d", options.MaxWorkers))
	}

	// Always output JSON for parsing
	args = append(args, "--json")

	return args
}

func (ji *JestIntegrator) parseJestOutput(output []byte) (*JestTestResult, error) {
	var result JestTestResult
	
	// Jest outputs both regular logs and JSON result
	// We need to extract the JSON part
	lines := strings.Split(string(output), "\n")
	var jsonLine string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") && strings.Contains(line, "testResults") {
			jsonLine = line
			break
		}
	}

	if jsonLine == "" {
		return nil, fmt.Errorf("no JSON output found in Jest result")
	}

	if err := json.Unmarshal([]byte(jsonLine), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Jest JSON output: %w", err)
	}

	return &result, nil
}

func (ji *JestIntegrator) findComponentPath(component string) string {
	// Search patterns for React components
	searchPaths := []string{
		filepath.Join(ji.projectRoot, "frontend", "src", "components", component+".tsx"),
		filepath.Join(ji.projectRoot, "frontend", "src", "components", component+".jsx"),
		filepath.Join(ji.projectRoot, "frontend", "src", "components", component, "index.tsx"),
		filepath.Join(ji.projectRoot, "frontend", "src", "components", component, "index.jsx"),
		filepath.Join(ji.projectRoot, "frontend", "src", "pages", component+".tsx"),
		filepath.Join(ji.projectRoot, "frontend", "src", "pages", component+".jsx"),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

type TestPattern struct {
	Name         string
	Code         string
	Dependencies []string
}

type TestComplexity string

const (
	SimpleTest    TestComplexity = "simple"
	IntermediateTest TestComplexity = "intermediate"
	ComplexTest   TestComplexity = "complex"
)

type GeneratedTest struct {
	Name          string         `json:"name"`
	TestFramework string         `json:"test_framework"`
	Code          string         `json:"code"`
	Dependencies  []string       `json:"dependencies"`
	Complexity    TestComplexity `json:"complexity"`
	GeneratedAt   time.Time      `json:"generated_at"`
}

type CoverageThreshold struct {
	Lines      int `json:"lines"`
	Functions  int `json:"functions"`
	Branches   int `json:"branches"`
	Statements int `json:"statements"`
}

func (ji *JestIntegrator) generateJestTestPatterns(source string, complexity TestComplexity) []TestPattern {
	var patterns []TestPattern

	// Basic patterns for all complexities
	patterns = append(patterns, TestPattern{
		Name: "Component Rendering Test",
		Code: ji.generateRenderingTest(source),
		Dependencies: []string{"@testing-library/react", "@testing-library/jest-dom"},
	})

	if complexity == IntermediateTest || complexity == ComplexTest {
		patterns = append(patterns, TestPattern{
			Name: "User Interaction Test",
			Code: ji.generateInteractionTest(source),
			Dependencies: []string{"@testing-library/react", "@testing-library/user-event"},
		})
	}

	if complexity == ComplexTest {
		patterns = append(patterns, TestPattern{
			Name: "Integration Test",
			Code: ji.generateIntegrationTest(source),
			Dependencies: []string{"@testing-library/react", "msw"},
		})
	}

	return patterns
}

func (ji *JestIntegrator) generateRenderingTest(source string) string {
	return `import { render, screen } from '@testing-library/react';
import Component from './Component';

describe('Component', () => {
  test('renders without crashing', () => {
    render(<Component />);
    expect(screen.getByRole('main')).toBeInTheDocument();
  });
});`
}

func (ji *JestIntegrator) generateInteractionTest(source string) string {
	return `import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Component from './Component';

describe('Component Interactions', () => {
  test('handles user interactions correctly', async () => {
    const user = userEvent.setup();
    render(<Component />);
    
    const button = screen.getByRole('button');
    await user.click(button);
    
    expect(screen.getByText(/clicked/i)).toBeInTheDocument();
  });
});`
}

func (ji *JestIntegrator) generateIntegrationTest(source string) string {
	return `import { render, screen, waitFor } from '@testing-library/react';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import Component from './Component';

const server = setupServer(
  rest.get('/api/data', (req, res, ctx) => {
    return res(ctx.json({ data: 'test' }));
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('Component Integration', () => {
  test('integrates with API correctly', async () => {
    render(<Component />);
    
    await waitFor(() => {
      expect(screen.getByText('test')).toBeInTheDocument();
    });
  });
});`
}

func (ji *JestIntegrator) validateCoverage(coverageMap map[string]CoverageInfo, threshold *CoverageThreshold) bool {
	// Calculate aggregate coverage
	totalLines, coveredLines := 0, 0
	totalFunctions, coveredFunctions := 0, 0
	totalBranches, coveredBranches := 0, 0
	totalStatements, coveredStatements := 0, 0

	for _, coverage := range coverageMap {
		for _, count := range coverage.Lines {
			totalLines++
			if count > 0 {
				coveredLines++
			}
		}
		for _, count := range coverage.Functions {
			totalFunctions++
			if count > 0 {
				coveredFunctions++
			}
		}
		for _, count := range coverage.Branches {
			totalBranches++
			if count > 0 {
				coveredBranches++
			}
		}
		for _, count := range coverage.Statements {
			totalStatements++
			if count > 0 {
				coveredStatements++
			}
		}
	}

	linesCoverage := float64(coveredLines) / float64(totalLines) * 100
	functionsCoverage := float64(coveredFunctions) / float64(totalFunctions) * 100
	branchesCoverage := float64(coveredBranches) / float64(totalBranches) * 100
	statementsCoverage := float64(coveredStatements) / float64(totalStatements) * 100

	return linesCoverage >= float64(threshold.Lines) &&
		   functionsCoverage >= float64(threshold.Functions) &&
		   branchesCoverage >= float64(threshold.Branches) &&
		   statementsCoverage >= float64(threshold.Statements)
}

func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Look for package.json to identify project root
	current := wd
	for {
		if _, err := os.Stat(filepath.Join(current, "package.json")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return wd
}