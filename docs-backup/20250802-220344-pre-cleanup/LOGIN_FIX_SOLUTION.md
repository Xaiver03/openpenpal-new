# 登录问题修复方案文档

## 问题总结

1. **问题现象**：除了 admin 用户外，其他用户（alice, courier_level1等）无法登录，返回 401 错误
2. **根本原因**：数据库中的密码哈希格式不正确，需要使用 Go 的 bcrypt 库生成的哈希值

## 已实施的修复

### 1. 更新密码哈希值

所有用户的密码已更新为正确的 bcrypt 哈希值：

```javascript
// 正确的 bcrypt 哈希值（Go 生成）
const passwords = {
  'secret': '$2a$10$KuNOKKOmFExYEe/BYHOQWOtuwywR3mHeOeBm7On0ZAozMWVqcmoU.',
  'admin123': '$2a$10$cH8Xq3cHw.nxkHBtepdYBekdP/85F1cn1LMBqii7tjB.VSmjInf/i'
};
```

### 2. 测试账号状态

现在所有测试账号都可以正常登录：

| 用户名 | 密码 | 角色 | 状态 |
|--------|------|------|------|
| admin | admin123 | super_admin | ✅ 可登录 |
| alice | secret | user | ✅ 可登录 |
| courier_level1 | secret | courier_level1 | ✅ 可登录 |
| courier_level2 | secret | courier_level2 | ✅ 可登录 |
| courier_level3 | secret | courier_level3 | ✅ 可登录 |
| courier_level4 | secret | courier_level4 | ✅ 可登录 |

## 系统架构说明

### 服务端口

- **网关服务**: 8000（路由转发）
- **后端服务**: 8080（实际处理）
- **前端应用**: 3000

### 登录流程

1. 前端发送 POST 请求到 `/api/v1/auth/login`
2. 请求被发送到后端服务（8080）
3. 后端验证用户名和密码
4. 生成 JWT token 并返回

## 预防措施

### 1. 密码生成标准化

创建新用户时，必须使用 Go 的 bcrypt 库生成密码哈希：

```go
import "golang.org/x/crypto/bcrypt"

hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
```

### 2. 定期验证脚本

创建了验证脚本 `test-backend-direct.js` 用于测试所有用户登录：

```bash
node test-backend-direct.js
```

### 3. 监控要点

- 监控登录失败率
- 检查 401 错误日志
- 定期验证测试账号可用性

## 故障排查步骤

如果再次出现登录问题：

1. **检查服务状态**
   ```bash
   ./startup/check-status.sh
   ```

2. **验证数据库连接**
   ```bash
   node check-user-passwords.js
   ```

3. **测试登录功能**
   ```bash
   node test-backend-direct.js
   ```

4. **检查密码哈希**
   ```bash
   cd backend && go run debug-login.go
   ```

5. **查看日志**
   ```bash
   tail -f logs/backend.log
   ```

## 注意事项

1. **速率限制**：生产环境有速率限制（每分钟6次），测试时注意间隔
2. **环境变量**：可设置 `TEST_MODE=true` 放宽速率限制
3. **密码规则**：必须使用 bcrypt cost 10 生成哈希

## 联系方式

如有问题，请查看：
- 系统日志：`logs/` 目录
- 测试脚本：`test-*.js` 文件
- 调试工具：`debug-*.go` 文件

---
更新时间：2025-08-01
版本：1.0