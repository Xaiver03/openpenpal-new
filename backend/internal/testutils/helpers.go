package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	return db
}

// SetupTestRouter 创建测试路由
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	return router
}

// MakeRequest 执行HTTP请求
func MakeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}, headers ...http.Header) *httptest.ResponseRecorder {
	var req *http.Request

	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	// 添加额外的请求头
	for _, h := range headers {
		for k, v := range h {
			req.Header[k] = v
		}
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// ParseResponse 解析响应
func ParseResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	assert.NoError(t, err)
}

// CreateAuthHeader 创建认证头
func CreateAuthHeader(token string) http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	return header
}

// AssertErrorResponse 断言错误响应
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	assert.Equal(t, expectedStatus, w.Code)

	var response map[string]interface{}
	ParseResponse(t, w, &response)

	assert.False(t, response["success"].(bool))
	if expectedMessage != "" {
		assert.Contains(t, response["message"].(string), expectedMessage)
	}
}

// AssertSuccessResponse 断言成功响应
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, w.Code)

	var response map[string]interface{}
	ParseResponse(t, w, &response)

	assert.True(t, response["success"].(bool))
}
