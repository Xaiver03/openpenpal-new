# 🧪 OpenPenPal 测试指南

> 本指南旨在帮助开发团队系统性地提升 OpenPenPal 项目的测试覆盖率，从当前的 <10% 提升到 80% 以上。

## 📋 目录

1. [测试策略概述](#测试策略概述)
2. [测试环境搭建](#测试环境搭建)
3. [后端测试指南](#后端测试指南)
4. [前端测试指南](#前端测试指南)
5. [E2E测试指南](#e2e测试指南)
6. [测试最佳实践](#测试最佳实践)
7. [CI/CD集成](#cicd集成)
8. [测试计划时间表](#测试计划时间表)

## 测试策略概述

### 测试金字塔

```
         /\
        /E2E\         (10%) - 用户场景测试
       /------\
      /  集成  \      (20%) - API和服务集成测试
     /----------\
    /   单元测试   \   (70%) - 函数和组件级测试
   /--------------\
```

### 测试优先级

1. **P0 - 立即实施**（第1-2周）
   - 认证授权系统
   - 四级信使核心功能
   - 信件管理基础功能

2. **P1 - 核心功能**（第3-4周）
   - 支付系统
   - OP Code系统
   - 实时通信功能

3. **P2 - 完善覆盖**（第5-8周）
   - 管理后台
   - 辅助功能
   - 性能测试

## 测试环境搭建

### 1. 安装测试依赖

#### 后端（Go）
```bash
# 测试框架
go get -u github.com/stretchr/testify
go get -u github.com/golang/mock/mockgen
go get -u github.com/DATA-DOG/go-sqlmock

# 在项目根目录创建 Makefile
cat > Makefile << 'EOF'
.PHONY: test test-coverage mock

test:
	go test -v ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

mock:
	mockgen -source=internal/services/auth_service.go -destination=internal/mocks/mock_auth_service.go -package=mocks
EOF
```

#### 前端（TypeScript/React）
```bash
cd frontend

# 测试工具已安装，确认版本
npm list @testing-library/react @testing-library/jest-dom jest

# 如需更新
npm install --save-dev @testing-library/react@latest
npm install --save-dev @testing-library/user-event@latest
npm install --save-dev @testing-library/jest-dom@latest
```

### 2. 配置测试数据库

创建测试配置文件：
```bash
# backend/internal/config/test_config.go
cat > backend/internal/config/test_config.go << 'EOF'
package config

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	// 自动迁移测试所需的表
	// db.AutoMigrate(&models.User{}, &models.Letter{}, ...)
	
	return db, nil
}

func GetTestConfig() *Config {
	return &Config{
		JWTSecret:        "test-secret-key",
		DatabaseType:     "sqlite",
		DatabaseDSN:      ":memory:",
		AppEnv:           "test",
		FrontendURL:      "http://localhost:3000",
		ServerPort:       "8080",
		CSRFSecret:       "test-csrf-secret",
		UploadDir:        "/tmp/test-uploads",
		MaxUploadSize:    10 * 1024 * 1024, // 10MB
		AllowedFileTypes: []string{".jpg", ".jpeg", ".png", ".pdf"},
	}
}
EOF
```

### 3. 创建测试工具函数

```bash
# backend/internal/testutils/helpers.go
mkdir -p backend/internal/testutils
cat > backend/internal/testutils/helpers.go << 'EOF'
package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupTestRouter 创建测试路由
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	return router
}

// MakeRequest 执行HTTP请求
func MakeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
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
EOF
```

## 后端测试指南

### 1. Handler 测试示例

创建第一个测试文件：
```bash
# backend/internal/handlers/auth_handler_test.go
cat > backend/internal/handlers/auth_handler_test.go << 'EOF'
package handlers

import (
	"net/http"
	"testing"
	
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/testutils"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService 模拟认证服务
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(username, password string) (*models.User, string, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.User), args.String(1), args.Error(2)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	// 1. 设置
	cfg := config.GetTestConfig()
	db, _ := config.SetupTestDB()
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService, nil, cfg, db)
	router := testutils.SetupTestRouter()
	router.POST("/api/v1/auth/login", handler.Login)
	
	// 2. 准备测试数据
	user := &models.User{
		ID:       "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}
	
	// 3. 设置 Mock 期望
	mockService.On("Login", "testuser", "password123").
		Return(user, "mock-jwt-token", nil)
	
	// 4. 执行请求
	body := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	w := testutils.MakeRequest(t, router, "POST", "/api/v1/auth/login", body)
	
	// 5. 断言
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	testutils.ParseResponse(t, w, &response)
	
	assert.True(t, response["success"].(bool))
	assert.Equal(t, "登录成功", response["message"])
	assert.NotEmpty(t, response["data"].(map[string]interface{})["token"])
	
	// 6. 验证 Mock 调用
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	// 测试无效凭证的情况
	// ... 类似的测试结构
}

func TestAuthHandler_Login_EmptyFields(t *testing.T) {
	// 测试空字段验证
	// ... 类似的测试结构
}
EOF
```

### 2. Service 测试示例

```bash
# backend/internal/services/letter_service_test.go
cat > backend/internal/services/letter_service_test.go << 'EOF'
package services

import (
	"testing"
	"time"
	
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LetterServiceTestSuite 使用 testify suite 组织测试
type LetterServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *LetterService
	user    *models.User
}

// SetupSuite 在所有测试前运行
func (suite *LetterServiceTestSuite) SetupSuite() {
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db
	
	// 迁移测试表
	err = db.AutoMigrate(&models.User{}, &models.Letter{})
	suite.NoError(err)
	
	cfg := config.GetTestConfig()
	suite.service = NewLetterService(db, cfg)
}

// SetupTest 在每个测试前运行
func (suite *LetterServiceTestSuite) SetupTest() {
	// 创建测试用户
	suite.user = &models.User{
		ID:       "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
		IsActive: true,
	}
	suite.db.Create(suite.user)
}

// TearDownTest 在每个测试后运行
func (suite *LetterServiceTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM letters")
	suite.db.Exec("DELETE FROM users")
}

// TestCreateLetter 测试创建信件
func (suite *LetterServiceTestSuite) TestCreateLetter() {
	// 准备
	req := &models.CreateLetterRequest{
		Title:   "测试信件",
		Content: "这是一封测试信件",
		Style:   models.StyleClassic,
	}
	
	// 执行
	letter, err := suite.service.CreateLetter(suite.user.ID, req)
	
	// 断言
	suite.NoError(err)
	suite.NotNil(letter)
	suite.Equal("测试信件", letter.Title)
	suite.Equal("这是一封测试信件", letter.Content)
	suite.Equal(models.StatusDraft, letter.Status)
	suite.Equal(suite.user.ID, letter.UserID)
}

// TestGetLettersByUser 测试获取用户信件列表
func (suite *LetterServiceTestSuite) TestGetLettersByUser() {
	// 准备 - 创建多封信件
	for i := 0; i < 3; i++ {
		letter := &models.Letter{
			ID:      fmt.Sprintf("letter-%d", i),
			UserID:  suite.user.ID,
			Title:   fmt.Sprintf("信件 %d", i),
			Content: "测试内容",
			Status:  models.StatusDraft,
		}
		suite.db.Create(letter)
	}
	
	// 执行
	letters, total, err := suite.service.GetLettersByUser(suite.user.ID, 1, 10, "")
	
	// 断言
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(letters, 3)
}

// TestUpdateLetterStatus 测试更新信件状态
func (suite *LetterServiceTestSuite) TestUpdateLetterStatus() {
	// 准备
	letter := &models.Letter{
		ID:      "test-letter-id",
		UserID:  suite.user.ID,
		Title:   "测试信件",
		Content: "测试内容",
		Status:  models.StatusDraft,
	}
	suite.db.Create(letter)
	
	// 执行
	err := suite.service.UpdateLetterStatus(letter.ID, models.StatusGenerated)
	
	// 断言
	suite.NoError(err)
	
	// 验证状态已更新
	var updated models.Letter
	suite.db.First(&updated, "id = ?", letter.ID)
	suite.Equal(models.StatusGenerated, updated.Status)
}

// 运行测试套件
func TestLetterServiceSuite(t *testing.T) {
	suite.Run(t, new(LetterServiceTestSuite))
}
EOF
```

### 3. 四级信使系统测试

```bash
# backend/internal/services/courier_hierarchy_test.go
cat > backend/internal/services/courier_hierarchy_test.go << 'EOF'
package services

import (
	"testing"
	
	"openpenpal-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCourierHierarchy_Permissions(t *testing.T) {
	tests := []struct {
		name           string
		courierLevel   int
		targetOPCode   string
		ownOPCode      string
		expectedResult bool
	}{
		{
			name:           "L4可以访问任何区域",
			courierLevel:   4,
			targetOPCode:   "QH3B02",
			ownOPCode:      "BJ0000",
			expectedResult: true,
		},
		{
			name:           "L3可以访问同学校",
			courierLevel:   3,
			targetOPCode:   "PK5F01",
			ownOPCode:      "PK0000",
			expectedResult: true,
		},
		{
			name:           "L3不能访问其他学校",
			courierLevel:   3,
			targetOPCode:   "QH3B02",
			ownOPCode:      "PK0000",
			expectedResult: false,
		},
		{
			name:           "L1只能访问相同前缀",
			courierLevel:   1,
			targetOPCode:   "PK5F01",
			ownOPCode:      "PK5F",
			expectedResult: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试服务
			service := &CourierService{}
			
			// 执行权限检查
			result := service.ValidateOPCodeAccess(
				tt.courierLevel,
				tt.targetOPCode,
				tt.ownOPCode,
			)
			
			// 断言
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCourierHierarchy_TaskAssignment(t *testing.T) {
	// 测试任务分配算法
	t.Run("应该优先分配给最近的信使", func(t *testing.T) {
		// TODO: 实现测试
	})
	
	t.Run("应该考虑信使负载均衡", func(t *testing.T) {
		// TODO: 实现测试
	})
	
	t.Run("应该遵守层级限制", func(t *testing.T) {
		// TODO: 实现测试
	})
}
EOF
```

## 前端测试指南

### 1. 组件测试示例

```bash
# frontend/src/components/letter/LetterEditor.test.tsx
cat > frontend/src/components/letter/LetterEditor.test.tsx << 'EOF'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LetterEditor } from './LetterEditor'
import { useLetterStore } from '@/stores/letter-store'
import { api } from '@/lib/api'

// Mock 依赖
jest.mock('@/stores/letter-store')
jest.mock('@/lib/api')

describe('LetterEditor', () => {
  const mockSaveDraft = jest.fn()
  const mockAutoSave = jest.fn()
  
  beforeEach(() => {
    // 重置 mocks
    jest.clearAllMocks()
    
    // 设置 store mock
    (useLetterStore as jest.Mock).mockReturnValue({
      currentLetter: null,
      saveDraft: mockSaveDraft,
      autoSave: mockAutoSave,
    })
  })
  
  it('应该渲染编辑器界面', () => {
    render(<LetterEditor />)
    
    expect(screen.getByPlaceholderText('给你的信起个标题...')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('开始写信...')).toBeInTheDocument()
    expect(screen.getByText('保存草稿')).toBeInTheDocument()
  })
  
  it('应该处理标题输入', async () => {
    const user = userEvent.setup()
    render(<LetterEditor />)
    
    const titleInput = screen.getByPlaceholderText('给你的信起个标题...')
    await user.type(titleInput, '给朋友的信')
    
    expect(titleInput).toHaveValue('给朋友的信')
  })
  
  it('应该自动保存内容', async () => {
    jest.useFakeTimers()
    const user = userEvent.setup({ delay: null })
    
    render(<LetterEditor />)
    
    const contentInput = screen.getByPlaceholderText('开始写信...')
    await user.type(contentInput, '亲爱的朋友...')
    
    // 快进时间触发自动保存
    jest.advanceTimersByTime(3000)
    
    await waitFor(() => {
      expect(mockAutoSave).toHaveBeenCalledWith({
        content: '亲爱的朋友...',
      })
    })
    
    jest.useRealTimers()
  })
  
  it('应该处理保存草稿', async () => {
    const user = userEvent.setup()
    mockSaveDraft.mockResolvedValue({ success: true })
    
    render(<LetterEditor />)
    
    // 输入内容
    await user.type(screen.getByPlaceholderText('给你的信起个标题...'), '测试信件')
    await user.type(screen.getByPlaceholderText('开始写信...'), '测试内容')
    
    // 点击保存
    await user.click(screen.getByText('保存草稿'))
    
    expect(mockSaveDraft).toHaveBeenCalledWith({
      title: '测试信件',
      content: '测试内容',
    })
  })
  
  it('应该显示字数统计', async () => {
    const user = userEvent.setup()
    render(<LetterEditor />)
    
    const contentInput = screen.getByPlaceholderText('开始写信...')
    await user.type(contentInput, '这是一封测试信件')
    
    expect(screen.getByText('8 字')).toBeInTheDocument()
  })
})
EOF
```

### 2. Hook 测试示例

```bash
# frontend/src/hooks/useAuth.test.ts
cat > frontend/src/hooks/useAuth.test.ts << 'EOF'
import { renderHook, act } from '@testing-library/react'
import { useAuth } from './useAuth'
import { AuthService } from '@/lib/services/auth-service'

jest.mock('@/lib/services/auth-service')

describe('useAuth', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
  })
  
  it('应该初始化为未认证状态', () => {
    const { result } = renderHook(() => useAuth())
    
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
    expect(result.current.isLoading).toBe(true)
  })
  
  it('应该处理登录成功', async () => {
    const mockUser = {
      id: 'test-user-id',
      username: 'testuser',
      role: 'user',
    }
    
    (AuthService.login as jest.Mock).mockResolvedValue({
      success: true,
      data: {
        user: mockUser,
        token: 'test-token',
      },
    })
    
    const { result } = renderHook(() => useAuth())
    
    await act(async () => {
      await result.current.login({
        username: 'testuser',
        password: 'password123',
      })
    })
    
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toEqual(mockUser)
  })
  
  it('应该处理登出', async () => {
    const { result } = renderHook(() => useAuth())
    
    // 先登录
    act(() => {
      result.current.setUser({
        id: 'test-user-id',
        username: 'testuser',
        role: 'user',
      })
    })
    
    expect(result.current.isAuthenticated).toBe(true)
    
    // 登出
    await act(async () => {
      await result.current.logout()
    })
    
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
  })
})
EOF
```

### 3. Store 测试示例

```bash
# frontend/src/stores/courier-store.test.ts
cat > frontend/src/stores/courier-store.test.ts << 'EOF'
import { act, renderHook } from '@testing-library/react'
import { useCourierStore } from './courier-store'
import { api } from '@/lib/api'

jest.mock('@/lib/api')

describe('CourierStore', () => {
  beforeEach(() => {
    const { result } = renderHook(() => useCourierStore.getState())
    act(() => {
      result.current.reset()
    })
  })
  
  it('应该获取待处理任务', async () => {
    const mockTasks = [
      {
        id: 'task-1',
        letterCode: 'LC123456',
        status: 'pending',
        pickupOPCode: 'PK5F01',
        deliveryOPCode: 'PK3D12',
      },
    ]
    
    (api.get as jest.Mock).mockResolvedValue({
      data: { success: true, data: mockTasks },
    })
    
    const { result } = renderHook(() => useCourierStore())
    
    await act(async () => {
      await result.current.fetchPendingTasks()
    })
    
    expect(result.current.pendingTasks).toEqual(mockTasks)
    expect(result.current.isLoading).toBe(false)
  })
  
  it('应该处理任务接受', async () => {
    (api.post as jest.Mock).mockResolvedValue({
      data: { success: true },
    })
    
    const { result } = renderHook(() => useCourierStore())
    
    await act(async () => {
      await result.current.acceptTask('task-1')
    })
    
    expect(api.post).toHaveBeenCalledWith('/api/v1/courier/tasks/task-1/accept')
  })
})
EOF
```

## E2E测试指南

### 1. 完整用户流程测试

```bash
# frontend/tests/e2e/tests/complete-flow.spec.ts
cat > frontend/tests/e2e/tests/complete-flow.spec.ts << 'EOF'
import { test, expect } from '@playwright/test'
import { LoginPage } from '../pages/login.page'
import { DashboardPage } from '../pages/dashboard.page'
import { LetterPage } from '../pages/letter.page'
import { testUsers } from '../fixtures/test-users'

test.describe('完整信件投递流程', () => {
  test('用户写信 -> 信使取件 -> 投递完成', async ({ page }) => {
    // 1. 用户登录并写信
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(testUsers.writer.username, testUsers.writer.password)
    
    const dashboard = new DashboardPage(page)
    await dashboard.waitForLoad()
    await dashboard.navigateToWriteLetter()
    
    const letterPage = new LetterPage(page)
    await letterPage.writeLetter({
      title: 'E2E测试信件',
      content: '这是一封端到端测试信件',
      recipientOPCode: 'PK3D12',
    })
    await letterPage.submitLetter()
    
    // 获取信件编号
    const letterCode = await letterPage.getLetterCode()
    expect(letterCode).toMatch(/^LC\d{6}$/)
    
    // 2. 登出并切换到信使账号
    await dashboard.logout()
    
    await loginPage.login(testUsers.courier.username, testUsers.courier.password)
    await dashboard.waitForLoad()
    await dashboard.navigateToCourierTasks()
    
    // 3. 信使接受任务
    const courierPage = await page.locator('.courier-dashboard')
    await courierPage.locator(`[data-letter-code="${letterCode}"]`).click()
    await page.getByRole('button', { name: '接受任务' }).click()
    
    // 4. 扫码取件
    await page.getByRole('button', { name: '扫码取件' }).click()
    // 模拟扫码
    await page.fill('[data-testid="qr-input"]', letterCode)
    await page.getByRole('button', { name: '确认取件' }).click()
    
    // 5. 投递完成
    await page.getByRole('button', { name: '扫码投递' }).click()
    await page.fill('[data-testid="qr-input"]', letterCode)
    await page.getByRole('button', { name: '确认投递' }).click()
    
    // 6. 验证状态更新
    await expect(page.locator('.task-status')).toHaveText('已完成')
  })
})
EOF
```

### 2. 四级信使系统测试

```bash
# frontend/tests/e2e/tests/courier-hierarchy.spec.ts
cat > frontend/tests/e2e/tests/courier-hierarchy.spec.ts << 'EOF'
import { test, expect } from '@playwright/test'

test.describe('四级信使权限系统', () => {
  test('L4信使可以创建L3信使', async ({ page }) => {
    // 使用L4信使账号登录
    await page.goto('/login')
    await page.fill('[name="username"]', 'courier_level4')
    await page.fill('[name="password"]', 'secret')
    await page.click('[type="submit"]')
    
    // 导航到信使管理
    await page.click('[data-testid="courier-management"]')
    
    // 创建L3信使
    await page.click('[data-testid="create-courier"]')
    await page.fill('[name="username"]', 'new_l3_courier')
    await page.fill('[name="name"]', '新L3信使')
    await page.selectOption('[name="level"]', '3')
    await page.selectOption('[name="school"]', 'BJDX')
    await page.click('[type="submit"]')
    
    // 验证创建成功
    await expect(page.locator('.success-message')).toBeVisible()
  })
  
  test('L1信使只能查看自己区域的任务', async ({ page }) => {
    // L1信使登录
    await page.goto('/login')
    await page.fill('[name="username"]', 'courier_level1')
    await page.fill('[name="password"]', 'secret')
    await page.click('[type="submit"]')
    
    // 查看任务列表
    await page.click('[data-testid="courier-tasks"]')
    
    // 验证只显示PK5F前缀的任务
    const tasks = page.locator('.task-item')
    const count = await tasks.count()
    
    for (let i = 0; i < count; i++) {
      const opCode = await tasks.nth(i).locator('.op-code').textContent()
      expect(opCode).toMatch(/^PK5F/)
    }
  })
})
EOF
```

## 测试最佳实践

### 1. 测试命名规范

```go
// Go测试命名
TestFunctionName_StateUnderTest_ExpectedBehavior
// 例如：
TestLogin_WithValidCredentials_ReturnsToken
TestLogin_WithEmptyUsername_ReturnsError

// TypeScript测试命名
describe('ComponentName', () => {
  it('should behavior when condition', () => {})
  it('should handle error when invalid input', () => {})
})
```

### 2. 测试数据管理

```bash
# backend/internal/testdata/fixtures.go
cat > backend/internal/testdata/fixtures.go << 'EOF'
package testdata

import (
	"openpenpal-backend/internal/models"
	"time"
)

// GetTestUser 获取测试用户
func GetTestUser(role models.UserRole) *models.User {
	return &models.User{
		ID:       "test-" + string(role),
		Username: "test_" + string(role),
		Email:    string(role) + "@test.com",
		Role:     role,
		IsActive: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestLetter 获取测试信件
func GetTestLetter(userID string) *models.Letter {
	return &models.Letter{
		ID:      "test-letter-id",
		UserID:  userID,
		Title:   "测试信件",
		Content: "测试内容",
		Status:  models.StatusDraft,
		Style:   models.StyleClassic,
	}
}
EOF
```

### 3. Mock 最佳实践

```go
// 使用接口而不是具体实现
type AuthService interface {
    Login(username, password string) (*models.User, string, error)
    Register(req *models.RegisterRequest) (*models.User, error)
}

// 生成 mock
//go:generate mockgen -source=auth_service.go -destination=../mocks/mock_auth_service.go
```

### 4. 测试隔离

```typescript
// 每个测试应该独立
beforeEach(() => {
  // 重置所有 mocks
  jest.clearAllMocks()
  
  // 清理 localStorage
  localStorage.clear()
  
  // 重置 store 状态
  useStore.setState(initialState)
})

afterEach(() => {
  // 清理副作用
  cleanup()
})
```

## CI/CD集成

### GitHub Actions 配置

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  backend-test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Install dependencies
      run: |
        cd backend
        go mod download
    
    - name: Run tests
      run: |
        cd backend
        go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./backend/coverage.out
        flags: backend
  
  frontend-test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    
    - name: Install dependencies
      run: |
        cd frontend
        npm ci
    
    - name: Run tests
      run: |
        cd frontend
        npm run test:coverage
    
    - name: Run E2E tests
      run: |
        cd frontend
        npx playwright install --with-deps
        npm run test:e2e
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./frontend/coverage/lcov.info
        flags: frontend
```

## 测试计划时间表

### 第1周：基础设施搭建
- [ ] 安装测试框架和依赖
- [ ] 配置测试环境
- [ ] 创建测试工具函数
- [ ] 设置 CI/CD

### 第2周：P0 核心功能测试
- [ ] 认证系统测试（5个测试文件）
- [ ] 信件基础功能测试（8个测试文件）
- [ ] 基础组件测试（10个测试文件）

### 第3周：信使系统测试
- [ ] 四级层级测试（6个测试文件）
- [ ] 任务分配测试（4个测试文件）
- [ ] 权限验证测试（5个测试文件）

### 第4周：集成测试
- [ ] API集成测试（10个测试文件）
- [ ] E2E核心流程（5个测试场景）
- [ ] 性能基准测试

### 第5-6周：完善覆盖
- [ ] 支付系统测试
- [ ] 管理后台测试
- [ ] 辅助功能测试

### 第7-8周：优化和文档
- [ ] 测试重构
- [ ] 性能优化
- [ ] 测试文档完善

## 测试覆盖率目标

| 时间节点 | 目标覆盖率 | 关键指标 |
|---------|-----------|---------|
| 第2周末 | 30% | 核心功能有测试 |
| 第4周末 | 60% | 主要API覆盖 |
| 第8周末 | 80% | 达到行业标准 |

## 总结

通过遵循本指南，团队可以系统性地提升 OpenPenPal 项目的测试覆盖率。记住：
- 先测试最关键的功能
- 保持测试简单和可维护
- 使用 mock 隔离依赖
- 持续运行和更新测试

良好的测试覆盖率是项目长期成功的关键保障。