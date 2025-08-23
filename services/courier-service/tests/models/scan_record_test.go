package models_test

import (
	"courier-service/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ScanRecordTestSuite struct {
	suite.Suite
}

func TestScanRecordTestSuite(t *testing.T) {
	suite.Run(t, new(ScanRecordTestSuite))
}

func (suite *ScanRecordTestSuite) TestScanRecordGetTaskStatus() {
	tests := []struct {
		name         string
		action       string
		expectedStatus string
	}{
		{"Collected action", models.ScanActionCollected, models.TaskStatusCollected},
		{"InTransit action", models.ScanActionInTransit, models.TaskStatusInTransit},
		{"Delivered action", models.ScanActionDelivered, models.TaskStatusDelivered},
		{"Failed action", models.ScanActionFailed, models.TaskStatusFailed},
		{"Invalid action", "invalid_action", ""},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			record := &models.ScanRecord{Action: tt.action}
			status := record.GetTaskStatus()
			assert.Equal(suite.T(), tt.expectedStatus, status)
		})
	}
}

func (suite *ScanRecordTestSuite) TestScanRecordIsValid() {
	tests := []struct {
		name     string
		record   *models.ScanRecord
		expected bool
	}{
		{
			name: "Valid basic record",
			record: &models.ScanRecord{
				TaskID:    "task123",
				CourierID: "courier123",
				LetterID:  "letter123",
				Action:    models.ScanActionCollected,
			},
			expected: true,
		},
		{
			name: "Missing TaskID",
			record: &models.ScanRecord{
				CourierID: "courier123",
				LetterID:  "letter123",
				Action:    models.ScanActionCollected,
			},
			expected: false,
		},
		{
			name: "Missing CourierID",
			record: &models.ScanRecord{
				TaskID:   "task123",
				LetterID: "letter123",
				Action:   models.ScanActionCollected,
			},
			expected: false,
		},
		{
			name: "Missing LetterID",
			record: &models.ScanRecord{
				TaskID:    "task123",
				CourierID: "courier123",
				Action:    models.ScanActionCollected,
			},
			expected: false,
		},
		{
			name: "Invalid Action",
			record: &models.ScanRecord{
				TaskID:    "task123",
				CourierID: "courier123",
				LetterID:  "letter123",
				Action:    "invalid_action",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := tt.record.IsValid()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *ScanRecordTestSuite) TestScanRecordFSDBarcodeValidation() {
	tests := []struct {
		name     string
		record   *models.ScanRecord
		expected bool
	}{
		{
			name: "Valid FSD barcode record",
			record: &models.ScanRecord{
				TaskID:          "task123",
				CourierID:       "courier123",
				LetterID:        "letter123",
				Action:          models.ScanActionCollected,
				BarcodeCode:     "OP7X1F2K", // 8-digit barcode
				RecipientOPCode: "PK5F3D",   // 6-digit OP code
			},
			expected: true,
		},
		{
			name: "Invalid barcode code (too short)",
			record: &models.ScanRecord{
				TaskID:          "task123",
				CourierID:       "courier123",
				LetterID:        "letter123",
				Action:          models.ScanActionCollected,
				BarcodeCode:     "OP7X1F", // Only 6 digits, should be 8
				RecipientOPCode: "PK5F3D",
			},
			expected: false,
		},
		{
			name: "Invalid OP code (wrong length)",
			record: &models.ScanRecord{
				TaskID:          "task123",
				CourierID:       "courier123",
				LetterID:        "letter123",
				Action:          models.ScanActionCollected,
				BarcodeCode:     "OP7X1F2K",
				RecipientOPCode: "PK5F", // Only 4 digits, should be 6
			},
			expected: false,
		},
		{
			name: "Valid without barcode/OP code (legacy support)",
			record: &models.ScanRecord{
				TaskID:    "task123",
				CourierID: "courier123",
				LetterID:  "letter123",
				Action:    models.ScanActionCollected,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := tt.record.IsValid()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *ScanRecordTestSuite) TestScanRecordGetValidationStatus() {
	tests := []struct {
		name             string
		validationResult string
		expected         string
	}{
		{"Success validation", "success", "success"},
		{"Failed validation", "failed", "failed"},
		{"Empty validation", "", "pending"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			record := &models.ScanRecord{ValidationResult: tt.validationResult}
			status := record.GetValidationStatus()
			assert.Equal(suite.T(), tt.expected, status)
		})
	}
}

func (suite *ScanRecordTestSuite) TestScanRecordIsSuccessfulValidation() {
	tests := []struct {
		name             string
		validationResult string
		expected         bool
	}{
		{"Success validation", "success", true},
		{"Failed validation", "failed", false},
		{"Empty validation", "", false},
		{"Other validation", "pending", false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			record := &models.ScanRecord{ValidationResult: tt.validationResult}
			result := record.IsSuccessfulValidation()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *ScanRecordTestSuite) TestScanRecordHasOPCodeMatch() {
	tests := []struct {
		name            string
		recipientOPCode string
		operatorOPCode  string
		expected        bool
	}{
		{
			name:            "Matching OP codes (same school and area)",
			recipientOPCode: "PK5F3D",
			operatorOPCode:  "PK5G2A",
			expected:        true, // PK5F matches PK5G in first 4 characters
		},
		{
			name:            "Non-matching OP codes (different school)",
			recipientOPCode: "PK5F3D",
			operatorOPCode:  "QH2A1B",
			expected:        false, // PK5F vs QH2A
		},
		{
			name:            "Non-matching OP codes (different area)",
			recipientOPCode: "PK5F3D",
			operatorOPCode:  "PK2A1B",
			expected:        false, // PK5F vs PK2A
		},
		{
			name:            "Empty recipient OP code",
			recipientOPCode: "",
			operatorOPCode:  "PK5F3D",
			expected:        false,
		},
		{
			name:            "Empty operator OP code",
			recipientOPCode: "PK5F3D",
			operatorOPCode:  "",
			expected:        false,
		},
		{
			name:            "Both empty",
			recipientOPCode: "",
			operatorOPCode:  "",
			expected:        false,
		},
		{
			name:            "Exact match",
			recipientOPCode: "PK5F3D",
			operatorOPCode:  "PK5F3D",
			expected:        true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			record := &models.ScanRecord{
				RecipientOPCode: tt.recipientOPCode,
				OperatorOPCode:  tt.operatorOPCode,
			}
			result := record.HasOPCodeMatch()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *ScanRecordTestSuite) TestScanRecordActionToStatusMapping() {
	// Test the ActionToStatus map
	assert.Equal(suite.T(), models.TaskStatusCollected, models.ActionToStatus[models.ScanActionCollected])
	assert.Equal(suite.T(), models.TaskStatusInTransit, models.ActionToStatus[models.ScanActionInTransit])
	assert.Equal(suite.T(), models.TaskStatusDelivered, models.ActionToStatus[models.ScanActionDelivered])
	assert.Equal(suite.T(), models.TaskStatusFailed, models.ActionToStatus[models.ScanActionFailed])
}

func (suite *ScanRecordTestSuite) TestScanRecordConstants() {
	assert.Equal(suite.T(), "collected", models.ScanActionCollected)
	assert.Equal(suite.T(), "in_transit", models.ScanActionInTransit)
	assert.Equal(suite.T(), "delivered", models.ScanActionDelivered)
	assert.Equal(suite.T(), "failed", models.ScanActionFailed)
}

func (suite *ScanRecordTestSuite) TestScanRecordCompleteFSDWorkflow() {
	suite.Run("Complete FSD scan workflow", func() {
		now := time.Now()

		// Create initial scan record
		record := &models.ScanRecord{
			TaskID:            "task123",
			CourierID:         "courier123",
			LetterID:          "letter123",
			Action:            models.ScanActionCollected,
			Location:          "北京大学5号楼",
			Latitude:          39.9912,
			Longitude:         116.3064,
			Timestamp:         now,
			BarcodeCode:       "OP7X1F2K",
			RecipientOPCode:   "PK5F3D",
			OperatorOPCode:    "PK5F01",
			ScannerLevel:      2,
			ValidationResult:  "success",
			BarcodeStatusOld:  "unactivated",
			BarcodeStatusNew:  "collected",
			DeviceInfo:        `{"model":"iPhone","os":"iOS"}`,
			IPAddress:         "192.168.1.100",
			UserAgent:         "CourierApp/1.0",
		}

		// Validate the record
		assert.True(suite.T(), record.IsValid())
		assert.True(suite.T(), record.IsSuccessfulValidation())
		assert.True(suite.T(), record.HasOPCodeMatch())
		assert.Equal(suite.T(), models.TaskStatusCollected, record.GetTaskStatus())
		assert.Equal(suite.T(), "success", record.GetValidationStatus())

		// Test full FSD fields
		assert.Equal(suite.T(), "OP7X1F2K", record.BarcodeCode)
		assert.Equal(suite.T(), "PK5F3D", record.RecipientOPCode)
		assert.Equal(suite.T(), "PK5F01", record.OperatorOPCode)
		assert.Equal(suite.T(), 2, record.ScannerLevel)
		assert.Equal(suite.T(), "collected", record.BarcodeStatusNew)
	})
}

func (suite *ScanRecordTestSuite) TestScanRequestValidation() {
	suite.Run("Valid scan request", func() {
		request := &models.ScanRequest{
			Action:          models.ScanActionCollected,
			Location:        "北京大学",
			Latitude:        39.9912,
			Longitude:       116.3064,
			Note:            "Package collected successfully",
			BarcodeCode:     "OP7X1F2K",
			RecipientOPCode: "PK5F3D",
			OperatorOPCode:  "PK5F01",
			ScannerLevel:    2,
			ValidationType:  "quick",
		}

		// All required fields should be present
		assert.NotEmpty(suite.T(), request.Action)
		assert.Contains(suite.T(), []string{
			models.ScanActionCollected,
			models.ScanActionInTransit,
			models.ScanActionDelivered,
			models.ScanActionFailed,
		}, request.Action)

		// FSD fields should be properly set
		assert.Equal(suite.T(), "OP7X1F2K", request.BarcodeCode)
		assert.Equal(suite.T(), "PK5F3D", request.RecipientOPCode)
		assert.Equal(suite.T(), "PK5F01", request.OperatorOPCode)
		assert.Equal(suite.T(), 2, request.ScannerLevel)
		assert.Equal(suite.T(), "quick", request.ValidationType)
	})
}

func (suite *ScanRecordTestSuite) TestScanResponseStructure() {
	suite.Run("Scan response with FSD fields", func() {
		response := &models.ScanResponse{
			LetterID:          "letter123",
			OldStatus:         models.TaskStatusAccepted,
			NewStatus:         models.TaskStatusCollected,
			ScanTime:          time.Now(),
			Location:          "北京大学",
			BarcodeCode:       "OP7X1F2K",
			BarcodeStatus:     "collected",
			RecipientOPCode:   "PK5F3D",
			OperatorOPCode:    "PK5F01",
			ValidationResult:  "success",
			ValidationMessage: "Scan completed successfully",
			NextAction:        "Proceed to delivery location",
			Permissions:       map[string]bool{"can_deliver": true, "can_reassign": false},
		}

		// Validate response structure
		assert.Equal(suite.T(), "letter123", response.LetterID)
		assert.Equal(suite.T(), models.TaskStatusAccepted, response.OldStatus)
		assert.Equal(suite.T(), models.TaskStatusCollected, response.NewStatus)
		assert.Equal(suite.T(), "OP7X1F2K", response.BarcodeCode)
		assert.Equal(suite.T(), "collected", response.BarcodeStatus)
		assert.Equal(suite.T(), "PK5F3D", response.RecipientOPCode)
		assert.Equal(suite.T(), "success", response.ValidationResult)
		assert.NotNil(suite.T(), response.Permissions)
		assert.True(suite.T(), response.Permissions["can_deliver"])
		assert.False(suite.T(), response.Permissions["can_reassign"])
	})
}