# OpenPenPal 测试账号文档

## 标准化测试账号

遵循SOTA原则，所有测试账号使用统一的密码哈希算法和标准化配置。

### 管理员账号

| 用户名 | 密码 | 角色 | 邮箱 | 昵称 | 学校 |
|--------|------|------|------|------|------|
| `admin` | `admin123` | super_admin | admin@openpenpal.com | 系统管理员 | SYSTEM |
| `school_admin` | `secret` | school_admin | school@openpenpal.com | 学校管理员 | PKU001 |

### 四级信使系统测试账号

| 用户名 | 密码 | 角色 | 邮箱 | 昵称 | 学校 | 信使等级 |
|--------|------|------|------|------|------|----------|
| `courier_level1` | `secret` | courier_level1 | courier1@openpenpal.com | 一级信使 | PKU001 | 1 |
| `courier_level2` | `secret` | courier_level2 | courier2@openpenpal.com | 二级信使 | PKU001 | 2 |
| `courier_level3` | `secret` | courier_level3 | courier3@openpenpal.com | 三级信使 | PKU001 | 3 |
| `courier_level4` | `secret` | courier_level4 | courier4@openpenpal.com | 四级信使 | PKU001 | 4 |

### 通用角色兼容性测试账号

| 用户名 | 密码 | 角色 | 邮箱 | 昵称 | 学校 | 对应等级 |
|--------|------|------|------|------|------|----------|
| `courier` | `secret` | courier | courier@openpenpal.com | 通用信使 | PKU001 | 1 |
| `senior_courier` | `secret` | senior_courier | senior@openpenpal.com | 高级信使 | PKU001 | 2 |
| `coordinator` | `secret` | courier_coordinator | coordinator@openpenpal.com | 信使协调员 | PKU001 | 3 |

### 普通用户账号

| 用户名 | 密码 | 角色 | 邮箱 | 昵称 | 学校 |
|--------|------|------|------|------|------|
| `alice` | `secret` | user | alice@pku.edu.cn | Alice | PKU001 |
| `bob` | `secret` | user | bob@tsinghua.edu.cn | Bob | THU001 |

## 权限等级说明

### 信使等级权限

- **Level 1 (一级信使)**: 楼栋/班级级别，负责基础投递
  - 权限：写信、读信、管理档案、投递信件、扫码、查看任务
  
- **Level 2 (二级信使)**: 区域/年级级别，可以分配任务
  - 权限：Level 1 + 查看报告
  
- **Level 3 (三级信使)**: 学校级别，管理学校信使团队
  - 权限：Level 2 + 管理信使、分配任务
  
- **Level 4 (四级信使)**: 城市级别，协调跨校投递
  - 权限：Level 3 + 管理学校

### 管理员权限

- **school_admin**: 学校管理员，管理本校用户和信使
- **platform_admin**: 平台管理员，跨校管理权限
- **super_admin**: 超级管理员，全平台最高权限

## API测试示例

### 登录测试
```bash
# 管理员登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 四级信使登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "courier_level4", "password": "secret"}'

# 普通用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret"}'
```

### 权限测试
```bash
# 获取当前用户信息（需要认证token）
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 测试信使功能（需要信使权限）
curl -X GET http://localhost:8080/api/v1/courier/tasks \
  -H "Authorization: Bearer COURIER_JWT_TOKEN"

# 测试管理员功能（需要管理员权限）
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

## 开发环境配置

测试账号在开发环境下自动创建。如需重新初始化：

1. 删除数据库文件：`rm openpenpal.db`
2. 重启应用：`go run main.go`
3. 测试账号将自动重新创建

## 密码哈希验证

系统使用bcrypt算法进行密码哈希，成本因子为10。所有测试账号的密码哈希都经过验证确保正确性。

## 前后端角色兼容性

系统支持两套角色名称：
- **前端兼容**: `courier`, `senior_courier`, `courier_coordinator`
- **后端标准**: `courier_level1`, `courier_level2`, `courier_level3`, `courier_level4`

角色映射自动处理，确保前后端无缝兼容。