# 商城管理后台API文档

## 概述

OpenPenPal商城管理后台是基于B2B2C模式的现代化电商管理系统。系统支持平台管理和商户管理双重角色，提供完整的权限控制、商品管理、订单处理等功能。

## 🏗️ 系统架构

### 角色权限架构
```
平台层级
├── 平台超级管理员 (PLATFORM_SUPER_ADMIN)
│   └── 完全权限：系统管理、商户管理、数据统计
├── 平台管理员 (PLATFORM_ADMIN) 
│   └── 平台权限：商户管理、分类管理、数据查看
└── 客服/财务等 (CUSTOMER_SERVICE/FINANCE_STAFF)
    └── 专项权限：订单处理、财务管理

商户层级  
├── 商城管理员 (SHOP_ADMIN)
│   └── 店铺完全权限：商品、订单、库存、营销
├── 商城运营 (SHOP_OPERATOR)
│   └── 运营权限：商品管理、订单处理
└── 其他角色
    └── 专项权限：客服、财务等
```

### 菜单权限体系
```
平台管理 (/platform)
├── 系统管理 (platform:system)
│   ├── 用户管理 (platform:system:user)
│   ├── 角色管理 (platform:system:role)
│   └── 菜单管理 (platform:system:menu)
├── 商户管理 (platform:shop)
├── 分类管理 (platform:category)
└── 数据统计 (platform:dashboard)

商城管理 (/shop)
├── 商品管理 (shop:product)
│   ├── SPU管理 (shop:product:spu)
│   └── SKU管理 (shop:product:sku)
├── 订单管理 (shop:order)
├── 库存管理 (shop:inventory)
└── 营销工具 (shop:marketing)
```

## 🔐 RBAC权限管理API

### 基础端点: `/api/v1/rbac`

#### 1. 菜单管理

##### 创建菜单
```http
POST /api/v1/rbac/menus
Content-Type: application/json
Authorization: Bearer {token}

{
    "parent_id": 1,
    "menu_name": "商品管理",
    "menu_code": "shop:product",
    "menu_type": 1,
    "biz_type": 2,
    "path": "/shop/products",
    "component": "shop/products/index",
    "icon": "product",
    "order_num": 1,
    "permission": "shop:product:list",
    "http_method": "GET",
    "api_url": "/api/spu"
}
```

**响应:**
```json
{
    "code": 0,
    "msg": "菜单创建成功",
    "data": {
        "menu_id": 15,
        "menu_name": "商品管理",
        "menu_code": "shop:product",
        "menu_type": 1,
        "biz_type": 2,
        "path": "/shop/products",
        "component": "shop/products/index",
        "icon": "product",
        "order_num": 1,
        "permission": "shop:product:list",
        "status": 1,
        "created_at": "2024-01-20T10:00:00Z"
    }
}
```

##### 获取菜单树
```http
GET /api/v1/rbac/menus/tree?biz_type=2&include_buttons=true
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "tree": [
            {
                "menu_id": 10,
                "menu_name": "商城管理",
                "menu_code": "shop",
                "biz_type": 2,
                "path": "/shop",
                "icon": "shop-manage",
                "children": [
                    {
                        "menu_id": 11,
                        "menu_name": "商品管理", 
                        "menu_code": "shop:product",
                        "path": "/shop/products",
                        "permission": "shop:product:list"
                    }
                ]
            }
        ],
        "biz_type": 2,
        "include_buttons": true
    }
}
```

##### 获取用户菜单
```http
GET /api/v1/rbac/users/{user_id}/menus?biz_type=2
```

#### 2. 角色管理

##### 创建角色
```http
POST /api/v1/rbac/roles
Content-Type: application/json

{
    "role_name": "店铺管理员",
    "role_code": "SHOP_MANAGER",
    "role_desc": "店铺管理员角色",
    "biz_type": 2,
    "is_admin": false,
    "status": 1
}
```

##### 分配角色权限
```http
POST /api/v1/rbac/roles/{role_id}/menus
Content-Type: application/json

{
    "menu_ids": [10, 11, 12, 13]
}
```

#### 3. 用户管理

##### 创建用户
```http
POST /api/v1/rbac/users
Content-Type: application/json

{
    "user_id": "USR001",
    "username": "shop_admin",
    "password": "password123",
    "email": "admin@shop.com",
    "real_name": "张店长",
    "status": 1,
    "user_type": 1
}
```

##### 分配用户角色
```http
POST /api/v1/rbac/users/{user_id}/roles
Content-Type: application/json

{
    "role_ids": [3, 4]
}
```

#### 4. 权限验证

##### 检查权限
```http
GET /api/v1/rbac/permissions/check?permission=shop:product:list&http_method=GET
```

##### 检查角色
```http
GET /api/v1/rbac/roles/check?role_code=SHOP_ADMIN
```

##### 检查管理员
```http
GET /api/v1/rbac/admin/check
```

## 📦 商品管理API

### 分类管理: `/api/v1/categories`

#### 创建分类
```http
POST /api/v1/categories
Content-Type: application/json

{
    "name": "文具用品",
    "parent_id": null,
    "description": "各类文具用品",
    "icon": "stationery",
    "is_active": true,
    "attribute_template": {
        "attributes": [
            {
                "name": "材质",
                "type": "basic",
                "required": true,
                "options": ["塑料", "金属", "木质"]
            },
            {
                "name": "颜色",
                "type": "sale", 
                "required": true,
                "options": ["红色", "蓝色", "黑色"]
            }
        ]
    }
}
```

#### 获取分类树
```http
GET /api/v1/categories/tree/full?include_inactive=false
```

### 商品属性: `/api/v1/product-attributes`

#### 创建商品属性
```http
POST /api/v1/product-attributes/spu/{spu_id}
Content-Type: application/json

[
    {
        "name": "材质",
        "value": "优质牛皮纸",
        "type": "basic",
        "required": true
    },
    {
        "name": "颜色",
        "value": "红色,蓝色,黑色",
        "type": "sale",
        "required": true
    }
]
```

#### 生成SKU组合
```http
GET /api/v1/product-attributes/spu/{spu_id}/sku-combinations
```

#### 自动生成SKU
```http
POST /api/v1/product-attributes/spu/{spu_id}/generate-skus
Content-Type: application/json

{
    "base_price": 25.0,
    "price_rules": {
        "颜色:红色": 2.0,
        "尺寸:A4": 3.0
    },
    "stock_quantity": 100
}
```

## 📊 操作日志API

### 获取操作日志
```http
GET /api/v1/rbac/logs/operations?user_id={user_id}&start_time=2024-01-01T00:00:00Z&page=1&size=20
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "logs": [
            {
                "oper_id": 1,
                "title": "商品管理",
                "business_type": 1,
                "method": "create_spu",
                "request_method": "POST",
                "oper_name": "admin",
                "oper_url": "/api/spu",
                "oper_ip": "192.168.1.1",
                "status": 0,
                "oper_time": "2024-01-20T10:00:00Z"
            }
        ],
        "pagination": {
            "page": 1,
            "size": 20,
            "total": 100,
            "pages": 5
        }
    }
}
```

## 👥 在线用户管理

### 获取在线用户
```http
GET /api/v1/rbac/online/users
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "online_count": 5,
        "users": [
            {
                "session_id": "session123",
                "user_id": "USR001",
                "login_name": "admin",
                "real_name": "系统管理员",
                "ipaddr": "192.168.1.1",
                "login_location": "北京",
                "browser": "Chrome",
                "os": "Windows 10",
                "status": "on_line",
                "start_timestamp": "2024-01-20T09:00:00Z",
                "last_access_time": "2024-01-20T10:30:00Z"
            }
        ]
    }
}
```

## 📈 统计信息API

### RBAC统计
```http
GET /api/v1/rbac/statistics
```

**响应:**
```json
{
    "code": 0,
    "msg": "success", 
    "data": {
        "user_total": 150,
        "user_active": 145,
        "role_total": 8,
        "role_active": 6,
        "menu_total": 25,
        "menu_active": 23,
        "online_users": 12
    }
}
```

## 🚀 系统初始化API

### 初始化RBAC数据
```http
POST /api/v1/rbac/init
```

**响应:**
```json
{
    "code": 0,
    "msg": "RBAC初始化数据创建成功",
    "data": {
        "roles_created": 6,
        "menus_created": 20,
        "roles": [
            {
                "role_id": 1,
                "role_name": "平台超级管理员",
                "role_code": "PLATFORM_SUPER_ADMIN"
            }
        ]
    }
}
```

## 💰 价格管理API

### 基础端点: `/api/v1/pricing`

#### 1. 价格策略管理

##### 创建价格策略
```http
POST /api/v1/pricing/policies
Content-Type: application/json
Authorization: Bearer {token}

{
    "policy_name": "会员专享价格",
    "policy_code": "VIP_PRICING",
    "description": "VIP会员专享优惠价格",
    "price_type": 2,
    "is_active": true,
    "priority": 200,
    "apply_to_all": true,
    "member_level_required": "VIP",
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-12-31T23:59:59Z"
}
```

**响应:**
```json
{
    "code": 0,
    "msg": "价格策略创建成功",
    "data": {
        "policy_id": 1,
        "policy_name": "会员专享价格",
        "policy_code": "VIP_PRICING",
        "price_type": 2,
        "is_active": true,
        "priority": 200,
        "created_at": "2024-01-20T10:00:00Z"
    }
}
```

##### 获取价格策略列表
```http
GET /api/v1/pricing/policies?price_type=2&is_active=true&page=1&size=20
```

#### 2. SKU价格管理

##### 设置SKU价格
```http
POST /api/v1/pricing/sku-prices
Content-Type: application/json

{
    "sku_id": "SKU001",
    "policy_id": 1,
    "original_price": 99.90,
    "current_price": 89.90,
    "cost_price": 45.00,
    "min_price": 60.00,
    "max_price": 120.00,
    "vip_price": 79.90,
    "member_price": 84.90,
    "bulk_config": "{\"tiers\":[{\"min_quantity\":10,\"discount\":5}]}"
}
```

##### 批量设置SKU价格
```http
POST /api/v1/pricing/sku-prices/batch
Content-Type: application/json

{
    "prices": [
        {
            "sku_id": "SKU001",
            "policy_id": 1,
            "original_price": 99.90,
            "current_price": 89.90
        },
        {
            "sku_id": "SKU002",
            "policy_id": 1,
            "original_price": 199.90,
            "current_price": 179.90
        }
    ]
}
```

#### 3. 价格计算引擎

##### 单个商品价格计算
```http
POST /api/v1/pricing/calculate
Content-Type: application/json

{
    "sku_id": "SKU001",
    "quantity": 5,
    "user_id": "USER123",
    "member_level": "VIP"
}
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "sku_id": "SKU001",
        "quantity": 5,
        "original_price": 99.90,
        "base_price": 89.90,
        "final_price": 79.90,
        "total_amount": 399.50,
        "discount_amount": 50.00,
        "applied_promotions": [
            {
                "type": "member",
                "name": "VIP会员价",
                "discount_amount": 50.00
            }
        ]
    }
}
```

##### 批量价格计算
```http
POST /api/v1/pricing/calculate/batch
Content-Type: application/json

[
    {
        "sku_id": "SKU001",
        "quantity": 2,
        "member_level": "VIP"
    },
    {
        "sku_id": "SKU002",
        "quantity": 1,
        "member_level": "VIP"
    }
]
```

#### 4. 促销活动管理

##### 创建促销活动
```http
POST /api/v1/pricing/promotions
Content-Type: application/json

{
    "activity_name": "双十一狂欢节",
    "activity_code": "DOUBLE_ELEVEN_2024",
    "activity_type": "festival",
    "description": "双十一特惠，全场5折起",
    "start_time": "2024-11-11T00:00:00Z",
    "end_time": "2024-11-11T23:59:59Z",
    "discount_config": "{\"type\":\"percentage\",\"value\":20,\"max_discount\":100}",
    "target_config": "{\"apply_to_all\":true}",
    "condition_config": "{\"min_amount\":100}",
    "max_participants": 10000,
    "priority": 500
}
```

##### 获取促销活动列表
```http
GET /api/v1/pricing/promotions?status=active&page=1&size=20
```

#### 5. 价格历史查询

##### 获取SKU价格历史
```http
GET /api/v1/pricing/history/SKU001?start_time=2024-01-01T00:00:00Z&page=1&size=20
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "sku_id": "SKU001",
        "histories": [
            {
                "history_id": 1,
                "price_type": "current_price",
                "old_price": 99.90,
                "new_price": 89.90,
                "change_reason": "促销调价",
                "effective_time": "2024-01-20T10:00:00Z",
                "operator_id": "ADM001"
            }
        ],
        "pagination": {
            "page": 1,
            "size": 20,
            "total": 10,
            "pages": 1
        }
    }
}
```

#### 6. 价格管理工具

##### 获取价格建议
```http
GET /api/v1/pricing/tools/price-suggestion/SKU001
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "sku_id": "SKU001",
        "suggestions": [
            {
                "strategy": "market_based",
                "suggested_price": 99.9,
                "reason": "基于市场价格分析",
                "confidence": 0.85
            },
            {
                "strategy": "cost_plus",
                "suggested_price": 89.9,
                "reason": "成本加成定价",
                "confidence": 0.92
            }
        ]
    }
}
```

##### 检查价格竞争力
```http
POST /api/v1/pricing/tools/price-check?sku_id=SKU001&proposed_price=85.00
```

## 🔨 批量操作API

### 批量删除
```http
POST /api/batch/letters/delete
Content-Type: application/json

{
    "target_ids": ["LETTER001", "LETTER002"],
    "soft_delete": true,
    "delete_reason": "批量清理"
}
```

### 批量状态更新
```http
POST /api/batch/letters/status
Content-Type: application/json

{
    "target_ids": ["LETTER001", "LETTER002"], 
    "new_status": "generated",
    "reason": "批量生成"
}
```

## 🛡️ 认证和授权

### JWT Token格式
```javascript
// Header
{
    "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

// Token Payload
{
    "user_id": "USR001",
    "username": "admin", 
    "roles": ["PLATFORM_SUPER_ADMIN"],
    "permissions": ["platform:system:user:list"],
    "exp": 1640995200
}
```

### 权限检查中间件
前端路由守卫示例：
```javascript
// router/permission.js
import { useUserStore } from '@/stores/user'

export async function checkPermission(to, from, next) {
    const userStore = useUserStore()
    
    // 检查登录状态
    if (!userStore.token) {
        return next('/login')
    }
    
    // 检查页面权限
    const permission = to.meta?.permission
    if (permission && !userStore.hasPermission(permission)) {
        return next('/403')
    }
    
    next()
}
```

## 📋 前端接口调用示例

### Vue 3 Composition API
```typescript
// composables/useRBAC.ts
import { ref, reactive } from 'vue'
import { rbacAPI } from '@/api/rbac'

export function useRBAC() {
    const loading = ref(false)
    const menuTree = ref([])
    const userRoles = ref([])
    
    // 获取菜单树
    const getMenuTree = async (bizType: number) => {
        loading.value = true
        try {
            const response = await rbacAPI.getMenuTree({ biz_type: bizType })
            menuTree.value = response.data.tree
        } finally {
            loading.value = false
        }
    }
    
    // 检查权限
    const checkPermission = async (permission: string, method: string = 'GET') => {
        const response = await rbacAPI.checkPermission({ permission, http_method: method })
        return response.data.has_permission
    }
    
    return {
        loading,
        menuTree,
        userRoles,
        getMenuTree,
        checkPermission
    }
}
```

### API服务封装
```typescript
// api/rbac.ts
import request from '@/utils/request'

export const rbacAPI = {
    // 菜单管理
    createMenu: (data: MenuCreateRequest) => 
        request.post('/api/v1/rbac/menus', data),
    
    getMenuTree: (params: { biz_type?: number, include_buttons?: boolean }) =>
        request.get('/api/v1/rbac/menus/tree', { params }),
    
    getUserMenus: (userId: string, bizType: number) =>
        request.get(`/api/v1/rbac/users/${userId}/menus`, { 
            params: { biz_type: bizType } 
        }),
    
    // 角色管理
    createRole: (data: RoleCreateRequest) =>
        request.post('/api/v1/rbac/roles', data),
    
    assignRoleMenus: (roleId: number, menuIds: number[]) =>
        request.post(`/api/v1/rbac/roles/${roleId}/menus`, { menu_ids: menuIds }),
    
    // 权限验证
    checkPermission: (params: { permission: string, http_method: string }) =>
        request.get('/api/v1/rbac/permissions/check', { params }),
    
    checkRole: (roleCode: string) =>
        request.get('/api/v1/rbac/roles/check', { params: { role_code: roleCode } })
}
```

## 🚨 错误处理

### 常见错误码
| 错误码 | 描述 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数无效 | 检查请求格式和必填字段 |
| 401 | 未授权访问 | 提供有效的JWT Token |
| 403 | 权限不足 | 检查用户角色和权限配置 |
| 404 | 资源不存在 | 确认资源ID正确性 |
| 409 | 资源冲突 | 检查唯一约束（如用户名重复） |
| 422 | 业务逻辑错误 | 查看错误消息详情 |
| 500 | 服务器内部错误 | 联系技术支持 |

### 错误响应格式
```json
{
    "code": 400,
    "msg": "菜单编码已存在",
    "data": null,
    "timestamp": "2024-01-20T10:00:00Z"
}
```

## 🔧 开发工具

### 数据库初始化
```bash
# 初始化RBAC数据
python migrations/init_rbac_data.py

# 迁移到SPU+SKU模型
python migrations/migrate_to_spu_sku.py
```

### API测试
```bash
# 健康检查
curl -X GET "http://localhost:8001/health"

# 获取菜单树
curl -X GET "http://localhost:8001/api/v1/rbac/menus/tree?biz_type=2" \
     -H "Authorization: Bearer {token}"

# 创建角色
curl -X POST "http://localhost:8001/api/v1/rbac/roles" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer {token}" \
     -d '{"role_name":"测试角色","role_code":"TEST_ROLE"}'
```

## 📚 相关文档

- [商品管理API文档](./PRODUCT_MANAGEMENT_API.md)
- [批量操作API文档](./BATCH_OPERATIONS_API.md) 
- [SPU+SKU迁移指南](./SPU_SKU_MIGRATION_GUIDE.md)
- [前端协同开发文档](../multi-agent/mall_admin_collaboration.md)

---

## 💡 最佳实践

### 1. 权限设计原则
- **最小权限原则**: 用户只获得执行任务所需的最小权限
- **角色继承**: 合理设计角色层级，避免权限冗余
- **动态权限**: 支持运行时权限变更，无需重启服务

### 2. API使用建议
- **权限预检**: 前端展示前先检查权限，避免无效操作
- **批量操作**: 大量数据操作使用批量接口，提高效率
- **缓存策略**: 菜单和权限信息适度缓存，减少频繁请求

### 3. 安全注意事项
- **Token管理**: 及时刷新Token，设置合理过期时间
- **敏感操作**: 重要操作需要二次确认和操作日志
- **数据加密**: 敏感数据传输使用HTTPS加密

*本文档随系统更新持续维护，如有疑问请联系开发团队。*