# 快速启动指南

## 🚀 启动方式

### 方式一：图形化启动（推荐）
```bash
# macOS 用户推荐
双击 "启动 OpenPenPal 集成.command"
```

### 方式二：命令行启动
```bash
# 演示模式（最简单）
./startup/quick-start.sh demo --auto-open

# 开发模式（完整微服务）
./startup/quick-start.sh development --auto-open

# 使用 Makefile（统一构建系统）
make dev
```

### 方式三：前端独立启动
```bash
# 进入前端目录
cd frontend
npm install
npm run dev
```

访问：
- 🌐 **前端**: http://localhost:3000 (主要入口)
- 🔗 **API网关**: http://localhost:8000 (微服务模式)
- 🔍 **服务状态**: ./startup/check-status.sh

## 📖 详细文档

- [🎯 统一文档中心](./UNIFIED_DOC_CENTER.md) - 一站式文档入口
- [📚 完整文档导航](./docs/README.md) - 所有文档索引
- [📖 项目总览](./README.md) - 项目详细说明
- [🚀 启动脚本指南](./STARTUP_SCRIPTS.md) - 启动系统详解
- [🔧 故障排查](./docs/troubleshooting/) - 常见问题解决

## 🧪 测试账户

### 快速测试
| 用户名 | 密码 | 角色 | 功能 |
|--------|------|------|------|
| admin | admin123 | 超级管理员 | 全部权限 |
| courier_building | courier123 | 信使 | 信件配送 |
| senior_courier | senior123 | 高级信使 | 信使+报告查看 |
| coordinator | coord123 | 信使协调员 | 信使管理 |

> 📋 完整测试账号说明：[test-accounts.md](./docs/getting-started/test-accounts.md)

## 🔍 主要功能模块

| 模块 | 访问路径 | 功能说明 | 状态 |
|------|----------|----------|------|
| 🏠 首页 | `/` | 项目介绍和导航 | ✅ |
| 🔐 用户系统 | `/login`, `/register` | 登录注册(支持邮箱验证) | ✅ |
| ✍️ 写信系统 | `/write` | 信件创作和发送 | ✅ |
| 🏃‍♂️ 信使中心 | `/courier` | 4级信使管理系统 | ✅ |
| 🏛️ 信件博物馆 | `/museum` | 信件展示和收藏 | ✅ |
| 🎪 Plaza广场 | `/plaza` | 社区交流平台 | ✅ |
| ⚙️ 管理后台 | `/admin` | 系统管理和配置 | ✅ |

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

---

**💡 提示**: 如果是第一次使用，建议按顺序阅读：
1. 此快速指南 → 2. [README.md](./README.md) → 3. [统一文档中心](./UNIFIED_DOC_CENTER.md)