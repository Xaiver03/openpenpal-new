# OpenPenPal 启动脚本使用指南

> 🔄 **最新更新**: 2024-01-20 - 重新整理启动脚本，简化使用流程

## 📋 当前可用的启动脚本

| 脚本名称 | 用途 | 推荐场景 | 状态 |
|---------|------|----------|------|
| `start-all.command` | 同时启动前后端 | **首次启动**或完整重启 | ✅ 推荐 |
| `start-backend.command` | 只启动Go后端 | 前端已运行，重启后端 | ✅ 可用 |
| `start-frontend.command` | 只启动Next.js前端 | 后端已运行，启动前端 | ✅ 推荐 |
| `fix-npm.command` | 修复npm权限问题 | npm安装失败时 | ✅ 工具 |

## 🚀 推荐的启动方式

### 🎯 首次启动或完整重启
```bash
./start-all.command
```

**适用场景：**
- 第一次启动项目
- 两个服务都需要重启
- 不确定当前服务状态

**功能：**
- ✅ 检查Go和Node.js环境
- ✅ 检查端口占用
- ✅ 安装/更新依赖
- ✅ 启动Go后端服务 (8080端口)
- ✅ 启动Next.js前端服务 (3000端口)
- ✅ 自动打开浏览器
- ✅ 提供完整的日志和错误提示

### 🌐 只启动前端（当前推荐）
```bash
./start-frontend.command
```

**适用场景：**
- Go后端已在运行
- 只需要启动或重启前端
- 前端开发调试

**功能：**
- ✅ 检查Node.js环境
- ✅ 检查后端服务状态
- ✅ 安装前端依赖
- ✅ 创建环境配置文件
- ✅ 启动Next.js开发服务器
- ✅ 自动打开浏览器

### 🔧 只启动后端
```bash
./start-backend.command
```

**适用场景：**
- 前端已在运行
- 只需要启动或重启后端
- 后端开发调试

## 🛠️ 手动启动方式

如果不想使用启动脚本，也可以手动启动：

### 启动后端
```bash
cd backend
go mod download
cp .env.example .env
go run main.go
```

### 启动前端
```bash
cd frontend
npm install
echo "NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1" > .env.local
npm run dev
```

## 🌐 服务地址

启动成功后，可以访问：

- **前端界面**: http://localhost:3000
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health
- **API文档**: http://localhost:8080/api/v1/

## 🔍 可用页面路由

目前已实现的页面：

| 路由 | 页面 | 状态 |
|------|------|------|
| `/` | 首页 (纸黄色主题) | ✅ |
| `/write` | 写信页面 | ✅ |
| `/mailbox` | 我的信箱 | ✅ |
| `/profile` | 个人资料页面 | ✅ |
| `/settings` | 设置页面 | ✅ |
| `/about` | 关于我们页面 | ✅ |
| `/courier/scan` | 信使扫码 | ✅ |
| `/404` | 自定义404页面 | ✅ |

**注意**: 所有页面已采用纸黄色 (amber) 主题设计，提供温暖舒适的用户体验。

## 🐛 故障排除

### 1. Go环境问题
```bash
# 检查Go是否安装
go version

# 如果未安装，需要先安装Go
brew install go  # macOS
```

### 2. 端口被占用
```bash
# 查看端口占用
lsof -i :8080  # 后端端口
lsof -i :3000  # 前端端口

# 杀死占用进程
kill -9 <PID>
```

### 3. npm权限问题
```bash
# 使用修复脚本
./fix-npm.command

# 或手动修复
sudo chown -R $(whoami) ~/.npm
npm cache clean --force
```

### 4. 数据库问题
```bash
# 删除数据库文件重新初始化
rm backend/openpenpal.db
# 重新启动后端服务
```

## 📝 测试账户

开发环境默认创建的测试账户：

| 用户名 | 密码 | 角色 | 学校代码 |
|--------|------|------|----------|
| alice | secret | 用户 | BJDX01 |
| bob | secret | 用户 | BJDX01 |
| courier1 | secret | 信使 | BJDX01 |

## 🔧 开发调试

### 查看日志
```bash
# 查看后端日志（如果使用start-all.command启动）
tail -f logs/backend.log

# 查看前端日志
tail -f logs/frontend.log
```

### API测试
```bash
# 测试后端健康检查
curl http://localhost:8080/health

# 测试用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123","nickname":"Test","school_code":"BJDX01"}'
```

## 📞 获得帮助

### 启动失败？
1. 查看错误日志
2. 检查Go和Node.js环境
3. 尝试使用 `fix-npm.command` 修复npm问题
4. 查看故障排除章节

### 需要支持？
- 📖 查看[开发文档](./开发文档.md)
- 🗺️ 查看[开发计划](./开发计划.md)
- 🏠 返回[项目首页](../README.md)

---

*让启动变得简单，让开发更加专注！* 🚀✨