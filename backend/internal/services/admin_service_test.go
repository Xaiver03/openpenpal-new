package services

import (
	"testing"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// AdminServiceTestSuite 管理服务测试套件
type AdminServiceTestSuite struct {
	suite.Suite
	db           *gorm.DB
	adminService *AdminService
	config       *config.Config
	testUsers    []*models.User
	testLetters  []*models.Letter
	testAdmin    *models.User
}

func (suite *AdminServiceTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.adminService = NewAdminService(db, suite.config)

	// 创建测试管理员
	suite.testAdmin = config.CreateTestUser(db, "testadmin", models.RoleSuperAdmin)

	// 创建多个测试用户
	for i, username := range []string{"alice", "bob", "carol", "david"} {
		role := models.RoleUser
		if i == 0 {
			role = models.RoleCourierLevel1 // alice 是信使
		}
		user := config.CreateTestUser(db, username, role)
		suite.testUsers = append(suite.testUsers, user)
	}

	// 创建测试信件
	for i, user := range suite.testUsers {
		letter := &models.Letter{
			ID:        "test-letter-" + user.Username,
			UserID:    user.ID,
			Title:     "Test Letter " + user.Username,
			Content:   "This is a test letter content for " + user.Username,
			Style:     models.StyleClassic,
			Status:    []models.LetterStatus{models.StatusDraft, models.StatusGenerated, models.StatusInTransit, models.StatusDelivered}[i%4],
			CreatedAt: time.Now().AddDate(0, 0, -i),
			UpdatedAt: time.Now(),
		}
		db.Create(letter)
		suite.testLetters = append(suite.testLetters, letter)
	}
}

func (suite *AdminServiceTestSuite) TearDownTest() {
	// 清理非测试用户的数据
	suite.db.Where("username NOT IN ?", []string{"testadmin", "alice", "bob", "carol", "david"}).Delete(&models.User{})
	suite.db.Where("id NOT IN ?", []string{"test-letter-alice", "test-letter-bob", "test-letter-carol", "test-letter-david"}).Delete(&models.Letter{})
	suite.db.Delete(&models.MuseumItem{}, "1=1")
	suite.db.Delete(&models.EnvelopeOrder{}, "1=1")
	suite.db.Delete(&models.Notification{}, "1=1")
	suite.db.Delete(&models.EnvelopeDesign{}, "1=1")
}

// TestGetDashboardStats_Success 测试获取仪表板统计
func (suite *AdminServiceTestSuite) TestGetDashboardStats_Success() {
	stats, err := suite.adminService.GetDashboardStats()

	suite.NoError(err)
	suite.NotNil(stats)

	// 验证用户统计
	suite.GreaterOrEqual(stats.TotalUsers, int64(5)) // 至少有测试用户
	suite.GreaterOrEqual(stats.NewUsersToday, int64(0))

	// 验证信件统计
	suite.GreaterOrEqual(stats.TotalLetters, int64(4)) // 测试信件
	suite.GreaterOrEqual(stats.LettersToday, int64(0))

	// 验证信使统计（alice是信使）
	suite.GreaterOrEqual(stats.ActiveCouriers, int64(1))

	// 验证状态分布
	suite.NotNil(stats.LetterStatusDistribution)

	// 验证系统健康状态
	suite.NotNil(stats.SystemHealth)
	suite.Equal("healthy", stats.SystemHealth.DatabaseStatus)
	suite.Equal("running", stats.SystemHealth.ServiceStatus)
}

// TestGetRecentActivities_Success 测试获取最近活动
func (suite *AdminServiceTestSuite) TestGetRecentActivities_Success() {
	activities, err := suite.adminService.GetRecentActivities(10)

	suite.NoError(err)
	suite.NotNil(activities)
	suite.LessOrEqual(len(activities), 10)

	// 验证活动内容
	if len(activities) > 0 {
		activity := activities[0]
		suite.NotEmpty(activity.ID)
		suite.NotEmpty(activity.Type)
		suite.NotEmpty(activity.Description)
		suite.NotEmpty(activity.UserID)
		suite.False(activity.CreatedAt.IsZero())
	}
}

// TestGetRecentActivities_InvalidLimit 测试无效的限制参数
func (suite *AdminServiceTestSuite) TestGetRecentActivities_InvalidLimit() {
	// 测试负数限制
	activities, err := suite.adminService.GetRecentActivities(-1)
	suite.NoError(err)
	suite.LessOrEqual(len(activities), 20) // 应该使用默认值20

	// 测试过大的限制
	activities, err = suite.adminService.GetRecentActivities(200)
	suite.NoError(err)
	suite.LessOrEqual(len(activities), 20) // 应该使用默认值20
}

// TestInjectSeedData_Success 测试注入种子数据成功
func (suite *AdminServiceTestSuite) TestInjectSeedData_Success() {
	// 创建一个独立的测试数据库连接
	db, err := config.SetupTestDB()
	suite.NoError(err)
	
	// 使用独立的service实例
	adminService := NewAdminService(db, suite.config)

	err = adminService.InjectSeedData()
	suite.NoError(err)

	// 验证种子用户是否创建成功
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	suite.Equal(int64(5), userCount) // 应该有5个种子用户

	// 验证种子信件是否创建成功
	var letterCount int64
	db.Model(&models.Letter{}).Count(&letterCount)
	suite.Equal(int64(3), letterCount) // 应该有3封种子信件

	// 验证信封设计是否创建成功
	var designCount int64
	db.Model(&models.EnvelopeDesign{}).Count(&designCount)
	suite.Equal(int64(3), designCount) // 应该有3个信封设计

	// 验证博物馆展品是否创建成功
	var exhibitCount int64
	db.Model(&models.MuseumItem{}).Count(&exhibitCount)
	suite.Equal(int64(2), exhibitCount) // 应该有2个展品
}

// TestInjectSeedData_AlreadyExists 测试种子数据已存在
func (suite *AdminServiceTestSuite) TestInjectSeedData_AlreadyExists() {
	// 确保有超过5个用户来触发"已存在"逻辑
	for i := 0; i < 3; i++ {
		extraUser := config.CreateTestUser(suite.db, "extra"+string(rune(i+'1')), models.RoleUser)
		suite.NotNil(extraUser)
	}
	
	// 验证用户数量超过5个
	var userCount int64
	suite.db.Model(&models.User{}).Count(&userCount)
	suite.Greater(userCount, int64(5), "Should have more than 5 users to trigger 'already exists' condition")

	err := suite.adminService.InjectSeedData()
	suite.Error(err)
	suite.Contains(err.Error(), "seed data already exists")
}

// TestGetUserManagement_Success 测试获取用户管理数据
func (suite *AdminServiceTestSuite) TestGetUserManagement_Success() {
	response, err := suite.adminService.GetUserManagement(1, 10)

	suite.NoError(err)
	suite.NotNil(response)
	suite.GreaterOrEqual(response.Total, int64(5)) // 至少有测试用户
	suite.Equal(1, response.Page)
	suite.Equal(10, response.Limit)
	suite.NotEmpty(response.Users)
	
	// 验证用户按创建时间倒序排列
	if len(response.Users) > 1 {
		suite.True(response.Users[0].CreatedAt.After(response.Users[1].CreatedAt) || 
			response.Users[0].CreatedAt.Equal(response.Users[1].CreatedAt))
	}
}

// TestGetUserManagement_Pagination 测试分页功能
func (suite *AdminServiceTestSuite) TestGetUserManagement_Pagination() {
	// 测试第一页
	response1, err := suite.adminService.GetUserManagement(1, 2)
	suite.NoError(err)
	suite.LessOrEqual(len(response1.Users), 2)

	// 测试第二页
	response2, err := suite.adminService.GetUserManagement(2, 2)
	suite.NoError(err)
	suite.Equal(response1.Total, response2.Total) // 总数应该相同

	// 验证页面数据不同（如果有足够的数据）
	if response1.Total > 2 {
		if len(response1.Users) > 0 && len(response2.Users) > 0 {
			suite.NotEqual(response1.Users[0].ID, response2.Users[0].ID)
		}
	}
}

// TestGetUserManagement_InvalidParameters 测试无效参数
func (suite *AdminServiceTestSuite) TestGetUserManagement_InvalidParameters() {
	// 测试无效页数和限制
	response, err := suite.adminService.GetUserManagement(0, -1)
	suite.NoError(err)
	suite.Equal(1, response.Page)   // 应该使用默认页数1
	suite.Equal(20, response.Limit) // 应该使用默认限制20

	// 测试过大的限制
	response, err = suite.adminService.GetUserManagement(1, 200)
	suite.NoError(err)
	suite.Equal(20, response.Limit) // 应该限制为20
}

// TestGetSystemSettings_Success 测试获取系统设置
func (suite *AdminServiceTestSuite) TestGetSystemSettings_Success() {
	settings, err := suite.adminService.GetSystemSettings()

	suite.NoError(err)
	suite.NotNil(settings)
	suite.Equal("OpenPenPal", settings.SiteName)
	suite.Equal("手写信的温暖传递平台", settings.SiteDescription)
	suite.True(settings.RegistrationOpen)
	suite.False(settings.MaintenanceMode)
	suite.Equal(10, settings.MaxLettersPerDay)
	suite.Equal(20, settings.MaxEnvelopesPerOrder)
	suite.True(settings.EmailEnabled)
	suite.False(settings.SMSEnabled)
	suite.False(settings.LastUpdated.IsZero())
}

// TestUpdateUser_Success 测试更新用户成功
func (suite *AdminServiceTestSuite) TestUpdateUser_Success() {
	// 创建待更新的用户
	testUser := config.CreateTestUser(suite.db, "updatetest", models.RoleUser)

	req := &models.AdminUpdateUserRequest{
		Nickname:   "Updated Nickname",
		Email:      "updated@example.com",
		Role:       "courier_level1",
		SchoolCode: "NEWSCH",
		IsActive:   false,
	}

	updatedUser, err := suite.adminService.UpdateUser(testUser.ID, req)

	suite.NoError(err)
	suite.NotNil(updatedUser)
	suite.Equal("Updated Nickname", updatedUser.Nickname)
	suite.Equal("updated@example.com", updatedUser.Email)
	suite.Equal(models.RoleCourierLevel1, updatedUser.Role)
	suite.Equal("NEWSCH", updatedUser.SchoolCode)
	suite.False(updatedUser.IsActive)
}

// TestUpdateUser_UserNotFound 测试更新不存在的用户
func (suite *AdminServiceTestSuite) TestUpdateUser_UserNotFound() {
	req := &models.AdminUpdateUserRequest{
		Nickname: "Test",
	}

	updatedUser, err := suite.adminService.UpdateUser("nonexistent", req)

	suite.Error(err)
	suite.Nil(updatedUser)
	suite.Contains(err.Error(), "用户不存在")
}

// TestUpdateUser_DuplicateEmail 测试更新重复邮箱
func (suite *AdminServiceTestSuite) TestUpdateUser_DuplicateEmail() {
	// 创建两个测试用户
	user1 := config.CreateTestUser(suite.db, "duptest1", models.RoleUser)
	user2 := config.CreateTestUser(suite.db, "duptest2", models.RoleUser)

	// 尝试将user2的邮箱改为user1的邮箱
	req := &models.AdminUpdateUserRequest{
		Email: user1.Email,
	}

	updatedUser, err := suite.adminService.UpdateUser(user2.ID, req)

	suite.Error(err)
	suite.Nil(updatedUser)
	suite.Contains(err.Error(), "邮箱已被使用")
}

// TestUpdateUser_InvalidRole 测试无效角色
func (suite *AdminServiceTestSuite) TestUpdateUser_InvalidRole() {
	testUser := config.CreateTestUser(suite.db, "invalidrole", models.RoleUser)

	req := &models.AdminUpdateUserRequest{
		Role: "invalid_role",
	}

	updatedUser, err := suite.adminService.UpdateUser(testUser.ID, req)

	suite.Error(err)
	suite.Nil(updatedUser)
	suite.Contains(err.Error(), "无效的角色")
}

// TestUpdateUser_PartialUpdate 测试部分更新
func (suite *AdminServiceTestSuite) TestUpdateUser_PartialUpdate() {
	testUser := config.CreateTestUser(suite.db, "partialupdate", models.RoleUser)
	originalEmail := testUser.Email
	originalRole := testUser.Role

	// 只更新昵称
	req := &models.AdminUpdateUserRequest{
		Nickname: "Only Nickname Updated",
	}

	updatedUser, err := suite.adminService.UpdateUser(testUser.ID, req)

	suite.NoError(err)
	suite.NotNil(updatedUser)
	suite.Equal("Only Nickname Updated", updatedUser.Nickname)
	suite.Equal(originalEmail, updatedUser.Email) // 邮箱应该保持不变
	suite.Equal(originalRole, updatedUser.Role)   // 角色应该保持不变
}

// TestDashboardStats_WithRealData 测试带真实数据的仪表板统计
func (suite *AdminServiceTestSuite) TestDashboardStats_WithRealData() {
	// 创建一些今天的数据
	today := time.Now().Truncate(24 * time.Hour)
	todayUser := &models.User{
		ID:        "today-user",
		Username:  "todayuser",
		Email:     "today@example.com",
		Nickname:  "Today User",
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: today,
	}
	suite.db.Create(todayUser)

	todayLetter := &models.Letter{
		ID:        "today-letter",
		UserID:    todayUser.ID,
		Title:     "Today Letter",
		Content:   "Letter created today",
		Status:    models.StatusGenerated,
		CreatedAt: today,
	}
	suite.db.Create(todayLetter)

	stats, err := suite.adminService.GetDashboardStats()

	suite.NoError(err)
	suite.GreaterOrEqual(stats.NewUsersToday, int64(1))
	suite.GreaterOrEqual(stats.LettersToday, int64(1))
}

// TestUpdateUser_ValidRoles 测试所有有效角色
func (suite *AdminServiceTestSuite) TestUpdateUser_ValidRoles() {
	validRoles := []string{
		"user", "courier", "courier_level1", "courier_level2", 
		"courier_level3", "courier_level4", "school_admin", "admin", "super_admin",
	}

	for _, role := range validRoles {
		testUser := config.CreateTestUser(suite.db, "role-test-"+role, models.RoleUser)
		
		req := &models.AdminUpdateUserRequest{
			Role: role,
		}

		updatedUser, err := suite.adminService.UpdateUser(testUser.ID, req)

		suite.NoError(err, "Role %s should be valid", role)
		suite.NotNil(updatedUser, "Role %s should be valid", role)
		suite.Equal(role, string(updatedUser.Role), "Role should be updated to %s", role)
	}
}

// 运行测试套件
func TestAdminServiceSuite(t *testing.T) {
	suite.Run(t, new(AdminServiceTestSuite))
}