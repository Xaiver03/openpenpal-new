# OpenPenPal AI子系统技术设计文档

## 1. 概述

### 1.1 设计目标
- 创建可扩展的AI服务抽象层，支持多种AI服务提供商
- 实现开发阶段Mock服务和生产环境真实AI API的无缝切换
- 支持OpenPenPal四大AI核心功能：笔友匹配、笔友模拟器、写作灵感、信件策展

### 1.2 技术架构
```
前端界面 → 后端API → AI服务抽象层 → AI实现层 (Mock/真实API)
```

## 2. AI服务抽象层设计

### 2.1 核心接口定义

```go
// AIService AI服务核心接口
type AIService interface {
    // 笔友匹配：根据信件内容和用户画像匹配合适的笔友
    MatchPenpal(ctx context.Context, req *PenpalMatchRequest) (*PenpalMatchResponse, error)
    
    // AI笔友模拟器：根据人格设定生成回信
    GenerateReply(ctx context.Context, req *AIReplyRequest) (*AIReplyResponse, error)
    
    // 写作灵感卡片：生成个性化写作提示
    GenerateWritingPrompt(ctx context.Context, req *WritingPromptRequest) (*WritingPromptResponse, error)
    
    // 信件分类：为博物馆策展分析信件主题
    CategorizeLetter(ctx context.Context, req *LetterCategoryRequest) (*LetterCategoryResponse, error)
}
```

### 2.2 数据结构定义

```go
// 笔友匹配请求
type PenpalMatchRequest struct {
    LetterContent   string      `json:"letter_content"`   // 信件内容
    SenderProfile   UserProfile `json:"sender_profile"`   // 发送者画像
    MatchType       string      `json:"match_type"`       // 匹配类型：random/interest_based
    ExcludeUserIDs  []string    `json:"exclude_user_ids"` // 排除的用户ID
}

// 笔友匹配响应
type PenpalMatchResponse struct {
    MatchedUserID   string  `json:"matched_user_id"`  // 匹配到的用户ID
    MatchScore      float64 `json:"match_score"`      // 匹配分数 0-1
    MatchReason     string  `json:"match_reason"`     // 匹配原因
    LetterThemes    []string `json:"letter_themes"`   // 信件主题标签
}

// AI回信请求
type AIReplyRequest struct {
    PersonaID       string `json:"persona_id"`       // 人格ID
    OriginalLetter  string `json:"original_letter"`  // 原始信件
    ConversationID  string `json:"conversation_id"`  // 对话ID
    DelayHours      int    `json:"delay_hours"`      // 延迟回信时间
}

// AI回信响应
type AIReplyResponse struct {
    ReplyContent    string    `json:"reply_content"`    // 回信内容
    ReplyTone       string    `json:"reply_tone"`       // 回信语调
    ScheduledTime   time.Time `json:"scheduled_time"`   // 预定发送时间
    ConversationID  string    `json:"conversation_id"`  // 对话ID
}

// 写作提示请求
type WritingPromptRequest struct {
    UserInterests   []string `json:"user_interests"`   // 用户兴趣
    PromptType      string   `json:"prompt_type"`      // 提示类型：daily/themed/challenge
    Difficulty      string   `json:"difficulty"`       // 难度：easy/medium/hard
}

// 写作提示响应
type WritingPromptResponse struct {
    PromptTitle     string   `json:"prompt_title"`     // 提示标题
    PromptContent   string   `json:"prompt_content"`   // 提示内容
    Keywords        []string `json:"keywords"`         // 关键词
    EstimatedTime   int      `json:"estimated_time"`   // 预估写作时间(分钟)
}

// 信件分类请求
type LetterCategoryRequest struct {
    LetterContent   string `json:"letter_content"`   // 信件内容
    SenderAge       int    `json:"sender_age"`       // 发送者年龄
    SenderGender    string `json:"sender_gender"`    // 发送者性别
}

// 信件分类响应
type LetterCategoryResponse struct {
    PrimaryCategory string            `json:"primary_category"`  // 主要分类
    SubCategories   []string          `json:"sub_categories"`    // 子分类
    Themes          []string          `json:"themes"`            // 主题标签
    Emotions        map[string]float64 `json:"emotions"`         // 情感分析
    MuseumSuitable  bool             `json:"museum_suitable"`   // 是否适合博物馆展示
}

// 用户画像
type UserProfile struct {
    UserID          string   `json:"user_id"`
    Age             int      `json:"age"`
    Gender          string   `json:"gender"`
    Interests       []string `json:"interests"`
    WritingStyle    string   `json:"writing_style"`    // formal/casual/poetic
    PreferredTopics []string `json:"preferred_topics"`
}
```

## 3. 实现层设计

### 3.1 Mock实现 (开发阶段)

```go
type MockAIService struct {
    config *MockConfig
}

type MockConfig struct {
    ResponseDelay   time.Duration `json:"response_delay"`
    ErrorRate       float64       `json:"error_rate"`
    EnableLogging   bool          `json:"enable_logging"`
}

func NewMockAIService(config *MockConfig) *MockAIService {
    return &MockAIService{config: config}
}

func (m *MockAIService) MatchPenpal(ctx context.Context, req *PenpalMatchRequest) (*PenpalMatchResponse, error) {
    // 模拟网络延迟
    time.Sleep(m.config.ResponseDelay)
    
    // 模拟错误
    if rand.Float64() < m.config.ErrorRate {
        return nil, errors.New("mock AI service error")
    }
    
    // 返回预设匹配结果
    mockUsers := []string{"user_001", "user_002", "user_003"}
    selectedUser := mockUsers[rand.Intn(len(mockUsers))]
    
    return &PenpalMatchResponse{
        MatchedUserID: selectedUser,
        MatchScore:    0.75 + rand.Float64()*0.25,
        MatchReason:   "都喜欢文学和旅行，写作风格相近",
        LetterThemes:  []string{"友情", "成长", "梦想"},
    }, nil
}
```

### 3.2 真实AI实现 (生产阶段)

```go
// OpenAI实现
type OpenAIService struct {
    client    *openai.Client
    config    *OpenAIConfig
    templates *PromptTemplates
}

type OpenAIConfig struct {
    APIKey      string `json:"api_key"`
    Model       string `json:"model"`
    MaxTokens   int    `json:"max_tokens"`
    Temperature float32 `json:"temperature"`
}

// 智谱AI实现  
type ZhipuAIService struct {
    client *zhipu.Client
    config *ZhipuConfig
}

// 统一工厂方法
func NewAIService(provider string, config interface{}) (AIService, error) {
    switch provider {
    case "mock":
        return NewMockAIService(config.(*MockConfig)), nil
    case "openai":
        return NewOpenAIService(config.(*OpenAIConfig)), nil
    case "zhipu":
        return NewZhipuAIService(config.(*ZhipuConfig)), nil
    default:
        return nil, fmt.Errorf("unsupported AI provider: %s", provider)
    }
}
```

## 4. 数据库设计

### 4.1 AI相关表结构

```sql
-- AI笔友模拟器人格表
CREATE TABLE ai_personas (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    personality_traits JSON,
    writing_style VARCHAR(50),
    response_patterns JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- AI对话记录表
CREATE TABLE ai_conversations (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    persona_id VARCHAR(36),
    conversation_type ENUM('penpal_match', 'ai_simulator') NOT NULL,
    status ENUM('active', 'paused', 'ended') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (persona_id) REFERENCES ai_personas(id)
);

-- AI消息记录表
CREATE TABLE ai_messages (
    id VARCHAR(36) PRIMARY KEY,
    conversation_id VARCHAR(36) NOT NULL,
    sender_type ENUM('user', 'ai') NOT NULL,
    content TEXT NOT NULL,
    scheduled_time TIMESTAMP NULL,
    sent_time TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES ai_conversations(id)
);

-- 写作提示表
CREATE TABLE writing_prompts (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    prompt_type VARCHAR(50),
    difficulty ENUM('easy', 'medium', 'hard'),
    keywords JSON,
    estimated_time INT,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 信件分类记录表
CREATE TABLE letter_categories (
    id VARCHAR(36) PRIMARY KEY,
    letter_id VARCHAR(36) NOT NULL,
    primary_category VARCHAR(100),
    sub_categories JSON,
    themes JSON,
    emotions JSON,
    museum_suitable BOOLEAN DEFAULT FALSE,
    ai_confidence FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (letter_id) REFERENCES letters(id)
);
```

## 5. API接口设计

### 5.1 RESTful API端点

```
POST /api/v1/ai/penpal/match          # 笔友匹配
POST /api/v1/ai/penpal/reply          # AI笔友回信
GET  /api/v1/ai/writing/prompts       # 获取写作提示
POST /api/v1/ai/writing/prompts       # 生成个性化写作提示
POST /api/v1/ai/letters/categorize    # 信件分类
GET  /api/v1/ai/personas              # 获取AI人格列表
POST /api/v1/ai/conversations         # 创建AI对话
GET  /api/v1/ai/conversations/:id     # 获取对话历史
```

### 5.2 接口实现示例

```go
// AI控制器
type AIController struct {
    aiService     AIService
    letterService LetterService
    userService   UserService
}

// 笔友匹配接口
func (c *AIController) MatchPenpal(ctx *gin.Context) {
    var req PenpalMatchRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        response.Error(ctx, "参数错误", err)
        return
    }
    
    // 获取用户画像
    userProfile, err := c.userService.GetUserProfile(req.SenderProfile.UserID)
    if err != nil {
        response.Error(ctx, "获取用户信息失败", err)
        return
    }
    req.SenderProfile = userProfile
    
    // 调用AI服务
    result, err := c.aiService.MatchPenpal(ctx, &req)
    if err != nil {
        response.Error(ctx, "AI匹配服务异常", err)
        return
    }
    
    response.Success(ctx, result)
}
```

## 6. 配置管理

### 6.1 配置文件结构

```yaml
ai:
  provider: "mock"  # mock/openai/zhipu
  mock:
    response_delay: "500ms"
    error_rate: 0.1
    enable_logging: true
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-3.5-turbo"
    max_tokens: 1000
    temperature: 0.7
  zhipu:
    api_key: "${ZHIPU_API_KEY}"
    model: "glm-4"
```

### 6.2 环境变量管理

```bash
# 开发环境
AI_PROVIDER=mock
AI_MOCK_DELAY=500ms
AI_MOCK_ERROR_RATE=0.1

# 生产环境
AI_PROVIDER=openai
OPENAI_API_KEY=sk-xxxxx
AI_MODEL=gpt-3.5-turbo
```

## 7. 测试策略

### 7.1 单元测试
- AI服务接口的Mock实现测试
- 数据结构序列化/反序列化测试
- 错误处理测试

### 7.2 集成测试
- 完整AI功能流程测试
- 不同AI提供商切换测试
- 性能和并发测试

### 7.3 用户验收测试
- 使用Mock数据验证前端交互流程
- AI功能的用户体验测试

## 8. 部署和监控

### 8.1 分阶段部署
1. **开发阶段**: 使用Mock服务完成功能开发
2. **测试阶段**: 接入真实AI API进行功能验证
3. **生产阶段**: 完整AI服务上线

### 8.2 监控指标
- AI API调用成功率
- 响应时间分布
- 成本监控（API调用费用）
- 用户满意度指标

## 9. 未来扩展

### 9.1 支持更多AI服务商
- 百度千帆
- 阿里通义
- 自部署开源模型

### 9.2 高级功能
- 用户偏好学习
- 个性化模型微调
- 多模态支持（图像、语音）

---

*本文档版本: v1.0*  
*最后更新: 2025-07-25*