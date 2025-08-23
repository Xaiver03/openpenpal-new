package services_test

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"courier-service/internal/utils"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AssignmentServiceTestSuite struct {
	suite.Suite
	db                *gorm.DB
	mock              sqlmock.Sqlmock
	assignmentService *services.AssignmentService
	locationService   *services.LocationService
	wsManager         *utils.WebSocketManager
}

func TestAssignmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AssignmentServiceTestSuite))
}

func (suite *AssignmentServiceTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	
	suite.db = db
	
	// Migrate test tables
	err = db.AutoMigrate(&models.Task{}, &models.Courier{}, &models.ScanRecord{})
	suite.Require().NoError(err)
	
	// Create services
	suite.locationService = services.NewLocationService()
	suite.wsManager = &utils.WebSocketManager{} // Mock WebSocket manager
	suite.assignmentService = services.NewAssignmentService(db, suite.locationService, suite.wsManager)
}

func (suite *AssignmentServiceTestSuite) TearDownTest() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

func (suite *AssignmentServiceTestSuite) TestCalculateIndividualScore() {
	// Create test courier
	courier := &models.Courier{
		UserID:     "courier123",
		Rating:     4.5,
		Zone:       "北京大学",
		ZoneType:   models.ZoneTypeBuilding,
		ZoneCode:   "PKU001-AREA001-BUILDING001",
		Points:     1500,
		Experience: "Experienced courier with 2 years of service in campus delivery",
	}

	// Create test task at PKU
	taskLat, taskLng := 39.9912, 116.3064

	score := suite.assignmentService.CalculateIndividualScore(courier, taskLat, taskLng)

	// Verify score components
	assert.Equal(suite.T(), *courier, score.Courier)
	assert.Greater(suite.T(), score.Score, 0.0)
	assert.GreaterOrEqual(suite.T(), score.Distance, 0.0)
	assert.GreaterOrEqual(suite.T(), score.CurrentTasks, 0)
}

func (suite *AssignmentServiceTestSuite) TestCalculateIndividualScoreWithHierarchy() {
	tests := []struct {
		name         string
		courier      *models.Courier
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name: "L4 City Director",
			courier: &models.Courier{
				UserID:   "courier_l4",
				Rating:   4.8,
				Zone:     "北京市",
				ZoneType: models.ZoneTypeCity,
				ZoneCode: "BJ001",
				Level:    4,
				Points:   2000,
				Experience: "Senior courier director with city-wide management experience",
			},
			expectedMin: 100.0,
			expectedMax: 200.0,
		},
		{
			name: "L3 School Courier",
			courier: &models.Courier{
				UserID:   "courier_l3",
				Rating:   4.5,
				Zone:     "北京大学",
				ZoneType: models.ZoneTypeSchool,
				ZoneCode: "PKU001",
				Level:    3,
				Points:   1500,
				Experience: "School-level courier with excellent delivery record",
			},
			expectedMin: 90.0,
			expectedMax: 180.0,
		},
		{
			name: "L2 Area Courier",
			courier: &models.Courier{
				UserID:   "courier_l2",
				Rating:   4.2,
				Zone:     "北京大学A区",
				ZoneType: models.ZoneTypeArea,
				ZoneCode: "PKU001-AREA001",
				Level:    2,
				Points:   1000,
				Experience: "Area courier specializing in dormitory deliveries",
			},
			expectedMin: 80.0,
			expectedMax: 170.0,
		},
		{
			name: "L1 Building Courier",
			courier: &models.Courier{
				UserID:   "courier_l1",
				Rating:   4.0,
				Zone:     "北京大学5号楼",
				ZoneType: models.ZoneTypeBuilding,
				ZoneCode: "PKU001-AREA001-BUILDING001",
				Level:    1,
				Points:   500,
				Experience: "Building courier",
			},
			expectedMin: 70.0,
			expectedMax: 160.0,
		},
	}

	task := &models.Task{
		TaskID:           "task123",
		PickupLocation:   "北京大学",
		DeliveryLocation: "清华大学",
		PickupOPCode:     "PK5F3D",
		DeliveryOPCode:   "PK5G2A",
	}
	taskLat, taskLng := 39.9912, 116.3064

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			score := suite.assignmentService.CalculateIndividualScoreWithHierarchy(tt.courier, task, taskLat, taskLng)
			
			assert.Equal(suite.T(), *tt.courier, score.Courier)
			assert.GreaterOrEqual(suite.T(), score.Score, tt.expectedMin)
			assert.LessOrEqual(suite.T(), score.Score, tt.expectedMax)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestGetLevelPriority() {
	tests := []struct {
		zoneType string
		expected int
	}{
		{models.ZoneTypeBuilding, 4},
		{models.ZoneTypeArea, 3},
		{models.ZoneTypeSchool, 2},
		{models.ZoneTypeCity, 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		suite.Run(tt.zoneType, func() {
			priority := suite.assignmentService.GetLevelPriority(tt.zoneType)
			assert.Equal(suite.T(), tt.expected, priority)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestGetHierarchyBonus() {
	tests := []struct {
		zoneType string
		expected float64
	}{
		{models.ZoneTypeBuilding, 40.0},
		{models.ZoneTypeArea, 30.0},
		{models.ZoneTypeSchool, 20.0},
		{models.ZoneTypeCity, 10.0},
		{"unknown", 0.0},
	}

	for _, tt := range tests {
		suite.Run(tt.zoneType, func() {
			bonus := suite.assignmentService.GetHierarchyBonus(tt.zoneType)
			assert.Equal(suite.T(), tt.expected, bonus)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestExtractZoneCodeFromLocation() {
	tests := []struct {
		name     string
		location string
		expected string
	}{
		{
			name:     "Standard zone code format",
			location: "PKU001-AREA001-BUILDING001",
			expected: "PKU001-AREA001-BUILDING001",
		},
		{
			name:     "Short location name",
			location: "北京大学",
			expected: "PKU001-AREA001-BUILDING001", // Default
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.assignmentService.ExtractZoneCodeFromLocation(tt.location)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestExtractParentZoneCode() {
	tests := []struct {
		name           string
		childZoneCode  string
		parentType     string
		expected       string
	}{
		{
			name:          "Building to Area",
			childZoneCode: "PKU001-AREA001-BUILDING001",
			parentType:    models.ZoneTypeArea,
			expected:      "PKU001-AREA001",
		},
		{
			name:          "Area to School",
			childZoneCode: "PKU001-AREA001",
			parentType:    models.ZoneTypeSchool,
			expected:      "PKU001",
		},
		{
			name:          "School to City",
			childZoneCode: "PKU001",
			parentType:    models.ZoneTypeCity,
			expected:      "BJ001",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.assignmentService.ExtractParentZoneCode(tt.childZoneCode, tt.parentType)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestGetSchoolToCityMapping() {
	tests := []struct {
		schoolCode string
		expected   string
	}{
		{"PKU001", "BJ001"},
		{"TSI001", "BJ001"},
		{"FUD001", "FJ001"},
		{"XMU001", "FJ001"},
		{"UNKNOWN", "CN001"},
		{"", "CN001"},
	}

	for _, tt := range tests {
		suite.Run(tt.schoolCode, func() {
			result := suite.assignmentService.GetSchoolToCityMapping(tt.schoolCode)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestValidateOPCodePermission() {
	tests := []struct {
		name                 string
		courier              *models.Courier
		task                 *models.Task
		expected             bool
		description          string
	}{
		{
			name: "L4 courier with city-wide permission",
			courier: &models.Courier{
				Level:                4,
				ManagedOPCodePrefix:  "PK",
			},
			task: &models.Task{
				DeliveryOPCode: "PK5F3D",
			},
			expected:    true,
			description: "L4 can handle same city tasks",
		},
		{
			name: "L3 courier with school permission",
			courier: &models.Courier{
				Level:                3,
				ManagedOPCodePrefix:  "PK",
			},
			task: &models.Task{
				DeliveryOPCode: "PK5F3D",
			},
			expected:    true,
			description: "L3 can handle same school tasks",
		},
		{
			name: "L2 courier with area permission",
			courier: &models.Courier{
				Level:                2,
				ManagedOPCodePrefix:  "PK5F",
			},
			task: &models.Task{
				DeliveryOPCode: "PK5F3D",
			},
			expected:    true,
			description: "L2 can handle same area tasks",
		},
		{
			name: "L1 courier with building permission",
			courier: &models.Courier{
				Level:                1,
				ManagedOPCodePrefix:  "PK5F3D",
			},
			task: &models.Task{
				DeliveryOPCode: "PK5F3D",
			},
			expected:    true,
			description: "L1 can handle exact match tasks",
		},
		{
			name: "L1 courier without permission",
			courier: &models.Courier{
				Level:                1,
				ManagedOPCodePrefix:  "PK5F3D",
			},
			task: &models.Task{
				DeliveryOPCode: "QH2A1B",
			},
			expected:    false,
			description: "L1 cannot handle different building tasks",
		},
		{
			name: "L2 courier without area permission",
			courier: &models.Courier{
				Level:                2,
				ManagedOPCodePrefix:  "PK5F",
			},
			task: &models.Task{
				DeliveryOPCode: "PK2A1B",
			},
			expected:    false,
			description: "L2 cannot handle different area tasks",
		},
		{
			name: "L3 courier without school permission",
			courier: &models.Courier{
				Level:                3,
				ManagedOPCodePrefix:  "PK",
			},
			task: &models.Task{
				DeliveryOPCode: "QH2A1B",
			},
			expected:    false,
			description: "L3 cannot handle different school tasks",
		},
		{
			name: "Missing OP Code configuration",
			courier: &models.Courier{
				Level:                2,
				ManagedOPCodePrefix:  "",
			},
			task: &models.Task{
				DeliveryOPCode: "PK5F3D",
			},
			expected:    false,
			description: "Missing OP Code configuration should fail",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.assignmentService.ValidateOPCodePermission(tt.courier, tt.task)
			assert.Equal(suite.T(), tt.expected, result, tt.description)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestCalculateZoneMatchScore() {
	courier := &models.Courier{
		ZoneCode: "PKU001-AREA001-BUILDING001",
		ZoneType: models.ZoneTypeBuilding,
	}

	tests := []struct {
		name                string
		task                *models.Task
		expectedScore       float64
		description         string
	}{
		{
			name: "Exact zone match",
			task: &models.Task{
				PickupLocation: "PKU001-AREA001-BUILDING001",
			},
			expectedScore: 50.0,
			description:   "Exact match should get highest score",
		},
		{
			name: "Same area different building",
			task: &models.Task{
				PickupLocation: "PKU001-AREA001-BUILDING002",
			},
			expectedScore: 20.0,
			description:   "Same area should get moderate score",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			score := suite.assignmentService.CalculateZoneMatchScore(courier, tt.task)
			assert.Equal(suite.T(), tt.expectedScore, score, tt.description)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestIsParentZone() {
	tests := []struct {
		name              string
		courierZoneCode   string
		courierZoneType   string
		taskZoneCode      string
		expected          bool
		description       string
	}{
		{
			name:             "Area manages building",
			courierZoneCode:  "PKU001-AREA001",
			courierZoneType:  models.ZoneTypeArea,
			taskZoneCode:     "PKU001-AREA001-BUILDING001",
			expected:         true,
			description:      "Area courier can manage buildings in same area",
		},
		{
			name:             "School manages area",
			courierZoneCode:  "PKU001",
			courierZoneType:  models.ZoneTypeSchool,
			taskZoneCode:     "PKU001-AREA001",
			expected:         true,
			description:      "School courier can manage areas in same school",
		},
		{
			name:             "Building cannot manage area",
			courierZoneCode:  "PKU001-AREA001-BUILDING001",
			courierZoneType:  models.ZoneTypeBuilding,
			taskZoneCode:     "PKU001-AREA002",
			expected:         false,
			description:      "Building courier cannot manage higher level zones",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.assignmentService.IsParentZone(tt.courierZoneCode, tt.courierZoneType, tt.taskZoneCode)
			assert.Equal(suite.T(), tt.expected, result, tt.description)
		})
	}
}

func (suite *AssignmentServiceTestSuite) TestGetCurrentTaskCount() {
	// Create test courier
	courierID := "courier123"
	
	// Create test tasks in database
	activeTasks := []models.Task{
		{TaskID: "task1", CourierID: &courierID, Status: models.TaskStatusAccepted},
		{TaskID: "task2", CourierID: &courierID, Status: models.TaskStatusCollected},
		{TaskID: "task3", CourierID: &courierID, Status: models.TaskStatusInTransit},
	}
	
	inactiveTasks := []models.Task{
		{TaskID: "task4", CourierID: &courierID, Status: models.TaskStatusDelivered},
		{TaskID: "task5", CourierID: &courierID, Status: models.TaskStatusFailed},
		{TaskID: "task6", CourierID: &courierID, Status: models.TaskStatusCanceled},
	}
	
	// Insert tasks
	for _, task := range activeTasks {
		suite.db.Create(&task)
	}
	for _, task := range inactiveTasks {
		suite.db.Create(&task)
	}
	
	// Test current task count
	count := suite.assignmentService.GetCurrentTaskCount(courierID)
	assert.Equal(suite.T(), 3, count, "Should count only active tasks")
	
	// Test with non-existent courier
	nonExistentCount := suite.assignmentService.GetCurrentTaskCount("non_existent")
	assert.Equal(suite.T(), 0, nonExistentCount, "Non-existent courier should have 0 tasks")
}

func (suite *AssignmentServiceTestSuite) TestFindOptimalCouriersForTask() {
	// Create test task
	task := models.Task{
		TaskID:         "task123",
		PickupLocation: "北京大学",
		Status:         models.TaskStatusAvailable,
	}
	suite.db.Create(&task)
	
	// Create test couriers
	couriers := []models.Courier{
		{
			UserID:   "courier1",
			Rating:   4.5,
			Zone:     "北京大学",
			ZoneType: models.ZoneTypeBuilding,
			Status:   models.CourierStatusApproved,
		},
		{
			UserID:   "courier2", 
			Rating:   4.2,
			Zone:     "清华大学",
			ZoneType: models.ZoneTypeSchool,
			Status:   models.CourierStatusApproved,
		},
	}
	
	for _, courier := range couriers {
		suite.db.Create(&courier)
	}
	
	// Find optimal couriers
	optimalCouriers, err := suite.assignmentService.FindOptimalCouriersForTask("task123", 5)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(optimalCouriers), 0)
	assert.LessOrEqual(suite.T(), len(optimalCouriers), 5)
}

func (suite *AssignmentServiceTestSuite) TestValidateAssignmentPermission() {
	tests := []struct {
		name     string
		courier  *models.Courier
		task     *models.Task
		expected bool
	}{
		{
			name: "Valid OP Code permission",
			courier: &models.Courier{
				ManagedOPCodePrefix: "PK5F",
				Level:              2,
			},
			task: &models.Task{
				DeliveryOPCode: "PK5F3D",
			},
			expected: true,
		},
		{
			name: "Invalid OP Code permission",
			courier: &models.Courier{
				ManagedOPCodePrefix: "PK5F",
				Level:              2,
			},
			task: &models.Task{
				DeliveryOPCode: "QH2A1B",
			},
			expected: false,
		},
		{
			name: "Legacy zone code validation",
			courier: &models.Courier{
				ManagedOPCodePrefix: "",
				ZoneCode:           "PKU001-AREA001",
				ZoneType:           models.ZoneTypeArea,
			},
			task: &models.Task{
				PickupLocation: "PKU001-AREA001-BUILDING001",
				DeliveryOPCode: "",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.assignmentService.ValidateAssignmentPermission(tt.courier, tt.task)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// Additional helper types for extending the AssignmentService for testing
type TestableAssignmentService struct {
	*services.AssignmentService
}

func (t *TestableAssignmentService) GetLevelPriority(zoneType string) int {
	// Access unexported method through reflection or create public wrapper
	switch zoneType {
	case models.ZoneTypeBuilding:
		return 4
	case models.ZoneTypeArea:
		return 3
	case models.ZoneTypeSchool:
		return 2
	case models.ZoneTypeCity:
		return 1
	default:
		return 0
	}
}

func (t *TestableAssignmentService) GetHierarchyBonus(zoneType string) float64 {
	switch zoneType {
	case models.ZoneTypeBuilding:
		return 40.0
	case models.ZoneTypeArea:
		return 30.0
	case models.ZoneTypeSchool:
		return 20.0
	case models.ZoneTypeCity:
		return 10.0
	default:
		return 0.0
	}
}

func (t *TestableAssignmentService) ExtractZoneCodeFromLocation(location string) string {
	if len(location) >= 20 {
		return location
	}
	return "PKU001-AREA001-BUILDING001"
}

func (t *TestableAssignmentService) ExtractParentZoneCode(childZoneCode string, parentType string) string {
	switch parentType {
	case models.ZoneTypeArea:
		if len(childZoneCode) > 12 {
			return childZoneCode[:12]
		}
	case models.ZoneTypeSchool:
		if len(childZoneCode) > 6 {
			return childZoneCode[:6]
		}
	case models.ZoneTypeCity:
		return t.GetSchoolToCityMapping(childZoneCode)
	}
	return childZoneCode
}

func (t *TestableAssignmentService) GetSchoolToCityMapping(schoolCode string) string {
	if len(schoolCode) >= 3 {
		prefix := schoolCode[:3]
		switch prefix {
		case "PKU", "TSI":
			return "BJ001"
		case "FUD", "XMU":
			return "FJ001"
		default:
			return "CN001"
		}
	}
	return "CN001"
}

func (t *TestableAssignmentService) CalculateZoneMatchScore(courier *models.Courier, task *models.Task) float64 {
	taskZoneCode := t.ExtractZoneCodeFromLocation(task.PickupLocation)

	if courier.ZoneCode == taskZoneCode {
		return 50.0
	}

	if t.IsParentZone(courier.ZoneCode, courier.ZoneType, taskZoneCode) {
		return 30.0
	}

	if courier.ZoneType == models.ZoneTypeBuilding &&
		t.ExtractParentZoneCode(courier.ZoneCode, models.ZoneTypeArea) ==
			t.ExtractParentZoneCode(taskZoneCode, models.ZoneTypeArea) {
		return 20.0
	}

	return 0.0
}

func (t *TestableAssignmentService) IsParentZone(courierZoneCode, courierZoneType, taskZoneCode string) bool {
	switch courierZoneType {
	case models.ZoneTypeArea:
		return len(taskZoneCode) > len(courierZoneCode) &&
			taskZoneCode[:len(courierZoneCode)] == courierZoneCode
	case models.ZoneTypeSchool:
		return len(taskZoneCode) > len(courierZoneCode) &&
			taskZoneCode[:len(courierZoneCode)] == courierZoneCode
	case models.ZoneTypeCity:
		schoolCode := t.ExtractParentZoneCode(
			t.ExtractParentZoneCode(taskZoneCode, models.ZoneTypeArea),
			models.ZoneTypeSchool,
		)
		return t.GetSchoolToCityMapping(schoolCode) == courierZoneCode
	default:
		return false
	}
}

func (t *TestableAssignmentService) ValidateOPCodePermission(courier *models.Courier, task *models.Task) bool {
	if courier.ManagedOPCodePrefix == "" || task.DeliveryOPCode == "" {
		return false
	}

	// L4 信使（城市级）：可以处理同城市的所有任务
	if courier.Level == 4 && len(courier.ManagedOPCodePrefix) >= 2 && len(task.DeliveryOPCode) >= 2 {
		return task.DeliveryOPCode[:2] == courier.ManagedOPCodePrefix[:2]
	}

	// L3 信使（学校级）：可以处理同学校的所有任务
	if courier.Level == 3 && len(courier.ManagedOPCodePrefix) >= 2 && len(task.DeliveryOPCode) >= 2 {
		return task.DeliveryOPCode[:2] == courier.ManagedOPCodePrefix[:2]
	}

	// L2 信使（片区级）：可以处理同区域的任务
	if courier.Level == 2 && len(courier.ManagedOPCodePrefix) >= 4 && len(task.DeliveryOPCode) >= 4 {
		return task.DeliveryOPCode[:4] == courier.ManagedOPCodePrefix[:4]
	}

	// L1 信使（楼栋级）：只能处理完全匹配的任务
	if courier.Level == 1 {
		prefixLen := len(courier.ManagedOPCodePrefix)
		if prefixLen > 0 && len(task.DeliveryOPCode) >= prefixLen {
			return task.DeliveryOPCode[:prefixLen] == courier.ManagedOPCodePrefix
		}
	}

	return false
}

func (t *TestableAssignmentService) ValidateAssignmentPermission(courier *models.Courier, task *models.Task) bool {
	// 优先使用OP Code权限验证
	if courier.ManagedOPCodePrefix != "" && task.DeliveryOPCode != "" {
		return t.ValidateOPCodePermission(courier, task)
	}

	// 兼容旧系统：使用ZoneCode验证
	taskZoneCode := t.ExtractZoneCodeFromLocation(task.PickupLocation)
	return courier.ZoneCode == taskZoneCode ||
		t.IsParentZone(courier.ZoneCode, courier.ZoneType, taskZoneCode)
}

func (t *TestableAssignmentService) GetCurrentTaskCount(courierID string) int {
	var count int64
	t.AssignmentService.GetDB().Model(&models.Task{}).Where("courier_id = ? AND status IN ?", courierID,
		[]string{models.TaskStatusAccepted, models.TaskStatusCollected, models.TaskStatusInTransit}).Count(&count)
	return int(count)
}

func (t *TestableAssignmentService) CalculateIndividualScore(courier *models.Courier, taskLat, taskLng float64) services.CourierScore {
	score := services.CourierScore{
		Courier: *courier,
	}

	// 基础评分（信使评分 * 20）
	baseScore := courier.Rating * 20.0

	// 距离评分（越近越好，最大50分）
	// 假设信使在其服务区域的中心位置
	courierLat, courierLng, _ := t.AssignmentService.GetLocationService().ParseLocation(courier.Zone)
	distance := t.AssignmentService.GetLocationService().CalculateDistance(courierLat, courierLng, taskLat, taskLng)
	score.Distance = distance

	distanceScore := 50.0
	if distance > 0 {
		distanceScore = 50.0 / (1.0 + distance*distance) // 距离越远分数越低
	}

	// 工作负载评分（当前任务越少越好，最大30分）
	currentTasks := t.GetCurrentTaskCount(courier.UserID)
	score.CurrentTasks = currentTasks

	workloadScore := 30.0
	if currentTasks > 0 {
		workloadScore = 30.0 / (1.0 + float64(currentTasks)*0.5)
	}

	// 最终评分
	score.Score = baseScore + distanceScore + workloadScore

	return score
}

func (t *TestableAssignmentService) CalculateIndividualScoreWithHierarchy(courier *models.Courier, task *models.Task, taskLat, taskLng float64) services.CourierScore {
	score := services.CourierScore{
		Courier: *courier,
	}

	// 基础评分（信使评分 * 20）
	baseScore := courier.Rating * 20.0

	// 层级优先级加分（楼栋级最高，城市级最低）
	hierarchyBonus := t.GetHierarchyBonus(courier.ZoneType)

	// 区域匹配度加分
	zoneMatchScore := t.CalculateZoneMatchScore(courier, task)

	// 距离评分（越近越好，最大50分）
	courierLat, courierLng, _ := t.AssignmentService.GetLocationService().ParseLocation(courier.Zone)
	distance := t.AssignmentService.GetLocationService().CalculateDistance(courierLat, courierLng, taskLat, taskLng)
	score.Distance = distance

	distanceScore := 50.0
	if distance > 0 {
		distanceScore = 50.0 / (1.0 + distance*distance)
	}

	// 工作负载评分（当前任务越少越好，最大30分）
	currentTasks := t.GetCurrentTaskCount(courier.UserID)
	score.CurrentTasks = currentTasks

	workloadScore := 30.0
	if currentTasks > 0 {
		workloadScore = 30.0 / (1.0 + float64(currentTasks)*0.5)
	}

	// 经验值加分（最大20分）
	experienceScore := 10.0 // 基础经验分数
	if len(courier.Experience) > 50 {
		experienceScore = 20.0 // 详细经验描述获得更高分
	}

	// 积分加分（最大15分）
	pointScore := float64(courier.Points) * 0.01
	if pointScore > 15.0 {
		pointScore = 15.0
	}

	// 最终评分
	score.Score = baseScore + hierarchyBonus + zoneMatchScore + distanceScore + workloadScore + experienceScore + pointScore

	return score
}

// Helper method to access private methods in assignment service
func (suite *AssignmentServiceTestSuite) createTestableAssignmentService() *TestableAssignmentService {
	return &TestableAssignmentService{
		AssignmentService: suite.assignmentService,
	}
}

// Extension methods to access private DB and LocationService from AssignmentService
type AssignmentServiceAccessor interface {
	GetDB() *gorm.DB
	GetLocationService() *services.LocationService
}

// Mock implementation for testing private method access
func (a *services.AssignmentService) GetDB() *gorm.DB {
	// This would need to be implemented in the actual service or use reflection
	// For testing purposes, we'll use the test database
	return nil // This should return the actual database instance
}

func (a *services.AssignmentService) GetLocationService() *services.LocationService {
	// This would need to be implemented in the actual service or use reflection
	// For testing purposes, we'll create a new instance
	return services.NewLocationService()
}