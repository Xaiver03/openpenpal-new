# OpenPenPal 自动化部署指南

## 概述

本指南详细介绍如何通过 GitHub Actions 实现 OpenPenPal 项目到腾讯云服务器的 CI/CD 自动化部署。

## 目录

1. [架构总览](#架构总览)
2. [准备工作](#准备工作)
3. [CI/CD 流程设计](#cicd-流程设计)
4. [GitHub Actions 配置](#github-actions-配置)
5. [腾讯云服务器配置](#腾讯云服务器配置)
6. [部署脚本配置](#部署脚本配置)
7. [监控与告警](#监控与告警)
8. [故障恢复](#故障恢复)
9. [最佳实践](#最佳实践)

## 架构总览

### 微服务架构

```
┌─────────────────────────────────────────────────────────┐
│                    Nginx (80/443)                       │
├─────────────────────────────────────────────────────────┤
│                  API Gateway (8000)                      │
├──────────┬──────────┬──────────┬───────────┬───────────┤
│ Frontend │  Admin   │ Backend  │  Write    │  Courier  │
│  (3000)  │  (3001)  │  (8080)  │  (8001)   │  (8002)   │
├──────────┴──────────┴──────────┴───────────┴───────────┤
│        Admin Service    │      OCR Service              │
│         (8003)          │        (8004)                 │
├─────────────────────────┴───────────────────────────────┤
│     PostgreSQL (5432)   │      Redis (6379)            │
├─────────────────────────┴───────────────────────────────┤
│   Prometheus (9090)     │     Grafana (3002)           │
└─────────────────────────────────────────────────────────┘
```

### 部署流程

```
GitHub Push → GitHub Actions → Build & Test → Docker Build → 
Push to Registry → Deploy to Tencent Cloud → Health Check → 
Notification
```

## 准备工作

### 1. 腾讯云服务器要求

- **系统**: Ubuntu 20.04 LTS 或更高版本
- **配置**: 最低 8GB RAM, 4 vCPU, 100GB SSD
- **网络**: 开放端口 80, 443, 22
- **域名**: 已备案的域名（如 openpenpal.com）

### 2. 必需的软件

在腾讯云服务器上安装：

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装其他工具
sudo apt install -y nginx certbot python3-certbot-nginx git
```

### 3. 配置 SSL 证书

```bash
# 使用 Let's Encrypt 获取免费 SSL 证书
sudo certbot --nginx -d openpenpal.com -d www.openpenpal.com
```

### 4. GitHub Secrets 配置

在 GitHub 仓库设置中添加以下 Secrets：

```yaml
# 腾讯云服务器连接
TENCENT_HOST: 你的服务器IP
TENCENT_USER: ubuntu
TENCENT_SSH_KEY: 服务器SSH私钥

# Docker Registry（使用腾讯云容器镜像服务）
DOCKER_REGISTRY: ccr.ccs.tencentyun.com
DOCKER_NAMESPACE: openpenpal
DOCKER_USERNAME: 腾讯云镜像服务用户名
DOCKER_PASSWORD: 腾讯云镜像服务密码

# 应用配置
JWT_SECRET: 生产环境JWT密钥
POSTGRES_PASSWORD: 数据库密码
REDIS_PASSWORD: Redis密码
GRAFANA_PASSWORD: Grafana管理员密码

# 通知配置（可选）
SLACK_WEBHOOK: Slack通知URL
EMAIL_NOTIFICATION: 通知邮箱
```

## CI/CD 流程设计

### 部署策略

采用**蓝绿部署**策略，确保零停机时间：

1. 构建新版本镜像
2. 部署到"绿"环境
3. 健康检查通过后切换流量
4. 保留"蓝"环境作为回滚备份

### 分支策略

- `main`: 生产环境
- `develop`: 开发环境
- `feature/*`: 功能分支
- `hotfix/*`: 紧急修复

## GitHub Actions 配置

创建以下工作流文件：

### .github/workflows/deploy-production.yml

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]
  workflow_dispatch:

env:
  DOCKER_REGISTRY: ${{ secrets.DOCKER_REGISTRY }}
  DOCKER_NAMESPACE: ${{ secrets.DOCKER_NAMESPACE }}

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [backend, write-service, courier-service, ocr-service]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Run tests for ${{ matrix.service }}
      run: |
        cd ${{ matrix.service }}
        if [ -f "go.mod" ]; then
          go test ./...
        elif [ -f "requirements.txt" ]; then
          pip install -r requirements.txt
          pytest
        elif [ -f "pom.xml" ]; then
          mvn test
        fi

  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: 
          - name: frontend
            context: ./frontend
            dockerfile: ./deploy/docker/Dockerfile.frontend.prod
          - name: admin-frontend
            context: ./services/admin-service/frontend
            dockerfile: ./services/admin-service/frontend/Dockerfile
          - name: gateway
            context: ./services/gateway
            dockerfile: ./services/gateway/Dockerfile
          - name: backend
            context: ./backend
            dockerfile: ./backend/Dockerfile
          - name: write-service
            context: ./services/write-service
            dockerfile: ./services/write-service/Dockerfile
          - name: courier-service
            context: ./services/courier-service
            dockerfile: ./services/courier-service/Dockerfile
          - name: admin-service
            context: ./services/admin-service/backend
            dockerfile: ./services/admin-service/backend/Dockerfile
          - name: ocr-service
            context: ./services/ocr-service
            dockerfile: ./services/ocr-service/Dockerfile
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Tencent Cloud Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.DOCKER_REGISTRY }}
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    
    - name: Build and push ${{ matrix.service.name }}
      uses: docker/build-push-action@v4
      with:
        context: ${{ matrix.service.context }}
        file: ${{ matrix.service.dockerfile }}
        push: true
        tags: |
          ${{ env.DOCKER_REGISTRY }}/${{ env.DOCKER_NAMESPACE }}/${{ matrix.service.name }}:latest
          ${{ env.DOCKER_REGISTRY }}/${{ env.DOCKER_NAMESPACE }}/${{ matrix.service.name }}:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Copy deployment files to server
      uses: appleboy/scp-action@v0.1.4
      with:
        host: ${{ secrets.TENCENT_HOST }}
        username: ${{ secrets.TENCENT_USER }}
        key: ${{ secrets.TENCENT_SSH_KEY }}
        source: "deploy/*,scripts/deploy/*"
        target: "/home/ubuntu/openpenpal"
    
    - name: Deploy to Tencent Cloud
      uses: appleboy/ssh-action@v0.1.5
      with:
        host: ${{ secrets.TENCENT_HOST }}
        username: ${{ secrets.TENCENT_USER }}
        key: ${{ secrets.TENCENT_SSH_KEY }}
        script: |
          cd /home/ubuntu/openpenpal
          
          # 更新环境变量
          cat > .env.production << EOF
          DOCKER_REGISTRY=${{ env.DOCKER_REGISTRY }}
          DOCKER_NAMESPACE=${{ env.DOCKER_NAMESPACE }}
          IMAGE_TAG=${{ github.sha }}
          JWT_SECRET=${{ secrets.JWT_SECRET }}
          POSTGRES_PASSWORD=${{ secrets.POSTGRES_PASSWORD }}
          REDIS_PASSWORD=${{ secrets.REDIS_PASSWORD }}
          GRAFANA_PASSWORD=${{ secrets.GRAFANA_PASSWORD }}
          EOF
          
          # 执行部署脚本
          chmod +x scripts/deploy/deploy.sh
          ./scripts/deploy/deploy.sh production
    
    - name: Health Check
      uses: appleboy/ssh-action@v0.1.5
      with:
        host: ${{ secrets.TENCENT_HOST }}
        username: ${{ secrets.TENCENT_USER }}
        key: ${{ secrets.TENCENT_SSH_KEY }}
        script: |
          cd /home/ubuntu/openpenpal
          ./scripts/deploy/health-check.sh
    
    - name: Notify Success
      if: success()
      uses: 8398a7/action-slack@v3
      with:
        status: success
        text: '🚀 OpenPenPal deployed successfully to production!'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
    
    - name: Notify Failure
      if: failure()
      uses: 8398a7/action-slack@v3
      with:
        status: failure
        text: '❌ OpenPenPal deployment failed!'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}

  cleanup:
    needs: deploy
    runs-on: ubuntu-latest
    if: success()
    
    steps:
    - name: Clean up old images
      uses: appleboy/ssh-action@v0.1.5
      with:
        host: ${{ secrets.TENCENT_HOST }}
        username: ${{ secrets.TENCENT_USER }}
        key: ${{ secrets.TENCENT_SSH_KEY }}
        script: |
          # 保留最近5个版本的镜像
          docker image prune -a --force --filter "until=24h"
```

### .github/workflows/deploy-staging.yml

```yaml
name: Deploy to Staging

on:
  push:
    branches: [develop]
  pull_request:
    branches: [main]

# 类似 production 配置，但部署到测试环境
```

## 腾讯云服务器配置

### 1. 创建部署目录结构

```bash
mkdir -p /home/ubuntu/openpenpal/{deploy,scripts,config,data,logs,backups}
```

### 2. 配置 Nginx

创建 `/etc/nginx/sites-available/openpenpal`:

```nginx
upstream frontend {
    server localhost:3000;
}

upstream api {
    server localhost:8000;
}

upstream admin {
    server localhost:3001;
}

server {
    listen 80;
    server_name openpenpal.com www.openpenpal.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name openpenpal.com www.openpenpal.com;
    
    ssl_certificate /etc/letsencrypt/live/openpenpal.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/openpenpal.com/privkey.pem;
    
    # 主应用
    location / {
        proxy_pass http://frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # API
    location /api {
        proxy_pass http://api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket
    location /ws {
        proxy_pass http://api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # 管理后台
    location /admin {
        proxy_pass http://admin;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # 静态文件
    location /uploads {
        alias /home/ubuntu/openpenpal/data/uploads;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

启用站点：

```bash
sudo ln -s /etc/nginx/sites-available/openpenpal /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## 部署脚本配置

### scripts/deploy/deploy.sh

```bash
#!/bin/bash
set -e

ENV=${1:-production}
DEPLOY_DIR="/home/ubuntu/openpenpal"
BACKUP_DIR="$DEPLOY_DIR/backups/$(date +%Y%m%d_%H%M%S)"

echo "🚀 Starting deployment for environment: $ENV"

# 加载环境变量
source $DEPLOY_DIR/.env.$ENV

# 创建备份
echo "📦 Creating backup..."
mkdir -p $BACKUP_DIR
docker-compose -f $DEPLOY_DIR/deploy/docker-compose.$ENV.yml ps -q | xargs docker inspect > $BACKUP_DIR/containers.json

# 拉取最新镜像
echo "🔄 Pulling latest images..."
docker-compose -f $DEPLOY_DIR/deploy/docker-compose.$ENV.yml pull

# 蓝绿部署
echo "🔵 Starting blue-green deployment..."

# 启动绿环境
docker-compose -f $DEPLOY_DIR/deploy/docker-compose.$ENV.yml \
  -p openpenpal_green \
  up -d

# 等待健康检查
echo "🏥 Waiting for health checks..."
sleep 30

# 健康检查
if ! $DEPLOY_DIR/scripts/deploy/health-check.sh green; then
  echo "❌ Health check failed! Rolling back..."
  docker-compose -f $DEPLOY_DIR/deploy/docker-compose.$ENV.yml \
    -p openpenpal_green \
    down
  exit 1
fi

# 切换流量到绿环境
echo "🔀 Switching traffic to green environment..."
docker-compose -f $DEPLOY_DIR/deploy/docker-compose.$ENV.yml \
  -p openpenpal_blue \
  down

# 重命名绿环境为蓝环境
docker-compose -f $DEPLOY_DIR/deploy/docker-compose.$ENV.yml \
  -p openpenpal_green \
  ps -q | xargs -I {} docker rename {} {}_blue

echo "✅ Deployment completed successfully!"

# 清理旧镜像
echo "🧹 Cleaning up old images..."
docker image prune -a --force --filter "until=72h"

# 保留最近10个备份
echo "📁 Cleaning old backups..."
ls -t $DEPLOY_DIR/backups | tail -n +11 | xargs -I {} rm -rf $DEPLOY_DIR/backups/{}
```

### scripts/deploy/health-check.sh

```bash
#!/bin/bash
set -e

ENV=${1:-blue}
BASE_URL="http://localhost"

echo "🏥 Running health checks for $ENV environment..."

# 检查所有服务
services=(
  "frontend:3000/"
  "api-gateway:8000/health"
  "backend:8080/health"
  "write-service:8001/health"
  "courier-service:8002/health"
  "admin-service:8003/actuator/health"
  "ocr-service:8004/health"
)

failed=0

for service in "${services[@]}"; do
  IFS=':' read -r name port_path <<< "$service"
  url="$BASE_URL:$port_path"
  
  echo -n "Checking $name... "
  
  if curl -f -s -o /dev/null -w "%{http_code}" --connect-timeout 5 --max-time 10 "$url" | grep -q "200\|204"; then
    echo "✅ OK"
  else
    echo "❌ FAILED"
    failed=$((failed + 1))
  fi
done

# 检查数据库连接
echo -n "Checking PostgreSQL... "
if docker exec openpenpal-postgres pg_isready -U openpenpal -d openpenpal > /dev/null 2>&1; then
  echo "✅ OK"
else
  echo "❌ FAILED"
  failed=$((failed + 1))
fi

# 检查 Redis
echo -n "Checking Redis... "
if docker exec openpenpal-redis redis-cli ping > /dev/null 2>&1; then
  echo "✅ OK"
else
  echo "❌ FAILED"
  failed=$((failed + 1))
fi

if [ $failed -gt 0 ]; then
  echo "❌ Health check failed! $failed services are down."
  exit 1
else
  echo "✅ All services are healthy!"
  exit 0
fi
```

### scripts/deploy/rollback.sh

```bash
#!/bin/bash
set -e

BACKUP_ID=${1:-latest}
DEPLOY_DIR="/home/ubuntu/openpenpal"

echo "🔄 Starting rollback to backup: $BACKUP_ID"

if [ "$BACKUP_ID" == "latest" ]; then
  BACKUP_DIR=$(ls -t $DEPLOY_DIR/backups | head -1)
else
  BACKUP_DIR=$BACKUP_ID
fi

if [ ! -d "$DEPLOY_DIR/backups/$BACKUP_DIR" ]; then
  echo "❌ Backup not found: $BACKUP_DIR"
  exit 1
fi

echo "📦 Restoring from backup: $BACKUP_DIR"

# 停止当前容器
docker-compose -f $DEPLOY_DIR/deploy/docker-compose.production.yml down

# 恢复容器配置
# 实现具体的恢复逻辑...

echo "✅ Rollback completed!"
```

## 监控与告警

### 1. 配置 Prometheus 告警规则

创建 `monitoring/alerts.yml`:

```yaml
groups:
  - name: openpenpal
    rules:
      - alert: ServiceDown
        expr: up{job=~"openpenpal-.*"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
          
      - alert: HighMemoryUsage
        expr: container_memory_usage_bytes{name=~"openpenpal-.*"} / container_spec_memory_limit_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage for {{ $labels.name }}"
          
      - alert: HighCPUUsage
        expr: rate(container_cpu_usage_seconds_total{name=~"openpenpal-.*"}[5m]) > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage for {{ $labels.name }}"
```

### 2. 配置日志收集

使用 Loki + Promtail 收集日志：

```yaml
# docker-compose.monitoring.yml 添加
loki:
  image: grafana/loki:latest
  ports:
    - "3100:3100"
  volumes:
    - ./monitoring/loki-config.yml:/etc/loki/local-config.yaml
    - loki_data:/loki

promtail:
  image: grafana/promtail:latest
  volumes:
    - ./monitoring/promtail-config.yml:/etc/promtail/config.yml
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
    - /var/run/docker.sock:/var/run/docker.sock
```

## 故障恢复

### 1. 自动故障转移

使用 Docker Swarm 或 Kubernetes 实现自动故障转移：

```bash
# 初始化 Swarm
docker swarm init

# 部署服务
docker stack deploy -c docker-compose.production.yml openpenpal
```

### 2. 数据备份策略

创建定时备份任务：

```bash
# 创建备份脚本
cat > /home/ubuntu/openpenpal/scripts/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/home/ubuntu/openpenpal/backups/data/$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR

# 备份数据库
docker exec openpenpal-postgres pg_dump -U openpenpal openpenpal | gzip > $BACKUP_DIR/postgres.sql.gz

# 备份上传文件
tar -czf $BACKUP_DIR/uploads.tar.gz /home/ubuntu/openpenpal/data/uploads

# 上传到对象存储（可选）
# coscli cp $BACKUP_DIR cos://backup-bucket/openpenpal/

# 清理旧备份
find /home/ubuntu/openpenpal/backups/data -type d -mtime +7 -exec rm -rf {} +
EOF

# 添加到 crontab
crontab -e
# 0 2 * * * /home/ubuntu/openpenpal/scripts/backup.sh
```

## 最佳实践

### 1. 安全建议

- 使用 HashiCorp Vault 管理敏感配置
- 启用容器安全扫描
- 限制容器权限
- 使用私有镜像仓库
- 定期更新基础镜像

### 2. 性能优化

- 使用 CDN 加速静态资源
- 启用 HTTP/2 和 Brotli 压缩
- 配置数据库连接池
- 使用 Redis 缓存热点数据
- 水平扩展无状态服务

### 3. 监控指标

关键指标：
- 响应时间 (P50, P95, P99)
- 错误率
- 并发用户数
- 数据库连接数
- 内存和 CPU 使用率

### 4. 灾难恢复计划

- RTO (恢复时间目标): < 30分钟
- RPO (恢复点目标): < 1小时
- 多地域备份
- 定期演练恢复流程

## 部署检查清单

- [ ] 所有 GitHub Secrets 已配置
- [ ] SSL 证书已安装
- [ ] 数据库备份策略已实施
- [ ] 监控和告警已配置
- [ ] 日志收集已启用
- [ ] 安全组规则已配置
- [ ] 域名解析已配置
- [ ] 健康检查端点可访问
- [ ] 回滚流程已测试
- [ ] 团队成员已培训

## 故障排查

### 常见问题

1. **镜像拉取失败**
   - 检查 Docker Registry 凭据
   - 确认网络连接正常
   - 检查镜像名称和标签

2. **服务启动失败**
   - 查看容器日志: `docker logs <container>`
   - 检查环境变量配置
   - 确认端口未被占用

3. **数据库连接失败**
   - 验证连接字符串
   - 检查网络配置
   - 确认数据库服务运行正常

4. **健康检查失败**
   - 增加启动等待时间
   - 检查服务依赖关系
   - 验证健康检查端点

## 联系支持

- 技术支持邮箱: tech@openpenpal.com
- 紧急联系电话: +86-xxx-xxxx-xxxx
- Slack 频道: #openpenpal-ops

---

最后更新: 2025-08-16