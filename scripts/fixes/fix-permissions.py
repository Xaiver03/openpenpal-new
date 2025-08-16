#!/usr/bin/env python3
import re

# 读取文件
with open('backend/internal/models/user.go', 'r') as f:
    content = f.read()

# 新的权限映射定义
new_permissions = '''// RolePermissions 角色权限映射 - 简化版本符合PRD
var RolePermissions = map[UserRole][]Permission{
	RoleUser: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
	},

	// 一级信使：基础投递
	RoleCourierLevel1: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
	},

	// 二级信使：片区协调员 - 可以分发任务给一级信使
	RoleCourierLevel2: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionAssignTasks, // 分发任务给一级信使
	},

	// 三级信使：校区负责人 - 可以任命二级信使，查看报告
	RoleCourierLevel3: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionAssignTasks,
		PermissionManageCouriers, // 任命二级信使
		PermissionViewReports,    // 查看报告
	},

	// 四级信使：城市负责人 - 拥有城市级权限
	RoleCourierLevel4: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionAssignTasks,
		PermissionManageCouriers,
		PermissionViewReports,
		PermissionManageSchool,   // 开通新学校
		PermissionViewAnalytics,  // 查看城市级数据
	},

	// 平台管理员
	RolePlatformAdmin: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionManageUsers,
		PermissionViewReports,
		PermissionViewAnalytics,
		PermissionManageSystem,
	},

	// 超级管理员 - 拥有所有权限
	RoleSuperAdmin: {
		PermissionWriteLetter,
		PermissionReadLetter,
		PermissionManageProfile,
		PermissionDeliverLetter,
		PermissionScanCode,
		PermissionViewTasks,
		PermissionManageCouriers,
		PermissionAssignTasks,
		PermissionViewReports,
		PermissionManageUsers,
		PermissionManageSchool,
		PermissionViewAnalytics,
		PermissionManageSystem,
		PermissionManagePlatform,
		PermissionManageAdmins,
		PermissionSystemConfig,
	},
}'''

# 使用正则表达式替换 RolePermissions 定义
pattern = r'// RolePermissions 角色权限映射\nvar RolePermissions = map\[UserRole\]\[\]Permission\{[\s\S]*?\n\}'
content = re.sub(pattern, new_permissions, content)

# 写回文件
with open('backend/internal/models/user.go', 'w') as f:
    f.write(content)

print("✅ Successfully updated RolePermissions!")