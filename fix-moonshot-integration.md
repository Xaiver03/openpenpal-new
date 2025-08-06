# Moonshot Kimi API 集成修复指南

## 当前状态

✅ **已配置**:
1. 后端 `.env` 文件已包含 Moonshot API Key: `sk-wQUdnk3pwUdEkAGQl85krQcmAE36eLl8DuKj1hLZZrjzuxvV`
2. AI Provider 已设置为 `moonshot`
3. 后端代码已实现 `callMoonshot` 方法
4. 前端组件已完整实现

⚠️ **问题**:
- AI 调用返回 fallback 内容，说明实际 API 调用失败

## 修复步骤

### 1. 验证 Moonshot API 配置

确保后端环境变量正确加载：

```bash
cd backend
cat .env | grep MOONSHOT
# 应该看到:
# MOONSHOT_API_KEY=sk-wQUdnk3pwUdEkAGQl85krQcmAE36eLl8DuKj1hLZZrjzuxvV
# AI_PROVIDER=moonshot
```

### 2. 启动后端服务并检查日志

```bash
cd backend
go run main.go
```

查看启动日志，确认 Moonshot 配置已加载。

### 3. 测试 Moonshot API 连接

运行测试脚本：

```bash
node test-moonshot-real-api.js
```

### 4. 检查后端日志

在测试时，后端会输出详细的日志：

- 🎯 [GetInspiration] Starting inspiration generation...
- 🔧 [GetInspiration] Using provider: moonshot
- 🌙 [Moonshot] 开始的请求日志
- ✅ 或 ❌ 的响应日志

### 5. 常见问题排查

#### A. API Key 无效
如果看到 401 错误，说明 API Key 无效。需要：
1. 登录 [Moonshot Platform](https://platform.moonshot.cn)
2. 生成新的 API Key
3. 更新 `backend/.env` 中的 `MOONSHOT_API_KEY`

#### B. 网络连接问题
如果看到连接超时，可能是网络问题：
1. 检查是否能访问 `https://api.moonshot.cn`
2. 检查防火墙设置
3. 尝试使用代理

#### C. 模型配置问题
默认使用 `moonshot-v1-8k` 模型，如果需要更改：
- 编辑 `backend/internal/services/ai_service.go` 第 84 行
- 可选模型：`moonshot-v1-8k`, `moonshot-v1-32k`, `moonshot-v1-128k`

### 6. 验证修复成功

修复后，再次运行测试：

```bash
# 1. 重启后端
cd backend && go run main.go

# 2. 运行测试
node test-moonshot-real-api.js
```

成功的标志：
- 返回实际 AI 生成的内容（而非 fallback）
- 内容有创意性和多样性
- 没有 "(fallback)" 标记

### 7. 前端验证

```bash
cd frontend && npm run dev
```

访问 http://localhost:3000/ai：
1. 点击"获取灵感"按钮
2. 应该看到 AI 生成的创意内容
3. 检查浏览器控制台没有错误

## 调试命令

### 直接测试 Moonshot API（绕过后端）

```bash
curl https://api.moonshot.cn/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-wQUdnk3pwUdEkAGQl85krQcmAE36eLl8DuKj1hLZZrjzuxvV" \
  -d '{
    "model": "moonshot-v1-8k",
    "messages": [
      {"role": "system", "content": "你是一个写作助手"},
      {"role": "user", "content": "给我一个关于日常生活的写信灵感"}
    ],
    "temperature": 0.7
  }'
```

## 期望结果

成功集成后，AI 功能应该：
1. 返回多样化的创意内容
2. 响应时间在 2-5 秒内
3. 内容质量高，符合用户需求
4. 支持所有 AI 功能：灵感、匹配、回信建议等

## 后续优化

1. **添加更多模型支持**：支持 Moonshot 的不同模型
2. **实现负载均衡**：在多个 AI Provider 之间切换
3. **缓存优化**：缓存常见请求结果
4. **监控和告警**：监控 API 使用量和错误率