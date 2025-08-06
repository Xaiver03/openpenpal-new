# 🔐 JWT安全配置部署指南

> **重要：生产环境必读**  
> **更新日期：** 2025-07-24  
> **状态：** 生产就绪

---

## 📋 概述

OpenPenPal使用JWT（JSON Web Token）进行用户认证和授权。为确保系统安全，生产环境必须使用强随机密钥。

## 🔑 密钥生成

### 1. 生成强随机密钥

```bash
# 生成JWT主密钥 (256位)
node -e "console.log('JWT_SECRET=' + require('crypto').randomBytes(32).toString('base64'))"

# 生成JWT刷新令牌密钥 (256位)  
node -e "console.log('JWT_REFRESH_SECRET=' + require('crypto').randomBytes(32).toString('base64'))"
```

### 2. 密钥要求

- **长度：** 32字节 (256位)
- **编码：** Base64
- **熵：** 高强度加密随机数
- **字符数：** 44个字符（Base64编码后）

### 3. 示例输出

```
JWT_SECRET=KY6QtIecDZocllQSYoqyTkYx8AuKDkpA7RfondzVB2Y=
JWT_REFRESH_SECRET=DLUW+DbjnEeVvKABqocQRdKPUqXrJmdbhoutikwukN4=
```

## ⚙️ 环境配置

### 生产环境 (.env.production)

```bash
# JWT安全配置
JWT_SECRET=your-production-jwt-secret-here
JWT_REFRESH_SECRET=your-production-refresh-secret-here
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=7d

# 数据库安全配置
DATABASE_PASSWORD=your-secure-production-password

# Redis配置
REDIS_URL=redis://your-redis-host:6379
REDIS_SESSION_TTL=3600
```

### 开发环境 (.env.local)

```bash
# JWT安全配置 - 已配置强密钥
JWT_SECRET=KY6QtIecDZocllQSYoqyTkYx8AuKDkpA7RfondzVB2Y=
JWT_REFRESH_SECRET=DLUW+DbjnEeVvKABqocQRdKPUqXrJmdbhoutikwukN4=
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=7d
```

## 🚀 部署步骤

### 1. 生产部署前检查

```bash
# 检查是否使用默认密钥
grep -r "your-super-secret" . || echo "✅ 无默认密钥"

# 检查JWT密钥长度
node -e "
const secret = process.env.JWT_SECRET;
if (!secret) {
  console.log('❌ JWT_SECRET未设置');
  process.exit(1);
}
if (secret.length < 40) {
  console.log('❌ JWT密钥太短:', secret.length);
  process.exit(1);
}
console.log('✅ JWT密钥长度正常:', secret.length);
"
```

### 2. 容器化部署

```dockerfile
# Dockerfile示例
FROM node:18-alpine

# 不在镜像中包含敏感信息
# JWT密钥通过环境变量注入

COPY package*.json ./
RUN npm ci --only=production

COPY . .
EXPOSE 3000

CMD ["npm", "start"]
```

```yaml
# docker-compose.yml示例
version: '3.8'
services:
  frontend:
    build: .
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
      - DATABASE_PASSWORD=${DATABASE_PASSWORD}
    env_file:
      - .env.production
```

### 3. Kubernetes部署

```yaml
# k8s-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: openpenpal-secrets
type: Opaque
stringData:
  jwt-secret: "your-base64-encoded-secret"
  jwt-refresh-secret: "your-base64-encoded-refresh-secret"
  database-password: "your-secure-db-password"
```

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openpenpal-frontend
spec:
  template:
    spec:
      containers:
      - name: frontend
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: openpenpal-secrets
              key: jwt-secret
        - name: JWT_REFRESH_SECRET
          valueFrom:
            secretKeyRef:
              name: openpenpal-secrets
              key: jwt-refresh-secret
```

## 🔒 安全最佳实践

### 密钥管理

1. **定期轮换：** 建议每90天轮换JWT密钥
2. **分离存储：** 生产密钥不存储在代码仓库
3. **访问控制：** 仅授权人员可访问生产密钥
4. **审计日志：** 记录所有密钥访问操作

### 令牌安全

```javascript
// jwt-utils.ts - 安全配置示例
const JWT_CONFIG = {
  secret: process.env.JWT_SECRET,
  refreshSecret: process.env.JWT_REFRESH_SECRET,
  expiresIn: process.env.JWT_EXPIRES_IN || '15m',
  refreshExpiresIn: process.env.JWT_REFRESH_EXPIRES_IN || '7d',
  issuer: 'openpenpal-auth',
  audience: 'openpenpal-client',
  algorithm: 'HS256'  // 使用HMAC SHA-256
}
```

## 🔍 验证部署

### 1. 密钥强度检查

```bash
# 检查JWT密钥熵
node -e "
const crypto = require('crypto');
const secret = process.env.JWT_SECRET;
const buffer = Buffer.from(secret, 'base64');
console.log('密钥字节数:', buffer.length);
console.log('预期字节数: 32');
console.log('状态:', buffer.length >= 32 ? '✅ 合格' : '❌ 不合格');
"
```

### 2. 令牌验证测试

```bash
# 测试JWT生成和验证
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}'

# 验证返回的token格式正确
```

### 3. 安全扫描

```bash
# 检查敏感信息泄露
git log --all --full-history -- "*.env*" | grep -i "secret\|password"

# 检查明文凭据
grep -r "password.*=" . --exclude-dir=node_modules
```

## ⚠️ 故障排除

### 常见问题

1. **Token无效错误**
   ```
   错误: "令牌格式无效"
   原因: JWT密钥与签发时不一致
   解决: 确认生产环境JWT_SECRET配置正确
   ```

2. **Token过期**
   ```
   错误: "令牌已过期"  
   原因: JWT_EXPIRES_IN配置过短
   解决: 调整过期时间或实现refresh token机制
   ```

3. **密钥格式错误**
   ```
   错误: 签名验证失败
   原因: Base64编码格式不正确
   解决: 重新生成密钥，确保Base64编码
   ```

## 📞 技术支持

- **文档版本：** v1.0
- **联系方式：** OpenPenPal开发团队
- **紧急联系：** 生产环境问题请立即联系运维团队

---

**🔐 记住：安全是我们的首要任务。永远不要在代码中硬编码密钥！**