# OpenPenPal 测试账号指南

> **最后更新**: 2025-08-15  
> **状态**: 基于实际数据库验证的有效测试账号  
> **验证状态**: ✅ 所有账号密码已验证可正常登录

## 📋 测试账号列表

### 🔑 默认密码说明
所有测试账号密码已标准化，符合安全要求（8位以上，包含大小写字母、数字、符号中至少两种）：
- **管理员账号 (admin)**: `Admin123!`
- **其他所有测试账号**: `Secret123!`

### 👤 用户账号

| 用户名 | 密码 | 邮箱 | 角色 | 昵称 | 学校代码 | 状态 |
|--------|------|------|------|------|----------|------|
| alice | Secret123! | alice@example.com | user | Alice | BJDX01 | ✅ 已验证 |
| bob | Secret123! | bob@example.com | user | Bob | BJDX01 | ✅ 已验证 |
| api_test_user_fixed | Secret123! | apitestfixed@example.com | user | API测试用户修复版 | TEST01 | ✅ 可用 |
| test_db_connection | Secret123! | test@example.com | user | 数据库测试用户 | - | ✅ 可用 |

### 🚴 信使账号（四级体系）

| 用户名 | 密码 | 邮箱 | 角色 | 昵称 | 学校代码 | 权限说明 | 状态 |
|--------|------|------|------|------|----------|----------|------|
| courier_level1 | Secret123! | courier1@openpenpal.com | courier_level1 | Level 1 Courier | BJDX01 | L1楼栋信使 - 基础配送权限 | ✅ 已验证 |
| courier_level2 | Secret123! | courier2@openpenpal.com | courier_level2 | Level 2 Courier | BJDX01 | L2片区信使 - 可管理L1信使 | ✅ 已验证 |
| courier_level3 | Secret123! | courier3@openpenpal.com | courier_level3 | Level 3 Courier | BJDX01 | L3校级信使 - **批量生成条码权限** | ✅ 已验证 |
| courier_level4 | Secret123! | courier4@openpenpal.com | courier_level4 | Level 4 Courier | SYSTEM | L4城市总代 - **全局批量生成权限** | ✅ 已验证 |

### 👨‍💼 管理员账号

| 用户名 | 密码 | 邮箱 | 角色 | 昵称 | 学校代码 | 权限说明 | 状态 |
|--------|------|------|------|------|----------|----------|------|
| admin | Admin123! | admin@openpenpal.com | super_admin | 系统管理员 | SYSTEM | 超级管理员 - 系统所有权限 | ✅ 已验证 |

## 🧪 测试场景指南

### 1. 基础功能测试

#### 📝 写信测试
```bash
# 登录alice账号进行写信测试
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"Secret123!"}'
```

1. 使用 `alice` 账号登录 ✅ 验证可用
2. 进入写信页面 (`/write`)
3. 选择信纸样式（4种可选）
4. 编写信件内容
5. 选择发送模式：
   - 定向发送：需要输入收件人OP Code
   - 匿名漂流：随机匹配收件人
   - 公开发布：发布到信件博物馆

#### 📮 收信测试
```bash
# 登录bob账号进行收信测试
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"Secret123!"}'
```

1. 使用 `bob` 账号登录 ✅ 验证可用
2. 查看收件箱
3. 如果alice发送了定向信件，bob应该能收到

### 2. 信使系统测试

#### 🏃 任务派送流程
```bash
# 登录L1信使账号
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"courier_level1","password":"Secret123!"}'
```

1. 使用 `courier_level1` 登录 ✅ 验证可用
2. 查看可接取的配送任务
3. 接取任务后扫描二维码
4. 更新配送状态：已取件 → 配送中 → 已送达

#### 📊 层级管理测试
```bash
# 登录L3校级信使（具有批量权限）
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"courier_level3","password":"Secret123!"}'
```

1. 使用 `courier_level3` 登录 ✅ 验证可用
2. 可以查看和管理下级信使
3. **批量生成条码功能**（L3/L4专属）
4. 分配任务给下级信使

#### 🌟 高级管理测试
```bash
# 登录L4城市总代（最高权限）
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"courier_level4","password":"Secret123!"}'
```

1. 使用 `courier_level4` 登录 ✅ 验证可用
2. 跨校管理权限
3. **全局批量生成权限**
4. 创建和管理L3信使

### 3. 管理员功能测试

#### 🔧 系统管理
```bash
# 登录超级管理员
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin123!"}'
```

1. 使用 `admin` 账号登录 ✅ 验证可用
2. 访问管理后台
3. 用户管理、系统配置
4. 敏感词管理（L4信使和管理员专属）
5. 安全监控和统计

### 4. AI功能测试

#### 🤖 AI写作助手
1. 在写信页面点击"AI助手"
2. 选择灵感类型
3. AI会提供写作建议

#### 💌 云中锦书
1. 选择AI笔友人设
2. 开启自动回信功能
3. AI会定期生成回信

### 5. 高级功能测试

#### 🏛️ 信件博物馆
1. 访问 `/museum`
2. 查看公开信件展览
3. 参与投票和评论

#### 📱 OP Code系统
- **格式**: 6位编码 (AABBCC)
- **示例**: PK5F3D = 北大5号楼303室
- **用途**: 精确定位收件地址
- **隐私**: 支持部分隐藏 (PK5F**)

## 🔒 安全功能测试

### XSS防护测试
```bash
# 测试XSS攻击防护
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"<script>alert(1)</script>","password":"test"}'
# 应该返回400错误，表示XSS检测生效
```

### 敏感词管理测试（需要L4或管理员权限）
```bash
# 获取admin令牌后测试敏感词管理
curl -X GET "http://localhost:8080/api/v1/admin/sensitive-words" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## 🐛 常见问题

### 登录失败
- ✅ **已修复**: 所有账号密码已同步为标准格式
- 确认密码正确：admin用`Admin123!`，其他用`Secret123!`
- 检查数据库连接是否正常
- 查看后端日志：`tail -f logs/go-backend.log`

### 信使任务不显示
- 确认信使账号级别正确
- 检查是否有待配送的信件
- L1只能看到自己楼栋的任务

### AI功能异常
- 检查AI服务配置
- 确认API密钥设置正确
- 查看环境变量：`AI_PROVIDER`

## 📝 测试数据准备

### 创建测试信件
```sql
-- 创建一封从alice到bob的测试信件
INSERT INTO letters (sender_id, recipient_op_code, title, content, style, status) 
VALUES ('test-user-alice', 'PK5F3D', '测试信件', '这是一封测试信件内容', 'classic', 'pending');
```

### 重置密码（如需要）
```bash
# 使用新的标准化密码更新脚本（推荐）
./scripts/update-test-passwords.sh

# 或者手动重置单个用户密码
cd backend && go run cmd/admin/reset_passwords.go -user=alice -password=Secret123!
```

### 快速验证所有账号
```bash
# 验证脚本 - 测试所有账号登录状态
cd backend/scripts && ./verify_test_accounts.sh
```

## 📊 数据库实际状态

**当前数据库用户总数**: 9个  
**验证状态**: 所有核心账号密码已标准化并符合安全要求  
**最后验证时间**: 2025-08-19 10:30

## 🔗 相关文档

- [用户手册](../getting-started/README.md)
- [信使系统详解](../getting-started/courier-levels.md)
- [API测试指南](../api/README.md)
- [安全系统文档](../security/)

---

**⚠️ 重要提示**: 
- ✅ 测试环境所有账号密码已标准化并符合安全要求（8位以上，包含大小写字母、数字、符号）
- 🔒 生产环境请使用不同的账号和密码
- 📝 如需添加新测试账号，请联系系统管理员
- 🛡️ 安全功能测试已包含XSS防护和敏感词管理验证
- 🔑 密码标准：最少8位，必须包含大写字母、小写字母、数字、符号中至少两种
- 🔄 数据库更新：运行 `./scripts/update-test-passwords.sh` 以更新数据库中的实际密码哈希