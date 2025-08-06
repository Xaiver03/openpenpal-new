package services

import (
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StorageService 文件存储服务
type StorageService struct {
	db     *gorm.DB
	config *config.Config
}

// StorageProvider 存储提供商接口
type StorageProvider interface {
	Upload(file *multipart.FileHeader, objectKey string) (*UploadResult, error)
	Download(objectKey string) (io.ReadCloser, error)
	Delete(objectKey string) error
	GetPublicURL(objectKey string) string
	GetPrivateURL(objectKey string, expiration time.Duration) string
	GenerateThumbnail(objectKey string, width, height int) (string, error)
}

// UploadResult 上传结果
type UploadResult struct {
	ObjectKey    string
	PublicURL    string
	PrivateURL   string
	ThumbnailURL string
	FileSize     int64
	MD5Hash      string
	SHA256Hash   string
}

// LocalStorageProvider 本地存储提供商
type LocalStorageProvider struct {
	basePath  string
	baseURL   string
	uploadDir string
}

// AliOSSProvider 阿里云OSS存储提供商
type AliOSSProvider struct {
	endpoint        string
	accessKeyID     string
	accessKeySecret string
	bucketName      string
	baseURL         string
}

// TencentCOSProvider 腾讯云COS存储提供商
type TencentCOSProvider struct {
	region     string
	secretID   string
	secretKey  string
	bucketName string
	baseURL    string
}

// NewStorageService 创建存储服务
func NewStorageService(db *gorm.DB, config *config.Config) *StorageService {
	return &StorageService{
		db:     db,
		config: config,
	}
}

// UploadFile 上传文件
func (s *StorageService) UploadFile(file *multipart.FileHeader, req *models.UploadRequest, userID string) (*models.UploadResponse, error) {
	// 验证文件
	if err := s.validateFile(file, req.Category); err != nil {
		return nil, err
	}

	// 获取默认存储提供商
	provider, err := s.getDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取存储提供商失败: %w", err)
	}

	// 生成文件ID和对象键
	fileID := uuid.New().String()
	objectKey := s.generateObjectKey(fileID, file.Filename, req.Category)

	// 上传文件
	result, err := s.uploadToProvider(provider, file, objectKey)
	if err != nil {
		return nil, fmt.Errorf("文件上传失败: %w", err)
	}

	// 设置过期时间
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &expiry
	}

	// 保存文件记录
	storageFile := &models.StorageFile{
		ID:           fileID,
		FileName:     objectKey,
		OriginalName: file.Filename,
		FileSize:     result.FileSize,
		MimeType:     file.Header.Get("Content-Type"),
		Extension:    strings.ToLower(filepath.Ext(file.Filename)),
		Category:     req.Category,
		Provider:     provider.Provider,
		BucketName:   provider.Config,
		ObjectKey:    result.ObjectKey,
		PublicURL:    result.PublicURL,
		PrivateURL:   result.PrivateURL,
		ThumbnailURL: result.ThumbnailURL,
		HashMD5:      result.MD5Hash,
		HashSHA256:   result.SHA256Hash,
		UploadedBy:   userID,
		RelatedType:  req.RelatedType,
		RelatedID:    req.RelatedID,
		IsPublic:     req.IsPublic,
		ExpiresAt:    expiresAt,
		Status:       models.FileStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(storageFile).Error; err != nil {
		return nil, fmt.Errorf("保存文件记录失败: %w", err)
	}

	// 记录操作
	s.recordOperation(fileID, "upload", userID, "success", result.FileSize, 0, "")

	// 更新存储配置使用量
	s.updateStorageUsage(provider.ID, result.FileSize)

	response := &models.UploadResponse{
		FileID:       fileID,
		FileName:     storageFile.FileName,
		PublicURL:    result.PublicURL,
		PrivateURL:   result.PrivateURL,
		ThumbnailURL: result.ThumbnailURL,
		FileSize:     result.FileSize,
		MimeType:     storageFile.MimeType,
	}

	return response, nil
}

// GetFile 获取文件信息
func (s *StorageService) GetFile(fileID string) (*models.StorageFile, error) {
	var file models.StorageFile
	if err := s.db.Where("id = ? AND status != ?", fileID, models.FileStatusDeleted).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, err
	}

	// 检查文件是否过期
	if file.ExpiresAt != nil && time.Now().After(*file.ExpiresAt) {
		return nil, fmt.Errorf("文件已过期")
	}

	return &file, nil
}

// GetFiles 获取文件列表
func (s *StorageService) GetFiles(query *models.FileQuery) ([]models.StorageFile, int64, error) {
	var files []models.StorageFile
	var total int64

	db := s.db.Model(&models.StorageFile{}).Where("status != ?", models.FileStatusDeleted)

	// 应用过滤条件
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Provider != "" {
		db = db.Where("provider = ?", query.Provider)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.RelatedType != "" {
		db = db.Where("related_type = ?", query.RelatedType)
	}
	if query.RelatedID != "" {
		db = db.Where("related_id = ?", query.RelatedID)
	}
	if query.UploadedBy != "" {
		db = db.Where("uploaded_by = ?", query.UploadedBy)
	}
	if query.StartDate != nil {
		db = db.Where("created_at >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("created_at <= ?", *query.EndDate)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC"
	if query.SortBy != "" {
		order := "ASC"
		if query.SortOrder == "desc" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", query.SortBy, order)
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order(orderBy).Offset(offset).Limit(query.PageSize).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// DeleteFile 删除文件
func (s *StorageService) DeleteFile(fileID, userID string) error {
	var file models.StorageFile
	if err := s.db.Where("id = ?", fileID).First(&file).Error; err != nil {
		return fmt.Errorf("文件不存在")
	}

	// 获取存储提供商
	provider, err := s.getProviderByType(file.Provider)
	if err != nil {
		return fmt.Errorf("获取存储提供商失败: %w", err)
	}

	// 从存储提供商删除文件
	storageProvider := s.createStorageProvider(provider)
	if err := storageProvider.Delete(file.ObjectKey); err != nil {
		// 记录错误但不阻止数据库删除
		s.recordOperation(fileID, "delete", userID, "failed", 0, 0, err.Error())
	}

	// 软删除文件记录
	if err := s.db.Model(&file).Updates(map[string]interface{}{
		"status":     models.FileStatusDeleted,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("删除文件记录失败: %w", err)
	}

	// 记录操作
	s.recordOperation(fileID, "delete", userID, "success", 0, 0, "")

	// 更新存储使用量
	s.updateStorageUsage(provider.ID, -file.FileSize)

	return nil
}

// GetStorageStats 获取存储统计信息
func (s *StorageService) GetStorageStats() (*models.StorageStats, error) {
	stats := &models.StorageStats{
		FilesByCategory: make(map[string]int64),
		FilesByProvider: make(map[string]int64),
		FilesByStatus:   make(map[string]int64),
		StorageUsage:    make(map[string]int64),
		LastUpdate:      time.Now(),
	}

	// 总文件数和大小
	s.db.Model(&models.StorageFile{}).
		Where("status != ?", models.FileStatusDeleted).
		Count(&stats.TotalFiles)

	var totalSize sql.NullInt64
	s.db.Model(&models.StorageFile{}).
		Where("status != ?", models.FileStatusDeleted).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize)
	stats.TotalSize = totalSize.Int64

	// 按分类统计
	var categoryStats []struct {
		Category string
		Count    int64
	}
	s.db.Model(&models.StorageFile{}).
		Select("category, count(*) as count").
		Where("status != ?", models.FileStatusDeleted).
		Group("category").
		Scan(&categoryStats)

	for _, stat := range categoryStats {
		stats.FilesByCategory[stat.Category] = stat.Count
	}

	// 按提供商统计
	var providerStats []struct {
		Provider string
		Count    int64
	}
	s.db.Model(&models.StorageFile{}).
		Select("provider, count(*) as count").
		Where("status != ?", models.FileStatusDeleted).
		Group("provider").
		Scan(&providerStats)

	for _, stat := range providerStats {
		stats.FilesByProvider[stat.Provider] = stat.Count
	}

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	s.db.Model(&models.StorageFile{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusStats)

	for _, stat := range statusStats {
		stats.FilesByStatus[stat.Status] = stat.Count
	}

	// 最近24小时上传数
	yesterday := time.Now().AddDate(0, 0, -1)
	s.db.Model(&models.StorageFile{}).
		Where("created_at >= ? AND status != ?", yesterday, models.FileStatusDeleted).
		Count(&stats.RecentUploads)

	// 存储使用量
	var usageStats []struct {
		Provider string
		Size     int64
	}
	s.db.Model(&models.StorageFile{}).
		Select("provider, COALESCE(SUM(file_size), 0) as size").
		Where("status != ?", models.FileStatusDeleted).
		Group("provider").
		Scan(&usageStats)

	for _, stat := range usageStats {
		stats.StorageUsage[stat.Provider] = stat.Size
	}

	return stats, nil
}

// 私有方法

// validateFile 验证文件
func (s *StorageService) validateFile(file *multipart.FileHeader, category models.FileCategory) error {
	// 检查文件大小
	if file.Size > 100*1024*1024 { // 100MB
		return fmt.Errorf("文件大小超过限制")
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := s.getAllowedExtensions(category)

	if len(allowedExts) > 0 {
		found := false
		for _, allowedExt := range allowedExts {
			if ext == allowedExt {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("文件类型不支持")
		}
	}

	return nil
}

// getAllowedExtensions 获取允许的文件扩展名
func (s *StorageService) getAllowedExtensions(category models.FileCategory) []string {
	switch category {
	case models.FileCategoryImage, models.FileCategoryAvatar, models.FileCategoryEnvelope, models.FileCategoryThumbnail:
		return []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	case models.FileCategoryDocument:
		return []string{".pdf", ".doc", ".docx", ".txt", ".md"}
	case models.FileCategoryQRCode:
		return []string{".png", ".svg"}
	default:
		return []string{} // 允许所有类型
	}
}

// generateObjectKey 生成对象键
func (s *StorageService) generateObjectKey(fileID, originalName string, category models.FileCategory) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Format("2006/01/02")
	return fmt.Sprintf("%s/%s/%s%s", category, timestamp, fileID, ext)
}

// getDefaultProvider 获取默认存储提供商
func (s *StorageService) getDefaultProvider() (*models.StorageConfig, error) {
	var config models.StorageConfig
	if err := s.db.Where("is_enabled = ? AND is_default = ?", true, true).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有默认提供商，使用本地存储
			return s.getOrCreateLocalProvider()
		}
		return nil, err
	}
	return &config, nil
}

// getOrCreateLocalProvider 获取或创建本地存储提供商
func (s *StorageService) getOrCreateLocalProvider() (*models.StorageConfig, error) {
	var config models.StorageConfig
	err := s.db.Where("provider = ?", models.StorageProviderLocal).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 创建本地存储配置
		config = models.StorageConfig{
			ID:           uuid.New().String(),
			Provider:     models.StorageProviderLocal,
			DisplayName:  "本地存储",
			Config:       `{"base_path": "./uploads", "base_url": "/uploads"}`,
			IsEnabled:    true,
			IsDefault:    true,
			Priority:     1,
			MaxFileSize:  100 * 1024 * 1024, // 100MB
			AllowedTypes: `[".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"]`,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.db.Create(&config).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &config, nil
}

// getProviderByType 根据类型获取存储提供商
func (s *StorageService) getProviderByType(providerType models.StorageProvider) (*models.StorageConfig, error) {
	var config models.StorageConfig
	if err := s.db.Where("provider = ? AND is_enabled = ?", providerType, true).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// createStorageProvider 创建存储提供商实例
func (s *StorageService) createStorageProvider(config *models.StorageConfig) StorageProvider {
	switch config.Provider {
	case models.StorageProviderLocal:
		return s.createLocalProvider(config)
	case models.StorageProviderAliOSS:
		return s.createAliOSSProvider(config)
	case models.StorageProviderTencentCOS:
		return s.createTencentCOSProvider(config)
	default:
		return s.createLocalProvider(config)
	}
}

// createLocalProvider 创建本地存储提供商
func (s *StorageService) createLocalProvider(config *models.StorageConfig) StorageProvider {
	var localConfig struct {
		BasePath string `json:"base_path"`
		BaseURL  string `json:"base_url"`
	}
	json.Unmarshal([]byte(config.Config), &localConfig)

	return &LocalStorageProvider{
		basePath:  localConfig.BasePath,
		baseURL:   localConfig.BaseURL,
		uploadDir: "uploads",
	}
}

// createAliOSSProvider 创建阿里云OSS提供商
func (s *StorageService) createAliOSSProvider(config *models.StorageConfig) StorageProvider {
	var ossConfig struct {
		Endpoint        string `json:"endpoint"`
		AccessKeyID     string `json:"access_key_id"`
		AccessKeySecret string `json:"access_key_secret"`
		BucketName      string `json:"bucket_name"`
		BaseURL         string `json:"base_url"`
	}
	json.Unmarshal([]byte(config.Config), &ossConfig)

	return &AliOSSProvider{
		endpoint:        ossConfig.Endpoint,
		accessKeyID:     ossConfig.AccessKeyID,
		accessKeySecret: ossConfig.AccessKeySecret,
		bucketName:      ossConfig.BucketName,
		baseURL:         ossConfig.BaseURL,
	}
}

// createTencentCOSProvider 创建腾讯云COS提供商
func (s *StorageService) createTencentCOSProvider(config *models.StorageConfig) StorageProvider {
	var cosConfig struct {
		Region     string `json:"region"`
		SecretID   string `json:"secret_id"`
		SecretKey  string `json:"secret_key"`
		BucketName string `json:"bucket_name"`
		BaseURL    string `json:"base_url"`
	}
	json.Unmarshal([]byte(config.Config), &cosConfig)

	return &TencentCOSProvider{
		region:     cosConfig.Region,
		secretID:   cosConfig.SecretID,
		secretKey:  cosConfig.SecretKey,
		bucketName: cosConfig.BucketName,
		baseURL:    cosConfig.BaseURL,
	}
}

// uploadToProvider 上传到存储提供商
func (s *StorageService) uploadToProvider(config *models.StorageConfig, file *multipart.FileHeader, objectKey string) (*UploadResult, error) {
	provider := s.createStorageProvider(config)

	result, err := provider.Upload(file, objectKey)
	if err != nil {
		return nil, err
	}

	// 计算文件哈希
	fileReader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	md5Hash, sha256Hash, err := s.calculateHashes(fileReader)
	if err != nil {
		return nil, err
	}

	result.MD5Hash = md5Hash
	result.SHA256Hash = sha256Hash

	return result, nil
}

// calculateHashes 计算文件哈希
func (s *StorageService) calculateHashes(reader io.Reader) (string, string, error) {
	md5Hash := md5.New()
	sha256Hash := sha256.New()

	multiWriter := io.MultiWriter(md5Hash, sha256Hash)

	if _, err := io.Copy(multiWriter, reader); err != nil {
		return "", "", err
	}

	return hex.EncodeToString(md5Hash.Sum(nil)), hex.EncodeToString(sha256Hash.Sum(nil)), nil
}

// recordOperation 记录操作
func (s *StorageService) recordOperation(fileID, operation, userID, status string, bytesTransferred int64, duration int, errorMsg string) {
	opRecord := &models.StorageOperation{
		ID:               uuid.New().String(),
		FileID:           fileID,
		Operation:        operation,
		UserID:           userID,
		Status:           status,
		BytesTransferred: bytesTransferred,
		Duration:         duration,
		ErrorMsg:         errorMsg,
		CreatedAt:        time.Now(),
	}

	s.db.Create(opRecord)
}

// updateStorageUsage 更新存储使用量
func (s *StorageService) updateStorageUsage(configID string, sizeDelta int64) {
	s.db.Model(&models.StorageConfig{}).
		Where("id = ?", configID).
		Update("current_size", gorm.Expr("current_size + ?", sizeDelta))
}
