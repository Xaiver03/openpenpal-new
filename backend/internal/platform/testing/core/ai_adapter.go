// Package core provides adapters for integrating AI modules with the testing engine
package core

import (
	"context"
	"log"

	"openpenpal-backend/internal/platform/testing/ai"
)

// EnhancedAIConfig provides configuration for the enhanced AI generator
type EnhancedAIConfig struct {
	ModelPath              string  `json:"model_path"`
	ConfidenceThreshold    float64 `json:"confidence_threshold"`
	MaxGeneratedTests      int     `json:"max_generated_tests"`
	LearningRate           float64 `json:"learning_rate"`
	EnableAdvancedFeatures bool    `json:"enable_advanced_features"`
	AnalysisDepth          string  `json:"analysis_depth"`
	LearningEnabled        bool    `json:"learning_enabled"`
}

// EnhancedAIGeneratorAdapter adapts the enhanced AI generator to the core interface
type EnhancedAIGeneratorAdapter struct {
	config    *EnhancedAIConfig
	generator *ai.EnhancedAIGenerator
}

// NewEnhancedAIGeneratorAdapter creates a new adapter
func NewEnhancedAIGeneratorAdapter(config *EnhancedAIConfig) *EnhancedAIGeneratorAdapter {
	// Convert to AI module configuration
	aiConfig := &ai.EnhancedConfig{
		GeneratorConfig: &ai.GeneratorConfig{
			MaxTestCases:         config.MaxGeneratedTests,
			MinCoverageTarget:    0.85,
			ComplexityThreshold:  15,
			EnablePatternLearning: config.LearningEnabled,
			EnableCoverageOpt:    true,
			TestDataSize:         1000,
			RandomSeed:          42,
		},
		EnableAdvancedFeatures: config.EnableAdvancedFeatures,
		AnalysisDepth:         config.AnalysisDepth,
		LearningEnabled:       config.LearningEnabled,
		CacheResults:          true,
		LogLevel:              "info",
	}
	
	generator := ai.NewEnhancedAIGenerator(aiConfig)
	
	return &EnhancedAIGeneratorAdapter{
		config:    config,
		generator: generator,
	}
}

// AnalyzeCode implements the AITestGenerator interface
func (a *EnhancedAIGeneratorAdapter) AnalyzeCode(ctx context.Context, codebase *Codebase) (*CodeAnalysis, error) {
	log.Println("🔍 Using enhanced AI analyzer for code analysis")
	return a.generator.AnalyzeCode(ctx, codebase)
}

// GenerateTestCases implements the AITestGenerator interface
func (a *EnhancedAIGeneratorAdapter) GenerateTestCases(ctx context.Context, analysis *CodeAnalysis) ([]*TestCase, error) {
	log.Println("🧠 Using enhanced AI generator for test case generation")
	return a.generator.GenerateTestCases(ctx, analysis)
}

// OptimizeCoverage implements the AITestGenerator interface
func (a *EnhancedAIGeneratorAdapter) OptimizeCoverage(ctx context.Context, testCases []*TestCase) ([]*TestCase, error) {
	log.Println("🎯 Using enhanced AI optimizer for coverage optimization")
	return a.generator.OptimizeCoverage(ctx, testCases)
}

// LearnFromResults implements the AITestGenerator interface
func (a *EnhancedAIGeneratorAdapter) LearnFromResults(ctx context.Context, results *TestResults) error {
	log.Println("📚 Using enhanced AI learner for results analysis")
	return a.generator.LearnFromResults(ctx, results)
}

// GetGenerationStats returns statistics from the enhanced AI generator
func (a *EnhancedAIGeneratorAdapter) GetGenerationStats() *ai.GenerationStats {
	return a.generator.GetGenerationStats()
}