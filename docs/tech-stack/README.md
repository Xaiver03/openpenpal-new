# 技术栈文档

本目录用于存放 OpenPenPal 项目的技术栈详细文档。

## 📚 技术栈概览

### 前端技术
- **框架**: Next.js 14 with App Router
- **语言**: TypeScript
- **样式**: Tailwind CSS + shadcn/ui
- **状态管理**: Zustand
- **组件库**: Radix UI

### 后端技术
- **主服务**: Go 1.21 + Gin + GORM
- **微服务**:
  - 写信服务: Python + FastAPI
  - 信使服务: Go + Gin
  - 管理服务: Java + Spring Boot
  - OCR服务: Python + OpenCV

### 基础设施
- **数据库**: PostgreSQL + Redis
- **容器化**: Docker + Docker Compose
- **监控**: Prometheus + Grafana
- **日志**: 结构化日志

## 📋 技术栈详细文档
- [web-first-tech-stack.md](./web-first-tech-stack.md) - Web优先技术栈建议

## 🔗 相关文档
- [系统架构](../architecture/)
- [开发指南](../development/)
- [API文档](../api/)

---

**最后更新**: 2025-01-23  
**状态**: 待完善