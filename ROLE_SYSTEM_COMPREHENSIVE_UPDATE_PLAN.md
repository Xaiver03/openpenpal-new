# OpenPenPal 角色系统全面更新计划

> 生成时间: 2025-08-09
> 目标: 统一前后端角色系统，严格遵循PRD设计

## 📊 当前状态分析

### 发现的问题

1. **后端问题**
   - ✅ `user.go` 已更新角色定义
   - ❌ `database.go` 包含旧角色的种子数据
   - ❌ 其他服务文件可能引用旧角色

2. **前端问题**
   - ❌ `roles.ts` 包含所有冗余角色定义
   - ❌ 前端类型定义与后端不一致

3. **数据库问题**
   - ✅ 实际数据已符合PRD（只有四级信使）
   - ❌ 种子数据脚本需要更新

## 🎯 目标角色体系

```
正确的角色：
- user              # 普通用户
- courier_level1    # 一级信使（基础投递信使）
- courier_level2    # 二级信使（片区协调员）  
- courier_level3    # 三级信使（校区负责人）
- courier_level4    # 四级信使（城市负责人）
- platform_admin    # 平台管理员
- super_admin       # 超级管理员

需要移除的角色：
- courier           ❌
- senior_courier    ❌
- courier_coordinator ❌
- school_admin      ❌
- admin             ❌ (统一使用 super_admin)
```

## 📝 实施步骤

### Step 1: 后端代码更新

#### 1.1 更新 database.go
```go
// 移除这些用户
- courier (ID: 5)
- senior_courier (ID: 6)
- courier_coordinator (ID: 7)
- school_admin (ID: 8)
- courier_1 (ID: 11)
- courier_2 (ID: 12)
- courier_3 (ID: 13)

// 保留并确保正确
- courier_level1 (ID: 9)
- courier_level2 (ID: 10)
- courier_level3 (ID: 3)
- courier_level4 (ID: 4)
```

#### 1.2 检查其他后端文件
需要检查并更新：
- `main.go` - 路由权限检查
- `courier_service.go` - 信使服务逻辑
- `websocket/client.go` - WebSocket权限
- `role_mapping.go` - 角色映射

### Step 2: 前端代码更新

#### 2.1 简化 roles.ts
```typescript
export type UserRole = 
  | 'user'
  | 'courier_level1'
  | 'courier_level2'
  | 'courier_level3'
  | 'courier_level4'
  | 'platform_admin'
  | 'super_admin'
```

#### 2.2 更新组件
- 移除对旧角色的引用
- 更新权限检查逻辑
- 确保导航菜单正确

### Step 3: 数据迁移

#### 3.1 数据库迁移SQL
```sql
-- 如果有使用旧角色的用户
UPDATE users SET role = 'courier_level1' WHERE role = 'courier';
UPDATE users SET role = 'courier_level2' WHERE role = 'senior_courier';
UPDATE users SET role = 'courier_level2' WHERE role = 'courier_coordinator';
UPDATE users SET role = 'courier_level3' WHERE role = 'school_admin';
UPDATE users SET role = 'super_admin' WHERE role = 'admin';
```

### Step 4: 测试验证

#### 4.1 功能测试
- [ ] 各角色登录测试
- [ ] 权限访问测试
- [ ] 信使功能测试
- [ ] 管理功能测试

#### 4.2 集成测试
- [ ] 前后端角色一致性
- [ ] API权限验证
- [ ] WebSocket权限

## 🚀 执行顺序

1. **Phase 1**: 后端更新（优先）
   - 更新database.go种子数据
   - 检查并更新其他服务文件
   - 运行单元测试

2. **Phase 2**: 前端同步
   - 更新roles.ts类型定义
   - 更新组件中的角色检查
   - 运行前端测试

3. **Phase 3**: 集成验证
   - 端到端测试
   - 用户体验测试
   - 性能测试

## ⚠️ 风险管理

1. **向后兼容性**
   - 保留角色映射逻辑一段时间
   - 提供数据迁移脚本

2. **用户影响**
   - 提前通知用户角色调整
   - 确保权限不会降级

3. **回滚计划**
   - 备份所有修改的文件
   - 准备回滚脚本

## 📊 预期结果

1. **代码简化**
   - 减少50%的角色定义代码
   - 更清晰的权限层级

2. **维护性提升**
   - 统一的角色体系
   - 更容易理解和扩展

3. **用户体验**
   - 更清晰的角色定位
   - 更合理的权限分配

## ✅ 完成标准

- [ ] 所有代码中只存在7种角色
- [ ] 前后端角色类型完全一致
- [ ] 所有测试通过
- [ ] 文档更新完成
- [ ] 没有功能退化