#!/usr/bin/env python3
"""
数据库迁移脚本：从旧商品模型迁移到SPU+SKU模型

该脚本负责：
1. 创建新的SPU+SKU表结构
2. 将现有商品数据迁移到新模型
3. 更新订单关联关系
4. 保持数据完整性

运行前请备份数据库！
"""

import asyncio
import logging
from typing import Dict, List, Optional
from sqlalchemy import create_engine, text
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from datetime import datetime
import json
import uuid

from app.core.config import settings
from app.core.database import Base
from app.models.shop import Product as OldProduct, Order as OldOrder, OrderItem as OldOrderItem
from app.models.product_new import (
    ProductSPU, ProductSKU, ProductCategory, ProductBrand, 
    ProductAttribute, StockRecord, PriceRecord
)
from app.models.order_new import Order as NewOrder, OrderItem as NewOrderItem

# 设置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class SPUSKUMigrator:
    """SPU+SKU迁移工具"""
    
    def __init__(self, database_url: str):
        self.engine = create_async_engine(database_url)
        self.SessionLocal = sessionmaker(
            bind=self.engine, 
            class_=AsyncSession, 
            expire_on_commit=False
        )
        
        # 映射关系存储
        self.product_to_spu_mapping: Dict[str, str] = {}
        self.product_to_sku_mapping: Dict[str, str] = {}
        self.category_mapping: Dict[str, str] = {}
        self.brand_mapping: Dict[str, str] = {}
    
    async def create_tables(self):
        """创建新表结构"""
        logger.info("🏗️ 创建新表结构...")
        
        async with self.engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)
        
        logger.info("✅ 新表结构创建完成")
    
    def generate_id(self, prefix: str = "") -> str:
        """生成唯一ID"""
        return f"{prefix}{uuid.uuid4().hex[:16].upper()}"
    
    async def migrate_categories(self, session: AsyncSession):
        """迁移商品分类"""
        logger.info("📁 迁移商品分类...")
        
        # 获取旧分类数据（从商品中提取）
        result = await session.execute(
            text("SELECT DISTINCT category FROM shop_products WHERE category IS NOT NULL")
        )
        categories = result.fetchall()
        
        for (category_name,) in categories:
            if not category_name or category_name in self.category_mapping:
                continue
            
            category_id = self.generate_id("CAT")
            
            new_category = ProductCategory(
                id=category_id,
                name=category_name,
                description=f"从旧系统迁移的分类：{category_name}",
                parent_id=None,
                path=f"/{category_id}/",
                level=0,
                is_active=True,
                is_visible=True,
                sort_order=0,
                spu_count=0
            )
            
            session.add(new_category)
            self.category_mapping[category_name] = category_id
        
        await session.commit()
        logger.info(f"✅ 迁移了 {len(self.category_mapping)} 个分类")
    
    async def migrate_brands(self, session: AsyncSession):
        """迁移品牌信息"""
        logger.info("🏷️ 迁移品牌信息...")
        
        # 获取旧品牌数据
        result = await session.execute(
            text("SELECT DISTINCT brand FROM shop_products WHERE brand IS NOT NULL")
        )
        brands = result.fetchall()
        
        for (brand_name,) in brands:
            if not brand_name or brand_name in self.brand_mapping:
                continue
            
            brand_id = self.generate_id("BRD")
            
            new_brand = ProductBrand(
                id=brand_id,
                name=brand_name,
                description=f"从旧系统迁移的品牌：{brand_name}",
                is_active=True,
                sort_order=0,
                spu_count=0
            )
            
            session.add(new_brand)
            self.brand_mapping[brand_name] = brand_id
        
        await session.commit()
        logger.info(f"✅ 迁移了 {len(self.brand_mapping)} 个品牌")
    
    async def migrate_products_to_spu_sku(self, session: AsyncSession):
        """将旧商品迁移为SPU+SKU模型"""
        logger.info("📦 迁移商品到SPU+SKU模型...")
        
        # 获取所有旧商品
        result = await session.execute(text("SELECT * FROM shop_products"))
        old_products = result.fetchall()
        
        migrated_count = 0
        
        for row in old_products:
            try:
                # 创建SPU
                spu_id = self.generate_id("SPU")
                
                category_id = self.category_mapping.get(row.category)
                brand_id = self.brand_mapping.get(row.brand)
                
                # 解析标签
                tags = []
                if row.tags:
                    tags = [tag.strip() for tag in row.tags.split(',') if tag.strip()]
                
                # 解析图片集
                gallery_images = []
                if row.gallery_images:
                    try:
                        gallery_images = json.loads(row.gallery_images)
                    except:
                        gallery_images = []
                
                new_spu = ProductSPU(
                    id=spu_id,
                    name=row.name,
                    subtitle=None,
                    description=row.description,
                    short_description=row.short_description,
                    category_id=category_id,
                    product_type=row.product_type,
                    brand_id=brand_id,
                    tags=tags,
                    status=row.status,
                    is_featured=row.is_featured,
                    is_digital=row.is_digital,
                    is_virtual=False,
                    main_image=row.main_image,
                    gallery_images=gallery_images,
                    video_url=row.video_url,
                    seo_title=row.seo_title,
                    seo_description=row.seo_description,
                    seo_keywords=row.seo_keywords.split(',') if row.seo_keywords else [],
                    total_stock=row.stock_quantity,
                    min_price=row.price,
                    max_price=row.price,
                    view_count=row.view_count,
                    sales_count=row.sales_count,
                    rating_avg=row.rating_avg,
                    rating_count=row.rating_count,
                    favorite_count=row.favorite_count,
                    creator_id=row.creator_id,
                    creator_name=row.creator_name,
                    created_at=row.created_at,
                    updated_at=row.updated_at,
                    published_at=row.published_at
                )
                
                session.add(new_spu)
                
                # 创建默认SKU
                sku_id = self.generate_id("SKU")
                
                new_sku = ProductSKU(
                    id=sku_id,
                    spu_id=spu_id,
                    sku_code=f"SKU-{row.id}",
                    name="默认规格",
                    price=row.price,
                    original_price=row.original_price,
                    cost_price=row.cost_price,
                    currency=row.currency,
                    stock_quantity=row.stock_quantity,
                    available_stock=row.stock_quantity,
                    reserved_stock=0,
                    min_stock=row.min_stock,
                    max_quantity_per_order=row.max_quantity_per_order,
                    status="active" if row.status == "active" else "inactive",
                    is_default=True,
                    weight=row.weight,
                    dimensions=row.dimensions,
                    sale_attributes={
                        "color": row.color,
                        "material": row.material
                    },
                    main_image=row.main_image,
                    sales_count=row.sales_count,
                    created_at=row.created_at,
                    updated_at=row.updated_at
                )
                
                session.add(new_sku)
                
                # 创建基本属性
                if row.material:
                    attr_id = self.generate_id("ATTR")
                    material_attr = ProductAttribute(
                        id=attr_id,
                        spu_id=spu_id,
                        attribute_name="材质",
                        attribute_value=row.material,
                        attribute_type="basic",
                        is_required=False,
                        is_searchable=True,
                        is_filterable=True,
                        sort_order=1
                    )
                    session.add(material_attr)
                
                # 记录映射关系
                self.product_to_spu_mapping[row.id] = spu_id
                self.product_to_sku_mapping[row.id] = sku_id
                
                migrated_count += 1
                
                if migrated_count % 100 == 0:
                    await session.commit()
                    logger.info(f"已迁移 {migrated_count} 个商品...")
                    
            except Exception as e:
                logger.error(f"迁移商品 {row.id} 失败: {e}")
                continue
        
        await session.commit()
        logger.info(f"✅ 成功迁移 {migrated_count} 个商品到SPU+SKU模型")
    
    async def migrate_orders(self, session: AsyncSession):
        """迁移订单数据"""
        logger.info("🛒 迁移订单数据...")
        
        # 获取所有旧订单
        result = await session.execute(text("SELECT * FROM shop_orders"))
        old_orders = result.fetchall()
        
        migrated_count = 0
        
        for row in old_orders:
            try:
                # 创建新订单
                new_order = NewOrder(
                    id=row.id,
                    order_no=f"ORD{row.id}",  # 生成订单编号
                    user_id=row.user_id,
                    user_name=row.user_name,
                    user_email=row.user_email,
                    user_phone=row.user_phone,
                    status=row.status,
                    payment_status=row.payment_status,
                    shipping_status="pending",
                    refund_status="none",
                    item_count=0,  # 稍后计算
                    subtotal=row.subtotal,
                    shipping_fee=row.shipping_fee,
                    tax_fee=row.tax_fee,
                    discount_amount=row.discount_amount,
                    coupon_discount=row.coupon_discount,
                    total_amount=row.total_amount,
                    actual_payment=row.total_amount,
                    currency=row.currency,
                    shipping_info={
                        "name": row.shipping_name,
                        "phone": row.shipping_phone,
                        "address": row.shipping_address,
                        "city": row.shipping_city,
                        "province": row.shipping_province,
                        "postal_code": row.shipping_postal_code
                    },
                    shipping_method=row.shipping_method,
                    user_note=row.user_note,
                    admin_note=row.admin_note,
                    coupon_info={
                        "code": row.coupon_code,
                        "discount": row.coupon_discount
                    },
                    payment_method=row.payment_method,
                    payment_info={
                        "transaction_id": row.payment_transaction_id
                    },
                    paid_at=row.paid_at,
                    shipped_at=row.shipped_at,
                    delivered_at=row.delivered_at,
                    created_at=row.created_at,
                    updated_at=row.updated_at
                )
                
                session.add(new_order)
                migrated_count += 1
                
            except Exception as e:
                logger.error(f"迁移订单 {row.id} 失败: {e}")
                continue
        
        await session.commit()
        logger.info(f"✅ 成功迁移 {migrated_count} 个订单")
    
    async def migrate_order_items(self, session: AsyncSession):
        """迁移订单项数据"""
        logger.info("📋 迁移订单项数据...")
        
        # 获取所有旧订单项
        result = await session.execute(text("SELECT * FROM shop_order_items"))
        old_order_items = result.fetchall()
        
        migrated_count = 0
        item_count_by_order = {}
        
        for row in old_order_items:
            try:
                # 获取对应的SPU和SKU ID
                spu_id = self.product_to_spu_mapping.get(row.product_id)
                sku_id = self.product_to_sku_mapping.get(row.product_id)
                
                if not spu_id or not sku_id:
                    logger.warning(f"找不到商品 {row.product_id} 的SPU/SKU映射")
                    continue
                
                # 解析商品属性
                product_attributes = {}
                if row.product_attributes:
                    try:
                        product_attributes = json.loads(row.product_attributes)
                    except:
                        product_attributes = {}
                
                new_order_item = NewOrderItem(
                    id=row.id,
                    order_id=row.order_id,
                    spu_id=spu_id,
                    sku_id=sku_id,
                    spu_name=row.product_name,
                    sku_name="默认规格",
                    sku_code=f"SKU-{row.product_id}",
                    spu_image=row.product_image,
                    sku_image=row.product_image,
                    sale_attributes=product_attributes,
                    basic_attributes={},
                    unit_price=row.unit_price,
                    original_price=row.unit_price,
                    quantity=row.quantity,
                    total_price=row.total_price,
                    item_discount=0.0,
                    promotion_info={},
                    item_status="normal",
                    refund_quantity=0,
                    refund_amount=0.0,
                    created_at=row.created_at
                )
                
                session.add(new_order_item)
                
                # 统计每个订单的商品件数
                if row.order_id not in item_count_by_order:
                    item_count_by_order[row.order_id] = 0
                item_count_by_order[row.order_id] += row.quantity
                
                migrated_count += 1
                
            except Exception as e:
                logger.error(f"迁移订单项 {row.id} 失败: {e}")
                continue
        
        await session.commit()
        
        # 更新订单的商品件数
        for order_id, item_count in item_count_by_order.items():
            await session.execute(
                text("UPDATE orders SET item_count = :count WHERE id = :id"),
                {"count": item_count, "id": order_id}
            )
        
        await session.commit()
        logger.info(f"✅ 成功迁移 {migrated_count} 个订单项")
    
    async def create_initial_stock_records(self, session: AsyncSession):
        """为所有SKU创建初始库存记录"""
        logger.info("📊 创建初始库存记录...")
        
        result = await session.execute(text("SELECT id, stock_quantity FROM product_sku"))
        skus = result.fetchall()
        
        created_count = 0
        
        for sku_id, stock_quantity in skus:
            if stock_quantity > 0:
                stock_record = StockRecord(
                    id=self.generate_id("STK"),
                    sku_id=sku_id,
                    change_type="initial",
                    change_quantity=stock_quantity,
                    before_quantity=0,
                    after_quantity=stock_quantity,
                    reference_type="migration",
                    reference_id="initial_migration",
                    operator_id="system",
                    operator_name="系统迁移",
                    remark="数据迁移时的初始库存"
                )
                session.add(stock_record)
                created_count += 1
        
        await session.commit()
        logger.info(f"✅ 创建了 {created_count} 条初始库存记录")
    
    async def update_spu_statistics(self, session: AsyncSession):
        """更新SPU统计信息"""
        logger.info("📈 更新SPU统计信息...")
        
        # 更新分类的SPU数量
        await session.execute(text("""
            UPDATE product_categories 
            SET spu_count = (
                SELECT COUNT(*) FROM product_spu 
                WHERE category_id = product_categories.id
            )
        """))
        
        # 更新品牌的SPU数量  
        await session.execute(text("""
            UPDATE product_brands 
            SET spu_count = (
                SELECT COUNT(*) FROM product_spu 
                WHERE brand_id = product_brands.id
            )
        """))
        
        await session.commit()
        logger.info("✅ SPU统计信息更新完成")
    
    async def verify_migration(self, session: AsyncSession):
        """验证迁移结果"""
        logger.info("🔍 验证迁移结果...")
        
        # 统计新表数据
        stats = {}
        
        for table in ['product_spu', 'product_sku', 'product_categories', 
                     'product_brands', 'orders', 'order_items', 'stock_records']:
            result = await session.execute(text(f"SELECT COUNT(*) FROM {table}"))
            count = result.fetchone()[0]
            stats[table] = count
        
        logger.info("📊 迁移统计:")
        for table, count in stats.items():
            logger.info(f"   {table}: {count} 条记录")
        
        # 验证数据一致性
        result = await session.execute(text("""
            SELECT 
                (SELECT COUNT(*) FROM shop_products) as old_products,
                (SELECT COUNT(*) FROM product_spu) as new_spu,
                (SELECT COUNT(*) FROM product_sku) as new_sku,
                (SELECT COUNT(*) FROM shop_orders) as old_orders,
                (SELECT COUNT(*) FROM orders) as new_orders,
                (SELECT COUNT(*) FROM shop_order_items) as old_order_items,
                (SELECT COUNT(*) FROM order_items) as new_order_items
        """))
        
        counts = result.fetchone()
        
        logger.info("🎯 数据一致性检查:")
        logger.info(f"   商品: {counts.old_products} → SPU: {counts.new_spu}, SKU: {counts.new_sku}")
        logger.info(f"   订单: {counts.old_orders} → {counts.new_orders}")  
        logger.info(f"   订单项: {counts.old_order_items} → {counts.new_order_items}")
        
        # 检查是否有数据丢失
        if counts.old_products != counts.new_spu:
            logger.warning(f"⚠️ 商品数量不匹配: {counts.old_products} vs {counts.new_spu}")
        
        if counts.old_orders != counts.new_orders:
            logger.warning(f"⚠️ 订单数量不匹配: {counts.old_orders} vs {counts.new_orders}")
            
        if counts.old_order_items != counts.new_order_items:
            logger.warning(f"⚠️ 订单项数量不匹配: {counts.old_order_items} vs {counts.new_order_items}")
    
    async def run_migration(self):
        """运行完整迁移流程"""
        logger.info("🚀 开始SPU+SKU模型迁移...")
        logger.info("=" * 60)
        
        start_time = datetime.now()
        
        try:
            # 创建新表结构
            await self.create_tables()
            
            async with self.SessionLocal() as session:
                # 1. 迁移分类和品牌
                await self.migrate_categories(session)
                await self.migrate_brands(session)
                
                # 2. 迁移商品到SPU+SKU
                await self.migrate_products_to_spu_sku(session)
                
                # 3. 迁移订单数据
                await self.migrate_orders(session)
                await self.migrate_order_items(session)
                
                # 4. 创建库存记录
                await self.create_initial_stock_records(session)
                
                # 5. 更新统计信息
                await self.update_spu_statistics(session)
                
                # 6. 验证迁移结果
                await self.verify_migration(session)
            
            end_time = datetime.now()
            duration = end_time - start_time
            
            logger.info("=" * 60)
            logger.info(f"🎉 迁移完成！耗时: {duration}")
            logger.info("✅ 所有数据已成功迁移到新的SPU+SKU模型")
            logger.info("💡 请测试新系统功能，确认无误后可删除旧表")
            
        except Exception as e:
            logger.error(f"❌ 迁移失败: {e}")
            raise
        
        finally:
            await self.engine.dispose()

async def main():
    """主函数"""
    print("SPU+SKU模型迁移工具")
    print("=" * 60)
    print("⚠️  警告：请在运行前备份数据库！")
    print("📋 本工具将执行以下操作：")
    print("   1. 创建新的SPU+SKU表结构")
    print("   2. 将现有商品迁移到SPU+SKU模型")  
    print("   3. 更新订单关联关系")
    print("   4. 创建库存记录")
    print("   5. 验证数据完整性")
    print()
    
    confirm = input("确认继续执行迁移？(y/N): ")
    if confirm.lower() != 'y':
        print("❌ 迁移已取消")
        return
    
    # 使用配置中的数据库URL
    database_url = settings.DATABASE_URL
    if database_url.startswith("postgresql://"):
        database_url = database_url.replace("postgresql://", "postgresql+asyncpg://")
    
    migrator = SPUSKUMigrator(database_url)
    await migrator.run_migration()

if __name__ == "__main__":
    asyncio.run(main())