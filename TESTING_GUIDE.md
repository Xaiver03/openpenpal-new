# OpenPenPal 测试指南 🧪

**Complete Testing Guide with All Test Accounts**  
*基于ULTRATHINK数据库分析和完整数据注入 - 2025-08-16*

---

## 🚀 快速启动测试环境

### 1. 启动系统
```bash
# 在项目根目录执行
./startup/quick-start.sh demo --auto-open

# 或者手动启动服务
cd backend && ./main          # 后端 (8080)
cd frontend && npm run dev    # 前端 (3000)
```

### 2. 系统状态检查
```bash
./startup/check-status.sh     # 检查所有服务状态
./startup/stop-all.sh         # 停止所有服务（如需重启）
```

---

## 👥 测试账号总览

### 🔐 **管理员账户**

| 用户名 | 密码 | 角色 | 邮箱 | 权限范围 |
|--------|------|------|------|----------|
| `admin` | `Admin123!` | super_admin | admin@openpenpal.com | 系统全部权限 |

**测试功能:**
- ✅ 用户管理、角色权限管理
- ✅ 系统配置、积分规则管理
- ✅ OP Code 审核与分配
- ✅ 博物馆内容审核
- ✅ 数据分析和报告查看

---

### 🚚 **四级信使系统账户**

#### **L4 - 城市总代 (最高级别)**
| 用户名 | 密码 | 角色 | 真实姓名 | 管理区域 | OP Code前缀 |
|--------|------|------|----------|----------|-------------|
| `courier_level4` | `Secret123!` | courier_level4 | 张明 | BEIJING (全市) | `PK` |

**测试权限:**
- ✅ 创建和管理 L3 校级信使
- ✅ 全北京地区信件调度
- ✅ 批量生成 OP Code
- ✅ 跨学校信件分配
- ✅ 系统级数据分析

#### **L3 - 校级信使**
| 用户名 | 密码 | 角色 | 真实姓名 | 管理区域 | OP Code前缀 |
|--------|------|------|----------|----------|-------------|
| `courier_level3` | `Secret123!` | courier_level3 | 李华 | BJDX (北大校内) | `PK5F` |

**测试权限:**
- ✅ 创建和管理 L2 区域信使
- ✅ 北大校内全区域分发
- ✅ 校园级批量 OP Code 生成
- ✅ 区域间信件调度

#### **L2 - 区域信使**
| 用户名 | 密码 | 角色 | 真实姓名 | 管理区域 | OP Code前缀 |
|--------|------|------|----------|----------|-------------|
| `courier_level2` | `Secret123!` | courier_level2 | 王芳 | BJDX-5F (5号楼区域) | `PK5F` |

**测试权限:**
- ✅ 创建和管理 L1 楼栋信使
- ✅ 5号楼及周边区域投递
- ✅ 区域内任务分配
- ✅ 本区域信件状态管理

#### **L1 - 楼栋信使**
| 用户名 | 密码 | 角色 | 真实姓名 | 管理区域 | OP Code前缀 |
|--------|------|------|----------|----------|-------------|
| `courier_level1` | `Secret123!` | courier_level1 | 赵强 | BJDX-5F-3D (宿舍楼) | `PK5F3D` |

**测试权限:**
- ✅ 直接投递任务执行
- ✅ 扫码收件和送达
- ✅ 楼栋内信件状态更新
- ✅ 基础任务管理

---

### 👤 **普通用户账户**

| 用户名 | 密码 | 角色 | 邮箱 | 真实姓名 | 个人简介 |
|--------|------|------|------|----------|----------|
| `alice` | `Secret123!` | user | alice@openpenpal.com | 爱丽丝 | 文学摄影爱好者，大二学生 |
| `bob` | `Secret123!` | user | bob@openpenpal.com | 小明 | 计算机系学生，热爱传统文化 |

**测试功能:**
- ✅ 写信、收信、回信
- ✅ OP Code 申请与使用
- ✅ 博物馆投稿与浏览
- ✅ 积分系统体验
- ✅ AI 辅助写作

---

## 📮 预置测试数据

### **样本信件 (3封)**

1. **信件1:** "致远方朋友的第一封信"
   - 作者: alice (爱丽丝)
   - 收件人: bob (PK5F3D)
   - 状态: delivered (已送达)
   - 风格: classic (经典)

2. **信件2:** "Re: 致远方朋友的第一封信"  
   - 作者: bob (小明)
   - 收件人: alice (PK5F01)
   - 状态: read (已阅读)
   - 风格: modern (现代)

3. **信件3:** "春日校园随想"
   - 作者: alice (爱丽丝)
   - 类型: public (公开信)
   - 状态: approved (已审核)
   - 收录: 博物馆展示

### **OP Code 示例**

- `PK5F01` - Alice 的OP Code (北大5号楼01室)
- `PK5F3D` - Bob 的OP Code (北大5号楼303室)  
- `PUBLIC` - 公开信件专用

---

## 🧪 系统测试场景

### **场景1: 信件创建与投递流程**

1. **用户写信 (alice):**
   ```
   登录: alice / Secret123!
   -> 写信页面
   -> 收件人OP: PK5F3D (bob的地址)
   -> 内容: 测试信件内容
   -> 提交并生成编号
   ```

2. **信使投递 (courier_level1):**
   ```
   登录: courier_level1 / Secret123!  
   -> 扫描信件条码
   -> 更新状态: collected → in_transit → delivered
   -> 获得积分奖励 (25分)
   ```

3. **收件人确认 (bob):**
   ```
   登录: bob / Secret123!
   -> 我的信件 → 收到新信
   -> 阅读信件 (获得12积分)
   -> 可选择回信
   ```

### **场景2: 四级信使权限测试**

1. **L4权限测试 (courier_level4):**
   ```
   登录: courier_level4 / Secret123!
   -> 信使管理 → 查看全市信使
   -> 批量OP Code生成 (PK前缀)
   -> 跨学校任务分配
   ```

2. **L3权限测试 (courier_level3):**
   ```
   登录: courier_level3 / Secret123!  
   -> 校园管理 → 北大区域任务
   -> L2信使创建与管理
   -> 校园级别数据报告
   ```

### **场景3: 博物馆系统测试**

1. **内容提交 (alice):**
   ```
   登录: alice / Secret123!
   -> 博物馆 → 投稿作品
   -> 上传"春日校园随想"类型内容
   -> 获得30积分奖励
   ```

2. **内容审核 (admin):**
   ```
   登录: admin / Admin123!
   -> 管理后台 → 博物馆审核
   -> 审核alice的投稿
   -> 批准发布(alice获得100积分)
   ```

### **场景4: 积分系统测试**

**高价值任务 (admin专用):**
- OP Code审核: 200积分/次 (限1次/日)
- 管理员奖励: 500积分/次 (限1次/日)
- 社区徽章: 150积分/次 (限1次/日)

**日常任务 (用户):**
- 创建信件: 10积分/次 (限5次/日)
- 信件送达: 15积分/次 (限8次/日)  
- 收到信件: 12积分/次 (限10次/日)
- 博物馆投稿: 30积分/次 (限3次/日)

**信使任务:**
- 首次任务奖励: 50积分 (一次性)
- 投递任务: 25积分/次 (限20次/日)

---

## 🔧 AI系统测试

### **AI配置状态**
- ✅ **Moonshot AI** (主要): moonshot-v1-8k, 活跃, 10000次/日
- ❌ **OpenAI** (备用): gpt-3.5-turbo, 非活跃, 5000次/日  
- ✅ **Local Dev** (开发): development-mode, 活跃, 99999次/日

### **AI功能测试**
1. **AI写作助手:**
   ```
   登录任何用户 → 写信页面
   → AI助手 → 选择人设(诗人/哲学家/艺术家)
   → 生成写作建议
   ```

2. **AI回信生成:**
   ```
   收到信件后 → 点击"AI辅助回信"
   → 系统分析原信内容
   → 生成个性化回信建议
   ```

---

## 📊 系统监控与验证

### **数据库状态检查**
```bash
# 在backend目录执行
PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal_user -d openpenpal -c "
SELECT 'SYSTEM STATUS' as type, COUNT(*) as count FROM users
UNION ALL
SELECT 'LETTERS', COUNT(*) FROM letters  
UNION ALL
SELECT 'COURIERS', COUNT(*) FROM couriers
UNION ALL  
SELECT 'CREDIT_RULES', COUNT(*) FROM credit_task_rules
ORDER BY type;"
```

### **服务健康检查**
```bash
# 检查后端API
curl http://localhost:8080/api/health

# 检查前端访问  
curl http://localhost:3000

# 检查数据库连接
./backend/main --test-db
```

### **日志监控**
```bash
# 后端日志
tail -f backend/backend.log

# 前端开发日志
# 查看浏览器控制台

# 数据库日志
tail -f /usr/local/var/log/postgresql@14.log
```

---

## 🎯 重点测试功能

### **✅ 必测功能清单**

- [ ] **用户认证**: 所有账号登录/登出
- [ ] **信件流程**: 写信→生成码→投递→送达→阅读
- [ ] **四级权限**: L4→L3→L2→L1 层级管理权限
- [ ] **OP Code**: 申请→审核→分配→使用
- [ ] **积分系统**: 任务完成→积分获得→限制检查
- [ ] **博物馆**: 投稿→审核→发布→浏览
- [ ] **AI辅助**: 写作建议→回信生成→内容优化
- [ ] **扫码功能**: 条码生成→扫描识别→状态更新

### **🚫 已知限制**

- **OpenAI配置**: 当前非活跃状态 (需要有效API密钥)
- **支付系统**: 积分商城仅演示功能
- **邮件通知**: 开发环境下邮件功能有限
- **文件上传**: 本地存储，生产环境需配置云存储

---

## 🆘 故障排除

### **常见问题解决**

1. **无法登录**
   - 检查密码是否正确 (Secret123!/Admin123!)
   - 确认用户状态: `is_active = true`

2. **积分不增加**
   - 检查每日限制是否达到
   - 确认任务规则是否启用

3. **信使权限不足**
   - 验证 OP Code 前缀权限
   - 检查信使等级和管理区域

4. **AI功能不工作**
   - 确认AI配置状态为 `is_active = true`
   - 检查Daily quota是否耗尽

### **重置方法**

```bash
# 重置用户密码 (在backend目录)
go run cmd/admin/reset_passwords.go -user=alice -password=newpass

# 重新注入测试数据
PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal_user -d openpenpal -f scripts/seed_essential_data_corrected.sql

# 清理并重启服务
./startup/force-cleanup.sh
./startup/quick-start.sh demo --auto-open
```

---

## 📋 测试记录模板

### **测试会话记录**

```markdown
**测试日期**: 2025-XX-XX
**测试人员**: [姓名]  
**测试账号**: [使用的账号]
**测试场景**: [场景描述]

**执行步骤**:
1. 登录系统
2. [具体操作步骤]
3. 验证结果

**测试结果**: ✅通过 / ❌失败
**发现问题**: [问题描述]
**改进建议**: [建议内容]
```

---

## 🎊 总结

OpenPenPal测试环境已完全配置，包含:
- ✅ **7个完整测试账号** (1管理员 + 4信使 + 2用户)
- ✅ **161个数据库表** 完整结构
- ✅ **完整业务数据** 注入和持久化
- ✅ **四级信使系统** 完整权限体系  
- ✅ **积分经济系统** 16种任务奖励规则
- ✅ **AI辅助功能** 多提供商配置
- ✅ **样本内容** 3封示例信件

**开始测试**: 使用任何上述账号登录系统，体验完整的OpenPenPal校园手写信平台功能！

---

*🤖 Generated with ULTRATHINK deep analysis - 完整数据库分析和测试账号体系*