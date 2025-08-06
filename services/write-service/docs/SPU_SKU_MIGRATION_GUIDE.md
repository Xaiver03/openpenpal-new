# SPU+SKU模型迁移指南

## 概述

本指南详细说明如何从传统的单一商品模型迁移到现代的SPU+SKU分离架构。新架构提供了更好的商品管理能力，支持多规格商品、库存管理、价格体系等高级功能。

## 🏗️ 架构变化

### 传统模型 vs SPU+SKU模型

| 特性 | 传统模型 | SPU+SKU模型 |
|------|----------|-------------|
| 商品结构 | 单一商品表 | SPU(商品) + SKU(规格) |
| 规格管理 | 字段存储 | 独立SKU管理 |
| 库存管理 | 简单数量 | 多规格库存+记录 |
| 价格体系 | 单一价格 | 多规格价格+历史 |
| 订单关联 | 商品ID | SKU ID |
| 属性管理 | 混合存储 | 分类属性系统 |

### 新模型优势

- ✅ **多规格支持**: 一个商品可以有多个规格（颜色、尺寸等）
- ✅ **精确库存管理**: 每个SKU独立库存，支持预占、锁定等
- ✅ **灵活价格体系**: 不同规格不同价格，支持价格历史
- ✅ **属性分离**: 销售属性和基本属性分开管理
- ✅ **多级分类**: 支持无限级分类和分类模板
- ✅ **品牌管理**: 独立的品牌体系
- ✅ **订单精确性**: 订单直接关联到具体SKU

## 📦 新模型结构

### 核心表结构

```sql
-- SPU表：标准商品单元
product_spu (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id VARCHAR(20),
    brand_id VARCHAR(20),
    status VARCHAR(20),
    -- ...其他字段
)

-- SKU表：库存保持单元  
product_sku (
    id VARCHAR(20) PRIMARY KEY,
    spu_id VARCHAR(20) REFERENCES product_spu(id),
    sku_code VARCHAR(100) UNIQUE,
    price FLOAT NOT NULL,
    stock_quantity INTEGER,
    sale_attributes JSON,
    -- ...其他字段
)

-- 商品属性表
product_attributes (
    id VARCHAR(20) PRIMARY KEY,
    spu_id VARCHAR(20) REFERENCES product_spu(id),
    attribute_name VARCHAR(100),
    attribute_value TEXT,
    attribute_type VARCHAR(20)
)

-- 分类表
product_categories (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    parent_id VARCHAR(20) REFERENCES product_categories(id),
    path VARCHAR(500),
    level INTEGER
)

-- 品牌表
product_brands (
    id VARCHAR(20) PRIMARY KEY, 
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
)
```

### 订单模型变化

```sql
-- 新订单项表
order_items (
    id VARCHAR(20) PRIMARY KEY,
    order_id VARCHAR(20),
    spu_id VARCHAR(20) REFERENCES product_spu(id),
    sku_id VARCHAR(20) REFERENCES product_sku(id),  -- 直接关联SKU
    spu_name VARCHAR(200),
    sku_name VARCHAR(200),
    sale_attributes JSON,  -- SKU属性快照
    unit_price FLOAT,
    quantity INTEGER,
    -- ...其他字段
)
```

## 🚀 迁移流程

### 第一步：数据备份

⚠️ **重要**：迁移前必须备份数据！

```bash
# 运行备份脚本
python migrations/backup_before_migration.py
```

备份内容：
- 所有原始表的完整副本
- 表结构、索引、约束定义
- 数据导出文件
- 备份验证报告

### 第二步：执行迁移

```bash
# 运行迁移脚本
python migrations/migrate_to_spu_sku.py
```

迁移过程：
1. ✅ 创建新表结构
2. ✅ 迁移分类和品牌数据
3. ✅ 将商品转换为SPU+SKU模型
4. ✅ 更新订单关联关系
5. ✅ 创建库存记录
6. ✅ 验证数据完整性

### 第三步：验证结果

迁移完成后检查：

```sql
-- 检查数据量一致性
SELECT 
    (SELECT COUNT(*) FROM shop_products) as old_products,
    (SELECT COUNT(*) FROM product_spu) as new_spu,
    (SELECT COUNT(*) FROM product_sku) as new_sku;

-- 检查订单项关联
SELECT COUNT(*) FROM order_items WHERE spu_id IS NOT NULL AND sku_id IS NOT NULL;

-- 检查库存记录
SELECT COUNT(*) FROM stock_records WHERE reference_type = 'migration';
```

## 📊 数据映射关系

### 商品数据映射

| 旧字段 (shop_products) | 新字段 | 目标表 | 说明 |
|------------------------|--------|--------|------|
| id | - | mapping | 用于关联映射 |
| name | name | product_spu | SPU名称 |
| description | description | product_spu | 商品描述 |
| category | name | product_categories | 转为分类表 |
| brand | name | product_brands | 转为品牌表 |
| price | price | product_sku | SKU价格 |
| stock_quantity | stock_quantity | product_sku | SKU库存 |
| color, material | sale_attributes | product_sku | 转为JSON属性 |
| tags | tags | product_spu | 转为JSON数组 |

### 订单数据映射

| 旧字段 | 新字段 | 说明 |
|--------|--------|------|
| product_id | spu_id, sku_id | 映射到对应的SPU和SKU |
| product_name | spu_name | SPU名称快照 |
| product_attributes | sale_attributes | SKU属性快照 |

## 🔧 API变化

### 商品查询API

```javascript
// 旧API
GET /api/products/{product_id}

// 新API  
GET /api/spu/{spu_id}              // 获取SPU基础信息
GET /api/spu/{spu_id}/skus         // 获取SPU的所有SKU
GET /api/sku/{sku_id}              // 获取具体SKU信息
```

### 购物车API

```javascript
// 旧API
POST /api/cart/add
{
    "product_id": "PROD001",
    "quantity": 2
}

// 新API
POST /api/cart/add  
{
    "sku_id": "SKU001",      // 直接指定SKU
    "quantity": 2
}
```

### 订单创建API

```javascript
// 旧API
POST /api/orders
{
    "items": [
        {
            "product_id": "PROD001",
            "quantity": 1,
            "unit_price": 99.00
        }
    ]
}

// 新API
POST /api/orders
{
    "items": [
        {
            "sku_id": "SKU001",          // 改为SKU ID
            "quantity": 1,
            "unit_price": 99.00
        }
    ]
}
```

## 🎯 迁移后的新功能

### 1. 多规格商品支持

```python
# 创建一个信封商品，支持多种颜色和尺寸
spu = ProductSPU(
    name="经典牛皮信封",
    category_id="CAT_ENVELOPE"
)

# 创建多个SKU
sku1 = ProductSKU(
    spu_id=spu.id,
    sku_code="ENV_RED_A4",
    name="红色-A4",
    price=25.00,
    sale_attributes={"color": "红色", "size": "A4"}
)

sku2 = ProductSKU(
    spu_id=spu.id, 
    sku_code="ENV_BLUE_A5",
    name="蓝色-A5",
    price=20.00,
    sale_attributes={"color": "蓝色", "size": "A5"}
)
```

### 2. 精确库存管理

```python
# 库存操作自动记录
async def update_stock(sku_id: str, quantity_change: int, reason: str):
    sku = await get_sku(sku_id)
    old_quantity = sku.stock_quantity
    new_quantity = old_quantity + quantity_change
    
    # 更新库存
    sku.stock_quantity = new_quantity
    
    # 记录变更
    stock_record = StockRecord(
        sku_id=sku_id,
        change_type="adjust",
        change_quantity=quantity_change,
        before_quantity=old_quantity,
        after_quantity=new_quantity,
        remark=reason
    )
    
    await save(sku, stock_record)
```

### 3. 动态属性系统

```python
# 基本属性（不影响SKU）
basic_attr = ProductAttribute(
    spu_id=spu.id,
    attribute_name="材质",
    attribute_value="真皮",
    attribute_type="basic"
)

# 销售属性（影响SKU）
sale_attr = ProductAttribute(
    spu_id=spu.id,
    attribute_name="颜色",
    attribute_value="红色,蓝色,黑色",
    attribute_type="sale"
)
```

## 🛠️ 故障排除

### 常见问题

**Q: 迁移后找不到某些商品？**
A: 检查迁移日志，可能是数据不完整。使用映射关系表查找：
```sql
SELECT * FROM product_spu WHERE name LIKE '%关键词%';
```

**Q: 订单显示异常？**
A: 订单项可能没有正确关联SKU：
```sql
SELECT * FROM order_items WHERE sku_id IS NULL;
```

**Q: 库存数据不对？**
A: 检查库存记录：
```sql
SELECT * FROM stock_records WHERE sku_id = 'SKU001' ORDER BY created_at;
```

### 回滚操作

如果迁移出现严重问题，可以回滚：

```bash
# 回滚到旧模型（危险操作！）
python migrations/rollback_spu_sku.py
```

⚠️ **警告**: 回滚会丢失所有新模型的数据改动！

## 📈 性能优化建议

### 1. 索引优化

```sql
-- SKU查询优化
CREATE INDEX idx_sku_spu_status ON product_sku(spu_id, status);
CREATE INDEX idx_sku_price_stock ON product_sku(price, stock_quantity);

-- 订单查询优化  
CREATE INDEX idx_order_item_sku ON order_items(sku_id);
CREATE INDEX idx_order_item_order_sku ON order_items(order_id, sku_id);
```

### 2. 查询优化

```python
# 避免N+1查询，使用预加载
spus = await session.execute(
    select(ProductSPU)
    .options(
        selectinload(ProductSPU.skus),
        selectinload(ProductSPU.category),
        selectinload(ProductSPU.brand)
    )
)
```

### 3. 缓存策略

```python
# 热门商品SPU+SKU数据缓存
@cache_result(ttl=3600)
async def get_spu_with_skus(spu_id: str):
    return await session.execute(
        select(ProductSPU)
        .options(selectinload(ProductSPU.skus))
        .where(ProductSPU.id == spu_id)
    )
```

## 🎓 最佳实践

### 1. 商品创建流程

```python
async def create_product_with_skus(spu_data: dict, sku_list: list):
    # 1. 创建SPU
    spu = ProductSPU(**spu_data)
    session.add(spu)
    await session.flush()
    
    # 2. 创建SKU
    for sku_data in sku_list:
        sku_data['spu_id'] = spu.id
        sku = ProductSKU(**sku_data)
        session.add(sku)
    
    # 3. 更新SPU统计
    await update_spu_statistics(spu.id)
    
    await session.commit()
```

### 2. 库存管理

```python
# 下单时预占库存
async def reserve_stock(sku_id: str, quantity: int):
    sku = await get_sku_for_update(sku_id)
    
    if sku.available_stock < quantity:
        raise InsufficientStock()
    
    # 预占库存
    sku.available_stock -= quantity
    sku.reserved_stock += quantity
    
    # 记录预占
    await create_stock_record(
        sku_id, -quantity, "reserve", 
        f"预占库存 {quantity} 件"
    )
```

### 3. 价格管理

```python
# 价格变更记录
async def update_sku_price(sku_id: str, new_price: float, reason: str):
    sku = await get_sku(sku_id)
    old_price = sku.price
    
    # 记录价格变更
    price_record = PriceRecord(
        sku_id=sku_id,
        old_price=old_price,
        new_price=new_price,
        change_reason=reason
    )
    
    sku.price = new_price
    session.add_all([sku, price_record])
    await session.commit()
```

## 📚 相关文档

- [商品管理API文档](./PRODUCT_API.md)
- [订单系统API文档](./ORDER_API.md)
- [库存管理API文档](./INVENTORY_API.md)
- [批量操作API文档](./BATCH_OPERATIONS_API.md)

## 🆘 支持与反馈

如在迁移过程中遇到问题：

1. 查看迁移日志文件
2. 检查备份数据完整性
3. 参考故障排除章节
4. 联系技术支持团队

---

**迁移完成检查清单**:

- [ ] 数据备份已完成并验证
- [ ] 迁移脚本执行成功
- [ ] 数据一致性验证通过
- [ ] API接口测试正常
- [ ] 前端界面适配完成
- [ ] 性能测试通过
- [ ] 用户操作培训完成

*祝您迁移顺利！* 🎉