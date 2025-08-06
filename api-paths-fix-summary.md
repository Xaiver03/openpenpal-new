# API 路径修复总结

## 问题描述
浏览器显示多个 404 错误：
- `letters/popular?period=weekly&limit=6`
- `api/v1/shop/products?page=1&limit=12`

## 根本原因
当我将 `apiClient` 的基础路径改为空字符串 `''` 后，所有 API 调用都需要完整路径包含 `/api/v1` 前缀。

## 修复内容

### 1. Letter Service 路径修复
修改了所有缺少完整路径的 API 调用：
- `/letters/read/${code}` → `/api/v1/letters/read/${code}`
- `/letters` → `/api/v1/letters`
- `/letters/${letterId}` → `/api/v1/letters/${letterId}`
- `/letters/stats` → `/api/v1/letters/stats`
- `/letters/templates/${templateId}` → `/api/v1/letters/templates/${templateId}`
- `/letters/recommended` → `/api/v1/letters/recommended`
- `/letters/public` → `/api/v1/letters/public`

### 2. Shop Service 问题
Shop service 已经使用正确的路径 `/api/v1/shop/products`，但后端返回 500 错误：
```
ERROR: relation "products" does not exist
```
这是因为数据库中没有 products 表，属于功能未实现，不是路由问题。

## 验证

```bash
# 测试热门信件（现在应该正常）
curl "http://localhost:3000/api/v1/letters/popular?period=weekly&limit=6"

# 测试商品列表（会返回 500，因为表不存在）
curl "http://localhost:3000/api/v1/shop/products"
```

## 建议
1. 刷新浏览器页面以加载最新代码
2. Shop 功能的 500 错误需要后端实现 products 表才能解决
3. 所有信件相关的 API 现在应该正常工作