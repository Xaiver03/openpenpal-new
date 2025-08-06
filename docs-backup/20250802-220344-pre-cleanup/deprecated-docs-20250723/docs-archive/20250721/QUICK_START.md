# 快速启动指南

## 🚀 启动方式

### 首次启动或完整重启
```bash
# 同时启动 Go 后端 + Next.js 前端
./start-all.command
```

### 只启动前端（后端已运行时）
```bash
# 只启动 Next.js 前端
./start-frontend.command
```

### 只启动后端
```bash
# 只启动 Go 后端
./start-backend.command
```

访问：
- 🌐 **前端**: http://localhost:3000
- 🔗 **API**: http://localhost:8080
- 🔍 **健康检查**: http://localhost:8080/health

## 📖 详细文档

- [📚 文档中心](./docs/README.md) - 所有文档导航
- [📖 项目总览](./README.md) - 完整项目说明
- [🛠️ 开发文档](./docs/开发文档.md) - 开发指南
- [🚀 启动脚本说明](./docs/启动脚本使用指南.md) - 详细启动文档
- [💻 Command文件使用](./docs/command文件使用指南.md) - macOS启动文件
- [🔧 权限问题解决](./npm权限问题解决方案.md) - npm权限修复

## 🧪 测试账户

| 用户名 | 密码 | 角色 |
|--------|------|------|
| alice | secret | 用户 |
| bob | secret | 用户 |
| courier1 | secret | 信使 |

## 🔍 可用页面

| 路由 | 页面 | 状态 |
|------|------|------|
| `/` | 首页 | ✅ |
| `/login` | 登录页面 | ✅ |
| `/register` | 注册页面 | ✅ |
| `/write` | 写信页面 | ✅ |
| `/profile` | 个人资料 | ✅ |
| `/deliver` | 投递页面 | ✅ |
| `/mailbox` | 我的信箱 | ✅ |
| `/read/[code]` | 阅读信件 | ✅ |
| `/courier/scan` | 信使扫码 | ✅ |

## ❓ 遇到问题？

1. **Go环境问题**: 确保已安装 Go 1.21+
2. **端口被占用**: 检查 8080 和 3000 端口
3. **npm 权限错误**: 运行 `./fix-npm.command`
4. **更多问题**: 查看 [故障排除文档](./docs/启动脚本使用指南.md#故障排除)