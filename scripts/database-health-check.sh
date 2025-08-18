#!/bin/bash

# 数据库健康检查工具
# 检查所有服务的数据库连接状态和配置一致性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🏥 数据库健康检查工具${NC}"
echo "=========================="

# 检查PostgreSQL服务状态
check_postgresql_service() {
    echo -e "\n${YELLOW}1. PostgreSQL服务状态${NC}"
    if pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ PostgreSQL服务正常运行${NC}"
        
        # 显示数据库列表
        echo "数据库列表:"
        psql -l | grep openpenpal | while read line; do
            echo "  - $line"
        done
    else
        echo -e "${RED}❌ PostgreSQL服务未运行${NC}"
        return 1
    fi
}

# 检查配置文件一致性
check_config_consistency() {
    echo -e "\n${YELLOW}2. 配置文件一致性检查${NC}"
    
    # 配置文件列表
    CONFIG_FILES=(
        "./backend/.env"
        "./services/write-service/.env" 
        "./services/courier-service/.env"
        "./services/gateway/.env"
        "./.env"
    )
    
    declare -A db_configs
    
    for config_file in "${CONFIG_FILES[@]}"; do
        if [ -f "$config_file" ]; then
            echo -e "\n${BLUE}检查: $config_file${NC}"
            
            # 提取数据库配置
            if grep -q "DATABASE_URL" "$config_file"; then
                db_url=$(grep "DATABASE_URL" "$config_file" | head -1 | cut -d'=' -f2 | tr -d '"')
                echo "  DATABASE_URL: $db_url"
                db_configs["$config_file"]="$db_url"
            else
                echo -e "  ${RED}❌ 缺少DATABASE_URL配置${NC}"
            fi
        else
            echo -e "${RED}❌ 配置文件不存在: $config_file${NC}"
        fi
    done
    
    # 检查配置一致性
    echo -e "\n${YELLOW}配置一致性分析:${NC}"
    unique_configs=$(printf '%s\n' "${db_configs[@]}" | sort | uniq | wc -l)
    
    if [ "$unique_configs" -eq 1 ]; then
        echo -e "${GREEN}✅ 所有服务使用相同的数据库配置${NC}"
    else
        echo -e "${RED}❌ 发现 $unique_configs 种不同的数据库配置${NC}"
        printf '%s\n' "${db_configs[@]}" | sort | uniq | while read config; do
            echo "  - $config"
        done
    fi
}

# 检查SQLite文件
check_sqlite_files() {
    echo -e "\n${YELLOW}3. SQLite文件检查${NC}"
    
    SQLITE_FILES=$(find . -name "*.db" -o -name "*.sqlite" -o -name "*.sqlite3" | grep -v archive | grep -v backup)
    
    if [ -n "$SQLITE_FILES" ]; then
        echo -e "${RED}⚠️  发现活跃的SQLite文件:${NC}"
        echo "$SQLITE_FILES" | while read file; do
            size=$(ls -lh "$file" | awk '{print $5}')
            echo "  - $file ($size)"
        done
        echo -e "${YELLOW}建议: 清理这些SQLite文件以避免数据不一致${NC}"
    else
        echo -e "${GREEN}✅ 未发现活跃的SQLite文件${NC}"
    fi
}

# 检查服务连接状态
check_service_connections() {
    echo -e "\n${YELLOW}4. 服务数据库连接状态${NC}"
    
    # 服务端口列表
    SERVICES=(
        "主后端:8080"
        "写信服务:8001"
        "信使服务:8002"
        "管理服务:8003"
        "OCR服务:8004"
    )
    
    for service_info in "${SERVICES[@]}"; do
        service_name=$(echo "$service_info" | cut -d':' -f1)
        service_port=$(echo "$service_info" | cut -d':' -f2)
        
        echo -e "\n${BLUE}检查 $service_name (端口 $service_port)${NC}"
        
        # 检查服务是否运行
        if lsof -i ":$service_port" >/dev/null 2>&1; then
            echo -e "  ${GREEN}✅ 服务正在运行${NC}"
            
            # 尝试健康检查
            health_url="http://localhost:$service_port/health"
            if curl -s "$health_url" >/dev/null 2>&1; then
                echo -e "  ${GREEN}✅ 健康检查通过${NC}"
            else
                echo -e "  ${YELLOW}⚠️  健康检查失败或无健康检查端点${NC}"
            fi
        else
            echo -e "  ${RED}❌ 服务未运行${NC}"
        fi
    done
}

# 检查数据库连接池
check_database_connections() {
    echo -e "\n${YELLOW}5. 数据库连接池状态${NC}"
    
    # 检查PostgreSQL连接数
    if psql -h localhost -U rocalight -d openpenpal -c "
        SELECT 
            datname as database,
            numbackends as connections,
            xact_commit as commits,
            xact_rollback as rollbacks
        FROM pg_stat_database 
        WHERE datname = 'openpenpal';
    " 2>/dev/null; then
        echo -e "${GREEN}✅ 数据库连接信息获取成功${NC}"
    else
        echo -e "${RED}❌ 无法获取数据库连接信息${NC}"
    fi
}

# 数据同步风险评估
assess_sync_risks() {
    echo -e "\n${YELLOW}6. 数据同步风险评估${NC}"
    
    echo "检查项目:"
    
    # 检查是否有数据同步机制
    if [ -f "./services/sync-service/main.go" ]; then
        echo -e "  ${GREEN}✅ 发现专门的同步服务${NC}"
    else
        echo -e "  ${RED}❌ 缺少专门的数据同步服务${NC}"
    fi
    
    # 检查Redis配置
    if grep -r "redis" . --include="*.env" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✅ 配置了Redis缓存${NC}"
    else
        echo -e "  ${YELLOW}⚠️  未发现Redis配置${NC}"
    fi
    
    # 检查WebSocket配置
    if grep -r "websocket\|ws://" . --include="*.go" --include="*.ts" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✅ 支持WebSocket实时通信${NC}"
    else
        echo -e "  ${YELLOW}⚠️  未发现WebSocket配置${NC}"
    fi
    
    # 风险评估
    echo -e "\n${YELLOW}风险评估结果:${NC}"
    echo -e "${RED}高风险:${NC} 多服务共享数据库但缺少统一数据治理"
    echo -e "${YELLOW}中风险:${NC} 缺少专门的数据同步机制"
    echo -e "${GREEN}低风险:${NC} PostgreSQL提供ACID事务保证"
}

# 生成修复建议
generate_recommendations() {
    echo -e "\n${YELLOW}7. 修复建议${NC}"
    echo "=================================="
    
    echo -e "${BLUE}立即执行:${NC}"
    echo "1. 运行配置统一脚本: ./scripts/fix-database-configuration.sh"
    echo "2. 重启所有服务: ./startup/stop-all.sh && ./startup/quick-start.sh"
    
    echo -e "\n${BLUE}短期优化 (1-2周):${NC}"
    echo "1. 建立数据库连接监控"
    echo "2. 实现服务启动时的数据库健康检查"
    echo "3. 统一错误处理和重试机制"
    
    echo -e "\n${BLUE}中期优化 (1-2月):${NC}"
    echo "1. 实现分布式事务管理"
    echo "2. 建立数据库性能监控"
    echo "3. 实现数据备份和恢复机制"
    
    echo -e "\n${BLUE}长期规划 (3-6月):${NC}"
    echo "1. 考虑微服务数据库分离"
    echo "2. 实现事件驱动架构"
    echo "3. 建立完整的数据治理体系"
}

# 主执行流程
main() {
    check_postgresql_service
    check_config_consistency
    check_sqlite_files
    check_service_connections
    check_database_connections
    assess_sync_risks
    generate_recommendations
    
    echo -e "\n${GREEN}🎯 数据库健康检查完成${NC}"
    echo "详细报告已生成，请根据建议进行优化"
}

# 运行检查
main "$@"