#!/bin/bash

# 数据库配置统一修复脚本
# 解决PostgreSQL和SQLite并存以及配置不一致问题

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🛠️  数据库配置统一修复工具${NC}"
echo "=================================="

# 检查PostgreSQL是否运行
check_postgresql() {
    echo -e "\n${YELLOW}1. 检查PostgreSQL状态...${NC}"
    if pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ PostgreSQL正在运行${NC}"
        return 0
    else
        echo -e "${RED}❌ PostgreSQL未运行，请先启动PostgreSQL${NC}"
        return 1
    fi
}

# 备份当前配置
backup_configs() {
    echo -e "\n${YELLOW}2. 备份当前配置...${NC}"
    BACKUP_DIR="./archive/database_config_backup/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # 备份所有.env文件
    find . -name ".env" -type f | while read env_file; do
        if [ -f "$env_file" ]; then
            cp "$env_file" "$BACKUP_DIR/$(echo $env_file | sed 's/\//_/g')"
            echo "  - 备份: $env_file"
        fi
    done
    
    echo -e "${GREEN}✅ 配置已备份到: $BACKUP_DIR${NC}"
}

# 统一数据库配置
standardize_config() {
    echo -e "\n${YELLOW}3. 统一数据库配置...${NC}"
    
    # 标准配置参数
    DB_HOST="localhost"
    DB_PORT="5432"
    DB_NAME="openpenpal"
    DB_USER="rocalight"
    DB_PASSWORD="password"
    
    # 生成标准配置
    STANDARD_CONFIG="# OpenPenPal 统一数据库配置
# 生成时间: $(date)
DATABASE_TYPE=postgres
DATABASE_HOST=$DB_HOST
DATABASE_PORT=$DB_PORT
DATABASE_NAME=$DB_NAME
DATABASE_USER=$DB_USER
DATABASE_PASSWORD=$DB_PASSWORD
DATABASE_URL=postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable

# 兼容性配置
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME"

    echo -e "\n${BLUE}标准数据库配置:${NC}"
    echo "$STANDARD_CONFIG"
}

# 更新各服务配置
update_service_configs() {
    echo -e "\n${YELLOW}4. 更新各服务配置文件...${NC}"
    
    # 服务配置文件列表
    CONFIG_FILES=(
        "./backend/.env"
        "./services/write-service/.env" 
        "./services/courier-service/.env"
        "./services/gateway/.env"
        "./.env"
    )
    
    for config_file in "${CONFIG_FILES[@]}"; do
        if [ -f "$config_file" ]; then
            echo "  - 更新: $config_file"
            # 保留其他配置，只更新数据库相关
            if [ -f "$config_file" ]; then
                # 创建临时文件
                temp_file=$(mktemp)
                
                # 保留非数据库配置
                grep -v "^DATABASE" "$config_file" | grep -v "^DB_" > "$temp_file" 2>/dev/null || true
                
                # 添加标准数据库配置
                echo "" >> "$temp_file"
                echo "$STANDARD_CONFIG" >> "$temp_file"
                
                # 替换原文件
                mv "$temp_file" "$config_file"
            fi
        else
            echo "  - 创建: $config_file"
            mkdir -p "$(dirname "$config_file")"
            echo "$STANDARD_CONFIG" > "$config_file"
        fi
    done
    
    echo -e "${GREEN}✅ 所有服务配置已统一${NC}"
}

# 清理SQLite文件
cleanup_sqlite() {
    echo -e "\n${YELLOW}5. 清理历史SQLite文件...${NC}"
    
    # 只清理非备份的SQLite文件
    SQLITE_FILES=$(find . -name "*.db" -o -name "*.sqlite" -o -name "*.sqlite3" | grep -v archive | grep -v backup)
    
    if [ -n "$SQLITE_FILES" ]; then
        echo "发现的SQLite文件:"
        echo "$SQLITE_FILES"
        
        read -p "是否删除这些SQLite文件? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "$SQLITE_FILES" | xargs rm -f
            echo -e "${GREEN}✅ SQLite文件已清理${NC}"
        else
            echo "跳过清理SQLite文件"
        fi
    else
        echo "未发现需要清理的SQLite文件"
    fi
}

# 验证数据库连接
verify_connection() {
    echo -e "\n${YELLOW}6. 验证数据库连接...${NC}"
    
    # 检查数据库是否存在
    if psql -h localhost -U rocalight -d openpenpal -c "SELECT 1;" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ 数据库连接成功${NC}"
    else
        echo -e "${RED}❌ 数据库连接失败${NC}"
        echo "请检查PostgreSQL配置和用户权限"
        return 1
    fi
}

# 重启服务建议
restart_services() {
    echo -e "\n${YELLOW}7. 重启服务建议...${NC}"
    echo "配置更新完成，建议重启所有服务以使配置生效:"
    echo "  ./startup/stop-all.sh"
    echo "  ./startup/quick-start.sh"
}

# 主执行流程
main() {
    echo -e "${BLUE}开始数据库配置统一修复...${NC}"
    
    if ! check_postgresql; then
        exit 1
    fi
    
    backup_configs
    standardize_config
    update_service_configs
    cleanup_sqlite
    
    if verify_connection; then
        restart_services
        echo -e "\n${GREEN}🎉 数据库配置统一完成!${NC}"
    else
        echo -e "\n${RED}⚠️  配置已更新，但数据库连接验证失败${NC}"
        echo "请手动检查数据库配置和权限"
    fi
}

# 运行主程序
main "$@"