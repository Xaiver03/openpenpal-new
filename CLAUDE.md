# OpenPenPal - 校园手写信平台

**核心理念**: Git版本管理，Think before action, SOTA原则，谨慎删除，持续优化用户体验，禁止简化问题和跳过问题，禁止硬编码数据。

## 技术栈
- 前端：Next.js 14, TypeScript, Tailwind CSS, React 18
- 后端：Go (Gin), Python (FastAPI), Java (Spring Boot), PostgreSQL 15  
- 测试：Jest, React Testing Library, Go testing, Python pytest
- 架构：微服务 + WebSocket + JWT认证 + 四级信使系统

## 常用命令
- ./startup/quick-start.sh demo --auto-open: 启动演示模式（推荐）
- ./startup/quick-start.sh development --auto-open: 启动所有服务
- ./startup/check-status.sh: 检查服务状态
- ./startup/stop-all.sh: 停止所有服务
- ./startup/force-cleanup.sh: 强制清理端口
- npm run dev: 启动前端开发服务器（cd frontend）
- go run main.go: 启动后端服务（cd backend）
- npm run type-check: 运行TypeScript类型检查
- ./scripts/test-apis.sh: 运行API测试
- ./test-kimi/run_tests.sh: 运行集成测试

## 编码规范
- 使用严格的TypeScript模式，避免any类型
- Go代码遵循gofmt标准格式化
- 文件命名：snake_case.go, PascalCase.tsx, kebab-case.ts
- API字段命名：后端使用snake_case，前端完全匹配（不转换为camelCase）
- 数据库字段：GORM + snake_case JSON字段
- 导入：优先使用解构导入 import { foo } from 'bar'
- 配置：使用环境变量，禁止硬编码

## 工作流程
- 每次修改后运行type-check验证TypeScript
- Git分支管理：main为生产分支，feature/description为功能分支
- 提交格式：feat/fix/docs: message
- Think before action: 深度分析问题后再编码实现
- SOTA原则：追求最先进的技术实现，注重性能和用户体验
- 谨慎删除：删除代码前充分理解其作用和依赖关系
- PR前确保所有检查通过（类型检查、测试、代码规范）

## 架构设计

### 微服务架构与端口
- Frontend: Next.js 14 + TypeScript (3000)
- Backend: Go + Gin (8080)
- Write: Python/FastAPI (8001)
- Courier: Go (8002)
- Admin: Java/Spring Boot (8003)
- OCR: Python (8004)
- Gateway: Go (8000)

### 核心组件
- 认证：JWT + 四级角色权限（admin/courier/senior_courier/coordinator）
- 数据库：PostgreSQL（必需，不支持SQLite）
- 实时通信：WebSocket
- 存储：本地上传 + QR码生成
- 共享模块：`/shared/go/pkg/`

## 核心业务系统

### 积分活动系统（第三阶段已完成 ✅）
- **智能调度器**: 30秒间隔，5个并发任务，3次重试+指数退避
- **活动类型**: daily/weekly/monthly/seasonal/first_time/cumulative/time_limited  
- **API接口**: 20+个端点在 `/api/v1/credit-activities/` 和 `/admin/credit-activities/`
- **测试命令**: `./backend/scripts/test-credit-activity-scheduler.sh`

### 积分过期系统（Phase 4.1 已完成 ✅）
- **智能过期**: 基于积分类型的分级过期规则，支持12种积分类型
- **批量处理**: 高效的批次过期处理，完整的审计日志和通知系统
- **API接口**: 用户端点 `/api/v1/credits/expiring` 管理端点 `/admin/credits/expiration/*`
- **测试命令**: `./backend/scripts/test-credit-expiration.sh`

### 积分转赠系统（Phase 4.2 已完成 ✅）
- **安全转赠**: 支持直接转赠、礼物转赠、奖励转赠三种类型，带手续费机制
- **智能规则**: 基于用户角色的分级转赠规则，每日/每月限额控制
- **API接口**: 用户端点 `/api/v1/credits/transfer/*` 管理端点 `/admin/credits/transfers/*`
- **状态管理**: 完整的转赠生命周期：待处理→已处理/已拒绝/已取消/已过期

### 四级信使系统（核心架构）

**层级结构**:
1. **L4 城市总代**: 全市控制权，创建L3（区域：BEIJING）
2. **L3 校级信使**: 学校分发，创建L2（区域：BJDX）
3. **L2 片区信使**: 区域管理，创建L1（区域：District）
4. **L1 楼栋信使**: 直接投递（区域：BJDX-A-101）

**核心功能**:
- 智能分配（位置+负载均衡）
- QR扫描工作流（已收集→运输中→已投递）
- 基于表现的晋升机制
- 实时WebSocket追踪
- 游戏化+排行榜

**批量生成权限（L3/L4关键功能）**:
- **L3 校级信使**: 学校级批量生成，管理校园编码（AABBCC格式）
- **L4 城市总代**: 全市批量生成，跨学校操作
- **信号码系统**: 通过`GenerateCodeBatch` API完整批量生成
- **权限矩阵**: 层级继承（L4继承所有L3权限）
- **隐藏UI**: 批量功能存在但UI入口不明显
- **核心API**: POST `/api/signal-codes/batch`, POST `/api/signal-codes/assign`

**关键文件**:
- `services/courier-service/internal/services/hierarchy.go`
- `frontend/src/components/courier/CourierPermissionGuard.tsx`
- `services/courier-service/internal/models/courier.go`
- **批量生成系统（L3/L4）**:
  - `services/courier-service/internal/services/signal_code_service.go`（批量生成API）
  - `services/courier-service/internal/handlers/signal_code_handler.go`（批量端点）
  - `services/courier-service/internal/services/postal_management.go`（L3/L4权限）
  - `services/courier-service/internal/models/signal_code.go`（批量模型）

### 数据库设计
- 主要实体：User, Letter, Courier, Museum
- ORM：GORM + PostgreSQL（必需，不支持SQLite）
- 关系：四级信使层级、权限继承、地理位置映射

## 项目结构
- **Backend**: main.go, internal/{config,handlers,middleware,models,services}/
- **Frontend**: src/{app,components,hooks,lib,stores,types}/
- **Services**: courier-service/, write-service/, admin-service/, ocr-service/
- **Shared**: shared/go/pkg/ (共享Go模块)
- **Scripts**: startup/, scripts/, test-kimi/
- **Docs**: docs/, PRD-NEW/ (产品需求和技术文档)

## 环境设置

### PostgreSQL（必需）
```bash
# 启动数据库
brew services start postgresql  # macOS
sudo systemctl start postgresql # Linux

# 设置数据库
createdb openpenpal
export DATABASE_URL="postgres://$(whoami):password@localhost:5432/openpenpal"
export DB_TYPE="postgres"

# 数据库迁移
cd backend && go run main.go migrate
```
**注意**: macOS使用系统用户名(`whoami`)，Linux可能需要'postgres'

### 测试账户
- admin/admin123 (super_admin)
- alice/secret123 (student) - 已更新密码满足8位字符要求
- courier_level[1-4]/secret123 (L1-L4 courier) - 已更新密码

### 常见问题排查
- 端口冲突: `./startup/force-cleanup.sh`
- 权限问题: 检查middleware配置
- 数据库: 确保PostgreSQL正在运行
- 认证: 前端必须查询数据库，禁止硬编码
- 密码重置: `cd backend && go run cmd/admin/reset_passwords.go -user=username -password=newpass`
- React Hooks错误: 已修复条件hook调用，确保组件渲染一致性

## 开发原则与标准

### SOTA架构原则
1. 微服务清晰分离
2. 共享库在 `/shared/go/pkg/`
3. 四级RBAC权限控制
4. WebSocket实时通信
5. 多层测试策略

### Git版本管理
- `main`: 仅用于生产环境
- 功能分支: `feature/description`
- 提交格式: `feat/fix/docs: message`
- **Think before action**: 深度分析问题后再实施解决方案
- **谨慎删除**: 删除代码前充分理解其作用和依赖关系

### 配置管理
- 后端配置: `internal/config/config.go`
- 前端配置: `src/lib/api.ts`
- 使用环境变量，禁止硬编码

### 开发标准
- Go: gofmt格式化
- TypeScript: ESLint + 严格模式
- 数据库: 一致的GORM，snake_case JSON字段
- API: 统一响应格式
- 文件命名: snake_case.go, PascalCase.tsx, kebab-case.ts
- 字段命名: 后端使用snake_case，前端完全匹配（不转换camelCase）

## 测试与验证

### 信使系统验证
**关键文件**: services/courier-service/, role_compatibility.go, CourierPermissionGuard.tsx

**测试命令**:
```bash
./startup/tests/test-permissions.sh
cd services/courier-service && ./test_apis.sh
curl -X GET "http://localhost:8002/api/v1/courier/hierarchy/level/2"

# 测试L3/L4批量生成权限
curl -X POST "http://localhost:8002/api/signal-codes/batch" \
  -H "Authorization: Bearer $L3_TOKEN" \
  -d '{"batch_no":"B001","school_id":"BJDX","quantity":100}'
  
curl -X POST "http://localhost:8002/api/signal-codes/assign" \
  -H "Authorization: Bearer $L4_TOKEN" \
  -d '{"codes":["PK5F3D","PK5F3E"],"assignee_id":"courier123"}'
```

**层级规则**:
- L4→L3→L2→L1 创建链
- 任务流程: Available→Accepted→Collected→InTransit→Delivered
- 基于区域的权限
- 基于表现的晋升

**端点** (8002): /hierarchy, /tasks, /scan, /leaderboard

## OP Code编码系统（关键）

### 编码格式
**格式**: AABBCC（6位数字）
- AA: 学校（PK=北大, QH=清华, BD=北交大）
- BB: 区域（5F=5号楼, 3D=3食堂, 2G=2号门）
- CC: 位置（3D=303室, 1A=1层A区, 12=12号桌）

**示例**: PK5F3D = 北大5号楼303室

### 核心特性
- 统一6位编码
- 隐私控制（PK5F**隐藏后两位）
- 层级权限管理
- 复用SignalCode基础设施

**数据模型**: SignalCode（重用）, Letter（+OP Code字段）, Courier（+ManagedOPCodePrefix）

### API接口与服务

**服务**: opcode_service.go（Apply/Assign/Search/Validate/Stats/Migrate）
**处理器**: opcode_handler.go（隐私感知端点）

**API端点**:
```bash
# 公开接口
GET /api/v1/opcode/:code
GET /api/v1/opcode/validate

# 受保护接口  
POST /api/v1/opcode/apply
GET /api/v1/opcode/search
GET /api/v1/opcode/stats/:school_code

# 管理员接口
POST /api/v1/opcode/admin/applications/:id/review
```

**隐私级别**: 完全/部分（PK5F**）/公开
**权限控制**: L1受限，L2+前缀访问，管理员完全访问
**迁移映射**: Zone→OPCode（BEIJING→BJ, BJDX→BD）
**验证规则**: 6位大写字母数字，唯一性，层级结构

### OP Code集成状态（✅ 已完成）

**1. 信件服务**: RecipientOPCode/SenderOPCode字段，QR码包含OP数据
**2. 信使任务**: 取件/送达/当前OPCode，前缀权限，地理路由
**3. 博物馆**: OriginOPCode用于来源追踪
**4. QR增强**: JSON格式 + OP Code验证
**架构**: OPCode服务 → 信件/信使/博物馆/通知服务
**数据表**: signal_codes（重用）, letters, courier_tasks, museum_items（都含OP字段）

## FSD条码系统（增强型LetterCode）

### 设计原则
**原则**: 增强现有LetterCode而非创建新模型

**增强的LetterCode模型**:
- 保留原字段（ID, LetterID, Code, QRCodeURL等）
- FSD新增：Status, RecipientCode, EnvelopeID, 扫描追踪
- 状态生命周期：unactivated→bound→in_transit→delivered
- 生命周期方法：IsValidTransition(), IsActive(), CanBeBound()

### FSD服务集成

**请求模型**: BindBarcodeRequest, UpdateBarcodeStatusRequest, EnvelopeWithBarcodeResponse

**服务方法**:
- BindBarcodeToEnvelope() - FSD 6.2
- UpdateBarcodeStatus() - FSD 6.3
- GetBarcodeStatus()
- ValidateBarcodeOperation()

**三方绑定**: LetterCode ↔ Envelope ↔ OP Code
**流程**: 生成→绑定→关联→扫描→投递

### FSD信使集成

**增强模型**: ScanRequest/Response包含FSD字段（条码、OP码、验证）

**任务服务方法**:
- UpdateTaskStatus() - 增强扫描
- validateOPCodePermission() - 基于级别的访问
- getNextAction() - 智能推荐
- calculateEstimatedDelivery() - 时间估算

**OP Code权限**:
- L4: 任何地方
- L3: 同校
- L2: 同校+区域
- L1: 同4位前缀

### FSD端点

**信件条码** (8080):
- POST /api/barcodes (创建条码)
- PATCH /api/barcodes/:id/bind (绑定条码)
- PATCH /api/barcodes/:id/status (更新状态)
- GET /api/barcodes/:id/status (获取状态)
- POST /api/barcodes/:id/validate (验证操作)

**信使扫描** (8002):
- POST /api/v1/courier/scan/:code
- GET /api/v1/courier/scan/history/:id
- POST /api/v1/courier/barcode/:code/validate-access

**生命周期测试**: 绑定→扫描→更新→查询

### FSD优势与状态

**✅ 已实现**:
- 8位条码 + 生命周期管理
- OP Code集成 + 信封绑定
- 四级信使验证
- 实时追踪 + 智能推荐
- 向后兼容

**🔧 优雅**: 增强现有模型，无重复

**集成完成**: 所有系统都符合FSD标准

**测试QR码扫描与OP Code验证**:
```bash
curl -X POST "http://localhost:8080/api/v1/courier/scan" \
  -H "Authorization: Bearer $COURIER_TOKEN" \
  -d '{"qr_data":"...","current_op_code":"PK5F01"}'
```

**集成点**:
- ✅ 信件创建/投递使用OP Code寻址
- ✅ 基于OP Code前缀的信使任务分配
- ✅ 博物馆条目引用OP Code位置
- ✅ QR码包含结构化OP Code数据用于位置追踪
- ✅ 权限系统按OP Code区域验证信使访问
- ✅ 按OP Code区域进行地理分析和报告

### OP Code实现详情

**模型**: OPCodeApplication, OPCodeRequest, OPCodeAssignRequest, OPCodeSearchRequest, OPCodeStats
**类型**: dormitory/shop/box/club, pending/approved/rejected
**工具**: Generate/Parse/Validate/FormatOPCode
**服务**: Apply/Assign/Get/Search/Stats/ValidateAccess/Migrate
**处理器**: 用户端点 + 管理员审核

**状态**: ⚠️ 代码完成但数据库迁移缺失 - 模型、服务、处理器、路由、验证已实现，但OP Code模型未包含在数据库迁移中

**🔴 关键问题**: OP Code模型未包含在 `backend/internal/config/database.go` 的 `getAllModels()` 函数中，导致数据库表未创建

**测试**: 使用提供的curl命令和适当的认证令牌（需先修复数据库迁移）

## SOTA增强（最先进技术）

### React性能优化工具
- **位置**: `frontend/src/lib/utils/react-optimizer.ts`
- **特性**: 智能备忘，虚拟滚动，性能监控，懒加载
- **用法**: `useDebouncedValue`, `useThrottledCallback`, `useOptimizedState`, `smartMemo`

### 增强API客户端
- **位置**: `frontend/src/lib/utils/enhanced-api-client.ts`  
- **特性**: 断路器模式，请求去重，智能缓存
- **优势**: 提高可靠性，减少冗余请求，更好的用户体验

### 错误处理系统
- **增强错误边界**: `frontend/src/components/error-boundary/enhanced-error-boundary.tsx`
- **性能监控器**: `frontend/src/lib/utils/performance-monitor.ts`
- **缓存管理器**: `frontend/src/lib/utils/cache-manager.ts`

### 认证系统增强
- **增强提供者**: `frontend/src/app/providers/auth-provider-enhanced.tsx`
- **调试工具**: 仅开发环境的认证调试小部件
- **安全性**: CSRF保护，令牌轮换，安全存储

## 近期修复记录

### React Hooks错误解决
- **问题**: "渲染的hooks比前一次渲染多"
- **修复**: 一致的hook执行，正确的useCallback使用，清理处理
- **位置**: `auth-provider-enhanced.tsx:138-152`

### TypeScript一致性
- **问题**: 字段命名不匹配（camelCase ↔ snake_case）
- **修复**: 更新所有前端类型以完全匹配后端JSON
- **影响**: 用户类型，信件类型，API响应，状态管理

### 数据库连接
- **问题**: 连接字符串解析错误
- **修复**: 使用`config.DatabaseName`而非`config.DatabaseURL`
- **位置**: `backend/internal/config/database.go:45`

---

## 结语

**OpenPenPal**是一个现代化的校园手写信平台，采用微服务架构，集成了先进的四级信使系统、OP Code编码、条码追踪等创新功能。本文档旨在为开发者提供完整的项目理解和开发指导。

## 技术债务状态（2025-08-16 最新验证）

### ✅ 已完成的数据库迁移 (2025-08-15)
- **积分系统数据库**: 全部24个积分系统表已成功创建和迁移 ✅
- **迁移脚本**: 创建了PostgreSQL兼容的迁移脚本 `backend/scripts/migrate-database.sh`
- **表覆盖**: Phase 1-4 所有积分功能的数据库表已就绪
- **验证命令**: `./backend/scripts/migrate-database.sh` 显示 24/24 表存在

### 🔴 新发现的关键问题 (2025-08-16)
- **OP Code数据库缺失**: OP Code模型完整但未包含在数据库迁移中，需添加到 `getAllModels()` 函数
- **硬编码JWT令牌**: 10个测试文件仍包含硬编码JWT令牌，存在安全风险  
- **技术债务**: 171个TODO/FIXME注释分布在80个文件中，需要逐步清理

### 🔴 剩余高优先级问题  
- **禁用服务**: 12个 `.disabled` 服务文件需要重新启用和测试（非15个）
- **路径错误**: 更正脚本路径引用从 `/scripts/` 到 `/backend/scripts/`

### ✅ 已修复的安全问题  
- `.broken` 文件已全部修复（0个残留）



```