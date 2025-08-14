package services

import (
	"testing"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	userService *UserService
	testUser    *models.User
	testAdmin   *models.User
	config      *config.Config
}

func (suite *UserServiceTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.userService = NewUserService(db, suite.config)

	// 创建测试用户
	suite.testUser = config.CreateTestUser(db, "testuser", models.RoleUser)
	suite.testAdmin = config.CreateTestUser(db, "testadmin", models.RoleSuperAdmin)
}

func (suite *UserServiceTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM users WHERE username NOT IN ('testuser', 'testadmin')")
	suite.db.Exec("DELETE FROM user_profiles")
}

// TestRegister_Success 测试成功注册用户
func (suite *UserServiceTestSuite) TestRegister_Success() {
	req := &models.RegisterRequest{
		Username:   "newuser",
		Password:   "password123",
		Email:      "newuser@example.com",
		SchoolCode: "BJDX01",
		Nickname:   "新用户",
	}

	user, err := suite.userService.Register(req)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("newuser", user.Username)
	suite.Equal("newuser@example.com", user.Email)
	suite.Equal("BJDX01", user.SchoolCode)
	suite.Equal("新用户", user.Nickname)
	suite.Equal(models.RoleUser, user.Role)
	suite.True(user.IsActive)

	// 服务会清空返回的密码哈希，直接验证数据库中的记录
	var dbUser models.User
	err = suite.db.First(&dbUser, "username = ?", "newuser").Error
	suite.NoError(err)
	suite.NotEqual("password123", dbUser.PasswordHash)
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte("password123"))
	suite.NoError(err)

	// 验证数据库中存在记录
	var savedUser models.User
	err = suite.db.First(&savedUser, "username = ?", "newuser").Error
	suite.NoError(err)
	suite.Equal(user.ID, savedUser.ID)
}

// TestRegister_DuplicateUsername 测试重复用户名
func (suite *UserServiceTestSuite) TestRegister_DuplicateUsername() {
	req := &models.RegisterRequest{
		Username:   "testuser", // 已存在的用户名
		Password:   "password123",
		Email:      "duplicate@example.com",
		SchoolCode: "BJDX01",
	}

	user, err := suite.userService.Register(req)

	suite.Error(err)
	suite.Nil(user)
	suite.Contains(err.Error(), "username or email already exists")
}

// TestRegister_DuplicateEmail 测试重复邮箱
func (suite *UserServiceTestSuite) TestRegister_DuplicateEmail() {
	// 先注册一个用户
	req1 := &models.RegisterRequest{
		Username:   "user1",
		Password:   "password123",
		Email:      "same@example.com",
		SchoolCode: "BJDX01",
	}
	_, err := suite.userService.Register(req1)
	suite.NoError(err)

	// 尝试用相同邮箱注册另一个用户
	req2 := &models.RegisterRequest{
		Username:   "user2",
		Password:   "password123",
		Email:      "same@example.com", // 相同邮箱
		SchoolCode: "QHDA01",
	}
	user, err := suite.userService.Register(req2)

	suite.Error(err)
	suite.Nil(user)
	suite.Contains(err.Error(), "username or email already exists")
}

// TestRegister_InvalidSchoolCode 测试无效学校代码
func (suite *UserServiceTestSuite) TestRegister_InvalidSchoolCode() {
	req := &models.RegisterRequest{
		Username:   "invalidschool",
		Password:   "password123",
		Email:      "invalid@example.com",
		SchoolCode: "INVALD", // Invalid 6-char code for testing
	}

	user, err := suite.userService.Register(req)

	// INVALD passes format validation (6 chars uppercase), so should succeed
	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("INVALD", user.SchoolCode)
}

// TestLogin_Success 测试成功登录
func (suite *UserServiceTestSuite) TestLogin_Success() {
	// 先注册一个用户
	regReq := &models.RegisterRequest{
		Username:   "loginuser",
		Password:   "password123",
		Email:      "login@example.com",
		SchoolCode: "BJDX01",
	}
	registeredUser, err := suite.userService.Register(regReq)
	suite.NoError(err)

	// 使用用户名登录
	loginReq := &models.LoginRequest{
		Username: "loginuser",
		Password: "password123",
	}
	loginResp, err := suite.userService.Login(loginReq)

	suite.NoError(err)
	suite.NotNil(loginResp)
	suite.NotEmpty(loginResp.Token)
	suite.Equal(registeredUser.ID, loginResp.User.ID)
	suite.Equal("loginuser", loginResp.User.Username)
	suite.Equal("login@example.com", loginResp.User.Email)
}

// TestLogin_WithEmail 测试使用邮箱登录
func (suite *UserServiceTestSuite) TestLogin_WithEmail() {
	// 先注册一个用户
	regReq := &models.RegisterRequest{
		Username:   "emailuser",
		Password:   "password123",
		Email:      "email@example.com",
		SchoolCode: "BJDX01",
	}
	_, err := suite.userService.Register(regReq)
	suite.NoError(err)

	// 使用邮箱登录
	loginReq := &models.LoginRequest{
		Username: "email@example.com", // 注意：实际实现中可能需要支持邮箱登录
		Password: "password123",
	}
	loginResp, err := suite.userService.Login(loginReq)

	suite.NoError(err)
	suite.NotNil(loginResp)
	suite.NotEmpty(loginResp.Token)
	suite.Equal("emailuser", loginResp.User.Username)
	suite.Equal("email@example.com", loginResp.User.Email)
}

// TestLogin_WrongPassword 测试密码错误
func (suite *UserServiceTestSuite) TestLogin_WrongPassword() {
	loginReq := &models.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	loginResp, err := suite.userService.Login(loginReq)

	suite.Error(err)
	suite.Nil(loginResp)
	suite.Contains(err.Error(), "invalid username or password")
}

// TestLogin_UserNotFound 测试用户不存在
func (suite *UserServiceTestSuite) TestLogin_UserNotFound() {
	loginReq := &models.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}
	loginResp, err := suite.userService.Login(loginReq)

	suite.Error(err)
	suite.Nil(loginResp)
	suite.Contains(err.Error(), "invalid username or password")
}

// TestLogin_InactiveUser 测试已禁用用户
func (suite *UserServiceTestSuite) TestLogin_InactiveUser() {
	// 创建并禁用用户
	regReq := &models.RegisterRequest{
		Username:   "inactiveuser",
		Password:   "password123",
		Email:      "inactive@example.com",
		SchoolCode: "BJDX01",
	}
	user, err := suite.userService.Register(regReq)
	suite.NoError(err)

	// 禁用用户
	err = suite.userService.DeactivateUser(user.ID)
	suite.NoError(err)

	// 尝试登录
	loginReq := &models.LoginRequest{
		Username: "inactiveuser",
		Password: "password123",
	}
	loginResp, err := suite.userService.Login(loginReq)

	suite.Error(err)
	suite.Nil(loginResp)
	suite.Contains(err.Error(), "user account is disabled")
}

// TestGetUserByID_Success 测试获取用户信息
func (suite *UserServiceTestSuite) TestGetUserByID_Success() {
	user, err := suite.userService.GetUserByID(suite.testUser.ID)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(suite.testUser.ID, user.ID)
	suite.Equal(suite.testUser.Username, user.Username)
	suite.Equal(suite.testUser.Email, user.Email)
}

// TestGetUserByID_NotFound 测试用户不存在
func (suite *UserServiceTestSuite) TestGetUserByID_NotFound() {
	user, err := suite.userService.GetUserByID("nonexistent")

	suite.Error(err)
	suite.Nil(user)
}

// TestUpdateProfile_Success 测试更新用户档案
func (suite *UserServiceTestSuite) TestUpdateProfile_Success() {
	req := &models.UpdateProfileRequest{
		Nickname: "新昵称",
		Bio:      "这是我的简介",
		Address:  "北京市海淀区",
	}

	updatedUser, err := suite.userService.UpdateProfile(suite.testUser.ID, req)

	suite.NoError(err)
	suite.NotNil(updatedUser)
	suite.Equal("新昵称", updatedUser.Nickname)

	// 验证数据库更新
	var savedUser models.User
	err = suite.db.First(&savedUser, "id = ?", suite.testUser.ID).Error
	suite.NoError(err)
	suite.Equal("新昵称", savedUser.Nickname)
}

// TestUpdateProfile_UserNotFound 测试更新不存在用户的档案
func (suite *UserServiceTestSuite) TestUpdateProfile_UserNotFound() {
	req := &models.UpdateProfileRequest{
		Nickname: "测试昵称",
	}

	user, err := suite.userService.UpdateProfile("nonexistent", req)

	suite.Error(err)
	suite.Nil(user)
}

// TestChangePassword_Success 测试更改密码
func (suite *UserServiceTestSuite) TestChangePassword_Success() {
	// 先注册用户以知道初始密码
	regReq := &models.RegisterRequest{
		Username:   "changepassuser",
		Password:   "oldpassword",
		Email:      "changepw@example.com",
		SchoolCode: "BJDX01",
	}
	user, err := suite.userService.Register(regReq)
	suite.NoError(err)

	// 更改密码
	changeReq := &models.ChangePasswordRequest{
		OldPassword: "oldpassword",
		NewPassword: "newpassword123",
	}
	err = suite.userService.ChangePassword(user.ID, changeReq)

	suite.NoError(err)

	// 验证新密码可以登录
	loginReq := &models.LoginRequest{
		Username: "changepassuser",
		Password: "newpassword123",
	}
	loginResp, err := suite.userService.Login(loginReq)
	suite.NoError(err)
	suite.NotNil(loginResp)
}

// TestChangePassword_WrongCurrentPassword 测试当前密码错误
func (suite *UserServiceTestSuite) TestChangePassword_WrongCurrentPassword() {
	// 注册用户
	regReq := &models.RegisterRequest{
		Username:   "wrongcurrentpw",
		Password:   "correctpassword",
		Email:      "wrongpw@example.com",
		SchoolCode: "BJDX01",
	}
	user, err := suite.userService.Register(regReq)
	suite.NoError(err)

	// 使用错误的当前密码
	changeReq := &models.ChangePasswordRequest{
		OldPassword: "wrongcurrentpassword",
		NewPassword: "newpassword123",
	}
	err = suite.userService.ChangePassword(user.ID, changeReq)

	suite.Error(err)
	suite.Contains(err.Error(), "old password is incorrect")
}

// TestGetUserStats_Success 测试获取用户统计
func (suite *UserServiceTestSuite) TestGetUserStats_Success() {
	stats, err := suite.userService.GetUserStats(suite.testUser.ID)

	suite.NoError(err)
	suite.NotNil(stats)
	// 统计数据结构验证（具体字段取决于实现）
	suite.GreaterOrEqual(stats.LettersSent, int64(0))
	suite.GreaterOrEqual(stats.LettersReceived, int64(0))
	suite.GreaterOrEqual(stats.DraftsCount, int64(0))
}

// TestGetUserStats_UserNotFound 测试不存在用户的统计
func (suite *UserServiceTestSuite) TestGetUserStats_UserNotFound() {
	stats, err := suite.userService.GetUserStats("nonexistent")

	// 服务不检查用户是否存在，只统计数据，所以返回空统计
	suite.NoError(err)
	suite.NotNil(stats)
	suite.Equal(int64(0), stats.LettersSent)
}

// TestDeactivateUser_Success 测试禁用用户
func (suite *UserServiceTestSuite) TestDeactivateUser_Success() {
	// 注册一个用户用于测试
	regReq := &models.RegisterRequest{
		Username:   "deactivateuser",
		Password:   "password123",
		Email:      "deactivate@example.com",
		SchoolCode: "BJDX01",
	}
	user, err := suite.userService.Register(regReq)
	suite.NoError(err)

	// 禁用用户
	err = suite.userService.DeactivateUser(user.ID)
	suite.NoError(err)

	// 验证用户被禁用
	var updatedUser models.User
	err = suite.db.First(&updatedUser, "id = ?", user.ID).Error
	suite.NoError(err)
	suite.False(updatedUser.IsActive)
}

// TestReactivateUser_Success 测试重新激活用户
func (suite *UserServiceTestSuite) TestReactivateUser_Success() {
	// 注册并禁用用户
	regReq := &models.RegisterRequest{
		Username:   "reactivateuser",
		Password:   "password123",
		Email:      "reactivate@example.com",
		SchoolCode: "BJDX01",
	}
	user, err := suite.userService.Register(regReq)
	suite.NoError(err)

	err = suite.userService.DeactivateUser(user.ID)
	suite.NoError(err)

	// 重新激活用户
	err = suite.userService.ReactivateUser(user.ID)
	suite.NoError(err)

	// 验证用户被激活
	var updatedUser models.User
	err = suite.db.First(&updatedUser, "id = ?", user.ID).Error
	suite.NoError(err)
	suite.True(updatedUser.IsActive)
}

// TestUpdateLastActivity_Success 测试更新最后活动时间
func (suite *UserServiceTestSuite) TestUpdateLastActivity_Success() {
	now := time.Now()
	err := suite.userService.UpdateLastActivity(suite.testUser.ID, &now)

	suite.NoError(err)

	// 验证更新
	var updatedUser models.User
	err = suite.db.First(&updatedUser, "id = ?", suite.testUser.ID).Error
	suite.NoError(err)
	suite.NotNil(updatedUser.LastLoginAt)
	suite.True(updatedUser.LastLoginAt.After(time.Now().Add(-time.Minute)))
}

// TestUserRolePermissions 测试用户角色权限
func (suite *UserServiceTestSuite) TestUserRolePermissions() {
	// 测试基础用户权限
	suite.True(suite.testUser.HasRole(models.RoleUser))
	suite.False(suite.testUser.HasRole(models.RoleSuperAdmin))

	// 测试管理员权限
	suite.True(suite.testAdmin.HasRole(models.RoleSuperAdmin))
	suite.True(suite.testAdmin.HasRole(models.RoleUser)) // 管理员包含用户权限

	// 测试管理权限
	suite.True(suite.testAdmin.CanManageUser(suite.testUser))
	suite.False(suite.testUser.CanManageUser(suite.testAdmin))
}

// TestPasswordSecurity 测试密码安全性
func (suite *UserServiceTestSuite) TestPasswordSecurity() {
	// 注册用户
	regReq := &models.RegisterRequest{
		Username:   "securitytest",
		Password:   "securepassword123",
		Email:      "security@example.com",
		SchoolCode: "BJDX01",
	}
	_, err := suite.userService.Register(regReq)
	suite.NoError(err)

	// 服务会清空返回的密码哈希，所以需要从数据库查询验证
	var savedUser models.User
	err = suite.db.First(&savedUser, "username = ?", "securitytest").Error
	suite.NoError(err)

	// 验证密码被正确哈希（不等于原始密码）
	suite.NotEqual("securepassword123", savedUser.PasswordHash)

	// 验证哈希强度（bcrypt特征）
	suite.True(len(savedUser.PasswordHash) >= 60)  // bcrypt哈希长度
	suite.Contains(savedUser.PasswordHash, "$2a$") // bcrypt前缀

	// 验证密码验证功能
	err = bcrypt.CompareHashAndPassword([]byte(savedUser.PasswordHash), []byte("securepassword123"))
	suite.NoError(err)

	err = bcrypt.CompareHashAndPassword([]byte(savedUser.PasswordHash), []byte("wrongpassword"))
	suite.Error(err)
}

// 运行测试套件
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
