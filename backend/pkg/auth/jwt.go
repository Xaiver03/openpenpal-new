package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string          `json:"userId"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成JWT令牌
func GenerateJWT(userID string, role models.UserRole, secret string, expiresAt time.Time) (string, error) {
	// 生成唯一的JWT ID
	jti, err := generateJTI()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT ID: %w", err)
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti, // 添加JWT ID用于黑名单功能
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "openpenpal",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// generateJTI 生成唯一的JWT ID
func generateJTI() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateJWT 验证JWT令牌
func ValidateJWT(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	return claims, nil
}

// ExtractTokenFromHeader 从请求头中提取令牌
func ExtractTokenFromHeader(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "

	if len(authHeader) < len(bearerPrefix) {
		return "", fmt.Errorf("invalid authorization header format")
	}

	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}

	return authHeader[len(bearerPrefix):], nil
}
