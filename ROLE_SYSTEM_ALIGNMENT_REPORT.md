# OpenPenPal 角色系统对齐报告

## 📋 PRD要求 vs 现状对比

### ✅ 应保留的角色

| 角色 | 说明 | PRD依据 |
|------|------|---------|
| `user` | 普通用户 | 基础用户角色 |
| `courier_level1` | 一级信使（基础投递信使） | 楼栋/商店路径投递 |
| `courier_level2` | 二级信使（片区协调员） | 管理5-6位编码段 |
| `courier_level3` | 三级信使（校区负责人） | 管理3-4位编码段，任命二级信使 |
| `courier_level4` | 四级信使（城市负责人） | 城市级物流调度，开通新学校 |
| `platform_admin` | 平台管理员 | 平台运营管理 |
| `super_admin` | 超级管理员 | 系统最高权限 |

### ❌ 需要移除的冗余角色

| 冗余角色 | 移除原因 | 迁移方案 |
|----------|----------|----------|
| `courier` | 与四级信使体系重复 | → `courier_level1` |
| `senior_courier` | 不符合四级体系 | → `courier_level2` |
| `courier_coordinator` | 角色定位与二级信使重复 | → `courier_level2` |
| `school_admin` | 三级信使已是校区负责人 | → `courier_level3` |

## 🔧 角色权限对应关系

### 四级信使权限层级（根据PRD）

```
courier_level4 (城市负责人)
├── 开通新学校
├── 城市级物流调度
├── 设计城市活动信封
└── 管理前两位编码

courier_level3 (校区负责人)
├── 任命二级信使
├── 设计校内信封
├── 管理3-4位编码段
├── 调度本校物流
└── 审核信使成长申请

courier_level2 (片区协调员)
├── 管理5-6位编码段
├── 审核新点位申请
└── 分发任务给一级信使

courier_level1 (基础投递信使)
├── 领取任务
├── 扫码更新条码状态
├── 完成实际派送流程
└── 投递反馈
```

## 📊 数据库清理SQL

```sql
-- 1. 合并冗余角色到四级体系
UPDATE users SET role = 'courier_level1' WHERE role = 'courier';
UPDATE users SET role = 'courier_level2' WHERE role IN ('senior_courier', 'courier_coordinator');
UPDATE users SET role = 'courier_level3' WHERE role = 'school_admin';

-- 2. 验证角色分布
SELECT role, COUNT(*) as count 
FROM users 
GROUP BY role 
ORDER BY 
  CASE role
    WHEN 'super_admin' THEN 1
    WHEN 'platform_admin' THEN 2
    WHEN 'courier_level4' THEN 3
    WHEN 'courier_level3' THEN 4
    WHEN 'courier_level2' THEN 5
    WHEN 'courier_level1' THEN 6
    WHEN 'user' THEN 7
  END;
```

## 🎯 代码层面需要的调整

### 1. 更新角色常量定义
```go
// internal/models/user.go
const (
    RoleUser          UserRole = "user"           // 普通用户
    RoleCourierLevel1 UserRole = "courier_level1" // 一级信使
    RoleCourierLevel2 UserRole = "courier_level2" // 二级信使
    RoleCourierLevel3 UserRole = "courier_level3" // 三级信使
    RoleCourierLevel4 UserRole = "courier_level4" // 四级信使
    RolePlatformAdmin UserRole = "platform_admin" // 平台管理员
    RoleSuperAdmin    UserRole = "super_admin"    // 超级管理员
)
```

### 2. 更新角色层级
```go
var RoleHierarchy = map[UserRole]int{
    RoleUser:          1,
    RoleCourierLevel1: 2,
    RoleCourierLevel2: 3,
    RoleCourierLevel3: 4,
    RoleCourierLevel4: 5,
    RolePlatformAdmin: 6,
    RoleSuperAdmin:    7,
}
```

### 3. 更新权限映射
根据PRD，每级信使的权限应该是累加的，高级别信使拥有低级别的所有权限。

## ✅ 实施建议

1. **数据迁移**：先运行SQL脚本统一现有用户角色
2. **代码更新**：更新后端角色定义和权限映射
3. **前端同步**：确保前端角色常量与后端一致
4. **测试验证**：验证各级信使权限正确性

## 📝 总结

通过此次调整，系统角色将完全符合PRD设计：
- 保留必要的管理角色（platform_admin、super_admin）
- 严格遵循四级信使体系
- 移除与四级体系重复或不符的角色
- 权限体系更加清晰和模块化