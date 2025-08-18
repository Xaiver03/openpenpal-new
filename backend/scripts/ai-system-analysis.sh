#!/bin/bash

# OpenPenPal AI系统全面分析脚本
# 用于系统性检测和分析AI服务的各个组件

set -e

PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"
BACKEND_DIR="$PROJECT_ROOT/backend"
ANALYSIS_LOG="$PROJECT_ROOT/ai-system-analysis.log"

echo "🧠 OpenPenPal AI系统全面分析"
echo "================================"
echo "开始时间: $(date)"
echo ""

# 清理之前的日志
> "$ANALYSIS_LOG"

log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$ANALYSIS_LOG"
}

log "🔍 第一阶段：AI系统架构分析"
echo "========================================="

# 1. 检查AI相关文件完整性
log "检查AI核心文件..."
AI_FILES=(
    "internal/models/ai.go"
    "internal/services/ai_service.go"
    "internal/services/ai_provider_interface.go"
    "internal/services/ai_provider_manager.go"
    "internal/services/ai_provider_openai.go"
    "internal/services/ai_provider_claude.go"
    "internal/services/ai_provider_moonshot.go"
    "internal/services/ai_provider_local.go"
    "internal/handlers/ai_handler.go"
    "internal/routes/ai_routes.go"
    "internal/services/delay_queue_service_fixed.go"
)

missing_files=0
for file in "${AI_FILES[@]}"; do
    if [ -f "$BACKEND_DIR/$file" ]; then
        log "✅ $file"
    else
        log "❌ $file (缺失)"
        ((missing_files++))
    fi
done

log ""
log "📊 第二阶段：AI提供商配置分析"
echo "========================================="

# 2. 检查数据库AI配置
log "查询数据库AI配置..."
cd "$BACKEND_DIR"

DB_QUERY_RESULT=$(psql "postgres://openpenpal_user@localhost:5432/openpenpal?sslmode=disable" -c "
SELECT 
    provider,
    model,
    is_active,
    priority,
    daily_quota,
    used_quota,
    api_endpoint
FROM ai_configs 
WHERE provider IS NOT NULL AND provider != '' 
ORDER BY priority DESC, provider;
" -t 2>/dev/null || echo "数据库查询失败")

log "AI提供商配置状态:"
echo "$DB_QUERY_RESULT" | while read line; do
    if [[ -n "$line" && "$line" != *"数据库查询失败"* ]]; then
        log "  $line"
    fi
done

# 3. 检查环境变量配置
log ""
log "检查AI API密钥配置..."
if [ -f "$BACKEND_DIR/.env" ]; then
    log "环境变量状态:"
    grep -E "(AI_|OPENAI|CLAUDE|MOONSHOT)" "$BACKEND_DIR/.env" | while read var; do
        key=$(echo "$var" | cut -d'=' -f1)
        value=$(echo "$var" | cut -d'=' -f2)
        if [[ -n "$value" ]]; then
            log "  ✅ $key: 已配置"
        else
            log "  ⚠️  $key: 未配置"
        fi
    done
else
    log "❌ .env文件不存在"
fi

log ""
log "🔧 第三阶段：AI服务编译测试"
echo "========================================="

# 4. 编译测试
log "测试AI服务编译..."
if go vet ./internal/services/ai_*.go 2>/dev/null; then
    log "✅ AI服务代码验证通过"
else
    log "⚠️  AI服务代码存在潜在问题"
fi

# 编译测试主要AI组件
log "测试核心AI组件编译..."
AI_COMPONENTS=(
    "internal/services/ai_service.go"
    "internal/services/ai_provider_manager.go"
    "internal/handlers/ai_handler.go"
)

for component in "${AI_COMPONENTS[@]}"; do
    if go build -o /tmp/ai_test_$(basename "$component" .go) "$component" 2>/dev/null; then
        log "✅ $(basename "$component") 编译成功"
        rm -f "/tmp/ai_test_$(basename "$component" .go)"
    else
        log "❌ $(basename "$component") 编译失败"
    fi
done

log ""
log "🎯 第四阶段：AI API端点检查"
echo "========================================="

# 5. 检查AI路由和处理器
log "分析AI API端点..."
if [ -f "$BACKEND_DIR/internal/routes/ai_routes.go" ]; then
    ENDPOINTS=$(grep -E "(POST|GET|PUT|DELETE)" "$BACKEND_DIR/internal/routes/ai_routes.go" | grep -o '"/[^"]*"' | sort -u)
    log "发现的AI API端点:"
    echo "$ENDPOINTS" | while read endpoint; do
        log "  📍 $endpoint"
    done
else
    log "❌ AI路由文件不存在"
fi

log ""
log "⚙️  第五阶段：AI任务处理分析"
echo "========================================="

# 6. 检查延迟队列AI任务
log "检查AI任务队列状态..."
if command -v redis-cli >/dev/null 2>&1; then
    if redis-cli ping >/dev/null 2>&1; then
        OLD_QUEUE_SIZE=$(redis-cli ZCARD delay_queue 2>/dev/null || echo "0")
        NEW_QUEUE_SIZE=$(redis-cli ZCARD delay_queue_fixed 2>/dev/null || echo "0")
        log "旧AI任务队列: $OLD_QUEUE_SIZE 个任务"
        log "新AI任务队列: $NEW_QUEUE_SIZE 个任务"
        
        # 检查任务类型
        if [ "$NEW_QUEUE_SIZE" -gt 0 ]; then
            log "检查队列中的AI任务类型..."
            redis-cli ZRANGE delay_queue_fixed 0 -1 | head -3 | while read task; do
                if echo "$task" | grep -q "ai_reply"; then
                    log "  🤖 发现AI回信任务"
                fi
            done
        fi
    else
        log "⚠️  Redis服务未运行"
    fi
else
    log "⚠️  Redis客户端未安装"
fi

log ""
log "📈 第六阶段：AI使用统计分析"
echo "========================================="

# 7. 检查AI使用日志
log "查询AI使用统计..."
AI_USAGE_QUERY=$(psql "postgres://openpenpal_user@localhost:5432/openpenpal?sslmode=disable" -c "
SELECT 
    task_type,
    provider,
    COUNT(*) as usage_count,
    AVG(response_time) as avg_response_time,
    SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count
FROM ai_usage_logs 
GROUP BY task_type, provider
ORDER BY usage_count DESC;
" -t 2>/dev/null || echo "无法查询AI使用统计")

if [[ "$AI_USAGE_QUERY" != *"无法查询"* ]]; then
    log "AI使用统计:"
    echo "$AI_USAGE_QUERY" | while read line; do
        if [[ -n "$line" ]]; then
            log "  $line"
        fi
    done
else
    log "⚠️  AI使用统计表可能不存在或无数据"
fi

log ""
log "🚨 第七阶段：AI系统健康检查"
echo "========================================="

# 8. 系统健康状态汇总
log "生成AI系统健康报告..."

# 统计问题
total_files=${#AI_FILES[@]}
healthy_files=$((total_files - missing_files))
health_score=$((healthy_files * 100 / total_files))

log ""
log "🎯 AI系统分析总结"
log "=================="
log "文件完整性: $healthy_files/$total_files ($health_score%)"
log "编译状态: 已测试主要组件"
log "数据库配置: 已检查AI配置表"
log "环境变量: 已验证API密钥配置"
log "任务队列: 已检查Redis队列状态"
log "使用统计: 已查询AI使用数据"

log ""
log "📋 建议优化项:"
if [ $missing_files -gt 0 ]; then
    log "1. 修复缺失的AI组件文件"
fi
log "2. 验证AI API密钥的有效性"
log "3. 测试AI提供商的实际响应"
log "4. 监控AI任务处理性能"
log "5. 实施AI服务的端到端测试"

log ""
log "分析完成时间: $(date)"
log "详细日志已保存至: $ANALYSIS_LOG"

echo ""
echo "🎉 AI系统分析完成！"
echo "📄 完整报告: $ANALYSIS_LOG"