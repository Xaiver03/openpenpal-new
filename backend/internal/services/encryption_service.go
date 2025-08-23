package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/logger"
	"openpenpal-backend/internal/models"
	"gorm.io/gorm"
)

// EncryptionService 敏感数据加密服务
type EncryptionService struct {
	config *config.Config
	gcm    cipher.AEAD
	logger *logger.SmartLogger
}

// NewEncryptionService 创建加密服务
func NewEncryptionService(cfg *config.Config, logger *logger.SmartLogger) (*EncryptionService, error) {
	// 生成32字节密钥
	key := sha256.Sum256([]byte(cfg.EncryptionKey))
	
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	
	return &EncryptionService{
		config: cfg,
		gcm:    gcm,
		logger: logger,
	}, nil
}

// EncryptSensitiveData 加密敏感数据
func (s *EncryptionService) EncryptSensitiveData(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	
	// 生成随机nonce
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		s.logger.Error("Failed to generate nonce: %v", err)
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// 加密数据
	ciphertext := s.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Base64编码
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	
	s.logger.Debug("Successfully encrypted sensitive data")
	return encoded, nil
}

// DecryptSensitiveData 解密敏感数据
func (s *EncryptionService) DecryptSensitiveData(encryptedData string) (string, error) {
	if encryptedData == "" {
		return "", nil
	}
	
	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		s.logger.Error("Failed to decode base64 data: %v", err)
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}
	
	// 检查密文长度
	nonceSize := s.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	
	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	// 解密
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		s.logger.Error("Failed to decrypt data: %v", err)
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}
	
	s.logger.Debug("Successfully decrypted sensitive data")
	return string(plaintext), nil
}

// EncryptedUserProfile 加密的用户档案
type EncryptedUserProfile struct {
	UserID             string `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	RealNameEncrypted  string `json:"-" gorm:"type:text;column:real_name_encrypted"`  // 加密的真实姓名
	PhoneEncrypted     string `json:"-" gorm:"type:text;column:phone_encrypted"`     // 加密的手机号
	AddressEncrypted   string `json:"-" gorm:"type:text;column:address_encrypted"`   // 加密的地址
	Bio                string `json:"bio" gorm:"type:text"`                          // 简介不加密
	Preferences        string `json:"preferences" gorm:"type:json"`                  // 偏好设置不加密
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName 指定表名
func (EncryptedUserProfile) TableName() string {
	return "user_profiles"
}

// GetRealName 获取解密后的真实姓名
func (p *EncryptedUserProfile) GetRealName(encService *EncryptionService) (string, error) {
	return encService.DecryptSensitiveData(p.RealNameEncrypted)
}

// SetRealName 设置加密的真实姓名
func (p *EncryptedUserProfile) SetRealName(realName string, encService *EncryptionService) error {
	encrypted, err := encService.EncryptSensitiveData(realName)
	if err != nil {
		return err
	}
	p.RealNameEncrypted = encrypted
	return nil
}

// GetPhone 获取解密后的手机号
func (p *EncryptedUserProfile) GetPhone(encService *EncryptionService) (string, error) {
	return encService.DecryptSensitiveData(p.PhoneEncrypted)
}

// SetPhone 设置加密的手机号
func (p *EncryptedUserProfile) SetPhone(phone string, encService *EncryptionService) error {
	encrypted, err := encService.EncryptSensitiveData(phone)
	if err != nil {
		return err
	}
	p.PhoneEncrypted = encrypted
	return nil
}

// GetAddress 获取解密后的地址
func (p *EncryptedUserProfile) GetAddress(encService *EncryptionService) (string, error) {
	return encService.DecryptSensitiveData(p.AddressEncrypted)
}

// SetAddress 设置加密的地址
func (p *EncryptedUserProfile) SetAddress(address string, encService *EncryptionService) error {
	encrypted, err := encService.EncryptSensitiveData(address)
	if err != nil {
		return err
	}
	p.AddressEncrypted = encrypted
	return nil
}

// ToPublicProfile 转换为公开档案格式（解密敏感字段）
func (p *EncryptedUserProfile) ToPublicProfile(encService *EncryptionService) (*models.UserProfile, error) {
	realName, err := p.GetRealName(encService)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt real name: %w", err)
	}
	
	phone, err := p.GetPhone(encService)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt phone: %w", err)
	}
	
	address, err := p.GetAddress(encService)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt address: %w", err)
	}
	
	return &models.UserProfile{
		UserID:      p.UserID,
		RealName:    realName,
		Phone:       phone,
		Address:     address,
		Bio:         p.Bio,
		Preferences: p.Preferences,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}, nil
}

// FromPublicProfile 从公开档案创建加密档案
func (p *EncryptedUserProfile) FromPublicProfile(profile *models.UserProfile, encService *EncryptionService) error {
	p.UserID = profile.UserID
	p.Bio = profile.Bio
	p.Preferences = profile.Preferences
	p.CreatedAt = profile.CreatedAt
	p.UpdatedAt = profile.UpdatedAt
	
	if err := p.SetRealName(profile.RealName, encService); err != nil {
		return fmt.Errorf("failed to encrypt real name: %w", err)
	}
	
	if err := p.SetPhone(profile.Phone, encService); err != nil {
		return fmt.Errorf("failed to encrypt phone: %w", err)
	}
	
	if err := p.SetAddress(profile.Address, encService); err != nil {
		return fmt.Errorf("failed to encrypt address: %w", err)
	}
	
	return nil
}

// EncryptionMigrationService 数据加密迁移服务
type EncryptionMigrationService struct {
	db         *gorm.DB
	encService *EncryptionService
	logger     *logger.SmartLogger
}

// NewEncryptionMigrationService 创建迁移服务
func NewEncryptionMigrationService(db *gorm.DB, encService *EncryptionService, logger *logger.SmartLogger) *EncryptionMigrationService {
	return &EncryptionMigrationService{
		db:         db,
		encService: encService,
		logger:     logger,
	}
}

// MigrateUserProfiles 迁移用户档案数据到加密存储
func (s *EncryptionMigrationService) MigrateUserProfiles() error {
	s.logger.Info("Starting user profile encryption migration...")
	
	// 检查是否已经有加密列
	if !s.db.Migrator().HasColumn(&EncryptedUserProfile{}, "real_name_encrypted") {
		if err := s.db.Migrator().AddColumn(&EncryptedUserProfile{}, "real_name_encrypted"); err != nil {
			return fmt.Errorf("failed to add real_name_encrypted column: %w", err)
		}
	}
	
	if !s.db.Migrator().HasColumn(&EncryptedUserProfile{}, "phone_encrypted") {
		if err := s.db.Migrator().AddColumn(&EncryptedUserProfile{}, "phone_encrypted"); err != nil {
			return fmt.Errorf("failed to add phone_encrypted column: %w", err)
		}
	}
	
	if !s.db.Migrator().HasColumn(&EncryptedUserProfile{}, "address_encrypted") {
		if err := s.db.Migrator().AddColumn(&EncryptedUserProfile{}, "address_encrypted"); err != nil {
			return fmt.Errorf("failed to add address_encrypted column: %w", err)
		}
	}
	
	// 查询所有未加密的用户档案
	var profiles []models.UserProfile
	if err := s.db.Where("real_name != '' OR phone != '' OR address != ''").Find(&profiles).Error; err != nil {
		return fmt.Errorf("failed to query user profiles: %w", err)
	}
	
	s.logger.Info("Found %d user profiles to encrypt", len(profiles))
	
	// 批量加密处理
	for _, profile := range profiles {
		var encryptedProfile EncryptedUserProfile
		if err := encryptedProfile.FromPublicProfile(&profile, s.encService); err != nil {
			s.logger.Error("Failed to encrypt profile for user %s: %v", profile.UserID, err)
			continue
		}
		
		// 更新数据库
		if err := s.db.Model(&encryptedProfile).Where("user_id = ?", profile.UserID).Updates(map[string]interface{}{
			"real_name_encrypted": encryptedProfile.RealNameEncrypted,
			"phone_encrypted":     encryptedProfile.PhoneEncrypted,
			"address_encrypted":   encryptedProfile.AddressEncrypted,
		}).Error; err != nil {
			s.logger.Error("Failed to update encrypted profile for user %s: %v", profile.UserID, err)
			continue
		}
		
		s.logger.Debug("Successfully encrypted profile for user: %s", profile.UserID)
	}
	
	s.logger.Info("User profile encryption migration completed")
	return nil
}