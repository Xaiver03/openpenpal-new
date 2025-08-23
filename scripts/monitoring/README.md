# OpenPenPal 监控系统文档

## 概述

本监控系统是为解决OpenPenPal后端服务中的数据库连接池问题而设计的综合监控解决方案。系统主要针对HikariCP线程饥饿、连接泄漏、以及相关的性能问题提供实时监控和告警。

## 问题背景

### 原始问题
- **HikariPool线程饥饿**：Admin Service出现16-34分钟的延迟警告
- **连接池配置不当**：最大连接数过低(10)，空闲超时过长(10分钟)
- **缺乏监控**：无法及时发现和诊断连接池问题

### 解决方案
本监控系统提供了以下核心功能：
1. 实时数据库连接监控
2. HikariCP指标收集和告警
3. 慢查询检测和分析
4. 连接泄漏检测
5. 系统资源监控
6. 自动告警系统

## 目录结构

```
scripts/monitoring/
├── README.md                           # 本文档
├── database/                          # 数据库监控
│   ├── monitor-connections.sh         # 数据库连接监控
│   ├── monitor-slow-queries.sh        # 慢查询监控
│   └── configure-slow-query-logging.sh # 慢查询配置
└── alert-system.sh                    # 自动告警系统
```

## 核心组件

### 1. 数据库连接监控 (`database/monitor-connections.sh`)

**功能**：
- 连接状态统计（活跃/空闲/事务中）
- 应用程序连接分析
- 长时间运行查询检测
- 空闲连接分析
- 锁分析
- 连接池使用率计算

**使用方法**：
```bash
./scripts/monitoring/database/monitor-connections.sh
```

**输出示例**：
```
=== OpenPenPal Database Connection Monitor ===
Database: openpenpal | Time: 2025-08-20 09:42:39

1. Connection Summary:
 idle   |     9 | -00:00:00.033018
 active |     2 | -00:00:00.000006

6. Connection Pool Analysis:
Total Connections:     10
Max Connections:  100
Usage: 10.0%
Connection usage is healthy
```

### 2. HikariCP指标收集 (Java组件)

**位置**：`services/admin-service/backend/src/main/java/com/openpenpal/admin/monitoring/HikariMetrics.java`

**功能**：
- 每分钟记录连接池统计信息
- 实时告警检测（高使用率、线程等待、快速增长）
- Spring Boot Actuator健康检查集成
- JMX监控支持

**关键指标**：
- 活跃连接数
- 空闲连接数
- 等待线程数
- 池使用率
- 连接泄漏检测

**日志示例**：
```
POOL_STATS timestamp='2025-08-20 09:42:39' active=5 idle=3 total=8 waiting=0 
max_pool_size=30 utilization=26.7% active_percent=16.7% pool_name='OpenPenPal-AdminService-Pool'
```

### 3. 慢查询监控

**配置脚本**：`database/configure-slow-query-logging.sh`
**监控脚本**：`database/monitor-slow-queries.sh`

**功能**：
- 配置PostgreSQL慢查询日志（>1秒）
- 启用pg_stat_statements扩展
- 创建查询分析视图和函数
- 实时慢查询监控

**分析视图**：
```sql
-- 慢查询分析
SELECT * FROM slow_query_analysis;

-- 当前活动连接
SELECT * FROM connection_activity;

-- 分析最近慢查询
SELECT * FROM analyze_recent_slow_queries();
```

### 4. 连接泄漏检测

**配置**：`services/admin-service/backend/src/main/java/com/openpenpal/admin/config/ConnectionLeakDetectionConfig.java`

**功能**：
- 开发环境：30秒泄漏检测
- 生产环境：60秒泄漏检测
- 增强的HikariCP配置
- 连接验证和超时优化

**关键配置**：
```yaml
hikari:
  maximum-pool-size: 30          # 增加到30
  minimum-idle: 3               # 减少到3
  connection-timeout: 20000     # 20秒
  idle-timeout: 120000         # 2分钟
  leak-detection-threshold: 60000  # 60秒
```

### 5. 自动告警系统 (`alert-system.sh`)

**功能**：
- 数据库连接监控
- 长时间运行查询检测
- HikariCP健康检查
- 系统资源监控
- 服务可用性检查
- 日志文件大小检查

**告警阈值**：
- 总连接数 > 25
- 池使用率 > 80%
- 查询运行时间 > 30秒
- 空闲连接 > 10个（超过5分钟）
- 线程等待连接 > 0

**使用方法**：
```bash
# 单次监控
./scripts/monitoring/alert-system.sh monitor

# 持续监控（每5分钟）
./scripts/monitoring/alert-system.sh continuous

# 生成健康报告
./scripts/monitoring/alert-system.sh report
```

## JVM调优配置

**配置文件**：`services/admin-service/jvm-tuning.conf`

**关键优化**：
- G1GC垃圾收集器
- 最大GC暂停时间：200ms
- 堆内存：2GB
- 元空间：512MB
- GC日志和监控
- JMX远程监控

**应用方法**：
```bash
# 启动时应用JVM参数
java @jvm-tuning.conf -jar admin-service.jar
```

## Spring Boot Actuator集成

**健康检查端点**：
- `/actuator/health` - 总体健康状态
- `/actuator/health/hikaricp` - HikariCP池状态
- `/actuator/metrics` - 详细指标

**示例响应**：
```json
{
  "status": "UP",
  "components": {
    "hikaricp": {
      "status": "UP",
      "details": {
        "active_connections": 5,
        "idle_connections": 3,
        "total_connections": 8,
        "max_pool_size": 30,
        "utilization_percent": "26.7"
      }
    }
  }
}
```

## 监控工作流程

### 日常监控
1. **自动化监控**：运行`alert-system.sh continuous`
2. **连接状态检查**：每小时运行`monitor-connections.sh`
3. **慢查询分析**：每日运行`monitor-slow-queries.sh`

### 问题诊断
1. **连接池问题**：
   ```bash
   # 检查当前连接状态
   ./scripts/monitoring/database/monitor-connections.sh
   
   # 检查HikariCP健康
   curl http://localhost:8003/api/admin/actuator/health/hikaricp
   ```

2. **性能问题**：
   ```bash
   # 检查慢查询
   ./scripts/monitoring/database/monitor-slow-queries.sh
   
   # 分析查询性能
   psql -d openpenpal -c "SELECT * FROM slow_query_analysis;"
   ```

3. **连接泄漏**：
   ```bash
   # 检查长时间运行的连接
   psql -d openpenpal -c "SELECT * FROM connection_activity;"
   
   # 检查应用日志中的泄漏警告
   grep "leak" logs/admin-service.log
   ```

## 告警规则

### 严重告警 (CRITICAL)
- 数据库连接数超过阈值
- 服务不可用
- HikariCP池状态DOWN
- 系统资源耗尽（内存>90%，磁盘>90%）

### 警告告警 (WARNING)  
- 长时间运行查询
- 多个空闲连接
- 线程等待连接
- 系统资源高使用率（>80%）
- 大日志文件

### 信息 (INFO)
- 正常运行状态
- 定期健康检查
- 指标收集

## 故障排除

### 常见问题

1. **"Thread starvation or clock leap detected"**
   - **原因**：连接池耗尽或长时间GC暂停
   - **解决**：检查连接泄漏，应用JVM调优，增加池大小

2. **连接超时**
   - **原因**：连接池配置不当或数据库负载高
   - **解决**：优化查询，调整超时设置，检查网络

3. **慢查询**
   - **原因**：缺少索引或查询优化不当
   - **解决**：分析执行计划，添加索引，优化查询

### 调试命令

```bash
# 查看当前数据库活动
psql -d openpenpal -c "SELECT * FROM pg_stat_activity;"

# 检查锁等待
psql -d openpenpal -c "SELECT * FROM pg_locks WHERE NOT granted;"

# 查看连接池JMX指标
jconsole localhost:9999  # 连接到JMX端口
```

## 配置参考

### PostgreSQL配置
```sql
-- 慢查询日志
ALTER SYSTEM SET log_min_duration_statement = 1000;

-- 连接日志
ALTER SYSTEM SET log_connections = on;
ALTER SYSTEM SET log_disconnections = on;

-- 锁等待日志
ALTER SYSTEM SET log_lock_waits = on;

-- 重新加载配置
SELECT pg_reload_conf();
```

### HikariCP最佳实践
```yaml
spring:
  datasource:
    hikari:
      maximum-pool-size: 30
      minimum-idle: 3
      connection-timeout: 20000
      idle-timeout: 120000
      max-lifetime: 1800000
      leak-detection-threshold: 60000
      connection-test-query: "SELECT 1"
```

## 维护

### 定期任务
- **每日**：检查告警日志，运行健康报告
- **每周**：分析慢查询趋势，清理大日志文件
- **每月**：审查连接池配置，优化JVM参数

### 日志管理
- 告警日志：`logs/monitoring-alerts.log`
- 健康报告：`logs/health-report-*.txt`
- HikariCP日志：应用日志中的`METRICS.HikariCP`

## 扩展

### 添加新监控指标
1. 在`HikariMetrics.java`中添加新的度量
2. 更新`alert-system.sh`中的检查逻辑
3. 在健康检查端点中暴露新指标

### 集成外部监控
- Prometheus指标导出
- Grafana仪表板
- PagerDuty告警集成
- ELK Stack日志聚合

## 总结

本监控系统通过多层次的监控和告警机制，有效解决了OpenPenPal项目中的HikariCP连接池问题。系统提供了：

- **实时监控**：连接池状态、查询性能、系统资源
- **预防性告警**：提前发现潜在问题
- **诊断工具**：快速定位和解决问题
- **性能优化**：JVM调优和连接池配置优化

通过持续使用此监控系统，可以确保数据库连接池的稳定运行，避免线程饥饿等严重问题的再次发生。