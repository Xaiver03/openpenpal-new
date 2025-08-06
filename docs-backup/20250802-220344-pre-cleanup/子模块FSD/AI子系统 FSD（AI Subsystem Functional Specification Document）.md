
> 面向 LLM API 调用设计，强调解耦合与功能注入式结构。

---

## **一、模块定位**

OpenPenPal 的 AI 子系统通过接入大模型 API，为写信体验提供辅助与增强功能。它不参与核心社交关系构建，而是作为 “温柔、慢节奏、可控”的智能助手，服务于匿名匹配、笔友模拟、写作引导与信件策展等场景。

---

## **二、系统结构概览**

```
flowchart TD
  User -->|写匿名信| AI_Matcher -->|返回目标编码| Barcode_Module
  User -->|开启虚拟笔友| AI_Reply -->|周期性生成信件| Barcode_Module
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

### **3.2 AI 笔友模拟器（Virtual Penpal）**

**用途**：用户自选一个笔友人格，开启长期通信，AI 每隔 1–3 天自动回信。

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

### **3.4 信件策展引擎（Letter Curation）**

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
|虚拟笔友回信|gpt-4, claude, moonshot 等|多轮历史记忆支持|
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

是否接下来输出 FSD + PRD 的整合目录结构文档？或者开始准备 API 网关的标准路由定义？如需输出 OpenAPI（Swagger）结构，也可继续指示。✅