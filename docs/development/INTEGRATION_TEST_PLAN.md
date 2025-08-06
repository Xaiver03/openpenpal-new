# OpenPenPal 集成测试计划

## 测试概述

本测试计划旨在全面验证 OpenPenPal 系统的功能完整性，重点关注：
- 前端按钮与后端 API 的交互
- 前后端与数据库的实际交互
- 各测试账号的登录和权限功能
- 用户体验的完整性

## 测试环境

- **前端**: http://localhost:3000
- **后端**: http://localhost:8080
- **数据库**: SQLite (openpenpal.db)
- **测试浏览器**: Chrome/Firefox 最新版本

## 测试账号

### 1. 管理员账号
- 用户名: `admin`
- 密码: `admin123`
- 角色: admin
- 权限: 系统管理、用户管理、数据统计

### 2. 四级信使账号
| 级别 | 用户名 | 密码 | 权限范围 |
|------|--------|------|----------|
| Level 4 (城市总代) | `courier_level4_city` | `courier123` | 城市级管理、创建L3信使 |
| Level 3 (校级信使) | `courier_level3_school` | `courier123` | 学校级管理、创建L2信使 |
| Level 2 (片区信使) | `courier_level2_zone` | `courier123` | 片区管理、创建L1信使 |
| Level 1 (楼栋信使) | `courier_level1_building` | `courier123` | 信件收发、扫码确认 |

### 3. 普通用户账号
- 用户名: `test_user`
- 密码: `user123`
- 角色: user
- 权限: 写信、查看信件、购买信封

## 测试用例

### 1. 用户认证测试

#### TC-AUTH-001: 登录功能测试
**测试步骤:**
1. 访问 http://localhost:3000/login
2. 输入各测试账号的用户名和密码
3. 点击"登录"按钮
4. 验证跳转到对应的首页

**预期结果:**
- 管理员跳转到管理后台
- 信使跳转到信使工作台
- 普通用户跳转到用户首页

**数据库验证:**
```sql
SELECT * FROM users WHERE username = 'admin';
SELECT * FROM user_sessions WHERE user_id = ?;
```

#### TC-AUTH-002: 注册功能测试
**测试步骤:**
1. 访问 http://localhost:3000/register
2. 填写注册表单
3. 选择学校代码
4. 提交注册

**预期结果:**
- 新用户记录创建成功
- 自动登录并跳转

### 2. 写信功能测试

#### TC-LETTER-001: 创建信件
**测试步骤:**
1. 登录普通用户账号
2. 点击"写信"按钮
3. 填写信件内容
4. 选择信件样式
5. 点击"保存草稿"
6. 点击"生成编号"

**API 交互验证:**
- POST /api/v1/letters
- POST /api/v1/letters/{id}/generate-code

**数据库验证:**
```sql
SELECT * FROM letters WHERE user_id = ?;
SELECT * FROM letter_codes WHERE letter_id = ?;
```

#### TC-LETTER-002: 绑定信封
**测试步骤:**
1. 在已生成编号的信件页面
2. 点击"绑定信封"按钮
3. 选择可用信封
4. 确认绑定

**API 交互验证:**
- POST /api/v1/letters/{id}/bind-envelope

### 3. 信使功能测试

#### TC-COURIER-001: 四级信使权限测试
**测试步骤:**
1. 分别登录四个级别的信使账号
2. 验证各自的功能菜单
3. 测试创建下级信使功能

**权限验证矩阵:**
| 功能 | L4 | L3 | L2 | L1 |
|------|----|----|----|----|
| 创建L3信使 | ✓ | ✗ | ✗ | ✗ |
| 创建L2信使 | ✗ | ✓ | ✗ | ✗ |
| 创建L1信使 | ✗ | ✗ | ✓ | ✗ |
| 扫码收件 | ✓ | ✓ | ✓ | ✓ |
| 查看统计 | ✓ | ✓ | ✓ | ✗ |

#### TC-COURIER-002: 信件扫码流转
**测试步骤:**
1. L1信使扫码收件
2. 更新状态为"已收取"
3. L2信使扫码接收
4. 更新状态为"运输中"
5. L1信使扫码派送
6. 更新状态为"已送达"

**WebSocket 验证:**
- 实时状态更新推送
- 通知消息推送

### 4. 编码系统测试

#### TC-CODE-001: 信件编码生成
**测试步骤:**
1. 创建新信件
2. 点击"生成编号"
3. 验证编号格式 (OP + 时间戳 + 随机码)
4. 验证二维码生成

**文件系统验证:**
```bash
ls -la backend/uploads/qrcodes/
```

#### TC-CODE-002: Postcode 地址选择
**测试步骤:**
1. 在写信页面选择收件地址
2. 测试分级选择模式
3. 测试搜索模式
4. 验证6位编码生成

**API 交互验证:**
- GET /api/postcode/schools
- GET /api/postcode/search

### 5. 信封系统测试

#### TC-ENVELOPE-001: 信封设计投票
**测试步骤:**
1. 浏览信封设计列表
2. 点击投票按钮
3. 验证投票数更新

#### TC-ENVELOPE-002: 信封订购
**测试步骤:**
1. 选择信封设计
2. 设置购买数量
3. 提交订单
4. 验证订单状态

### 6. 管理后台测试

#### TC-ADMIN-001: 统计仪表板
**测试步骤:**
1. 登录管理员账号
2. 查看统计数据
3. 验证数据准确性

**数据验证查询:**
```sql
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM letters;
SELECT status, COUNT(*) FROM letters GROUP BY status;
```

#### TC-ADMIN-002: 用户管理
**测试步骤:**
1. 查看用户列表
2. 搜索用户
3. 修改用户角色
4. 禁用/启用用户

### 7. 用户体验完整性测试

#### TC-UX-001: 普通用户完整流程
**测试流程:**
1. 注册新账号
2. 完善个人信息
3. 写第一封信
4. 生成编号
5. 选择并绑定信封
6. 提交投递
7. 追踪信件状态
8. 收到回信通知
9. 查看信件历史

#### TC-UX-002: 信使工作流程
**测试流程:**
1. 登录信使账号
2. 查看待处理任务
3. 扫码收取信件
4. 批量处理信件
5. 更新投递状态
6. 查看绩效统计
7. 申请晋升

#### TC-UX-003: 响应式测试
**测试设备:**
- 桌面端 (1920x1080)
- 平板端 (768x1024)
- 移动端 (375x667)

### 8. 性能测试

#### TC-PERF-001: 并发登录测试
- 同时登录10个用户
- 验证系统响应时间

#### TC-PERF-002: 数据加载测试
- 加载1000条信件记录
- 验证分页功能
- 测试搜索性能

## 测试执行记录

### 测试环境准备
```bash
# 1. 启动所有服务
cd /Users/rocalight/同步空间/opplc/openpenpal
./startup/quick-start.sh development --auto-open

# 2. 检查服务状态
./startup/check-status.sh

# 3. 初始化测试数据
curl -X POST http://localhost:8080/api/v1/admin/inject-seed-data \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### 测试结果记录表

| 测试用例ID | 测试日期 | 执行人 | 测试结果 | 问题描述 | 截图 |
|-----------|---------|--------|---------|---------|------|
| TC-AUTH-001 | | | | | |
| TC-AUTH-002 | | | | | |
| TC-LETTER-001 | | | | | |
| TC-LETTER-002 | | | | | |
| TC-COURIER-001 | | | | | |
| TC-COURIER-002 | | | | | |
| TC-CODE-001 | | | | | |
| TC-CODE-002 | | | | | |
| TC-ENVELOPE-001 | | | | | |
| TC-ENVELOPE-002 | | | | | |
| TC-ADMIN-001 | | | | | |
| TC-ADMIN-002 | | | | | |
| TC-UX-001 | | | | | |
| TC-UX-002 | | | | | |
| TC-UX-003 | | | | | |
| TC-PERF-001 | | | | | |
| TC-PERF-002 | | | | | |

## 测试工具脚本

### 1. 自动化登录测试
```bash
# test-login.sh
#!/bin/bash
BASE_URL="http://localhost:8080"

test_login() {
    local username=$1
    local password=$2
    
    response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    echo "Testing login for $username: $response"
}

# 测试所有账号
test_login "admin" "admin123"
test_login "courier_level4_city" "courier123"
test_login "courier_level3_school" "courier123"
test_login "courier_level2_zone" "courier123"
test_login "courier_level1_building" "courier123"
test_login "test_user" "user123"
```

### 2. API 交互测试
```javascript
// test-api-interactions.js
const testCases = [
    {
        name: "Create Letter",
        method: "POST",
        url: "/api/v1/letters",
        body: {
            title: "测试信件",
            content: "这是一封测试信件",
            style: "classic"
        }
    },
    {
        name: "Generate Code",
        method: "POST",
        url: "/api/v1/letters/{letterID}/generate-code"
    },
    {
        name: "Update Status",
        method: "PUT",
        url: "/api/v1/letters/scan/{code}",
        body: {
            status: "collected",
            location: "北京大学"
        }
    }
];
```

### 3. 数据库验证脚本
```sql
-- verify-database.sql
-- 验证用户数据
SELECT role, COUNT(*) as count FROM users GROUP BY role;

-- 验证信件状态
SELECT status, COUNT(*) as count FROM letters GROUP BY status;

-- 验证信使层级
SELECT level, COUNT(*) as count FROM couriers GROUP BY level;

-- 验证信封使用情况
SELECT status, COUNT(*) as count FROM envelopes GROUP BY status;

-- 验证编码生成
SELECT COUNT(*) as total_codes FROM letter_codes;
SELECT COUNT(*) as total_postcodes FROM postcode_rooms;
```

## 问题跟踪

### 已知问题
1. 
2. 
3. 

### 改进建议
1. 
2. 
3. 

## 测试总结

**测试覆盖率:**
- 功能测试: __%
- 集成测试: __%
- 用户体验: __%

**测试结论:**
- [ ] 所有测试账号可正常登录
- [ ] 前后端交互正常
- [ ] 数据库操作正确
- [ ] 权限控制有效
- [ ] 用户体验流畅

**发布建议:**
- [ ] 可以发布
- [ ] 需要修复关键问题后发布
- [ ] 需要重大改进

---

测试执行时间: ____年__月__日 __:__ - __:__
测试执行人: _____________
审核人: _____________