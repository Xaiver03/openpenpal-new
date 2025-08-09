package services

import (
	"fmt"
	"testing"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LetterServiceTestSuite 信件服务测试套件
type LetterServiceTestSuite struct {
	suite.Suite
	db            *gorm.DB
	letterService *LetterService
	userService   *UserService
	testUser      *models.User
	config        *config.Config
}

func (suite *LetterServiceTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.userService = NewUserService(db, suite.config)
	suite.letterService = NewLetterService(db, suite.config)

	// 创建测试用户
	suite.testUser = config.CreateTestUser(db, "letteruser", models.RoleUser)
}

func (suite *LetterServiceTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM letters")
	suite.db.Exec("DELETE FROM letter_codes")
	suite.db.Exec("DELETE FROM status_logs")
}

// TestCreateDraft_Success 测试成功创建草稿
func (suite *LetterServiceTestSuite) TestCreateDraft_Success() {
	req := &models.CreateLetterRequest{
		Title:   "Test Letter",
		Content: "This is a test letter content.",
		Style:   models.StyleClassic,
	}

	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	
	suite.NoError(err)
	suite.NotNil(letter)
	suite.Equal("Test Letter", letter.Title)
	suite.Equal("This is a test letter content.", letter.Content)
	suite.Equal(models.StatusDraft, letter.Status)
	suite.Equal(suite.testUser.ID, letter.UserID)
	suite.Equal(models.VisibilityPrivate, letter.Visibility) // Default visibility

	// 验证数据库中存在记录
	var savedLetter models.Letter
	err = suite.db.First(&savedLetter, "id = ?", letter.ID).Error
	suite.NoError(err)
	suite.Equal(letter.ID, savedLetter.ID)
}

// TestCreateDraft_EmptyContent 测试空内容（应该被允许）
func (suite *LetterServiceTestSuite) TestCreateDraft_EmptyContent() {
	req := &models.CreateLetterRequest{
		Title:   "Test Letter",
		Content: "", // 空内容在草稿中应该被允许
		Style:   models.StyleClassic,
	}

	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	
	suite.NoError(err) // 应该成功
	suite.NotNil(letter)
	suite.Equal("", letter.Content) // 内容为空
	suite.Equal(models.StatusDraft, letter.Status)
}

// TestCreateDraft_ValidEmptyTitle 测试空标题（允许）
func (suite *LetterServiceTestSuite) TestCreateDraft_ValidEmptyTitle() {
	req := &models.CreateLetterRequest{
		Title:   "", // 空标题应该被允许
		Content: "Content",
		Style:   models.StyleClassic,
	}

	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	
	suite.NoError(err)
	suite.NotNil(letter)
	suite.Equal("", letter.Title)
	suite.Equal("Content", letter.Content)
}

// TestGenerateCode_Success 测试成功生成二维码
func (suite *LetterServiceTestSuite) TestGenerateCode_Success() {
	// 先创建草稿
	req := &models.CreateLetterRequest{
		Title:   "Test Letter",
		Content: "Content",
		Style:   models.StyleClassic,
	}
	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	suite.NoError(err)

	// 生成二维码
	letterCode, err := suite.letterService.GenerateCode(letter.ID)
	
	suite.NoError(err)
	suite.NotNil(letterCode)
	suite.Equal(letter.ID, letterCode.LetterID)
	suite.NotEmpty(letterCode.Code)
	suite.NotEmpty(letterCode.QRCodeURL)

	// 验证信件状态更新为generated
	var updatedLetter models.Letter
	err = suite.db.First(&updatedLetter, "id = ?", letter.ID).Error
	suite.NoError(err)
	suite.Equal(models.StatusGenerated, updatedLetter.Status)
}

// TestGenerateCode_AlreadyGenerated 测试重复生成二维码
func (suite *LetterServiceTestSuite) TestGenerateCode_AlreadyGenerated() {
	// 创建草稿并生成二维码
	req := &models.CreateLetterRequest{
		Title:   "Test Letter",
		Content: "Content",
		Style:   models.StyleClassic,
	}
	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	suite.NoError(err)
	
	_, err = suite.letterService.GenerateCode(letter.ID)
	suite.NoError(err)

	// 尝试重复生成（应该返回现有的code）
	letterCode, err := suite.letterService.GenerateCode(letter.ID)
	
	suite.NoError(err) // 应该成功返回现有code
	suite.NotNil(letterCode)
	suite.NotEmpty(letterCode.Code)
}

// TestUpdateStatus_Success 测试成功更新状态
func (suite *LetterServiceTestSuite) TestUpdateStatus_Success() {
	// 创建生成状态的信件
	letter := suite.createGeneratedLetter()

	// 获取letter code
	var letterCode models.LetterCode
	err := suite.db.First(&letterCode, "letter_id = ?", letter.ID).Error
	suite.NoError(err)

	// 更新状态为collected
	req := &models.UpdateLetterStatusRequest{
		Status:   models.StatusCollected,
		Location: "Test Location",
		Note:     "Collected for delivery",
	}
	err = suite.letterService.UpdateStatus(letterCode.Code, req, "test-courier")
	
	suite.NoError(err)

	// 验证状态更新
	var updatedLetter models.Letter
	err = suite.db.First(&updatedLetter, "id = ?", letter.ID).Error
	suite.NoError(err)
	suite.Equal(models.StatusCollected, updatedLetter.Status)

	// 验证状态日志
	var statusLog models.StatusLog
	err = suite.db.First(&statusLog, "letter_id = ?", letter.ID).Error
	suite.NoError(err)
	suite.Equal(models.StatusCollected, statusLog.Status)
	suite.Equal("test-courier", statusLog.UpdatedBy)
	suite.Equal("Collected for delivery", statusLog.Note)
}

// TestUpdateStatus_LetterNotFound 测试信件不存在
func (suite *LetterServiceTestSuite) TestUpdateStatus_LetterNotFound() {
	// 使用不存在的code
	req := &models.UpdateLetterStatusRequest{
		Status:   models.StatusCollected,
		Location: "Test Location",
		Note:     "Test Note",
	}
	err := suite.letterService.UpdateStatus("NONEXISTENT", req, "test-courier")
	
	suite.Error(err)
	suite.Contains(err.Error(), "letter not found")
}

// TestGetUserLetters_Success 测试获取用户信件列表
func (suite *LetterServiceTestSuite) TestGetUserLetters_Success() {
	// 创建多封信件
	for i := 0; i < 5; i++ {
		req := &models.CreateLetterRequest{
			Title:   fmt.Sprintf("Test Letter %d", i+1),
			Content: fmt.Sprintf("Content %d", i+1),
			Style:   models.StyleClassic,
		}
		_, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
		suite.NoError(err)
	}

	// 获取信件列表
	params := &models.LetterListParams{
		Page:      1,
		Limit:     10,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	letters, total, err := suite.letterService.GetUserLetters(suite.testUser.ID, params)
	
	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(letters, 5)

	// 验证按时间倒序排列
	for i := 0; i < len(letters)-1; i++ {
		suite.True(letters[i].CreatedAt.After(letters[i+1].CreatedAt) || letters[i].CreatedAt.Equal(letters[i+1].CreatedAt))
	}
}

// TestGetUserLetters_Pagination 测试分页功能
func (suite *LetterServiceTestSuite) TestGetUserLetters_Pagination() {
	// 创建10封信件
	for i := 0; i < 10; i++ {
		req := &models.CreateLetterRequest{
			Title:   fmt.Sprintf("Test Letter %d", i+1),
			Content: fmt.Sprintf("Content %d", i+1),
			Style:   models.StyleClassic,
		}
		_, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
		suite.NoError(err)
	}

	// 获取第一页（每页5条）
	params1 := &models.LetterListParams{
		Page:      1,
		Limit:     5,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	page1Letters, total, err := suite.letterService.GetUserLetters(suite.testUser.ID, params1)
	
	suite.NoError(err)
	suite.Equal(int64(10), total)
	suite.Len(page1Letters, 5)

	// 获取第二页
	params2 := &models.LetterListParams{
		Page:      2,
		Limit:     5,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	page2Letters, total, err := suite.letterService.GetUserLetters(suite.testUser.ID, params2)
	
	suite.NoError(err)
	suite.Equal(int64(10), total)
	suite.Len(page2Letters, 5)

	// 验证页面不重复
	for _, letter1 := range page1Letters {
		for _, letter2 := range page2Letters {
			suite.NotEqual(letter1.ID, letter2.ID)
		}
	}
}

// TestMarkAsRead_Success 测试标记为已读
func (suite *LetterServiceTestSuite) TestMarkAsRead_Success() {
	// 创建delivered状态的信件
	letter := suite.createDeliveredLetter()

	// 获取letter code
	var letterCode models.LetterCode
	err := suite.db.First(&letterCode, "letter_id = ?", letter.ID).Error
	suite.NoError(err)

	// 标记为已读
	err = suite.letterService.MarkAsRead(letterCode.Code, suite.testUser.ID)
	
	suite.NoError(err)

	// 验证状态更新
	var updatedLetter models.Letter
	err = suite.db.First(&updatedLetter, "id = ?", letter.ID).Error
	suite.NoError(err)
	suite.Equal(models.StatusRead, updatedLetter.Status)
}

// TestGetLetterByCode_Success 测试通过码获取信件
func (suite *LetterServiceTestSuite) TestGetLetterByCode_Success() {
	// 创建生成状态的信件
	letter := suite.createGeneratedLetter()

	// 获取信件码
	var letterCode models.LetterCode
	err := suite.db.First(&letterCode, "letter_id = ?", letter.ID).Error
	suite.NoError(err)

	// 通过码获取信件
	retrievedLetter, err := suite.letterService.GetLetterByCode(letterCode.Code)
	
	suite.NoError(err)
	suite.NotNil(retrievedLetter)
	suite.Equal(letter.ID, retrievedLetter.Letter.ID)
	suite.Equal(letter.Title, retrievedLetter.Letter.Title)
	suite.Equal(letter.Content, retrievedLetter.Letter.Content)
}

// TestGetLetterByCode_NotFound 测试无效码
func (suite *LetterServiceTestSuite) TestGetLetterByCode_NotFound() {
	// 使用不存在的码
	letter, err := suite.letterService.GetLetterByCode("NONEXISTENT")
	
	suite.Error(err)
	suite.Nil(letter)
	suite.Contains(err.Error(), "letter not found")
}

// Helper methods

func (suite *LetterServiceTestSuite) createGeneratedLetter() *models.Letter {
	req := &models.CreateLetterRequest{
		Title:   "Generated Letter",
		Content: "Generated content",
		Style:   models.StyleClassic,
	}
	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	suite.NoError(err)

	_, err = suite.letterService.GenerateCode(letter.ID)
	suite.NoError(err)

	// 重新获取更新后的信件
	err = suite.db.First(letter, "id = ?", letter.ID).Error
	suite.NoError(err)

	return letter
}

func (suite *LetterServiceTestSuite) createDeliveredLetter() *models.Letter {
	letter := suite.createGeneratedLetter()

	// 获取letter code
	var letterCode models.LetterCode
	err := suite.db.First(&letterCode, "letter_id = ?", letter.ID).Error
	suite.NoError(err)

	// 更新状态到delivered
	req1 := &models.UpdateLetterStatusRequest{
		Status:   models.StatusCollected,
		Location: "Test Location",
		Note:     "Collected",
	}
	err = suite.letterService.UpdateStatus(letterCode.Code, req1, "test-courier")
	suite.NoError(err)
	
	req2 := &models.UpdateLetterStatusRequest{
		Status:   models.StatusInTransit,
		Location: "Test Location",
		Note:     "In transit",
	}
	err = suite.letterService.UpdateStatus(letterCode.Code, req2, "test-courier")
	suite.NoError(err)
	
	req3 := &models.UpdateLetterStatusRequest{
		Status:   models.StatusDelivered,
		Location: "Test Location",
		Note:     "Delivered",
	}
	err = suite.letterService.UpdateStatus(letterCode.Code, req3, "test-courier")
	suite.NoError(err)

	// 重新获取更新后的信件
	err = suite.db.First(letter, "id = ?", letter.ID).Error
	suite.NoError(err)

	return letter
}

// 运行测试套件
func TestLetterServiceSuite(t *testing.T) {
	suite.Run(t, new(LetterServiceTestSuite))
}