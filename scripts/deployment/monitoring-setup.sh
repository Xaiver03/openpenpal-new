#!/bin/bash
# 监控系统设置和管理脚本
# 用于快速部署和管理 OpenPenPal 监控堆栈
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

# 检查 Docker 和 Docker Compose
check_dependencies() {
    echo -e "${BLUE}🔍 检查依赖...${NC}"
    
    if ! command -v docker >/dev/null 2>&1; then
        echo -e "${RED}❌ Docker 未安装${NC}"
        exit 1
    fi
    
    if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
        echo -e "${RED}❌ Docker Compose 未安装${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 依赖检查通过${NC}"
}

# 创建必要的目录
setup_directories() {
    echo -e "${BLUE}📁 创建监控目录结构...${NC}"
    
    # 创建 Grafana 仪表板目录
    mkdir -p "$SCRIPT_DIR/monitoring/grafana/dashboards"/{system,microservices,database,business,observability}
    
    # 创建数据目录
    mkdir -p "$SCRIPT_DIR/monitoring/data"/{prometheus,grafana,loki,jaeger}
    
    # 设置权限
    sudo chown -R 472:472 "$SCRIPT_DIR/monitoring/data/grafana" 2>/dev/null || true
    sudo chown -R 65534:65534 "$SCRIPT_DIR/monitoring/data/prometheus" 2>/dev/null || true
    
    echo -e "${GREEN}✅ 目录结构创建完成${NC}"
}

# 生成环境变量文件
generate_env_file() {
    if [ ! -f "$SCRIPT_DIR/.env.monitoring" ]; then
        echo -e "${BLUE}⚙️ 生成监控环境变量...${NC}"
        
        cat > "$SCRIPT_DIR/.env.monitoring" << EOF
# OpenPenPal 监控环境变量
GRAFANA_ADMIN_PASSWORD=admin123
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=openpenpal
POSTGRES_READONLY_USER=grafana_readonly
POSTGRES_READONLY_PASSWORD=readonly_password

# 告警通知配置
SMTP_PASSWORD=your_smtp_password
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK

# Jaeger 配置
JAEGER_AGENT_HOST=jaeger
JAEGER_AGENT_PORT=6831

# 外部服务配置
EXTERNAL_API_URL=https://api.openpenpal.com
EOF
        
        echo -e "${YELLOW}⚠️  请编辑 .env.monitoring 文件配置通知设置${NC}"
    else
        echo -e "${GREEN}✅ 环境变量文件已存在${NC}"
    fi
}

# 下载 Grafana 仪表板模板
download_dashboards() {
    echo -e "${BLUE}📊 下载 Grafana 仪表板...${NC}"
    
    DASHBOARD_DIR="$SCRIPT_DIR/monitoring/grafana/dashboards"
    
    # 系统监控仪表板
    curl -s https://grafana.com/api/dashboards/1860/revisions/27/download > "$DASHBOARD_DIR/system/node-exporter.json" || true
    curl -s https://grafana.com/api/dashboards/893/revisions/4/download > "$DASHBOARD_DIR/system/docker-monitoring.json" || true
    
    # 数据库监控仪表板
    curl -s https://grafana.com/api/dashboards/9628/revisions/7/download > "$DASHBOARD_DIR/database/postgresql.json" || true
    curl -s https://grafana.com/api/dashboards/763/revisions/5/download > "$DASHBOARD_DIR/database/redis.json" || true
    
    echo -e "${GREEN}✅ 仪表板下载完成${NC}"
}

# 创建简单的系统概览仪表板
create_system_dashboard() {
    cat > "$SCRIPT_DIR/monitoring/grafana/dashboards/system/openpenpal-overview.json" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "OpenPenPal System Overview",
    "tags": ["openpenpal", "system"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Service Status",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=~\"frontend|backend|.*-service\"}",
            "legendFormat": "{{job}}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {
                  "color": "red",
                  "value": 0
                },
                {
                  "color": "green",
                  "value": 1
                }
              ]
            }
          }
        },
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 0,
          "y": 0
        }
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "5s"
  }
}
EOF
}

# 配置 Blackbox Exporter
setup_blackbox_config() {
    mkdir -p "$SCRIPT_DIR/monitoring/blackbox"
    
    cat > "$SCRIPT_DIR/monitoring/blackbox/blackbox.yml" << 'EOF'
modules:
  http_2xx:
    prober: http
    timeout: 5s
    http:
      valid_http_versions: ["HTTP/1.1", "HTTP/2.0"]
      valid_status_codes: []
      method: GET
      headers:
        Host: vhost.example.com
        Accept-Language: en-US
      no_follow_redirects: false
      fail_if_ssl: false
      fail_if_not_ssl: false
      
  https_2xx:
    prober: http
    timeout: 5s
    http:
      valid_status_codes: []
      method: GET
      no_follow_redirects: false
      fail_if_not_ssl: true
      
  tcp_connect:
    prober: tcp
    timeout: 5s
    
  websocket_connect:
    prober: http
    timeout: 5s
    http:
      valid_status_codes: []
      method: GET
      headers:
        Host: vhost.example.com
        Origin: example.com
        Sec-WebSocket-Protocol: chat, superchat
        Sec-WebSocket-Version: 13
        Connection: Upgrade
        Upgrade: websocket
        Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==
      fail_if_not_ssl: false
EOF
}

# 配置 Loki
setup_loki_config() {
    mkdir -p "$SCRIPT_DIR/monitoring/loki"
    
    cat > "$SCRIPT_DIR/monitoring/loki/loki-config.yml" << 'EOF'
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    instance_addr: 127.0.0.1
    kvstore:
      store: inmemory

query_range:
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 100

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://localhost:9093

limits_config:
  reject_old_samples: true
  reject_old_samples_max_age: 168h

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: false
  retention_period: 0s

compactor:
  working_directory: /loki/boltdb-shipper-compactor
  shared_store: filesystem

ingester:
  max_chunk_age: 1h
  chunk_idle_period: 3m
  chunk_block_size: 262144
  chunk_retain_period: 1m
  max_transfer_retries: 0
  wal:
    enabled: true
    dir: /loki/wal
  lifecycler:
    addr: 127.0.0.1
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
    final_sleep: 0s
EOF
}

# 配置 Promtail
setup_promtail_config() {
    mkdir -p "$SCRIPT_DIR/monitoring/promtail"
    
    cat > "$SCRIPT_DIR/monitoring/promtail/promtail-config.yml" << 'EOF'
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # 系统日志
  - job_name: system
    static_configs:
      - targets:
          - localhost
        labels:
          job: varlogs
          __path__: /var/log/*log

  # Docker 容器日志
  - job_name: containers
    static_configs:
      - targets:
          - localhost
        labels:
          job: containerlogs
          __path__: /var/lib/docker/containers/*/*log
    
    pipeline_stages:
      - json:
          expressions:
            output: log
            stream: stream
            attrs:
      - json:
          expressions:
            tag:
          source: attrs
      - regex:
          expression: (?P<container_name>(?:[^|]|\|\|)+)
          source: tag
      - timestamp:
          format: RFC3339Nano
          source: time
      - labels:
          stream:
          container_name:
      - output:
          source: output

  # OpenPenPal 应用日志
  - job_name: openpenpal-logs
    static_configs:
      - targets:
          - localhost
        labels:
          job: openpenpal
          __path__: /var/log/openpenpal/*.log
EOF
}

# 启动监控堆栈
start_monitoring() {
    echo -e "${BLUE}🚀 启动监控堆栈...${NC}"
    
    cd "$SCRIPT_DIR"
    
    # 加载环境变量
    if [ -f ".env.monitoring" ]; then
        export $(cat .env.monitoring | grep -v '^#' | xargs)
    fi
    
    # 启动监控服务
    docker-compose -f monitoring-stack.yml --env-file .env.monitoring up -d
    
    echo -e "${GREEN}✅ 监控堆栈启动完成${NC}"
    echo ""
    echo -e "${BLUE}📊 访问地址:${NC}"
    echo "  Grafana:      http://localhost:3001 (admin/admin123)"
    echo "  Prometheus:   http://localhost:9090"
    echo "  AlertManager: http://localhost:9093"
    echo "  Jaeger:       http://localhost:16686"
    echo "  Loki:         http://localhost:3100"
}

# 停止监控堆栈
stop_monitoring() {
    echo -e "${BLUE}🛑 停止监控堆栈...${NC}"
    
    cd "$SCRIPT_DIR"
    docker-compose -f monitoring-stack.yml down
    
    echo -e "${GREEN}✅ 监控堆栈已停止${NC}"
}

# 查看监控状态
status_monitoring() {
    echo -e "${BLUE}📊 监控服务状态:${NC}"
    
    cd "$SCRIPT_DIR"
    docker-compose -f monitoring-stack.yml ps
}

# 查看日志
logs_monitoring() {
    local service="${1:-}"
    
    cd "$SCRIPT_DIR"
    
    if [ -n "$service" ]; then
        docker-compose -f monitoring-stack.yml logs -f "$service"
    else
        docker-compose -f monitoring-stack.yml logs -f
    fi
}

# 重启监控服务
restart_monitoring() {
    echo -e "${BLUE}🔄 重启监控堆栈...${NC}"
    
    stop_monitoring
    sleep 2
    start_monitoring
}

# 清理监控数据
clean_monitoring() {
    echo -e "${YELLOW}⚠️  确认清理所有监控数据? (y/N)${NC}"
    read -r response
    
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${BLUE}🧹 清理监控数据...${NC}"
        
        cd "$SCRIPT_DIR"
        docker-compose -f monitoring-stack.yml down -v
        sudo rm -rf monitoring/data/* 2>/dev/null || true
        
        echo -e "${GREEN}✅ 监控数据清理完成${NC}"
    else
        echo -e "${BLUE}ℹ️  取消清理操作${NC}"
    fi
}

# 主函数
main() {
    case "${1:-}" in
        "start"|"up")
            check_dependencies
            setup_directories
            generate_env_file
            setup_blackbox_config
            setup_loki_config
            setup_promtail_config
            download_dashboards
            create_system_dashboard
            start_monitoring
            ;;
        "stop"|"down")
            stop_monitoring
            ;;
        "restart")
            restart_monitoring
            ;;
        "status")
            status_monitoring
            ;;
        "logs")
            logs_monitoring "${2:-}"
            ;;
        "clean")
            clean_monitoring
            ;;
        "setup")
            check_dependencies
            setup_directories
            generate_env_file
            setup_blackbox_config
            setup_loki_config
            setup_promtail_config
            download_dashboards
            create_system_dashboard
            echo -e "${GREEN}✅ 监控环境设置完成，运行 '$0 start' 启动${NC}"
            ;;
        *)
            echo -e "${BLUE}OpenPenPal 监控管理脚本${NC}"
            echo ""
            echo "用法: $0 {start|stop|restart|status|logs|clean|setup}"
            echo ""
            echo "命令:"
            echo "  start   - 启动监控堆栈"
            echo "  stop    - 停止监控堆栈"
            echo "  restart - 重启监控堆栈"
            echo "  status  - 查看服务状态"
            echo "  logs    - 查看日志 (可指定服务名)"
            echo "  clean   - 清理所有监控数据"
            echo "  setup   - 仅设置环境，不启动服务"
            echo ""
            echo "示例:"
            echo "  $0 start                    # 启动监控"
            echo "  $0 logs prometheus          # 查看 Prometheus 日志"
            echo "  $0 logs                     # 查看所有服务日志"
            ;;
    esac
}

main "$@"