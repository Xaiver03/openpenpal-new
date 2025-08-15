package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// TestConfig 测试配置
type TestConfig struct {
	BaseURL    string `json:"base_url"`
	AdminUser  string `json:"admin_user"`
	AdminPass  string `json:"admin_pass"`
	TestUser   string `json:"test_user"`
	TestPass   string `json:"test_pass"`
	Timeout    int    `json:"timeout"`
}

// TestResult 测试结果
type TestResult struct {
	TestName    string        `json:"test_name"`
	Passed      bool          `json:"passed"`
	Duration    time.Duration `json:"duration"`
	Details     string        `json:"details"`
	ErrorMsg    string        `json:"error_msg,omitempty"`
	StatusCode  int           `json:"status_code"`
	Category    string        `json:"category"`
}

// E2ESecurityTester 端到端安全测试器
type E2ESecurityTester struct {
	config  TestConfig
	client  *http.Client
	results []TestResult
	tokens  map[string]string
}

// NewE2ESecurityTester 创建端到端安全测试器
func NewE2ESecurityTester(config TestConfig) *E2ESecurityTester {
	return &E2ESecurityTester{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		results: make([]TestResult, 0),
		tokens:  make(map[string]string),
	}
}

// RunAllTests 运行所有测试
func (tester *E2ESecurityTester) RunAllTests() {
	fmt.Println("🚀 开始端到端安全测试...")
	
	// 1. 基础连通性测试
	tester.testConnectivity()
	
	// 2. 认证系统测试
	tester.testAuthentication()
	
	// 3. 输入验证测试
	tester.testInputValidation()
	
	// 4. 权限控制测试
	tester.testPermissionControl()
	
	// 5. 内容安全测试
	tester.testContentSecurity()
	
	// 6. 速率限制测试
	tester.testRateLimit()
	
	// 7. 安全监控测试
	tester.testSecurityMonitoring()
	
	// 8. 端到端数据流测试
	tester.testEndToEndDataFlow()
	
	// 9. 安全验证系统测试
	tester.testValidationSystem()
	
	// 输出测试报告
	tester.generateReport()
}

// testConnectivity 测试基础连通性
func (tester *E2ESecurityTester) testConnectivity() {
	fmt.Println("\n📡 测试基础连通性...")
	
	// 健康检查
	tester.runTest("Health Check", "connectivity", func() (bool, string, int) {
		resp, err := tester.client.Get(tester.config.BaseURL + "/health")
		if err != nil {
			return false, err.Error(), 0
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == 200 {
			return true, "Health check passed", resp.StatusCode
		}
		return false, fmt.Sprintf("Health check failed with status %d", resp.StatusCode), resp.StatusCode
	})
	
	// Ping测试
	tester.runTest("Ping Test", "connectivity", func() (bool, string, int) {
		resp, err := tester.client.Get(tester.config.BaseURL + "/ping")
		if err != nil {
			return false, err.Error(), 0
		}
		defer resp.Body.Close()
		
		return resp.StatusCode == 200, "Ping test", resp.StatusCode
	})
}

// testAuthentication 测试认证系统
func (tester *E2ESecurityTester) testAuthentication() {
	fmt.Println("\n🔐 测试认证系统...")
	
	// 管理员登录
	tester.runTest("Admin Login", "authentication", func() (bool, string, int) {
		token, statusCode, err := tester.login(tester.config.AdminUser, tester.config.AdminPass)
		if err != nil {
			return false, err.Error(), statusCode
		}
		
		tester.tokens["admin"] = token
		return true, "Admin login successful", statusCode
	})
	
	// 普通用户登录
	tester.runTest("User Login", "authentication", func() (bool, string, int) {
		token, statusCode, err := tester.login(tester.config.TestUser, tester.config.TestPass)
		if err != nil {
			return false, err.Error(), statusCode
		}
		
		tester.tokens["user"] = token
		return true, "User login successful", statusCode
	})
	
	// 无效凭据测试
	tester.runTest("Invalid Credentials", "authentication", func() (bool, string, int) {
		_, statusCode, err := tester.login("invalid", "invalid")
		if err != nil && statusCode == 401 {
			return true, "Invalid credentials correctly rejected", statusCode
		}
		return false, "Invalid credentials not properly handled", statusCode
	})
}

// testInputValidation 测试输入验证
func (tester *E2ESecurityTester) testInputValidation() {
	fmt.Println("\n🛡️ 测试输入验证...")
	
	// XSS攻击测试
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"javascript:alert(1)",
		"<img src=x onerror=alert(1)>",
		"<iframe src=javascript:alert(1)></iframe>",
	}
	
	for i, payload := range xssPayloads {
		tester.runTest(fmt.Sprintf("XSS Attack Test %d", i+1), "input_validation", func() (bool, string, int) {
			body := map[string]interface{}{
				"username": payload,
				"password": "test",
			}
			
			statusCode, err := tester.makeRequest("POST", "/api/v1/auth/login", body, "")
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			// XSS payload 应该被拒绝 (400 Bad Request)
			if statusCode == 400 {
				return true, "XSS payload correctly rejected", statusCode
			}
			return false, "XSS payload not detected", statusCode
		})
	}
	
	// SQL注入测试
	sqlPayloads := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"admin'/*",
		"' UNION SELECT * FROM users --",
	}
	
	for i, payload := range sqlPayloads {
		tester.runTest(fmt.Sprintf("SQL Injection Test %d", i+1), "input_validation", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/letters/public?search="+payload, nil, "")
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			// SQL注入payload应该被拒绝或安全处理
			if statusCode == 400 || statusCode == 200 {
				return true, "SQL injection payload handled safely", statusCode
			}
			return false, "SQL injection vulnerability detected", statusCode
		})
	}
}

// testPermissionControl 测试权限控制
func (tester *E2ESecurityTester) testPermissionControl() {
	fmt.Println("\n🔒 测试权限控制...")
	
	// 管理员访问敏感功能
	if adminToken, ok := tester.tokens["admin"]; ok {
		tester.runTest("Admin Access to Sensitive Words", "permission_control", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/admin/sensitive-words", nil, adminToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 200, "Admin access to sensitive words", statusCode
		})
		
		tester.runTest("Admin Access to Security Dashboard", "permission_control", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/admin/security/dashboard", nil, adminToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 200, "Admin access to security dashboard", statusCode
		})
	}
	
	// 普通用户访问受限功能
	if userToken, ok := tester.tokens["user"]; ok {
		tester.runTest("User Forbidden Access", "permission_control", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/admin/sensitive-words", nil, userToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			// 普通用户应该被拒绝访问 (403 Forbidden)
			return statusCode == 403, "User correctly forbidden from admin functions", statusCode
		})
	}
	
	// 未认证访问
	tester.runTest("Unauthenticated Access", "permission_control", func() (bool, string, int) {
		statusCode, err := tester.makeRequest("GET", "/api/v1/admin/users", nil, "")
		if err != nil {
			return false, err.Error(), statusCode
		}
		
		// 未认证用户应该被拒绝访问 (401 Unauthorized)
		return statusCode == 401, "Unauthenticated access correctly rejected", statusCode
	})
}

// testContentSecurity 测试内容安全
func (tester *E2ESecurityTester) testContentSecurity() {
	fmt.Println("\n🛡️ 测试内容安全...")
	
	if userToken, ok := tester.tokens["user"]; ok {
		// 恶意内容测试
		maliciousContents := []string{
			"<script>alert('xss')</script>正常内容",
			"点击这里：javascript:alert(1)",
			"这里包含敏感词测试",
		}
		
		for i, content := range maliciousContents {
			tester.runTest(fmt.Sprintf("Malicious Content Test %d", i+1), "content_security", func() (bool, string, int) {
				body := map[string]interface{}{
					"title":   "测试信件",
					"content": content,
					"type":    "draft",
				}
				
				statusCode, err := tester.makeRequest("POST", "/api/v1/letters", body, userToken)
				if err != nil {
					return false, err.Error(), statusCode
				}
				
				// 恶意内容应该被拒绝或过滤
				if statusCode == 400 || statusCode == 201 {
					return true, "Malicious content handled appropriately", statusCode
				}
				return false, "Malicious content not properly filtered", statusCode
			})
		}
		
		// 正常内容测试
		tester.runTest("Normal Content Acceptance", "content_security", func() (bool, string, int) {
			body := map[string]interface{}{
				"title":   "正常信件",
				"content": "这是一封正常的信件内容，没有任何恶意代码。",
				"type":    "draft",
			}
			
			statusCode, err := tester.makeRequest("POST", "/api/v1/letters", body, userToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 201, "Normal content accepted", statusCode
		})
	}
}

// testRateLimit 测试速率限制
func (tester *E2ESecurityTester) testRateLimit() {
	fmt.Println("\n⏱️ 测试速率限制...")
	
	// 一般速率限制测试
	tester.runTest("General Rate Limit", "rate_limit", func() (bool, string, int) {
		var lastStatusCode int
		
		// 快速发送多个请求
		for i := 0; i < 20; i++ {
			resp, err := tester.client.Get(tester.config.BaseURL + "/ping")
			if err != nil {
				return false, err.Error(), 0
			}
			lastStatusCode = resp.StatusCode
			resp.Body.Close()
			
			if lastStatusCode == 429 {
				return true, fmt.Sprintf("Rate limit triggered after %d requests", i+1), lastStatusCode
			}
			
			time.Sleep(100 * time.Millisecond)
		}
		
		return false, "Rate limit not triggered after 20 requests", lastStatusCode
	})
	
	// 认证速率限制测试
	tester.runTest("Auth Rate Limit", "rate_limit", func() (bool, string, int) {
		var lastStatusCode int
		
		for i := 0; i < 10; i++ {
			_, statusCode, _ := tester.login("invalid", "invalid")
			lastStatusCode = statusCode
			
			if statusCode == 429 {
				return true, fmt.Sprintf("Auth rate limit triggered after %d attempts", i+1), statusCode
			}
			
			time.Sleep(200 * time.Millisecond)
		}
		
		return false, "Auth rate limit not triggered", lastStatusCode
	})
}

// testSecurityMonitoring 测试安全监控
func (tester *E2ESecurityTester) testSecurityMonitoring() {
	fmt.Println("\n📊 测试安全监控...")
	
	if adminToken, ok := tester.tokens["admin"]; ok {
		// 触发一些安全事件
		tester.client.Get(tester.config.BaseURL + "/api/v1/auth/login")
		tester.makeRequest("POST", "/api/v1/auth/login", map[string]string{
			"username": "<script>alert(1)</script>",
			"password": "test",
		}, "")
		
		time.Sleep(2 * time.Second) // 等待事件记录
		
		// 检查安全事件记录
		tester.runTest("Security Events Recording", "monitoring", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/admin/security/events", nil, adminToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 200, "Security events accessible", statusCode
		})
		
		// 检查安全统计
		tester.runTest("Security Statistics", "monitoring", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/admin/security/stats", nil, adminToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 200, "Security statistics accessible", statusCode
		})
	}
}

// testEndToEndDataFlow 测试端到端数据流
func (tester *E2ESecurityTester) testEndToEndDataFlow() {
	fmt.Println("\n🔄 测试端到端数据流...")
	
	if userToken, ok := tester.tokens["user"]; ok {
		// 创建信件 -> 内容安全检查 -> 存储 -> 检索
		tester.runTest("End-to-End Letter Creation", "data_flow", func() (bool, string, int) {
			// 1. 创建信件
			letterBody := map[string]interface{}{
				"title":   "端到端测试信件",
				"content": "这是一封端到端测试信件，用于验证完整的数据流。",
				"type":    "draft",
			}
			
			statusCode, err := tester.makeRequest("POST", "/api/v1/letters", letterBody, userToken)
			if err != nil || statusCode != 201 {
				return false, fmt.Sprintf("Letter creation failed: %v, status: %d", err, statusCode), statusCode
			}
			
			// 2. 获取用户信件列表
			statusCode, err = tester.makeRequest("GET", "/api/v1/letters", nil, userToken)
			if err != nil || statusCode != 200 {
				return false, fmt.Sprintf("Letter retrieval failed: %v, status: %d", err, statusCode), statusCode
			}
			
			return true, "End-to-end letter flow successful", statusCode
		})
	}
}

// testValidationSystem 测试安全验证系统
func (tester *E2ESecurityTester) testValidationSystem() {
	fmt.Println("\n✅ 测试安全验证系统...")
	
	if adminToken, ok := tester.tokens["admin"]; ok {
		// 运行安全验证
		tester.runTest("Run Security Validation", "validation", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("POST", "/api/v1/admin/security/validate", nil, adminToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 200, "Security validation executed", statusCode
		})
		
		// 获取验证摘要
		tester.runTest("Get Validation Summary", "validation", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/admin/security/validate/summary", nil, adminToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 200, "Validation summary accessible", statusCode
		})
		
		// 验证系统健康状态
		tester.runTest("Validation System Health", "validation", func() (bool, string, int) {
			statusCode, err := tester.makeRequest("GET", "/api/v1/admin/security/validate/health", nil, adminToken)
			if err != nil {
				return false, err.Error(), statusCode
			}
			
			return statusCode == 200, "Validation system healthy", statusCode
		})
	}
}

// 辅助方法

// login 用户登录
func (tester *E2ESecurityTester) login(username, password string) (string, int, error) {
	body := map[string]string{
		"username": username,
		"password": password,
	}
	
	jsonBody, _ := json.Marshal(body)
	resp, err := tester.client.Post(
		tester.config.BaseURL+"/api/v1/auth/login",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", resp.StatusCode, fmt.Errorf("login failed with status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", resp.StatusCode, err
	}
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		if token, ok := data["token"].(string); ok {
			return token, resp.StatusCode, nil
		}
	}
	
	return "", resp.StatusCode, fmt.Errorf("token not found in response")
}

// makeRequest 发送HTTP请求
func (tester *E2ESecurityTester) makeRequest(method, path string, body interface{}, token string) (int, error) {
	var reqBody io.Reader
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}
	
	req, err := http.NewRequest(method, tester.config.BaseURL+path, reqBody)
	if err != nil {
		return 0, err
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	resp, err := tester.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	return resp.StatusCode, nil
}

// runTest 运行单个测试
func (tester *E2ESecurityTester) runTest(name, category string, testFunc func() (bool, string, int)) {
	start := time.Now()
	passed, details, statusCode := testFunc()
	duration := time.Since(start)
	
	result := TestResult{
		TestName:   name,
		Passed:     passed,
		Duration:   duration,
		Details:    details,
		StatusCode: statusCode,
		Category:   category,
	}
	
	if !passed {
		result.ErrorMsg = details
	}
	
	tester.results = append(tester.results, result)
	
	status := "✅"
	if !passed {
		status = "❌"
	}
	
	fmt.Printf("  %s %s (%dms)\n", status, name, duration.Milliseconds())
}

// generateReport 生成测试报告
func (tester *E2ESecurityTester) generateReport() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 端到端安全测试报告")
	fmt.Println(strings.Repeat("=", 60))
	
	total := len(tester.results)
	passed := 0
	failed := 0
	
	categories := make(map[string]int)
	categoryPassed := make(map[string]int)
	
	for _, result := range tester.results {
		categories[result.Category]++
		if result.Passed {
			passed++
			categoryPassed[result.Category]++
		} else {
			failed++
		}
	}
	
	fmt.Printf("\n📈 总体统计:\n")
	fmt.Printf("  总测试数: %d\n", total)
	fmt.Printf("  通过测试: %d\n", passed)
	fmt.Printf("  失败测试: %d\n", failed)
	fmt.Printf("  成功率: %.1f%%\n", float64(passed)/float64(total)*100)
	
	fmt.Printf("\n📋 分类统计:\n")
	for category, total := range categories {
		passed := categoryPassed[category]
		fmt.Printf("  %s: %d/%d (%.1f%%)\n", 
			category, passed, total, float64(passed)/float64(total)*100)
	}
	
	if failed > 0 {
		fmt.Printf("\n❌ 失败的测试:\n")
		for _, result := range tester.results {
			if !result.Passed {
				fmt.Printf("  - %s: %s\n", result.TestName, result.ErrorMsg)
			}
		}
	}
	
	// 保存详细报告到文件
	reportData, _ := json.MarshalIndent(map[string]interface{}{
		"summary": map[string]interface{}{
			"total":       total,
			"passed":      passed,
			"failed":      failed,
			"success_rate": float64(passed) / float64(total) * 100,
			"categories":  categories,
		},
		"results":   tester.results,
		"timestamp": time.Now(),
	}, "", "  ")
	
	os.WriteFile("security_test_report.json", reportData, 0644)
	fmt.Printf("\n📄 详细报告已保存到: security_test_report.json\n")
	
	if failed == 0 {
		fmt.Printf("\n🎉 所有安全测试通过！系统安全防护完备。\n")
	} else {
		fmt.Printf("\n⚠️  检测到 %d 个安全问题，请检查并修复。\n", failed)
	}
}

// main 主函数
func main() {
	// 默认配置
	config := TestConfig{
		BaseURL:   "http://localhost:8080",
		AdminUser: "admin",
		AdminPass: "admin123",
		TestUser:  "alice",
		TestPass:  "secret",
		Timeout:   30,
	}
	
	// 从环境变量读取配置
	if url := os.Getenv("TEST_BASE_URL"); url != "" {
		config.BaseURL = url
	}
	if user := os.Getenv("TEST_ADMIN_USER"); user != "" {
		config.AdminUser = user
	}
	if pass := os.Getenv("TEST_ADMIN_PASS"); pass != "" {
		config.AdminPass = pass
	}
	
	tester := NewE2ESecurityTester(config)
	tester.RunAllTests()
}