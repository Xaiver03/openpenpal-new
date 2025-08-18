package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateSignatureKey 生成32位随机签名密钥
func GenerateSignatureKey() string {
	bytes := make([]byte, 16) // 16字节 = 32字符的hex
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateSecurityHash 生成SHA256安全哈希
func GenerateSecurityHash(code string, timestamp time.Time, signatureKey string) string {
	// 构造签名数据：条码+时间戳+密钥
	data := fmt.Sprintf("%s:%d:%s", code, timestamp.Unix(), signatureKey)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifyBarcodeSignature 验证条码签名
func VerifyBarcodeSignature(code string, securityHash string, signatureKey string, createdAt time.Time) bool {
	expectedHash := GenerateSecurityHash(code, createdAt, signatureKey)
	return securityHash == expectedHash
}

// ValidateBarcodeIntegrity 验证条码完整性（防篡改）
func ValidateBarcodeIntegrity(code string, securityHash string, signatureKey string, createdAt time.Time) error {
	if !VerifyBarcodeSignature(code, securityHash, signatureKey, createdAt) {
		return fmt.Errorf("invalid barcode signature - potential forgery detected")
	}
	
	// 检查时间戳是否合理（不能太旧或太新）
	now := time.Now()
	maxAge := 365 * 24 * time.Hour // 1年
	if now.Sub(createdAt) > maxAge {
		return fmt.Errorf("barcode too old - expired")
	}
	
	if createdAt.After(now.Add(time.Hour)) {
		return fmt.Errorf("barcode timestamp is in the future - invalid")
	}
	
	return nil
}