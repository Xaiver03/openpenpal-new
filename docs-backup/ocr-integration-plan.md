# 手写信OCR功能集成计划

## 当前状态分析

### 已有基础设施 ✅
1. **OCR服务已部署**
   - 位置：`/services/ocr-service/`
   - 端口：8004
   - 技术：Python + Tesseract/PaddleOCR

2. **图片上传组件完善**
   - 博物馆贡献页面已实现
   - 支持多图上传、预览、删除

3. **后端存储能力**
   - 文件上传API已实现
   - 图片存储路径配置完成

### 缺失功能 ❌
1. **主写信流程未集成OCR**
   - 写信页面无上传手写信入口
   - OCR服务未与写信API对接

2. **用户体验不完整**
   - 缺少OCR识别结果编辑界面
   - 无法保存原始手写信图片与文字对照

---

## 实现方案

### 第一阶段：基础集成（2天）

#### 1.1 写信页面增加上传入口
```typescript
// 位置：/frontend/src/app/(main)/write/page.tsx
// 新增"上传手写信"标签页
interface WritePageTabs {
  'compose': '在线编写',
  'upload': '上传手写信',  // 新增
  'template': '使用模板',
  'ai': 'AI助手'
}
```

#### 1.2 创建手写信上传组件
```typescript
// 新建：/frontend/src/components/write/handwritten-upload.tsx
interface HandwrittenUploadProps {
  onTextExtracted: (text: string) => void
  onImageUploaded: (url: string) => void
}

// 功能包括：
// - 图片上传（支持多张）
// - 图片预览与编辑
// - OCR处理进度显示
// - 识别结果预览
```

#### 1.3 集成OCR服务API
```typescript
// 新建：/frontend/src/lib/api/ocr.ts
export const ocrApi = {
  // 上传图片并识别
  async recognizeHandwriting(image: File): Promise<{
    text: string
    confidence: number
    segments: TextSegment[]
  }> {
    const formData = new FormData()
    formData.append('image', image)
    return apiClient.post('/ocr/recognize', formData)
  },
  
  // 批量识别
  async batchRecognize(images: File[]): Promise<RecognitionResult[]> {
    // ...
  }
}
```

### 第二阶段：体验优化（3天）

#### 2.1 OCR结果编辑器
```typescript
// 新建：/frontend/src/components/write/ocr-editor.tsx
interface OCREditorProps {
  originalImage: string
  recognizedText: string
  segments: TextSegment[]
  onConfirm: (editedText: string) => void
}

// 功能：
// - 原图与识别文本对照显示
// - 分段编辑能力
// - 错误标注与修正
// - 一键应用到编辑器
```

#### 2.2 手写信存档功能
```typescript
// 扩展信件模型
interface Letter {
  // ... 现有字段
  handwrittenImages?: string[]      // 手写信原图
  ocrText?: string                  // OCR识别原文
  ocrConfidence?: number            // 识别置信度
  preserveHandwritten: boolean      // 是否保留手写版
}
```

#### 2.3 后端API扩展
```go
// 扩展：/backend/internal/handlers/letter.go
// 新增手写信相关接口
func (h *LetterHandler) UploadHandwritten(c *gin.Context) {
    // 1. 接收图片文件
    // 2. 调用OCR服务
    // 3. 返回识别结果
    // 4. 保存原图关联
}
```

### 第三阶段：高级功能（2天）

#### 3.1 OCR优化
- 针对手写体优化识别模型
- 支持多语言识别（中英文混合）
- 智能纠错功能
- 用户反馈训练

#### 3.2 批量处理
- 多页信纸批量上传
- 后台异步处理
- 进度实时推送
- 结果批量下载

#### 3.3 历史信件数字化
- 为已有纸质信件提供数字化入口
- 批量导入功能
- 自动分类整理

---

## 技术细节

### OCR服务优化
```python
# /services/ocr-service/main.py 优化建议
class HandwritingOCR:
    def __init__(self):
        # 使用 PaddleOCR 提升中文识别率
        self.ocr = PaddleOCR(
            use_angle_cls=True,
            lang='ch',
            det_model_dir='./models/ch_det',
            rec_model_dir='./models/ch_rec'
        )
    
    def preprocess_image(self, image):
        """图片预处理：去噪、增强对比度"""
        # 灰度化
        # 二值化
        # 去噪
        # 倾斜矫正
        pass
    
    def recognize(self, image_path):
        """识别手写文字"""
        # 预处理
        # 文字检测
        # 文字识别
        # 后处理
        pass
```

### 前端交互流程
```mermaid
graph LR
    A[选择上传手写信] --> B[上传图片]
    B --> C[显示上传进度]
    C --> D[调用OCR服务]
    D --> E[显示识别结果]
    E --> F[用户编辑校对]
    F --> G[确认导入编辑器]
    G --> H[保存信件]
```

### 数据库更新
```sql
-- 新增手写信相关字段
ALTER TABLE letters ADD COLUMN handwritten_images TEXT[];
ALTER TABLE letters ADD COLUMN ocr_text TEXT;
ALTER TABLE letters ADD COLUMN ocr_confidence FLOAT;
ALTER TABLE letters ADD COLUMN preserve_handwritten BOOLEAN DEFAULT false;

-- 创建手写信识别记录表
CREATE TABLE ocr_records (
    id SERIAL PRIMARY KEY,
    letter_id VARCHAR(255) REFERENCES letters(id),
    image_url TEXT NOT NULL,
    recognized_text TEXT,
    confidence FLOAT,
    segments JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 实施计划

### 第1周
- [ ] 完成写信页面UI改造
- [ ] 实现基础上传组件
- [ ] 对接OCR服务API
- [ ] 基础功能测试

### 第2周  
- [ ] 开发OCR编辑器
- [ ] 完善用户体验
- [ ] 性能优化
- [ ] 集成测试

### 第3周
- [ ] 部署上线
- [ ] 用户培训
- [ ] 收集反馈
- [ ] 迭代优化

---

## 预期效果

### 用户价值
1. **便捷性提升** - 直接上传手写信，无需手动录入
2. **情感保留** - 保存手写原稿，留存书写温度
3. **效率提升** - OCR快速转文字，方便编辑分享

### 业务价值
1. **用户粘性** - 独特功能提升平台吸引力
2. **内容丰富** - 更多优质手写信内容
3. **技术领先** - 行业内首创手写信数字化

### 技术指标
- OCR识别准确率 > 95%（印刷体）
- OCR识别准确率 > 85%（手写体）
- 单张图片处理时间 < 3秒
- 支持图片格式：JPG/PNG/HEIC
- 最大图片大小：10MB

---

## 风险与对策

### 技术风险
- **风险**：手写体识别准确率低
- **对策**：提供便捷的编辑工具，支持用户快速修正

### 性能风险
- **风险**：大量图片上传造成服务器压力
- **对策**：使用队列异步处理，CDN分发，限流控制

### 用户体验风险
- **风险**：操作流程复杂，用户不会使用
- **对策**：提供新手引导，视频教程，简化操作步骤

---

## 总结

手写信OCR功能是OpenPenPal平台的重要补充，将传统手写信与数字化完美结合。通过分阶段实施，可以快速上线基础功能，逐步优化体验，最终打造行业领先的手写信数字化解决方案。

**预计总工期**：7个工作日
**所需资源**：前端1人、后端1人、测试0.5人
**优先级**：高