# 后端API重构计划

## 1. API路由对应调整

### 信使系统API重构

**当前API结构：**
```go
// 分散的信使API
/api/v1/courier/tasks
/api/v1/courier/scan/:code
/api/v1/courier/growth
/api/v1/courier/points
/api/v1/courier/promotion/apply
/api/v1/courier/level/:level/couriers
/api/v1/courier/management/building
/api/v1/courier/management/zone
/api/v1/courier/management/school
/api/v1/courier/management/city
/api/v1/courier/batch/generate
/api/v1/courier/opcode/applications
```

**新API结构：**
```go
// 整合后的信使API
/api/v1/courier/dashboard/stats              // 仪表板统计
/api/v1/courier/tasks                        // 任务管理（包含扫描）
/api/v1/courier/tasks/scan                   // 扫码功能（作为子路由）
/api/v1/courier/management/hierarchy         // 层级管理（统一管理各级）
/api/v1/courier/management/opcode            // OP Code管理
/api/v1/courier/management/batch             // 批量操作
/api/v1/courier/management/credits           // 积分管理
/api/v1/courier/profile/info                 // 个人信息
/api/v1/courier/profile/growth               // 成长记录
/api/v1/courier/profile/points               // 积分详情
/api/v1/courier/profile/promotion            // 晋升管理
```

### 商城系统API重构

**当前API结构：**
```go
/api/v1/shop/products
/api/v1/shop/categories
/api/v1/credit-shop/products
/api/v1/credit-shop/exchange
/api/v1/cart/items
/api/v1/checkout/process
/api/v1/orders
```

**新API结构：**
```go
// 统一商城API
/api/v1/shop/products?type=regular|credit    // 统一商品接口
/api/v1/shop/categories                      // 分类
/api/v1/shop/cart                           // 购物车
/api/v1/shop/checkout                       // 结算
/api/v1/shop/orders                         // 订单
/api/v1/shop/credit/exchange                 // 积分兑换
```

### 信件系统API重构

**当前API结构：**
```go
/api/v1/letters/write
/api/v1/letters/send
/api/v1/deliver/options
/api/v1/mailbox/letters
/api/v1/letters/read/:code
```

**新API结构：**
```go
// 统一信件API
/api/v1/letters/compose                      // 撰写
/api/v1/letters/drafts                       // 草稿
/api/v1/letters/send                         // 发送
/api/v1/letters/inbox                        // 收件箱
/api/v1/letters/sent                         // 已发送
/api/v1/letters/read/:code                   // 阅读
/api/v1/letters/tracking/:id                 // 追踪
```

## 2. 后端Handler重构

### CourierHandler整合

```go
// 当前：多个分散的handler
type CourierHandler struct {}
type CourierGrowthHandler struct {}
type CourierPromotionHandler struct {}
type CourierManagementHandler struct {}

// 新结构：模块化handler
type CourierDashboardHandler struct {}     // 仪表板
type CourierTaskHandler struct {}          // 任务管理
type CourierManagementHandler struct {}    // 管理功能
type CourierProfileHandler struct {}       // 个人中心
```

### 路由注册优化

```go
// backend/main.go 中的路由注册优化

// 信使路由组
courier := protected.Group("/courier")
{
    // 仪表板
    dashboard := courier.Group("/dashboard")
    {
        dashboard.GET("/stats", courierDashboardHandler.GetStats)
        dashboard.GET("/summary", courierDashboardHandler.GetSummary)
    }
    
    // 任务管理
    tasks := courier.Group("/tasks")
    {
        tasks.GET("", courierTaskHandler.GetTasks)
        tasks.POST("/scan", courierTaskHandler.ScanCode)
        tasks.PUT("/:id/status", courierTaskHandler.UpdateStatus)
    }
    
    // 管理中心
    management := courier.Group("/management")
    {
        // 层级管理
        management.GET("/hierarchy", courierManagementHandler.GetHierarchy)
        management.POST("/hierarchy/assign", courierManagementHandler.AssignCourier)
        
        // OP Code管理
        management.GET("/opcode/applications", courierManagementHandler.GetOPCodeApplications)
        management.POST("/opcode/review", courierManagementHandler.ReviewOPCode)
        
        // 批量操作
        management.POST("/batch/generate", courierManagementHandler.BatchGenerate)
        management.POST("/batch/assign", courierManagementHandler.BatchAssign)
    }
    
    // 个人中心
    profile := courier.Group("/profile")
    {
        profile.GET("/info", courierProfileHandler.GetInfo)
        profile.GET("/growth", courierProfileHandler.GetGrowth)
        profile.GET("/points", courierProfileHandler.GetPoints)
        profile.POST("/promotion/apply", courierProfileHandler.ApplyPromotion)
    }
}
```

## 3. Service层重构

### 整合Service

```go
// 当前：分散的service
courierService
courierGrowthService
courierPromotionService
courierManagementService

// 新结构：模块化service
type CourierCoreService struct {}        // 核心功能
type CourierManagementService struct {}  // 管理功能
type CourierProfileService struct {}     // 个人中心
type CourierStatsService struct {}       // 统计分析
```

## 4. 数据库查询优化

### 统一查询接口

```go
// 优化前：多次查询
courier := GetCourierByUserID(userID)
tasks := GetCourierTasks(courierID)
stats := GetCourierStats(courierID)

// 优化后：聚合查询
type CourierDashboard struct {
    Courier *Courier
    Tasks   []Task
    Stats   *Stats
}
dashboard := GetCourierDashboard(userID)
```

## 5. 中间件优化

### 权限中间件整合

```go
// 统一的信使权限中间件
func CourierAuthMiddleware(minLevel int) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUserFromContext(c)
        level := GetCourierLevel(user.Role)
        
        if level < minLevel {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "权限不足",
                "required_level": minLevel,
                "current_level": level,
            })
            c.Abort()
            return
        }
        
        c.Set("courier_level", level)
        c.Next()
    }
}

// 使用示例
management.Use(CourierAuthMiddleware(2)) // L2及以上可访问
```

## 6. 向后兼容策略

### API版本控制

```go
// 保留旧版本API
v1 := r.Group("/api/v1")
{
    // 旧路由 - 标记为deprecated
    v1.GET("/courier/scan/:code", DeprecatedMiddleware(), oldHandler)
}

// 新版本API
v2 := r.Group("/api/v2")
{
    // 新路由
    v2.POST("/courier/tasks/scan", newHandler)
}
```

### 重定向处理

```go
// 自动重定向旧API到新API
func APIRedirectMiddleware() gin.HandlerFunc {
    redirectMap := map[string]string{
        "/api/v1/mailbox/letters": "/api/v1/letters/inbox",
        "/api/v1/deliver/options": "/api/v1/letters/send/options",
        // ... 更多映射
    }
    
    return func(c *gin.Context) {
        if newPath, exists := redirectMap[c.Request.URL.Path]; exists {
            c.Redirect(http.StatusMovedPermanently, newPath)
            return
        }
        c.Next()
    }
}
```

## 7. 测试策略

1. 单元测试：为每个新handler编写测试
2. 集成测试：测试完整的API流程
3. 兼容性测试：确保旧API仍可用
4. 性能测试：验证优化效果

## 8. 实施步骤

1. **Phase 1**: 创建新的handler和service结构
2. **Phase 2**: 实现新API端点
3. **Phase 3**: 添加向后兼容层
4. **Phase 4**: 迁移前端调用
5. **Phase 5**: 监控和优化
6. **Phase 6**: 废弃旧API（3个月后）