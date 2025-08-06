# 部署运维指南

本目录包含OpenPenPal项目的部署和运维相关文档。

## 🚀 部署文档

### 环境部署
- [Docker部署指南](./docker-guide.md) - 使用Docker进行项目部署
- [生产环境部署](./production.md) - 生产环境部署配置和注意事项

### 运维管理
- [监控配置](./monitoring.md) - 系统监控和日志配置
- [故障排查](./troubleshooting.md) - 常见问题和解决方案

## 🔧 快速部署

### 开发环境
```bash
# 克隆项目
git clone [project-url]
cd openpenpal

# 启动开发环境
docker-compose up -d
```

### 生产环境
参考 [生产环境部署](./production.md) 文档进行详细配置。

## 📊 运维工具

- **Docker**: 容器化部署
- **Docker Compose**: 多服务编排
- **脚本工具**: 参考 [操作脚本](../operations/scripts-usage.md)

## 🆘 故障处理

遇到问题时请按以下顺序排查：

1. 检查 [故障排查文档](./troubleshooting.md)
2. 查看 [系统日志](./monitoring.md)
3. 参考 [开发文档](../development/)

---

**最后更新**: 2025-01-23  
**维护**: OpenPenPal运维团队