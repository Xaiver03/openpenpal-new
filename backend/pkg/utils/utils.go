package utils

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// GenerateLetterCode 生成信件编号
func GenerateLetterCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 12

	rand.Seed(time.Now().UnixNano())

	code := make([]byte, codeLength)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}

	// 格式化为 OP + 时间戳 + 随机数
	timestamp := time.Now().Unix()
	return fmt.Sprintf("OP%d%s", timestamp%10000, string(code[:8]))
}

// IsValidSchoolCode 验证学校代码格式
func IsValidSchoolCode(code string) bool {
	// 学校代码规则：6位字符，支持字母和数字
	pattern := `^[A-Z0-9]{6}$`
	matched, _ := regexp.MatchString(pattern, code)
	return matched
}

// EnsureDir 确保目录存在
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// Contains 检查切片是否包含指定元素
func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetPage 计算分页参数
func GetPage(page, limit int) (offset int, adjustedLimit int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset = (page - 1) * limit
	return offset, limit
}

// Pagination 分页信息
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// CalculatePagination 计算分页信息
func CalculatePagination(page, limit int, total int64) *Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// TruncateString 截断字符串
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// SanitizeString 清理字符串（移除特殊字符）
func SanitizeString(s string) string {
	// 移除HTML标签和特殊字符
	pattern := `<[^>]*>|[<>&"']`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, "")
}

// ParseUUID 解析UUID字符串
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
