# 部署目录整合迁移指南

## 📁 整合说明

我们已经将原有的分散部署文件整合到统一的 `scripts/deployment/` 目录中：

### 原有目录结构
```
openpenpal/
├── deploy/                    # 原部署配置目录（较完善）
├── deployment/               # 原简单k8s配置目录  
└── scripts/deployment/       # 新的统一部署目录
```

### 整合后的统一结构
```
scripts/deployment/
├── README.md                          # 完整使用文档
├── MIGRATION_GUIDE.md                 # 本迁移指南
├── validate-env.js                   # 环境变量校验
├── db-migrate.sh                     # 数据库迁移
├── local-dev.sh                      # 本地联调
├── build-verify.sh                   # 构建验证
├── pre-release-checklist.md          # 发布检查清单
├── ci-cd-workflow.yml                # CI/CD工作流
├── docker-compose.production.yml     # 生产环境配置
├── docker-compose.dev.yml            # 开发环境配置
├── docker-compose.microservices.yml  # 微服务配置
├── docker-compose.monitoring.yml     # 监控配置
├── DEPLOYMENT_GUIDE.md               # 详细部署文档
├── QUICK_START.md                    # 快速开始
├── github-workflows/                 # GitHub Actions
│   ├── deploy-production.yml
│   └── deploy-staging.yml
├── scripts/                          # 部署脚本
│   ├── deploy-blue-green.sh
│   ├── health-check.sh
│   ├── backup.sh
│   ├── rollback.sh
│   └── tencent-cloud-setup.sh
├── docker/                          # Docker配置
│   └── Dockerfile.frontend.prod
└── kubernetes/                      # K8s配置
    ├── deployment.yaml
    └── namespace.yaml
```

## 🔄 迁移步骤

### 1. 更新脚本路径引用

如果你的项目中有引用原来路径的地方，需要更新：

```bash
# 旧路径引用
./deploy/scripts/deploy-blue-green.sh
./deployment/kubernetes/deployment.yaml

# 新路径引用
./scripts/deployment/scripts/deploy-blue-green.sh
./scripts/deployment/kubernetes/deployment.yaml
```

### 2. 更新GitHub Actions工作流

如果你已经在使用GitHub Actions，需要更新路径：

```bash
# 复制新的工作流文件
mkdir -p .github/workflows
cp scripts/deployment/github-workflows/*.yml .github/workflows/

# 或者手动更新现有工作流中的路径引用
```

### 3. 更新CI/CD脚本

检查你的CI/CD配置文件，更新部署脚本路径：

```yaml
# 示例：更新GitHub Actions中的路径
- name: Deploy
  run: ./scripts/deployment/scripts/deploy-blue-green.sh production
```

### 4. 更新环境变量和配置

新的统一配置提供了更完善的环境变量管理：

```bash
# 使用新的环境变量校验
./scripts/deployment/validate-env.js

# 使用新的本地开发环境
./scripts/deployment/local-dev.sh
```

## 🆕 新增功能

整合后增加了以下新功能：

### 1. 环境变量校验
- 自动校验42个环境变量
- 格式验证和安全检查
- 自动生成.env.example

### 2. 本地一键联调
- 自动启动所有依赖服务
- 智能健康检查
- 集成测试流程

### 3. 构建验证
- 完整的构建流水线
- 安全扫描集成
- 性能测试支持

### 4. 发布检查清单
- 98项检查要点
- 标准化发布流程
- Go/No-Go决策支持

### 5. 增强的CI/CD
- 蓝绿部署支持
- 安全扫描集成
- 多环境部署管理

## 📝 配置兼容性

### Docker Compose配置升级

新的docker-compose配置支持完整的微服务架构：

```yaml
# 新增服务
- API Gateway (8000)
- Write Service (8001) 
- Courier Service (8002)
- Admin Service (8003)
- OCR Service (8004)

# 增强监控
- Prometheus + Grafana
- Loki + Promtail日志收集
- 健康检查配置
```

### 环境变量兼容

保持了原有环境变量的兼容性，同时新增：

```bash
# 原有变量仍然支持
JWT_SECRET=
POSTGRES_PASSWORD=
REDIS_PASSWORD=

# 新增变量
MOONSHOT_API_KEY=
SILICON_FLOW_API_KEY=
PROMETHEUS_ENABLED=
GRAFANA_ENABLED=
```

## 🧹 清理原有目录

完成迁移验证后，可以清理原有目录：

```bash
# 确认新配置工作正常后执行
# rm -rf deploy/
# rm -rf deployment/

# 建议先重命名，观察一段时间无问题再删除
mv deploy/ deploy.backup/
mv deployment/ deployment.backup/
```

## ⚠️ 注意事项

### 1. 生产环境迁移

**在生产环境应用新配置前，务必：**

- 在测试环境完整验证新流程
- 备份现有配置和数据
- 准备回滚方案
- 通知相关团队

### 2. 路径依赖检查

检查以下可能的路径依赖：

- [ ] Dockerfile中的COPY路径
- [ ] 脚本中的相对路径引用
- [ ] 配置文件中的路径设置
- [ ] 文档中的路径说明

### 3. 权限设置

确保新脚本有执行权限：

```bash
chmod +x scripts/deployment/*.sh
chmod +x scripts/deployment/scripts/*.sh
```

### 4. 环境变量迁移

使用新的环境变量校验工具检查配置：

```bash
# 校验现有环境变量
./scripts/deployment/validate-env.js

# 生成新的配置模板
# 会自动创建 .env.example 文件
```

## 🔧 故障排查

### 1. 路径不存在错误

```bash
# 错误：./deploy/scripts/xxx.sh: No such file or directory
# 解决：更新为新路径
./scripts/deployment/scripts/xxx.sh
```

### 2. 权限被拒绝

```bash
# 错误：Permission denied
# 解决：设置执行权限
chmod +x scripts/deployment/*.sh
```

### 3. 环境变量错误

```bash
# 使用新的校验工具检查
./scripts/deployment/validate-env.js
```

### 4. Docker配置问题

```bash
# 使用新的本地联调工具测试
./scripts/deployment/local-dev.sh
```

## 📞 获取帮助

如果在迁移过程中遇到问题：

1. 查看 `scripts/deployment/README.md` 的详细文档
2. 运行 `./scripts/deployment/local-dev.sh --help` 查看帮助
3. 检查项目的 `CLAUDE.md` 文件
4. 在项目仓库提交Issue

## ✅ 迁移验证清单

- [ ] 更新了所有脚本路径引用
- [ ] 测试了本地开发环境启动
- [ ] 验证了环境变量配置
- [ ] 运行了构建验证流程
- [ ] 测试了数据库迁移脚本
- [ ] 更新了CI/CD工作流
- [ ] 通知了团队成员
- [ ] 备份了原有配置

---

*这次整合统一了部署工具，提供了更完善的开发和运维体验。*