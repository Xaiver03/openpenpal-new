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

// CourierHandlerTestSuite 信使处理器测试套件
type CourierHandlerTestSuite struct {
	suite.Suite
	db             *gorm.DB
	handler        *CourierHandler
	router         *gin.Engine
	courierService *services.CourierService
	userService    *services.UserService
	testUser       *models.User
	testCourier    *models.User
	testAdmin      *models.User
	config         *config.Config
}

func (suite *CourierHandlerTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.userService = services.NewUserService(db, suite.config)
	suite.courierService = services.NewCourierService(db)

	// 创建处理器
	suite.handler = NewCourierHandler(suite.courierService)

	// 创建测试用户
	suite.testUser = config.CreateTestUser(db, "courieruser", models.RoleUser)
	suite.testCourier = config.CreateTestUser(db, "courier_test", models.RoleCourier1)
	suite.testAdmin = config.CreateTestUser(db, "admin_test", models.RoleAdmin)

	// 设置路由
	suite.setupRoutes()
}

func (suite *CourierHandlerTestSuite) setupRoutes() {
	suite.router = testutils.SetupTestRouter()

	// 受保护路由（需要认证）
	authGroup := suite.router.Group("/api/v1/courier")
	authGroup.Use(middleware.AuthMiddleware(suite.config, suite.db))
	authGroup.POST("/apply", suite.handler.ApplyCourier)
	authGroup.GET("/status", suite.handler.GetCourierStatus)
	authGroup.GET("/profile", suite.handler.GetCourierProfile)
	authGroup.GET("/stats", suite.handler.GetCourierStats)
	authGroup.GET("/subordinates", suite.handler.GetSubordinates)
	authGroup.GET("/info", suite.handler.GetCourierInfo)
	authGroup.GET("/tasks", suite.handler.GetCourierTasks)

	// 管理员路由
	adminGroup := suite.router.Group("/api/v1/courier/admin")
	adminGroup.Use(middleware.AuthMiddleware(suite.config, suite.db))
	adminGroup.GET("/applications/pending", suite.handler.GetPendingApplications)
	adminGroup.PUT("/applications/:id/approve", suite.handler.ApproveCourierApplication)
	adminGroup.PUT("/applications/:id/reject", suite.handler.RejectCourierApplication)
	adminGroup.POST("/create", suite.handler.CreateCourier)
	adminGroup.GET("/candidates", suite.handler.GetCourierCandidates)

	// 等级统计路由
	levelGroup := suite.router.Group("/api/v1/courier/level")
	levelGroup.Use(middleware.AuthMiddleware(suite.config, suite.db))
	levelGroup.GET("/1/stats", suite.handler.GetFirstLevelStats)
	levelGroup.GET("/1/couriers", suite.handler.GetFirstLevelCouriers)
	levelGroup.GET("/2/stats", suite.handler.GetSecondLevelStats)
	levelGroup.GET("/2/couriers", suite.handler.GetSecondLevelCouriers)
	levelGroup.GET("/3/stats", suite.handler.GetThirdLevelStats)
	levelGroup.GET("/3/couriers", suite.handler.GetThirdLevelCouriers)
	levelGroup.GET("/4/stats", suite.handler.GetFourthLevelStats)
	levelGroup.GET("/4/couriers", suite.handler.GetFourthLevelCouriers)
}

func (suite *CourierHandlerTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM couriers")
	suite.db.Exec("DELETE FROM courier_tasks")
}

// TestApplyCourier_Success 测试成功申请信使
func (suite *CourierHandlerTestSuite) TestApplyCourier_Success() {
	// 准备请求体
	body := map[string]interface{}{
		"reason_for_application": "我想帮助同学传递信件",
		"contact":                "test@example.com",
		"available_time":         "周末",
		"delivery_area":          "北京大学",
		"zone":                   "BJDX",
		"contact_method":         "Email",
	}

	// 执行请求
	headers := suite.createAuthHeader(suite.testUser)
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/courier/apply", body, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	assert.NotEmpty(suite.T(), response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testUser.ID, data["user_id"])
	assert.Equal(suite.T(), "test@example.com", data["contact"])
	assert.Equal(suite.T(), "pending", data["status"])
	assert.Equal(suite.T(), float64(1), data["level"])
}

// TestApplyCourier_InvalidRequest 测试无效请求
func (suite *CourierHandlerTestSuite) TestApplyCourier_InvalidRequest() {
	// 准备无效请求体（缺少必要字段）
	body := map[string]interface{}{
		"reason_for_application": "我想申请",
		// 缺少其他必要字段
	}

	// 执行请求
	headers := suite.createAuthHeader(suite.testUser)
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/courier/apply", body, headers)

	// 断言：应该返回400错误
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestApplyCourier_Unauthorized 测试未授权访问
func (suite *CourierHandlerTestSuite) TestApplyCourier_Unauthorized() {
	// 准备请求体
	body := map[string]interface{}{
		"reason_for_application": "我想申请",
		"contact":                "test@example.com",
		"available_time":         "周末",
		"delivery_area":          "北京大学",
		"zone":                   "BJDX",
		"contact_method":         "Email",
	}

	// 执行请求（不带认证头）
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/courier/apply", body)

	// 断言：应该返回401未授权
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// TestGetCourierStatus_Success 测试获取信使状态
func (suite *CourierHandlerTestSuite) TestGetCourierStatus_Success() {
	// 先申请成为信使
	suite.createTestCourierApplication(suite.testUser)

	// 执行请求
	headers := suite.createAuthHeader(suite.testUser)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/status", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "pending", data["status"])
	assert.Equal(suite.T(), float64(1), data["level"])
}

// TestGetCourierStatus_NotApplied 测试未申请用户获取状态
func (suite *CourierHandlerTestSuite) TestGetCourierStatus_NotApplied() {
	// 执行请求（未申请过信使）
	headers := suite.createAuthHeader(suite.testUser)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/status", nil, headers)

	// 断言：应该返回404
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestGetCourierProfile_Success 测试获取信使档案
func (suite *CourierHandlerTestSuite) TestGetCourierProfile_Success() {
	// 创建并审批通过信使
	courier := suite.createApprovedCourier(suite.testUser)

	// 执行请求
	headers := suite.createAuthHeader(suite.testUser)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/profile", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), courier.Contact, data["contact"])
	assert.Equal(suite.T(), "approved", data["status"])
}

// TestGetCourierStats_Success 测试获取信使统计（需要管理员权限）
func (suite *CourierHandlerTestSuite) TestGetCourierStats_Success() {
	// 创建一些测试数据
	for i := 0; i < 3; i++ {
		user := config.CreateTestUser(suite.db, "stats_user_"+string(rune(i)), models.RoleUser)
		suite.createTestCourierApplication(user)
	}

	// 执行请求（使用管理员权限）
	headers := suite.createAuthHeader(suite.testAdmin)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/stats", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "pending_count")
	assert.Contains(suite.T(), data, "total_count")
}

// TestGetPendingApplications_Success 测试获取待审批申请
func (suite *CourierHandlerTestSuite) TestGetPendingApplications_Success() {
	// 创建多个待审批申请
	for i := 0; i < 3; i++ {
		user := config.CreateTestUser(suite.db, "pending_user_"+string(rune(i)), models.RoleUser)
		suite.createTestCourierApplication(user)
	}

	// 执行请求（使用管理员权限）
	headers := suite.createAuthHeader(suite.testAdmin)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/admin/applications/pending", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 3)
}

// TestApproveCourierApplication_Success 测试审批通过申请
func (suite *CourierHandlerTestSuite) TestApproveCourierApplication_Success() {
	// 创建待审批申请
	courier := suite.createTestCourierApplication(suite.testUser)

	// 执行审批请求（使用管理员权限）
	headers := suite.createAuthHeader(suite.testAdmin)
	url := "/api/v1/courier/admin/applications/" + string(rune(courier.ID)) + "/approve"
	w := testutils.MakeRequest(suite.T(), suite.router, "PUT", url, nil, headers)

	// 断言
	testutils.AssertSuccessResponse(suite.T(), w, http.StatusOK)

	// 验证数据库中状态已更新
	var updatedCourier models.Courier
	err := suite.db.First(&updatedCourier, "id = ?", courier.ID).Error
	suite.NoError(err)
	assert.Equal(suite.T(), models.CourierStatusApproved, updatedCourier.Status)
}

// TestRejectCourierApplication_Success 测试拒绝申请
func (suite *CourierHandlerTestSuite) TestRejectCourierApplication_Success() {
	// 创建待审批申请
	courier := suite.createTestCourierApplication(suite.testUser)

	// 执行拒绝请求（使用管理员权限）
	headers := suite.createAuthHeader(suite.testAdmin)
	url := "/api/v1/courier/admin/applications/" + string(rune(courier.ID)) + "/reject"
	w := testutils.MakeRequest(suite.T(), suite.router, "PUT", url, nil, headers)

	// 断言
	testutils.AssertSuccessResponse(suite.T(), w, http.StatusOK)

	// 验证数据库中状态已更新
	var updatedCourier models.Courier
	err := suite.db.First(&updatedCourier, "id = ?", courier.ID).Error
	suite.NoError(err)
	assert.Equal(suite.T(), models.CourierStatusRejected, updatedCourier.Status)
}

// TestCreateCourier_Success 测试创建信使
func (suite *CourierHandlerTestSuite) TestCreateCourier_Success() {
	// 准备请求体
	body := map[string]interface{}{
		"username":               "new_courier",
		"password":               "password123",
		"email":                  "new@example.com",
		"school_code":            "BJDX",
		"level":                  1,
		"zone":                   "BJDX-A",
		"reason_for_application": "管理员创建",
		"contact":                "new@example.com",
		"available_time":         "全时",
		"delivery_area":          "北京大学A区",
		"contact_method":         "Email",
	}

	// 执行请求（使用管理员权限）
	headers := suite.createAuthHeader(suite.testAdmin)
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/courier/admin/create", body, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "new_courier", data["username"])
	assert.Equal(suite.T(), "new@example.com", data["email"])
}

// TestCreateCourier_Unauthorized 测试非管理员创建信使
func (suite *CourierHandlerTestSuite) TestCreateCourier_Unauthorized() {
	// 准备请求体
	body := map[string]interface{}{
		"username": "unauthorized_courier",
		"email":    "unauth@example.com",
		"level":    1,
	}

	// 执行请求（使用普通用户权限）
	headers := suite.createAuthHeader(suite.testUser)
	w := testutils.MakeRequest(suite.T(), suite.router, "POST", "/api/v1/courier/admin/create", body, headers)

	// 断言：应该被拒绝（具体状态码取决于权限验证实现）
	assert.NotEqual(suite.T(), http.StatusCreated, w.Code)
}

// TestGetSubordinates_Success 测试获取下级信使
func (suite *CourierHandlerTestSuite) TestGetSubordinates_Success() {
	// 创建高级信使
	parentCourier := config.CreateTestUser(suite.db, "parent_courier", models.RoleCourier3)

	// 创建一些下级信使（通过服务层模拟）
	for i := 0; i < 2; i++ {
		req := &models.CreateCourierRequest{
			Username:             "sub_courier_" + string(rune(i)),
			Password:             "password123",
			Email:                "sub" + string(rune(i)) + "@example.com",
			SchoolCode:           "BJDX",
			Level:                2,
			Zone:                 "BJDX-" + string(rune(i)),
			ReasonForApplication: "下级信使",
			Contact:              "sub" + string(rune(i)) + "@example.com",
			AvailableTime:        "全时",
			DeliveryArea:         "测试区域" + string(rune(i)),
			ContactMethod:        "Email",
		}
		_, err := suite.courierService.CreateSubordinateCourier(parentCourier, req)
		suite.NoError(err)
	}

	// 执行请求
	headers := suite.createAuthHeader(parentCourier)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/subordinates", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

// TestGetCourierInfo_Success 测试获取信使信息
func (suite *CourierHandlerTestSuite) TestGetCourierInfo_Success() {
	// 创建信使用户
	courier := suite.createApprovedCourier(suite.testCourier)

	// 执行请求
	headers := suite.createAuthHeader(suite.testCourier)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/info", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "level")
	assert.Contains(suite.T(), data, "can_create_levels")
}

// TestGetFirstLevelStats_Success 测试获取一级信使统计
func (suite *CourierHandlerTestSuite) TestGetFirstLevelStats_Success() {
	// 创建一些1级信使
	for i := 0; i < 3; i++ {
		user := config.CreateTestUser(suite.db, "level1_user_"+string(rune(i)), models.RoleCourier1)
		suite.createApprovedCourier(user)
	}

	// 执行请求
	headers := suite.createAuthHeader(suite.testAdmin)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/level/1/stats", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "total_count")
	assert.Contains(suite.T(), data, "active_count")
}

// TestGetFirstLevelCouriers_Success 测试获取一级信使列表
func (suite *CourierHandlerTestSuite) TestGetFirstLevelCouriers_Success() {
	// 创建一些1级信使
	for i := 0; i < 2; i++ {
		user := config.CreateTestUser(suite.db, "l1_courier_"+string(rune(i)), models.RoleCourier1)
		suite.createApprovedCourier(user)
	}

	// 执行请求
	headers := suite.createAuthHeader(suite.testAdmin)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/level/1/couriers", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

// TestGetCourierTasks_Success 测试获取信使任务
func (suite *CourierHandlerTestSuite) TestGetCourierTasks_Success() {
	// 创建信使
	courier := suite.createApprovedCourier(suite.testCourier)

	// 执行请求
	headers := suite.createAuthHeader(suite.testCourier)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/tasks", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
	// 任务数据可能为空，但结构应该正确
	assert.Contains(suite.T(), response, "data")
}

// TestGetCourierTasks_WithFilters 测试带过滤器的任务查询
func (suite *CourierHandlerTestSuite) TestGetCourierTasks_WithFilters() {
	// 创建信使
	suite.createApprovedCourier(suite.testCourier)

	// 执行带过滤器的请求
	headers := suite.createAuthHeader(suite.testCourier)
	w := testutils.MakeRequest(suite.T(), suite.router, "GET", "/api/v1/courier/tasks?status=available&priority=high&page=1&limit=10", nil, headers)

	// 断言
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	testutils.ParseResponse(suite.T(), w, &response)

	assert.True(suite.T(), response["success"].(bool))
}

// Helper methods

func (suite *CourierHandlerTestSuite) createAuthHeader(user *models.User) http.Header {
	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := auth.GenerateJWT(user.ID, user.Role, suite.config.JWTSecret, expiresAt)
	suite.NoError(err)

	return testutils.CreateAuthHeader(token)
}

func (suite *CourierHandlerTestSuite) createTestCourierApplication(user *models.User) *models.Courier {
	req := &models.CourierApplication{
		ReasonForApplication: "测试申请",
		Contact:              "test_" + user.Username + "@example.com",
		AvailableTime:        "周末",
		DeliveryArea:         "北京大学",
		Zone:                 "BJDX",
		ContactMethod:        "Email",
	}
	courier, err := suite.courierService.ApplyCourier(user.ID, req)
	suite.NoError(err)
	return courier
}

func (suite *CourierHandlerTestSuite) createApprovedCourier(user *models.User) *models.Courier {
	courier := suite.createTestCourierApplication(user)
	err := suite.courierService.ApproveCourier(courier.ID)
	suite.NoError(err)

	// 重新获取更新后的信使
	err = suite.db.First(courier, "id = ?", courier.ID).Error
	suite.NoError(err)

	return courier
}

// 运行测试套件
func TestCourierHandlerSuite(t *testing.T) {
	suite.Run(t, new(CourierHandlerTestSuite))
}
