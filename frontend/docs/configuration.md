# OpenPenPal 配置管理指南

## 📋 概述

OpenPenPal 使用统一的环境变量配置系统，支持开发、测试和生产环境的完整配置管理。

## 🏗️ 配置系统架构

```
src/lib/config/
├── env.ts              # 环境变量管理核心
├── validator.ts        # 配置验证和安全检查
├── initializer.ts      # 应用启动时配置初始化
└── index.ts           # 统一导出接口

src/config/             # 应用级配置
└── courier-test-accounts.ts  # 测试账号配置
```

## 🔧 使用方法

### 1. 基础使用

```typescript
import { getConfig } from '@/lib/config'

const config = getConfig()
console.log(config.backend.url)  // 获取后端服务地址
console.log(config.ai.provider)  // 获取AI提供商
```

### 2. 配置验证

```typescript
import { validateConfiguration, printValidationResult } from '@/lib/config'

const validation = validateConfiguration()
if (!validation.isValid) {
  printValidationResult(validation)
  throw new Error('配置验证失败')
}
```

### 3. 应用初始化

```typescript
import { initializeConfiguration } from '@/lib/config'

// 在应用启动时调用
await initializeConfiguration({
  validateConfig: true,
  printResults: true,
  throwOnErrors: process.env.NODE_ENV === 'production'
})
```

## ⚙️ 环境变量配置

### 必需的环境变量

#### 开发环境
```bash
# 基础配置
NODE_ENV=development
DEBUG=true

# 后端服务
BACKEND_URL=http://localhost:8080
```

#### 生产环境
```bash
# 基础配置（必需）
NODE_ENV=production
DEBUG=false

# 安全配置（必需）
JWT_SECRET=your_secure_64_char_secret_key_here
DB_PASSWORD=your_secure_database_password

# 数据库配置（必需）
DATABASE_URL=postgres://user:password@host:port/database
```

### 完整配置示例

参考 `.env.example` 文件获取完整的配置示例。

## 🔒 安全最佳实践

### 1. 环境变量安全

- ✅ **绝不提交** `.env` 文件到版本控制
- ✅ **使用强密钥** JWT密钥长度至少32个字符
- ✅ **分离环境** 开发/测试/生产使用不同的密钥
- ✅ **定期轮换** 定期更换敏感凭据

### 2. 配置验证

系统会自动验证：
- JWT密钥强度
- 必需的环境变量
- 数据库连接安全性
- CORS配置安全性

### 3. 生产环境检查

生产环境启动时会强制检查：
```bash
✅ JWT_SECRET 已设置且安全
✅ 数据库密码已设置
✅ 调试模式已关闭
✅ SSL连接已启用
```

## 🚀 部署配置

### 1. 环境文件管理

```bash
# 开发环境
cp .env.example .env.development

# 生产环境  
cp .env.example .env.production
# 编辑 .env.production 设置实际值
```

### 2. 容器化部署

Docker 环境变量注入：
```dockerfile
ENV NODE_ENV=production
ENV JWT_SECRET=${JWT_SECRET}
ENV DATABASE_URL=${DATABASE_URL}
```

### 3. 云平台部署

#### Vercel
```bash
vercel env add JWT_SECRET
vercel env add DATABASE_URL
```

#### Railway/Heroku
通过控制台或CLI设置环境变量。

## 🛠️ 开发工具

### 配置验证命令

```bash
# 验证当前配置
npm run config:validate

# 生成配置报告
npm run config:report

# 健康检查
npm run config:health
```

### 配置调试

开发环境下，启动时会自动打印配置摘要：
```
🔧 应用配置已加载:
  environment: development
  backend: http://localhost:8080
  ai: { provider: 'siliconflow', hasKey: true }
  database: { host: 'localhost', port: 5432, name: 'openpenpal' }
```

## 🔍 故障排除

### 常见问题

#### 1. 配置加载失败
```
❌ 配置加载失败: 缺少必需的环境变量: JWT_SECRET
```
**解决方案**: 检查 `.env` 文件是否存在并设置了所需变量。

#### 2. JWT密钥不安全
```
🔐 安全问题: JWT密钥长度不足32个字符
```
**解决方案**: 使用 `openssl rand -base64 32` 生成安全密钥。

#### 3. 数据库连接失败
```
⚠️ 后端服务连接失败: Connection refused
```
**解决方案**: 确保数据库服务正在运行且连接信息正确。

### 调试模式

设置 `DEBUG=true` 启用详细的配置调试信息：
```bash
DEBUG=true npm run dev
```

## 📚 API 参考

### getConfig()
```typescript
function getConfig(): AppConfig
```
获取完整的应用配置对象。

### validateConfiguration()
```typescript
function validateConfiguration(): ValidationResult
```
验证当前配置的完整性和安全性。

### initializeConfiguration()
```typescript
function initializeConfiguration(options?: {
  validateConfig?: boolean
  printResults?: boolean  
  throwOnErrors?: boolean
}): Promise<InitializationStatus>
```
初始化配置系统并验证外部服务连接。

## 🔄 迁移指南

### 从旧配置系统迁移

1. **安装新配置系统**:
   ```typescript
   import { getConfig } from '@/lib/config'
   ```

2. **替换硬编码值**:
   ```typescript
   // 旧方式
   const backendUrl = 'http://localhost:8080'
   
   // 新方式
   const config = getConfig()
   const backendUrl = config.backend.url
   ```

3. **添加配置验证**:
   ```typescript
   // 在应用启动时
   await initializeConfiguration()
   ```

## 📞 支持

如有配置相关问题，请：
1. 检查本文档的故障排除部分
2. 运行 `npm run config:health` 进行健康检查
3. 查看应用启动日志中的配置验证结果