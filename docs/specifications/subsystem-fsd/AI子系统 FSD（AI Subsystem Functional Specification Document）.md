
> 面向 LLM API 调用设计，强调解耦合与功能注入式结构。

---

## **一、模块定位**

OpenPenPal 的 AI 子系统通过接入大模型 API，为写信体验提供辅助与增强功能。它不参与核心社交关系构建，而是作为 “温柔、慢节奏、可控”的智能助手，服务于匿名匹配、笔友模拟、写作引导与信件策展等场景。

---

## **二、系统结构概览**

```
flowchart TD
  User -->|写匿名信| AI_Matcher -->|返回目标编码| Barcode_Module
  User -->|开启云中锦书| AI_Reply -->|周期性生成信件| Barcode_Module
  User -->|获取灵感| AI_Inspiration
  User -->|提交公开信| AI_Curation --> Museum_System
```

---

## **三、功能模块与接口设计**

### **3.1 自由笔友匹配（AI Matching）**

**用途**：匿名信用户选择“由平台匹配笔友”时，AI 根据信件内容与用户画像返回最合适收件人编码（不返回 ID）。

**调用链路**：

- 输入：用户信件内容 + 元数据（如学校、性别、活跃度标签）
    
- 输出：推荐编码（OP Code） + 匹配置信度

**API Schema**：

```
POST /api/ai/match
{
  "letter_id": "abc123",
  "content": "string",
  "meta": {
    "school_code": "PK",
    "user_tags": ["夜猫子", "情绪细腻"]
  }
}
```

**返回**：

```
{
  "matched_code": "PK5FSP",
  "reason": "匹配对方喜好：深夜写信、文学风格",
  "confidence": 0.84
}
```

---

### **3.2 云中锦书（Cloud Letter Companion）**

**用途**：用户选择一个长期陪伴的AI笔友人设，建立持续的书信往来关系，AI保持一致的性格和记忆，每隔 1–3 天自动回信。

**调用链路**：

- 用户首封信 + 已设定角色人格描述
    
- 返回内容为回信草稿（由平台贴条码后投递）

**API Schema**：

```
POST /api/ai/virtual-reply
{
  "session_id": "user_abc__writer_poetic",
  "letter_content": "string",
  "context": ["历史回信内容"]
}
```

**返回**：

```
{
  "reply_content": "亲爱的你：看到你说....",
  "persona": "高冷作家",
  "delay_days": 2
}
```

---

### **3.3 写作灵感推送（Inspiration Cards）**

**用途**：用户点击获取每日灵感，AI 提供 prompt + 配图。

**API Schema**：

```
GET /api/ai/inspiration

Response:
{
  "prompt": "如果你能寄一封信给一年前的自己，你会写什么？",
  "type": "回忆",
  "image_url": "cdn.openpenpal.ai/prompt42.jpg"
}
```

---

### **3.4 角色驿站（Character Station）**

**用途**：基于不同角色视角为用户的回信提供思路和建议，支持自定义角色和情感引导。

**调用链路**：

- 输入：来信内容 + 自定义角色描述 + 关系类型
    
- 输出：回信角度建议 + 情感策略 + 语气建议

**API Schema**：

```
POST /api/ai/reply-advice
{
  "original_letter": "string",
  "relationship": "friend|elder|classmate|custom",
  "custom_persona": "温柔的学姐，善于倾听",
  "response_style": "warm|formal|casual"
}
```

**返回**：

```
{
  "advice": {
    "perspective": "从学姐的角度",
    "key_points": ["回应对方的困惑", "分享相似经历", "给予温暖鼓励"],
    "tone_suggestion": "温柔而有力量",
    "emotion_strategy": "共情 + 支持"
  },
  "sample_openings": ["看到你的困惑，让我想起了...", "作为过来人..."]
}
```

---

### **3.5 信件策展引擎（Letter Curation）**

**用途**：用户公开信件时，AI 自动归类主题并推荐到博物馆栏目（如“告别特展”“暗恋”）。

**调用链路**：

- 输入：信件内容 + 可选标签
    
- 输出：主题归类、展示排版建议

**API Schema**：

```
POST /api/ai/curate
{
  "letter_id": "abc123",
  "content": "string",
  "tags": ["毕业", "送别"]
}
```

**返回**：

```
{
  "category": "告别",
  "section": "夏天的最后一封信",
  "recommended_display": {
    "layout": "双栏",
    "quote_highlight": "我们可能不会再见，但我不会忘记你"
  }
}
```

---

## **四、模型调用策略建议（可部署型 / 云API型兼容）**

|**功能点**|**推荐模型能力**|**说明**|
|---|---|---|
|AI匹配|embedding + 向量召回|支持 Faiss/Weaviate，本地或云|
|云中锦书回信|gpt-4, claude, moonshot 等|多轮历史记忆支持，长期人设一致性|
|角色驿站建议|gpt-3.5-turbo, claude-haiku|角色视角分析，情感引导|
|灵感生成|prompt + fine-tuned RAG|可本地生成，不强依赖大模型|
|策展与分类|gpt + rule-based logic|可部署半结构化 prompt 分类器|

---

## **五、安全与节奏控制机制**

|**控制点**|**描述**|
|---|---|
|匿名匹配保护|匹配仅返回编码，不泄露 ID|
|写作内容预审|所有 AI 公开内容需人工二审|
|AI身份标签显示|每封 AI 生成信需标记“由 AI 笔友生成”|
|节奏延迟机制|回信需延迟 1–3 天发送，模拟真实等待|
|情绪内容拦截|自动检测极端情绪，提示用户谨慎使用|

---

## **六、前后端接口对接图**

```
flowchart TD
A[用户写匿名信] --> B[POST /ai/match]
B --> C[返回匹配 Code] --> D[生成条码并投递]

E[用户开启 AI 笔友] --> F[POST /ai/virtual-reply]
F --> G[生成草稿 → 加条码 → 发出]

H[用户点击写作灵感] --> I[GET /ai/inspiration]
```

---

## **七、未来可扩展接口（预留）**

|**功能**|**描述**|
|---|---|
|情绪缓解建议|信件中如含强烈负面情绪，返回慰问语或建议|
|AI写信挑战|每周主题挑战，如“孤独信挑战”，AI参与创作|
|多语言笔友|跨语种信件翻译/AI生成|

---

## **八、当前实现状态（2025年1月）**

### **8.1 API实现对照表**

| API接口 | 设计路径 | 实际实现 | 状态 | 说明 |
|---------|----------|----------|------|------|
| AI匹配 | POST /api/ai/match | POST /api/ai/match | ✅ 完成 | 完全符合设计 |
| 云中锦书 | POST /api/ai/virtual-reply | POST /api/ai/reply | ✅ 完成 | 路径略有差异 |
| 写作灵感 | GET /api/ai/inspiration | POST /api/ai/inspiration | ✅ 完成 | 改为POST支持参数 |
| 角色驿站 | POST /api/ai/reply-advice | POST /api/ai/reply-advice | ✅ 完成 | 完全符合设计 |
| 信件策展 | POST /api/ai/curate | POST /api/ai/curate | ✅ 完成 | 后端完成，前端待集成 |

### **8.2 数据模型实现**

**已实现的核心数据表**：
```sql
-- AI配置表
CREATE TABLE ai_configs (
    id VARCHAR(36) PRIMARY KEY,
    provider VARCHAR(50),        -- openai/claude/siliconflow
    api_key TEXT,
    api_endpoint TEXT,
    model VARCHAR(100),
    temperature FLOAT,
    max_tokens INT,
    is_active BOOLEAN,
    priority INT,
    daily_quota INT,
    used_quota INT,
    quota_reset_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- AI使用日志表
CREATE TABLE ai_usage_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    task_type VARCHAR(50),       -- match/reply/inspiration/curate
    task_id VARCHAR(36),
    provider VARCHAR(50),
    model VARCHAR(100),
    input_tokens INT,
    output_tokens INT,
    total_tokens INT,
    status VARCHAR(20),
    error_message TEXT,
    created_at TIMESTAMP
);

-- AI回信建议表
CREATE TABLE ai_reply_advices (
    id VARCHAR(36) PRIMARY KEY,
    letter_id VARCHAR(36),
    user_id VARCHAR(36),
    persona_type VARCHAR(50),
    persona_name VARCHAR(100),
    persona_desc TEXT,
    perspectives TEXT,           -- JSON数组
    emotional_tone VARCHAR(100),
    suggested_topics TEXT,
    writing_style VARCHAR(100),
    key_points TEXT,
    delivery_delay INT,
    scheduled_for TIMESTAMP,
    provider VARCHAR(50),
    created_at TIMESTAMP,
    used_at TIMESTAMP
);
```

### **8.3 安全机制实现**

| 安全要求 | 设计要求 | 实现情况 | 状态 |
|----------|----------|----------|------|
| 匿名保护 | 仅返回编码 | ✅ 已实现 | 完全符合 |
| 内容预审 | 人工二审 | ⚠️ 基础实现 | 需加强审核能力 |
| AI标记 | 标记AI生成 | ✅ 已实现 | 前端有标识 |
| 延迟机制 | 1-3天延迟 | ⚠️ 模拟延迟 | 需改为真实队列 |
| 情绪拦截 | 极端情绪检测 | ❌ 未实现 | 待开发 |

### **8.4 模型集成状态**

**已集成的AI提供商**：
1. **OpenAI**
   - 模型：GPT-3.5-turbo, GPT-4
   - 用途：全功能支持
   - 状态：✅ 稳定运行

2. **Claude (Anthropic)**
   - 模型：Claude-3-sonnet
   - 用途：高质量对话生成
   - 状态：✅ 已集成

3. **SiliconFlow**
   - 模型：Qwen2.5-7B-Instruct
   - 用途：国产模型备份
   - 状态：✅ 已集成

### **8.5 性能优化实现**

1. **多提供商故障转移**
   ```go
   // 自动切换逻辑已实现
   if provider1.Failed() {
       useProvider2()
   }
   ```

2. **配额管理**
   - 每日配额限制
   - 自动重置机制
   - 用户级别控制

3. **响应缓存**
   - ❌ 待实现
   - 建议使用Redis缓存常见请求

### **8.6 前端集成状态**

**AI功能页面** (`/ai`):
- ✅ 写作灵感模块
- ✅ 云中锦书对话
- ✅ 笔友匹配界面
- ✅ 角色驿站建议
- ⚠️ 策展功能待完善

**组件实现**：
```typescript
// 已实现的核心组件
- AIWritingInspiration    // 写作灵感
- CloudLetterCompanion    // 云中锦书
- AIPenpalMatch          // 笔友匹配
- AIReplyAdvice          // 回信建议
- AIPersonaSelector      // 人设选择器
```

---

## **九、技术债务与优化建议**

### **9.1 待解决的技术债务**

1. **TypeScript类型错误**
   - 161个编译错误待修复
   - 主要集中在UI组件props

2. **延迟机制**
   - 当前使用setTimeout模拟
   - 需实现基于消息队列的真实延迟

3. **缓存策略**
   - AI响应未缓存
   - 重复请求浪费资源

### **9.2 架构优化建议**

1. **引入消息队列**
   ```
   用户请求 → RabbitMQ/Redis → 延迟投递
   ```

2. **实现响应缓存**
   ```
   请求 → Redis缓存检查 → AI调用 → 缓存存储
   ```

3. **增强监控**
   - API调用统计
   - 错误率监控
   - 成本分析报表

### **9.3 功能增强建议**

1. **情感分析增强**
   - 集成专业情感分析模型
   - 实现多维度情绪检测

2. **个性化推荐**
   - 基于用户历史构建画像
   - 实现智能内容推荐

3. **多模态支持**
   - 支持图片理解
   - 语音信件转文字

---

是否接下来输出 FSD + PRD 的整合目录结构文档？或者开始准备 API 网关的标准路由定义？如需输出 OpenAPI（Swagger）结构，也可继续指示。✅