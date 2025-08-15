# OpenPenPal 测试账号指南

> **最后更新**: 2025-08-14  
> **状态**: 当前数据库中的有效测试账号

## 📋 测试账号列表

### 🔑 默认密码说明
根据系统配置，所有测试账号的默认密码为：
- 管理员账号 (admin): `admin123`
- 其他测试账号: `secret123`（至少8位字符要求）

### 👤 用户账号

| 用户名 | 密码 | 邮箱 | 角色 | 说明 |
|--------|------|------|------|------|
| alice | secret123 | alice@example.com | user | 普通学生用户，用于测试写信、收信等基础功能 |
| bob | secret123 | bob@example.com | user | 普通学生用户，可与alice互发信件测试 |

### 🚴 信使账号（四级体系）

| 用户名 | 密码 | 邮箱 | 角色 | 权限说明 |
|--------|------|------|------|----------|
| courier_level1 | secret123 | courier1@openpenpal.com | courier_level1 | L1楼栋信使 - 仅可配送指定楼栋区域 |
| courier_level2 | secret123 | courier2@openpenpal.com | courier_level2 | L2片区信使 - 管理多个楼栋，可创建L1信使 |
| courier_level3 | secret123 | courier3@openpenpal.com | courier_level3 | L3校级信使 - 管理整个学校，可创建L2信使，**批量生成条码权限** |
| courier_level4 | secret123 | courier4@openpenpal.com | courier_level4 | L4城市总代 - 跨校管理，可创建L3信使，**全局批量生成权限** |

### 👨‍💼 管理员账号

| 用户名 | 密码 | 邮箱 | 角色 | 权限说明 |
|--------|------|------|------|----------|
| admin | admin123 | admin@openpenpal.com | super_admin | 超级管理员 - 系统所有权限 |

## 🧪 测试场景指南

### 1. 基础功能测试

#### 📝 写信测试
1. 使用 `alice` 账号登录
2. 进入写信页面 (`/write`)
3. 选择信纸样式（4种可选）
4. 编写信件内容
5. 选择发送模式：
   - 定向发送：需要输入收件人OP Code
   - 匿名漂流：随机匹配收件人
   - 公开发布：发布到信件博物馆

#### 📮 收信测试
1. 使用 `bob` 账号登录
2. 查看收件箱
3. 如果alice发送了定向信件，bob应该能收到

### 2. 信使系统测试

#### 🏃 任务派送流程
1. 使用 `courier_level1` 登录
2. 查看可接取的配送任务
3. 接取任务后扫描二维码
4. 更新配送状态：已取件 → 配送中 → 已送达

#### 📊 层级管理测试
1. 使用 `courier_level3` 登录（校级信使）
2. 可以查看和管理下级信使
3. **批量生成条码功能**（L3/L4专属）
4. 分配任务给下级信使

### 3. AI功能测试

#### 🤖 AI写作助手
1. 在写信页面点击"AI助手"
2. 选择灵感类型
3. AI会提供写作建议

#### 💌 云中锦书
1. 选择AI笔友人设
2. 开启自动回信功能
3. AI会定期生成回信

### 4. 高级功能测试

#### 🏛️ 信件博物馆
1. 访问 `/museum`
2. 查看公开信件展览
3. 参与投票和评论

#### 📱 OP Code系统
- 格式：6位编码 (XXYYZI)
- 示例：PK5F3D = 北大5号楼303室
- 用于精确定位收件地址

## 🐛 常见问题

### 登录失败
- 确认密码至少8位字符
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

### 重置密码
如果忘记密码，可以使用管理工具重置：
```bash
cd backend && go run cmd/admin/reset_passwords.go -user=alice -password=newpassword
```

## 🔗 相关文档

- [用户手册](../getting-started/README.md)
- [信使系统详解](../getting-started/courier-levels.md)
- [API测试指南](../api/README.md)

---

**提示**: 
- 测试环境数据会定期清理，请勿存储重要信息
- 生产环境请使用不同的账号和密码
- 如需添加新测试账号，请联系系统管理员