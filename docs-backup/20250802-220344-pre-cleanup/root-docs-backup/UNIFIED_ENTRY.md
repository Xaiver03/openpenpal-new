# OpenPenPal 统一入口文档

<p align="center">
  <strong>让每一封信都有温度 | 校园手写信件平台</strong>
  <br/>
  <em>实体手写信 + 数字跟踪平台，为校园学生打造有温度、有深度的慢节奏社交体验</em>
</p>

<p align="center">
  <a href="#快速启动">🚀 快速启动</a> •
  <a href="#核心功能">⭐ 核心功能</a> •
  <a href="#系统架构">🏗️ 系统架构</a> •
  <a href="#文档导航">📚 文档导航</a> •
  <a href="#开发指南">🛠️ 开发指南</a>
</p>

> 🎉 **最新状态** (2025-07-24): 系统集成验证完成，所有核心功能已实现并测试通过！
> 
> ✅ **演示就绪**: 前后端完整集成，6个微服务正常运行  
> ✅ **功能完整**: 信使系统、博物馆、Plaza等核心模块全部实现  
> ✅ **API验证**: 100% API测试通过，支持完整业务流程

## 🚀 快速启动

### ⚡ 一键启动（推荐）

#### 方式一：图形化启动（最简单）
```bash
# macOS 用户推荐
双击 "启动 OpenPenPal 集成.command"
```

#### 方式二：命令行启动
```bash
# 演示模式（最简单）
./startup/quick-start.sh demo --auto-open

# 开发模式（完整微服务）
./startup/quick-start.sh development --auto-open

# 使用 Makefile（统一构建系统）
make dev
```

#### 方式三：前端独立启动
```bash
cd frontend && npm install && npm run dev
```

### 🌐 访问地址
- **前端应用**: http://localhost:3000 (主要入口)
- **API网关**: http://localhost:8000 (微服务模式)
- **服务状态**: `./startup/check-status.sh`

### 🧪 快速测试账户

| 用户名 | 密码 | 角色 | 功能 |
|--------|------|------|------|
| admin | admin123 | 超级管理员 | 全部权限 |
| courier_building | courier123 | 信使 | 信件配送 |
| senior_courier | senior123 | 高级信使 | 信使+报告查看 |
| coordinator | coord123 | 信使协调员 | 信使管理 |

> 📋 完整测试账号说明：[test-accounts.md](./docs/getting-started/test-accounts.md)

## ⭐ 核心功能

### ✍️ 信件写作系统
- 在线创作，支持图片OCR识别
- 主题信件和个人信件分类
- 信件草稿保存和模板功能

### 📮 4级信使体系
- **城市总代 → 校级 → 片区/年级 → 楼栋/班级** 四级管理架构
- 🏃‍♂️ **智能任务分配**: 基于地理位置和负载均衡
- 📊 **实时状态追踪**: 扫码确认、状态更新、异常处理
- 🏆 **积分成长系统**: 等级晋升、排行榜、成就系统

### 🏛️ 信件博物馆
- 主题展览、时间信轴、用户典藏
- 历史信件展示和互动体验

### 🎪 Plaza广场
- 信件分享、互动交流、社区建设
- 实时聊天和信件推荐功能

### 👥 用户管理
- 📧 **邮箱注册**: 完整的邮箱验证注册流程
- 🔐 **多角色权限**: 支持管理员、信使、高级信使、协调员等角色
- ✅ **权限控制**: 基于RBAC的细粒度权限管理

### 📊 管理后台
- ⚙️ **系统设置**: 可配置站点信息、安全策略等
- 📨 **邮件配置**: 支持Gmail SMTP配置和测试邮件发送
- 👨‍💼 **用户管理**: 用户信息管理和权限分配

## 🏗️ 系统架构

### 📁 目录结构
```
openpenpal/
├── 🎨 frontend/              # Next.js 14 前端应用
├── ⚙️ backend/               # Go 主后端服务
├── 🔧 services/              # 微服务架构
│   ├── write-service/       # Python - 写信服务
│   ├── courier-service/     # Go - 信使服务  
│   ├── admin-service/       # Java - 管理服务
│   ├── ocr-service/         # Python - OCR识别
│   └── gateway/            # Go - API网关
├── 📚 docs/                 # 项目文档中心
├── 🚀 startup/              # 统一启动系统
└── 🧪 test-kimi/           # 测试套件
```

### 🔗 微服务架构
| 服务 | 技术栈 | 端口 | 状态 | 功能描述 |
|------|--------|------|------|----------|
| Frontend | Next.js 14 + TypeScript | 3000 | ✅ 运行中 | 用户界面和交互 |
| API Gateway | Express.js + Proxy | 8000 | ✅ 运行中 | API网关和路由统一入口 |
| Write Service | Express.js + Mock | 8001 | ✅ 运行中 | 信件创作、Plaza、博物馆 |
| Courier Service | Express.js + Mock | 8002 | ✅ 运行中 | 4级信使管理和任务分配 |
| Admin Service | Express.js + Mock | 8003 | ✅ 运行中 | 系统管理和数据统计 |
| OCR Service | Express.js + Mock | 8004 | ✅ 运行中 | 图像识别和扫码功能 |

### 💻 技术栈

**前端技术栈**
- **框架**: Next.js 14 with App Router
- **语言**: TypeScript
- **样式**: Tailwind CSS + shadcn/ui
- **状态管理**: Zustand
- **UI组件**: Radix UI

**后端技术栈**
- **主服务**: Go 1.21 + Gin + GORM
- **写信服务**: Python 3.8 + FastAPI
- **管理服务**: Java 17 + Spring Boot
- **OCR服务**: Python + OpenCV + Tesseract
- **信使服务**: Go + Redis + WebSocket

**数据存储**
- **主数据库**: PostgreSQL 13+
- **缓存**: Redis 6+
- **文件存储**: 本地存储 + 七牛云OSS
- **消息队列**: Redis Pub/Sub

## 📚 文档导航

### 🚀 新手起步
- **[📋 快速启动](./QUICK_START.md)** - 立即启动项目
- **[🎯 5分钟新手指南](./docs/getting-started/5min-guide.md)** - 零基础入门
- **[🧪 测试账号](./docs/getting-started/test-accounts.md)** - 快速测试功能

### 🛠️ 开发者指南
- **[⚙️ 开发环境搭建](./docs/development/README.md)** - 完整开发指南
- **[🚀 启动脚本详解](./STARTUP_SCRIPTS.md)** - 启动系统说明
- **[📝 代码规范](./docs/development/coding-standards.md)** - 编码标准

### 📡 API 文档
- **[📊 API 总览](./docs/api/README.md)** - API 概述和索引
- **[📋 统一规范](./docs/api/unified-specification.md)** - API 设计规范
- **[🧪 接口测试](./docs/api/testing.md)** - API 测试指南

### 🏗️ 系统架构
- **[🎯 整体架构](./docs/architecture/README.md)** - 系统架构概览
- **[🔧 微服务设计](./docs/architecture/microservices.md)** - 服务拆分策略
- **[📊 数据模型](./docs/architecture/data-models.md)** - 数据库设计

### 🧪 测试指南
- **[🧪 测试概览](./docs/testing/README.md)** - 测试策略
- **[📊 覆盖率报告](./docs/testing/coverage.md)** - 测试覆盖率
- **[🔄 集成测试](./test-kimi/README.md)** - 端到端测试

### 🚀 部署运维
- **[🏠 本地部署](./docs/deployment/local.md)** - 本地环境部署
- **[☁️ 云端部署](./docs/deployment/cloud.md)** - 云服务部署
- **[🔧 运维手册](./docs/operations/README.md)** - 运维操作指南

### 📚 产品文档
- **[📋 产品需求](./docs/product/README.md)** - PRD 文档中心
- **[🏃‍♂️ 信使系统](./docs/product/prd/penpal-messenger-system-prd.md)** - 信使业务流程
- **[🏛️ 博物馆模块](./docs/product/prd/letter-museum-module-prd.md)** - 展览展示功能

### 🤝 团队协作
- **[👥 团队规范](./docs/team-collaboration/README.md)** - 协作指南
- **[🔄 多人协作](./docs/team-collaboration/MULTI_AGENT_SYNC_SYSTEM.md)** - 并行开发
- **[📋 任务管理](./docs/project/README.md)** - 项目管理流程

## 🛠️ 开发指南

### 🔧 环境要求
- **Node.js**: 18+
- **Go**: 1.21+
- **Python**: 3.8+
- **Java**: 17+
- **PostgreSQL**: 13+
- **Redis**: 6+

### 📝 开发流程
1. **环境准备**: 安装 Node.js 18+ (其他依赖可选)
2. **快速启动**: `./startup/quick-start.sh demo --auto-open`
3. **依赖管理**: `./startup/install-deps.sh` (自动检测和安装)
4. **服务监控**: `./startup/check-status.sh --continuous`
5. **问题排查**: 查看 [LAUNCH_FIX_SUMMARY.md](./LAUNCH_FIX_SUMMARY.md)

### 🧪 测试
```bash
# 测试启动配置 (不实际启动)
./startup/quick-start.sh demo --dry-run

# 验证服务状态
./startup/check-status.sh --detailed

# 运行完整测试套件
./test-kimi/run_tests.sh

# API集成测试
./scripts/test-apis.sh
```

## 🔍 快速查找

### 按功能查找
- **用户系统**: [API文档](./docs/api/auth.md) | [测试账号](./docs/getting-started/test-accounts.md)
- **信使系统**: [业务文档](./docs/product/prd/penpal-messenger-system-prd.md) | [API接口](./services/courier-service/README.md)
- **写信功能**: [产品设计](./docs/product/write-features.md) | [API接口](./services/write-service/README.md)
- **管理后台**: [管理指南](./docs/admin/README.md) | [权限说明](./docs/admin/permissions.md)

### 按角色查找
- **🆕 新用户**: [5分钟指南](./docs/getting-started/5min-guide.md) → [功能概览](#核心功能)
- **👨‍💻 开发者**: [开发指南](./docs/development/README.md) → [API文档](./docs/api/README.md)
- **🚀 部署者**: [部署指南](./docs/deployment/README.md) → [运维手册](./docs/operations/README.md)
- **🧪 测试者**: [测试指南](./docs/testing/README.md) → [测试账号](./docs/getting-started/test-accounts.md)

## ❓ 遇到问题？

### 🚨 常见启动问题
1. **端口占用**: 运行 `./startup/force-cleanup.sh` 清理
2. **权限问题**: 确保脚本有执行权限 `chmod +x startup/*.sh`
3. **依赖问题**: 运行 `./startup/install-deps.sh` 安装依赖
4. **服务状态**: 运行 `./startup/check-status.sh` 检查服务

### 📞 获取帮助
- 🔍 **检查服务状态**: `make status`
- 📋 **查看日志**: `tail -f logs/*.log`
- 🧹 **强制清理**: `make clean-all`
- 📚 **详细故障排查**: [LAUNCH_FIX_SUMMARY.md](./LAUNCH_FIX_SUMMARY.md)

### 问题反馈
- **🐛 报告 Bug**: [Bug 报告模板](./.github/ISSUE_TEMPLATE/bug_report.md)
- **✨ 功能建议**: [功能请求模板](./.github/ISSUE_TEMPLATE/feature_request.md)
- **❓ 使用问题**: [GitHub Discussions](https://github.com/openpenpal/discussions)

## 📊 项目状态

### 最新验证结果
- **系统集成**: ✅ 6个微服务正常运行
- **API测试**: ✅ 100% 接口测试通过
- **权限系统**: ✅ 4级信使权限正确实现
- **用户注册**: ✅ 完整邮箱验证流程
- **管理后台**: ✅ 系统配置和邮件发送功能

### 文档统计
- **总文档数**: 55+ 个文档文件
- **覆盖模块**: 8 个主要功能模块
- **维护状态**: ✅ 持续更新中
- **最后更新**: 2025-07-24

---

<div align="center">

**⭐ 如果 OpenPenPal 对你有帮助，请给我们一个 Star！**

**让每一封信都成为连接心灵的桥梁** ✨

</div>

---

**💡 使用提示**: 
- 使用 Ctrl/Cmd + F 快速搜索关键词
- 建议按照"快速启动 → 核心功能 → 开发指南"的顺序阅读
- 文档问题请提交 [Issue](https://github.com/openpenpal/issues)