# OpenPenPal 部署配置

## 目录结构

```
deploy/
├── README.md                          # 部署使用指南
├── DEPLOYMENT_GUIDE.md               # 详细部署文档  
├── docker-compose.production.yml     # 生产环境配置
├── docker-compose.microservices.yml  # 微服务开发配置
├── docker-compose.dev.yml           # 开发环境配置
├── docker-compose.monitoring.yml    # 监控服务配置
├── github-workflows/                 # GitHub Actions 工作流
│   ├── deploy-production.yml        # 生产环境部署
│   └── deploy-staging.yml           # 测试环境部署
├── scripts/                         # 部署脚本
│   ├── deploy-blue-green.sh         # 蓝绿部署脚本
│   ├── health-check.sh              # 健康检查脚本
│   ├── backup.sh                    # 备份脚本
│   ├── rollback.sh                  # 回滚脚本
│   └── tencent-cloud-setup.sh       # 腾讯云初始化脚本
├── docker/                          # Docker 配置
│   └── Dockerfile.frontend.prod     # 前端生产镜像
└── k8s/                             # Kubernetes 配置
    └── namespace.yaml               # 命名空间配置
```

## 快速开始

### 1. 使用 GitHub Actions 自动部署

**前提条件**：
- 腾讯云服务器已准备就绪
- GitHub Secrets 已配置完成
- 域名已解析到服务器

**部署步骤**：

1. **配置 GitHub Secrets**：
   ```bash
   # 在 GitHub 仓库的 Settings > Secrets 中添加：
   TENCENT_HOST=你的服务器IP
   TENCENT_USER=ubuntu
   TENCENT_SSH_KEY=SSH私钥内容
   DOCKER_REGISTRY=ccr.ccs.tencentyun.com
   DOCKER_NAMESPACE=openpenpal
   DOCKER_USERNAME=腾讯云镜像服务用户名
   DOCKER_PASSWORD=腾讯云镜像服务密码
   JWT_SECRET=生产环境JWT密钥
   POSTGRES_PASSWORD=数据库密码
   REDIS_PASSWORD=Redis密码
   GRAFANA_PASSWORD=Grafana密码
   ```

2. **复制工作流文件**：
   ```bash
   # 复制到 GitHub Actions 目录
   mkdir -p .github/workflows
   cp deploy/github-workflows/*.yml .github/workflows/
   ```

3. **推送代码触发部署**：
   ```bash
   git add .
   git commit -m "feat: add CI/CD deployment"
   git push origin main  # 触发生产环境部署
   ```

### 2. 手动部署

**生产环境部署**：
```bash
# 1. 拉取代码
git clone https://github.com/your-username/openpenpal.git
cd openpenpal

# 2. 配置环境变量
cp deploy/.env.example deploy/.env.production
vim deploy/.env.production  # 修改配置

# 3. 执行部署
cd deploy
./scripts/deploy-blue-green.sh production
```

**开发环境部署**：
```bash
# 使用开发配置
docker-compose -f deploy/docker-compose.dev.yml up -d
```

## 配置说明

### 环境变量配置

创建 `.env.production` 文件：

```bash
# Docker 镜像配置
DOCKER_REGISTRY=ccr.ccs.tencentyun.com
DOCKER_NAMESPACE=openpenpal
IMAGE_TAG=latest

# 应用配置
JWT_SECRET=your-super-secret-jwt-key-32-chars
POSTGRES_PASSWORD=your-secure-postgres-password
REDIS_PASSWORD=your-secure-redis-password
GRAFANA_PASSWORD=your-secure-grafana-password

# 域名配置
DOMAIN=openpenpal.com
EMAIL=admin@openpenpal.com

# 腾讯云配置
TENCENT_COS_BUCKET=openpenpal-backups
TENCENT_COS_REGION=ap-guangzhou

# 通知配置
SLACK_WEBHOOK=https://hooks.slack.com/...
EMAIL_NOTIFICATION=alerts@openpenpal.com
```

### 服务配置

#### 微服务架构

| 服务 | 端口 | 描述 |
|------|------|------|
| Nginx | 80/443 | 反向代理和 SSL 终端 |
| Frontend | 3000 | Next.js 前端应用 |
| Admin Frontend | 3001 | Vue 3 管理后台 |
| API Gateway | 8000 | 统一 API 网关 |
| Backend | 8080 | Go 主后端服务 |
| Write Service | 8001 | Python 写信服务 |
| Courier Service | 8002 | Go 信使服务 |
| Admin Service | 8003 | Java 管理服务 |
| OCR Service | 8004 | Python OCR 服务 |
| PostgreSQL | 5432 | 主数据库 |
| Redis | 6379 | 缓存和队列 |
| Prometheus | 9090 | 监控数据收集 |
| Grafana | 3002 | 监控可视化 |

#### 健康检查端点

所有服务都提供健康检查端点：
- `/health` - 基础健康检查
- `/ready` - 就绪检查
- `/live` - 存活检查

### 监控配置

#### Prometheus 监控指标

- **系统指标**：CPU、内存、磁盘、网络
- **应用指标**：请求数量、响应时间、错误率
- **业务指标**：用户注册、信件发送、信使任务

#### Grafana 仪表板

- **系统概览**：整体系统状态
- **服务监控**：各微服务详细指标
- **业务分析**：用户行为和业务数据
- **告警视图**：实时告警状态

## 部署策略

### 蓝绿部署

支持零停机时间的蓝绿部署：

```bash
# 执行蓝绿部署
./scripts/deploy-blue-green.sh production

# 部署流程：
# 1. 拉取新镜像
# 2. 启动绿环境
# 3. 健康检查
# 4. 切换流量
# 5. 停止蓝环境
```

### 滚动更新

逐个更新服务实例：

```bash
# 执行滚动更新
./scripts/deploy-rolling.sh production
```

### 金丝雀发布

小流量验证新版本：

```bash
# 执行金丝雀发布
./scripts/deploy-canary.sh production
```

## 运维操作

### 健康检查

```bash
# 检查所有服务状态
./scripts/health-check.sh production

# 检查特定服务
docker-compose -f docker-compose.production.yml ps
```

### 备份操作

```bash
# 手动备份
./scripts/backup.sh manual

# 预部署备份
./scripts/backup.sh pre-deployment

# 查看备份列表
ls -la /home/ubuntu/openpenpal/backups/
```

### 回滚操作

```bash
# 回滚到最新备份
./scripts/rollback.sh latest

# 回滚到特定版本
./scripts/rollback.sh 20240816_143000

# 强制回滚（跳过确认）
FORCE_ROLLBACK=true ./scripts/rollback.sh latest
```

### 日志查看

```bash
# 查看所有服务日志
docker-compose -f docker-compose.production.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.production.yml logs -f backend

# 查看错误日志
docker-compose -f docker-compose.production.yml logs --tail=100 | grep ERROR
```

## 故障排查

### 常见问题

1. **镜像拉取失败**
   ```bash
   # 检查镜像仓库凭据
   docker login ccr.ccs.tencentyun.com
   
   # 检查镜像是否存在
   docker images | grep openpenpal
   ```

2. **服务启动失败**
   ```bash
   # 查看容器状态
   docker ps -a
   
   # 查看启动日志
   docker logs <container-name>
   
   # 检查资源使用
   docker stats
   ```

3. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker exec openpenpal-postgres pg_isready
   
   # 检查连接配置
   docker exec openpenpal-postgres psql -U openpenpal -d openpenpal -c "SELECT 1;"
   ```

4. **网络问题**
   ```bash
   # 检查网络连通性
   docker network ls
   docker network inspect openpenpal_openpenpal-network
   
   # 测试服务间连接
   docker exec openpenpal-backend nc -zv openpenpal-postgres 5432
   ```

### 性能调优

1. **数据库优化**
   ```sql
   -- 检查慢查询
   SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;
   
   -- 检查索引使用
   SELECT schemaname, tablename, indexname, idx_scan FROM pg_stat_user_indexes ORDER BY idx_scan;
   ```

2. **内存优化**
   ```bash
   # 调整 PostgreSQL 内存
   # shared_buffers = 256MB
   # effective_cache_size = 1GB
   
   # 调整 Redis 内存
   # maxmemory 512mb
   # maxmemory-policy allkeys-lru
   ```

3. **网络优化**
   ```bash
   # 调整 Nginx 配置
   # worker_connections 2048;
   # keepalive_timeout 65;
   # client_max_body_size 100M;
   ```

## 安全配置

### SSL/TLS 配置

```bash
# 生成 Diffie-Hellman 参数
openssl dhparam -out dhparam.pem 2048

# 配置 SSL 证书自动续期
systemctl enable certbot-renewal.timer
systemctl start certbot-renewal.timer
```

### 防火墙配置

```bash
# 仅开放必要端口
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw enable
```

### 容器安全

```bash
# 扫描镜像漏洞
trivy image openpenpal/backend:latest

# 检查容器配置
docker-bench-security
```

## 监控告警

### 告警规则

- **服务可用性**：服务下线超过 2 分钟
- **响应时间**：P95 响应时间超过 5 秒
- **错误率**：错误率超过 5%
- **资源使用**：CPU/内存使用率超过 80%
- **磁盘空间**：可用空间少于 20%

### 通知渠道

- **Slack**：实时告警通知
- **邮件**：重要告警和日报
- **短信**：严重故障（可选）

## 备份策略

### 备份内容

- **数据库**：完整备份 + 增量备份
- **文件**：用户上传文件、日志文件
- **配置**：环境配置、证书信息
- **容器**：镜像版本信息

### 备份频率

- **数据库**：每 6 小时增量，每天完整
- **文件**：每天一次
- **配置**：变更时自动备份

### 备份保留

- **本地**：保留 7 天
- **远程**：保留 30 天
- **归档**：月度备份保留 1 年

## 扩容指南

### 水平扩容

```bash
# 增加服务实例
docker-compose -f docker-compose.production.yml up -d --scale backend=3

# 使用 Docker Swarm 扩容
docker service scale openpenpal_backend=3
```

### 垂直扩容

```bash
# 修改资源限制
resources:
  limits:
    memory: 2G
    cpus: '1.0'
  reservations:
    memory: 1G
    cpus: '0.5'
```

## 联系信息

- **技术支持**：tech@openpenpal.com
- **紧急联系**：+86-xxx-xxxx-xxxx
- **文档反馈**：请在 GitHub 提交 Issue

---

*最后更新：2025-08-16*