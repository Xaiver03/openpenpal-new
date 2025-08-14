package handlers

import (
	"net/http"
	"testing"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/testutils"
	"openpenpal-backend/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LetterHandlerTestSuite 信件处理器测试套件
type LetterHandlerTestSuite struct {
	suite.Suite
	db              *gorm.DB
	handler         *LetterHandler
	router          *gin.Engine
	letterService   *services.LetterService
	userService     *services.UserService
	envelopeService *services.EnvelopeService
	testUser        *models.User
	config          *config.Config
}

func (suite *LetterHandlerTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.userService = services.NewUserService(db, suite.config)
	suite.letterService = services.NewLetterService(db, suite.config)
	suite.envelopeService = services.NewEnvelopeService(db)

	// 创建处理器
	suite.handler = NewLetterHandler(suite.letterService, suite.envelopeService)

	// 创建测试用户
	suite.testUser = config.CreateTestUser(db, "letteruser", models.RoleUser)

	// 设置路由
	suite.setupRoutes()
}

func (suite *LetterHandlerTestSuite) setupRoutes() {
	suite.router = testutils.SetupTestRouter()

	// 公开路由（无需认证）
	suite.router.GET("/api/v1/letters/read/:code", suite.handler.GetLetterByCode)
	suite.router.GET("/api/v1/letters/public", suite.handler.GetPublicLetters)

	// 受保护路由（需要认证）
	authGroup := suite.router.Group("/api/v1/letters")
	authGroup.Use(middleware.AuthMiddleware(suite.config, suite.db))
	authGroup.POST("/", suite.handler.CreateDraft)
	authGroup.GET("/", suite.handler.GetUserLetters)
	authGroup.GET("/:id", suite.handler.GetLetter)
	authGroup.PUT("/:id", suite.handler.UpdateLetter)
	authGroup.DELETE("/:id", suite.handler.DeleteLetter)
	authGroup.POST("/:id/generate-code", suite.handler.GenerateCode)
	authGroup.PUT("/:id/read", suite.handler.MarkAsRead)
}

func (suite *LetterHandlerTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM letters")
	suite.db.Exec("DELETE FROM letter_codes")
	suite.db.Exec("DELETE FROM status_logs")
}

// TestCreateDraft_Success 测试成功创建草稿
func (suite *LetterHandlerTestSuite) TestCreateDraft_Success() {
	// 准备请求体
	body := map[string]interface{}{
		"title":   "Test Letter",
		"content": "This is a test letter content.",
		"style":   "classic",
	}

	// 执行请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/letters/", body, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	assert.NotEmpty(suite.T(), response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Test Letter", data["title"])
	assert.Equal(suite.T(), "This is a test letter content.", data["content"])
	assert.Equal(suite.T(), "draft", data["status"])
	assert.Equal(suite.T(), suite.testUser.ID, data["user_id"])
}

// TestCreateDraft_EmptyTitle 测试空标题
func (suite *LetterHandlerTestSuite) TestCreateDraft_EmptyTitle() {
	// 准备请求体（空标题）
	body := map[string]interface{}{
		"title":   "",
		"content": "Content with empty title",
		"style":   "classic",
	}

	// 执行请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/letters/", body, headers)

	// 断言：应该接受空标题（草稿状态）
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)
	assert.True(suite.T(), response["success"].(bool))
}

// TestCreateDraft_Unauthorized 测试未授权访问
func (suite *LetterHandlerTestSuite) TestCreateDraft_Unauthorized() {
	// 准备请求体
	body := map[string]interface{}{
		"title":   "Test Letter",
		"content": "Content",
		"style":   "classic",
	}

	// 执行请求（不带认证头）
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/letters/", body)

	// 断言：应该返回401未授权
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// TestGenerateCode_Success 测试成功生成二维码
func (suite *LetterHandlerTestSuite) TestGenerateCode_Success() {
	// 先创建草稿
	letter := suite.createTestLetter()

	// 执行生成二维码请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/letters/"+letter.ID+"/generate-code", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	assert.NotEmpty(suite.T(), response["message"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(suite.T(), data["letter_code"])
	assert.NotEmpty(suite.T(), data["qr_code_url"])
}

// TestGenerateCode_LetterNotFound 测试信件不存在
func (suite *LetterHandlerTestSuite) TestGenerateCode_LetterNotFound() {
	// 执行请求（使用不存在的ID）
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/letters/nonexistent/generate-code", nil, headers)

	// 断言：应该返回500（服务内部错误，因为信件不存在）
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)
	assert.False(suite.T(), response["success"].(bool))
	assert.NotEmpty(suite.T(), response["error"])
}

// TestGetUserLetters_Success 测试获取用户信件列表
func (suite *LetterHandlerTestSuite) TestGetUserLetters_Success() {
	// 创建多封信件
	for i := 0; i < 3; i++ {
		suite.createTestLetter()
	}

	// 执行请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/letters/", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	letters := data["data"].([]interface{})
	assert.Len(suite.T(), letters, 3)

	// 验证分页信息
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), pagination["page"])
	assert.Equal(suite.T(), float64(20), pagination["limit"])
	assert.Equal(suite.T(), float64(3), pagination["total"])
}

// TestGetUserLetters_Pagination 测试分页功能
func (suite *LetterHandlerTestSuite) TestGetUserLetters_Pagination() {
	// 创建10封信件
	for i := 0; i < 10; i++ {
		suite.createTestLetter()
	}

	// 测试第一页
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/letters/?page=1&limit=5", nil, headers)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	letters := data["data"].([]interface{})
	assert.Len(suite.T(), letters, 5)

	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), pagination["page"])
	assert.Equal(suite.T(), float64(5), pagination["limit"])
	assert.Equal(suite.T(), float64(10), pagination["total"])
}

// TestGetLetter_Success 测试获取信件详情
func (suite *LetterHandlerTestSuite) TestGetLetter_Success() {
	// 创建测试信件
	letter := suite.createTestLetter()

	// 执行请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/letters/"+letter.ID, nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), letter.ID, data["id"])
	assert.Equal(suite.T(), letter.Title, data["title"])
	assert.Equal(suite.T(), letter.Content, data["content"])
}

// TestGetLetter_NotFound 测试信件不存在
func (suite *LetterHandlerTestSuite) TestGetLetter_NotFound() {
	// 执行请求（使用不存在的ID）
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/letters/nonexistent", nil, headers)

	// 断言：应该返回404
	testutils.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "")
}

// TestUpdateLetter_Success 测试成功更新信件
func (suite *LetterHandlerTestSuite) TestUpdateLetter_Success() {
	// 创建测试信件
	letter := suite.createTestLetter()

	// 准备更新数据
	body := map[string]interface{}{
		"title":   "Updated Title",
		"content": "Updated content",
		"style":   "modern",
	}

	// 执行请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "PUT", "/api/v1/letters/"+letter.ID, body, headers)

	// 断言
	testutils.AssertSuccessResponse(suite.T(), w, http.StatusOK)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Updated Title", data["title"])
	assert.Equal(suite.T(), "Updated content", data["content"])
	assert.Equal(suite.T(), "modern", data["style"])
}

// TestDeleteLetter_Success 测试成功删除信件
func (suite *LetterHandlerTestSuite) TestDeleteLetter_Success() {
	// 创建测试信件
	letter := suite.createTestLetter()

	// 执行删除请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "DELETE", "/api/v1/letters/"+letter.ID, nil, headers)

	// 断言
	testutils.AssertSuccessResponse(suite.T(), w, http.StatusOK)

	// 验证信件已被软删除
	var deletedLetter models.Letter
	err := suite.db.Unscoped().First(&deletedLetter, "id = ?", letter.ID).Error
	suite.NoError(err)
	suite.NotNil(deletedLetter.DeletedAt)
}

// TestGetLetterByCode_Success 测试通过代码获取信件（公开接口）
func (suite *LetterHandlerTestSuite) TestGetLetterByCode_Success() {
	// 创建测试信件并生成代码
	letter := suite.createTestLetter()
	letterCode, err := suite.letterService.GenerateCode(letter.ID)
	suite.NoError(err)

	// 执行公开请求（无需认证）
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/letters/read/"+letterCode.Code, nil)

	// 断言
	testutils.AssertSuccessResponse(suite.T(), w, http.StatusOK)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), letter.Title, data["title"])
	assert.Equal(suite.T(), letter.Content, data["content"])
}

// TestGetLetterByCode_NotFound 测试无效代码
func (suite *LetterHandlerTestSuite) TestGetLetterByCode_NotFound() {
	// 执行请求（使用不存在的代码）
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/letters/read/INVALID_CODE", nil)

	// 断言：应该返回404
	testutils.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "")
}

// TestMarkAsRead_Success 测试标记为已读
func (suite *LetterHandlerTestSuite) TestMarkAsRead_Success() {
	// 创建测试信件并生成代码
	letter := suite.createTestLetter()
	letterCode, err := suite.letterService.GenerateCode(letter.ID)
	suite.NoError(err)

	// 首先更新状态到delivered
	err = suite.letterService.UpdateStatus(letterCode.Code, &models.UpdateLetterStatusRequest{
		Status: models.StatusDelivered,
	}, "test-courier")
	suite.NoError(err)

	// 执行标记已读请求
	headers := suite.createAuthHeader()
	w := testutils.MakeRequest(suite.T(), suite.router, "PUT", "/api/v1/letters/"+letter.ID+"/read", nil, headers)

	// 断言
	testutils.AssertSuccessResponse(suite.T(), w, http.StatusOK)
}

// Helper methods

func (suite *LetterHandlerTestSuite) createAuthHeader() http.Header {
	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := auth.GenerateJWT(suite.testUser.ID, suite.testUser.Role, suite.config.JWTSecret, expiresAt)
	suite.NoError(err)

	return testutils.CreateAuthHeader(token)
}

func (suite *LetterHandlerTestSuite) createTestLetter() *models.Letter {
	req := &models.CreateLetterRequest{
		Title:   "Test Letter",
		Content: "Test content",
		Style:   models.StyleClassic,
	}

	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	suite.NoError(err)

	return letter
}

// 运行测试套件
func TestLetterHandlerSuite(t *testing.T) {
	suite.Run(t, new(LetterHandlerTestSuite))
}
