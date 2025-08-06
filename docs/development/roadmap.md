# OpenPenPal 信使计划 - 开发计划

## 🎯 项目目标

基于产品需求文档，开发一个以"实体手写信 + 数字跟踪平台"为核心的校园慢社交平台，实现完整的写信→投递→收信→回信闭环体验。

## 🚀 即时开发计划 (下一步行动)

### 第一优先级：Go环境配置和后端服务启动
1. **安装Go环境** (预计15分钟)
   ```bash
   # macOS 安装 Go
   brew install go
   # 或者下载官方安装包
   ```

2. **测试Go后端服务** (预计20分钟)
   ```bash
   cd backend
   go mod download    # 下载依赖
   cp .env.example .env  # 创建环境配置
   go run main.go     # 启动服务
   ```

3. **验证前后端连接** (预计15分钟)
   - 测试后端健康检查: `http://localhost:8080/health`
   - 测试前端API调用
   - 验证用户注册登录流程

### 第二优先级：核心功能验证 (预计1-2小时)
1. **写信功能完整测试**
   - 编写信件内容
   - 生成编号和二维码
   - 验证数据保存

2. **信使扫码功能测试**
   - 验证二维码扫描
   - 测试状态更新API

3. **信件查看功能测试**
   - 通过编号查看信件
   - 验证权限控制

### 第三优先级：问题修复和优化 (预计2-3小时)
1. **修复前端构建警告**
2. **优化用户体验**
3. **完善错误处理**

### 预期成果
- ✅ 完整的开发环境配置
- ✅ 前后端服务正常运行
- ✅ 核心业务流程验证通过
- ✅ 为后续开发奠定基础

## 📋 开发阶段规划

### 🎯 当前状态评估（2024年1月）

#### ✅ 已完成
- [x] 项目脚手架搭建（Go + Next.js架构）
- [x] UI组件库集成（ShadCN/UI + TailwindCSS）
- [x] 基础路由结构设计
- [x] 认证系统架构（AuthProvider + JWT）
- [x] 基础页面开发（写信、登录、注册等）
- [x] Docker容器化配置
- [x] 启动脚本配置
- [x] 文档维护规范建立

#### 🔧 需要修复的问题
- [ ] Go环境安装和配置
- [ ] 前后端API连接测试
- [ ] 数据库初始化
- [ ] 图片上传和文件服务
- [ ] metadata viewport警告修复

### 阶段一：环境修复和核心功能完善（2-3周）

#### Week 1: 环境修复和基础设施
- [ ] Go环境安装配置
- [ ] 后端服务启动测试
- [ ] 数据库连接和初始化
- [ ] 前后端API联调
- [ ] 文件上传服务配置
- [ ] 修复前端构建警告

#### Week 2: 核心功能完善和测试
**写信模块优化 (`/write`)**
- [x] 基础写信页面（已完成）
- [ ] 编号生成API连接测试
- [ ] 二维码生成功能测试
- [ ] 草稿保存与状态管理优化
- [ ] 用户体验优化

**其他核心页面测试**
- [ ] 信使扫码页面功能测试
- [ ] 信件投递页面完善
- [ ] 信箱页面数据展示
- [ ] 用户认证流程测试

#### Week 3: 核心业务流程打通
**信件生命周期完整测试**
- [ ] 写信→生成编号→扫码→状态更新完整流程
- [ ] 信件查看功能测试
- [ ] 信使系统功能验证
- [ ] 用户体验优化和错误处理

### 阶段二：功能增强和优化（3-4周）

#### Week 7-8: 信使端基础功能
**信使任务系统 (`/courier/*`)**
- [ ] 信使注册与认证
- [ ] 任务列表界面
- [ ] 扫码录入功能
- [ ] 状态更新系统

#### Week 9-10: 信使管理与激励
- [ ] 信使绩效统计
- [ ] 积分排行榜
- [ ] 任务地图可视化
- [ ] 信使等级系统

### 阶段三：高级功能开发（3-4周）

#### Week 11-12: 核心API开发
- [ ] 用户认证系统（JWT + 微信OAuth）
- [ ] 编号生成与管理API
- [ ] 信件状态追踪API
- [ ] 文件上传服务

#### Week 13-14: 数据库设计与优化
- [ ] PostgreSQL数据库设计
- [ ] Prisma ORM配置
- [ ] 数据迁移脚本
- [ ] 索引优化

#### Week 15: 第三方服务集成
- [ ] 文件存储服务（OSS/COS）
- [ ] 内容安全审核API
- [ ] 推送通知服务

### 阶段四：管理后台开发（2-3周）

#### Week 16-17: 后台管理系统
- [ ] 管理员认证系统
- [ ] 编号管理界面
- [ ] 用户/信使管理
- [ ] 内容审核系统
- [ ] 数据统计看板

#### Week 18: 运营工具完善
- [ ] 信件追踪工具
- [ ] 异常处理机制
- [ ] 客服支持系统

### 阶段五：测试与优化（2-3周）

#### Week 19-20: 全面测试
- [ ] 单元测试编写
- [ ] 集成测试
- [ ] 端到端测试
- [ ] 性能测试
- [ ] 安全性测试

#### Week 21: 上线准备
- [ ] 生产环境部署
- [ ] 监控系统配置
- [ ] 错误日志收集
- [ ] 用户文档完善

## 🛠️ 当前技术架构

### 后端技术栈 (Go)
```go
// 主要依赖 (go.mod)
require (
    github.com/gin-gonic/gin v1.9.1           // Web框架
    github.com/gin-contrib/cors v1.4.0        // CORS中间件
    gorm.io/gorm v1.25.5                      // ORM
    gorm.io/driver/sqlite v1.5.4              // SQLite驱动
    gorm.io/driver/postgres v1.5.4            // PostgreSQL驱动
    github.com/golang-jwt/jwt/v5 v5.1.0       // JWT认证
    github.com/skip2/go-qrcode v0.0.0         // 二维码生成
    golang.org/x/crypto v0.14.0               // 密码加密
)
```

### 前端技术栈 (Next.js)
```typescript
// 当前架构 (package.json)
"dependencies": {
  "next": "14.2.30",                      // React框架
  "react": "^18.2.0",                     // UI库
  "@radix-ui/*": "latest",                // UI组件基础
  "tailwindcss": "^3.4.0",               // CSS框架
  "lucide-react": "^0.263.1",             // 图标库
  "zustand": "^4.4.6"                     // 状态管理
}
```

### 状态管理架构

```typescript
// stores/letterStore.ts
interface LetterStore {
  // 写信状态
  currentDraft: LetterDraft | null;
  savedDrafts: LetterDraft[];
  
  // 编号管理
  generatedCodes: LetterCode[];
  
  // 信件状态
  sentLetters: SentLetter[];
  receivedLetters: ReceivedLetter[];
  
  // 操作方法
  saveDraft: (draft: LetterDraft) => void;
  generateCode: () => Promise<LetterCode>;
  updateLetterStatus: (codeId: string, status: LetterStatus) => void;
}
```

### 组件库设计

```typescript
// 基础UI组件（基于ShadCN/UI）
- Button, Input, Card, Dialog
- Form, Select, Tabs, Badge
- Toast, Loading, Modal

// 业务组件
- LetterEditor         // 富文本编辑器
- LetterPaper         // 信纸样式选择
- QRCodeGenerator     // 二维码生成
- LetterPreview       // 信件预览
- CourierScanner      // 信使扫码
- StatusTimeline      // 状态时间线
```

### 路由结构

```
/                     # 首页
/write               # 写信页面
/deliver             # 投递页面
/read/[code]         # 收信页面
/mailbox             # 我的信箱
/courier
  ├── /tasks         # 信使任务
  ├── /scan          # 扫码录入
  └── /history       # 投递记录
/admin
  ├── /dashboard     # 管理面板
  ├── /letters       # 信件管理
  ├── /users         # 用户管理
  └── /couriers      # 信使管理
```

## 📊 开发里程碑（更新版）

| 里程碑 | 目标日期 | 关键交付物 | 状态 |
|--------|----------|------------|------|
| M0 | 已完成 | 基础架构搭建，Go+Next.js分离架构 | ✅ |
| M1 | Week 1 | 环境配置完成，前后端连接成功 | 🔄 |
| M2 | Week 2 | 核心功能测试通过，主要业务流程验证 | 📅 |
| M3 | Week 3 | 完整信件生命周期流程打通 | 📅 |
| M4 | Week 6 | 功能增强，用户体验优化 | 📅 |
| M5 | Week 10 | 高级功能开发，管理后台 | 📅 |
| M6 | Week 12 | 测试优化，准备部署 | 📅 |

## 🔧 开发环境配置

### 必需工具
- **Go 1.21+** ⚠️ *需要安装*
- **Node.js 18+** ✅ *已安装*
- **npm** ✅ *已配置*
- **VS Code + 相关插件** ✅
- **Git** ✅
- **Docker** ✅ *已配置*

### 当前环境状态
- ✅ Next.js 前端环境已配置
- ✅ TypeScript 配置完成
- ✅ ShadCN/UI 组件库集成
- ✅ 启动脚本已配置
- ⚠️ Go 后端需要安装配置
- ⚠️ 数据库需要初始化

### 推荐插件
- TypeScript Importer
- Tailwind CSS IntelliSense
- ESLint
- Prettier
- Auto Rename Tag

### 代码规范
- TypeScript严格模式
- ESLint + Prettier格式化
- 组件命名采用PascalCase
- 文件命名采用kebab-case
- 提交信息遵循Conventional Commits

## 🚀 部署策略

### 开发环境
- Vercel/Netlify自动部署
- 每次push自动触发构建

### 生产环境
- Docker容器化部署
- CI/CD自动化流水线
- 监控和日志收集

## 📝 文档计划

- [ ] API接口文档
- [ ] 组件库文档
- [ ] 部署运维文档
- [ ] 用户使用指南
- [ ] 开发者贡献指南

## 🎯 成功指标

### 技术指标
- 页面加载时间 < 2秒
- 代码测试覆盖率 > 80%
- TypeScript严格模式无错误
- Lighthouse性能评分 > 90

### 功能指标
- 写信转化率 ≥ 30%
- 实体信投递率 ≥ 60%
- 回信参与率 ≥ 20%
- 信使留存率 ≥ 50%

---

*本开发计划将根据实际进度和需求变化进行调整*