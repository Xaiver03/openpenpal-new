/**
 * 权限系统核心功能单元测试
 * 覆盖SOTA权限系统的主要功能点
 */

package permissions

import (
	"testing"
)

// TestUserRolePermissions 测试用户角色权限
func TestUserRolePermissions(t *testing.T) {
	service := NewService()
	
	tests := []struct {
		name       string
		user       User
		permission string
		expected   bool
	}{
		{
			name: "普通用户基础权限",
			user: User{
				Role: RoleUser,
			},
			permission: "READ_LETTER",
			expected:   true,
		},
		{
			name: "普通用户无信使权限",
			user: User{
				Role: RoleUser,
			},
			permission: "COURIER_SCAN_CODE",
			expected:   false,
		},
		{
			name: "信使有信使权限",
			user: User{
				Role: RoleCourier,
				CourierInfo: &CourierInfo{
					Level: CourierLevel1,
				},
			},
			permission: "COURIER_SCAN_CODE",
			expected:   true,
		},
		{
			name: "管理员有所有权限",
			user: User{
				Role: RoleSuperAdmin,
			},
			permission: "SYSTEM_ADMIN",
			expected:   true,
		},
		{
			name: "信使等级权限检查",
			user: User{
				Role: RoleCourier,
				CourierInfo: &CourierInfo{
					Level: CourierLevel3,
				},
			},
			permission: "MANAGE_SUBORDINATES",
			expected:   true,
		},
		{
			name: "低级信使无管理权限",
			user: User{
				Role: RoleCourier,
				CourierInfo: &CourierInfo{
					Level: CourierLevel1,
				},
			},
			permission: "MANAGE_SUBORDINATES",
			expected:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.HasPermission(tt.user, tt.permission)
			if result != tt.expected {
				t.Errorf("HasPermission() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetUserPermissions 测试获取用户权限列表
func TestGetUserPermissions(t *testing.T) {
	service := NewService()
	
	tests := []struct {
		name     string
		user     User
		minPerms int
	}{
		{
			name: "普通用户权限数量",
			user: User{
				Role: RoleUser,
			},
			minPerms: 5, // 基础权限
		},
		{
			name: "信使权限数量",
			user: User{
				Role: RoleCourier,
				CourierInfo: &CourierInfo{
					Level: CourierLevel1,
				},
			},
			minPerms: 10, // 基础权限 + 信使权限
		},
		{
			name: "高级信使权限数量",
			user: User{
				Role: RoleSeniorCourier,
				CourierInfo: &CourierInfo{
					Level: CourierLevel3,
				},
			},
			minPerms: 15, // 更多管理权限
		},
		{
			name: "超级管理员权限数量",
			user: User{
				Role: RoleSuperAdmin,
			},
			minPerms: 25, // 所有权限
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := service.GetUserPermissions(tt.user)
			if len(perms) < tt.minPerms {
				t.Errorf("GetUserPermissions() returned %d permissions, want at least %d", len(perms), tt.minPerms)
			}
		})
	}
}

// TestPermissionModules 测试权限模块定义
func TestPermissionModules(t *testing.T) {
	service := NewService()
	
	// 测试所有权限模块都已定义
	allModules := service.GetAllPermissionModules()
	if len(allModules) != 29 {
		t.Errorf("Expected 29 permission modules, got %d", len(allModules))
	}
	
	// 测试特定权限模块
	courierScan := service.GetPermissionModule("COURIER_SCAN_CODE")
	if courierScan == nil {
		t.Error("COURIER_SCAN_CODE module should exist")
	}
	if courierScan.Category != CategoryCourier {
		t.Errorf("COURIER_SCAN_CODE should be in courier category, got %s", courierScan.Category)
	}
	
	// 测试系统核心权限
	sysAdmin := service.GetPermissionModule("SYSTEM_ADMIN")
	if sysAdmin == nil {
		t.Error("SYSTEM_ADMIN module should exist")
	}
	if sysAdmin.RiskLevel != RiskCritical {
		t.Errorf("SYSTEM_ADMIN should have critical risk level, got %s", sysAdmin.RiskLevel)
	}
}

// TestDynamicPermissionUpdate 测试动态权限更新
func TestDynamicPermissionUpdate(t *testing.T) {
	service := NewService()
	
	// 测试更新角色权限
	newPerms := []string{"READ_LETTER", "WRITE_LETTER", "COURIER_SCAN_CODE"}
	err := service.UpdateRolePermissions(RoleUser, newPerms, "test_admin")
	if err != nil {
		t.Fatalf("UpdateRolePermissions failed: %v", err)
	}
	
	// 验证更新后的权限
	user := User{Role: RoleUser}
	if !service.HasPermission(user, "COURIER_SCAN_CODE") {
		t.Error("User should have COURIER_SCAN_CODE permission after update")
	}
	
	// 重置权限
	service.ResetRolePermissions(RoleUser)
	if service.HasPermission(user, "COURIER_SCAN_CODE") {
		t.Error("User should not have COURIER_SCAN_CODE permission after reset")
	}
}

// TestPermissionAnalysis 测试权限分析
func TestPermissionAnalysis(t *testing.T) {
	service := NewService()
	
	tests := []struct {
		name         string
		user         User
		minCoverage  float64
		expectedRisk RiskLevel
	}{
		{
			name: "普通用户分析",
			user: User{
				Role: RoleUser,
			},
			minCoverage:  10.0,
			expectedRisk: RiskLow,
		},
		{
			name: "管理员分析",
			user: User{
				Role: RoleAdmin,
			},
			minCoverage:  60.0,
			expectedRisk: RiskHigh,
		},
		{
			name: "超级管理员分析",
			user: User{
				Role: RoleSuperAdmin,
			},
			minCoverage:  90.0,
			expectedRisk: RiskCritical,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := service.AnalyzeUserPermissions(tt.user)
			if analysis.Coverage < tt.minCoverage {
				t.Errorf("Coverage = %.2f%%, want at least %.2f%%", analysis.Coverage, tt.minCoverage)
			}
			if analysis.RiskLevel != tt.expectedRisk {
				t.Errorf("RiskLevel = %s, want %s", analysis.RiskLevel, tt.expectedRisk)
			}
		})
	}
}

// TestCourierLevelPermissions 测试信使等级权限
func TestCourierLevelPermissions(t *testing.T) {
	service := NewService()
	
	// 测试不同等级的信使权限差异
	level1Perms := service.GetCourierLevelPermissions(CourierLevel1)
	level3Perms := service.GetCourierLevelPermissions(CourierLevel3)
	level4Perms := service.GetCourierLevelPermissions(CourierLevel4)
	
	if len(level1Perms) >= len(level3Perms) {
		t.Error("Level 3 courier should have more permissions than level 1")
	}
	
	if len(level3Perms) >= len(level4Perms) {
		t.Error("Level 4 courier should have more permissions than level 3")
	}
	
	// 验证管理权限只给高级信使
	hasManagePerms := false
	for _, perm := range level1Perms {
		if perm == "MANAGE_SUBORDINATES" {
			hasManagePerms = true
			break
		}
	}
	if hasManagePerms {
		t.Error("Level 1 courier should not have MANAGE_SUBORDINATES permission")
	}
}

// TestPermissionDependencies 测试权限依赖关系
func TestPermissionDependencies(t *testing.T) {
	service := NewService()
	
	// 测试权限依赖 - MANAGE_PERMISSIONS 需要 SYSTEM_ADMIN
	managePermsModule := service.GetPermissionModule("MANAGE_PERMISSIONS")
	if managePermsModule == nil {
		t.Fatal("MANAGE_PERMISSIONS module should exist")
	}
	
	foundDep := false
	for _, dep := range managePermsModule.Dependencies {
		if dep == "SYSTEM_ADMIN" {
			foundDep = true
			break
		}
	}
	if !foundDep {
		t.Error("MANAGE_PERMISSIONS should depend on SYSTEM_ADMIN")
	}
}

// TestQuickCheckFunctions 测试便捷函数
func TestQuickCheckFunctions(t *testing.T) {
	// 测试快速权限检查
	user := User{
		Role: RoleAdmin,
	}
	
	if !QuickCheck(user, "MANAGE_USERS") {
		t.Error("Admin should have MANAGE_USERS permission")
	}
	
	if !QuickCanAccessAdmin(user) {
		t.Error("Admin should be able to access admin panel")
	}
	
	// 测试快速分析
	analysis := QuickAnalyze(user)
	if analysis == nil {
		t.Error("QuickAnalyze should return analysis result")
	}
}

// BenchmarkPermissionCheck 性能测试
func BenchmarkPermissionCheck(b *testing.B) {
	service := NewService()
	user := User{
		Role: RoleCourier,
		CourierInfo: &CourierInfo{
			Level: CourierLevel2,
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.HasPermission(user, "COURIER_SCAN_CODE")
	}
}

// BenchmarkGetUserPermissions 获取权限列表性能测试
func BenchmarkGetUserPermissions(b *testing.B) {
	service := NewService()
	user := User{
		Role: RoleAdmin,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetUserPermissions(user)
	}
}