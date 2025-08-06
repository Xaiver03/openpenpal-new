/**
 * JWT认证中间件单元测试
 * 测试统一认证中间件的核心功能
 */

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/openpenpal/shared/go/pkg/permissions"
)

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// generateTestToken 生成测试JWT token
func generateTestToken(claims *JWTClaims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// TestJWTMiddleware 测试JWT中间件基本功能
func TestJWTMiddleware(t *testing.T) {
	router := setupTestRouter()
	secret := []byte("test-secret-key")
	
	// 配置JWT中间件
	config := &JWTConfig{
		SigningKey: secret,
		SkipperFunc: func(c *gin.Context) bool {
			return c.Request.URL.Path == "/skip"
		},
	}
	
	router.Use(JWTMiddleware(config))
	
	// 测试路由
	router.GET("/protected", func(c *gin.Context) {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no claims"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"claims": claims})
	})
	
	router.GET("/skip", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "skipped"})
	})
	
	tests := []struct {
		name       string
		path       string
		token      string
		setupToken func() string
		wantStatus int
	}{
		{
			name:       "无token访问保护路由",
			path:       "/protected",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "跳过认证的路由",
			path:       "/skip",
			token:      "",
			wantStatus: http.StatusOK,
		},
		{
			name: "有效token访问",
			path: "/protected",
			setupToken: func() string {
				claims := &JWTClaims{
					UserID:   "test-user",
					Username: "testuser",
					Role:     permissions.RoleUser,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				token, _ := generateTestToken(claims, secret)
				return token
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "过期token",
			path: "/protected",
			setupToken: func() string {
				claims := &JWTClaims{
					UserID: "test-user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					},
				}
				token, _ := generateTestToken(claims, secret)
				return token
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "无效token格式",
			path:       "/protected",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			
			// 设置token
			if tt.setupToken != nil {
				tt.token = tt.setupToken()
			}
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestRequirePermission 测试权限要求中间件
func TestRequirePermission(t *testing.T) {
	router := setupTestRouter()
	secret := []byte("test-secret-key")
	
	// 配置JWT中间件
	config := &JWTConfig{
		SigningKey: secret,
	}
	
	router.Use(JWTMiddleware(config))
	
	// 需要特定权限的路由
	router.GET("/courier", RequirePermission(permissions.PermissionCourierScanCode), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "courier access"})
	})
	
	tests := []struct {
		name       string
		userRole   permissions.UserRole
		hasCourier bool
		wantStatus int
	}{
		{
			name:       "普通用户无权限",
			userRole:   permissions.RoleUser,
			hasCourier: false,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "信使有权限",
			userRole:   permissions.RoleCourier,
			hasCourier: true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "管理员有权限",
			userRole:   permissions.RoleAdmin,
			hasCourier: true,
			wantStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成token
			claims := &JWTClaims{
				UserID:   "test-user",
				Username: "testuser",
				Role:     tt.userRole,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}
			
			if tt.hasCourier && tt.userRole == permissions.RoleCourier {
				claims.CourierInfo = &permissions.CourierInfo{
					Level: permissions.CourierLevel1,
				}
			}
			
			token, _ := generateTestToken(claims, secret)
			
			req, _ := http.NewRequest("GET", "/courier", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestRequireRole 测试角色要求中间件
func TestRequireRole(t *testing.T) {
	router := setupTestRouter()
	secret := []byte("test-secret-key")
	
	// 配置JWT中间件
	config := &JWTConfig{
		SigningKey: secret,
	}
	
	router.Use(JWTMiddleware(config))
	
	// 需要管理员角色的路由
	router.GET("/admin", RequireRole(permissions.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})
	
	tests := []struct {
		name       string
		userRole   permissions.UserRole
		wantStatus int
	}{
		{
			name:       "普通用户无权限",
			userRole:   permissions.RoleUser,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "信使无权限",
			userRole:   permissions.RoleCourier,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "管理员有权限",
			userRole:   permissions.RoleAdmin,
			wantStatus: http.StatusOK,
		},
		{
			name:       "超级管理员有权限",
			userRole:   permissions.RoleSuperAdmin,
			wantStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成token
			claims := &JWTClaims{
				UserID:   "test-user",
				Username: "testuser",
				Role:     tt.userRole,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}
			
			token, _ := generateTestToken(claims, secret)
			
			req, _ := http.NewRequest("GET", "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestTokenExtraction 测试不同方式的token提取
func TestTokenExtraction(t *testing.T) {
	router := setupTestRouter()
	secret := []byte("test-secret-key")
	
	// 测试从header提取
	configHeader := &JWTConfig{
		SigningKey:  secret,
		TokenLookup: "header:Authorization",
	}
	
	// 测试从query提取
	configQuery := &JWTConfig{
		SigningKey:  secret,
		TokenLookup: "query:token",
	}
	
	// 测试从cookie提取
	configCookie := &JWTConfig{
		SigningKey:  secret,
		TokenLookup: "cookie:jwt",
	}
	
	tests := []struct {
		name   string
		config *JWTConfig
		setup  func(*http.Request, string)
	}{
		{
			name:   "从header提取",
			config: configHeader,
			setup: func(req *http.Request, token string) {
				req.Header.Set("Authorization", "Bearer "+token)
			},
		},
		{
			name:   "从query提取",
			config: configQuery,
			setup: func(req *http.Request, token string) {
				q := req.URL.Query()
				q.Add("token", token)
				req.URL.RawQuery = q.Encode()
			},
		},
		{
			name:   "从cookie提取",
			config: configCookie,
			setup: func(req *http.Request, token string) {
				req.AddCookie(&http.Cookie{
					Name:  "jwt",
					Value: token,
				})
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			r.Use(JWTMiddleware(tt.config))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"success": true})
			})
			
			// 生成有效token
			claims := &JWTClaims{
				UserID: "test-user",
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}
			token, _ := generateTestToken(claims, secret)
			
			req, _ := http.NewRequest("GET", "/test", nil)
			tt.setup(req, token)
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

// TestGenerateToken 测试token生成
func TestGenerateToken(t *testing.T) {
	secret := []byte("test-secret-key")
	
	user := &permissions.User{
		Role: permissions.RoleCourier,
		CourierInfo: &permissions.CourierInfo{
			Level: permissions.CourierLevel2,
		},
	}
	
	// 生成token
	token, err := GenerateToken(user, "test-user", "testuser", "test@example.com", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	
	if token == "" {
		t.Error("Generated token should not be empty")
	}
	
	// 解析token验证内容
	claims := &JWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	
	if !parsedToken.Valid {
		t.Error("Token should be valid")
	}
	
	if claims.UserID != "test-user" {
		t.Errorf("UserID = %s, want test-user", claims.UserID)
	}
	
	if claims.Role != permissions.RoleCourier {
		t.Errorf("Role = %s, want %s", claims.Role, permissions.RoleCourier)
	}
	
	if claims.CourierInfo == nil || claims.CourierInfo.Level != permissions.CourierLevel2 {
		t.Error("CourierInfo not properly set")
	}
}

// BenchmarkJWTMiddleware 性能测试
func BenchmarkJWTMiddleware(b *testing.B) {
	router := setupTestRouter()
	secret := []byte("test-secret-key")
	
	config := &JWTConfig{
		SigningKey: secret,
	}
	
	router.Use(JWTMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	
	// 生成token
	claims := &JWTClaims{
		UserID: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, _ := generateTestToken(claims, secret)
	
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}