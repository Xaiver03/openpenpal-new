# 数据库测试用户系统

## 概述

为了提高测试稳定性，我们已将测试用户数据从内存存储迁移到数据库持久化存储。系统现在支持两种模式：

1. **数据库模式**（推荐）- 使用PostgreSQL存储测试用户
2. **内存模式**（fallback）- 当数据库不可用时的兜底方案

## 系统架构

### 数据库优先策略

```
登录请求 → CSRF验证 → 网关认证（如果可用）
                    ↓（网关不可用）
          数据库认证 → 内存认证（fallback）
```

### 自动降级机制

系统启动时会按以下顺序检查：
1. 检查数据库连接是否可用
2. 检查数据库中是否存在测试用户
3. 如果数据库可用且有测试用户，使用数据库模式
4. 否则自动降级到内存模式

## 数据库初始化

### 1. 创建测试用户

```bash
# 运行数据库初始化脚本
node scripts/init-database-test-users.js
```

### 2. 数据库表结构

脚本会自动创建 `test_users` 表：

```sql
CREATE TABLE test_users (
  id VARCHAR(50) PRIMARY KEY,
  username VARCHAR(100) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  real_name VARCHAR(100) NOT NULL,
  password_hash TEXT NOT NULL,
  role VARCHAR(50) NOT NULL,
  permissions JSONB DEFAULT '[]'::jsonb,
  school_code VARCHAR(20),
  school_name VARCHAR(100),
  status VARCHAR(20) DEFAULT 'active',
  courier_level INTEGER,
  courier_info JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 测试账户

### 基础管理员账户

| 用户名 | 密码 | 角色 | 描述 |
|--------|------|------|------|
| admin | admin123 | super_admin | 系统管理员 |
| courier_building | courier123 | courier | 建筑楼信使 |
| senior_courier | senior123 | senior_courier | 高级信使 |
| coordinator | coord123 | courier_coordinator | 信使协调员 |

### 层级信使账户

| 用户名 | 密码 | 等级 | 描述 |
|--------|------|------|------|
| courier_level4_city | city123 | 4 | 四级信使（城市总代） |
| courier_level3_school | school123 | 3 | 三级信使（校级） |
| courier_level2_zone | zone123 | 2 | 二级信使（片区/年级） |
| courier_level1_building | building123 | 1 | 一级信使（楼栋/班级） |

## 环境配置

### 数据库连接

在 `.env.local` 中配置：

```env
# PostgreSQL Database Settings
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=openpenpal
DATABASE_USER=postgres
DATABASE_PASSWORD=OpenPenPal_Secure_DB_P@ssw0rd_2025

# Test Data Configuration
ENABLE_TEST_DATA=true
```

### 测试账户密码

支持通过环境变量自定义密码：

```env
TEST_ACCOUNT_ADMIN_PASSWORD=admin123
TEST_ACCOUNT_COURIER_LEVEL2_ZONE_PASSWORD=zone123
TEST_ACCOUNT_COURIER_LEVEL3_SCHOOL_PASSWORD=school123
# ... 其他账户密码
```

## 使用方式

### 1. 开发环境

正常启动应用，系统会自动检测数据库可用性：

```bash
npm run dev
```

### 2. 查看初始化状态

启动时查看日志：

```bash
# 数据库模式
✅ 使用数据库中的测试用户

# 内存模式（fallback）
⚠️  数据库不可用，使用内存存储
📄 内存初始化完成：8 个账户
```

### 3. 登录测试

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_TOKEN" \
  -d '{"username":"courier_level2_zone","password":"zone123"}'
```

## 优势对比

### 数据库模式 vs 内存模式

| 特性 | 数据库模式 | 内存模式 |
|------|------------|----------|
| 数据持久性 | ✅ 持久化存储 | ❌ 重启丢失 |
| 启动速度 | ✅ 快速启动 | ❌ 需重新生成 |
| 测试稳定性 | ✅ 高稳定性 | ❌ 数据不一致 |
| 真实环境测试 | ✅ 真实DB查询 | ❌ 内存模拟 |
| 多实例一致性 | ✅ 数据统一 | ❌ 各自独立 |

## 监控和调试

### 1. 检查系统状态

```bash
# 查看用户统计
curl http://localhost:3000/api/auth/debug-users
```

### 2. 查看日志

关注以下日志信息：
- `✅ 使用数据库中的测试用户` - 数据库模式启动成功
- `⚠️  数据库不可用，使用内存存储` - 降级到内存模式
- `数据库认证失败，降级到内存认证` - 运行时降级

### 3. 数据库查询

```sql
-- 查看所有测试用户
SELECT username, real_name, role, courier_level, status 
FROM test_users 
ORDER BY courier_level DESC NULLS LAST;

-- 查看用户统计
SELECT 
  COUNT(*) as total_users,
  COUNT(*) FILTER (WHERE role = 'courier') as courier_users,
  COUNT(*) FILTER (WHERE status = 'active') as active_users
FROM test_users;
```

## 故障排除

### 1. 数据库连接失败

检查：
- PostgreSQL服务是否运行
- 数据库连接配置是否正确
- 防火墙设置

### 2. 用户不存在

重新运行初始化脚本：
```bash
node scripts/init-database-test-users.js
```

### 3. 密码错误

检查环境变量配置是否正确：
```bash
echo $TEST_ACCOUNT_COURIER_LEVEL2_ZONE_PASSWORD
```

## 生产环境注意事项

⚠️ **安全警告**：
- 生产环境必须禁用测试数据：`ENABLE_TEST_DATA=false`
- 测试账户仅适用于开发和测试环境
- 生产环境应使用独立的用户管理系统

## API接口

### 数据库用户服务

```typescript
import { DatabaseUserService } from '@/lib/services/database-user-service'

// 用户认证
const user = await DatabaseUserService.authenticate(username, password)

// 获取用户信息
const user = await DatabaseUserService.getUserByUsername(username)

// 检查数据库中是否有测试用户
const hasUsers = await DatabaseUserService.hasTestUsers()
```

## 总结

通过数据库持久化测试用户数据，我们实现了：

1. **更高的测试稳定性** - 数据不会因重启而丢失
2. **更快的启动速度** - 避免每次启动时重新生成密码哈希
3. **更真实的测试环境** - 使用真实的数据库查询路径
4. **更好的开发体验** - 支持自动降级，确保系统可用性

系统现在具备了生产级别的稳定性，同时保持了开发环境的灵活性。