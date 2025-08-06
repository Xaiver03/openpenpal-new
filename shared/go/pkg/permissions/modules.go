/**
 * 权限模块定义 - SOTA权限系统核心模块
 */

package permissions

import "sync"

var (
	permissionModules map[string]*PermissionModule
	modulesOnce      sync.Once
)

// GetPermissionModules 获取所有权限模块
func GetPermissionModules() map[string]*PermissionModule {
	modulesOnce.Do(initPermissionModules)
	return permissionModules
}

// GetPermissionModule 获取指定权限模块
func GetPermissionModule(id string) *PermissionModule {
	modules := GetPermissionModules()
	return modules[id]
}

// initPermissionModules 初始化权限模块定义
func initPermissionModules() {
	permissionModules = map[string]*PermissionModule{
		// 基础权限模块
		"READ_LETTER": {
			ID:          "READ_LETTER",
			Name:        "阅读信件",
			Description: "查看和阅读收到的信件内容",
			Category:    CategoryBasic,
			RiskLevel:   RiskLow,
		},
		"WRITE_LETTER": {
			ID:          "WRITE_LETTER",
			Name:        "写信",
			Description: "撰写和发送信件给其他用户",
			Category:    CategoryBasic,
			RiskLevel:   RiskLow,
		},
		"MANAGE_PROFILE": {
			ID:          "MANAGE_PROFILE",
			Name:        "管理个人资料",
			Description: "修改个人基本信息和设置",
			Category:    CategoryBasic,
			RiskLevel:   RiskLow,
		},
		"VIEW_PLAZA": {
			ID:          "VIEW_PLAZA",
			Name:        "浏览广场",
			Description: "查看公共信件广场内容",
			Category:    CategoryBasic,
			RiskLevel:   RiskLow,
		},
		"PARTICIPATE_PLAZA": {
			ID:          "PARTICIPATE_PLAZA",
			Name:        "参与广场互动",
			Description: "在广场发表评论和互动",
			Category:    CategoryBasic,
			RiskLevel:   RiskLow,
		},

		// 信使权限模块
		"COURIER_SCAN_CODE": {
			ID:          "COURIER_SCAN_CODE",
			Name:        "扫描信件",
			Description: "扫描信件二维码进行配送",
			Category:    CategoryCourier,
			RiskLevel:   RiskMedium,
		},
		"COURIER_DELIVER_LETTER": {
			ID:          "COURIER_DELIVER_LETTER",
			Name:        "投递信件",
			Description: "完成信件的最终投递",
			Category:    CategoryCourier,
			RiskLevel:   RiskMedium,
		},
		"COURIER_VIEW_TASKS": {
			ID:          "COURIER_VIEW_TASKS",
			Name:        "查看任务",
			Description: "查看分配给自己的配送任务",
			Category:    CategoryCourier,
			RiskLevel:   RiskLow,
		},
		"COURIER_UPDATE_STATUS": {
			ID:          "COURIER_UPDATE_STATUS",
			Name:        "更新配送状态",
			Description: "更新信件配送进度状态",
			Category:    CategoryCourier,
			RiskLevel:   RiskMedium,
		},
		"COURIER_VIEW_POINTS": {
			ID:          "COURIER_VIEW_POINTS",
			Name:        "查看积分",
			Description: "查看个人配送积分和等级",
			Category:    CategoryCourier,
			RiskLevel:   RiskLow,
		},

		// 管理权限模块
		"MANAGE_SUBORDINATES": {
			ID:          "MANAGE_SUBORDINATES",
			Name:        "管理下级信使",
			Description: "管理和指导下级信使工作",
			Category:    CategoryManagement,
			RiskLevel:   RiskMedium,
			Dependencies: []string{"COURIER_VIEW_TASKS"},
		},
		"ASSIGN_TASKS": {
			ID:          "ASSIGN_TASKS",
			Name:        "分配任务",
			Description: "为信使分配配送任务",
			Category:    CategoryManagement,
			RiskLevel:   RiskMedium,
		},
		"VIEW_REGION_STATS": {
			ID:          "VIEW_REGION_STATS",
			Name:        "查看区域统计",
			Description: "查看负责区域的配送统计",
			Category:    CategoryManagement,
			RiskLevel:   RiskLow,
		},
		"MANAGE_POSTAL_CODES": {
			ID:          "MANAGE_POSTAL_CODES",
			Name:        "管理邮政编码",
			Description: "管理配送区域的邮政编码",
			Category:    CategoryManagement,
			RiskLevel:   RiskMedium,
		},
		"APPROVE_COURIER_APPLICATIONS": {
			ID:          "APPROVE_COURIER_APPLICATIONS",
			Name:        "审批信使申请",
			Description: "审核和批准新信使的申请",
			Category:    CategoryManagement,
			RiskLevel:   RiskHigh,
		},

		// 管理员权限模块
		"MANAGE_USERS": {
			ID:          "MANAGE_USERS",
			Name:        "管理用户",
			Description: "管理平台用户账户和权限",
			Category:    CategoryAdmin,
			RiskLevel:   RiskHigh,
		},
		"MANAGE_LETTERS": {
			ID:          "MANAGE_LETTERS",
			Name:        "管理信件",
			Description: "管理和审核平台信件内容",
			Category:    CategoryAdmin,
			RiskLevel:   RiskHigh,
		},
		"MANAGE_COURIERS": {
			ID:          "MANAGE_COURIERS",
			Name:        "管理信使",
			Description: "管理信使账户和等级",
			Category:    CategoryAdmin,
			RiskLevel:   RiskHigh,
		},
		"MANAGE_SCHOOLS": {
			ID:          "MANAGE_SCHOOLS",
			Name:        "管理学校",
			Description: "管理学校信息和配置",
			Category:    CategoryAdmin,
			RiskLevel:   RiskHigh,
		},
		"VIEW_ANALYTICS": {
			ID:          "VIEW_ANALYTICS",
			Name:        "查看分析报告",
			Description: "查看平台运营数据和分析",
			Category:    CategoryAdmin,
			RiskLevel:   RiskMedium,
		},
		"AUDIT_LOGS": {
			ID:          "AUDIT_LOGS",
			Name:        "审计日志",
			Description: "查看系统操作审计日志",
			Category:    CategoryAdmin,
			RiskLevel:   RiskHigh,
		},

		// 系统权限模块
		"MANAGE_SYSTEM_SETTINGS": {
			ID:          "MANAGE_SYSTEM_SETTINGS",
			Name:        "管理系统设置",
			Description: "修改系统配置和参数",
			Category:    CategorySystem,
			RiskLevel:   RiskCritical,
			IsSystemCore: true,
		},
		"MANAGE_PERMISSIONS": {
			ID:          "MANAGE_PERMISSIONS",
			Name:        "管理权限",
			Description: "动态配置用户角色权限",
			Category:    CategorySystem,
			RiskLevel:   RiskCritical,
			IsSystemCore: true,
		},
		"SYSTEM_ADMIN": {
			ID:          "SYSTEM_ADMIN",
			Name:        "系统管理",
			Description: "最高级别的系统管理权限",
			Category:    CategorySystem,
			RiskLevel:   RiskCritical,
			IsSystemCore: true,
		},
		"DATABASE_ACCESS": {
			ID:          "DATABASE_ACCESS",
			Name:        "数据库访问",
			Description: "直接访问数据库的权限",
			Category:    CategorySystem,
			RiskLevel:   RiskCritical,
			IsSystemCore: true,
		},
		"API_ADMIN": {
			ID:          "API_ADMIN",
			Name:        "API管理",
			Description: "管理API接口和访问控制",
			Category:    CategorySystem,
			RiskLevel:   RiskCritical,
			IsSystemCore: true,
		},
	}
}

// GetModulesByCategory 按类别获取权限模块
func GetModulesByCategory() map[PermissionCategory][]*PermissionModule {
	modules := GetPermissionModules()
	result := make(map[PermissionCategory][]*PermissionModule)
	
	for _, module := range modules {
		result[module.Category] = append(result[module.Category], module)
	}
	
	return result
}

// ValidatePermissionDependencies 验证权限依赖关系
func ValidatePermissionDependencies(permissions []string) error {
	modules := GetPermissionModules()
	
	for _, permission := range permissions {
		module := modules[permission]
		if module == nil {
			continue
		}
		
		// 检查依赖
		for _, dep := range module.Dependencies {
			found := false
			for _, p := range permissions {
				if p == dep {
					found = true
					break
				}
			}
			if !found {
				return &PermissionError{
					Message:    "Missing dependency: " + dep + " for permission: " + permission,
					Permission: permission,
					Code:       "DEPENDENCY_MISSING",
				}
			}
		}
		
		// 检查冲突
		for _, conflict := range module.Conflicts {
			for _, p := range permissions {
				if p == conflict {
					return &PermissionError{
						Message:    "Permission conflict: " + permission + " conflicts with " + conflict,
						Permission: permission,
						Code:       "PERMISSION_CONFLICT",
					}
				}
			}
		}
	}
	
	return nil
}

// CalculatePermissionRiskLevel 计算权限组合的风险等级
func CalculatePermissionRiskLevel(permissions []string) RiskLevel {
	modules := GetPermissionModules()
	riskScore := 0
	
	for _, permission := range permissions {
		module := modules[permission]
		if module == nil {
			continue
		}
		
		switch module.RiskLevel {
		case RiskLow:
			riskScore += 1
		case RiskMedium:
			riskScore += 3
		case RiskHigh:
			riskScore += 7
		case RiskCritical:
			riskScore += 15
		}
	}
	
	if riskScore >= 50 {
		return RiskCritical
	} else if riskScore >= 20 {
		return RiskHigh
	} else if riskScore >= 8 {
		return RiskMedium
	}
	return RiskLow
}