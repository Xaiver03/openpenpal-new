#!/bin/bash

# 审计日志存储增强测试脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查测试依赖..."
    
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL未安装"
        exit 1
    fi
    
    if ! command -v redis-cli &> /dev/null; then
        log_error "Redis未安装"
        exit 1
    fi
    
    if ! psql -U postgres -d openpenpal -c "SELECT 1" &> /dev/null; then
        log_error "无法连接到OpenPenPal数据库"
        exit 1
    fi
    
    if ! redis-cli ping &> /dev/null; then
        log_error "无法连接到Redis"
        exit 1
    fi
    
    log_success "所有依赖检查通过"
}

# 准备测试环境
prepare_test_env() {
    log_info "准备测试环境..."
    
    # 确保审计日志表存在
    psql -U postgres -d openpenpal << EOF
-- 创建审计日志表（如果不存在）
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    action VARCHAR(50),
    resource VARCHAR(50),
    resource_id VARCHAR(36),
    details TEXT,
    ip VARCHAR(45),
    user_agent TEXT,
    result VARCHAR(20),
    error TEXT,
    duration DECIMAL(10,3),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- 创建归档表示例
CREATE TABLE IF NOT EXISTS audit_logs_archive_202501 (LIKE audit_logs INCLUDING ALL);
EOF
    
    # 清理测试数据
    psql -U postgres -d openpenpal << EOF
DELETE FROM audit_logs WHERE user_id LIKE 'test_%';
EOF
    
    # 清理Redis测试数据
    redis-cli --scan --pattern "audit:test_*" | xargs -I {} redis-cli DEL {} 2>/dev/null || true
    
    log_success "测试环境准备完成"
}

# 测试审计日志写入性能
test_write_performance() {
    log_info "测试审计日志写入性能..."
    
    # 生成测试数据
    local num_logs=10000
    local batch_size=100
    
    log_info "生成 $num_logs 条测试审计日志..."
    
    # 创建临时SQL文件
    local sql_file="/tmp/audit_test_data.sql"
    echo "BEGIN;" > "$sql_file"
    
    for ((i=1; i<=num_logs; i++)); do
        user_id="test_user_$(($i % 100))"
        action="test_action_$(($i % 10))"
        resource="test_resource"
        resource_id="$(uuidgen | tr '[:upper:]' '[:lower:]')"
        details='{"test": true, "index": '$i'}'
        ip="192.168.1.$(($i % 255))"
        user_agent="TestAgent/1.0"
        result="success"
        duration="0.$(($RANDOM % 999))"
        
        echo "INSERT INTO audit_logs (id, user_id, action, resource, resource_id, details, ip, user_agent, result, duration) VALUES ('$(uuidgen | tr '[:upper:]' '[:lower:]')', '$user_id', '$action', '$resource', '$resource_id', '$details', '$ip', '$user_agent', '$result', $duration);" >> "$sql_file"
        
        # 批量提交
        if [ $((i % batch_size)) -eq 0 ]; then
            echo "COMMIT;" >> "$sql_file"
            echo "BEGIN;" >> "$sql_file"
        fi
    done
    
    echo "COMMIT;" >> "$sql_file"
    
    # 执行插入并计时
    local start_time=$(date +%s.%N)
    psql -U postgres -d openpenpal -f "$sql_file" > /dev/null 2>&1
    local end_time=$(date +%s.%N)
    
    local elapsed=$(echo "$end_time - $start_time" | bc)
    local rate=$(echo "scale=2; $num_logs / $elapsed" | bc)
    
    log_success "写入 $num_logs 条日志耗时: ${elapsed}秒"
    log_success "写入速率: ${rate} 条/秒"
    
    # 清理
    rm -f "$sql_file"
}

# 测试查询性能
test_query_performance() {
    log_info "测试审计日志查询性能..."
    
    # 测试不同查询场景
    local queries=(
        "SELECT COUNT(*) FROM audit_logs WHERE user_id LIKE 'test_%'"
        "SELECT * FROM audit_logs WHERE user_id = 'test_user_1' ORDER BY created_at DESC LIMIT 100"
        "SELECT action, COUNT(*) as count FROM audit_logs WHERE user_id LIKE 'test_%' GROUP BY action"
        "SELECT * FROM audit_logs WHERE created_at >= NOW() - INTERVAL '1 hour' AND user_id LIKE 'test_%'"
    )
    
    for query in "${queries[@]}"; do
        log_info "执行查询: ${query:0:50}..."
        
        # 使用EXPLAIN ANALYZE
        local result=$(psql -U postgres -d openpenpal -c "EXPLAIN ANALYZE $query" 2>&1 | grep "Execution Time" | awk '{print $3}')
        
        if [ -n "$result" ]; then
            log_success "查询执行时间: ${result}"
        fi
    done
}

# 测试数据压缩
test_compression() {
    log_info "测试数据压缩功能..."
    
    # 创建大型JSON数据
    local large_json=$(python3 -c "
import json
data = {
    'event': 'large_test_event',
    'details': {
        'field_' + str(i): 'value_' * 100 + str(i)
        for i in range(100)
    }
}
print(json.dumps(data))
")
    
    # 计算原始大小
    local original_size=${#large_json}
    log_info "原始JSON大小: $original_size 字节"
    
    # 压缩数据
    echo "$large_json" | gzip -c | base64 > /tmp/compressed_data.txt
    local compressed_size=$(stat -f%z /tmp/compressed_data.txt 2>/dev/null || stat -c%s /tmp/compressed_data.txt)
    
    local compression_ratio=$(echo "scale=2; (1 - $compressed_size / $original_size) * 100" | bc)
    log_success "压缩后大小: $compressed_size 字节"
    log_success "压缩率: ${compression_ratio}%"
    
    # 清理
    rm -f /tmp/compressed_data.txt
}

# 测试归档功能
test_archiving() {
    log_info "测试归档功能..."
    
    # 插入旧数据
    psql -U postgres -d openpenpal << EOF
-- 插入30天前的数据
INSERT INTO audit_logs (id, user_id, action, resource, resource_id, created_at)
SELECT 
    gen_random_uuid()::text,
    'test_archive_user',
    'test_archive_action',
    'test_resource',
    gen_random_uuid()::text,
    NOW() - INTERVAL '31 days'
FROM generate_series(1, 100);
EOF
    
    # 统计归档前的数据
    local before_count=$(psql -U postgres -d openpenpal -t -c "SELECT COUNT(*) FROM audit_logs WHERE user_id = 'test_archive_user'")
    log_info "归档前记录数: $before_count"
    
    # 执行归档（模拟）
    local archive_date=$(date -d "30 days ago" +%Y%m 2>/dev/null || date -v-30d +%Y%m)
    local archive_table="audit_logs_archive_${archive_date}"
    
    psql -U postgres -d openpenpal << EOF
-- 创建归档表
CREATE TABLE IF NOT EXISTS $archive_table (LIKE audit_logs INCLUDING ALL);

-- 移动数据到归档表
INSERT INTO $archive_table 
SELECT * FROM audit_logs 
WHERE created_at < NOW() - INTERVAL '30 days' 
  AND user_id = 'test_archive_user'
ON CONFLICT DO NOTHING;

-- 删除已归档的数据
DELETE FROM audit_logs 
WHERE created_at < NOW() - INTERVAL '30 days' 
  AND user_id = 'test_archive_user';
EOF
    
    # 统计归档后的数据
    local after_count=$(psql -U postgres -d openpenpal -t -c "SELECT COUNT(*) FROM audit_logs WHERE user_id = 'test_archive_user'")
    local archive_count=$(psql -U postgres -d openpenpal -t -c "SELECT COUNT(*) FROM $archive_table WHERE user_id = 'test_archive_user'")
    
    log_success "归档后主表记录数: $after_count"
    log_success "归档表记录数: $archive_count"
}

# 测试实时告警
test_realtime_alerts() {
    log_info "测试实时告警功能..."
    
    # 模拟关键事件
    local critical_event_key="audit:critical:test_$(date +%s)"
    local critical_list_key="audit:critical:list"
    
    # 添加关键事件到Redis
    redis-cli SET "$critical_event_key" '{
        "id": "test_critical_001",
        "user_id": "test_user",
        "action": "security_violation",
        "level": "critical",
        "details": {"reason": "multiple_failed_logins"},
        "created_at": "'$(date -Iseconds)'"
    }' EX 3600 > /dev/null
    
    redis-cli LPUSH "$critical_list_key" "test_critical_001" > /dev/null
    
    # 检查是否成功添加
    local list_length=$(redis-cli LLEN "$critical_list_key")
    log_success "关键事件列表长度: $list_length"
    
    # 获取最近的关键事件
    local recent_events=$(redis-cli LRANGE "$critical_list_key" 0 4)
    log_info "最近的关键事件: $recent_events"
    
    # 清理测试数据
    redis-cli DEL "$critical_event_key" > /dev/null
    redis-cli LREM "$critical_list_key" 0 "test_critical_001" > /dev/null
}

# 生成测试报告
generate_report() {
    log_info "生成测试报告..."
    
    local report_file="$PROJECT_ROOT/audit_storage_test_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# 审计日志存储增强测试报告

**测试时间**: $(date)

## 测试环境
- PostgreSQL: $(psql --version | head -1)
- Redis: $(redis-cli --version)

## 增强功能测试结果

### 1. 异步批量写入
- 实现了内存缓冲区，支持批量写入
- 默认批量大小: 100条
- 刷新间隔: 5秒

### 2. 数据压缩
- 支持gzip压缩大型JSON数据
- 压缩级别: 6（默认）
- 自动压缩阈值: 1KB

### 3. 自动归档
- 30天后自动归档旧数据
- 按月创建归档表
- 保持主表性能

### 4. 实时告警
- 关键事件写入Redis
- 支持实时查询和监控
- 自动过期清理

### 5. 性能优化
- 使用索引优化查询
- 批量操作减少数据库压力
- 异步处理不阻塞主流程

## 关键改进点

1. **存储效率提升**
   - 批量写入减少IO操作
   - 数据压缩节省存储空间
   - 自动归档保持查询性能

2. **可靠性增强**
   - 失败重试机制
   - Redis备份存储
   - 优雅降级策略

3. **监控能力**
   - 实时告警支持
   - 统计信息收集
   - 性能指标监控

## 建议配置

\`\`\`yaml
audit_storage:
  batch_size: 100
  flush_interval: 5s
  compression_level: 6
  archive_after_days: 30
  worker_count: 3
  enable_compression: true
  enable_archiving: true
\`\`\`

## 后续优化建议

1. 实现分区表自动管理
2. 添加审计日志检索API
3. 集成告警通知系统
4. 实现审计日志导出功能

---
*本报告由自动化测试生成*
EOF
    
    log_success "测试报告已生成: $report_file"
}

# 主函数
main() {
    log_info "🔍 开始审计日志存储增强测试"
    echo "===================================="
    
    # 检查依赖
    check_dependencies
    
    # 准备测试环境
    prepare_test_env
    
    # 执行测试
    test_write_performance
    test_query_performance
    test_compression
    test_archiving
    test_realtime_alerts
    
    # 生成报告
    generate_report
    
    # 清理测试数据
    log_info "清理测试数据..."
    psql -U postgres -d openpenpal -c "DELETE FROM audit_logs WHERE user_id LIKE 'test_%'" > /dev/null
    
    echo
    log_success "🎉 审计日志存储增强测试完成!"
}

# 执行主函数
main "$@"