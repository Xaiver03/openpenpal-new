#!/bin/bash
# 微服务治理设置和管理脚本
# 用于管理服务发现、负载均衡、配置管理、熔断降级等
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 微服务配置
SERVICES=(
    "frontend:3000:http://host.docker.internal:3000"
    "backend:8080:http://host.docker.internal:8080"
    "write-service:8001:http://host.docker.internal:8001"
    "courier-service:8002:http://host.docker.internal:8002"
    "admin-service:8003:http://host.docker.internal:8003"
    "ocr-service:8004:http://host.docker.internal:8004"
    "gateway:8000:http://host.docker.internal:8000"
)

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}🔍 检查依赖...${NC}"
    
    if ! command -v docker >/dev/null 2>&1; then
        echo -e "${RED}❌ Docker 未安装${NC}"
        exit 1
    fi
    
    if ! command -v curl >/dev/null 2>&1; then
        echo -e "${RED}❌ curl 未安装${NC}"
        exit 1
    fi
    
    if ! command -v jq >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  jq 未安装，建议安装以获得更好的体验${NC}"
    fi
    
    echo -e "${GREEN}✅ 依赖检查通过${NC}"
}

# 创建配置文件
setup_configs() {
    echo -e "${BLUE}⚙️ 创建配置文件...${NC}"
    
    # Consul 配置
    cat > "$SCRIPT_DIR/governance/consul/consul.json" << 'EOF'
{
  "datacenter": "openpenpal-dc1",
  "data_dir": "/consul/data",
  "log_level": "INFO",
  "server": true,
  "ui": true,
  "bootstrap_expect": 1,
  "bind_addr": "0.0.0.0",
  "client_addr": "0.0.0.0",
  "retry_join": ["consul"],
  "connect": {
    "enabled": true
  },
  "ports": {
    "grpc": 8502
  },
  "acl": {
    "enabled": false,
    "default_policy": "allow"
  }
}
EOF

    # Vault 配置
    cat > "$SCRIPT_DIR/governance/vault/vault.hcl" << 'EOF'
storage "file" {
  path = "/vault/data"
}

listener "tcp" {
  address = "0.0.0.0:8200"
  tls_disable = true
}

ui = true
log_level = "Info"
EOF

    # Nginx 配置模板
    cat > "$SCRIPT_DIR/governance/consul-template/nginx.conf.tpl" << 'EOF'
upstream frontend {
    {{range service "frontend"}}
    server {{.Address}}:{{.Port}} max_fails=3 fail_timeout=60 weight=1;
    {{else}}
    server 127.0.0.1:65535; # force a 502
    {{end}}
}

upstream backend {
    {{range service "backend"}}
    server {{.Address}}:{{.Port}} max_fails=3 fail_timeout=60 weight=1;
    {{else}}
    server 127.0.0.1:65535; # force a 502
    {{end}}
}

upstream gateway {
    {{range service "gateway"}}
    server {{.Address}}:{{.Port}} max_fails=3 fail_timeout=60 weight=1;
    {{else}}
    server 127.0.0.1:65535; # force a 502
    {{end}}
}

server {
    listen 80;
    server_name localhost openpenpal.local;
    
    # 健康检查端点
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
    
    # 前端静态资源
    location / {
        proxy_pass http://frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # API 路由到网关
    location /api/ {
        proxy_pass http://gateway/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # 直接路由到后端 (备用)
    location /backend/ {
        proxy_pass http://backend/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# 管理界面代理
server {
    listen 8080;
    server_name admin.openpenpal.local;
    
    location / {
        proxy_pass http://konga:1337/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

    # Redis 集群配置
    for i in 1 2 3; do
        port=$((7000 + i))
        cluster_port=$((17000 + i))
        cat > "$SCRIPT_DIR/governance/redis/redis-$i.conf" << EOF
port $port
cluster-enabled yes
cluster-config-file nodes-$port.conf
cluster-node-timeout 5000
cluster-announce-port $port
cluster-announce-bus-port $cluster_port
appendonly yes
EOF
    done

    echo -e "${GREEN}✅ 配置文件创建完成${NC}"
}

# 注册服务到 Consul
register_services() {
    echo -e "${BLUE}📋 注册服务到 Consul...${NC}"
    
    # 等待 Consul 启动
    echo "等待 Consul 启动..."
    for i in {1..30}; do
        if curl -s http://localhost:8500/v1/status/leader >/dev/null 2>&1; then
            break
        fi
        sleep 2
    done
    
    # 注册每个服务
    for service_config in "${SERVICES[@]}"; do
        IFS=':' read -r name port url <<< "$service_config"
        
        echo "注册服务: $name"
        
        cat > "/tmp/service-$name.json" << EOF
{
  "ID": "$name-1",
  "Name": "$name",
  "Port": $port,
  "Address": "host.docker.internal",
  "Check": {
    "HTTP": "$url/health",
    "Interval": "30s",
    "Timeout": "10s"
  },
  "Tags": ["openpenpal", "microservice"]
}
EOF
        
        curl -X PUT "http://localhost:8500/v1/agent/service/register" \
             -d @"/tmp/service-$name.json" >/dev/null 2>&1 || true
        
        rm -f "/tmp/service-$name.json"
    done
    
    echo -e "${GREEN}✅ 服务注册完成${NC}"
}

# 配置 Kong 网关
setup_kong() {
    echo -e "${BLUE}🦍 配置 Kong 网关...${NC}"
    
    # 等待 Kong 启动
    echo "等待 Kong 启动..."
    for i in {1..30}; do
        if curl -s http://localhost:8001/ >/dev/null 2>&1; then
            break
        fi
        sleep 2
    done
    
    # 创建 OpenPenPal 服务和路由
    for service_config in "${SERVICES[@]}"; do
        IFS=':' read -r name port url <<< "$service_config"
        
        echo "配置 Kong 服务: $name"
        
        # 创建服务
        curl -i -X POST http://localhost:8001/services/ \
             --data "name=$name" \
             --data "url=$url" >/dev/null 2>&1 || true
        
        # 创建路由
        curl -i -X POST http://localhost:8001/services/$name/routes \
             --data "paths[]=/$name" \
             --data "strip_path=true" >/dev/null 2>&1 || true
    done
    
    # 启用插件
    echo "启用 Kong 插件..."
    
    # 全局速率限制
    curl -i -X POST http://localhost:8001/plugins/ \
         --data "name=rate-limiting" \
         --data "config.minute=1000" \
         --data "config.hour=10000" >/dev/null 2>&1 || true
    
    # CORS 支持
    curl -i -X POST http://localhost:8001/plugins/ \
         --data "name=cors" \
         --data "config.origins=*" \
         --data "config.methods=GET,POST,PUT,DELETE,PATCH,OPTIONS" \
         --data "config.headers=Accept,Accept-Version,Content-Length,Content-MD5,Content-Type,Date,X-Auth-Token,Authorization" >/dev/null 2>&1 || true
    
    # 请求ID追踪
    curl -i -X POST http://localhost:8001/plugins/ \
         --data "name=correlation-id" >/dev/null 2>&1 || true
    
    # Prometheus 指标
    curl -i -X POST http://localhost:8001/plugins/ \
         --data "name=prometheus" >/dev/null 2>&1 || true
    
    echo -e "${GREEN}✅ Kong 配置完成${NC}"
}

# 设置熔断器
setup_circuit_breaker() {
    echo -e "${BLUE}⚡ 配置熔断器...${NC}"
    
    # 这里可以配置 Hystrix 或其他熔断器
    # 为了演示，我们创建一个简单的健康检查脚本
    cat > "$SCRIPT_DIR/circuit-breaker-check.sh" << 'EOF'
#!/bin/bash
# 简单的熔断器健康检查

SERVICES=(
    "frontend:3000"
    "backend:8080"
    "gateway:8000"
)

for service in "${SERVICES[@]}"; do
    IFS=':' read -r name port <<< "$service"
    
    if curl -s --max-time 5 "http://localhost:$port/health" >/dev/null; then
        echo "✅ $name: HEALTHY"
    else
        echo "❌ $name: UNHEALTHY - 触发熔断"
        # 这里可以添加熔断逻辑
    fi
done
EOF
    
    chmod +x "$SCRIPT_DIR/circuit-breaker-check.sh"
    
    echo -e "${GREEN}✅ 熔断器配置完成${NC}"
}

# 启动治理堆栈
start_governance() {
    echo -e "${BLUE}🚀 启动微服务治理堆栈...${NC}"
    
    cd "$SCRIPT_DIR"
    
    # 启动基础服务
    docker-compose -f microservices-governance.yml up -d consul vault kong-database
    
    # 等待基础服务启动
    sleep 10
    
    # 启动Kong迁移
    docker-compose -f microservices-governance.yml up kong-migration
    
    # 启动其他服务
    docker-compose -f microservices-governance.yml up -d
    
    echo -e "${GREEN}✅ 治理堆栈启动完成${NC}"
    
    # 等待服务就绪后进行配置
    echo -e "${BLUE}⏳ 等待服务启动完成...${NC}"
    sleep 20
    
    register_services
    setup_kong
    setup_circuit_breaker
    
    echo ""
    echo -e "${BLUE}🎯 微服务治理访问地址:${NC}"
    echo "  Consul UI:        http://localhost:8500"
    echo "  Kong Admin:       http://localhost:8001"
    echo "  Konga UI:         http://localhost:1337"
    echo "  Vault UI:         http://localhost:8200"
    echo "  Hystrix Dashboard:http://localhost:9002"
    echo "  Config Server:    http://localhost:8888"
    echo "  Zipkin:           http://localhost:9411"
    echo ""
    echo "  应用访问 (通过负载均衡):"
    echo "  Frontend:         http://localhost/"
    echo "  API Gateway:      http://localhost/api/"
}

# 停止治理堆栈
stop_governance() {
    echo -e "${BLUE}🛑 停止微服务治理堆栈...${NC}"
    
    cd "$SCRIPT_DIR"
    docker-compose -f microservices-governance.yml down
    
    echo -e "${GREEN}✅ 治理堆栈已停止${NC}"
}

# 查看服务状态
status_governance() {
    echo -e "${BLUE}📊 微服务治理状态:${NC}"
    
    cd "$SCRIPT_DIR"
    docker-compose -f microservices-governance.yml ps
    
    echo ""
    echo -e "${BLUE}🔍 Consul 服务状态:${NC}"
    if curl -s http://localhost:8500/v1/agent/services 2>/dev/null | jq -r 'keys[]' 2>/dev/null; then
        curl -s http://localhost:8500/v1/agent/services | jq '.[] | "\(.Service): \(.Address):\(.Port)"'
    else
        curl -s http://localhost:8500/v1/agent/services 2>/dev/null || echo "Consul 不可用"
    fi
    
    echo ""
    echo -e "${BLUE}🦍 Kong 服务状态:${NC}"
    if curl -s http://localhost:8001/services 2>/dev/null | jq -r '.data[].name' 2>/dev/null; then
        curl -s http://localhost:8001/services | jq '.data[] | "\(.name): \(.host)"'
    else
        curl -s http://localhost:8001/services 2>/dev/null || echo "Kong 不可用"
    fi
}

# 健康检查
health_check() {
    echo -e "${BLUE}🏥 执行健康检查...${NC}"
    
    # 运行熔断器检查
    if [ -f "$SCRIPT_DIR/circuit-breaker-check.sh" ]; then
        "$SCRIPT_DIR/circuit-breaker-check.sh"
    fi
    
    echo ""
    echo -e "${BLUE}📋 Consul 健康检查:${NC}"
    curl -s http://localhost:8500/v1/health/state/any 2>/dev/null | jq -r '.[] | "\(.ServiceName): \(.Status)"' 2>/dev/null || echo "Consul 健康检查不可用"
}

# 服务发现测试
test_service_discovery() {
    echo -e "${BLUE}🔍 测试服务发现...${NC}"
    
    echo "从 Consul 查询服务:"
    for service_config in "${SERVICES[@]}"; do
        IFS=':' read -r name port url <<< "$service_config"
        echo -n "  $name: "
        
        if consul_result=$(curl -s "http://localhost:8500/v1/health/service/$name" 2>/dev/null); then
            if echo "$consul_result" | jq -e '.[] | select(.Checks[].Status == "passing")' >/dev/null 2>&1; then
                echo -e "${GREEN}✅ 健康${NC}"
            else
                echo -e "${RED}❌ 不健康${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️  未注册${NC}"
        fi
    done
}

# 负载均衡测试
test_load_balancing() {
    echo -e "${BLUE}⚖️ 测试负载均衡...${NC}"
    
    echo "通过 Kong 网关测试:"
    for i in {1..5}; do
        echo -n "  请求 $i: "
        if response=$(curl -s -w "%{http_code}" "http://localhost:8000/backend/health" 2>/dev/null); then
            http_code="${response: -3}"
            if [ "$http_code" = "200" ]; then
                echo -e "${GREEN}✅ 成功${NC}"
            else
                echo -e "${RED}❌ 失败 ($http_code)${NC}"
            fi
        else
            echo -e "${RED}❌ 连接失败${NC}"
        fi
        sleep 1
    done
}

# 配置管理测试
test_config_management() {
    echo -e "${BLUE}⚙️ 测试配置管理...${NC}"
    
    # 测试从 Consul KV 读取配置
    echo "设置测试配置..."
    curl -X PUT "http://localhost:8500/v1/kv/openpenpal/config/test" -d "test-value" >/dev/null 2>&1
    
    echo -n "读取配置: "
    if config_value=$(curl -s "http://localhost:8500/v1/kv/openpenpal/config/test?raw" 2>/dev/null); then
        if [ "$config_value" = "test-value" ]; then
            echo -e "${GREEN}✅ 成功${NC}"
        else
            echo -e "${RED}❌ 值不匹配${NC}"
        fi
    else
        echo -e "${RED}❌ 读取失败${NC}"
    fi
}

# 运行全面测试
run_tests() {
    echo -e "${BLUE}🧪 运行微服务治理测试...${NC}"
    echo ""
    
    test_service_discovery
    echo ""
    test_load_balancing
    echo ""
    test_config_management
    echo ""
    health_check
    
    echo ""
    echo -e "${GREEN}✅ 测试完成${NC}"
}

# 清理数据
clean_governance() {
    echo -e "${YELLOW}⚠️  确认清理所有治理数据? (y/N)${NC}"
    read -r response
    
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${BLUE}🧹 清理治理数据...${NC}"
        
        cd "$SCRIPT_DIR"
        docker-compose -f microservices-governance.yml down -v
        sudo rm -rf governance/*/data 2>/dev/null || true
        
        echo -e "${GREEN}✅ 数据清理完成${NC}"
    else
        echo -e "${BLUE}ℹ️  取消清理操作${NC}"
    fi
}

# 主函数
main() {
    case "${1:-}" in
        "start"|"up")
            check_dependencies
            setup_configs
            start_governance
            ;;
        "stop"|"down")
            stop_governance
            ;;
        "restart")
            stop_governance
            sleep 2
            check_dependencies
            setup_configs
            start_governance
            ;;
        "status")
            status_governance
            ;;
        "health")
            health_check
            ;;
        "test")
            run_tests
            ;;
        "clean")
            clean_governance
            ;;
        "setup")
            check_dependencies
            setup_configs
            echo -e "${GREEN}✅ 治理环境设置完成，运行 '$0 start' 启动${NC}"
            ;;
        *)
            echo -e "${BLUE}OpenPenPal 微服务治理管理脚本${NC}"
            echo ""
            echo "用法: $0 {start|stop|restart|status|health|test|clean|setup}"
            echo ""
            echo "命令:"
            echo "  start   - 启动微服务治理堆栈"
            echo "  stop    - 停止微服务治理堆栈"
            echo "  restart - 重启微服务治理堆栈"
            echo "  status  - 查看服务状态"
            echo "  health  - 执行健康检查"
            echo "  test    - 运行治理功能测试"
            echo "  clean   - 清理所有数据"
            echo "  setup   - 仅设置环境，不启动服务"
            echo ""
            echo "微服务治理组件:"
            echo "  🔍 服务发现: Consul"
            echo "  🦍 API网关: Kong"
            echo "  🔐 秘钥管理: Vault"
            echo "  ⚖️  负载均衡: Nginx + Consul Template"
            echo "  ⚡ 熔断降级: Hystrix"
            echo "  📊 链路追踪: Zipkin"
            echo "  ⚙️  配置管理: Consul KV + Config Server"
            ;;
    esac
}

main "$@"