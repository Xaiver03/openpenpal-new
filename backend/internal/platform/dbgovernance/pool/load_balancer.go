// Package pool provides AI-driven load prediction for connection pools
package pool

import (
	"fmt"
	"math"
	"sync"
	"time"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// LoadPredictor uses AI algorithms to predict optimal connection pool sizes
type LoadPredictor struct {
	// Historical data
	loadHistory     []LoadDataPoint
	predictionModel *PredictionModel
	
	// Configuration
	historyWindow   time.Duration
	predictionWindow time.Duration
	
	// Thread safety
	mu sync.RWMutex
}

// LoadDataPoint represents a single load measurement
type LoadDataPoint struct {
	Timestamp       time.Time
	ActiveConns     int
	WaitingRequests int
	AverageWaitTime float64
	QueriesPerSecond float64
	CPUUsage        float64
	MemoryUsage     float64
	ErrorRate       float64
}

// PredictionModel contains the AI model for load prediction
type PredictionModel struct {
	// Neural network weights (simplified)
	weights          map[string][]float64
	biases           map[string]float64
	learningRate     float64
	
	// Pattern recognition
	hourlyPatterns   [24]float64
	weeklyPatterns   [7]float64
	seasonalFactors  map[string]float64
	
	// Performance metrics
	accuracy         float64
	lastTraining     time.Time
	trainingCount    int
}

// LoadPrediction represents a load prediction result
type LoadPrediction struct {
	Timestamp          time.Time     `json:"timestamp"`
	PredictedLoad      float64       `json:"predicted_load"`
	OptimalPoolSize    int           `json:"optimal_pool_size"`
	Confidence         float64       `json:"confidence"`
	RecommendedAction  string        `json:"recommended_action"`
	TimeHorizon        time.Duration `json:"time_horizon"`
	Factors            map[string]float64 `json:"factors"`
}

// NewLoadPredictor creates a new load predictor
func NewLoadPredictor() *LoadPredictor {
	return &LoadPredictor{
		loadHistory:      make([]LoadDataPoint, 0),
		historyWindow:    24 * time.Hour,
		predictionWindow: 1 * time.Hour,
		predictionModel:  NewPredictionModel(),
	}
}

// RecordLoad records a load measurement
func (lp *LoadPredictor) RecordLoad(stats *core.PoolStats) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	dataPoint := LoadDataPoint{
		Timestamp:       time.Now(),
		ActiveConns:     stats.ActiveConnections,
		WaitingRequests: stats.WaitingRequests,
		AverageWaitTime: stats.AverageWaitTime,
		QueriesPerSecond: calculateQPS(stats),
		CPUUsage:        getCPUUsage(),
		MemoryUsage:     getMemoryUsage(),
		ErrorRate:       calculateErrorRate(stats),
	}
	
	lp.loadHistory = append(lp.loadHistory, dataPoint)
	
	// Keep only recent history
	cutoff := time.Now().Add(-lp.historyWindow)
	lp.trimHistory(cutoff)
	
	// Update prediction model
	lp.updateModel(dataPoint)
}

// PredictOptimalPoolSize predicts the optimal pool size
func (lp *LoadPredictor) PredictOptimalPoolSize(stats *core.PoolStats) int {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	if len(lp.loadHistory) < 10 {
		// Not enough data, return current size
		return stats.TotalConnections
	}
	
	// Generate prediction
	prediction := lp.generatePrediction()
	
	// Calculate optimal pool size based on prediction
	optimalSize := lp.calculateOptimalSize(prediction, stats)
	
	return optimalSize
}

// PredictLoad predicts future load patterns
func (lp *LoadPredictor) PredictLoad(timeHorizon time.Duration) *LoadPrediction {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	if len(lp.loadHistory) < 5 {
		// Not enough data for prediction
		return &LoadPrediction{
			Timestamp:         time.Now(),
			PredictedLoad:     1.0,
			OptimalPoolSize:   10,
			Confidence:        0.0,
			RecommendedAction: "collect_more_data",
			TimeHorizon:       timeHorizon,
		}
	}
	
	// Use the prediction model
	prediction := lp.generateAdvancedPrediction(timeHorizon)
	
	return prediction
}

// GetLoadTrends returns current load trends
func (lp *LoadPredictor) GetLoadTrends() *LoadTrends {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	if len(lp.loadHistory) < 2 {
		return &LoadTrends{}
	}
	
	// Calculate trends over different time periods
	trends := &LoadTrends{
		LastHour:   lp.calculateTrend(1 * time.Hour),
		Last6Hours: lp.calculateTrend(6 * time.Hour),
		Last24Hours: lp.calculateTrend(24 * time.Hour),
		LastWeek:   lp.calculateTrend(7 * 24 * time.Hour),
	}
	
	return trends
}

// TrainModel trains the prediction model with historical data
func (lp *LoadPredictor) TrainModel() error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	if len(lp.loadHistory) < 100 {
		// Need more data for training
		return nil
	}
	
	// Prepare training data
	trainingData := lp.prepareTrainingData()
	
	// Train the model using simple gradient descent
	lp.trainNeuralNetwork(trainingData)
	
	// Update patterns
	lp.updatePatterns()
	
	lp.predictionModel.lastTraining = time.Now()
	lp.predictionModel.trainingCount++
	
	return nil
}

// Private methods

func (lp *LoadPredictor) trimHistory(cutoff time.Time) {
	for i, point := range lp.loadHistory {
		if point.Timestamp.After(cutoff) {
			lp.loadHistory = lp.loadHistory[i:]
			return
		}
	}
	lp.loadHistory = make([]LoadDataPoint, 0)
}

func (lp *LoadPredictor) updateModel(dataPoint LoadDataPoint) {
	// Update hourly patterns
	hour := dataPoint.Timestamp.Hour()
	lp.predictionModel.hourlyPatterns[hour] = 
		lp.predictionModel.hourlyPatterns[hour]*0.9 + dataPoint.QueriesPerSecond*0.1
	
	// Update weekly patterns
	weekday := int(dataPoint.Timestamp.Weekday())
	lp.predictionModel.weeklyPatterns[weekday] = 
		lp.predictionModel.weeklyPatterns[weekday]*0.9 + dataPoint.QueriesPerSecond*0.1
}

func (lp *LoadPredictor) generatePrediction() *LoadPrediction {
	now := time.Now()
	
	// Get current patterns
	hourPattern := lp.predictionModel.hourlyPatterns[now.Hour()]
	weekPattern := lp.predictionModel.weeklyPatterns[int(now.Weekday())]
	
	// Simple prediction based on patterns
	predictedLoad := (hourPattern + weekPattern) / 2
	
	// Apply neural network prediction
	if len(lp.predictionModel.weights) > 0 {
		networkPrediction := lp.applyNeuralNetwork(lp.getLatestFeatures())
		predictedLoad = predictedLoad*0.6 + networkPrediction*0.4
	}
	
	return &LoadPrediction{
		Timestamp:     now,
		PredictedLoad: predictedLoad,
		Confidence:    lp.predictionModel.accuracy,
		Factors: map[string]float64{
			"hourly_pattern":  hourPattern,
			"weekly_pattern":  weekPattern,
		},
	}
}

func (lp *LoadPredictor) generateAdvancedPrediction(timeHorizon time.Duration) *LoadPrediction {
	prediction := lp.generatePrediction()
	
	// Adjust for time horizon
	futureTime := time.Now().Add(timeHorizon)
	
	// Get future patterns
	futureHour := futureTime.Hour()
	futureWeekday := int(futureTime.Weekday())
	
	hourlyFactor := lp.predictionModel.hourlyPatterns[futureHour]
	weeklyFactor := lp.predictionModel.weeklyPatterns[futureWeekday]
	
	// Adjust prediction
	adjustedLoad := prediction.PredictedLoad
	if hourlyFactor > 0 {
		adjustedLoad *= hourlyFactor / lp.predictionModel.hourlyPatterns[time.Now().Hour()]
	}
	if weeklyFactor > 0 {
		adjustedLoad *= weeklyFactor / lp.predictionModel.weeklyPatterns[int(time.Now().Weekday())]
	}
	
	prediction.PredictedLoad = adjustedLoad
	prediction.TimeHorizon = timeHorizon
	prediction.OptimalPoolSize = int(math.Ceil(adjustedLoad * 1.2))
	
	// Determine recommended action
	if adjustedLoad > prediction.PredictedLoad*1.5 {
		prediction.RecommendedAction = "scale_up"
	} else if adjustedLoad < prediction.PredictedLoad*0.7 {
		prediction.RecommendedAction = "scale_down"
	} else {
		prediction.RecommendedAction = "maintain"
	}
	
	return prediction
}

func (lp *LoadPredictor) calculateOptimalSize(prediction *LoadPrediction, stats *core.PoolStats) int {
	// Base size on predicted load
	baseSize := int(math.Ceil(prediction.PredictedLoad))
	
	// Apply safety margin
	safetyMargin := 1.2
	if prediction.Confidence < 0.7 {
		safetyMargin = 1.5 // Larger margin for uncertain predictions
	}
	
	optimalSize := int(float64(baseSize) * safetyMargin)
	
	// Consider current load
	currentUtilization := float64(stats.ActiveConnections) / float64(stats.TotalConnections)
	if currentUtilization > 0.8 {
		optimalSize = int(float64(optimalSize) * 1.3)
	} else if currentUtilization < 0.3 {
		optimalSize = int(float64(optimalSize) * 0.8)
	}
	
	// Apply bounds
	if optimalSize < 5 {
		optimalSize = 5
	}
	if optimalSize > 200 {
		optimalSize = 200
	}
	
	return optimalSize
}

func (lp *LoadPredictor) calculateTrend(period time.Duration) float64 {
	cutoff := time.Now().Add(-period)
	var values []float64
	
	for _, point := range lp.loadHistory {
		if point.Timestamp.After(cutoff) {
			values = append(values, point.QueriesPerSecond)
		}
	}
	
	if len(values) < 2 {
		return 0
	}
	
	// Simple linear trend calculation
	n := float64(len(values))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Calculate slope (trend)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	return slope
}

func (lp *LoadPredictor) prepareTrainingData() []TrainingExample {
	var examples []TrainingExample
	
	// Create training examples from historical data
	for i := 10; i < len(lp.loadHistory); i++ {
		// Use last 10 data points as features
		features := make([]float64, 0)
		for j := i - 10; j < i; j++ {
			point := lp.loadHistory[j]
			features = append(features, 
				point.QueriesPerSecond,
				point.CPUUsage,
				point.MemoryUsage,
				float64(point.ActiveConns),
			)
		}
		
		// Target is the next load value
		target := lp.loadHistory[i].QueriesPerSecond
		
		examples = append(examples, TrainingExample{
			Features: features,
			Target:   target,
		})
	}
	
	return examples
}

func (lp *LoadPredictor) trainNeuralNetwork(examples []TrainingExample) {
	if len(examples) == 0 {
		return
	}
	
	// Simple gradient descent training
	learningRate := lp.predictionModel.learningRate
	
	for epoch := 0; epoch < 100; epoch++ {
		totalError := 0.0
		
		for _, example := range examples {
			prediction := lp.applyNeuralNetwork(example.Features)
			error := example.Target - prediction
			totalError += error * error
			
			// Update weights (simplified backpropagation)
			lp.updateWeights(example.Features, error, learningRate)
		}
		
		// Calculate accuracy
		avgError := totalError / float64(len(examples))
		lp.predictionModel.accuracy = 1.0 / (1.0 + avgError)
		
		// Early stopping if converged
		if avgError < 0.01 {
			break
		}
	}
}

func (lp *LoadPredictor) applyNeuralNetwork(features []float64) float64 {
	if len(lp.predictionModel.weights) == 0 {
		lp.initializeWeights(len(features))
	}
	
	// Simple feedforward network
	hiddenLayer := make([]float64, 10)
	
	// Calculate hidden layer
	for i := 0; i < 10; i++ {
		sum := 0.0
		for j, feature := range features {
			if weights, exists := lp.predictionModel.weights[fmt.Sprintf("input_%d_%d", j, i)]; exists && len(weights) > 0 {
				sum += feature * weights[0]
			}
		}
		sum += lp.predictionModel.biases[fmt.Sprintf("hidden_%d", i)]
		hiddenLayer[i] = sigmoid(sum)
	}
	
	// Calculate output
	output := 0.0
	for i, h := range hiddenLayer {
		if weights, exists := lp.predictionModel.weights[fmt.Sprintf("hidden_%d_output", i)]; exists && len(weights) > 0 {
			output += h * weights[0]
		}
	}
	output += lp.predictionModel.biases["output"]
	
	return output
}

func (lp *LoadPredictor) initializeWeights(inputSize int) {
	lp.predictionModel.weights = make(map[string][]float64)
	lp.predictionModel.biases = make(map[string]float64)
	
	// Initialize input to hidden weights
	for i := 0; i < inputSize; i++ {
		for j := 0; j < 10; j++ {
			key := fmt.Sprintf("input_%d_%d", i, j)
			lp.predictionModel.weights[key] = []float64{randomWeight()}
		}
	}
	
	// Initialize hidden to output weights
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("hidden_%d_output", i)
		lp.predictionModel.weights[key] = []float64{randomWeight()}
	}
	
	// Initialize biases
	for i := 0; i < 10; i++ {
		lp.predictionModel.biases[fmt.Sprintf("hidden_%d", i)] = randomWeight()
	}
	lp.predictionModel.biases["output"] = randomWeight()
}

func (lp *LoadPredictor) updateWeights(features []float64, error, learningRate float64) {
	// Simplified weight update
	for i, feature := range features {
		for j := 0; j < 10; j++ {
			key := fmt.Sprintf("input_%d_%d", i, j)
			if weights := lp.predictionModel.weights[key]; len(weights) > 0 {
				weights[0] += learningRate * error * feature
			}
		}
	}
}

func (lp *LoadPredictor) updatePatterns() {
	// Update seasonal patterns based on recent data
	now := time.Now()
	month := now.Month().String()
	
	if lp.predictionModel.seasonalFactors == nil {
		lp.predictionModel.seasonalFactors = make(map[string]float64)
	}
	
	// Calculate average load for current month
	cutoff := now.AddDate(0, -1, 0)
	var monthlyLoads []float64
	
	for _, point := range lp.loadHistory {
		if point.Timestamp.After(cutoff) {
			monthlyLoads = append(monthlyLoads, point.QueriesPerSecond)
		}
	}
	
	if len(monthlyLoads) > 0 {
		avgLoad := 0.0
		for _, load := range monthlyLoads {
			avgLoad += load
		}
		lp.predictionModel.seasonalFactors[month] = avgLoad / float64(len(monthlyLoads))
	}
}

func (lp *LoadPredictor) getLatestFeatures() []float64 {
	if len(lp.loadHistory) < 10 {
		return []float64{}
	}
	
	features := make([]float64, 0)
	start := len(lp.loadHistory) - 10
	
	for i := start; i < len(lp.loadHistory); i++ {
		point := lp.loadHistory[i]
		features = append(features,
			point.QueriesPerSecond,
			point.CPUUsage,
			point.MemoryUsage,
			float64(point.ActiveConns),
		)
	}
	
	return features
}

// Helper functions

func NewPredictionModel() *PredictionModel {
	return &PredictionModel{
		weights:         make(map[string][]float64),
		biases:          make(map[string]float64),
		learningRate:    0.01,
		seasonalFactors: make(map[string]float64),
		accuracy:        0.5,
	}
}

func calculateQPS(stats *core.PoolStats) float64 {
	// This would be calculated from actual query metrics
	// For now, estimate based on active connections
	return float64(stats.ActiveConnections) * 10.0
}

func getCPUUsage() float64 {
	// This would get actual CPU usage
	// For now, return simulated value
	return 50.0
}

func getMemoryUsage() float64 {
	// This would get actual memory usage
	// For now, return simulated value
	return 60.0
}

func calculateErrorRate(stats *core.PoolStats) float64 {
	if stats.ConnectionsCreated == 0 {
		return 0
	}
	return float64(stats.FailedConnections) / float64(stats.ConnectionsCreated)
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func randomWeight() float64 {
	return (2.0*math.Sin(float64(time.Now().UnixNano())) - 1.0) * 0.1
}

// Data structures

type TrainingExample struct {
	Features []float64
	Target   float64
}

type LoadTrends struct {
	LastHour    float64 `json:"last_hour"`
	Last6Hours  float64 `json:"last_6_hours"`
	Last24Hours float64 `json:"last_24_hours"`
	LastWeek    float64 `json:"last_week"`
}

