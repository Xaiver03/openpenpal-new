package main

import (
    "fmt"
    "openpenpal-backend/internal/models"
)

func main() {
    // Test role hierarchy
    roles := []models.UserRole{
        models.RoleUser,
        models.RoleCourierLevel1,
        models.RoleCourierLevel2,
        models.RoleCourierLevel3,
        models.RoleCourierLevel4,
        models.RolePlatformAdmin,
        models.RoleSuperAdmin,
    }
    
    fmt.Println("Role Hierarchy Verification:")
    fmt.Println("==========================")
    
    // Test each role's permissions
    for _, role := range roles {
        user := &models.User{Role: role}
        fmt.Printf("\nRole: %s\n", role)
        fmt.Printf("- Can manage users: %v\n", user.HasPermission(models.PermissionManageUsers))
        fmt.Printf("- Can manage couriers: %v\n", user.HasPermission(models.PermissionManageCouriers))
        fmt.Printf("- Can manage platform: %v\n", user.HasPermission(models.PermissionManagePlatform))
        fmt.Printf("- Can deliver letters: %v\n", user.HasPermission(models.PermissionDeliverLetter))
    }
}
