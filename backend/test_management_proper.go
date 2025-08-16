package main

import (
    "fmt"
    "openpenpal-backend/internal/models"
)

func main() {
    superAdmin := &models.User{ID: "admin-1", Role: models.RoleSuperAdmin}
    platformAdmin := &models.User{ID: "platform-1", Role: models.RolePlatformAdmin}
    courier4 := &models.User{ID: "courier4-1", Role: models.RoleCourierLevel4}
    courier3 := &models.User{ID: "courier3-1", Role: models.RoleCourierLevel3}
    regularUser := &models.User{ID: "user-1", Role: models.RoleUser}
    
    fmt.Println("Role Hierarchy Levels:")
    fmt.Printf("SuperAdmin (Level %d)\n", models.RoleHierarchy[models.RoleSuperAdmin])
    fmt.Printf("PlatformAdmin (Level %d)\n", models.RoleHierarchy[models.RolePlatformAdmin])
    fmt.Printf("CourierLevel4 (Level %d)\n", models.RoleHierarchy[models.RoleCourierLevel4])
    fmt.Printf("CourierLevel3 (Level %d)\n", models.RoleHierarchy[models.RoleCourierLevel3])
    fmt.Printf("RegularUser (Level %d)\n", models.RoleHierarchy[models.RoleUser])
    
    fmt.Println("\nUser Management Test Results:")
    fmt.Printf("SuperAdmin can manage PlatformAdmin: %v\n", superAdmin.CanManageUser(platformAdmin))
    fmt.Printf("SuperAdmin can manage CourierL4: %v\n", superAdmin.CanManageUser(courier4))
    fmt.Printf("PlatformAdmin can manage CourierL4: %v\n", platformAdmin.CanManageUser(courier4))
    fmt.Printf("PlatformAdmin can manage SuperAdmin: %v\n", platformAdmin.CanManageUser(superAdmin))
    fmt.Printf("CourierL4 can manage CourierL3: %v\n", courier4.CanManageUser(courier3))
    fmt.Printf("CourierL3 can manage RegularUser: %v\n", courier3.CanManageUser(regularUser))
    fmt.Printf("RegularUser can manage SuperAdmin: %v\n", regularUser.CanManageUser(superAdmin))
    
    // Test self-management prevention
    fmt.Printf("\nSelf-Management Prevention:")
    fmt.Printf("\nSuperAdmin can manage self: %v\n", superAdmin.CanManageUser(superAdmin))
}
