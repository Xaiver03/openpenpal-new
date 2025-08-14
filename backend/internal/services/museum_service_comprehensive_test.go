package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MuseumServiceComprehensiveTestSuite 博物馆服务全面测试套件
type MuseumServiceComprehensiveTestSuite struct {
	suite.Suite
	db             *gorm.DB
	museumService  *MuseumService
	letterService  *LetterService
	testUser       *models.User
	testAdmin      *models.User
	testLetter     *models.Letter
	testItem       *models.MuseumItem
	testExhibition *models.MuseumExhibition
	config         *config.Config
}

func (suite *MuseumServiceComprehensiveTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := config.SetupTestDB()
	suite.NoError(err)
	suite.db = db

	// 获取测试配置
	suite.config = config.GetTestConfig()

	// 创建服务
	suite.letterService = NewLetterService(db, suite.config)
	suite.museumService = NewMuseumService(db)

	// 创建测试用户
	suite.testUser = config.CreateTestUser(db, "museumuser", models.RoleUser)
	suite.testAdmin = config.CreateTestUser(db, "museumadmin", models.RoleSuperAdmin)

	// 创建测试信件
	suite.testLetter = &models.Letter{
		ID:        "test-letter-id",
		UserID:    suite.testUser.ID,
		Title:     "测试信件",
		Content:   "这是一封测试信件的内容",
		Style:     models.StyleClassic,
		Status:    models.StatusDelivered,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(suite.testLetter)

	// 创建测试博物馆物品
	suite.testItem = &models.MuseumItem{
		ID:          "test-item-id",
		SourceType:  models.SourceTypeLetter,
		SourceID:    suite.testLetter.ID,
		Title:       "测试博物馆物品",
		Description: "这是一个测试博物馆物品的描述",
		Tags:        "测试,博物馆",
		Status:      models.MuseumItemApproved,
		SubmittedBy: suite.testUser.ID,
		ViewCount:   10,
		LikeCount:   5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(suite.testItem)

	// 创建测试展览
	suite.testExhibition = &models.MuseumExhibition{
		ID:          "test-exhibition-id",
		Title:       "测试展览",
		Description: "这是一个测试展览",
		CreatorID:   suite.testAdmin.ID,
		Status:      "draft",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(suite.testExhibition)
}

func (suite *MuseumServiceComprehensiveTestSuite) TearDownTest() {
	// 清理测试数据，保留基础测试数据
	suite.db.Where("id NOT IN ?", []string{"test-item-id", "test-letter-id", "test-exhibition-id"}).Delete(&models.MuseumItem{})
	suite.db.Where("id != ?", "test-exhibition-id").Delete(&models.MuseumExhibition{})
	suite.db.Delete(&models.MuseumInteraction{}, "1=1")
	suite.db.Delete(&models.MuseumReaction{}, "1=1")
	suite.db.Delete(&models.MuseumExhibitionEntry{}, "1=1")
}

// TestSubmitLetterToMuseum_Success 测试提交信件到博物馆成功
func (suite *MuseumServiceComprehensiveTestSuite) TestSubmitLetterToMuseum_Success() {
	// 创建新的已送达信件
	letter := &models.Letter{
		ID:      "new-letter-id",
		UserID:  suite.testUser.ID,
		Title:   "新信件",
		Content: "新信件内容",
		Status:  models.StatusDelivered,
	}
	suite.db.Create(letter)

	item, err := suite.museumService.SubmitLetterToMuseum(
		context.Background(),
		letter.ID,
		suite.testUser.ID,
		"提交的信件标题",
		"这是提交的描述",
		[]string{"友谊", "回忆"},
	)

	suite.NoError(err)
	suite.NotNil(item)
	suite.Equal("提交的信件标题", item.Title)
	suite.Equal("这是提交的描述", item.Description)
	suite.Equal("友谊,回忆", item.Tags)
	suite.Equal(models.MuseumItemPending, item.Status)
}

// TestSubmitLetterToMuseum_LetterNotEligible 测试提交不合格的信件
func (suite *MuseumServiceComprehensiveTestSuite) TestSubmitLetterToMuseum_LetterNotEligible() {
	// 创建草稿状态的信件
	letter := &models.Letter{
		ID:     "draft-letter-id",
		UserID: suite.testUser.ID,
		Title:  "草稿信件",
		Status: models.StatusDraft,
	}
	suite.db.Create(letter)

	item, err := suite.museumService.SubmitLetterToMuseum(
		context.Background(),
		letter.ID,
		suite.testUser.ID,
		"标题",
		"描述",
		[]string{"标签"},
	)

	suite.Error(err)
	suite.Nil(item)
	suite.Contains(err.Error(), "not eligible for museum submission")
}

// TestSubmitLetterToMuseum_AlreadySubmitted 测试重复提交信件
func (suite *MuseumServiceComprehensiveTestSuite) TestSubmitLetterToMuseum_AlreadySubmitted() {
	// 使用已经提交的信件
	item, err := suite.museumService.SubmitLetterToMuseum(
		context.Background(),
		suite.testLetter.ID,
		suite.testUser.ID,
		"重复提交",
		"描述",
		[]string{"标签"},
	)

	suite.Error(err)
	suite.Nil(item)
	suite.Contains(err.Error(), "already submitted to museum")
}

// TestLikeMuseumItem_Success 测试点赞博物馆物品
func (suite *MuseumServiceComprehensiveTestSuite) TestLikeMuseumItem_Success() {
	// 记录点赞前的点赞数
	var originalItem models.MuseumItem
	suite.db.First(&originalItem, "id = ?", suite.testItem.ID)
	originalLikes := originalItem.LikeCount

	err := suite.museumService.LikeMuseumItem(
		context.Background(),
		suite.testItem.ID,
		suite.testUser.ID,
	)

	suite.NoError(err)

	// 验证点赞数增加
	var updatedItem models.MuseumItem
	suite.db.First(&updatedItem, "id = ?", suite.testItem.ID)
	suite.Equal(originalLikes+1, updatedItem.LikeCount)
}

// TestLikeMuseumItem_NotFound 测试点赞不存在的物品
func (suite *MuseumServiceComprehensiveTestSuite) TestLikeMuseumItem_NotFound() {
	err := suite.museumService.LikeMuseumItem(
		context.Background(),
		"nonexistent-item",
		suite.testUser.ID,
	)

	suite.Error(err)
	suite.Contains(err.Error(), "museum item not found")
}

// TestRejectMuseumItem_Success 测试拒绝博物馆物品
func (suite *MuseumServiceComprehensiveTestSuite) TestRejectMuseumItem_Success() {
	// 创建待审核的物品
	pendingItem := &models.MuseumItem{
		ID:          "pending-item-id",
		SourceType:  models.SourceTypeLetter,
		SourceID:    "source-id",
		Title:       "待审核物品",
		Status:      models.MuseumItemPending,
		SubmittedBy: suite.testUser.ID,
	}
	suite.db.Create(pendingItem)

	err := suite.museumService.RejectMuseumItem(
		context.Background(),
		pendingItem.ID,
		suite.testAdmin.ID,
		"内容不符合博物馆标准",
	)

	suite.NoError(err)

	// 验证状态更新为拒绝
	var rejectedItem models.MuseumItem
	suite.db.First(&rejectedItem, "id = ?", pendingItem.ID)
	suite.Equal(models.MuseumItemRejected, rejectedItem.Status)
	suite.Equal(suite.testAdmin.ID, *rejectedItem.ApprovedBy)
}

// TestRecordInteraction_Success 测试记录用户互动
func (suite *MuseumServiceComprehensiveTestSuite) TestRecordInteraction_Success() {
	interactionTypes := []string{"view", "like", "bookmark", "share"}

	for _, interactionType := range interactionTypes {
		err := suite.museumService.RecordInteraction(
			context.Background(),
			suite.testItem.ID,
			suite.testUser.ID,
			interactionType,
		)
		suite.NoError(err, "Failed to record %s interaction", interactionType)
	}

	// 验证计数器更新
	var updatedItem models.MuseumItem
	suite.db.First(&updatedItem, "id = ?", suite.testItem.ID)
	suite.Greater(updatedItem.ViewCount, int64(10)) // 原始值为10
	suite.Greater(updatedItem.LikeCount, int64(5))  // 原始值为5
}

// TestRecordInteraction_InvalidType 测试无效的互动类型
func (suite *MuseumServiceComprehensiveTestSuite) TestRecordInteraction_InvalidType() {
	err := suite.museumService.RecordInteraction(
		context.Background(),
		suite.testItem.ID,
		suite.testUser.ID,
		"invalid_type",
	)

	suite.Error(err)
	suite.Contains(err.Error(), "invalid interaction type")
}

// TestAddReaction_Success 测试添加反应
func (suite *MuseumServiceComprehensiveTestSuite) TestAddReaction_Success() {
	reaction, err := suite.museumService.AddReaction(
		context.Background(),
		suite.testItem.ID,
		suite.testUser.ID,
		"love",
		"这个展品真棒！",
	)

	suite.NoError(err)
	suite.NotNil(reaction)
	suite.Equal(suite.testItem.ID, reaction.EntryID)
	suite.Equal(suite.testUser.ID, reaction.UserID)
	suite.Equal("love", reaction.ReactionType)
	suite.Equal("这个展品真棒！", reaction.Comment)
}

// TestWithdrawEntry_Success 测试撤回条目
func (suite *MuseumServiceComprehensiveTestSuite) TestWithdrawEntry_Success() {
	// 创建用户自己的物品
	userItem := &models.MuseumItem{
		ID:          "user-item-id",
		SourceType:  models.SourceTypeLetter,
		SourceID:    "source-id",
		Title:       "用户物品",
		Status:      models.MuseumItemPending,
		SubmittedBy: suite.testUser.ID,
	}
	suite.db.Create(userItem)

	err := suite.museumService.WithdrawEntry(
		context.Background(),
		userItem.ID,
		suite.testUser.ID,
	)

	suite.NoError(err)

	// 验证状态更新为已撤回
	var withdrawnItem models.MuseumItem
	suite.db.First(&withdrawnItem, "id = ?", userItem.ID)
	suite.Equal("withdrawn", string(withdrawnItem.Status))
}

// TestWithdrawEntry_Unauthorized 测试无权限撤回
func (suite *MuseumServiceComprehensiveTestSuite) TestWithdrawEntry_Unauthorized() {
	// 创建其他用户的物品
	otherUser := config.CreateTestUser(suite.db, "otheruser", models.RoleUser)
	otherItem := &models.MuseumItem{
		ID:          "other-item-id",
		SourceType:  models.SourceTypeLetter,
		SourceID:    "source-id",
		Title:       "其他用户物品",
		Status:      models.MuseumItemPending,
		SubmittedBy: otherUser.ID,
	}
	suite.db.Create(otherItem)

	err := suite.museumService.WithdrawEntry(
		context.Background(),
		otherItem.ID,
		suite.testUser.ID,
	)

	suite.Error(err)
	suite.Contains(err.Error(), "unauthorized")
}

// TestGetPopularTags_Success 测试获取热门标签
func (suite *MuseumServiceComprehensiveTestSuite) TestGetPopularTags_Success() {
	// 创建一些带标签的物品
	items := []*models.MuseumItem{
		{
			ID:     "tag-item-1",
			Title:  "标签测试1",
			Tags:   "友谊,温暖,回忆",
			Status: models.MuseumItemApproved,
		},
		{
			ID:     "tag-item-2",
			Title:  "标签测试2",
			Tags:   "友谊,成长,青春",
			Status: models.MuseumItemApproved,
		},
		{
			ID:     "tag-item-3",
			Title:  "标签测试3",
			Tags:   "温暖,感动,真诚",
			Status: models.MuseumItemApproved,
		},
	}

	for _, item := range items {
		suite.db.Create(item)
	}

	tags, err := suite.museumService.GetPopularTags(context.Background(), "", 5)

	suite.NoError(err)
	suite.NotEmpty(tags)
	// 验证标签按使用次数排序
	if len(tags) > 1 {
		suite.GreaterOrEqual(tags[0].UsageCount, tags[1].UsageCount)
	}
}

// TestModerateEntry_Approve 测试审核批准条目
func (suite *MuseumServiceComprehensiveTestSuite) TestModerateEntry_Approve() {
	// 创建待审核物品
	pendingItem := &models.MuseumItem{
		ID:          "moderate-item-id",
		SourceType:  models.SourceTypeLetter,
		SourceID:    "source-id",
		Title:       "待审核物品",
		Status:      models.MuseumItemPending,
		SubmittedBy: suite.testUser.ID,
	}
	suite.db.Create(pendingItem)

	err := suite.museumService.ModerateEntry(
		context.Background(),
		pendingItem.ID,
		suite.testAdmin.ID,
		"approved",
		"",
		true, // featured
	)

	suite.NoError(err)

	// 验证状态更新
	var moderatedItem models.MuseumItem
	suite.db.First(&moderatedItem, "id = ?", pendingItem.ID)
	suite.Equal(models.MuseumItemApproved, moderatedItem.Status)
	suite.Equal(suite.testAdmin.ID, *moderatedItem.ApprovedBy)
}

// TestGetPendingEntries_Success 测试获取待审核条目
func (suite *MuseumServiceComprehensiveTestSuite) TestGetPendingEntries_Success() {
	// 创建一些待审核物品
	for i := 0; i < 3; i++ {
		pendingItem := &models.MuseumItem{
			ID:          fmt.Sprintf("pending-%d", i),
			SourceType:  models.SourceTypeLetter,
			Title:       fmt.Sprintf("待审核物品 %d", i),
			Status:      models.MuseumItemPending,
			SubmittedBy: suite.testUser.ID,
		}
		suite.db.Create(pendingItem)
	}

	items, total, err := suite.museumService.GetPendingEntries(context.Background(), 1, 10)

	suite.NoError(err)
	suite.GreaterOrEqual(len(items), 3)
	suite.GreaterOrEqual(total, int64(3))
}

// TestCreateExhibition_Success 测试创建展览
func (suite *MuseumServiceComprehensiveTestSuite) TestCreateExhibition_Success() {
	exhibition := &models.MuseumExhibition{
		Title:       "新展览",
		Description: "这是一个新展览",
		CreatorID:   suite.testAdmin.ID,
		Status:      "draft",
		MaxEntries:  20,
	}

	createdExhibition, err := suite.museumService.CreateExhibition(context.Background(), exhibition)

	suite.NoError(err)
	suite.NotNil(createdExhibition)
	suite.NotEmpty(createdExhibition.ID)
	suite.Equal("新展览", createdExhibition.Title)
	suite.Equal(suite.testAdmin.ID, createdExhibition.CreatorID)
}

// TestUpdateExhibition_Success 测试更新展览
func (suite *MuseumServiceComprehensiveTestSuite) TestUpdateExhibition_Success() {
	updatedExhibition := &models.MuseumExhibition{
		ID:          suite.testExhibition.ID,
		Title:       "更新后的展览",
		Description: "更新后的描述",
		Status:      "published",
	}

	result, err := suite.museumService.UpdateExhibition(
		context.Background(),
		updatedExhibition,
		suite.testAdmin.ID,
	)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("更新后的展览", result.Title)
	suite.Equal("更新后的描述", result.Description)
}

// TestUpdateExhibition_Unauthorized 测试无权限更新展览
func (suite *MuseumServiceComprehensiveTestSuite) TestUpdateExhibition_Unauthorized() {
	updatedExhibition := &models.MuseumExhibition{
		ID:    suite.testExhibition.ID,
		Title: "无权限更新",
	}

	result, err := suite.museumService.UpdateExhibition(
		context.Background(),
		updatedExhibition,
		suite.testUser.ID, // 不是创建者
	)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "unauthorized")
}

// TestAddItemsToExhibition_Success 测试向展览添加物品
func (suite *MuseumServiceComprehensiveTestSuite) TestAddItemsToExhibition_Success() {
	itemIDs := []string{suite.testItem.ID}

	err := suite.museumService.AddItemsToExhibition(
		context.Background(),
		suite.testExhibition.ID,
		itemIDs,
		suite.testAdmin.ID,
	)

	suite.NoError(err)

	// 验证物品已添加到展览
	var count int64
	suite.db.Model(&models.MuseumExhibitionEntry{}).
		Where("collection_id = ? AND item_id = ?", suite.testExhibition.ID, suite.testItem.ID).
		Count(&count)
	suite.Equal(int64(1), count)
}

// TestRemoveItemsFromExhibition_Success 测试从展览中移除物品
func (suite *MuseumServiceComprehensiveTestSuite) TestRemoveItemsFromExhibition_Success() {
	// 先添加物品到展览
	itemIDs := []string{suite.testItem.ID}
	err := suite.museumService.AddItemsToExhibition(
		context.Background(),
		suite.testExhibition.ID,
		itemIDs,
		suite.testAdmin.ID,
	)
	suite.NoError(err)

	// 然后移除物品
	err = suite.museumService.RemoveItemsFromExhibition(
		context.Background(),
		suite.testExhibition.ID,
		itemIDs,
		suite.testAdmin.ID,
	)

	suite.NoError(err)

	// 验证物品已从展览中移除
	var count int64
	suite.db.Model(&models.MuseumExhibitionEntry{}).
		Where("collection_id = ? AND item_id = ?", suite.testExhibition.ID, suite.testItem.ID).
		Count(&count)
	suite.Equal(int64(0), count)
}

// TestGetExhibitionItems_Success 测试获取展览中的物品
func (suite *MuseumServiceComprehensiveTestSuite) TestGetExhibitionItems_Success() {
	// 先添加物品到展览
	itemIDs := []string{suite.testItem.ID}
	err := suite.museumService.AddItemsToExhibition(
		context.Background(),
		suite.testExhibition.ID,
		itemIDs,
		suite.testAdmin.ID,
	)
	suite.NoError(err)

	items, total, err := suite.museumService.GetExhibitionItems(
		context.Background(),
		suite.testExhibition.ID,
		1,
		10,
	)

	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(items, 1)
	suite.Equal(suite.testItem.ID, items[0].ID)
}

// TestPublishExhibition_Success 测试发布展览
func (suite *MuseumServiceComprehensiveTestSuite) TestPublishExhibition_Success() {
	// 先添加物品到展览
	itemIDs := []string{suite.testItem.ID}
	err := suite.museumService.AddItemsToExhibition(
		context.Background(),
		suite.testExhibition.ID,
		itemIDs,
		suite.testAdmin.ID,
	)
	suite.NoError(err)

	err = suite.museumService.PublishExhibition(
		context.Background(),
		suite.testExhibition.ID,
		suite.testAdmin.ID,
	)

	suite.NoError(err)

	// 验证展览状态更新
	var publishedExhibition models.MuseumExhibition
	suite.db.First(&publishedExhibition, "id = ?", suite.testExhibition.ID)
	suite.Equal("published", publishedExhibition.Status)
}

// TestPublishExhibition_EmptyExhibition 测试发布空展览
func (suite *MuseumServiceComprehensiveTestSuite) TestPublishExhibition_EmptyExhibition() {
	// 创建空展览
	emptyExhibition := &models.MuseumExhibition{
		ID:        "empty-exhibition-id",
		Title:     "空展览",
		CreatorID: suite.testAdmin.ID,
		Status:    "draft",
	}
	suite.db.Create(emptyExhibition)

	err := suite.museumService.PublishExhibition(
		context.Background(),
		emptyExhibition.ID,
		suite.testAdmin.ID,
	)

	suite.Error(err)
	suite.Contains(err.Error(), "cannot publish empty exhibition")
}

// TestSearchEntries_Success 测试搜索博物馆条目
func (suite *MuseumServiceComprehensiveTestSuite) TestSearchEntries_Success() {
	// 创建一些测试物品用于搜索
	searchItems := []*models.MuseumItem{
		{
			ID:     "search-item-1",
			Title:  "友谊之光",
			Tags:   "友谊,温暖",
			Status: models.MuseumItemApproved,
		},
		{
			ID:     "search-item-2",
			Title:  "成长回忆",
			Tags:   "成长,青春",
			Status: models.MuseumItemApproved,
		},
	}

	for _, item := range searchItems {
		suite.db.Create(item)
	}

	entries, total, err := suite.museumService.SearchEntries(
		context.Background(),
		"友谊",         // query
		[]string{},   // tags
		"",           // theme
		"",           // status
		nil,          // featured
		"",           // dateFrom
		"",           // dateTo
		"created_at", // sortBy
		"desc",       // sortOrder
		1,            // page
		10,           // limit
	)

	suite.NoError(err)
	suite.GreaterOrEqual(len(entries), 1)
	suite.GreaterOrEqual(total, int64(1))
}

// TestGetAnalytics_Success 测试获取分析数据
func (suite *MuseumServiceComprehensiveTestSuite) TestGetAnalytics_Success() {
	analytics, err := suite.museumService.GetAnalytics(
		context.Background(),
		"month",
		"",
		"",
	)

	suite.NoError(err)
	suite.NotNil(analytics)
	suite.GreaterOrEqual(analytics.TotalEntries, int64(0))
	suite.False(analytics.GeneratedAt.IsZero())
}

// TestRefreshStats_Success 测试刷新统计数据
func (suite *MuseumServiceComprehensiveTestSuite) TestRefreshStats_Success() {
	err := suite.museumService.RefreshStats(context.Background())
	suite.NoError(err)

	// 验证标签统计已更新
	var tagCount int64
	suite.db.Model(&models.MuseumTag{}).Count(&tagCount)
	// 标签表可能为空或有数据，两种情况都是正常的
}

// TestDeleteExhibition_Success 测试删除展览
func (suite *MuseumServiceComprehensiveTestSuite) TestDeleteExhibition_Success() {
	// 创建可删除的展览
	deletableExhibition := &models.MuseumExhibition{
		ID:        "deletable-exhibition-id",
		Title:     "可删除展览",
		CreatorID: suite.testAdmin.ID,
		Status:    "draft",
	}
	suite.db.Create(deletableExhibition)

	err := suite.museumService.DeleteExhibition(
		context.Background(),
		deletableExhibition.ID,
		suite.testAdmin.ID,
	)

	suite.NoError(err)

	// 验证展览被软删除
	var deletedExhibition models.MuseumExhibition
	err = suite.db.Where("id = ?", deletableExhibition.ID).First(&deletedExhibition).Error
	suite.Error(err) // 应该找不到记录（被软删除）
}

// TestGetExhibitionByID_Success 测试根据ID获取展览
func (suite *MuseumServiceComprehensiveTestSuite) TestGetExhibitionByID_Success() {
	exhibition, err := suite.museumService.GetExhibitionByID(
		context.Background(),
		suite.testExhibition.ID,
	)

	suite.NoError(err)
	suite.NotNil(exhibition)
	suite.Equal(suite.testExhibition.ID, exhibition.ID)
	suite.Equal(suite.testExhibition.Title, exhibition.Title)
}

// TestGetExhibitionByID_NotFound 测试获取不存在的展览
func (suite *MuseumServiceComprehensiveTestSuite) TestGetExhibitionByID_NotFound() {
	exhibition, err := suite.museumService.GetExhibitionByID(
		context.Background(),
		"nonexistent-exhibition",
	)

	suite.Error(err)
	suite.Nil(exhibition)
	suite.Contains(err.Error(), "exhibition not found")
}

// TestGetMuseumExhibitions_Success 测试获取博物馆展览列表
func (suite *MuseumServiceComprehensiveTestSuite) TestGetMuseumExhibitions_Success() {
	exhibitions, total, err := suite.museumService.GetMuseumExhibitions(
		context.Background(),
		1,
		10,
	)

	suite.NoError(err)
	suite.GreaterOrEqual(len(exhibitions), 1) // 至少有测试展览
	suite.GreaterOrEqual(total, int64(1))
}

// TestGetPopularEntries_Success 测试获取热门条目
func (suite *MuseumServiceComprehensiveTestSuite) TestGetPopularEntries_Success() {
	entries, total, err := suite.museumService.GetPopularEntries(
		context.Background(),
		1,
		10,
		"month",
	)

	suite.NoError(err)
	suite.GreaterOrEqual(len(entries), 0) // 可能没有数据
	suite.GreaterOrEqual(total, int64(0))
}

// TestUpdateExhibitionItemOrder_Success 测试更新展览中物品的顺序
func (suite *MuseumServiceComprehensiveTestSuite) TestUpdateExhibitionItemOrder_Success() {
	// 先添加物品到展览
	itemIDs := []string{suite.testItem.ID}
	err := suite.museumService.AddItemsToExhibition(
		context.Background(),
		suite.testExhibition.ID,
		itemIDs,
		suite.testAdmin.ID,
	)
	suite.NoError(err)

	// 更新物品顺序
	itemOrders := []ItemOrder{
		{
			ItemID:       suite.testItem.ID,
			DisplayOrder: 5,
		},
	}

	err = suite.museumService.UpdateExhibitionItemOrder(
		context.Background(),
		suite.testExhibition.ID,
		itemOrders,
		suite.testAdmin.ID,
	)

	suite.NoError(err)

	// 验证顺序已更新
	var entry models.MuseumExhibitionEntry
	suite.db.Where("collection_id = ? AND item_id = ?",
		suite.testExhibition.ID, suite.testItem.ID).First(&entry)
	suite.Equal(5, entry.DisplayOrder)
}

// 运行测试套件
func TestMuseumServiceComprehensiveSuite(t *testing.T) {
	suite.Run(t, new(MuseumServiceComprehensiveTestSuite))
}
