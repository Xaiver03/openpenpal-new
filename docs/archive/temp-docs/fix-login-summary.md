# 登录问题修复总结

## 当前状态

1. **密码已修复** ✅
   - 所有测试账号的密码都已更新为 `password`
   - 包括: admin, courier_level1, courier1, user1

2. **前端运行在端口 3001**
   - 不是默认的 3000 端口
   - 这可能导致 CORS 问题

3. **CORS 错误**
   ```
   Access to XMLHttpRequest at 'http://localhost:8003/api/admin/auth/login' 
   from origin 'http://localhost:3001' has been blocked by CORS policy
   ```

## 解决方案

### 方案 1: 使用正确的端口访问前端
访问 http://localhost:3000 而不是 http://localhost:3001

### 方案 2: 配置前端使用 API 网关
确保所有 API 调用都通过网关 (端口 8000) 而不是直接访问服务

### 方案 3: 临时解决 - 添加 CORS 支持
在 admin-service (Java Spring Boot) 中添加 CORS 配置以支持端口 3001

## 测试步骤

1. 首先尝试访问 http://localhost:3000/login
2. 使用以下账号登录：
   - 管理员: `admin` / `password`
   - 信使: `courier_level1` / `password`
   - 普通用户: `user1` / `password`

## 服务状态检查

```bash
# 检查所有服务
lsof -i :3000  # 前端 (正常端口)
lsof -i :3001  # 前端 (当前运行端口)
lsof -i :8000  # API 网关
lsof -i :8080  # 主后端
lsof -i :8003  # Admin 服务
```

## 问题根源

前端似乎配置了直接访问 admin-service (8003) 而不是通过 API 网关。这违反了微服务架构的最佳实践。所有外部请求应该通过 API 网关统一路由。