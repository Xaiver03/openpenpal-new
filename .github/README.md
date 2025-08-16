# GitHub Actions 工作流说明

## 工作流概览

本项目使用 GitHub Actions 实现完整的 CI/CD 流水线，包含以下工作流：

### 🚀 部署工作流

#### `deploy-production.yml` - 生产环境部署
**触发条件**：
- 推送到 `main` 分支
- 手动触发（可选择部署模式）

**部署模式**：
- `blue-green`：蓝绿部署（默认，零停机）
- `rolling`：滚动更新
- `canary`：金丝雀发布

**流程**：
1. 代码质量检查（安全扫描）
2. 单元测试（所有微服务）
3. 构建镜像并推送到腾讯云镜像仓库
4. 集成测试
5. 部署到生产环境
6. 健康检查
7. 通知和回滚准备

#### `deploy-staging.yml` - 测试环境部署
**触发条件**：
- 推送到 `develop` 或 `feature/*` 分支
- Pull Request 到 `main` 分支

**流程**：
1. 代码检查（Lint + 安全扫描）
2. 构建和测试
3. 构建镜像
4. 部署到测试环境
5. 冒烟测试
6. 性能测试（仅 PR）
7. PR 评论反馈

### 🔧 持续集成工作流

#### `ci-enhanced.yml` - 增强版 CI 流水线
**触发条件**：
- 推送到 `main`、`develop`、`feature/*` 分支
- Pull Request

**阶段**：
1. **代码质量检查**
   - ESLint（前端）
   - TypeScript 检查
   - Go Lint（golangci-lint）
   - Python Lint（flake8, black, pylint）
   - 安全扫描（Bandit, npm audit, Trivy）

2. **测试执行**
   - 前端测试（Jest + 覆盖率）
   - 后端测试（Go + 覆盖率）
   - Python 服务测试（pytest + 覆盖率）
   - Java 服务测试（Maven + JaCoCo）

3. **集成测试**
   - 跨服务集成验证
   - 数据库集成测试

4. **构建验证**
   - 所有服务构建检查
   - 构建产物上传

### 🔒 安全工作流

#### `security-scan.yml` - 安全扫描
**触发条件**：
- 推送到 `main`、`develop` 分支
- Pull Request 到 `main`
- 定时任务（每日凌晨 2 点）
- 手动触发

**扫描类型**：
1. **依赖漏洞扫描**
   - npm 依赖（frontend）
   - Go modules（backend services）
   - Python packages（Python services）
   - Maven 依赖（Java service）

2. **代码安全分析**
   - CodeQL 静态分析
   - 多语言支持（JavaScript, Go, Python, Java）

3. **秘密检测**
   - TruffleHog 扫描
   - 检测泄露的凭据和密钥

4. **容器安全扫描**
   - Trivy 漏洞扫描
   - Dockle 最佳实践检查

5. **基础设施安全**
   - Checkov IaC 扫描
   - Terrascan 配置检查
   - 网络配置安全检查

### 📊 工作流矩阵

| 工作流 | 触发条件 | 运行时间 | 关键功能 |
|--------|----------|----------|----------|
| `ci-enhanced.yml` | Push/PR | ~15-20分钟 | 代码质量 + 测试 |
| `security-scan.yml` | Push/PR/定时 | ~10-15分钟 | 安全扫描 |
| `deploy-staging.yml` | develop/feature/PR | ~20-25分钟 | 测试环境部署 |
| `deploy-production.yml` | main | ~25-30分钟 | 生产环境部署 |

## 🔑 必需的 GitHub Secrets

### 腾讯云配置
```bash
TENCENT_HOST=你的服务器IP
TENCENT_USER=ubuntu
TENCENT_SSH_KEY=SSH私钥内容

# 镜像仓库
DOCKER_REGISTRY=ccr.ccs.tencentyun.com
DOCKER_NAMESPACE=openpenpal
DOCKER_USERNAME=腾讯云用户名
DOCKER_PASSWORD=腾讯云密码
```

### 应用配置
```bash
JWT_SECRET=生产环境JWT密钥
POSTGRES_PASSWORD=数据库密码
REDIS_PASSWORD=Redis密码
GRAFANA_PASSWORD=Grafana密码

# 测试环境
STAGING_HOST=测试服务器IP
STAGING_USER=ubuntu
STAGING_SSH_KEY=测试服务器SSH私钥
STAGING_JWT_SECRET=测试环境JWT密钥
STAGING_POSTGRES_PASSWORD=测试数据库密码
STAGING_REDIS_PASSWORD=测试Redis密码
```

### 通知配置（可选）
```bash
SLACK_WEBHOOK=Slack通知URL
EMAIL_USERNAME=邮件用户名
EMAIL_PASSWORD=邮件密码
EMAIL_NOTIFICATION=通知邮箱

# 安全扫描
NVD_API_KEY=NVD漏洞数据库API密钥
```

## 🎯 最佳实践

### 分支策略
- `main`：生产环境，稳定版本
- `develop`：开发环境，集成分支
- `feature/*`：功能开发分支
- `hotfix/*`：紧急修复分支

### 提交规范
- `feat: 新功能`
- `fix: 修复问题`
- `docs: 文档更新`
- `style: 代码格式`
- `refactor: 重构`
- `test: 测试相关`
- `chore: 构建/工具相关`

### PR 工作流
1. 创建 feature 分支
2. 开发并提交代码
3. 推送触发 staging 部署
4. 创建 PR 到 main
5. 自动测试和安全扫描
6. 代码审查
7. 合并到 main 触发生产部署

## 🚨 故障处理

### 部署失败
1. 查看 GitHub Actions 日志
2. 检查服务器状态
3. 执行回滚命令：
   ```bash
   ssh ubuntu@服务器IP "./scripts/rollback.sh latest"
   ```

### 安全告警
1. 查看 Security 标签页
2. 修复高危漏洞
3. 更新依赖版本
4. 重新运行安全扫描

### 测试失败
1. 查看失败的测试日志
2. 本地复现问题
3. 修复并重新推送
4. 等待自动重新测试

## 📈 监控指标

每个工作流都会生成以下指标：
- **构建时间**：各阶段耗时统计
- **测试覆盖率**：代码覆盖率报告
- **安全评分**：漏洞扫描结果
- **部署成功率**：成功/失败统计

## 🔄 工作流优化

### 缓存策略
- Node modules 缓存
- Go modules 缓存
- pip 缓存
- Maven 缓存
- Docker 层缓存

### 并行执行
- 多服务并行测试
- 多平台并行构建
- 矩阵策略优化

### 资源限制
- 合理的超时设置
- 适当的并发限制
- 缓存有效期管理

---

## 📞 支持

如有问题，请：
1. 查看工作流运行日志
2. 检查本文档的故障处理部分
3. 在 GitHub Issues 中报告问题
4. 联系技术团队

*最后更新：2025-08-16*