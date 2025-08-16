package main

import (
    "fmt"
    "openpenpal-backend/internal/models"
)

func main() {
    superAdmin := &models.User{Role: models.RoleSuperAdmin}
    platformAdmin := &models.User{Role: models.RolePlatformAdmin}
    courier4 := &models.User{Role: models.RoleCourierLevel4}
    regularUser := &models.User{Role: models.RoleUser}
    
    fmt.Println("User Management Test Results:")
    fmt.Println("SuperAdmin can manage PlatformAdmin:", superAdmin.CanManageUser(platformAdmin))
    fmt.Println("SuperAdmin can manage CourierL4:", superAdmin.CanManageUser(courier4))
    fmt.Println("PlatformAdmin can manage CourierL4:", platformAdmin.CanManageUser(courier4))
    fmt.Println("PlatformAdmin can manage SuperAdmin:", platformAdmin.CanManageUser(superAdmin))
    fmt.Println("CourierL4 can manage RegularUser:", courier4.CanManageUser(regularUser))
    fmt.Println("RegularUser can manage SuperAdmin:", regularUser.CanManageUser(superAdmin))
}
