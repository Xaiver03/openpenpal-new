# OpenPenPal 部署快速指南

## 🚀 一键部署到腾讯云

### 步骤一：服务器初始化

```bash
# 在腾讯云服务器上运行
curl -fsSL https://raw.githubusercontent.com/your-repo/openpenpal/main/deploy/scripts/tencent-cloud-setup.sh | sudo bash -s openpenpal.com admin@openpenpal.com
```

### 步骤二：配置 GitHub Secrets

在 GitHub 仓库的 `Settings > Secrets and variables > Actions` 中添加：

```bash
# 🔐 服务器连接
TENCENT_HOST=你的服务器IP
TENCENT_USER=ubuntu  
TENCENT_SSH_KEY=SSH私钥内容

# 🐳 Docker 镜像仓库
DOCKER_REGISTRY=ccr.ccs.tencentyun.com
DOCKER_NAMESPACE=openpenpal
DOCKER_USERNAME=腾讯云用户名
DOCKER_PASSWORD=腾讯云密码

# 🔑 应用密钥
JWT_SECRET=至少32位的随机字符串
POSTGRES_PASSWORD=数据库密码
REDIS_PASSWORD=Redis密码
GRAFANA_PASSWORD=Grafana密码

# 📢 通知配置（可选）
SLACK_WEBHOOK=https://hooks.slack.com/...
EMAIL_NOTIFICATION=alerts@openpenpal.com
```

### 步骤三：部署工作流

```bash
# 1. 复制 GitHub Actions 配置
mkdir -p .github/workflows
cp deploy/github-workflows/*.yml .github/workflows/

# 2. 提交并推送到 main 分支
git add .
git commit -m "feat: add production deployment"
git push origin main

# 🎉 自动触发部署！
```

## 📊 部署状态检查

### 实时监控
- **应用主页**：https://openpenpal.com
- **管理后台**：https://openpenpal.com/admin  
- **监控面板**：https://openpenpal.com/grafana
- **健康检查**：https://openpenpal.com/health

### 命令行检查
```bash
# SSH 到服务器
ssh ubuntu@你的服务器IP

# 检查服务状态
cd /home/ubuntu/openpenpal
./scripts/health-check.sh

# 查看部署日志
docker-compose logs -f --tail=100
```

## 🛠 常用运维命令

### 部署管理
```bash
# 手动触发部署
gh workflow run "Deploy to Production" --ref main

# 查看部署历史
gh run list --workflow="Deploy to Production"

# 紧急回滚
ssh ubuntu@服务器IP "./scripts/rollback.sh latest"
```

### 服务管理
```bash
# 重启特定服务
docker-compose restart backend

# 查看服务日志
docker-compose logs -f backend

# 进入容器调试
docker exec -it openpenpal-backend /bin/bash
```

### 备份恢复
```bash
# 创建备份
./scripts/backup.sh manual

# 查看备份列表
ls -la /home/ubuntu/openpenpal/backups/

# 恢复备份
./scripts/rollback.sh backup_20240816_143000
```

## 🔧 环境配置

### 生产环境 (.env.production)
```bash
# 基础配置
ENVIRONMENT=production
DOMAIN=openpenpal.com
SSL_ENABLED=true

# 数据库配置
DATABASE_URL=postgres://openpenpal:PASSWORD@postgres:5432/openpenpal
REDIS_URL=redis://:PASSWORD@redis:6379

# 性能配置
MAX_CONNECTIONS=100
POOL_SIZE=20
CACHE_TTL=3600

# 安全配置
JWT_EXPIRY=24h
SESSION_TIMEOUT=1h
RATE_LIMIT=1000/hour
```

### 开发环境 (.env.development)
```bash
# 基础配置
ENVIRONMENT=development
DOMAIN=localhost
SSL_ENABLED=false

# 调试配置
DEBUG=true
LOG_LEVEL=debug
HOT_RELOAD=true

# 测试配置
TEST_DB_URL=postgres://test:test@postgres:5432/openpenpal_test
MOCK_SERVICES=true
```

## 📈 扩容指南

### 自动扩容
```yaml
# docker-compose.yml 中添加
deploy:
  replicas: 3
  update_config:
    parallelism: 1
    delay: 10s
  restart_policy:
    condition: on-failure
```

### 负载均衡
```nginx
# nginx.conf
upstream backend_pool {
    least_conn;
    server backend-1:8080 weight=3;
    server backend-2:8080 weight=2;
    server backend-3:8080 weight=1;
}
```

### 数据库扩容
```bash
# 添加只读副本
docker run -d \
  --name postgres-replica \
  --env PGUSER=replica \
  --env POSTGRES_PASSWORD=password \
  postgres:15-alpine

# 配置流复制
echo "host replication replica 0.0.0.0/0 md5" >> pg_hba.conf
```

## 🔍 故障排查

### 常见问题速查表

| 问题 | 症状 | 解决方案 |
|------|------|----------|
| 服务无法访问 | 502/503 错误 | 检查容器状态，重启服务 |
| 数据库连接失败 | 连接超时 | 检查数据库容器，验证配置 |
| 内存不足 | 服务响应慢 | 增加内存限制，清理缓存 |
| 磁盘空间不足 | 无法写入 | 清理日志，扩容磁盘 |
| SSL 证书过期 | HTTPS 警告 | 运行 certbot renew |

### 紧急处理流程
```bash
# 1. 快速诊断
./scripts/health-check.sh | grep "FAILED"

# 2. 查看错误日志
docker-compose logs --tail=50 | grep ERROR

# 3. 重启问题服务
docker-compose restart <service-name>

# 4. 如果问题持续，执行回滚
./scripts/rollback.sh latest

# 5. 通知团队
curl -X POST $SLACK_WEBHOOK -d '{"text":"🚨 Production issue detected"}'
```

## 📋 部署检查清单

### 部署前检查
- [ ] GitHub Secrets 已配置
- [ ] 服务器资源充足（8GB+ RAM, 100GB+ 磁盘）
- [ ] 域名已解析到服务器
- [ ] SSL 证书已配置
- [ ] 备份策略已实施

### 部署后验证
- [ ] 所有服务健康检查通过
- [ ] 前端页面可正常访问
- [ ] 用户注册登录功能正常
- [ ] 数据库连接正常
- [ ] 监控指标正常
- [ ] 备份任务正常运行

### 性能检查
- [ ] 页面加载时间 < 3 秒
- [ ] API 响应时间 < 500ms
- [ ] 内存使用率 < 80%
- [ ] CPU 使用率 < 70%
- [ ] 磁盘使用率 < 80%

## 🎯 性能优化建议

### 前端优化
```javascript
// next.config.js
module.exports = {
  experimental: {
    optimizeCss: true,
    optimizeImages: true,
  },
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },
}
```

### 数据库优化
```sql
-- 添加必要索引
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY idx_letters_created_at ON letters(created_at);

-- 配置连接池
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
```

### 缓存策略
```bash
# Redis 配置优化
maxmemory 1gb
maxmemory-policy allkeys-lru
timeout 300
tcp-keepalive 60
```

## 📞 支持联系

### 获取帮助
- **📚 文档**：查看 [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md)
- **🐛 问题反馈**：在 GitHub 提交 Issue
- **💬 技术交流**：加入 Slack 频道 #openpenpal-ops
- **📧 邮件支持**：tech@openpenpal.com

### 紧急联系
- **🚨 生产故障**：+86-xxx-xxxx-xxxx
- **🔒 安全问题**：security@openpenpal.com
- **📊 性能问题**：performance@openpenpal.com

---

**🎉 恭喜！你已完成 OpenPenPal 的生产环境部署。现在可以体验完整的校园手写信平台了！**

*最后更新：2025-08-16*