# 项目重组计划

## 当前问题
1. 多个独立项目混杂在一起（mall4cloud、funNLP、wangEditor等）
2. OpenPenPal项目的核心目录结构不够清晰
3. 缺少统一的项目管理和构建系统

## 重组方案

### 第一步：分离无关项目
将以下目录移出主项目：
- `mall4cloud-master/` → 移至独立仓库或archive目录
- `funNLP-master/` → 移至独立仓库或archive目录
- `wangEditor-master/` → 移至独立仓库或archive目录
- `ui-main/` → 评估是否属于核心项目
- `multi-agent/` → 移至tools或独立仓库

### 第二步：规范化项目结构
```
openpenpal/
├── apps/                    # 应用层
│   ├── web/                # Web前端 (当前的frontend/)
│   ├── mobile/             # 移动端（预留）
│   └── admin/              # 管理后台
├── services/               # 微服务层
│   ├── api-gateway/        # API网关
│   ├── user-service/       # 用户服务
│   ├── courier-service/    # 信使服务
│   ├── write-service/      # 写信服务
│   ├── admin-service/      # 管理服务
│   └── ocr-service/        # OCR服务
├── packages/               # 共享包
│   ├── common/             # 公共工具库
│   ├── ui-components/      # UI组件库
│   └── api-client/         # API客户端SDK
├── infrastructure/         # 基础设施
│   ├── docker/            # Docker配置
│   ├── k8s/               # Kubernetes配置
│   ├── terraform/         # 基础设施即代码
│   └── monitoring/        # 监控配置
├── tools/                 # 开发工具
│   ├── scripts/           # 自动化脚本
│   ├── cli/               # 命令行工具
│   └── generators/        # 代码生成器
├── docs/                  # 文档中心
│   ├── architecture/      # 架构文档
│   ├── api/              # API文档
│   ├── guides/           # 使用指南
│   └── adr/              # 架构决策记录
├── tests/                # 测试套件
│   ├── unit/             # 单元测试
│   ├── integration/      # 集成测试
│   ├── e2e/              # 端到端测试
│   └── performance/      # 性能测试
├── .github/              # GitHub配置
│   ├── workflows/        # CI/CD工作流
│   ├── ISSUE_TEMPLATE/   # Issue模板
│   └── PULL_REQUEST_TEMPLATE.md
├── scripts/              # 顶层脚本
│   ├── setup.sh         # 环境设置
│   ├── build.sh         # 构建脚本
│   └── deploy.sh        # 部署脚本
├── Makefile             # 统一构建入口
├── docker-compose.yml   # 开发环境编排
├── docker-compose.prod.yml # 生产环境编排
├── lerna.json          # Monorepo配置
├── package.json        # 根package.json
├── .gitignore
├── .env.example
├── LICENSE
├── CONTRIBUTING.md
└── README.md
```

### 第三步：引入Monorepo管理
使用 Lerna + Yarn Workspaces 管理多包项目：
- 统一依赖管理
- 统一版本控制
- 统一构建流程
- 跨包引用支持

### 第四步：标准化开发流程
1. **统一构建系统**：使用Makefile作为顶层构建入口
2. **统一脚本命令**：所有项目使用相同的命令规范
3. **统一配置管理**：环境变量和配置文件标准化
4. **统一代码规范**：ESLint、Prettier、Golint等

## 实施步骤

### Phase 1: 清理和分离（1天）
- [ ] 备份当前项目
- [ ] 分离无关项目到archive目录
- [ ] 清理.DS_Store等系统文件
- [ ] 更新.gitignore

### Phase 2: 重构目录结构（2天）
- [ ] 创建新的目录结构
- [ ] 移动现有代码到新结构
- [ ] 更新所有import路径
- [ ] 验证各服务能正常启动

### Phase 3: 引入工具链（2天）
- [ ] 配置Lerna和Yarn Workspaces
- [ ] 创建统一的Makefile
- [ ] 配置CI/CD工作流
- [ ] 设置代码质量检查工具

### Phase 4: 文档和测试（1天）
- [ ] 更新所有文档
- [ ] 确保测试通过
- [ ] 创建迁移指南
- [ ] 更新README

## 预期收益

1. **开发效率提升**
   - 统一的构建和启动流程
   - 更快的依赖安装
   - 更好的IDE支持

2. **维护性改善**
   - 清晰的项目边界
   - 规范的目录结构
   - 统一的配置管理

3. **团队协作增强**
   - 标准化的开发流程
   - 清晰的代码组织
   - 更好的代码重用

4. **部署简化**
   - 统一的容器化方案
   - 标准化的部署流程
   - 更好的环境隔离