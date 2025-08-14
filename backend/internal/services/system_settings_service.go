package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemSettingsService 系统设置服务
type SystemSettingsService struct {
	db     *gorm.DB
	config *config.Config
}

// NewSystemSettingsService 创建系统设置服务实例
func NewSystemSettingsService(db *gorm.DB, config *config.Config) *SystemSettingsService {
	return &SystemSettingsService{
		db:     db,
		config: config,
	}
}

// GetSystemConfig 获取完整的系统配置
func (s *SystemSettingsService) GetSystemConfig() (*models.SystemConfig, error) {
	// 从数据库加载所有配置项
	var settings []models.SystemSettings
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	// 如果数据库中没有配置，返回默认配置
	if len(settings) == 0 {
		return models.DefaultSystemConfig(), nil
	}

	// 将配置项转换为 SystemConfig 结构体
	config := models.DefaultSystemConfig()
	for _, setting := range settings {
		s.applySetting(config, setting)
	}

	return config, nil
}

// UpdateSystemConfig 更新系统配置
func (s *SystemSettingsService) UpdateSystemConfig(newConfig *models.SystemConfig) (*models.SystemConfig, error) {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 将 SystemConfig 转换为配置项并保存
	settings := s.configToSettings(newConfig)
	for _, setting := range settings {
		// 检查是否已存在
		var existing models.SystemSettings
		err := tx.Where("key = ?", setting.Key).First(&existing).Error

		if err == nil {
			// 更新现有配置
			existing.Value = setting.Value
			existing.DataType = setting.DataType
			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update setting %s: %w", setting.Key, err)
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新配置
			setting.ID = uuid.New().String()
			if err := tx.Create(&setting).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create setting %s: %w", setting.Key, err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("database error for setting %s: %w", setting.Key, err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit settings: %w", err)
	}

	// 返回更新后的配置
	return s.GetSystemConfig()
}

// ResetToDefaults 重置为默认配置
func (s *SystemSettingsService) ResetToDefaults() (*models.SystemConfig, error) {
	// 清空所有配置
	if err := s.db.Exec("DELETE FROM system_settings").Error; err != nil {
		return nil, fmt.Errorf("failed to clear settings: %w", err)
	}

	// 返回默认配置
	return models.DefaultSystemConfig(), nil
}

// GetSetting 获取单个配置项
func (s *SystemSettingsService) GetSetting(key string) (string, error) {
	var setting models.SystemSettings
	err := s.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("setting not found: %s", key)
		}
		return "", fmt.Errorf("failed to get setting: %w", err)
	}
	return setting.Value, nil
}

// SetSetting 设置单个配置项
func (s *SystemSettingsService) SetSetting(key, value, category, dataType string) error {
	var setting models.SystemSettings
	err := s.db.Where("key = ?", key).First(&setting).Error

	if err == nil {
		// 更新现有配置
		setting.Value = value
		if category != "" {
			setting.Category = category
		}
		if dataType != "" {
			setting.DataType = dataType
		}
		return s.db.Save(&setting).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新配置
		setting = models.SystemSettings{
			ID:       uuid.New().String(),
			Key:      key,
			Value:    value,
			Category: category,
			DataType: dataType,
		}
		return s.db.Create(&setting).Error
	}

	return fmt.Errorf("failed to set setting: %w", err)
}

// applySetting 将单个配置项应用到 SystemConfig
func (s *SystemSettingsService) applySetting(config *models.SystemConfig, setting models.SystemSettings) {
	switch setting.Key {
	// 基本设置
	case models.KeySiteName:
		config.SiteName = setting.Value
	case models.KeySiteDescription:
		config.SiteDescription = setting.Value
	case models.KeySiteLogo:
		config.SiteLogo = setting.Value
	case models.KeyMaintenanceMode:
		config.MaintenanceMode = s.parseBool(setting.Value)

	// 邮件设置
	case models.KeySMTPHost:
		config.SMTPHost = setting.Value
	case models.KeySMTPPort:
		config.SMTPPort = s.parseInt(setting.Value, 587)
	case models.KeySMTPUsername:
		config.SMTPUsername = setting.Value
	case models.KeySMTPPassword:
		config.SMTPPassword = setting.Value
	case models.KeySMTPEncryption:
		config.SMTPEncryption = setting.Value
	case models.KeyEmailFromName:
		config.EmailFromName = setting.Value
	case models.KeyEmailFromAddress:
		config.EmailFromAddress = setting.Value

	// 信件设置
	case models.KeyMaxLetterLength:
		config.MaxLetterLength = s.parseInt(setting.Value, 5000)
	case models.KeyAllowedFileTypes:
		config.AllowedFileTypes = s.parseStringArray(setting.Value)
	case models.KeyMaxFileSize:
		config.MaxFileSize = s.parseInt(setting.Value, 10)
	case models.KeyLetterReviewRequired:
		config.LetterReviewRequired = s.parseBool(setting.Value)
	case models.KeyAutoDeliveryEnabled:
		config.AutoDeliveryEnabled = s.parseBool(setting.Value)

	// 用户设置
	case models.KeyUserRegistrationEnabled:
		config.UserRegistrationEnabled = s.parseBool(setting.Value)
	case models.KeyEmailVerificationRequired:
		config.EmailVerificationRequired = s.parseBool(setting.Value)
	case models.KeyMaxUsersPerSchool:
		config.MaxUsersPerSchool = s.parseInt(setting.Value, 10000)
	case models.KeyUserInactiveDays:
		config.UserInactiveDays = s.parseInt(setting.Value, 90)

	// 信使设置
	case models.KeyCourierApplicationEnabled:
		config.CourierApplicationEnabled = s.parseBool(setting.Value)
	case models.KeyCourierAutoApproval:
		config.CourierAutoApproval = s.parseBool(setting.Value)
	case models.KeyMaxDeliveryDistance:
		config.MaxDeliveryDistance = s.parseInt(setting.Value, 10)
	case models.KeyCourierRatingRequired:
		config.CourierRatingRequired = s.parseBool(setting.Value)

	// 安全设置
	case models.KeyPasswordMinLength:
		config.PasswordMinLength = s.parseInt(setting.Value, 6)
	case models.KeyPasswordRequireSymbols:
		config.PasswordRequireSymbols = s.parseBool(setting.Value)
	case models.KeyPasswordRequireNumbers:
		config.PasswordRequireNumbers = s.parseBool(setting.Value)
	case models.KeySessionTimeout:
		config.SessionTimeout = s.parseInt(setting.Value, 3600)
	case models.KeyMaxLoginAttempts:
		config.MaxLoginAttempts = s.parseInt(setting.Value, 5)

	// 通知设置
	case models.KeyEmailNotifications:
		config.EmailNotifications = s.parseBool(setting.Value)
	case models.KeySMSNotifications:
		config.SMSNotifications = s.parseBool(setting.Value)
	case models.KeyPushNotifications:
		config.PushNotifications = s.parseBool(setting.Value)
	case models.KeyAdminNotifications:
		config.AdminNotifications = s.parseBool(setting.Value)
	}
}

// configToSettings 将 SystemConfig 转换为配置项列表
func (s *SystemSettingsService) configToSettings(config *models.SystemConfig) []models.SystemSettings {
	settings := []models.SystemSettings{
		// 基本设置
		{Key: models.KeySiteName, Value: config.SiteName, Category: models.CategoryGeneral, DataType: "string"},
		{Key: models.KeySiteDescription, Value: config.SiteDescription, Category: models.CategoryGeneral, DataType: "string"},
		{Key: models.KeySiteLogo, Value: config.SiteLogo, Category: models.CategoryGeneral, DataType: "string"},
		{Key: models.KeyMaintenanceMode, Value: s.boolToString(config.MaintenanceMode), Category: models.CategoryGeneral, DataType: "boolean"},

		// 邮件设置
		{Key: models.KeySMTPHost, Value: config.SMTPHost, Category: models.CategoryEmail, DataType: "string"},
		{Key: models.KeySMTPPort, Value: strconv.Itoa(config.SMTPPort), Category: models.CategoryEmail, DataType: "number"},
		{Key: models.KeySMTPUsername, Value: config.SMTPUsername, Category: models.CategoryEmail, DataType: "string"},
		{Key: models.KeySMTPPassword, Value: config.SMTPPassword, Category: models.CategoryEmail, DataType: "string"},
		{Key: models.KeySMTPEncryption, Value: config.SMTPEncryption, Category: models.CategoryEmail, DataType: "string"},
		{Key: models.KeyEmailFromName, Value: config.EmailFromName, Category: models.CategoryEmail, DataType: "string"},
		{Key: models.KeyEmailFromAddress, Value: config.EmailFromAddress, Category: models.CategoryEmail, DataType: "string"},

		// 信件设置
		{Key: models.KeyMaxLetterLength, Value: strconv.Itoa(config.MaxLetterLength), Category: models.CategoryLetter, DataType: "number"},
		{Key: models.KeyAllowedFileTypes, Value: s.stringArrayToJSON(config.AllowedFileTypes), Category: models.CategoryLetter, DataType: "json"},
		{Key: models.KeyMaxFileSize, Value: strconv.Itoa(config.MaxFileSize), Category: models.CategoryLetter, DataType: "number"},
		{Key: models.KeyLetterReviewRequired, Value: s.boolToString(config.LetterReviewRequired), Category: models.CategoryLetter, DataType: "boolean"},
		{Key: models.KeyAutoDeliveryEnabled, Value: s.boolToString(config.AutoDeliveryEnabled), Category: models.CategoryLetter, DataType: "boolean"},

		// 用户设置
		{Key: models.KeyUserRegistrationEnabled, Value: s.boolToString(config.UserRegistrationEnabled), Category: models.CategoryUser, DataType: "boolean"},
		{Key: models.KeyEmailVerificationRequired, Value: s.boolToString(config.EmailVerificationRequired), Category: models.CategoryUser, DataType: "boolean"},
		{Key: models.KeyMaxUsersPerSchool, Value: strconv.Itoa(config.MaxUsersPerSchool), Category: models.CategoryUser, DataType: "number"},
		{Key: models.KeyUserInactiveDays, Value: strconv.Itoa(config.UserInactiveDays), Category: models.CategoryUser, DataType: "number"},

		// 信使设置
		{Key: models.KeyCourierApplicationEnabled, Value: s.boolToString(config.CourierApplicationEnabled), Category: models.CategoryCourier, DataType: "boolean"},
		{Key: models.KeyCourierAutoApproval, Value: s.boolToString(config.CourierAutoApproval), Category: models.CategoryCourier, DataType: "boolean"},
		{Key: models.KeyMaxDeliveryDistance, Value: strconv.Itoa(config.MaxDeliveryDistance), Category: models.CategoryCourier, DataType: "number"},
		{Key: models.KeyCourierRatingRequired, Value: s.boolToString(config.CourierRatingRequired), Category: models.CategoryCourier, DataType: "boolean"},

		// 安全设置
		{Key: models.KeyPasswordMinLength, Value: strconv.Itoa(config.PasswordMinLength), Category: models.CategorySecurity, DataType: "number"},
		{Key: models.KeyPasswordRequireSymbols, Value: s.boolToString(config.PasswordRequireSymbols), Category: models.CategorySecurity, DataType: "boolean"},
		{Key: models.KeyPasswordRequireNumbers, Value: s.boolToString(config.PasswordRequireNumbers), Category: models.CategorySecurity, DataType: "boolean"},
		{Key: models.KeySessionTimeout, Value: strconv.Itoa(config.SessionTimeout), Category: models.CategorySecurity, DataType: "number"},
		{Key: models.KeyMaxLoginAttempts, Value: strconv.Itoa(config.MaxLoginAttempts), Category: models.CategorySecurity, DataType: "number"},

		// 通知设置
		{Key: models.KeyEmailNotifications, Value: s.boolToString(config.EmailNotifications), Category: models.CategoryNotification, DataType: "boolean"},
		{Key: models.KeySMSNotifications, Value: s.boolToString(config.SMSNotifications), Category: models.CategoryNotification, DataType: "boolean"},
		{Key: models.KeyPushNotifications, Value: s.boolToString(config.PushNotifications), Category: models.CategoryNotification, DataType: "boolean"},
		{Key: models.KeyAdminNotifications, Value: s.boolToString(config.AdminNotifications), Category: models.CategoryNotification, DataType: "boolean"},
	}

	return settings
}

// 辅助方法

func (s *SystemSettingsService) parseBool(value string) bool {
	return value == "true" || value == "1" || value == "yes" || value == "on"
}

func (s *SystemSettingsService) boolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (s *SystemSettingsService) parseInt(value string, defaultValue int) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

func (s *SystemSettingsService) parseStringArray(value string) []string {
	// 尝试解析为 JSON 数组
	var arr []string
	if err := json.Unmarshal([]byte(value), &arr); err == nil {
		return arr
	}
	// 否则按逗号分隔
	return strings.Split(value, ",")
}

func (s *SystemSettingsService) stringArrayToJSON(arr []string) string {
	data, _ := json.Marshal(arr)
	return string(data)
}

// ValidateEmailConfig 验证邮件配置
func (s *SystemSettingsService) ValidateEmailConfig(config *models.SystemConfig) error {
	if config.SMTPHost == "" {
		return errors.New("SMTP host is required")
	}
	if config.SMTPPort <= 0 || config.SMTPPort > 65535 {
		return errors.New("invalid SMTP port")
	}
	if config.EmailFromAddress == "" {
		return errors.New("from email address is required")
	}
	// 验证邮箱格式
	if !strings.Contains(config.EmailFromAddress, "@") {
		return errors.New("invalid from email address format")
	}
	return nil
}
