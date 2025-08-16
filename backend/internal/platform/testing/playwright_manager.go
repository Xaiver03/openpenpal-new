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

type PlaywrightManager struct {
	config         *config.Config
	logger         *zap.Logger
	projectRoot    string
	configPath     string
	mu             sync.RWMutex
}

type PlaywrightTestResult struct {
	TestResults     []PlaywrightSuite `json:"suites"`
	Stats          PlaywrightStats   `json:"stats"`
	Config         PlaywrightConfig  `json:"config"`
	Duration       int               `json:"duration"`
	Workers        int               `json:"workers"`
	Status         string            `json:"status"`
	StartTime      time.Time         `json:"startTime"`
	FullReport     string            `json:"fullReport,omitempty"`
}

type PlaywrightSuite struct {
	Title      string            `json:"title"`
	File       string            `json:"file"`
	Tests      []PlaywrightTest  `json:"tests"`
	Duration   int               `json:"duration"`
	Status     string            `json:"status"`
}

type PlaywrightTest struct {
	Title          string                   `json:"title"`
	Location       PlaywrightLocation       `json:"location"`
	ProjectName    string                   `json:"projectName"`
	Status         string                   `json:"status"`
	Duration       int                      `json:"duration"`
	Errors         []PlaywrightError        `json:"errors,omitempty"`
	Attachments    []PlaywrightAttachment   `json:"attachments,omitempty"`
	Steps          []PlaywrightStep         `json:"steps,omitempty"`
}

type PlaywrightLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type PlaywrightError struct {
	Message string                `json:"message"`
	Stack   string                `json:"stack"`
	Location PlaywrightLocation   `json:"location"`
}

type PlaywrightAttachment struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Path        string `json:"path"`
}

type PlaywrightStep struct {
	Title    string `json:"title"`
	Duration int    `json:"duration"`
	Error    string `json:"error,omitempty"`
}

type PlaywrightStats struct {
	Expected   int `json:"expected"`
	Unexpected int `json:"unexpected"`
	Flaky      int `json:"flaky"`
	Skipped    int `json:"skipped"`
}

type PlaywrightConfig struct {
	Projects []PlaywrightProject `json:"projects"`
	Workers  int                 `json:"workers"`
	Timeout  int                 `json:"timeout"`
}

type PlaywrightProject struct {
	Name        string            `json:"name"`
	TestDir     string            `json:"testDir"`
	Use         map[string]interface{} `json:"use"`
}

type PlaywrightExecutionOptions struct {
	TestPattern    string            `json:"testPattern,omitempty"`
	Project        string            `json:"project,omitempty"`
	Headed         bool              `json:"headed"`
	Debug          bool              `json:"debug"`
	UI             bool              `json:"ui"`
	Reporter       string            `json:"reporter,omitempty"`
	Workers        int               `json:"workers,omitempty"`
	MaxFailures    int               `json:"maxFailures,omitempty"`
	UpdateSnapshots bool             `json:"updateSnapshots"`
	Grep           string            `json:"grep,omitempty"`
	GrepInvert     string            `json:"grepInvert,omitempty"`
	Environment    map[string]string `json:"environment,omitempty"`
}

type AIE2ETestTemplate struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	UserFlow     []string `json:"userFlow"`
	Assertions   []string `json:"assertions"`
	TestData     map[string]interface{} `json:"testData"`
	Complexity   TestComplexity `json:"complexity"`
	BrowserTypes []string `json:"browserTypes"`
}

func NewPlaywrightManager(cfg *config.Config, logger *zap.Logger) *PlaywrightManager {
	projectRoot := findProjectRoot()
	
	return &PlaywrightManager{
		config:      cfg,
		logger:      logger,
		projectRoot: projectRoot,
		configPath:  filepath.Join(projectRoot, "frontend", "playwright.config.ts"),
	}
}

func (pm *PlaywrightManager) ExecuteTests(ctx context.Context, options *PlaywrightExecutionOptions) (*PlaywrightTestResult, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.logger.Info("Executing Playwright tests", 
		zap.String("project_root", pm.projectRoot),
		zap.Any("options", options))

	// Build Playwright command
	args := pm.buildPlaywrightCommand(options)
	
	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = filepath.Join(pm.projectRoot, "frontend")
	
	// Set environment variables
	env := append(os.Environ(), 
		"NODE_ENV=test",
		"CI=true",
		"PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1",
	)
	
	// Add custom environment variables
	if options != nil && options.Environment != nil {
		for key, value := range options.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	cmd.Env = env

	// Execute Playwright and capture output
	output, err := cmd.CombinedOutput()
	
	pm.logger.Info("Playwright execution completed",
		zap.Error(err),
		zap.String("output_preview", truncateString(string(output), 500)))

	// Parse Playwright output
	result, parseErr := pm.parsePlaywrightOutput(output, options)
	if parseErr != nil {
		pm.logger.Error("Failed to parse Playwright output", zap.Error(parseErr))
		return nil, fmt.Errorf("failed to parse Playwright output: %w", parseErr)
	}

	pm.logger.Info("Playwright tests completed",
		zap.Int("expected", result.Stats.Expected),
		zap.Int("unexpected", result.Stats.Unexpected),
		zap.Int("flaky", result.Stats.Flaky),
		zap.String("status", result.Status))

	return result, err
}

func (pm *PlaywrightManager) ExecuteTestFile(ctx context.Context, testFile string, options *PlaywrightExecutionOptions) (*PlaywrightTestResult, error) {
	if options == nil {
		options = &PlaywrightExecutionOptions{}
	}
	options.TestPattern = testFile
	
	return pm.ExecuteTests(ctx, options)
}

func (pm *PlaywrightManager) ExecuteCrossBrowserTests(ctx context.Context, testPattern string) (*PlaywrightTestResult, error) {
	options := &PlaywrightExecutionOptions{
		TestPattern: testPattern,
		Reporter:    "json",
		Workers:     3, // Parallel execution across browsers
	}
	
	return pm.ExecuteTests(ctx, options)
}

func (pm *PlaywrightManager) ExecuteVisualRegressionTests(ctx context.Context, updateBaseline bool) (*PlaywrightTestResult, error) {
	options := &PlaywrightExecutionOptions{
		TestPattern:     "**/visual.spec.ts",
		UpdateSnapshots: updateBaseline,
		Reporter:        "html",
	}
	
	return pm.ExecuteTests(ctx, options)
}

func (pm *PlaywrightManager) GenerateAIE2ETests(ctx context.Context, userFlow []string, complexity TestComplexity) ([]*AIE2ETestTemplate, error) {
	pm.logger.Info("Generating AI-enhanced E2E tests", 
		zap.Strings("user_flow", userFlow),
		zap.String("complexity", string(complexity)))

	var templates []*AIE2ETestTemplate

	for i, flow := range userFlow {
		template := &AIE2ETestTemplate{
			Name:         fmt.Sprintf("AI Generated E2E Test %d", i+1),
			Description:  fmt.Sprintf("Tests user flow: %s", flow),
			UserFlow:     []string{flow},
			Assertions:   pm.generateAssertions(flow, complexity),
			TestData:     pm.generateTestData(flow),
			Complexity:   complexity,
			BrowserTypes: []string{"chromium", "firefox", "webkit"},
		}
		templates = append(templates, template)
	}

	return templates, nil
}

func (pm *PlaywrightManager) CreateE2ETestFromTemplate(template *AIE2ETestTemplate) (string, error) {
	testCode := pm.generateE2ETestCode(template)
	
	// Create test file
	testFileName := fmt.Sprintf("%s.spec.ts", strings.ToLower(strings.ReplaceAll(template.Name, " ", "-")))
	testFilePath := filepath.Join(pm.projectRoot, "frontend", "tests", "e2e", "ai-generated", testFileName)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(testFilePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write test file
	if err := os.WriteFile(testFilePath, []byte(testCode), 0644); err != nil {
		return "", fmt.Errorf("failed to write test file: %w", err)
	}
	
	pm.logger.Info("Created AI-generated E2E test",
		zap.String("test_file", testFilePath),
		zap.String("template_name", template.Name))
	
	return testFilePath, nil
}

func (pm *PlaywrightManager) buildPlaywrightCommand(options *PlaywrightExecutionOptions) []string {
	args := []string{"playwright", "test"}

	if options == nil {
		return args
	}

	if options.TestPattern != "" {
		args = append(args, options.TestPattern)
	}

	if options.Project != "" {
		args = append(args, "--project", options.Project)
	}

	if options.Headed {
		args = append(args, "--headed")
	}

	if options.Debug {
		args = append(args, "--debug")
	}

	if options.UI {
		args = append(args, "--ui")
	}

	if options.Reporter != "" {
		args = append(args, "--reporter", options.Reporter)
	} else {
		args = append(args, "--reporter", "json")
	}

	if options.Workers > 0 {
		args = append(args, "--workers", fmt.Sprintf("%d", options.Workers))
	}

	if options.MaxFailures > 0 {
		args = append(args, "--max-failures", fmt.Sprintf("%d", options.MaxFailures))
	}

	if options.UpdateSnapshots {
		args = append(args, "--update-snapshots")
	}

	if options.Grep != "" {
		args = append(args, "--grep", options.Grep)
	}

	if options.GrepInvert != "" {
		args = append(args, "--grep-invert", options.GrepInvert)
	}

	return args
}

func (pm *PlaywrightManager) parsePlaywrightOutput(output []byte, options *PlaywrightExecutionOptions) (*PlaywrightTestResult, error) {
	result := &PlaywrightTestResult{
		StartTime: time.Now(),
		Status:    "unknown",
	}

	outputStr := string(output)
	
	// Try to parse JSON output if reporter was set to json
	if options != nil && options.Reporter == "json" {
		var jsonResult PlaywrightTestResult
		if err := json.Unmarshal(output, &jsonResult); err == nil {
			return &jsonResult, nil
		}
	}

	// Parse text output for key information
	lines := strings.Split(outputStr, "\n")
	
	var stats PlaywrightStats
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "passed") {
			// Parse test stats from summary line
			stats = pm.parseStatsFromLine(line)
		}
		
		if strings.Contains(line, "failed") || strings.Contains(line, "error") {
			result.Status = "failed"
		}
	}

	if result.Status == "unknown" {
		if stats.Unexpected == 0 {
			result.Status = "passed"
		} else {
			result.Status = "failed"
		}
	}

	result.Stats = stats
	result.FullReport = outputStr
	
	return result, nil
}

func (pm *PlaywrightManager) parseStatsFromLine(line string) PlaywrightStats {
	stats := PlaywrightStats{}
	
	// Simple parsing of common Playwright output patterns
	if strings.Contains(line, "passed") {
		// Extract numbers using simple string operations
		parts := strings.Fields(line)
		for i, part := range parts {
			if part == "passed" && i > 0 {
				fmt.Sscanf(parts[i-1], "%d", &stats.Expected)
			}
			if part == "failed" && i > 0 {
				fmt.Sscanf(parts[i-1], "%d", &stats.Unexpected)
			}
			if part == "skipped" && i > 0 {
				fmt.Sscanf(parts[i-1], "%d", &stats.Skipped)
			}
		}
	}
	
	return stats
}

func (pm *PlaywrightManager) generateAssertions(flow string, complexity TestComplexity) []string {
	assertions := []string{
		"expect(page).toHaveTitle(/OpenPenPal/)",
		"expect(page.locator('[data-testid=\"main-content\"]')).toBeVisible()",
	}

	if complexity == IntermediateTest || complexity == ComplexTest {
		assertions = append(assertions, 
			"expect(page.locator('.error')).not.toBeVisible()",
			"expect(page).toHaveURL(/dashboard/)",
		)
	}

	if complexity == ComplexTest {
		assertions = append(assertions,
			"await expect(page.locator('[data-testid=\"success-message\"]')).toBeVisible()",
			"expect(page.locator('.loading')).not.toBeVisible()",
		)
	}

	return assertions
}

func (pm *PlaywrightManager) generateTestData(flow string) map[string]interface{} {
	return map[string]interface{}{
		"testUser": map[string]string{
			"email":    "test@example.com",
			"password": "testpassword123",
		},
		"testData": map[string]string{
			"subject": "Test Letter Subject",
			"content": "This is a test letter content for automated testing.",
		},
		"timeout": 30000,
	}
}

func (pm *PlaywrightManager) generateE2ETestCode(template *AIE2ETestTemplate) string {
	return fmt.Sprintf(`import { test, expect } from '@playwright/test';

test.describe('%s', () => {
  test.beforeEach(async ({ page }) => {
    // Setup test environment
    await page.goto('/');
  });

  test('%s', async ({ page }) => {
    // Test implementation for: %s
    %s
    
    // Assertions
    %s
  });
});

test.describe('Cross-browser compatibility', () => {
  ['chromium', 'firefox', 'webkit'].forEach(browserName => {
    test('¥s works on ¥s', async ({ page, browserName: currentBrowser }) => {
      if (currentBrowser !== browserName) {
        test.skip();
      }
      
      await page.goto('/');
      %s
    });
  });
});`,
		template.Name,
		template.Description,
		strings.Join(template.UserFlow, ", "),
		pm.generateStepsCode(template.UserFlow),
		pm.generateAssertionsCode(template.Assertions),
		template.Name,
		pm.generateAssertionsCode(template.Assertions))
}

func (pm *PlaywrightManager) generateStepsCode(userFlow []string) string {
	steps := []string{}
	
	for _, flow := range userFlow {
		switch {
		case strings.Contains(strings.ToLower(flow), "login"):
			steps = append(steps, `
    // Login flow
    await page.fill('[data-testid="email-input"]', 'test@example.com');
    await page.fill('[data-testid="password-input"]', 'testpassword123');
    await page.click('[data-testid="login-button"]');
    await page.waitForURL('**/dashboard');`)
			
		case strings.Contains(strings.ToLower(flow), "letter"):
			steps = append(steps, `
    // Letter creation flow
    await page.click('[data-testid="new-letter-button"]');
    await page.fill('[data-testid="subject-input"]', 'Test Letter Subject');
    await page.fill('[data-testid="content-textarea"]', 'Test letter content');
    await page.click('[data-testid="send-button"]');`)
			
		case strings.Contains(strings.ToLower(flow), "navigate"):
			steps = append(steps, `
    // Navigation flow
    await page.click('[data-testid="navigation-menu"]');
    await page.click('[data-testid="menu-item"]');`)
			
		default:
			steps = append(steps, fmt.Sprintf(`
    // %s
    // TODO: Implement specific steps for this flow`, flow))
		}
	}
	
	return strings.Join(steps, "\n")
}

func (pm *PlaywrightManager) generateAssertionsCode(assertions []string) string {
	assertionCode := []string{}
	
	for _, assertion := range assertions {
		assertionCode = append(assertionCode, fmt.Sprintf("    %s;", assertion))
	}
	
	return strings.Join(assertionCode, "\n")
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}