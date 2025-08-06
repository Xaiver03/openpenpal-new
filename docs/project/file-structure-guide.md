# OpenPenPal 项目文件结构规范

## 📁 根目录结构 (ROOT/)

### 核心目录
```
openpenpal/
├── frontend/              # Next.js前端应用
├── backend/               # Go后端API
├── services/              # 微服务目录
│   ├── admin-service/     # 管理服务
│   ├── courier-service/   # 信使服务
│   ├── gateway/          # API网关
│   ├── ocr-service/      # OCR服务
│   └── write-service/    # 写信服务
├── test-kimi/            # 测试套件
├── docs/                 # 项目文档
├── scripts/              # 部署和工具脚本
├── config/               # 配置文件
├── nginx/                # Nginx配置
├── tools/                # 开发工具
├── temp/                 # 临时文件
└── archive/              # 归档文件
```

### 根目录保留文件
- `README.md` - 项目主文档
- `docker-compose.yml` - 生产环境配置
- `docker-compose.dev.yml` - 开发环境配置
- `.env.example` - 环境变量模板
- `.gitignore` - Git忽略规则
- `package.json` - 根npm配置

### 需要移动的松散文件分类

#### 📋 文档类 → docs/
- `COMPONENT_MANAGEMENT.md`
- `DETAILED_AGENT_TASKS.md`
- `MCPBROWSER_USAGE.md`
- `MULTI_AGENT_COORDINATION.md`
- `NEXT_PHASE_TASKS.md`
- `BROWSERMCP_SETUP.md`

#### 🧪 测试类 → test-kimi/
- `test_user_registration.sh`
- `comprehensive_mcp_test.js`
- `mcp_test_script.js`
- `MCP_TEST_REPORT.md`

#### 🖼️ 图片类 → docs/images/
- `auth-nav-check.png`
- `debug-homepage.png`
- `openpenpal-improved-test.png`

#### ⚙️ 脚本类 → scripts/
- `start.command`
- `js-launcher.js`
- `npm权限问题解决方案.md`

#### 🗂️ 配置文件 → config/
- 移动相关配置文件

## 🎯 文件命名规范

### 文档文件
- 使用小写连字符命名法
- README.md, CONTRIBUTING.md, CHANGELOG.md
- 技术文档: `tech-[主题].md`
- 部署文档: `deploy-[环境].md`

### 脚本文件
- Shell脚本: `*.sh`
- Node.js脚本: `*.js`
- 配置文件: `*.yml` 或 `*.json`

### 图片文件
- 截图: `screenshot-[功能]-[日期].png`
- 图表: `diagram-[主题].png`

### 测试文件
- 测试脚本: `test-[功能]-[场景].sh`
- 测试报告: `report-[类型]-[日期].md`

## 📊 文件移动映射表

### 立即移动的文件
| 当前位置 | 目标位置 | 说明 |
|---|---|---|
| `COMPONENT_MANAGEMENT.md` | `docs/development/` | 组件管理文档 |
| `DETAILED_AGENT_TASKS.md` | `docs/agents/` | Agent任务文档 |
| `MCPBROWSER_USAGE.md` | `docs/tools/` | MCP浏览器使用说明 |
| `MULTI_AGENT_COORDINATION.md` | `docs/agents/` | 多Agent协调文档 |
| `NEXT_PHASE_TASKS.md` | `docs/project/` | 下一阶段任务 |
| `BROWSERMCP_SETUP.md` | `docs/tools/` | MCP浏览器设置 |
| `test_user_registration.sh` | `test-kimi/scripts/` | 用户注册测试脚本 |
| `comprehensive_mcp_test.js` | `test-kimi/scripts/` | MCP综合测试 |
| `mcp_test_script.js` | `test-kimi/scripts/` | MCP测试脚本 |
| `MCP_TEST_REPORT.md` | `test-kimi/reports/` | MCP测试报告 |

### 图片文件整理
| 当前位置 | 目标位置 | 说明 |
|---|---|---|
| `auth-nav-check.png` | `docs/images/auth/` | 认证导航截图 |
| `debug-homepage.png` | `docs/images/debug/` | 调试首页截图 |
| `openpenpal-improved-test.png` | `docs/images/testing/` | 改进测试截图 |

### 脚本文件整理
| 当前位置 | 目标位置 | 说明 |
|---|---|---|
| `start.command` | `scripts/start.sh` | 启动脚本 |
| `js-launcher.js` | `scripts/launcher.js` | JS启动器 |
| `npm权限问题解决方案.md` | `docs/guides/npm-permissions.md` | NPM权限指南 |

## 🔄 临时文件和缓存

### 需要清理的文件
- `*.log` - 日志文件 → `logs/` 或清理
- `*.tsbuildinfo` - TypeScript缓存 → `.gitignore`
- `.DS_Store` - macOS系统文件 → `.gitignore`
- `node_modules/` - 已存在于各目录中

### 归档文件
- 旧的测试报告 → `archive/reports/`
- 历史文档版本 → `archive/docs/`
- 临时图片文件 → 清理或移动到`temp/`

## 📋 整理执行清单

### Phase 1: 核心文档整理
- [ ] 移动所有Markdown文档到对应docs子目录
- [ ] 移动测试文件到test-kimi对应目录
- [ ] 移动图片文件到docs/images/

### Phase 2: 脚本文件整理
- [ ] 移动启动脚本到scripts/
- [ ] 重命名文件遵循命名规范
- [ ] 更新脚本引用路径

### Phase 3: 配置文件整理
- [ ] 统一配置文件位置
- [ ] 创建环境配置模板
- [ ] 验证所有配置路径

### Phase 4: 清理和验证
- [ ] 清理临时文件
- [ ] 验证所有文件移动成功
- [ ] 更新README文档
- [ ] 运行测试验证环境