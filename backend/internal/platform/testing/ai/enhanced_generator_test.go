package ai

import (
	"context"
	"testing"
	"time"

	"openpenpal-backend/internal/platform/testing/core"
)

func TestNewEnhancedAIGenerator(t *testing.T) {
	config := &EnhancedConfig{
		GeneratorConfig: &GeneratorConfig{
			MaxTestCases:      50,
			ComplexityThreshold: 10,
		},
		EnableAdvancedFeatures: true,
		AnalysisDepth:         "deep",
		LearningEnabled:       true,
	}

	generator := NewEnhancedAIGenerator(config)

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if !generator.config.EnableAdvancedFeatures {
		t.Error("Expected EnableAdvancedFeatures to be true")
	}

	if generator.config.AnalysisDepth != "deep" {
		t.Errorf("Expected AnalysisDepth to be 'deep', got %s", generator.config.AnalysisDepth)
	}
}

func TestNewEnhancedAIGenerator_WithNilConfig(t *testing.T) {
	generator := NewEnhancedAIGenerator(nil)

	if generator == nil {
		t.Fatal("Expected generator to be created with default config, got nil")
	}

	// Check default values
	if !generator.config.EnableAdvancedFeatures {
		t.Error("Expected default EnableAdvancedFeatures to be true")
	}

	if generator.config.AnalysisDepth != "deep" {
		t.Error("Expected default AnalysisDepth to be 'deep'")
	}
}

func TestEnhancedAIGenerator_AnalyzeCode(t *testing.T) {
	generator := NewEnhancedAIGenerator(&EnhancedConfig{
		EnableAdvancedFeatures: true,
		AnalysisDepth:         "deep",
	})

	ctx := context.Background()
	codebase := &core.Codebase{
		Path:     "./test",
		Language: "go",
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	analysis, err := generator.AnalyzeCode(ctx, codebase)

	// The actual analysis will use mock data
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected analysis result, got nil")
	}

	// Verify enhanced features were applied
	if generator.stats.AnalysisTime <= 0 {
		t.Error("Expected analysis time to be recorded")
	}
}

func TestEnhancedAIGenerator_GenerateTestCases(t *testing.T) {
	generator := NewEnhancedAIGenerator(&EnhancedConfig{
		GeneratorConfig: &GeneratorConfig{
			MaxTestCases: 10,
		},
		EnableAdvancedFeatures: true,
	})

	ctx := context.Background()
	analysis := createMockAnalysis()

	testCases, err := generator.GenerateTestCases(ctx, analysis)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testCases) == 0 {
		t.Error("Expected test cases to be generated")
	}

	// Verify AI enhancements were applied
	if generator.stats.GenerationTime <= 0 {
		t.Error("Expected generation time to be recorded")
	}

	// Check that AI metadata was added
	if len(testCases) > 0 {
		testCase := testCases[0]
		if testCase.Metadata == nil {
			t.Error("Expected test case metadata to be set")
		} else {
			if aiGenerated, ok := testCase.Metadata["ai_generated"]; !ok || aiGenerated != true {
				t.Error("Expected ai_generated metadata to be true")
			}
		}
	}
}

func TestEnhancedAIGenerator_OptimizeCoverage(t *testing.T) {
	generator := NewEnhancedAIGenerator(&EnhancedConfig{
		GeneratorConfig: &GeneratorConfig{
			MaxTestCases:      5,
			EnableCoverageOpt: true,
		},
		EnableAdvancedFeatures: true,
	})

	ctx := context.Background()
	testCases := createMockTestCases()

	optimizedTestCases, err := generator.OptimizeCoverage(ctx, testCases)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if optimizedTestCases == nil {
		t.Fatal("Expected optimized test cases, got nil")
	}

	// Verify optimization was applied
	if generator.stats.OptimizationTime <= 0 {
		t.Error("Expected optimization time to be recorded")
	}
}

func TestEnhancedAIGenerator_LearnFromResults(t *testing.T) {
	generator := NewEnhancedAIGenerator(&EnhancedConfig{
		LearningEnabled: true,
	})

	ctx := context.Background()
	results := createMockTestResults()

	err := generator.LearnFromResults(ctx, results)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify pattern statistics were updated
	if len(generator.stats.PatternMatches) == 0 {
		t.Error("Expected pattern matches to be recorded")
	}
}

func TestEnhancedAIGenerator_LearnFromResults_Disabled(t *testing.T) {
	generator := NewEnhancedAIGenerator(&EnhancedConfig{
		LearningEnabled: false,
	})

	ctx := context.Background()
	results := createMockTestResults()

	err := generator.LearnFromResults(ctx, results)

	// Should not error even when learning is disabled
	if err != nil {
		t.Errorf("Expected no error when learning disabled, got %v", err)
	}
}

func TestEnhancedAIGenerator_GetGenerationStats(t *testing.T) {
	generator := NewEnhancedAIGenerator(nil)

	// Simulate some activity
	generator.stats.TotalGenerations = 5
	generator.stats.TotalTestsGenerated = 100
	generator.stats.AnalysisTime = 100 * time.Millisecond
	generator.stats.GenerationTime = 200 * time.Millisecond
	generator.stats.OptimizationTime = 50 * time.Millisecond

	stats := generator.GetGenerationStats()

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	if stats.TotalGenerations != 5 {
		t.Errorf("Expected 5 generations, got %d", stats.TotalGenerations)
	}

	if stats.TotalTestsGenerated != 100 {
		t.Errorf("Expected 100 tests generated, got %d", stats.TotalTestsGenerated)
	}

	// Check that average time was calculated
	expectedAvg := (100 + 200 + 50) * time.Millisecond / 5
	if stats.AverageGenTime != expectedAvg {
		t.Errorf("Expected average time %v, got %v", expectedAvg, stats.AverageGenTime)
	}
}

func TestCalculateTestConfidence(t *testing.T) {
	generator := NewEnhancedAIGenerator(nil)
	analysis := createMockAnalysis()

	tests := []struct {
		name     string
		testCase *core.TestCase
		expected float64
	}{
		{
			name: "well structured test",
			testCase: &core.TestCase{
				Setup:      []string{"setup1"},
				Actions:    []string{"action1"},
				Assertions: []string{"assert1"},
			},
			expected: 1.0, // 0.5 + 0.1 + 0.2 + 0.2 = 1.0
		},
		{
			name: "minimal test",
			testCase: &core.TestCase{
				Actions: []string{"action1"},
			},
			expected: 0.7, // 0.5 + 0.2 = 0.7
		},
		{
			name: "empty test",
			testCase: &core.TestCase{},
			expected: 0.5, // base confidence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := generator.calculateTestConfidence(tt.testCase, analysis)
			if confidence != tt.expected {
				t.Errorf("Expected confidence %f, got %f", tt.expected, confidence)
			}
		})
	}
}

func TestFindPatternMatches(t *testing.T) {
	generator := NewEnhancedAIGenerator(nil)

	tests := []struct {
		name     string
		testCase *core.TestCase
		expected []string
	}{
		{
			name: "happy path test",
			testCase: &core.TestCase{
				Name: "Test Function - Happy Path",
			},
			expected: []string{"happy_path"},
		},
		{
			name: "boundary test",
			testCase: &core.TestCase{
				Name: "Test Function - Boundary Values",
			},
			expected: []string{"boundary_testing"},
		},
		{
			name: "error test",
			testCase: &core.TestCase{
				Name: "Test Function - Error Handling",
			},
			expected: []string{"error_handling"},
		},
		{
			name: "no patterns",
			testCase: &core.TestCase{
				Name: "Test Function - Basic",
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := generator.findPatternMatches(tt.testCase)
			if len(patterns) != len(tt.expected) {
				t.Errorf("Expected %d patterns, got %d", len(tt.expected), len(patterns))
			}
			for i, pattern := range patterns {
				if i < len(tt.expected) && pattern != tt.expected[i] {
					t.Errorf("Expected pattern %s, got %s", tt.expected[i], pattern)
				}
			}
		})
	}
}

func TestCalculateComplexityScore(t *testing.T) {
	generator := NewEnhancedAIGenerator(nil)

	tests := []struct {
		name     string
		testCase *core.TestCase
		expected float64
	}{
		{
			name: "complex test case",
			testCase: &core.TestCase{
				Setup:      []string{"s1", "s2"},      // 2 * 0.1 = 0.2
				Actions:    []string{"a1", "a2", "a3"}, // 3 * 0.2 = 0.6
				Assertions: []string{"as1", "as2"},     // 2 * 0.2 = 0.4
				Teardown:   []string{"t1"},             // 1 * 0.1 = 0.1
			},
			expected: 1.3, // 0.2 + 0.6 + 0.4 + 0.1 = 1.3
		},
		{
			name: "simple test case",
			testCase: &core.TestCase{
				Actions:    []string{"a1"},  // 1 * 0.2 = 0.2
				Assertions: []string{"as1"}, // 1 * 0.2 = 0.2
			},
			expected: 0.4, // 0.2 + 0.2 = 0.4
		},
		{
			name:     "empty test case",
			testCase: &core.TestCase{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := generator.calculateComplexityScore(tt.testCase)
			if score != tt.expected {
				t.Errorf("Expected complexity score %f, got %f", tt.expected, score)
			}
		})
	}
}

func TestEstimateTestCoverage(t *testing.T) {
	generator := NewEnhancedAIGenerator(nil)

	tests := []struct {
		name     string
		testCase *core.TestCase
		expected float64
	}{
		{
			name: "critical priority test",
			testCase: &core.TestCase{
				Priority: core.TestPriorityCritical,
			},
			expected: 0.6, // 0.3 + 0.3 = 0.6
		},
		{
			name: "high priority test",
			testCase: &core.TestCase{
				Priority: core.TestPriorityHigh,
			},
			expected: 0.5, // 0.3 + 0.2 = 0.5
		},
		{
			name: "medium priority test",
			testCase: &core.TestCase{
				Priority: core.TestPriorityMedium,
			},
			expected: 0.3, // base coverage
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage := generator.estimateTestCoverage(tt.testCase)
			if coverage != tt.expected {
				t.Errorf("Expected coverage %f, got %f", tt.expected, coverage)
			}
		})
	}
}

func TestGetPriorityWeight(t *testing.T) {
	generator := NewEnhancedAIGenerator(nil)

	tests := []struct {
		name     string
		priority core.TestPriority
		expected int
	}{
		{
			name:     "critical priority",
			priority: core.TestPriorityCritical,
			expected: 4,
		},
		{
			name:     "high priority",
			priority: core.TestPriorityHigh,
			expected: 3,
		},
		{
			name:     "medium priority",
			priority: core.TestPriorityMedium,
			expected: 2,
		},
		{
			name:     "low priority",
			priority: core.TestPriorityLow,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := generator.getPriorityWeight(tt.priority)
			if weight != tt.expected {
				t.Errorf("Expected weight %d, got %d", tt.expected, weight)
			}
		})
	}
}

// Helper functions for tests

func createMockAnalysis() *core.CodeAnalysis {
	return &core.CodeAnalysis{
		CodebaseID: "test_analysis",
		TotalFiles: 5,
		TotalLines: 1000,
		Complexity: 50,
		TestableUnits: []*core.TestableUnit{
			{
				ID:         "unit_1",
				Name:       "TestFunction",
				Complexity: 10,
				Priority:   core.TestPriorityHigh,
			},
		},
		Patterns: []*core.CodePattern{
			{
				Name:        "constructor_pattern",
				Type:        "creational",
				Occurrences: 5,
			},
		},
		RiskAreas: []*core.RiskArea{
			{
				Name:     "High Complexity",
				Severity: core.RiskLevelHigh,
			},
		},
	}
}

func createMockTestCases() []*core.TestCase {
	return []*core.TestCase{
		{
			ID:       "test_1",
			Name:     "Test 1",
			Type:     core.TestCaseTypeUnit,
			Priority: core.TestPriorityHigh,
			Setup:    []string{"setup"},
			Actions:  []string{"action"},
			Assertions: []string{"assert"},
		},
		{
			ID:       "test_2",
			Name:     "Test 2",
			Type:     core.TestCaseTypeIntegration,
			Priority: core.TestPriorityMedium,
			Actions:  []string{"action"},
			Assertions: []string{"assert"},
		},
	}
}

func createMockTestResults() *core.TestResults {
	return &core.TestResults{
		ExecutionID:     "exec_1",
		TotalTests:      10,
		PassedTests:     8,
		FailedTests:     2,
		CoveragePercent: 85.0,
		TestCaseResults: []*core.TestCaseResult{
			{
				TestCaseID: "test_1",
				Status:     core.TestStatusPassed,
			},
			{
				TestCaseID: "test_2",
				Status:     core.TestStatusFailed,
			},
		},
	}
}

// Benchmark tests
func BenchmarkGenerateTestCases(b *testing.B) {
	generator := NewEnhancedAIGenerator(nil)
	ctx := context.Background()
	analysis := createMockAnalysis()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateTestCases(ctx, analysis)
	}
}

func BenchmarkOptimizeCoverage(b *testing.B) {
	generator := NewEnhancedAIGenerator(nil)
	ctx := context.Background()
	testCases := createMockTestCases()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.OptimizeCoverage(ctx, testCases)
	}
}

func BenchmarkCalculateTestConfidence(b *testing.B) {
	generator := NewEnhancedAIGenerator(nil)
	analysis := createMockAnalysis()
	testCase := &core.TestCase{
		Setup:      []string{"setup"},
		Actions:    []string{"action"},
		Assertions: []string{"assert"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.calculateTestConfidence(testCase, analysis)
	}
}