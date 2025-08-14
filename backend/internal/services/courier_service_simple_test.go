package services

import (
	"testing"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// CourierServiceSimpleTestSuite 简化的信使服务测试套件
type CourierServiceSimpleTestSuite struct {
	suite.Suite
	db             *gorm.DB
	courierService *CourierService
	userService    *UserService
	testUser       *models.User
	config         *config.Config
}

func (suite *CourierServiceSimpleTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.userService = NewUserService(db, suite.config)
	suite.courierService = NewCourierService(db)

	// 创建测试用户
	suite.testUser = config.CreateTestUser(db, "courieruser", models.RoleUser)
}

func (suite *CourierServiceSimpleTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM couriers")
	suite.db.Exec("DELETE FROM courier_tasks")
}

// TestApplyCourier_Success 测试成功申请信使
func (suite *CourierServiceSimpleTestSuite) TestApplyCourier_Success() {
	req := &models.CourierApplication{
		Name:            "测试信使",
		Contact:         "test@example.com",
		School:          "北京大学",
		Zone:            "BJDX",
		HasPrinter:      "yes",
		SelfIntro:       "我想帮助同学传递信件",
		CanMentor:       "maybe",
		WeeklyHours:     10,
		MaxDailyTasks:   5,
		TransportMethod: "walk",
		TimeSlots:       []string{"morning", "afternoon"},
	}

	courier, err := suite.courierService.ApplyCourier(suite.testUser.ID, req)

	suite.NoError(err)
	suite.NotNil(courier)
	suite.Equal(suite.testUser.ID, courier.UserID)
	suite.Equal("test@example.com", courier.Contact)
	suite.Equal("北京大学", courier.School)
	suite.Equal("approved", courier.Status) // Auto-approved based on criteria
	suite.Equal(1, courier.Level)           // Default level 1
}

// TestApplyCourier_DuplicateApplication 测试重复申请
func (suite *CourierServiceSimpleTestSuite) TestApplyCourier_DuplicateApplication() {
	// 先申请一次
	req1 := &models.CourierApplication{
		Name:            "测试信使1",
		Contact:         "test1@example.com",
		School:          "北京大学",
		Zone:            "BJDX",
		HasPrinter:      "yes",
		SelfIntro:       "我想帮助同学传递信件",
		CanMentor:       "maybe",
		WeeklyHours:     10,
		MaxDailyTasks:   5,
		TransportMethod: "walk",
		TimeSlots:       []string{"morning"},
	}
	_, err := suite.courierService.ApplyCourier(suite.testUser.ID, req1)
	suite.NoError(err)

	// 尝试重复申请
	req2 := &models.CourierApplication{
		Name:            "测试信使2",
		Contact:         "test2@example.com",
		School:          "清华大学",
		Zone:            "QH",
		HasPrinter:      "no",
		SelfIntro:       "再次申请",
		CanMentor:       "no",
		WeeklyHours:     8,
		MaxDailyTasks:   3,
		TransportMethod: "bike",
		TimeSlots:       []string{"evening"},
	}
	courier, err := suite.courierService.ApplyCourier(suite.testUser.ID, req2)

	suite.Error(err)
	suite.Nil(courier)
	suite.Contains(err.Error(), "已经申请过信使")
}

// TestGetCourierStatus_NotApplied 测试未申请用户
func (suite *CourierServiceSimpleTestSuite) TestGetCourierStatus_NotApplied() {
	// 获取未申请用户的状态
	status, err := suite.courierService.GetCourierStatus(suite.testUser.ID)

	suite.NoError(err)
	suite.NotNil(status)
	suite.False(status.IsApplied)
	suite.Equal("", status.Status)
	suite.Equal(0, status.Level)
}

// TestGetCourierByUserID_Success 测试通过用户ID获取信使
func (suite *CourierServiceSimpleTestSuite) TestGetCourierByUserID_Success() {
	// 先申请成为信使
	req := &models.CourierApplication{
		Name:            "测试信使",
		Contact:         "test@example.com",
		School:          "北京大学",
		Zone:            "BJDX",
		HasPrinter:      "yes",
		SelfIntro:       "我想帮助同学传递信件",
		CanMentor:       "maybe",
		WeeklyHours:     10,
		MaxDailyTasks:   5,
		TransportMethod: "walk",
		TimeSlots:       []string{"morning"},
	}
	appliedCourier, err := suite.courierService.ApplyCourier(suite.testUser.ID, req)
	suite.NoError(err)

	// 通过用户ID获取信使
	courier, err := suite.courierService.GetCourierByUserID(suite.testUser.ID)

	suite.NoError(err)
	suite.NotNil(courier)
	suite.Equal(appliedCourier.ID, courier.ID)
	suite.Equal(suite.testUser.ID, courier.UserID)
}

// TestGetCourierByUserID_NotFound 测试不存在的用户
func (suite *CourierServiceSimpleTestSuite) TestGetCourierByUserID_NotFound() {
	// 使用不存在的用户ID
	courier, err := suite.courierService.GetCourierByUserID("nonexistent-user")

	suite.Error(err)
	suite.Nil(courier)
}

// TestGetPendingApplications_Success 测试获取待审批申请
func (suite *CourierServiceSimpleTestSuite) TestGetPendingApplications_Success() {
	// 创建多个待审批申请（使用不会自动审批的条件）
	for i := 0; i < 3; i++ {
		user := config.CreateTestUser(suite.db, "pending_user_"+string(rune('A'+i)), models.RoleUser)
		req := &models.CourierApplication{
			Name:            "申请信使" + string(rune('A'+i)),
			Contact:         "test" + string(rune('A'+i)) + "@example.com",
			School:          "北京大学",
			Zone:            "BJDX*", // 申请整层楼，不会自动审批
			HasPrinter:      "yes",
			SelfIntro:       "申请成为信使",
			CanMentor:       "maybe",
			WeeklyHours:     10,
			MaxDailyTasks:   5,
			TransportMethod: "walk",
			TimeSlots:       []string{"morning"},
		}
		_, err := suite.courierService.ApplyCourier(user.ID, req)
		suite.NoError(err)
	}

	// 获取待审批申请
	applications, err := suite.courierService.GetPendingApplications()

	suite.NoError(err)
	suite.Len(applications, 3)
	for _, app := range applications {
		suite.Equal("pending", app.Status)
	}
}

// TestGetCouriersByZone_Success 测试按区域获取信使
func (suite *CourierServiceSimpleTestSuite) TestGetCouriersByZone_Success() {
	// 创建不同区域的信使并直接插入数据库（跳过审批流程）
	zones := []string{"BJDX", "QH", "BJDX"}
	for i, zone := range zones {
		user := config.CreateTestUser(suite.db, "zone_user_"+string(rune('A'+i)), models.RoleUser)

		// 直接创建已审批的信使记录
		courier := &models.Courier{
			ID:      "courier-" + string(rune('A'+i)),
			UserID:  user.ID,
			Name:    "区域信使" + string(rune('A'+i)),
			Contact: "zone" + string(rune('A'+i)) + "@example.com",
			School:  zone,
			Zone:    zone,
			Status:  "approved",
			Level:   1,
		}
		err := suite.db.Create(courier).Error
		suite.NoError(err)
	}

	// 获取BJDX区域的信使
	couriers, err := suite.courierService.GetCouriersByZone("BJDX")

	suite.NoError(err)
	suite.Len(couriers, 2) // 应该有2个BJDX区域的信使
	for _, courier := range couriers {
		suite.Equal("BJDX", courier.Zone)
		suite.Equal("approved", courier.Status)
	}
}

// TestValidateOPCodeAccess_Basic 测试基本OP码访问权限验证
func (suite *CourierServiceSimpleTestSuite) TestValidateOPCodeAccess_Basic() {
	// 创建信使用户
	courierUser := config.CreateTestUser(suite.db, "courier_opcode", models.RoleCourierLevel2)

	// 直接创建已审批的信使记录
	courier := &models.Courier{
		ID:                  "courier-opcode",
		UserID:              courierUser.ID,
		Name:                "OP码测试信使",
		Contact:             "opcode@example.com",
		School:              "北京大学",
		Zone:                "BJDX",
		ManagedOPCodePrefix: "BJDX",
		Status:              "approved",
		Level:               2,
	}
	err := suite.db.Create(courier).Error
	suite.NoError(err)

	// 测试访问权限
	valid, err := suite.courierService.ValidateOPCodeAccess(courierUser.ID, "BJDX5F")

	suite.NoError(err)
	suite.True(valid) // 2级信使应该能访问BJDX开头的OP码
}

// Helper methods

func (suite *CourierServiceSimpleTestSuite) createBasicCourier(user *models.User) *models.Courier {
	req := &models.CourierApplication{
		Name:            "基础信使",
		Contact:         user.Username + "@example.com",
		School:          "测试学校",
		Zone:            "TEST",
		HasPrinter:      "yes",
		SelfIntro:       "测试申请",
		CanMentor:       "maybe",
		WeeklyHours:     10,
		MaxDailyTasks:   5,
		TransportMethod: "walk",
		TimeSlots:       []string{"morning"},
	}
	courier, err := suite.courierService.ApplyCourier(user.ID, req)
	suite.NoError(err)
	return courier
}

// 运行测试套件
func TestCourierServiceSimpleSuite(t *testing.T) {
	suite.Run(t, new(CourierServiceSimpleTestSuite))
}
