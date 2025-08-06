#!/bin/bash

# 测试不同启动模式的数据库配置

echo "OpenPenPal 启动模式数据库测试"
echo "=============================="
echo ""

# 测试函数
test_mode() {
    local mode=$1
    echo -n "测试 $mode 模式... "
    
    # 清除之前的环境变量
    unset DATABASE_TYPE
    
    # 运行启动脚本获取环境变量
    source ./startup/environment-vars.sh
    
    # 模拟启动脚本的模式设置
    case $mode in
        production)
            export DATABASE_TYPE="postgres"
            ;;
    esac
    
    # 再次加载环境变量以应用数据库配置
    source ./startup/environment-vars.sh
    
    echo "DATABASE_TYPE=$DATABASE_TYPE"
}

# 测试各种模式
echo "默认数据库配置："
test_mode "development"
test_mode "simple"
test_mode "demo"
test_mode "production"
test_mode "complete"

echo ""
echo "总结："
echo "- development/simple/demo/complete 模式: 使用 SQLite"
echo "- production 模式: 使用 PostgreSQL"
echo ""
echo "可以通过环境变量覆盖任何模式的数据库配置："
echo "  export DATABASE_TYPE=postgres"
echo "  ./startup/quick-start.sh development"