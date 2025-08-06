# OpenPenPal 项目启动指南

欢迎使用 OpenPenPal！本指南将帮助您快速启动和运行整个项目。

## 🎯 快速开始（推荐）

### 方式一：一键启动（macOS）
```bash
# 双击运行（图形界面）
./startup/openpenpal-launcher.command

# 或命令行运行
./startup/quick-start.sh
```

### 方式二：分步启动
```bash
# 1. 启动简化Mock服务
./startup/start-simple-mock.sh

# 2. 启动完整Mock服务 
./startup/start-complete-mock.sh

# 3. 启动前端开发服务器
./startup/start-frontend.sh

# 4. 启动生产环境
./startup/start-production.sh
```

## 📁 启动文件说明

### 🚀 核心启动脚本

| 文件名 | 功能 | 使用场景 |
|--------|------|----------|
| `quick-start.sh` | 一键启动所有服务 | 日常开发 |
| `openpenpal-launcher.command` | macOS图形界面启动器 | 演示/体验 |
| `start-simple-mock.sh` | 启动简化Mock服务 | 快速原型 |
| `start-complete-mock.sh` | 启动完整Mock服务 | 完整开发 |
| `start-frontend.sh` | 启动前端服务 | 前端开发 |
| `start-production.sh` | 启动生产环境 | 部署/演示 |

### 🔧 辅助工具

| 文件名 | 功能 | 使用场景 |
|--------|------|----------|
| `stop-all.sh` | 停止所有服务 | 清理环境 |
| `check-status.sh` | 检查服务状态 | 调试诊断 |
| `install-deps.sh` | 安装所有依赖 | 初次部署 |
| `migrate-mock.sh` | 迁移Mock服务配置 | 版本升级 |

### 📋 配置文件

| 文件名 | 功能 | 说明 |
|--------|------|------|
| `startup-config.json` | 启动配置 | 端口、服务配置 |
| `environment-vars.sh` | 环境变量设置 | 共享配置 |
| `service-urls.json` | 服务地址映射 | API路由配置 |

## 🛠️ 环境要求

### 必需软件
- **Node.js**: 18.0+ (推荐 LTS 版本)
- **npm**: 8.0+ (随Node.js安装)
- **Git**: 最新版本

### 推荐软件
- **Docker**: 用于容器化部署
- **VS Code**: 开发环境
- **Postman**: API测试

### 系统要求
- **内存**: 最少 4GB，推荐 8GB+
- **磁盘**: 最少 2GB 可用空间
- **网络**: 需要互联网连接下载依赖

## 🚀 详细启动步骤

### 第一次使用

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd openpenpal
   ```

2. **安装依赖**
   ```bash
   ./startup/install-deps.sh
   ```

3. **启动服务**
   ```bash
   ./startup/quick-start.sh
   ```

4. **验证启动**
   ```bash
   ./startup/check-status.sh
   ```

### 日常开发

1. **启动开发环境**
   ```bash
   ./startup/quick-start.sh --dev
   ```

2. **查看服务状态**
   ```bash
   ./startup/check-status.sh
   ```

3. **停止所有服务**
   ```bash
   ./startup/stop-all.sh
   ```

## 🌐 服务端口说明

| 服务 | 端口 | 地址 | 说明 |
|------|------|------|------|
| **前端应用** | 3000 | http://localhost:3000 | 主用户界面 |
| **管理后台** | 3001 | http://localhost:3001 | 管理员界面 |
| **API网关** | 8000 | http://localhost:8000 | 统一API入口 |
| **主后端** | 8080 | http://localhost:8080 | 核心业务服务 |
| **写信服务** | 8001 | http://localhost:8001 | 信件处理 |
| **信使服务** | 8002 | http://localhost:8002 | 配送管理 |
| **管理服务** | 8003 | http://localhost:8003 | 系统管理 |
| **OCR服务** | 8004 | http://localhost:8004 | 文字识别 |

## 🔐 测试账号

### 普通用户
- **用户名**: alice
- **密码**: secret
- **角色**: 学生
- **学校**: 北京大学

### 管理员
- **用户名**: admin
- **密码**: admin123
- **角色**: 超级管理员
- **权限**: 全部权限

### 信使
- **用户名**: courier1
- **密码**: courier123
- **角色**: 配送员
- **区域**: 北京大学

## 🎨 启动模式

### 开发模式
```bash
./startup/quick-start.sh --dev
```
- 启用热重载
- 详细日志输出
- 开发工具集成

### 生产模式
```bash
./startup/start-production.sh
```
- 优化构建
- 压缩资源
- 性能监控

### 测试模式
```bash
./startup/quick-start.sh --test
```
- 启动测试数据库
- 运行自动化测试
- 生成测试报告

### Demo模式
```bash
./startup/openpenpal-launcher.command
```
- 自动打开浏览器
- 预填测试数据
- 引导式体验

## 🐛 常见问题

### 端口冲突
**问题**: 提示端口被占用
**解决**: 
```bash
./startup/stop-all.sh
./startup/quick-start.sh
```

### 依赖安装失败
**问题**: npm install 报错
**解决**:
```bash
# 清理缓存
npm cache clean --force
rm -rf node_modules package-lock.json
./startup/install-deps.sh
```

### 服务启动失败
**问题**: 某个服务无法启动
**解决**:
```bash
# 检查详细状态
./startup/check-status.sh --verbose

# 查看日志
tail -f logs/*.log

# 重新启动
./startup/stop-all.sh
./startup/quick-start.sh
```

### 前端无法访问后端
**问题**: API调用失败
**解决**:
1. 检查后端服务是否运行: `./startup/check-status.sh`
2. 检查网络配置: 查看 `startup/service-urls.json`
3. 重启服务: `./startup/quick-start.sh`

## 📊 监控和日志

### 实时日志
```bash
# 查看所有服务日志
tail -f logs/*.log

# 查看特定服务日志
tail -f logs/frontend.log
tail -f logs/backend.log
tail -f logs/mock-services.log
```

### 性能监控
```bash
# 检查服务性能
./startup/check-status.sh --performance

# 查看资源使用
./startup/check-status.sh --resources
```

### 健康检查
```bash
# 自动健康检查
./startup/check-status.sh --health

# 手动API测试
curl http://localhost:8000/health
```

## 🔄 更新和维护

### 更新代码
```bash
git pull origin main
./startup/install-deps.sh
./startup/quick-start.sh
```

### 清理环境
```bash
./startup/stop-all.sh
./startup/clean-cache.sh
```

### 重置到初始状态
```bash
./startup/reset-environment.sh
```

## 📚 进阶使用

### Docker部署
```bash
# 构建Docker镜像
./startup/build-docker.sh

# 启动Docker容器
./startup/start-docker.sh
```

### 自定义配置
1. 编辑 `startup/startup-config.json`
2. 修改端口和服务配置
3. 重新启动服务

### 扩展服务
1. 在 `startup/` 目录添加新的启动脚本
2. 更新 `startup-config.json`
3. 参考现有脚本编写启动逻辑

## 🤝 支持和反馈

### 获取帮助
- 查看项目文档: `docs/`
- 查看API文档: http://localhost:8000/docs
- 提交问题: 项目Issues页面

### 贡献代码
1. Fork项目
2. 创建特性分支
3. 提交更改
4. 创建Pull Request

---

**Happy Coding! 🚀**

*如果您在使用过程中遇到任何问题，请参考常见问题部分或提交Issue。*