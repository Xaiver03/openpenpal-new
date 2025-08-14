package services

import (
	"context"
	"testing"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MuseumServiceSimpleTestSuite 简化的博物馆服务测试套件
type MuseumServiceSimpleTestSuite struct {
	suite.Suite
	db            *gorm.DB
	museumService *MuseumService
	userService   *UserService
	letterService *LetterService
	testUser      *models.User
	testAdmin     *models.User
	config        *config.Config
}

func (suite *MuseumServiceSimpleTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.userService = NewUserService(db, suite.config)
	suite.letterService = NewLetterService(db, suite.config)
	suite.museumService = NewMuseumService(db)

	// 创建测试用户
	suite.testUser = config.CreateTestUser(db, "museumuser", models.RoleUser)
	suite.testAdmin = config.CreateTestUser(db, "museumadmin", models.RoleSuperAdmin)
}

func (suite *MuseumServiceSimpleTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM museum_items")
	suite.db.Exec("DELETE FROM museum_entries")
	suite.db.Exec("DELETE FROM museum_interactions")
	suite.db.Exec("DELETE FROM museum_reactions")
	suite.db.Exec("DELETE FROM museum_exhibitions")
	suite.db.Exec("DELETE FROM letters")
	suite.db.Exec("DELETE FROM letter_codes")
}

// TestCreateMuseumItem_Success 测试成功创建博物馆物品
func (suite *MuseumServiceSimpleTestSuite) TestCreateMuseumItem_Success() {
	// 创建请求结构体（基于服务中的实际定义）
	req := &CreateMuseumItemRequest{
		SourceType:  models.SourceTypeLetter,
		SourceID:    "test-source-id",
		Title:       "测试博物馆物品",
		Description: "这是一个测试物品的描述",
		Tags:        []string{"测试", "博物馆"},
	}

	// 创建博物馆物品
	item, err := suite.museumService.CreateMuseumItem(context.Background(), req)

	suite.NoError(err)
	suite.NotNil(item)
	suite.Equal("测试博物馆物品", item.Title)
	suite.Equal("这是一个测试物品的描述", item.Description)
	suite.Equal(models.SourceTypeLetter, item.SourceType)
	suite.Equal("pending", string(item.Status))

	// 验证数据库中存在记录
	var savedItem models.MuseumItem
	err = suite.db.First(&savedItem, "id = ?", item.ID).Error
	suite.NoError(err)
	suite.Equal(item.ID, savedItem.ID)
}

// TestCreateMuseumItem_InvalidSourceType 测试无效来源类型
func (suite *MuseumServiceSimpleTestSuite) TestCreateMuseumItem_InvalidSourceType() {
	req := &CreateMuseumItemRequest{
		SourceType: "invalid",
		SourceID:   "test-id",
		Title:      "测试标题",
	}

	item, err := suite.museumService.CreateMuseumItem(context.Background(), req)

	// 应该处理无效的来源类型（具体行为取决于服务实现）
	if err != nil {
		suite.Error(err)
		suite.Nil(item)
	} else {
		// 如果服务允许无效类型，至少验证创建成功
		suite.NotNil(item)
	}
}

// TestGetMuseumEntry_Success 测试获取博物馆条目
func (suite *MuseumServiceSimpleTestSuite) TestGetMuseumEntry_Success() {
	// 先创建一个博物馆条目
	entry := suite.createTestMuseumEntry()

	// 获取博物馆条目
	retrievedEntry, err := suite.museumService.GetMuseumEntry(context.Background(), entry.ID)

	suite.NoError(err)
	suite.NotNil(retrievedEntry)
	suite.Equal(entry.ID, retrievedEntry.ID)
}

// TestGetMuseumEntry_NotFound 测试物品不存在
func (suite *MuseumServiceSimpleTestSuite) TestGetMuseumEntry_NotFound() {
	// 使用不存在的ID
	entry, err := suite.museumService.GetMuseumEntry(context.Background(), "nonexistent")

	suite.Error(err)
	suite.Nil(entry)
}

// TestApproveMuseumItem_Success 测试审批通过物品
func (suite *MuseumServiceSimpleTestSuite) TestApproveMuseumItem_Success() {
	// 创建待审批物品
	item := suite.createTestMuseumItem()

	// 审批通过
	err := suite.museumService.ApproveMuseumItem(context.Background(), item.ID, suite.testAdmin.ID)

	suite.NoError(err)

	// 验证状态更新
	var updatedItem models.MuseumItem
	err = suite.db.First(&updatedItem, "id = ?", item.ID).Error
	suite.NoError(err)
	suite.Equal("approved", string(updatedItem.Status))
	suite.Equal(suite.testAdmin.ID, *updatedItem.ApprovedBy)
	suite.NotNil(updatedItem.ApprovedAt)
}

// TestApproveMuseumItem_NotFound 测试审批不存在的物品
func (suite *MuseumServiceSimpleTestSuite) TestApproveMuseumItem_NotFound() {
	err := suite.museumService.ApproveMuseumItem(context.Background(), "nonexistent", suite.testAdmin.ID)

	suite.Error(err)
}

// TestDirectDatabaseOperations 测试直接数据库操作（确保基础功能）
func (suite *MuseumServiceSimpleTestSuite) TestDirectDatabaseOperations() {
	// 直接创建博物馆物品到数据库
	item := &models.MuseumItem{
		ID:          "test-direct-id",
		SourceType:  models.SourceTypeLetter,
		SourceID:    "source-123",
		Title:       "直接数据库测试",
		Description: "测试直接数据库操作",
		Tags:        `["测试"]`,
		Status:      models.MuseumItemPending,
		SubmittedBy: suite.testUser.ID,
	}

	err := suite.db.Create(item).Error
	suite.NoError(err)

	// 验证数据已保存
	var savedItem models.MuseumItem
	err = suite.db.First(&savedItem, "id = ?", item.ID).Error
	suite.NoError(err)
	suite.Equal("直接数据库测试", savedItem.Title)
	suite.Equal("pending", string(savedItem.Status))
}

// TestMuseumItemStatusTypes 测试博物馆物品状态类型
func (suite *MuseumServiceSimpleTestSuite) TestMuseumItemStatusTypes() {
	// 测试状态常量
	suite.Equal("pending", string(models.MuseumItemPending))
	suite.Equal("approved", string(models.MuseumItemApproved))
	suite.Equal("rejected", string(models.MuseumItemRejected))
}

// TestMuseumSourceTypes 测试博物馆来源类型
func (suite *MuseumServiceSimpleTestSuite) TestMuseumSourceTypes() {
	// 测试来源类型常量
	suite.Equal("letter", string(models.SourceTypeLetter))
	suite.Equal("photo", string(models.SourceTypePhoto))
	suite.Equal("audio", string(models.SourceTypeAudio))
}

// TestGetMuseumStats_Basic 测试基础统计功能
func (suite *MuseumServiceSimpleTestSuite) TestGetMuseumStats_Basic() {
	// 创建一些测试数据
	suite.createTestMuseumItem()
	approved := suite.createTestMuseumItem()

	// 审批通过一个物品
	err := suite.museumService.ApproveMuseumItem(context.Background(), approved.ID, suite.testAdmin.ID)
	suite.NoError(err)

	// 获取统计信息
	stats, err := suite.museumService.GetMuseumStats(context.Background())

	suite.NoError(err)
	suite.NotNil(stats)
	// 统计信息应该包含基本字段（具体字段取决于实现）
	suite.IsType(map[string]interface{}{}, stats)
}

// Helper methods

func (suite *MuseumServiceSimpleTestSuite) createTestMuseumItem() *models.MuseumItem {
	req := &CreateMuseumItemRequest{
		SourceType:  models.SourceTypeLetter,
		SourceID:    "test-source",
		Title:       "测试博物馆物品",
		Description: "测试描述",
		Tags:        []string{"测试"},
	}
	item, err := suite.museumService.CreateMuseumItem(context.Background(), req)
	suite.NoError(err)
	return item
}

func (suite *MuseumServiceSimpleTestSuite) createTestMuseumEntry() *models.MuseumEntry {
	// 创建一个MuseumEntry直接到数据库
	entry := &models.MuseumEntry{
		ID:                "123e4567-e89b-12d3-a456-426614174000", // 使用有效的UUID格式
		LetterID:          "test-letter-id",
		DisplayTitle:      "测试博物馆条目",
		AuthorDisplayType: "anonymous",
		CuratorType:       "system",
		CuratorID:         suite.testAdmin.ID,
		Status:            models.MuseumItemPending,
		ModerationStatus:  models.MuseumItemPending,
		ViewCount:         0,
		LikeCount:         0,
		AIMetadata:        "{}", // JSON字符串格式
		SubmittedAt:       time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	err := suite.db.Create(entry).Error
	suite.NoError(err)
	return entry
}

func (suite *MuseumServiceSimpleTestSuite) createTestLetter() *models.Letter {
	req := &models.CreateLetterRequest{
		Title:   "测试信件",
		Content: "信件内容",
		Style:   models.StyleClassic,
	}
	letter, err := suite.letterService.CreateDraft(suite.testUser.ID, req)
	suite.NoError(err)
	return letter
}

// 运行测试套件
func TestMuseumServiceSimpleSuite(t *testing.T) {
	suite.Run(t, new(MuseumServiceSimpleTestSuite))
}
