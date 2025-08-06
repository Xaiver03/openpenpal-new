# OpenPenPal 晋升系统 - 全面测试指南

## 测试环境准备

确保以下服务正在运行：
- 后端服务: `http://localhost:8080`
- 前端服务: `http://localhost:3000`

## 测试用户账号

- **学生用户**: alice / secret
- **一级信使**: courier_level1 / secret
- **二级信使**: courier_level2 / secret
- **三级信使**: courier_level3 / secret
- **四级信使**: courier_level4 / secret
- **管理员**: admin / admin123

## 1. CSRF保护测试 ✅

### 测试步骤
1. 打开浏览器访问 `http://localhost:3000`
2. 打开开发者工具 (F12)，切换到Network标签
3. 尝试登录，观察请求头中的CSRF token
4. 验证后端API要求CSRF token

### 预期结果
- 登录请求应包含 `X-CSRF-Token` 请求头
- 缺少CSRF token的请求应返回403错误

## 2. 数据库功能测试 🔄

### 2.1 晋升路径查询测试

**测试步骤**：
1. 登录为 courier_level1 用户
2. 访问 `http://localhost:3000/courier/growth`
3. 检查晋升路径是否正确显示

**预期结果**：
- 显示当前等级: 一级信使
- 显示下一级目标: 二级信使
- 显示晋升要求

### 2.2 晋升申请提交测试

**测试步骤**：
1. 在晋升页面点击"申请晋升"按钮
2. 填写晋升理由
3. 提交申请

**预期结果**：
- 申请成功提交
- 数据库中创建晋升记录
- 页面显示申请状态

### 2.3 数据库验证SQL

```sql
-- 查看所有信使用户
SELECT u.id, u.username, u.role, c.level, c.zone, c.status
FROM users u
LEFT JOIN couriers c ON u.id = c.user_id
WHERE u.role LIKE 'courier%';

-- 查看晋升申请记录
SELECT * FROM courier_upgrade_requests
ORDER BY created_at DESC;

-- 查看信使层级关系
SELECT 
    c1.id, c1.nickname as courier_name, c1.level,
    c2.nickname as parent_name, c2.level as parent_level
FROM couriers c1
LEFT JOIN couriers c2 ON c1.parent_id = c2.id
ORDER BY c1.level DESC, c1.created_at;
```

## 3. 权限边界测试

### 3.1 一级信使权限测试

**测试账号**: courier_level1 / secret

**测试内容**：
- [x] 能否访问晋升页面 → 应该可以
- [ ] 能否查看晋升申请管理 → 应该不能（需要3级权限）
- [ ] 能否批准他人晋升 → 应该不能

### 3.2 三级信使权限测试

**测试账号**: courier_level3 / secret

**测试内容**：
- [x] 能否访问晋升页面 → 应该可以
- [x] 能否查看晋升申请管理 → 应该可以
- [x] 能否批准下级晋升申请 → 应该可以（只能批准比自己低级的）

### 3.3 权限测试API

```bash
# 测试无权限访问管理接口
curl -H "Authorization: Bearer <courier_level1_token>" \
  http://localhost:8080/api/v1/courier/level/upgrade-requests

# 应返回403 Forbidden
```

## 4. 错误处理测试

### 4.1 无效晋升申请测试

**测试场景**：
1. 四级信使申请晋升（已是最高级）
2. 提交空的晋升理由
3. 申请跳级晋升（1级直接申请3级）

**预期结果**：
- 应返回合适的错误消息
- 前端应显示友好的错误提示

### 4.2 网络错误处理

**测试步骤**：
1. 停止后端服务
2. 尝试提交晋升申请
3. 检查前端错误处理

**预期结果**：
- 前端应显示网络错误提示
- 不应崩溃或显示技术性错误

## 5. 端到端流程测试

### 完整晋升流程

1. **登录一级信使** (courier_level1)
   - 查看当前等级和晋升要求
   - 提交晋升申请

2. **切换到三级信使** (courier_level3)
   - 访问晋升管理页面
   - 查看待审核申请
   - 批准晋升申请

3. **返回一级信使账号**
   - 刷新页面查看晋升结果
   - 验证等级是否更新

## 测试检查清单

- [ ] CSRF保护正常工作
- [ ] 晋升路径正确显示
- [ ] 晋升申请可以提交
- [ ] 权限控制正确实施
- [ ] 错误处理友好准确
- [ ] 完整流程可以走通

## 常见问题排查

1. **登录失败**
   - 检查后端是否运行
   - 检查数据库连接
   - 验证用户密码

2. **页面无法访问**
   - 检查前端服务是否运行
   - 清除浏览器缓存
   - 检查控制台错误

3. **权限错误**
   - 确认登录的用户角色
   - 检查JWT token是否过期
   - 验证后端权限中间件

## 性能测试建议

1. 测试并发晋升申请
2. 测试大量数据时的查询性能
3. 测试WebSocket实时通知功能