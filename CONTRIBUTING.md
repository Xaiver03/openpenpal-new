# 贡献指南

感谢您对 OpenPenPal 项目的关注！本指南将帮助您了解如何为项目做出贡献。

## 🚀 快速开始

### 环境要求

- **Node.js**: 18.0+ 
- **Go**: 1.21+
- **Python**: 3.8+
- **Java**: 17+ (可选，仅Admin服务)
- **Docker**: 最新版本
- **Git**: 2.20+

### 本地开发设置

1. **Fork 项目**
   ```bash
   git clone https://github.com/your-username/openpenpal.git
   cd openpenpal
   ```

2. **安装依赖**
   ```bash
   make install
   # 或者使用安装脚本
   ./startup/install-deps.sh
   ```

3. **启动开发环境**
   ```bash
   make dev
   # 或者使用启动脚本
   ./startup/quick-start.sh demo --auto-open
   ```

4. **验证安装**
   ```bash
   make status
   ```

## 📋 贡献类型

我们欢迎以下类型的贡献：

### 🐛 Bug 报告
- 使用 [Bug Report 模板](.github/ISSUE_TEMPLATE/bug_report.md)
- 提供详细的重现步骤
- 包含系统环境信息

### ✨ 功能请求
- 使用 [Feature Request 模板](.github/ISSUE_TEMPLATE/feature_request.md)
- 详细描述功能需求和使用场景
- 考虑对现有功能的影响

### 📚 文档改进
- 修复文档错误
- 添加示例和教程
- 翻译文档

### 💻 代码贡献
- 修复 Bug
- 实现新功能
- 性能优化
- 重构改进

## 🔄 开发流程

### 分支策略

我们使用 GitFlow 分支模型：

- `main`: 生产环境分支，始终保持稳定
- `develop`: 开发分支，功能集成
- `feature/*`: 功能分支
- `hotfix/*`: 紧急修复分支
- `release/*`: 发布分支

### 提交流程

1. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **开发功能**
   - 遵循代码规范
   - 编写测试
   - 更新文档

3. **代码检查**
   ```bash
   make lint      # 代码质量检查
   make test      # 运行测试
   make format    # 格式化代码
   ```

4. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin feature/your-feature-name
   ```

5. **创建 Pull Request**
   - 使用 [PR 模板](.github/PULL_REQUEST_TEMPLATE.md)
   - 详细描述变更内容
   - 关联相关 Issue

## 📝 代码规范

### 提交信息格式

我们使用 [Conventional Commits](https://conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer]
```

**类型 (type):**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

**示例:**
```
feat(auth): add JWT token validation
fix(api): handle null pointer exception in user service
docs(readme): update installation instructions
```

### 代码风格

#### TypeScript/JavaScript
- 使用 ESLint + Prettier
- 2 空格缩进
- 单引号字符串
- 分号结尾

#### Go
- 使用 `gofmt` 格式化
- 遵循 Go 官方编码规范
- 使用 `golangci-lint` 检查

#### Python
- 使用 Black 格式化
- PEP 8 编码规范
- 88 字符行长度限制

#### Java
- 使用 Spotless 格式化
- Google Java Style
- 4 空格缩进

### 测试要求

- **单元测试**: 核心逻辑必须有单元测试
- **集成测试**: API 接口需要集成测试
- **端到端测试**: 关键用户流程需要 E2E 测试
- **覆盖率**: 新代码测试覆盖率不低于 80%

```bash
# 运行所有测试
make test

# 运行覆盖率检查
./scripts/test-coverage.sh

# 运行特定类型测试
make test-unit          # 单元测试
make test-integration   # 集成测试
make test-e2e          # 端到端测试
```

## 🏗️ 项目架构

### 目录结构
```
openpenpal/
├── frontend/          # Next.js 前端
├── backend/           # Go 主后端
├── services/          # 微服务
├── config/           # 配置管理
├── scripts/          # 构建脚本
├── docs/             # 项目文档
└── tests/            # 测试套件
```

### 技术栈
- **前端**: Next.js 14, TypeScript, Tailwind CSS
- **后端**: Go, Gin, GORM
- **数据库**: PostgreSQL, Redis
- **部署**: Docker, Kubernetes

## 🧪 测试指南

### 本地测试

```bash
# 启动测试环境
make dev

# 运行所有测试
make test

# 运行前端测试
cd frontend && npm test

# 运行后端测试
cd backend && go test ./...

# 运行Python服务测试
cd services/write-service && python -m pytest

# 运行Java服务测试
cd services/admin-service && ./mvnw test
```

### 编写测试

#### 前端测试 (Jest + Testing Library)
```typescript
import { render, screen } from '@testing-library/react';
import { Button } from '@/components/Button';

describe('Button', () => {
  it('renders correctly', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });
});
```

#### 后端测试 (Go)
```go
func TestUserService_GetUser(t *testing.T) {
    // Given
    service := NewUserService()
    
    // When
    user, err := service.GetUser(1)
    
    // Then
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}
```

#### Python 测试 (pytest)
```python
import pytest
from app.services.write_service import WriteService

def test_create_letter():
    # Given
    service = WriteService()
    
    # When
    letter = service.create_letter("Hello World")
    
    # Then
    assert letter.content == "Hello World"
    assert letter.id is not None
```

## 🚀 部署和发布

### 本地构建

```bash
# 构建所有服务
make build

# 构建 Docker 镜像
make build-docker

# 运行生产环境
docker-compose up -d
```

### CI/CD 流程

我们的 CI/CD 流程包括：

1. **代码质量检查**: ESLint, golangci-lint, Black, Spotless
2. **自动化测试**: 单元测试、集成测试、E2E 测试
3. **安全扫描**: CodeQL, Trivy, Bandit
4. **构建部署**: Docker 镜像构建和部署

## 📞 获取帮助

### 沟通渠道

- **GitHub Issues**: 报告 Bug 和功能请求
- **GitHub Discussions**: 技术讨论和问答
- **Discord**: 实时聊天和社区交流

### 文档资源

- [架构文档](./docs/architecture/README.md)
- [API 文档](./docs/api/README.md)
- [部署指南](./docs/deployment/README.md)
- [故障排查](./docs/troubleshooting/README.md)

### 开发者工具

- [启动脚本指南](./STARTUP_SCRIPTS.md)
- [项目重组计划](./PROJECT_REORGANIZATION.md)
- [Makefile 命令](./Makefile)

## 🎯 贡献认可

我们感谢每一位贡献者的努力！贡献者将：

- 被列入项目 README 的贡献者列表
- 获得项目 Discord 的特殊角色
- 参与项目重要决策讨论
- 优先获得项目相关机会

## 📄 许可证

通过贡献代码，您同意您的贡献将按照 [MIT License](./LICENSE) 许可证进行许可。

---

**感谢您的贡献，让 OpenPenPal 变得更好！** 🎉