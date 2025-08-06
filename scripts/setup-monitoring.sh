#!/bin/bash

# OpenPenPal监控系统设置脚本
# 设置Prometheus + Grafana + AlertManager监控栈

echo "📊 设置OpenPenPal监控系统..."
echo "============================"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🔧 第1步: 创建监控配置目录${NC}"
echo "----------------------------------------"

# 创建监控相关目录
mkdir -p {config/grafana/dashboards,config/grafana/provisioning/{dashboards,datasources},config/alertmanager}
echo "✅ 创建监控目录结构"

echo ""

echo -e "${BLUE}📋 第2步: 创建Grafana配置${NC}"
echo "----------------------------------------"

# Grafana数据源配置
cat > config/grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
    basicAuth: false
    
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    isDefault: false
    editable: true
EOF

echo "✅ 创建Grafana数据源配置"

# Grafana仪表板配置
cat > config/grafana/provisioning/dashboards/openpenpal.yml << 'EOF'
apiVersion: 1

providers:
  - name: 'openpenpal'
    orgId: 1
    folder: 'OpenPenPal'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

echo "✅ 创建Grafana仪表板配置"

# 创建系统概览仪表板
cat > config/grafana/dashboards/system-overview.json << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "OpenPenPal系统概览",
    "tags": ["openpenpal", "overview"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "系统状态",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=~\"openpenpal.*\"}",
            "legendFormat": "{{job}}"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 0,
          "y": 0
        }
      },
      {
        "id": 2,
        "title": "HTTP请求率",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 12,
          "y": 0
        }
      },
      {
        "id": 3,
        "title": "响应时间",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 24,
          "x": 0,
          "y": 8
        }
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
EOF

echo "✅ 创建系统概览仪表板"

echo ""

echo -e "${BLUE}🚨 第3步: 创建告警规则${NC}"
echo "----------------------------------------"

# Prometheus告警规则
cat > config/alert_rules.yml << 'EOF'
groups:
  - name: openpenpal.alerts
    rules:
      # 服务可用性告警
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
          description: "Service {{ $labels.job }} has been down for more than 1 minute."

      # HTTP错误率告警
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} for {{ $labels.job }}"

      # 响应时间告警
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "95th percentile latency is {{ $value }}s for {{ $labels.job }}"

      # 数据库连接告警
      - alert: DatabaseConnectionHigh
        expr: pg_stat_activity_count > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High database connections"
          description: "Database has {{ $value }} active connections"

      # Redis内存使用告警
      - alert: RedisMemoryHigh
        expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.8
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "Redis memory usage high"
          description: "Redis memory usage is {{ $value | humanizePercentage }}"

      # 磁盘空间告警
      - alert: DiskSpaceHigh
        expr: (node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes > 0.85
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Disk space running low"
          description: "Disk usage is {{ $value | humanizePercentage }} on {{ $labels.device }}"

      # CPU使用率告警
      - alert: CPUUsageHigh
        expr: 100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage"
          description: "CPU usage is {{ $value }}% on {{ $labels.instance }}"

      # 内存使用告警
      - alert: MemoryUsageHigh
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value | humanizePercentage }} on {{ $labels.instance }}"

      # 业务指标告警
      - alert: UserRegistrationRateLow
        expr: rate(user_registrations_total[1h]) < 1
        for: 30m
        labels:
          severity: info
        annotations:
          summary: "Low user registration rate"
          description: "User registration rate is {{ $value }} per hour"

      - alert: LetterDeliveryRateLow
        expr: rate(letters_delivered_total[1h]) < 5
        for: 30m
        labels:
          severity: info
        annotations:
          summary: "Low letter delivery rate"
          description: "Letter delivery rate is {{ $value }} per hour"
EOF

echo "✅ 创建Prometheus告警规则"

# AlertManager配置
cat > config/alertmanager/alertmanager.yml << 'EOF'
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@openpenpal.com'
  smtp_auth_username: 'alerts@openpenpal.com'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    email_configs:
      - to: 'admin@openpenpal.com'
        subject: 'OpenPenPal Alert: {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
    webhook_configs:
      - url: 'http://localhost:5001/'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF

echo "✅ 创建AlertManager配置"

echo ""

echo -e "${BLUE}🐳 第4步: 创建监控Docker配置${NC}"
echo "----------------------------------------"

# 监控栈Docker Compose配置
cat > docker-compose.monitoring.yml << 'EOF'
version: '3.8'

services:
  # Prometheus时序数据库
  prometheus:
    image: prom/prometheus:latest
    container_name: openpenpal-prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    restart: unless-stopped
    ports:
      - '9090:9090'
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./config/alert_rules.yml:/etc/prometheus/alert_rules.yml
      - prometheus_data:/prometheus
    networks:
      - openpenpal-monitoring

  # Grafana可视化
  grafana:
    image: grafana/grafana:latest
    container_name: openpenpal-grafana
    restart: unless-stopped
    ports:
      - '3001:3000'
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./config/grafana/provisioning:/etc/grafana/provisioning
      - ./config/grafana/dashboards:/etc/grafana/provisioning/dashboards
    networks:
      - openpenpal-monitoring

  # AlertManager告警管理
  alertmanager:
    image: prom/alertmanager:latest
    container_name: openpenpal-alertmanager
    restart: unless-stopped
    ports:
      - '9093:9093'
    volumes:
      - ./config/alertmanager:/etc/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9093'
    networks:
      - openpenpal-monitoring

  # Node Exporter系统监控
  node-exporter:
    image: prom/node-exporter:latest
    container_name: openpenpal-node-exporter
    restart: unless-stopped
    ports:
      - '9100:9100'
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    networks:
      - openpenpal-monitoring

  # Loki日志聚合
  loki:
    image: grafana/loki:latest
    container_name: openpenpal-loki
    restart: unless-stopped
    ports:
      - '3100:3100'
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - openpenpal-monitoring

  # Promtail日志收集
  promtail:
    image: grafana/promtail:latest
    container_name: openpenpal-promtail
    restart: unless-stopped
    volumes:
      - ./logs:/var/log/openpenpal
      - ./config/log-monitoring.yml:/etc/promtail/config.yml
    command: -config.file=/etc/promtail/config.yml
    networks:
      - openpenpal-monitoring

volumes:
  prometheus_data:
  grafana_data:

networks:
  openpenpal-monitoring:
    driver: bridge
EOF

echo "✅ 创建监控Docker配置"

echo ""

echo -e "${YELLOW}🚀 第5步: 创建监控启动脚本${NC}"
echo "----------------------------------------"

# 监控启动脚本
cat > scripts/start-monitoring.sh << 'EOF'
#!/bin/bash

# OpenPenPal监控系统启动脚本

echo "📊 启动OpenPenPal监控系统..."

# 检查Docker
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请启动Docker"
    exit 1
fi

# 创建网络
docker network create openpenpal-monitoring 2>/dev/null || true

# 启动监控服务
echo "🏗️  启动监控服务..."
docker-compose -f docker-compose.monitoring.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 健康检查
echo "🔍 执行健康检查..."
services=("prometheus" "grafana" "alertmanager" "node-exporter")
for service in "${services[@]}"; do
    if docker-compose -f docker-compose.monitoring.yml ps -q $service > /dev/null 2>&1; then
        echo "✅ $service 运行正常"
    else
        echo "❌ $service 启动失败"
    fi
done

echo ""
echo "🎉 监控系统启动完成！"
echo "📋 访问地址:"
echo "   • Prometheus: http://localhost:9090"
echo "   • Grafana: http://localhost:3001 (admin/admin123)"
echo "   • AlertManager: http://localhost:9093"
echo "   • Node Exporter: http://localhost:9100"
EOF

chmod +x scripts/start-monitoring.sh
echo "✅ 创建监控启动脚本"

echo ""

echo -e "${GREEN}🎊 监控系统设置完成${NC}"
echo "============================"

echo "📋 已创建的文件:"
echo "   • config/prometheus.yml - Prometheus配置"
echo "   • config/alert_rules.yml - 告警规则"
echo "   • config/alertmanager/ - AlertManager配置"
echo "   • config/grafana/ - Grafana配置和仪表板"
echo "   • docker-compose.monitoring.yml - 监控容器编排"
echo "   • scripts/start-monitoring.sh - 监控启动脚本"

echo ""
echo -e "${YELLOW}📋 使用说明:${NC}"
echo "1. 启动监控系统: ./scripts/start-monitoring.sh"
echo "2. 访问Grafana: http://localhost:3001 (admin/admin123)"
echo "3. 查看告警: http://localhost:9093"
echo "4. 停止监控: docker-compose -f docker-compose.monitoring.yml down"

echo ""
echo -e "${GREEN}✨ 监控基础设施搭建完成！${NC}"