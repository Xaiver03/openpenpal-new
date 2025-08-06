/**
 * 统一响应处理单元测试
 * 测试标准化的HTTP响应功能
 */

package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestSuccessResponses 测试成功响应
func TestSuccessResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		handler        func(*gin.Context)
		expectedStatus int
		checkResponse  func(map[string]interface{}) error
	}{
		{
			name: "基本成功响应",
			handler: func(c *gin.Context) {
				Success(c, gin.H{"id": 123})
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(resp map[string]interface{}) error {
				if !resp["success"].(bool) {
					t.Error("success should be true")
				}
				if resp["code"].(float64) != 200 {
					t.Error("code should be 200")
				}
				data := resp["data"].(map[string]interface{})
				if data["id"].(float64) != 123 {
					t.Error("data.id should be 123")
				}
				return nil
			},
		},
		{
			name: "带消息的成功响应",
			handler: func(c *gin.Context) {
				SuccessWithMessage(c, "操作成功", gin.H{"count": 5})
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(resp map[string]interface{}) error {
				if resp["message"].(string) != "操作成功" {
					t.Error("message mismatch")
				}
				return nil
			},
		},
		{
			name: "创建成功响应",
			handler: func(c *gin.Context) {
				Created(c, gin.H{"id": "new-123"})
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(resp map[string]interface{}) error {
				if resp["code"].(float64) != 201 {
					t.Error("code should be 201")
				}
				return nil
			},
		},
		{
			name: "空数据成功响应",
			handler: func(c *gin.Context) {
				Success(c, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(resp map[string]interface{}) error {
				if resp["data"] != nil {
					t.Error("data should be nil")
				}
				return nil
			},
		},
		{
			name: "数组数据响应",
			handler: func(c *gin.Context) {
				Success(c, []string{"item1", "item2", "item3"})
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(resp map[string]interface{}) error {
				data := resp["data"].([]interface{})
				if len(data) != 3 {
					t.Error("data should have 3 items")
				}
				return nil
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			tt.handler(c)
			
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
			
			if err := tt.checkResponse(resp); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestErrorResponses 测试错误响应
func TestErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		handler        func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "基本错误响应",
			handler: func(c *gin.Context) {
				Error(c, http.StatusBadRequest, "请求参数错误")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "请求参数错误",
		},
		{
			name: "带详情的错误响应",
			handler: func(c *gin.Context) {
				ErrorWithMessage(c, http.StatusForbidden, "权限不足", "需要管理员权限")
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "权限不足",
		},
		{
			name: "未授权错误",
			handler: func(c *gin.Context) {
				Unauthorized(c, "请先登录")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "请先登录",
		},
		{
			name: "未找到错误",
			handler: func(c *gin.Context) {
				NotFound(c, "资源不存在")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "资源不存在",
		},
		{
			name: "服务器内部错误",
			handler: func(c *gin.Context) {
				InternalServerError(c, "服务器错误")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "服务器错误",
		},
		{
			name: "验证错误",
			handler: func(c *gin.Context) {
				ValidationError(c, gin.H{
					"email": "邮箱格式不正确",
					"phone": "手机号格式不正确",
				})
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/test", nil)
			
			tt.handler(c)
			
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
			
			if !resp["success"].(bool) == true {
				t.Error("success should be false for error responses")
			}
			
			if tt.expectedError != "" && resp["error"].(string) != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, resp["error"].(string))
			}
		})
	}
}

// TestPaginatedResponse 测试分页响应
func TestPaginatedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	items := []map[string]interface{}{
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
		{"id": 3, "name": "Item 3"},
	}
	
	Paginated(c, items, 1, 10, 3)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	data := resp["data"].(map[string]interface{})
	if len(data["items"].([]interface{})) != 3 {
		t.Error("Should have 3 items")
	}
	
	pagination := data["pagination"].(map[string]interface{})
	if pagination["page"].(float64) != 1 {
		t.Error("Page should be 1")
	}
	if pagination["pageSize"].(float64) != 10 {
		t.Error("PageSize should be 10")
	}
	if pagination["total"].(float64) != 3 {
		t.Error("Total should be 3")
	}
}

// TestResponseHeaders 测试响应头
func TestResponseHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name          string
		handler       func(*gin.Context)
		checkHeaders  func(http.Header) error
	}{
		{
			name: "成功响应头",
			handler: func(c *gin.Context) {
				Success(c, nil)
			},
			checkHeaders: func(headers http.Header) error {
				if headers.Get("Content-Type") != "application/json; charset=utf-8" {
					t.Error("Content-Type header mismatch")
				}
				if headers.Get("X-Content-Type-Options") != "nosniff" {
					t.Error("Security header missing")
				}
				return nil
			},
		},
		{
			name: "错误响应缓存控制",
			handler: func(c *gin.Context) {
				Error(c, http.StatusBadRequest, "error")
			},
			checkHeaders: func(headers http.Header) error {
				if headers.Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
					t.Error("Cache-Control header mismatch for error")
				}
				return nil
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/test", nil)
			
			tt.handler(c)
			
			if err := tt.checkHeaders(w.Header()); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestResponseWithAudit 测试带审计信息的响应
func TestResponseWithAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/test", nil)
	c.Set("user_id", "test-user-123")
	
	// 测试带审计信息的响应
	SuccessWithAudit(c, gin.H{"action": "test"}, AuditInfo{
		Action:   "TEST_ACTION",
		Resource: "test_resource",
		Result:   "success",
	})
	
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// 检查审计信息
	audit := resp["audit"].(map[string]interface{})
	if audit["action"].(string) != "TEST_ACTION" {
		t.Error("Audit action mismatch")
	}
	if audit["userId"].(string) != "test-user-123" {
		t.Error("Audit userId mismatch")
	}
	if audit["path"].(string) != "/api/test" {
		t.Error("Audit path mismatch")
	}
}

// TestConcurrentResponses 测试并发响应
func TestConcurrentResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建多个并发请求
	numRequests := 100
	done := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			if id%2 == 0 {
				Success(c, gin.H{"id": id})
			} else {
				Error(c, http.StatusBadRequest, "error")
			}
			
			done <- true
		}(i)
	}
	
	// 等待所有请求完成
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// BenchmarkSuccessResponse 性能测试 - 成功响应
func BenchmarkSuccessResponse(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		Success(c, gin.H{"id": i, "name": "test"})
	}
}

// BenchmarkErrorResponse 性能测试 - 错误响应
func BenchmarkErrorResponse(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)
		Error(c, http.StatusBadRequest, "error message")
	}
}

// BenchmarkPaginatedResponse 性能测试 - 分页响应
func BenchmarkPaginatedResponse(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	// 准备测试数据
	items := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = gin.H{"id": i, "name": "item"}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		Paginated(c, items, 1, 20, 100)
	}
}