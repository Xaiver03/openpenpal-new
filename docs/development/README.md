# 开发文档

本目录包含OpenPenPal项目的开发指南、代码规范和最佳实践。

## 💻 开发指南

### 核心文档
- [coding-standards.md](./coding-standards.md) - 编码标准和最佳实践
- [component-management.md](./component-management.md) - 前端组件开发规范
- [file-naming.md](./file-naming.md) - 文件和目录命名标准
- [roadmap.md](./roadmap.md) - 项目开发路线图
- [development-guide-zh.md](./development-guide-zh.md) - 中文开发指南

## 🏗️ 开发环境

### 前端开发
- **框架**: Next.js 14 with App Router
- **样式**: Tailwind CSS + shadcn/ui
- **状态管理**: Zustand
- **类型检查**: TypeScript

### 后端开发
- **Go服务**: 高性能微服务
- **Python服务**: AI和数据处理
- **Java服务**: 企业级业务逻辑
- **数据库**: PostgreSQL

## 📋 开发规范

### 代码质量
1. **代码审查**: 所有代码变更需要通过审查
2. **测试覆盖**: 核心功能需要单元测试
3. **文档更新**: 代码变更后及时更新文档
4. **性能考虑**: 注意代码性能和资源使用

### Git工作流
1. **分支管理**: 使用 feature/bugfix 分支
2. **提交规范**: 遵循 Conventional Commits
3. **代码同步**: 定期同步主分支变更

## 🔗 相关链接

- [API文档](../api/) - 接口开发规范
- [架构文档](../architecture/) - 系统架构设计
- [测试账号](../getting-started/test-accounts.md) - 开发测试用账号

---

**最后更新**: 2025-01-23  
**维护**: OpenPenPal开发团队