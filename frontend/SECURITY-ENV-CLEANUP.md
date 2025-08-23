# 环境变量安全清理指南

## 安全问题发现

在前端项目中发现了以下安全风险：

### 🔴 敏感信息暴露

以下敏感变量不应该存在于前端项目中：

1. **JWT密钥**
   - `JWT_SECRET` - 应移至后端
   - `JWT_REFRESH_SECRET` - 应移至后端

2. **数据库凭据**
   - `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_NAME`
   - `DATABASE_USER`, `DATABASE_PASSWORD`
   - 所有数据库连接相关配置

3. **Redis配置**
   - `REDIS_URL` - 应移至后端

4. **测试账户密码**
   - 所有 `TEST_ACCOUNT_*_PASSWORD` 变量

5. **CSRF密钥**
   - `CSRF_SECRET` - 应移至后端

## 修复措施

### ✅ 已完成

1. **创建清理后的环境文件**
   - 文件：`.env.local.clean`
   - 只包含 `NEXT_PUBLIC_` 前缀的安全变量

2. **Edge Runtime兼容性**
   - 使用 jose 库替换 jsonwebtoken
   - 创建了 `jwt-utils-edge.ts`

### 🔧 需要手动完成

**请按以下步骤完成清理：**

```bash
# 1. 备份当前环境文件
cp .env.local .env.local.backup

# 2. 使用清理后的环境文件
cp .env.local.clean .env.local

# 3. 将敏感变量移至后端
# 编辑 ../backend/.env 添加以下变量：
cat >> ../backend/.env << 'EOF'

# Moved from frontend for security
JWT_SECRET=KY6QtIecDZocllQSYoqyTkYx8AuKDkpA7RfondzVB2Y=
JWT_REFRESH_SECRET=DLUW+DbjnEeVvKABqocQRdKPUqXrJmdbhoutikwukN4=

DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=openpenpal
DATABASE_USER=rocalight
DATABASE_PASSWORD=password

REDIS_URL=redis://localhost:6379

CSRF_SECRET=dev-csrf-secret-key-change-in-production-32char

# Test account passwords (change in production)
TEST_ACCOUNT_ADMIN_PASSWORD=Admin123!
TEST_ACCOUNT_COURIER_BUILDING_PASSWORD=Secret123!
# ... other test passwords
EOF
```

## 安全最佳实践

### ✅ 前端环境变量原则

1. **只使用 NEXT_PUBLIC_ 前缀**
   - 只有 `NEXT_PUBLIC_` 开头的变量会被注入到浏览器
   - 其他变量只在构建时可用

2. **绝对禁止的内容**
   - 数据库凭据
   - JWT密钥
   - API密钥
   - 密码或敏感令牌
   - 后端专用配置

3. **允许的内容**
   - 公开的API端点URL
   - 功能开关标志
   - 环境标识（development/production）
   - 应用元数据（名称、版本等）

### 🔒 生产环境注意事项

1. **使用环境变量服务**
   - Vercel Dashboard
   - AWS Secrets Manager
   - Azure Key Vault

2. **定期轮换密钥**
   - JWT密钥应定期更换
   - 数据库密码应定期更新

3. **监控和审计**
   - 监控环境变量访问
   - 定期审计配置安全性

## 验证清理结果

```bash
# 检查前端环境文件只包含安全变量
grep -v "^#" .env.local | grep -v "^$" | grep -v "NEXT_PUBLIC_"

# 上述命令应该返回空结果，表示所有非公开变量都已移除
```

## 相关文件更新

- ✅ `middleware.ts` - 更新为使用Edge兼容的JWT工具
- ✅ `jwt-utils-edge.ts` - 新建Edge Runtime兼容版本
- ✅ `deliver/page.tsx` - 修复AuthProvider导入
- ✅ `mailbox/page.tsx` - 修复AuthProvider导入
- ✅ SSR错误修复 - 添加浏览器API防护

遵循这些安全实践可以确保应用程序的安全性，避免敏感信息泄露到客户端。