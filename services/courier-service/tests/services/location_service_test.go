package services_test

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LocationServiceTestSuite struct {
	suite.Suite
	locationService *services.LocationService
}

func TestLocationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LocationServiceTestSuite))
}

func (suite *LocationServiceTestSuite) SetupTest() {
	suite.locationService = services.NewLocationService()
}

func (suite *LocationServiceTestSuite) TestCalculateDistance() {
	tests := []struct {
		name      string
		lat1      float64
		lon1      float64
		lat2      float64
		lon2      float64
		expectedKm float64
		tolerance  float64
	}{
		{
			name:       "Same location",
			lat1:       39.9912,
			lon1:       116.3064,
			lat2:       39.9912,
			lon2:       116.3064,
			expectedKm: 0.0,
			tolerance:  0.001,
		},
		{
			name:       "Beijing University to Tsinghua University",
			lat1:       39.9912, // PKU
			lon1:       116.3064,
			lat2:       40.0038, // Tsinghua
			lon2:       116.3265,
			expectedKm: 2.1, // Approximately 2.1km
			tolerance:  0.5,
		},
		{
			name:       "Short distance within campus",
			lat1:       39.9910,
			lon1:       116.3060,
			lat2:       39.9920,
			lon2:       116.3070,
			expectedKm: 0.14, // Very short distance
			tolerance:  0.05,
		},
		{
			name:       "Longer distance across Beijing",
			lat1:       39.9042, // Beijing center
			lon1:       116.4074,
			lat2:       39.9912, // PKU
			lon2:       116.3064,
			expectedKm: 11.0, // About 11km
			tolerance:  2.0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			distance := suite.locationService.CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			assert.InDelta(suite.T(), tt.expectedKm, distance, tt.tolerance)
		})
	}
}

func (suite *LocationServiceTestSuite) TestFindNearbyTasks() {
	// Create test tasks
	tasks := []models.Task{
		{
			TaskID:    "task1",
			PickupLat: 39.9912, // PKU
			PickupLng: 116.3064,
		},
		{
			TaskID:    "task2",
			PickupLat: 40.0038, // Tsinghua
			PickupLng: 116.3265,
		},
		{
			TaskID:    "task3",
			PickupLat: 39.9042, // Beijing center
			PickupLng: 116.4074,
		},
		{
			TaskID:    "task4",
			PickupLat: 0.0, // Invalid coordinates
			PickupLng: 0.0,
		},
	}

	tests := []struct {
		name        string
		courierLat  float64
		courierLng  float64
		radiusKm    float64
		expectedIDs []string
	}{
		{
			name:        "Near PKU, small radius",
			courierLat:  39.9912,
			courierLng:  116.3064,
			radiusKm:    1.0,
			expectedIDs: []string{"task1"}, // Only PKU task
		},
		{
			name:        "Near PKU, medium radius",
			courierLat:  39.9912,
			courierLng:  116.3064,
			radiusKm:    5.0,
			expectedIDs: []string{"task1", "task2"}, // PKU and Tsinghua
		},
		{
			name:        "Near PKU, large radius",
			courierLat:  39.9912,
			courierLng:  116.3064,
			radiusKm:    15.0,
			expectedIDs: []string{"task1", "task2", "task3"}, // All valid tasks
		},
		{
			name:        "Invalid courier location",
			courierLat:  0.0,
			courierLng:  0.0,
			radiusKm:    10.0,
			expectedIDs: []string{}, // No tasks found
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			nearbyTasks := suite.locationService.FindNearbyTasks(tt.courierLat, tt.courierLng, tt.radiusKm, tasks)
			
			actualIDs := make([]string, len(nearbyTasks))
			for i, task := range nearbyTasks {
				actualIDs[i] = task.TaskID
			}

			assert.ElementsMatch(suite.T(), tt.expectedIDs, actualIDs)

			// Verify distance information is added
			for _, task := range nearbyTasks {
				assert.NotEmpty(suite.T(), task.EstimatedDistance)
			}
		})
	}
}

func (suite *LocationServiceTestSuite) TestEstimateDeliveryTime() {
	tests := []struct {
		name       string
		pickupLat  float64
		pickupLng  float64
		deliveryLat float64
		deliveryLng float64
		expected   string
	}{
		{
			name:        "Very short distance",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 39.9920,
			deliveryLng: 116.3070,
			expected:    "30分钟内",
		},
		{
			name:        "Short distance (PKU to Tsinghua)",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 40.0038,
			deliveryLng: 116.3265,
			expected:    "30分钟内", // About 2km, should be under 30 min
		},
		{
			name:        "Medium distance",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 39.9042,
			deliveryLng: 116.4074,
			expected:    "1小时内", // About 11km, should be under 1 hour
		},
		{
			name:        "Long distance",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 39.8,
			deliveryLng: 116.5,
			expected:    "2小时内", // Longer distance
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			time := suite.locationService.EstimateDeliveryTime(tt.pickupLat, tt.pickupLng, tt.deliveryLat, tt.deliveryLng)
			assert.Equal(suite.T(), tt.expected, time)
		})
	}
}

func (suite *LocationServiceTestSuite) TestCalculateReward() {
	tests := []struct {
		name        string
		pickupLat   float64
		pickupLng   float64
		deliveryLat float64
		deliveryLng float64
		priority    string
		minReward   float64
		maxReward   float64
	}{
		{
			name:        "Short distance, normal priority",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 39.9920,
			deliveryLng: 116.3070,
			priority:    models.TaskPriorityNormal,
			minReward:   5.0,  // Base reward
			maxReward:   8.0,  // Base + small distance
		},
		{
			name:        "Medium distance, urgent priority",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 40.0038,
			deliveryLng: 116.3265,
			priority:    models.TaskPriorityUrgent,
			minReward:   9.0,  // Base + distance + urgent bonus
			maxReward:   12.0,
		},
		{
			name:        "Long distance, express priority",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 39.9042,
			deliveryLng: 116.4074,
			priority:    models.TaskPriorityExpress,
			minReward:   15.0, // Base + distance + express bonus
			maxReward:   25.0,
		},
		{
			name:        "Very short distance should have minimum",
			pickupLat:   39.9912,
			pickupLng:   116.3064,
			deliveryLat: 39.9912,
			deliveryLng: 116.3064,
			priority:    models.TaskPriorityNormal,
			minReward:   5.0, // Base reward
			maxReward:   5.0, // No distance bonus
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reward := suite.locationService.CalculateReward(tt.pickupLat, tt.pickupLng, tt.deliveryLat, tt.deliveryLng, tt.priority)
			
			assert.GreaterOrEqual(suite.T(), reward, tt.minReward)
			assert.LessOrEqual(suite.T(), reward, tt.maxReward)
			
			// Reward should be between global min and max
			assert.GreaterOrEqual(suite.T(), reward, 3.0)
			assert.LessOrEqual(suite.T(), reward, 50.0)
			
			// Should be rounded to 2 decimal places
			assert.Equal(suite.T(), reward, math.Round(reward*100)/100)
		})
	}
}

func (suite *LocationServiceTestSuite) TestFormatDistance() {
	tests := []struct {
		name       string
		distanceKm float64
		expected   string
	}{
		{"Zero distance", 0.0, "0m"},
		{"Very short distance", 0.05, "50m"},
		{"Short distance", 0.5, "500m"},
		{"Just under 1km", 0.9, "900m"},
		{"Exactly 1km", 1.0, "1.0km"},
		{"Medium distance", 2.5, "2.5km"},
		{"Long distance", 15.3, "15.3km"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			formatted := suite.locationService.FormatDistance(tt.distanceKm)
			assert.Equal(suite.T(), tt.expected, formatted)
		})
	}
}

func (suite *LocationServiceTestSuite) TestParseLocation() {
	tests := []struct {
		name        string
		location    string
		expectedLat float64
		expectedLng float64
		shouldError bool
	}{
		{
			name:        "Beijing University",
			location:    "北京大学",
			expectedLat: 39.9912,
			expectedLng: 116.3064,
			shouldError: false,
		},
		{
			name:        "PKU abbreviation",
			location:    "北大",
			expectedLat: 39.9912,
			expectedLng: 116.3064,
			shouldError: false,
		},
		{
			name:        "Tsinghua University",
			location:    "清华大学",
			expectedLat: 40.0038,
			expectedLng: 116.3265,
			shouldError: false,
		},
		{
			name:        "Tsinghua abbreviation",
			location:    "清华",
			expectedLat: 40.0038,
			expectedLng: 116.3265,
			shouldError: false,
		},
		{
			name:        "Unknown location (default to Beijing center)",
			location:    "未知地点",
			expectedLat: 39.9042,
			expectedLng: 116.4074,
			shouldError: false,
		},
		{
			name:        "Empty location (default to Beijing center)",
			location:    "",
			expectedLat: 39.9042,
			expectedLng: 116.4074,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			lat, lng, err := suite.locationService.ParseLocation(tt.location)
			
			if tt.shouldError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expectedLat, lat)
				assert.Equal(suite.T(), tt.expectedLng, lng)
			}
		})
	}
}

func (suite *LocationServiceTestSuite) TestGetZoneFromCoordinate() {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		expected string
	}{
		{
			name:     "Beijing University coordinates",
			lat:      39.9912,
			lng:      116.3064,
			expected: "北京大学",
		},
		{
			name:     "Tsinghua University coordinates",
			lat:      40.0038,
			lng:      116.3265,
			expected: "清华大学",
		},
		{
			name:     "Beijing center coordinates",
			lat:      39.9042,
			lng:      116.4074,
			expected: "其他区域",
		},
		{
			name:     "Edge of PKU range",
			lat:      39.990,
			lng:      116.310,
			expected: "北京大学",
		},
		{
			name:     "Outside known zones",
			lat:      40.1,
			lng:      116.5,
			expected: "其他区域",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			zone := suite.locationService.GetZoneFromCoordinate(tt.lat, tt.lng)
			assert.Equal(suite.T(), tt.expected, zone)
		})
	}
}

func (suite *LocationServiceTestSuite) TestHaversineFormulaAccuracy() {
	suite.Run("Haversine formula should be accurate for known distances", func() {
		// Test with known coordinates and distances
		// New York to Los Angeles (approximate distance: 3944 km)
		nyLat, nyLng := 40.7128, -74.0060
		laLat, laLng := 34.0522, -118.2437
		
		distance := suite.locationService.CalculateDistance(nyLat, nyLng, laLat, laLng)
		
		// Should be approximately 3944 km (allow 5% tolerance)
		expected := 3944.0
		tolerance := expected * 0.05
		assert.InDelta(suite.T(), expected, distance, tolerance)
	})
}

func (suite *LocationServiceTestSuite) TestLocationServiceIntegration() {
	suite.Run("Complete location service workflow", func() {
		// Parse locations
		pkuLat, pkuLng, err := suite.locationService.ParseLocation("北京大学")
		assert.NoError(suite.T(), err)
		
		tsinghuaLat, tsinghuaLng, err := suite.locationService.ParseLocation("清华大学")
		assert.NoError(suite.T(), err)
		
		// Calculate distance
		distance := suite.locationService.CalculateDistance(pkuLat, pkuLng, tsinghuaLat, tsinghuaLng)
		assert.Greater(suite.T(), distance, 0.0)
		
		// Format distance
		formattedDistance := suite.locationService.FormatDistance(distance)
		assert.NotEmpty(suite.T(), formattedDistance)
		
		// Estimate delivery time
		deliveryTime := suite.locationService.EstimateDeliveryTime(pkuLat, pkuLng, tsinghuaLat, tsinghuaLng)
		assert.NotEmpty(suite.T(), deliveryTime)
		
		// Calculate reward
		reward := suite.locationService.CalculateReward(pkuLat, pkuLng, tsinghuaLat, tsinghuaLng, models.TaskPriorityNormal)
		assert.GreaterOrEqual(suite.T(), reward, 3.0)
		assert.LessOrEqual(suite.T(), reward, 50.0)
		
		// Get zones
		pkuZone := suite.locationService.GetZoneFromCoordinate(pkuLat, pkuLng)
		tsinghuaZone := suite.locationService.GetZoneFromCoordinate(tsinghuaLat, tsinghuaLng)
		assert.Equal(suite.T(), "北京大学", pkuZone)
		assert.Equal(suite.T(), "清华大学", tsinghuaZone)
	})
}