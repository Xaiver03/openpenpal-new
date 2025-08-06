# API文档

本目录包含OpenPenPal项目的所有API接口文档和规范。

## 📡 API规范

### 统一规范
- **[unified-specification.md](./unified-specification.md)** - 完整的API设计规范和接口定义 (v2.0)

### 服务接口
- **[写信服务](../../services/write-service/README.md)**: 处理信件创作、Plaza、博物馆功能
- **[信使服务](../../services/courier-service/README.md)**: 4级信使管理和任务分配系统
- **[管理服务](../../services/admin-service/README.md)**: 系统管理和用户权限控制
- **[OCR服务](../../services/ocr-service/README.md)**: 图像识别和扫码功能

## 🔐 认证机制

所有API请求都需要通过身份认证，支持以下认证方式：
- JWT Token认证
- Session认证
- API Key认证

## 📖 使用指南

### 快速开始
1. 查看 [unified-specification.md](./unified-specification.md) 了解整体API架构
2. 参考具体服务的接口文档
3. 使用测试账号进行接口测试

### 请求格式
```json
{
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer <token>"
  },
  "body": {
    "data": "request_data"
  }
}
```

### 响应格式
```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "timestamp": "2025-01-23T10:00:00Z"
}
```

## 🔗 相关链接

- [开发文档](../development/) - 开发环境配置
- [测试账号](../getting-started/test-accounts.md) - API测试账号
- [故障排查](../troubleshooting/) - 常见问题解决

---

**最后更新**: 2025-01-23  
**维护**: OpenPenPal API团队