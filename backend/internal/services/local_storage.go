package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"openpenpal-backend/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

// Upload 上传文件到本地存储
func (p *LocalStorageProvider) Upload(file *multipart.FileHeader, objectKey string) (*UploadResult, error) {
	// 构建完整的文件路径
	fullPath := filepath.Join(p.basePath, objectKey)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := utils.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	fileSize, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("复制文件失败: %w", err)
	}

	// 构建URL
	publicURL := fmt.Sprintf("%s/%s", p.baseURL, objectKey)
	privateURL := publicURL // 本地存储没有私有URL概念

	return &UploadResult{
		ObjectKey:    objectKey,
		PublicURL:    publicURL,
		PrivateURL:   privateURL,
		ThumbnailURL: "", // 缩略图需要单独生成
		FileSize:     fileSize,
	}, nil
}

// Download 从本地存储下载文件
func (p *LocalStorageProvider) Download(objectKey string) (io.ReadCloser, error) {
	fullPath := filepath.Join(p.basePath, objectKey)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	return file, nil
}

// Delete 从本地存储删除文件
func (p *LocalStorageProvider) Delete(objectKey string) error {
	fullPath := filepath.Join(p.basePath, objectKey)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，认为删除成功
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetPublicURL 获取公共访问URL
func (p *LocalStorageProvider) GetPublicURL(objectKey string) string {
	return fmt.Sprintf("%s/%s", p.baseURL, objectKey)
}

// GetPrivateURL 获取私有访问URL（本地存储不支持）
func (p *LocalStorageProvider) GetPrivateURL(objectKey string, expiration time.Duration) string {
	// 本地存储没有私有URL概念，返回公共URL
	return p.GetPublicURL(objectKey)
}

// GenerateThumbnail 生成缩略图（简化实现）
func (p *LocalStorageProvider) GenerateThumbnail(objectKey string, width, height int) (string, error) {
	// 简化实现：返回原图URL
	// 实际项目中可以使用图像处理库生成真正的缩略图
	return p.GetPublicURL(objectKey), nil
}
