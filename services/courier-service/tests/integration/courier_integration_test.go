package integration_test

import (
	"bytes"
	"courier-service/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CourierIntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
}

func TestCourierIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CourierIntegrationTestSuite))
}

func (suite *CourierIntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	
	suite.db = db
	
	// Migrate test models
	err = db.AutoMigrate(
		&models.Task{},
		&models.Courier{},
		&models.ScanRecord{},
	)
	suite.Require().NoError(err)
	
	// Setup test router
	suite.router = gin.New()
	suite.setupRoutes()
	
	// Seed test data
	suite.seedTestData()
}

func (suite *CourierIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

func (suite *CourierIntegrationTestSuite) setupRoutes() {
	// Mock courier service routes
	api := suite.router.Group("/api/v1")
	{
		api.GET("/tasks", suite.handleGetTasks)
		api.POST("/tasks/:id/accept", suite.handleAcceptTask)
		api.POST("/scan/:code", suite.handleScanCode)
		api.GET("/courier/hierarchy/level/:level", suite.handleGetCouriersByLevel)
		api.GET("/courier/stats/:id", suite.handleGetCourierStats)
	}
}

func (suite *CourierIntegrationTestSuite) seedTestData() {
	// Create test couriers
	couriers := []models.Courier{
		{
			UserID:              "courier_l1",
			Level:               1,
			Rating:              4.0,
			Zone:                "北京大学5号楼",
			ZoneType:            models.ZoneTypeBuilding,
			ZoneCode:            "PKU001-AREA001-BUILDING001",
			ManagedOPCodePrefix: "PK5F3D",
			Status:              models.CourierStatusApproved,
			Points:              500,
			Experience:          "Building courier with 6 months experience",
		},
		{
			UserID:              "courier_l2",
			Level:               2,
			Rating:              4.2,
			Zone:                "北京大学A区",
			ZoneType:            models.ZoneTypeArea,
			ZoneCode:            "PKU001-AREA001",
			ManagedOPCodePrefix: "PK5F",
			Status:              models.CourierStatusApproved,
			Points:              1000,
			Experience:          "Area courier with 1 year experience",
		},
		{
			UserID:              "courier_l3",
			Level:               3,
			Rating:              4.5,
			Zone:                "北京大学",
			ZoneType:            models.ZoneTypeSchool,
			ZoneCode:            "PKU001",
			ManagedOPCodePrefix: "PK",
			Status:              models.CourierStatusApproved,
			Points:              1500,
			Experience:          "School courier with 2 years experience",
		},
		{
			UserID:              "courier_l4",
			Level:               4,
			Rating:              4.8,
			Zone:                "北京市",
			ZoneType:            models.ZoneTypeCity,
			ZoneCode:            "BJ001",
			ManagedOPCodePrefix: "PK",
			Status:              models.CourierStatusApproved,
			Points:              2000,
			Experience:          "City director with 3+ years experience",
		},
	}
	
	for _, courier := range couriers {
		suite.db.Create(&courier)
	}
	
	// Create test tasks
	tasks := []models.Task{
		{
			TaskID:           "task_001",
			LetterID:         "letter_001",
			PickupLocation:   "北京大学5号楼301",
			DeliveryLocation: "北京大学5号楼302",
			PickupOPCode:     "PK5F3D",
			DeliveryOPCode:   "PK5F3E", 
			Status:           models.TaskStatusAvailable,
			Priority:         models.TaskPriorityNormal,
			Reward:           5.0,
			CreatedAt:        time.Now(),
		},
		{
			TaskID:           "task_002", 
			LetterID:         "letter_002",
			PickupLocation:   "北京大学A区",
			DeliveryLocation: "北京大学B区",
			PickupOPCode:     "PK5F01",
			DeliveryOPCode:   "PK2A01",
			Status:           models.TaskStatusAvailable,
			Priority:         models.TaskPriorityUrgent,
			Reward:           8.0,
			CreatedAt:        time.Now(),
		},
		{
			TaskID:           "task_003",
			LetterID:         "letter_003",
			PickupLocation:   "清华大学",
			DeliveryLocation: "北京大学",
			PickupOPCode:     "QH2A01",
			DeliveryOPCode:   "PK5F01",
			Status:           models.TaskStatusAvailable,
			Priority:         models.TaskPriorityExpress,
			Reward:           12.0,
			CreatedAt:        time.Now(),
		},
	}
	
	for _, task := range tasks {
		suite.db.Create(&task)
	}
}

func (suite *CourierIntegrationTestSuite) TestGetTasks() {
	suite.Run("Get all available tasks", func() {
		req, _ := http.NewRequest("GET", "/api/v1/tasks?status=available", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response struct {
			Tasks []models.Task `json:"tasks"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.GreaterOrEqual(suite.T(), len(response.Tasks), 3)
	})
	
	suite.Run("Get tasks with priority filter", func() {
		req, _ := http.NewRequest("GET", "/api/v1/tasks?priority=urgent", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response struct {
			Tasks []models.Task `json:"tasks"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		
		for _, task := range response.Tasks {
			assert.Equal(suite.T(), models.TaskPriorityUrgent, task.Priority)
		}
	})
}

func (suite *CourierIntegrationTestSuite) TestAcceptTask() {
	suite.Run("L1 courier accepts appropriate task", func() {
		requestBody := map[string]string{
			"estimated_time": "30分钟内",
			"note":          "Ready to deliver",
		}
		jsonBody, _ := json.Marshal(requestBody)
		
		req, _ := http.NewRequest("POST", "/api/v1/tasks/task_001/accept", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Courier-ID", "courier_l1") // Mock auth
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// Verify task was updated in database
		var task models.Task
		err := suite.db.Where("task_id = ?", "task_001").First(&task).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), models.TaskStatusAccepted, task.Status)
		assert.NotNil(suite.T(), task.CourierID)
		assert.Equal(suite.T(), "courier_l1", *task.CourierID)
	})
	
	suite.Run("L1 courier cannot accept cross-area task", func() {
		requestBody := map[string]string{
			"estimated_time": "1小时内",
			"note":          "Cross area delivery",
		}
		jsonBody, _ := json.Marshal(requestBody)
		
		req, _ := http.NewRequest("POST", "/api/v1/tasks/task_002/accept", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Courier-ID", "courier_l1") // Mock auth
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusForbidden, w.Code)
		
		var response struct {
			Error string `json:"error"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), response.Error, "权限不足")
	})
}

func (suite *CourierIntegrationTestSuite) TestScanCode() {
	// First accept a task
	courierID := "courier_l1"
	task := models.Task{
		TaskID:    "task_scan",
		LetterID:  "letter_scan",
		CourierID: &courierID,
		Status:    models.TaskStatusAccepted,
		PickupOPCode: "PK5F3D",
		DeliveryOPCode: "PK5F3E",
	}
	suite.db.Create(&task)
	
	suite.Run("Successful scan for collection", func() {
		scanRequest := models.ScanRequest{
			Action:          models.ScanActionCollected,
			Location:        "北京大学5号楼301",
			Latitude:        39.9912,
			Longitude:       116.3064,
			Note:           "Package collected",
			BarcodeCode:    "OP7X1F2K",
			RecipientOPCode: "PK5F3D",
			OperatorOPCode:  "PK5F3D",
			ScannerLevel:   1,
			ValidationType: "quick",
		}
		jsonBody, _ := json.Marshal(scanRequest)
		
		req, _ := http.NewRequest("POST", "/api/v1/scan/OP7X1F2K", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Courier-ID", "courier_l1")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response models.ScanResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "success", response.ValidationResult)
		assert.Equal(suite.T(), models.TaskStatusCollected, response.NewStatus)
		
		// Verify scan record was created
		var scanRecord models.ScanRecord
		err = suite.db.Where("barcode_code = ?", "OP7X1F2K").First(&scanRecord).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), models.ScanActionCollected, scanRecord.Action)
	})
	
	suite.Run("Invalid OP code mismatch", func() {
		scanRequest := models.ScanRequest{
			Action:          models.ScanActionCollected,
			Location:        "北京大学5号楼301",
			BarcodeCode:    "OP7X1F2L",
			RecipientOPCode: "QH2A01", // Different school
			OperatorOPCode:  "PK5F3D",
			ScannerLevel:   1,
		}
		jsonBody, _ := json.Marshal(scanRequest)
		
		req, _ := http.NewRequest("POST", "/api/v1/scan/OP7X1F2L", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Courier-ID", "courier_l1")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
		
		var response struct {
			Error string `json:"error"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), response.Error, "OP Code验证失败")
	})
}

func (suite *CourierIntegrationTestSuite) TestGetCouriersByLevel() {
	suite.Run("Get L2 couriers", func() {
		req, _ := http.NewRequest("GET", "/api/v1/courier/hierarchy/level/2", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response struct {
			Couriers []models.Courier `json:"couriers"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		
		for _, courier := range response.Couriers {
			assert.Equal(suite.T(), 2, courier.Level)
			assert.Equal(suite.T(), models.ZoneTypeArea, courier.ZoneType)
		}
	})
	
	suite.Run("Get L4 couriers (city directors)", func() {
		req, _ := http.NewRequest("GET", "/api/v1/courier/hierarchy/level/4", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response struct {
			Couriers []models.Courier `json:"couriers"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		
		assert.Len(suite.T(), response.Couriers, 1)
		assert.Equal(suite.T(), 4, response.Couriers[0].Level)
		assert.Equal(suite.T(), models.ZoneTypeCity, response.Couriers[0].ZoneType)
	})
}

func (suite *CourierIntegrationTestSuite) TestGetCourierStats() {
	suite.Run("Get courier statistics", func() {
		req, _ := http.NewRequest("GET", "/api/v1/courier/stats/courier_l1", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response struct {
			CourierID     string  `json:"courier_id"`
			Level         int     `json:"level"`
			Rating        float64 `json:"rating"`
			TotalTasks    int     `json:"total_tasks"`
			CompletedTasks int    `json:"completed_tasks"`
			Points        int     `json:"points"`
			Zone          string  `json:"zone"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		
		assert.Equal(suite.T(), "courier_l1", response.CourierID)
		assert.Equal(suite.T(), 1, response.Level)
		assert.Equal(suite.T(), 4.0, response.Rating)
		assert.Equal(suite.T(), 500, response.Points)
	})
}

func (suite *CourierIntegrationTestSuite) TestCompleteTaskWorkflow() {
	suite.Run("Complete end-to-end task workflow", func() {
		// 1. Create a new task
		task := models.Task{
			TaskID:           "workflow_task",
			LetterID:         "workflow_letter",
			PickupLocation:   "北京大学5号楼",
			DeliveryLocation: "北京大学5号楼",
			PickupOPCode:     "PK5F3D",
			DeliveryOPCode:   "PK5F3E",
			Status:           models.TaskStatusAvailable,
			Priority:         models.TaskPriorityNormal,
			Reward:           6.0,
		}
		suite.db.Create(&task)
		
		// 2. Accept the task
		acceptRequest := map[string]string{
			"estimated_time": "30分钟内",
			"note":          "Ready for delivery",
		}
		jsonBody, _ := json.Marshal(acceptRequest)
		
		req, _ := http.NewRequest("POST", "/api/v1/tasks/workflow_task/accept", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Courier-ID", "courier_l1")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// 3. Scan for collection
		collectScan := models.ScanRequest{
			Action:          models.ScanActionCollected,
			Location:        "北京大学5号楼",
			BarcodeCode:    "OP7X1F3K",
			RecipientOPCode: "PK5F3D",
			OperatorOPCode:  "PK5F3D",
			ScannerLevel:   1,
			ValidationType: "full",
		}
		jsonBody, _ = json.Marshal(collectScan)
		
		req, _ = http.NewRequest("POST", "/api/v1/scan/OP7X1F3K", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Courier-ID", "courier_l1")
		w = httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// 4. Scan for in-transit
		transitScan := models.ScanRequest{
			Action:          models.ScanActionInTransit,
			Location:        "途中",
			BarcodeCode:    "OP7X1F3K",
			RecipientOPCode: "PK5F3E",
			OperatorOPCode:  "PK5F3D",
			ScannerLevel:   1,
		}
		jsonBody, _ = json.Marshal(transitScan)
		
		req, _ = http.NewRequest("POST", "/api/v1/scan/OP7X1F3K", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Courier-ID", "courier_l1")
		w = httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// 5. Scan for delivery
		deliveryScan := models.ScanRequest{
			Action:          models.ScanActionDelivered,
			Location:        "北京大学5号楼302",
			BarcodeCode:    "OP7X1F3K",
			RecipientOPCode: "PK5F3E",
			OperatorOPCode:  "PK5F3D",
			ScannerLevel:   1,
		}
		jsonBody, _ = json.Marshal(deliveryScan)
		
		req, _ = http.NewRequest("POST", "/api/v1/scan/OP7X1F3K", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json") 
		req.Header.Set("X-Courier-ID", "courier_l1")
		w = httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// 6. Verify final task status
		var finalTask models.Task
		err := suite.db.Where("task_id = ?", "workflow_task").First(&finalTask).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), models.TaskStatusDelivered, finalTask.Status)
		assert.NotNil(suite.T(), finalTask.CompletedAt)
		
		// 7. Verify scan records exist
		var scanRecords []models.ScanRecord
		err = suite.db.Where("task_id = ?", "workflow_task").Find(&scanRecords).Error
		assert.NoError(suite.T(), err)
		assert.GreaterOrEqual(suite.T(), len(scanRecords), 3) // collected, in_transit, delivered
	})
}

// Mock handler implementations
func (suite *CourierIntegrationTestSuite) handleGetTasks(c *gin.Context) {
	status := c.Query("status")
	priority := c.Query("priority")
	
	query := suite.db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	
	var tasks []models.Task
	query.Find(&tasks)
	
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (suite *CourierIntegrationTestSuite) handleAcceptTask(c *gin.Context) {
	taskID := c.Param("id")
	courierID := c.GetHeader("X-Courier-ID")
	
	var task models.Task
	if err := suite.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	
	// Check courier permissions
	var courier models.Courier
	if err := suite.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Courier not found"})
		return
	}
	
	// Simple permission check (L1 can only handle same building)
	if courier.Level == 1 && task.DeliveryOPCode != "" {
		if len(courier.ManagedOPCodePrefix) > 0 && 
		   !suite.hasPermission(courier.ManagedOPCodePrefix, task.DeliveryOPCode, courier.Level) {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
	}
	
	// Accept the task
	now := time.Now()
	task.CourierID = &courierID
	task.Status = models.TaskStatusAccepted
	task.AcceptedAt = &now
	
	suite.db.Save(&task)
	
	c.JSON(http.StatusOK, gin.H{"message": "Task accepted", "task": task})
}

func (suite *CourierIntegrationTestSuite) handleScanCode(c *gin.Context) {
	code := c.Param("code")
	courierID := c.GetHeader("X-Courier-ID")
	
	var scanRequest models.ScanRequest
	if err := c.ShouldBindJSON(&scanRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate OP Code permissions
	if scanRequest.RecipientOPCode != "" && scanRequest.OperatorOPCode != "" {
		if !suite.validateOPCodeMatch(scanRequest.RecipientOPCode, scanRequest.OperatorOPCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OP Code验证失败"})
			return
		}
	}
	
	// Create scan record
	scanRecord := models.ScanRecord{
		TaskID:            "workflow_task", // Simplified for testing
		CourierID:         courierID,
		LetterID:          "workflow_letter",
		Action:            scanRequest.Action,
		Location:          scanRequest.Location,
		Latitude:          scanRequest.Latitude,
		Longitude:         scanRequest.Longitude,
		Note:             scanRequest.Note,
		BarcodeCode:      scanRequest.BarcodeCode,
		RecipientOPCode:  scanRequest.RecipientOPCode,
		OperatorOPCode:   scanRequest.OperatorOPCode,
		ScannerLevel:     scanRequest.ScannerLevel,
		ValidationResult: "success",
		Timestamp:        time.Now(),
		CreatedAt:        time.Now(),
	}
	
	suite.db.Create(&scanRecord)
	
	// Update task status if applicable
	newStatus := models.ActionToStatus[scanRequest.Action]
	if newStatus != "" {
		var task models.Task
		if err := suite.db.Where("task_id = ?", "workflow_task").First(&task).Error; err == nil {
			task.Status = newStatus
			if newStatus == models.TaskStatusDelivered {
				now := time.Now()
				task.CompletedAt = &now
			}
			suite.db.Save(&task)
		}
	}
	
	response := models.ScanResponse{
		LetterID:         scanRecord.LetterID,
		OldStatus:        models.TaskStatusAccepted,
		NewStatus:        newStatus,
		ScanTime:         scanRecord.Timestamp,
		Location:         scanRecord.Location,
		BarcodeCode:      scanRecord.BarcodeCode,
		ValidationResult: "success",
		NextAction:       suite.getNextAction(scanRequest.Action),
	}
	
	c.JSON(http.StatusOK, response)
}

func (suite *CourierIntegrationTestSuite) handleGetCouriersByLevel(c *gin.Context) {
	level := c.Param("level")
	
	var couriers []models.Courier
	suite.db.Where("level = ? AND status = ?", level, models.CourierStatusApproved).Find(&couriers)
	
	c.JSON(http.StatusOK, gin.H{"couriers": couriers})
}

func (suite *CourierIntegrationTestSuite) handleGetCourierStats(c *gin.Context) {
	courierID := c.Param("id")
	
	var courier models.Courier
	if err := suite.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Courier not found"})
		return
	}
	
	// Get task statistics
	var totalTasks int64
	var completedTasks int64
	
	suite.db.Model(&models.Task{}).Where("courier_id = ?", courierID).Count(&totalTasks)
	suite.db.Model(&models.Task{}).Where("courier_id = ? AND status = ?", courierID, models.TaskStatusDelivered).Count(&completedTasks)
	
	stats := gin.H{
		"courier_id":      courier.UserID,
		"level":          courier.Level,
		"rating":         courier.Rating,
		"total_tasks":    totalTasks,
		"completed_tasks": completedTasks,
		"points":         courier.Points,
		"zone":           courier.Zone,
	}
	
	c.JSON(http.StatusOK, stats)
}

// Helper functions
func (suite *CourierIntegrationTestSuite) hasPermission(prefix, opcode string, level int) bool {
	switch level {
	case 1:
		return opcode == prefix
	case 2:
		return len(opcode) >= len(prefix) && opcode[:len(prefix)] == prefix
	case 3, 4:
		return len(opcode) >= 2 && len(prefix) >= 2 && opcode[:2] == prefix[:2]
	}
	return false
}

func (suite *CourierIntegrationTestSuite) validateOPCodeMatch(recipient, operator string) bool {
	if len(recipient) >= 4 && len(operator) >= 4 {
		return recipient[:4] == operator[:4] // School + Area match
	}
	return false
}

func (suite *CourierIntegrationTestSuite) getNextAction(currentAction string) string {
	switch currentAction {
	case models.ScanActionCollected:
		return "proceed_to_delivery"
	case models.ScanActionInTransit:
		return "deliver_package"
	case models.ScanActionDelivered:
		return "task_completed"
	default:
		return "unknown"
	}
}