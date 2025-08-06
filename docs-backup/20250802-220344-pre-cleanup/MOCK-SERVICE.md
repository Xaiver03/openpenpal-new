# OpenPenPal Mock Services 

统一的微服务 Mock 框架，用于本地开发和测试。

## 🎯 功能特性

- **统一架构**: 所有微服务共享一套 Mock 框架
- **权限控制**: 完整的用户角色和权限管理
- **智能路由**: 自动路由分发和服务发现
- **响应模拟**: 支持延迟、错误率等真实场景模拟
- **开发友好**: 热重载、彩色日志、详细错误信息
- **测试集成**: 内置权限测试和 API 测试工具

## 📁 项目结构

```
apps/mock-services/
├── src/
│   ├── config/
│   │   ├── users.js          # 用户和权限配置
│   │   └── services.js       # 服务配置
│   ├── middleware/
│   │   ├── auth.js           # 认证和权限中间件
│   │   └── response.js       # 响应处理中间件
│   ├── api/
│   │   ├── auth/
│   │   │   └── login.js      # 认证 API
│   │   ├── write/
│   │   │   └── letters.js    # 写信服务 API
│   │   └── courier/
│   │       └── tasks.js      # 信使服务 API
│   ├── utils/
│   │   └── logger.js         # 日志工具
│   ├── router.js             # 路由管理
│   └── index.js              # 主入口
├── test/
│   ├── run-tests.js          # 测试运行器
│   └── test-permissions.js   # 权限测试
├── package.json
└── README.md
```

## 🚀 快速开始

### 1. 安装依赖

```bash
# 使用启动脚本安装
./scripts/start-mock.sh --install

# 或手动安装
cd apps/mock-services
npm install
```

### 2. 启动服务

```bash
# 启动所有服务
./scripts/start-mock.sh

# 启动特定服务
./scripts/start-mock.sh gateway        # 只启动 API Gateway
./scripts/start-mock.sh write --watch  # 启动写信服务并启用热重载

# 查看更多选项
./scripts/start-mock.sh --help
```

### 3. 验证服务

```bash
# 检查服务状态
./scripts/start-mock.sh --status

# 运行测试
./scripts/start-mock.sh --test
```

## 🔐 用户和权限系统

### 预配置用户

| 用户名 | 密码 | 角色 | 权限 | 说明 |
|--------|------|------|------|------|
| admin | admin123 | super_admin | ALL | 超级管理员 |
| alice | secret | student | 基础用户权限 | 北大学生 |
| bob | password123 | student | 基础用户权限 | 清华学生 |
| courier1 | courier123 | courier | 信使权限 | 北大区域信使 |
| courier2 | courier456 | courier | 信使权限 | 清华区域信使 |
| moderator | mod123 | moderator | 内容审核权限 | 内容审核员 |

### 权限说明

```javascript
// 服务权限映射
const SERVICE_PERMISSIONS = {
  'write-service': ['WRITE_READ', 'WRITE_CREATE', 'LETTER_READ', 'LETTER_SEND'],
  'courier-service': ['COURIER_READ', 'COURIER_WRITE', 'TASK_READ', 'TASK_ACCEPT'],
  'admin-service': ['ADMIN_READ', 'ADMIN_WRITE', 'USER_MANAGE', 'SYSTEM_CONFIG'],
  'main-backend': ['PROFILE_READ', 'PROFILE_UPDATE', 'USER_MANAGE'],
  'ocr-service': ['OCR_READ', 'OCR_PROCESS']
};
```

## 🌐 API 接口

### 认证接口

```bash
# 用户登录
POST http://localhost:8000/api/auth/login
{
  "username": "alice",
  "password": "secret"
}

# 获取当前用户信息
GET http://localhost:8000/api/auth/me
Authorization: Bearer <token>
```

### 写信服务接口

```bash
# 创建信件
POST http://localhost:8000/api/write/letters
Authorization: Bearer <token>
{
  "title": "给朋友的一封信",
  "content": "信件内容...",
  "receiverHint": "北京大学计算机系的朋友"
}

# 获取信件列表
GET http://localhost:8000/api/write/letters?page=0&limit=20
Authorization: Bearer <token>
```

### 信使服务接口

```bash
# 获取可用任务
GET http://localhost:8000/api/courier/tasks?page=0&limit=20
Authorization: Bearer <token>

# 接受任务
POST http://localhost:8000/api/courier/tasks/{id}/accept
Authorization: Bearer <token>

# 更新任务状态
PUT http://localhost:8000/api/courier/tasks/{id}/status
Authorization: Bearer <token>
{
  "status": "picked_up",
  "note": "已取件"
}
```

### 管理服务接口

```bash
# 获取用户列表 (需要管理员权限)
GET http://localhost:8000/api/admin/users?page=0&size=20
Authorization: Bearer <admin_token>

# 获取博物馆展览
GET http://localhost:8000/api/admin/museum/exhibitions
Authorization: Bearer <admin_token>
```

## 🔧 配置说明

### 服务配置 (`src/config/services.js`)

```javascript
export const SERVICES = {
  'gateway': {
    name: 'API Gateway',
    port: 8000,
    basePath: '/api',
    delay: { min: 50, max: 200 },
    enabled: true
  },
  'write-service': {
    name: 'Write Service',
    port: 8001,
    basePath: '/api',
    delay: { min: 150, max: 400 },
    enabled: true
  }
  // ...其他服务
};
```

### 响应延迟模拟

```bash
# 全局启用延迟模拟
export DEFAULT_CONFIG.globalDelay.enabled = true

# 或在请求中指定延迟
GET http://localhost:8000/api/write/letters?delay=500
```

### 错误模拟

```javascript
// 配置错误模拟
export const DEFAULT_CONFIG = {
  errorSimulation: {
    enabled: true,
    probability: 0.1, // 10% 概率返回错误
    types: ['network', 'server', 'timeout']
  }
};
```

## 🧪 测试

### 运行所有测试

```bash
./scripts/start-mock.sh --test
```

### 权限测试

```bash
cd apps/mock-services
npm run test:permissions
```

### 手动 API 测试

```bash
# 使用 curl 测试登录
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret"}'

# 使用返回的 token 访问受保护接口
curl -X GET http://localhost:8000/api/write/letters \
  -H "Authorization: Bearer <your_token_here>"
```

## 📊 监控和日志

### 日志级别

```bash
# 设置日志级别
./scripts/start-mock.sh all --log=debug
```

### 服务状态检查

```bash
# 查看所有服务状态
./scripts/start-mock.sh --status

# 输出示例:
# ✓ gateway (端口 8000) - PID: 12345
# ✓ write-service (端口 8001) - PID: 12346
# ✗ courier-service (端口 8002) - 未运行
```

## 🔄 开发模式

### 热重载

```bash
# 启用文件监听模式
./scripts/start-mock.sh write --watch
```

### 调试模式

```bash
# 启用详细日志
./scripts/start-mock.sh all --log=debug
```

## 🚦 常见使用场景

### 1. 前端开发联调

```bash
# 启动 API Gateway 提供统一入口
./scripts/start-mock.sh gateway --watch

# 前端配置 API 基础路径
VITE_API_BASE_URL=http://localhost:8000/api
```

### 2. 单服务开发

```bash
# 只启动写信服务进行开发
./scripts/start-mock.sh write --watch --log=debug
```

### 3. 集成测试

```bash
# 启动所有服务
./scripts/start-mock.sh all

# 运行集成测试
npm test
```

### 4. 压力测试

```bash
# 启用错误模拟测试容错性
./scripts/start-mock.sh all --env=production

# 配置高延迟测试性能
# 在代码中设置 DEFAULT_CONFIG.globalDelay = { enabled: true, min: 1000, max: 3000 }
```

## 🛠️ 扩展开发

### 添加新的 API 接口

1. 在 `src/api/{service}/` 目录下创建新的处理函数
2. 在 `src/router.js` 中注册路由
3. 更新权限配置（如需要）

### 添加新的微服务

1. 在 `src/config/services.js` 中添加服务配置
2. 在 `src/router.js` 中添加服务路由设置函数
3. 创建对应的 API 处理函数

### 自定义中间件

```javascript
// 在 src/middleware/ 目录下创建新中间件
export function customMiddleware() {
  return (req, res, next) => {
    // 自定义逻辑
    next();
  };
}
```

## 🐛 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 停止所有服务
   ./scripts/start-mock.sh --stop
   
   # 检查端口占用
   lsof -i :8000
   ```

2. **依赖安装失败**
   ```bash
   # 清理缓存重新安装
   cd apps/mock-services
   rm -rf node_modules package-lock.json
   npm install
   ```

3. **权限错误**
   ```bash
   # 检查用户配置
   cat src/config/users.js
   
   # 运行权限测试
   npm run test:permissions
   ```

### 调试技巧

```bash
# 启用详细日志
./scripts/start-mock.sh gateway --log=debug

# 查看网络请求
# 在浏览器开发者工具的 Network 标签中查看请求详情

# 使用 curl 测试
curl -v http://localhost:8000/api/auth/login
```

## 📚 参考资料

- [Express.js 文档](https://expressjs.com/)
- [JWT 规范](https://jwt.io/)
- [OpenPenPal 项目文档](../README.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](../LICENSE) 文件