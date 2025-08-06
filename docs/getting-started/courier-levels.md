# 信使层级体系与权限管理

## 概述

OpenPenPal 实现了完整的四级信使层级体系，每个层级都有明确的权限范围和管理职责。

## 层级结构

| **层级** | **角色名称** | **管辖范围** | **创建方式** | **管理权限** |
|---------|------------|-----------|-----------|-----------|
| 4级 | 四级信使（城市总代） | 整个城市的学校信件流转 | 系统管理员授权 | 管理三级信使，城市运营分析 |
| 3级 | 三级信使（校级） | 本校信件中转与派发 | 四级信使授权 | 管理二级信使，校园区域协调 |
| 2级 | 二级信使（片区/年级） | 片区内信件整合派送 | 三级信使授权 | 管理一级信使，任务分配 |
| 1级 | 一级信使（楼栋/班级） | 实际收集与投递信件 | 二级信使邀请 | 无管理权限，执行配送任务 |

## 权限详解

### 一级信使（楼栋/班级）
- **基础权限**：
  - `courier_scan_code` - 扫描二维码
  - `courier_deliver_letter` - 投递信件
  - `courier_view_own_tasks` - 查看个人任务
  - `courier_report_exception` - 异常报告
- **管理界面**：`/courier/tasks` - 个人任务界面

### 二级信使（片区/年级）
- **继承权限**：所有一级信使权限
- **管理权限**：
  - `courier_manage_subordinates` - 管理一级信使
  - `courier_assign_tasks` - 分配任务
  - `courier_view_subordinate_reports` - 查看下级报告
  - `courier_create_subordinate` - 创建一级信使账号
- **管理界面**：`/courier/zone-manage` - 片区管理后台

### 三级信使（校级）
- **继承权限**：所有二级信使权限
- **高级权限**：
  - `courier_manage_school_zone` - 管理校园区域
  - `courier_view_school_analytics` - 查看学校分析数据
  - `courier_coordinate_cross_zone` - 跨区域协调
- **管理界面**：`/courier/school-manage` - 学校管理后台

### 四级信使（城市总代）
- **继承权限**：所有三级信使权限
- **城市权限**：
  - `courier_manage_city_operations` - 城市运营管理
  - `courier_create_school_courier` - 创建校级信使
  - `courier_view_city_analytics` - 查看城市分析数据
- **管理界面**：`/courier/city-manage` - 城市管理后台

## 测试账号

### 层级信使测试账号

| **层级** | **用户名** | **密码** | **说明** |
|---------|----------|---------|---------|
| 4级 | `courier_level4_city` | `city123` | 城市总代，管理全市信使网络 |
| 3级 | `courier_level3_school` | `school123` | 校级信使，管理学校信使团队 |
| 2级 | `courier_level2_zone` | `zone123` | 片区信使，管理一级信使 |
| 1级 | `courier_level1_basic` | `basic123` | 基础信使，执行配送任务 |

### 快速测试步骤

1. **测试一级信使**：
   ```bash
   # 登录一级信使账号
   用户名: courier_level1_basic
   密码: basic123
   
   # 可访问页面
   - /courier/tasks (个人任务)
   - 无管理权限
   ```

2. **测试二级信使管理功能**：
   ```bash
   # 登录二级信使账号
   用户名: courier_level2_zone
   密码: zone123
   
   # 可访问页面
   - /courier/zone-manage (片区管理)
   - 可创建和管理一级信使
   - 可分配任务给一级信使
   ```

3. **测试三级信使校园管理**：
   ```bash
   # 登录三级信使账号
   用户名: courier_level3_school
   密码: school123
   
   # 可访问页面
   - /courier/school-manage (学校管理)
   - 可创建和管理二级信使
   - 查看学校配送分析
   ```

4. **测试四级信使城市运营**：
   ```bash
   # 登录四级信使账号
   用户名: courier_level4_city
   密码: city123
   
   # 可访问页面
   - /courier/city-manage (城市管理)
   - 可创建和管理三级信使
   - 查看城市运营数据
   ```

## 权限验证代码示例

```typescript
import { useCourierPermission } from '@/hooks/use-courier-permission'

function CourierManagementPage() {
  const { 
    courierInfo, 
    canManageSubordinates,
    getCourierLevelName,
    getManagementDashboardPath 
  } = useCourierPermission()

  // 检查是否可以管理下级
  if (!canManageSubordinates()) {
    return <div>您没有管理权限</div>
  }

  // 获取当前信使级别
  const levelName = getCourierLevelName() // 例如："二级信使（片区/年级）"

  // 获取管理后台路径
  const dashboardPath = getManagementDashboardPath() // 例如："/courier/zone-manage"

  return (
    <div>
      <h1>{levelName}管理后台</h1>
      {/* 管理界面内容 */}
    </div>
  )
}
```

## 注意事项

1. **权限继承**：高级别信使自动继承所有低级别权限
2. **管理范围**：每个级别只能管理直接下级（如三级管理二级，二级管理一级）
3. **创建限制**：信使只能创建比自己低一级的信使账号
4. **区域绑定**：每个信使都绑定特定区域（城市/学校/片区/楼栋）

## 与旧角色系统的区别

旧系统中的角色（如 `senior_courier`）是基于功能的静态角色，而新的层级信使体系是基于管理层级的动态权限系统：

- `courier` → 对应一级信使（基础配送）
- `senior_courier` → 功能增强但无管理权限
- `courier_coordinator` → 类似二级信使但不完全相同
- 层级信使系统 → 完整的四级管理体系，权限动态继承