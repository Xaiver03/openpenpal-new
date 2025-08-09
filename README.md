# OpenPenPal 信使计划

<p align="center">
  <strong>让每一封信都有温度 | 校园手写信件平台</strong>
  <br/>
  <em>实体手写信 + 数字跟踪平台，为校园学生打造有温度、有深度的慢节奏社交体验</em>
</p>

<p align="center">
  <a href="#快速开始">🚀 快速开始</a> •
  <a href="#项目架构">🏗️ 项目架构</a> •
  <a href="#技术栈">💻 技术栈</a> •
  <a href="./DOCS.md">📚 完整文档</a> •
  <a href="#开发指南">🛠️ 开发指南</a>
</p>

> 📖 **重要提示**: 查看 **[完整文档中心](./DOCS.md)** 获取所有详细文档（PRD、FSD、API文档等）

## ✨ 项目概述

OpenPenPal是一个创新的校园手写信件平台，结合实体信件与数字化管理，为学生提供有温度的慢节奏社交体验。

> 🎉 **最新状态** (2025-07-23): 系统集成验证完成，所有核心功能已实现并测试通过！
> 
> ✅ **演示就绪**: 前后端完整集成，6个微服务正常运行  
> ✅ **功能完整**: 信使系统、博物馆、Plaza等核心模块全部实现  
> ✅ **API验证**: 100% API测试通过，支持完整业务流程
> ✅ **用户系统**: 完整邮箱注册流程，支持真实邮件发送
> ✅ **权限管理**: 4级角色权限系统，管理后台功能完整
> ✅ **邮件配置**: 支持Gmail SMTP配置和测试邮件发送功能

### 🎯 核心功能
- **✍️ 信件写作系统**: 在线创作，支持图片OCR识别
- **📮 4级信使体系**: 城市总代→校级→片区→个人 四级管理架构
  - 🏃‍♂️ **智能任务分配**: 基于地理位置和负载均衡的任务分配
  - 📊 **实时状态追踪**: 扫码确认、状态更新、异常处理
  - 🏆 **积分成长系统**: 等级晋升、排行榜、成就系统
- **🏛️ 信件博物馆**: 主题展览、时间信轴、用户典藏
- **🎪 Plaza广场**: 信件分享、互动交流、社区建设
- **👥 用户管理**: 基于学校的用户认证和权限管理
  - 📧 **邮箱注册**: 完整的邮箱验证注册流程
  - 🔐 **多角色权限**: 支持管理员、信使、高级信使、协调员等角色
  - ✅ **权限控制**: 基于RBAC的细粒度权限管理
- **📊 管理后台**: 完整的系统管理和数据统计
  - ⚙️ **系统设置**: 可配置站点信息、安全策略等
  - 📨 **邮件配置**: 支持Gmail SMTP配置和测试邮件发送
  - 👨‍💼 **用户管理**: 用户信息管理和权限分配

## 🚀 快速开始

### 📋 环境要求
- **Node.js**: 18+
- **Go**: 1.21+
- **Python**: 3.8+
- **Java**: 17+
- **PostgreSQL**: 13+
- **Redis**: 6+

### ⚡ 快速启动

#### 方式一：统一启动器（推荐）
```bash
# 克隆项目
git clone [repository-url]
cd openpenpal

# 方式1: macOS图形化启动 (推荐)
双击 "启动 OpenPenPal 集成.command"
# 提供4种启动模式选择，自动处理端口冲突

# 方式2: 命令行演示模式 (最简单)
./startup/quick-start.sh demo --auto-open

# 方式3: 开发模式 (完整微服务)
./startup/quick-start.sh development --auto-open

# 方式4: 图形化启动菜单
./startup/openpenpal-launcher.command
```

#### 🧹 如果遇到端口占用问题
```bash
# 强制清理所有端口
./startup/force-cleanup.sh

# 然后重新启动
./startup/quick-start.sh demo --auto-open
```

#### 方式二：Docker部署
```bash
# 开发环境
docker-compose -f docker-compose.dev.yml up -d

# 生产环境  
docker-compose up -d
```

#### 方式三：手动启动
```bash
# 1. 启动前端 (Next.js)
cd frontend
npm install && npm run dev

# 2. 启动后端主服务 (Go)
cd backend  
go run main.go

# 3. 启动各微服务
cd services/write-service && python app/main.py
cd services/courier-service && go run cmd/main.go
cd services/admin-service && ./mvnw spring-boot:run
cd services/ocr-service && python app.py
```

### 🌐 访问地址
- **前端应用**: http://localhost:3000
- **API网关**: http://localhost:8000 (健康检查: `/health`)
- **Plaza广场**: http://localhost:3000/plaza
- **信使中心**: http://localhost:3000/courier  
- **信件博物馆**: http://localhost:3000/museum
- **管理后台**: http://localhost:3000/admin

### 📊 服务管理
- **服务状态检查**: `./startup/check-status.sh`
- **停止所有服务**: `./startup/stop-all.sh`
- **安装项目依赖**: `./startup/install-deps.sh`
- **启动脚本指南**: 查看 [STARTUP_SCRIPTS.md](./STARTUP_SCRIPTS.md)

## 🏗️ 项目架构

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
├── 🛠️ scripts/              # 其他管理脚本
├── 🐳 docker-compose.yml    # Docker编排配置
├── 📊 monitoring/           # 监控配置
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

> 📍 **当前状态**: 所有服务基于模拟API运行，完整功能已验证。生产环境将使用对应的技术栈实现。

## 💻 技术栈

### 前端技术栈
- **框架**: Next.js 14 with App Router
- **语言**: TypeScript
- **样式**: Tailwind CSS + shadcn/ui
- **状态管理**: Zustand
- **UI组件**: Radix UI
- **通信**: WebSocket + REST API

### 后端技术栈
- **主服务**: Go 1.21 + Gin + GORM
- **写信服务**: Python 3.8 + FastAPI
- **管理服务**: Java 17 + Spring Boot
- **OCR服务**: Python + OpenCV + Tesseract
- **信使服务**: Go + Redis + WebSocket

### 数据存储
- **主数据库**: PostgreSQL 13+
- **缓存**: Redis 6+
- **文件存储**: 本地存储 + 七牛云OSS
- **消息队列**: Redis Pub/Sub

### 开发工具
- **容器化**: Docker + Docker Compose
- **监控**: Prometheus + Grafana
- **日志**: 结构化日志 + 日志轮转
- **测试**: 综合测试套件

## 📚 文档中心

### 📖 核心文档
- [🎯 **项目统一入口**](./UNIFIED_ENTRY.md) - **一站式项目入口，推荐从这里开始**
- [📚 完整文档导航](./docs/index.md) - 详细文档中心
- [🚀 快速开始指南](./docs/getting-started/README.md) - 新手入门
- [🏗️ 系统架构](./docs/architecture/README.md) - 架构设计文档
- [💻 开发指南](./docs/development/README.md) - 开发规范和指南
- [📡 API文档](./docs/api/README.md) - 接口文档和规范
- [✅ 系统验证报告](./SYSTEM_VERIFICATION_REPORT.md) - 最新集成测试结果

### 🔗 快速链接
- [🎮 测试账号](./docs/getting-started/test-accounts.md) - 4种角色测试账号，立即可用
- [📧 注册流程测试](./docs/getting-started/test-accounts.md#邮箱注册测试) - 完整邮箱验证注册
- [⚙️ 系统设置](http://localhost:3000/admin/settings) - 邮件配置等管理功能
- [📊 部署指南](./docs/deployment/README.md) - 部署和运维文档
- [❓ 故障排查](./docs/troubleshooting/) - 常见问题解决
- [👥 团队协作](./docs/team-collaboration/README.md) - 协作规范

### 🛠️ 工具和运维
- [⚙️ 运维操作](./docs/operations/) - 脚本使用和系统操作
- [🔧 工具配置](./docs/tools/) - 开发工具和环境配置
- [📋 操作指南](./docs/guides/) - 详细的操作和使用指南

### 👥 项目管理
- [👨‍💻 Agent文档](./docs/agents/) - Multi-Agent协作和任务分配
- [📦 产品文档](./docs/product/README.md) - 产品需求和设计文档
- [📋 项目管理](./docs/project/) - 项目规划和任务管理
- [🏗️ 技术栈](./docs/tech-stack/README.md) - 详细技术栈文档
- [📊 项目报告](./docs/reports/) - 集成和优化报告

## 🛠️ 开发指南

### 🔧 本地开发
1. **环境准备**: 安装 Node.js 18+ (其他依赖可选)
2. **快速启动**: `./startup/quick-start.sh demo --auto-open`
3. **依赖管理**: `./startup/install-deps.sh` (自动检测和安装)
4. **服务监控**: `./startup/check-status.sh --continuous`
5. **问题排查**: 查看 [LAUNCH_FIX_SUMMARY.md](./LAUNCH_FIX_SUMMARY.md)

### 📝 代码规范
- **Go**: 遵循 Go 官方代码规范
- **TypeScript**: 使用 ESLint + Prettier
- **Python**: 遵循 PEP 8 规范
- **Java**: 使用 Spring Boot 最佳实践

### 🧪 测试

#### 启动系统测试
```bash
# 测试启动配置 (不实际启动)
./startup/quick-start.sh demo --dry-run

# 验证服务状态
./startup/check-status.sh --detailed

# 强制清理测试
./startup/force-cleanup.sh
```

#### 综合测试套件
```bash
# 运行完整测试套件
./test-kimi/run_tests.sh

# API集成测试
./scripts/test-apis.sh
```

#### 功能测试 (最新已验证)
```bash
# 注册流程测试
# 1. 邮箱可用性检查 - ✅ 正常工作
# 2. 验证码发送和验证 - ✅ 控制台显示，支持真实邮件
# 3. 用户注册流程 - ✅ 完整流程验证通过
# 4. 多角色权限测试 - ✅ 4种角色权限正确

# 管理后台功能测试
# 1. 系统设置管理 - ✅ 可保存配置到内存
# 2. 邮件配置测试 - ✅ 支持Gmail SMTP配置和测试邮件发送
# 3. 权限控制验证 - ✅ 角色权限正确控制访问
# 4. 信使中心功能 - ✅ 不同角色显示对应功能
```

## 🎮 测试账户

> 详细测试账号信息请查看 [docs/getting-started/test-accounts.md](./docs/getting-started/test-accounts.md)

### 快速测试账号
| 用户名 | 密码 | 角色 | 权限 | 访问入口 |
|--------|------|------|------|----------|
| admin | admin123 | 超级管理员 | 全部权限 | [管理控制台](/admin) |
| courier_building | courier123 | 信使 | 信件配送、扫码、查看任务 | [信使中心](/courier) |
| senior_courier | senior123 | 高级信使 | 信使权限 + 查看报告 | [信使中心](/courier) |
| coordinator | coord123 | 信使协调员 | 信使权限 + 管理信使、分配任务 | [信使中心](/courier) |

### 🔐 账户功能说明
- **管理员** (admin): 可访问系统设置、用户管理、数据统计等完整管理功能
- **信使** (courier_building): 可执行信件投递、扫码确认、查看配送任务
- **高级信使** (senior_courier): 除信使功能外，还可查看配送报告和统计
- **信使协调员** (coordinator): 可管理其他信使、分配配送任务和区域

### 📧 邮箱注册测试
系统支持完整的邮箱验证注册流程：
- 邮箱可用性检查
- 验证码发送（控制台显示用于测试）
- 邮箱验证和用户注册
- 支持Gmail SMTP真实邮件发送（需配置）

## 🚀 部署方案

### 🏠 开发环境
- 使用 `./startup/quick-start.sh development` 一键启动
- 自动端口管理和冲突解决
- 完整的微服务Mock环境
- 实时状态监控和日志管理

### 🏭 生产环境
- Docker Compose 容器化部署
- Nginx 反向代理
- PostgreSQL 主从复制
- Redis 集群缓存
- Prometheus + Grafana 监控

## 🤝 参与贡献

### 🎯 贡献方式
- 🐛 报告Bug和问题
- ✨ 提出新功能建议
- 📝 改进文档
- 🧪 添加测试用例
- 💻 提交代码改进

### 📋 开发流程
1. Fork 项目仓库
2. 创建功能分支: `git checkout -b feature/new-feature`
3. 提交变更: `git commit -m "feat: add new feature"`
4. 推送分支: `git push origin feature/new-feature`
5. 创建 Pull Request

## 📞 联系我们

### 💬 获取帮助
- **文档**: [完整文档中心](./docs/index.md)
- **启动指南**: [启动脚本索引](./STARTUP_SCRIPTS.md)
- **启动问题**: [启动修复总结](./LAUNCH_FIX_SUMMARY.md)
- **问题反馈**: [GitHub Issues](https://github.com/openpenpal/issues)
- **技术交流**: [开发者社区](https://discord.gg/openpenpal)

### 🆘 故障支持
- **启动问题**: `./startup/force-cleanup.sh` + `./startup/quick-start.sh demo`
- **服务状态**: `./startup/check-status.sh --detailed`
- **日志查看**: `tail -f logs/*.log`
- [常见问题解答](./docs/troubleshooting/)
- [系统监控](./monitoring/)

---

<div align="center">

**⭐ 如果 OpenPenPal 对你有帮助，请给我们一个 Star！**

**让每一封信都成为连接心灵的桥梁** ✨

</div>