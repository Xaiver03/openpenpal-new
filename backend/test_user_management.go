package main

import (
    "fmt"
    "openpenpal-backend/internal/models"
)

func main() {
    fmt.Println("User Management Hierarchy Test:")
    fmt.Println("==============================")
    
    superAdmin := &models.User{Role: models.RoleSuperAdmin}
    platformAdmin := &models.User{Role: models.RolePlatformAdmin}
    courier4 := &models.User{Role: models.RoleCourierLevel4}
    courier3 := &models.User{Role: models.RoleCourierLevel3}
    courier1 := &models.User{Role: models.RoleCourierLevel1}
    regularUser := &models.User{Role: models.RoleUser}
    
    testCases := []struct{
        manager *models.User
        target *models.User
        description string
    }{
        {superAdmin, platformAdmin, "SuperAdmin manages PlatformAdmin"},
        {superAdmin, courier4, "SuperAdmin manages CourierL4"},
        {platformAdmin, courier4, "PlatformAdmin manages CourierL4"},
        {platformAdmin, superAdmin, "PlatformAdmin manages SuperAdmin"},
        {courier4, courier3, "CourierL4 manages CourierL3"},
        {courier3, courier1, "CourierL3 manages CourierL1"},
        {courier1, regularUser, "CourierL1 manages RegularUser"},
        {regularUser, superAdmin, "RegularUser manages SuperAdmin"},
    }
    
    for _, tc := range testCases {
        result := tc.manager.CanManageUser(tc.target)
        status := "YES"
        if \!result {
            status = "NO"
        }
        fmt.Printf("[%s] %s\n", status, tc.description)
    }
}
