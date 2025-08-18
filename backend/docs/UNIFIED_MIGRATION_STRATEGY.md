# OpenPenPal 统一数据库迁移策略

## 概述

本文档描述了OpenPenPal项目的统一数据库迁移策略，该策略整合了多个现有的迁移方法，提供了一个可靠、可扩展、可回滚的迁移管理系统。

## 架构设计

### 核心组件

1. **统一迁移策略 (UnifiedMigrationStrategy)**
   - 位置: `internal/config/unified_migration_strategy.go`
   - 功能: 提供分阶段、可回滚的迁移执行计划
   - 特性: 预检查、备份、验证、优化

2. **迁移协调器 (MigrationCoordinator)**
   - 位置: `internal/config/migration_coordinator.go`
   - 功能: 跨服务迁移协调和依赖管理
   - 特性: 依赖分析、并行执行、状态监控

3. **统一迁移CLI工具**
   - 位置: `cmd/unified-migration/main.go`
   - 功能: 命令行界面，支持多种迁移策略
   - 特性: 干运行、回滚、详细日志

### 迁移策略层次

```
├── 统一迁移策略 (推荐)
│   ├── 预检查阶段
│   ├── 备份阶段
│   ├── 核心模型迁移
│   ├── 扩展模型迁移
│   ├── 数据完整性检查
│   ├── 索引优化
│   ├── 约束和触发器
│   ├── 视图和函数
│   ├── 性能优化
│   └── 最终验证
│
├── 协调迁移策略
│   ├── 依赖关系分析
│   ├── 预检查
│   ├── 按依赖顺序迁移
│   ├── 后迁移验证
│   └── 共享包集成
│
├── 安全迁移策略
│   ├── SafeAutoMigrate
│   └── 约束冲突处理
│
└── 扩展迁移策略
    ├── 核心模型
    └── 扩展模型
```

## 使用指南

### 命令行工具

#### 基本用法

```bash
# 完整统一迁移（推荐）
go run cmd/unified-migration/main.go --strategy=unified

# 干运行模式（预览更改）
go run cmd/unified-migration/main.go --dry-run

# 协调迁移（跨服务）
go run cmd/unified-migration/main.go --strategy=coordinated

# 安全迁移（仅核心模型）
go run cmd/unified-migration/main.go --strategy=safe

# 扩展迁移（核心+扩展模型）
go run cmd/unified-migration/main.go --strategy=extended
```

#### 高级选项

```bash
# 连接到生产数据库
go run cmd/unified-migration/main.go \
  --host=prod-db.example.com \
  --user=admin \
  --password=secure_password \
  --database=openpenpal_prod \
  --ssl=require

# 跳过性能优化
go run cmd/unified-migration/main.go --skip-optimization

# 启用详细日志
go run cmd/unified-migration/main.go --verbose

# 回滚模式（实验性）
go run cmd/unified-migration/main.go --rollback
```

### 编程接口

#### 统一迁移策略

```go
import "openpenpal-backend/internal/config"

// 创建迁移策略
opts := &config.MigrationOptions{
    DryRun:              false,
    RollbackMode:        false,
    SkipOptimizations:   false,
    ConcurrentIndexes:   true,
    Timeout:             30 * time.Minute,
    FailureStrategy:     "stop",
    BackupBeforeMigrate: true,
}

strategy := config.NewMigrationStrategy(db, dbConfig, opts)
err := strategy.ExecuteUnifiedMigration()
```

#### 迁移协调器

```go
// 创建协调器
coordinator := config.NewMigrationCoordinator(db, dbConfig)

// 执行协调迁移
err := coordinator.ExecuteCoordinatedMigration()

// 获取状态
status := coordinator.GetStatus()
fmt.Printf("Progress: %.2f%%\n", status.ProgressPercent)
```

## 迁移步骤详解

### 1. 预检查阶段 (Pre-check Phase)

- **数据库连接验证**: 确保数据库可访问
- **PostgreSQL版本检查**: 验证版本兼容性
- **扩展可用性检查**: 检查必需的PostgreSQL扩展
- **磁盘空间评估**: 评估可用磁盘空间
- **权限验证**: 确保用户具有必要权限

### 2. 备份阶段 (Backup Phase)

- **备份建议**: 提供pg_dump命令建议
- **元数据备份**: 备份模式信息
- **配置快照**: 记录当前配置状态

### 3. 核心模型迁移 (Core Models Migration)

- **SafeAutoMigrate**: 使用安全迁移处理约束冲突
- **模型覆盖**: 迁移所有核心业务模型
- **约束处理**: 智能处理外键约束冲突
- **数据保护**: 确保现有数据不丢失

### 4. 扩展模型迁移 (Extended Models Migration)

- **博物馆扩展**: 迁移博物馆相关扩展模型
- **信件扩展**: 迁移信件系统扩展功能
- **用户扩展**: 迁移用户系统扩展属性
- **特殊处理**: 处理复杂的模型关系

### 5. 数据完整性检查 (Data Integrity Check)

- **表存在性验证**: 确保所有必需表已创建
- **外键约束检查**: 验证外键约束完整性
- **数据一致性验证**: 检查数据逻辑一致性
- **记录统计**: 统计迁移后的数据量

### 6. 索引优化 (Index Optimization)

- **性能索引**: 创建优化查询性能的索引
- **并发创建**: 使用CONCURRENTLY避免锁表
- **复合索引**: 创建多列复合索引
- **哈希索引**: 为等值查询创建哈希索引

### 7. 约束和触发器 (Constraints and Triggers)

- **数据约束**: 添加业务逻辑约束
- **触发器**: 创建审计和自动化触发器
- **规则**: 设置数据验证规则

### 8. 视图和函数 (Views and Functions)

- **物化视图**: 创建性能优化的物化视图
- **监控视图**: 创建系统监控视图
- **存储函数**: 创建常用业务函数
- **扩展启用**: 启用pg_stat_statements等扩展

### 9. 性能优化 (Performance Optimization)

- **统计信息刷新**: 更新表统计信息
- **查询计划优化**: 优化常用查询的执行计划
- **内存配置**: 调整内存相关参数
- **监控设置**: 设置性能监控

### 10. 最终验证 (Final Validation)

- **功能测试**: 验证关键功能正常
- **性能测试**: 验证性能指标
- **连接测试**: 验证应用连接正常
- **健康检查**: 全面的系统健康检查

## 故障恢复

### 回滚策略

1. **自动回滚**: 在检测到致命错误时自动回滚
2. **手动回滚**: 提供手动回滚命令
3. **部分回滚**: 支持回滚特定步骤
4. **快照恢复**: 从备份快照恢复

### 错误处理

1. **继续策略**: 非关键错误继续执行
2. **停止策略**: 关键错误立即停止
3. **重试机制**: 自动重试临时性错误
4. **详细日志**: 记录详细的错误信息

## 监控和报告

### 进度监控

```bash
# 实时查看迁移进度
tail -f migration.log

# 检查特定服务状态
curl http://localhost:8080/migration/status/backend
```

### 迁移报告

迁移完成后会生成详细报告，包括：

- 执行时间统计
- 成功/失败步骤汇总
- 性能指标对比
- 优化建议
- 维护计划

## 最佳实践

### 迁移前准备

1. **完整备份**: 始终在迁移前创建完整备份
2. **测试环境**: 在测试环境先验证迁移过程
3. **维护窗口**: 在低峰期执行迁移
4. **资源准备**: 确保足够的CPU、内存、磁盘资源

### 迁移过程中

1. **监控日志**: 密切关注迁移日志
2. **资源监控**: 监控数据库资源使用情况
3. **性能监控**: 监控查询性能变化
4. **备用计划**: 准备回滚计划

### 迁移后验证

1. **功能测试**: 验证所有功能正常工作
2. **性能验证**: 确认性能满足要求
3. **数据验证**: 验证数据完整性和一致性
4. **监控设置**: 设置持续监控

## 维护计划

### 定期维护任务

```sql
-- 每小时: 刷新物化视图
SELECT refresh_materialized_views();

-- 每日: 清理软删除数据（30天以上）
SELECT cleanup_soft_deleted_data();

-- 每周: 更新表统计信息
ANALYZE;

-- 每月: 清理和重建索引
VACUUM ANALYZE;
REINDEX DATABASE openpenpal;
```

### 性能监控查询

```sql
-- 查看慢查询
SELECT * FROM slow_queries LIMIT 10;

-- 查看表大小
SELECT * FROM table_sizes;

-- 查看连接状态
SELECT * FROM connection_monitoring;

-- 查看锁状态
SELECT * FROM lock_monitoring;
```

## 故障排除

### 常见问题

1. **连接超时**: 检查网络连接和防火墙
2. **权限不足**: 确保用户有必要的数据库权限
3. **磁盘空间不足**: 清理临时文件或扩展磁盘
4. **版本不兼容**: 升级PostgreSQL版本
5. **约束冲突**: 检查数据一致性

### 调试技巧

1. **详细日志**: 使用--verbose参数
2. **干运行**: 使用--dry-run预览
3. **分步执行**: 使用不同的策略分步执行
4. **数据库日志**: 检查PostgreSQL日志

## 版本兼容性

- **PostgreSQL**: 12+ (推荐 14+)
- **Go**: 1.19+ 
- **GORM**: v1.25+
- **操作系统**: Linux, macOS, Windows

## 安全考虑

1. **密码安全**: 避免在命令行明文传递密码
2. **SSL连接**: 生产环境使用SSL
3. **权限最小化**: 使用最小必要权限
4. **审计日志**: 启用数据库审计日志

## 贡献指南

如需改进迁移策略，请：

1. 在测试环境充分测试
2. 添加适当的错误处理
3. 更新文档
4. 添加单元测试
5. 提交Pull Request

## 联系支持

如遇到问题，请：

1. 查看此文档
2. 检查GitHub Issues
3. 提交详细的错误报告
4. 包含迁移日志和系统信息

---

*最后更新: 2025-08-18*
*版本: 1.0.0*