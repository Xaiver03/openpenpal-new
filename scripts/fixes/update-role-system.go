package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	// 读取原始user.go文件
	userFilePath := "backend/internal/models/user.go"
	content, err := ioutil.ReadFile(userFilePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	originalContent := string(content)
	
	// 备份原文件
	backupPath := userFilePath + ".backup"
	err = ioutil.WriteFile(backupPath, content, 0644)
	if err != nil {
		fmt.Printf("Error creating backup: %v\n", err)
		return
	}
	fmt.Printf("✅ Created backup at: %s\n", backupPath)

	// 定义新的角色常量部分
	newRoleConstants := `// UserRole 用户角色枚举
type UserRole string

const (
	// 基础角色
	RoleUser UserRole = "user" // 普通用户
	
	// 四级信使体系
	RoleCourierLevel1 UserRole = "courier_level1" // 一级信使（基础投递信使）
	RoleCourierLevel2 UserRole = "courier_level2" // 二级信使（片区协调员）
	RoleCourierLevel3 UserRole = "courier_level3" // 三级信使（校区负责人）
	RoleCourierLevel4 UserRole = "courier_level4" // 四级信使（城市负责人）
	
	// 管理角色
	RolePlatformAdmin UserRole = "platform_admin" // 平台管理员
	RoleSuperAdmin    UserRole = "super_admin"    // 超级管理员
)`

	// 定义新的角色层级
	newRoleHierarchy := `// RoleHierarchy 角色层级（数字越大权限越高）
var RoleHierarchy = map[UserRole]int{
	RoleUser:          1,
	RoleCourierLevel1: 2,
	RoleCourierLevel2: 3,
	RoleCourierLevel3: 4,
	RoleCourierLevel4: 5,
	RolePlatformAdmin: 6,
	RoleSuperAdmin:    7,
}`

	// 查找并替换角色常量部分
	startMarker := "// UserRole 用户角色枚举"
	endMarker := "// String 返回角色字符串"
	
	startIndex := strings.Index(originalContent, startMarker)
	endIndex := strings.Index(originalContent, endMarker)
	
	if startIndex == -1 || endIndex == -1 {
		fmt.Println("❌ Could not find role constants section")
		return
	}

	// 构建新内容
	newContent := originalContent[:startIndex] + 
		newRoleConstants + "\n\n" +
		originalContent[endIndex:]

	// 替换RoleHierarchy
	hierarchyStart := strings.Index(newContent, "// RoleHierarchy 角色层级")
	hierarchyEnd := strings.Index(newContent[hierarchyStart:], "}")
	if hierarchyStart != -1 && hierarchyEnd != -1 {
		hierarchyEnd += hierarchyStart + 1
		newContent = newContent[:hierarchyStart] + 
			newRoleHierarchy + 
			newContent[hierarchyEnd:]
	}

	// 写入更新后的内容
	err = ioutil.WriteFile(userFilePath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing updated file: %v\n", err)
		return
	}

	fmt.Println("✅ Successfully updated role system!")
	fmt.Println("\n📋 Updated roles:")
	fmt.Println("  - user (普通用户)")
	fmt.Println("  - courier_level1 (一级信使)")
	fmt.Println("  - courier_level2 (二级信使)")
	fmt.Println("  - courier_level3 (三级信使)")
	fmt.Println("  - courier_level4 (四级信使)")
	fmt.Println("  - platform_admin (平台管理员)")
	fmt.Println("  - super_admin (超级管理员)")
	
	fmt.Println("\n❌ Removed redundant roles:")
	fmt.Println("  - courier")
	fmt.Println("  - senior_courier")
	fmt.Println("  - courier_coordinator")
	fmt.Println("  - school_admin")
}