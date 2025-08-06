# 商品管理系统API文档

## 概述

OpenPenPal商品管理系统基于现代SPU+SKU架构设计，提供完整的商品、分类、属性管理功能。系统支持多规格商品、动态属性配置、多级分类管理等高级电商功能。

## 🏗️ 系统架构

### 核心概念

- **SPU (Standard Product Unit)**: 标准商品单元，代表一个商品的抽象概念
- **SKU (Stock Keeping Unit)**: 库存保持单元，代表具体的商品规格和库存
- **属性系统**: 分为销售属性（影响SKU）和基本属性（描述性信息）
- **分类系统**: 支持无限级分类和分类属性模板

### 数据模型关系

```
ProductCategory (分类)
    ├── ProductSPU (商品)
    │   ├── ProductSKU (规格) × N
    │   ├── ProductAttribute (属性) × N
    │   └── StockRecord (库存记录) × N
    └── AttributeTemplate (属性模板)

Order (订单)
    └── OrderItem (订单项) × N
        └── 关联到具体的SKU
```

## 📦 商品分类API

### 基础端点: `/api/v1/categories`

#### 1. 创建分类
```http
POST /api/v1/categories
Content-Type: application/json

{
    "name": "信封",
    "parent_id": null,
    "description": "各种类型的信封产品",
    "icon": "envelope-icon",
    "is_active": true,
    "is_visible": true,
    "sort_order": 1,
    "attribute_template": {
        "attributes": [
            {
                "name": "材质",
                "type": "basic",
                "required": true,
                "options": ["牛皮纸", "珠光纸", "艺术纸"],
                "searchable": true,
                "filterable": true,
                "sort_order": 1
            },
            {
                "name": "颜色",
                "type": "sale",
                "required": true,
                "options": ["白色", "米色", "棕色"],
                "searchable": false,
                "filterable": true,
                "sort_order": 2
            }
        ]
    }
}
```

**响应:**
```json
{
    "code": 0,
    "msg": "分类创建成功",
    "data": {
        "id": "CAT001",
        "name": "信封",
        "level": 0,
        "path": "/CAT001/",
        "created_at": "2024-01-20T10:00:00Z"
    }
}
```

#### 2. 获取分类树
```http
GET /api/v1/categories/tree/full?include_inactive=false
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "tree": [
            {
                "id": "CAT001",
                "name": "信封",
                "level": 0,
                "children": [
                    {
                        "id": "CAT002", 
                        "name": "商务信封",
                        "level": 1,
                        "parent_id": "CAT001"
                    }
                ]
            }
        ],
        "total_nodes": 15
    }
}
```

#### 3. 获取分类路径（面包屑）
```http
GET /api/v1/categories/{category_id}/path
```

#### 4. 搜索分类
```http
GET /api/v1/categories/search?keyword=信封&parent_id=CAT001
```

#### 5. 分类统计信息
```http
GET /api/v1/categories/{category_id}/statistics
```

**响应:**
```json
{
    "code": 0,
    "msg": "success", 
    "data": {
        "category_id": "CAT001",
        "category_name": "信封",
        "statistics": {
            "spu_count": 25,
            "total_spu_including_descendants": 45,
            "total_sales": 1250,
            "direct_children_count": 3,
            "all_descendants_count": 8,
            "level": 0
        }
    }
}
```

## 🏷️ 商品属性API

### 基础端点: `/api/v1/product-attributes`

#### 1. 获取分类属性模板
```http
GET /api/v1/product-attributes/templates/{category_id}
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "category_id": "CAT001",
        "templates": [
            {
                "name": "材质",
                "type": "basic",
                "required": true,
                "options": ["牛皮纸", "珠光纸", "艺术纸"],
                "searchable": true,
                "filterable": true
            }
        ]
    }
}
```

#### 2. 为SPU创建属性
```http
POST /api/v1/product-attributes/spu/{spu_id}
Content-Type: application/json

[
    {
        "name": "材质",
        "value": "牛皮纸",
        "type": "basic",
        "required": true,
        "searchable": true,
        "filterable": true
    },
    {
        "name": "颜色",
        "value": "红色,蓝色,黑色",
        "type": "sale",
        "required": true,
        "filterable": true
    }
]
```

#### 3. 生成SKU组合
```http
GET /api/v1/product-attributes/spu/{spu_id}/sku-combinations
```

**响应:**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "spu_id": "SPU001",
        "combination_count": 9,
        "combinations": [
            {
                "name": "红色-A4",
                "sale_attributes": {
                    "颜色": "红色",
                    "尺寸": "A4"
                }
            },
            {
                "name": "蓝色-A5",
                "sale_attributes": {
                    "颜色": "蓝色", 
                    "尺寸": "A5"
                }
            }
        ]
    }
}
```

#### 4. 自动生成SKU
```http
POST /api/v1/product-attributes/spu/{spu_id}/generate-skus
Content-Type: application/json

{
    "auto_pricing": true,
    "base_price": 20.0,
    "price_rules": {
        "颜色:红色": 5.0,
        "尺寸:A4": 3.0
    },
    "stock_quantity": 100
}
```

#### 5. 批量创建SKU
```http
POST /api/v1/product-attributes/spu/{spu_id}/create-skus
Content-Type: application/json

{
    "skus": [
        {
            "name": "红色-A4",
            "price": 25.0,
            "stock_quantity": 100,
            "sale_attributes": {
                "颜色": "红色",
                "尺寸": "A4"
            }
        }
    ]
}
```

#### 6. 属性筛选搜索
```http
POST /api/v1/product-attributes/search
Content-Type: application/json

{
    "attributes": {
        "颜色": ["红色", "蓝色"],
        "尺寸": ["A4", "A5"]
    },
    "category_id": "CAT001"
}
```

#### 7. 获取筛选选项
```http
GET /api/v1/product-attributes/filter-options?category_id=CAT001
```

## 🛒 完整的商品创建流程

### 1. 创建SPU
```http
POST /api/spu
Content-Type: application/json

{
    "name": "经典牛皮信封",
    "description": "高质量牛皮纸材质，适合商务和个人使用",
    "category_id": "CAT001",
    "brand_id": "BRD001",
    "product_type": "envelope",
    "status": "active",
    "main_image": "https://example.com/envelope.jpg"
}
```

### 2. 添加商品属性
```http
POST /api/v1/product-attributes/spu/SPU001

[
    {
        "name": "材质",
        "value": "牛皮纸",
        "type": "basic"
    },
    {
        "name": "颜色", 
        "value": "红色,蓝色,黑色",
        "type": "sale"
    },
    {
        "name": "尺寸",
        "value": "A4,A5,C6",
        "type": "sale"
    }
]
```

### 3. 生成并创建SKU
```http
POST /api/v1/product-attributes/spu/SPU001/generate-skus

{
    "base_price": 20.0,
    "stock_quantity": 50
}
```

### 4. 验证和调整
```http
GET /api/spu/SPU001/skus
PUT /api/sku/{sku_id}
```

## 🔍 高级功能

### 分类管理

1. **无限级分类**: 支持任意深度的分类层次
2. **分类模板**: 每个分类可配置专属的属性模板
3. **分类统计**: 自动统计商品数量、销量等信息
4. **批量操作**: 支持批量重排序、统计刷新等

### 属性系统

1. **动态属性**: 基于分类模板自动生成属性选项
2. **属性类型**: 
   - **基本属性**: 描述性信息，不影响SKU生成
   - **销售属性**: 影响SKU生成的规格属性
   - **自定义属性**: 特殊用途的自定义属性
3. **属性验证**: 自动验证属性值的合法性
4. **搜索筛选**: 支持基于属性的商品搜索和筛选

### SKU管理

1. **自动生成**: 根据销售属性组合自动生成SKU
2. **批量创建**: 支持批量导入SKU数据
3. **价格策略**: 支持基于属性的动态定价
4. **库存管理**: 每个SKU独立库存管理

## 📊 数据统计API

### 分类统计
```http
GET /api/v1/categories/analytics/popular?limit=10
GET /api/v1/categories/{category_id}/statistics
POST /api/v1/categories/statistics/refresh
```

### 属性统计  
```http
GET /api/v1/product-attributes/statistics?category_id=CAT001
```

## 🔧 管理工具

### 批量操作
```http
POST /api/v1/categories/reorder
POST /api/v1/categories/{category_id}/template
POST /api/v1/product-attributes/spu/{spu_id}/generate-skus
```

### 数据验证
```http
POST /api/v1/product-attributes/validate-sku
GET /api/v1/categories/{category_id}/template
```

## 🚀 使用示例

### JavaScript客户端
```javascript
// 创建信封分类
const createCategory = async () => {
    const response = await fetch('/api/v1/categories', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer your-jwt-token'
        },
        body: JSON.stringify({
            name: '信封',
            description: '各种类型的信封产品',
            attribute_template: {
                attributes: [
                    {
                        name: '材质',
                        type: 'basic',
                        required: true,
                        options: ['牛皮纸', '珠光纸']
                    }
                ]
            }
        })
    });
    return await response.json();
};

// 为商品添加属性并生成SKU
const createProductWithSKUs = async (spuId) => {
    // 1. 添加属性
    await fetch(`/api/v1/product-attributes/spu/${spuId}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer your-jwt-token'
        },
        body: JSON.stringify([
            {
                name: '颜色',
                value: '红色,蓝色,黑色',
                type: 'sale'
            }
        ])
    });
    
    // 2. 生成SKU
    const skuResponse = await fetch(`/api/v1/product-attributes/spu/${spuId}/generate-skus`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer your-jwt-token'
        },
        body: JSON.stringify({
            base_price: 20.0,
            stock_quantity: 100
        })
    });
    
    return await skuResponse.json();
};
```

### Python客户端
```python
import aiohttp
import asyncio

class ProductManagementClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    async def create_category_with_template(self, name: str, template: dict):
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f'{self.base_url}/api/v1/categories',
                json={
                    'name': name,
                    'attribute_template': template
                },
                headers=self.headers
            ) as response:
                return await response.json()
    
    async def generate_skus_for_product(self, spu_id: str, base_price: float):
        async with aiohttp.ClientSession() as session:
            # 生成SKU组合
            async with session.get(
                f'{self.base_url}/api/v1/product-attributes/spu/{spu_id}/sku-combinations',
                headers=self.headers
            ) as response:
                combinations = await response.json()
            
            # 自动生成SKU
            async with session.post(
                f'{self.base_url}/api/v1/product-attributes/spu/{spu_id}/generate-skus',
                json={
                    'base_price': base_price,
                    'stock_quantity': 50
                },
                headers=self.headers
            ) as response:
                return await response.json()

# 使用示例
async def main():
    client = ProductManagementClient('http://localhost:8001', 'your-jwt-token')
    
    # 创建信封分类
    category_result = await client.create_category_with_template(
        '信封',
        {
            'attributes': [
                {
                    'name': '颜色',
                    'type': 'sale',
                    'options': ['红色', '蓝色', '黑色']
                }
            ]
        }
    )
    
    # 为商品生成SKU
    sku_result = await client.generate_skus_for_product('SPU001', 20.0)
    print(f"Generated {sku_result['data']['created_count']} SKUs")

asyncio.run(main())
```

## 📋 最佳实践

### 1. 分类设计
- 保持分类层次简洁（建议不超过4级）
- 每个分类配置合适的属性模板
- 定期更新分类统计信息

### 2. 属性管理
- 销售属性应简洁明了，避免过多组合
- 基本属性可以丰富详细，提供充分的商品信息
- 属性选项要标准化，避免重复和歧义

### 3. SKU生成
- 先设计好销售属性再生成SKU
- 合理设置价格规则和库存
- 生成后及时调整和优化

### 4. 性能优化
- 使用筛选API时合理设置分类范围
- 大量SKU的商品考虑分页加载
- 定期清理无效的属性和分类

## 🔒 权限控制

- 分类管理需要管理员权限
- 商品属性创建需要商品管理权限
- SKU生成需要商品编辑权限
- 统计查询支持只读权限

## 🚨 错误处理

常见错误码：
- 400: 请求参数错误
- 401: 未授权访问
- 403: 权限不足
- 404: 资源不存在
- 409: 资源冲突（如重复名称）
- 422: 业务逻辑错误

---

*本文档涵盖了商品管理系统的核心API，如有疑问请联系开发团队。*