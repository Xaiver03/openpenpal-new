#!/usr/bin/env python3
"""
价格体系初始化脚本

创建价格管理相关数据表和初始化数据：
1. 创建价格体系数据表
2. 初始化基础价格策略
3. 创建示例促销活动模板
4. 设置基础价格规则
"""

import asyncio
import logging
import json
from datetime import datetime, timedelta
from decimal import Decimal
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker

from app.core.config import settings
from app.core.database import Base
from app.models.pricing import (
    PricePolicy, PriceRule, SkuPrice, PriceHistory, PromotionActivity,
    PriceType, DiscountType
)

# 设置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class PricingSystemInitializer:
    """价格体系初始化工具"""

    def __init__(self, database_url: str):
        self.engine = create_async_engine(database_url)
        self.SessionLocal = sessionmaker(
            bind=self.engine,
            class_=AsyncSession,
            expire_on_commit=False
        )

    async def create_tables(self):
        """创建价格体系数据表"""
        logger.info("🏗️ 创建价格体系数据表...")
        
        async with self.engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)
        
        logger.info("✅ 价格体系数据表创建完成")

    async def init_price_policies(self, session: AsyncSession):
        """初始化基础价格策略"""
        logger.info("💰 初始化基础价格策略...")
        
        policies_data = [
            {
                "policy_name": "基础定价策略",
                "policy_code": "BASE_PRICING",
                "description": "商品基础定价策略，适用于所有商品的标准定价",
                "price_type": PriceType.BASE.value,
                "is_active": True,
                "priority": 100,
                "apply_to_all": True,
                "created_by": "system",
                "updated_by": "system"
            },
            {
                "policy_name": "会员定价策略",
                "policy_code": "MEMBER_PRICING",
                "description": "会员专享价格策略，为不同等级会员提供优惠价格",
                "price_type": PriceType.MEMBER.value,
                "is_active": True,
                "priority": 200,
                "apply_to_all": True,
                "member_level_required": "VIP",
                "created_by": "system",
                "updated_by": "system"
            },
            {
                "policy_name": "批量采购定价",
                "policy_code": "BULK_PRICING",
                "description": "批量采购价格策略，购买数量越多价格越优惠",
                "price_type": PriceType.BULK.value,
                "is_active": True,
                "priority": 300,
                "apply_to_all": True,
                "min_quantity": 10,
                "created_by": "system",
                "updated_by": "system"
            },
            {
                "policy_name": "促销定价策略",
                "policy_code": "PROMOTION_PRICING",
                "description": "促销活动价格策略，用于各类促销活动",
                "price_type": PriceType.PROMOTION.value,
                "is_active": True,
                "priority": 400,
                "apply_to_all": False,
                "created_by": "system",
                "updated_by": "system"
            },
            {
                "policy_name": "动态定价策略",
                "policy_code": "DYNAMIC_PRICING",
                "description": "动态价格策略，根据供需关系和市场情况调整价格",
                "price_type": PriceType.DYNAMIC.value,
                "is_active": False,
                "priority": 500,
                "apply_to_all": False,
                "created_by": "system",
                "updated_by": "system"
            }
        ]
        
        created_policies = {}
        for policy_data in policies_data:
            policy = PricePolicy(**policy_data)
            session.add(policy)
            await session.flush()
            created_policies[policy_data["policy_code"]] = policy.policy_id
            logger.info(f"✅ 创建价格策略: {policy.policy_name} ({policy.policy_id})")
        
        await session.commit()
        logger.info(f"✅ 共创建 {len(policies_data)} 个价格策略")
        return created_policies

    async def init_price_rules(self, session: AsyncSession, created_policies: dict):
        """初始化价格规则"""
        logger.info("📏 初始化价格规则...")
        
        rules_data = [
            # 会员定价规则
            {
                "policy_id": created_policies["MEMBER_PRICING"],
                "rule_name": "VIP会员9折优惠",
                "condition_type": "member_level",
                "condition_value": json.dumps({"level": "VIP"}),
                "discount_type": DiscountType.PERCENTAGE.value,
                "discount_value": Decimal("10.00"),
                "is_active": True,
                "sort_order": 1,
                "created_by": "system"
            },
            {
                "policy_id": created_policies["MEMBER_PRICING"],
                "rule_name": "黄金会员8.5折优惠",
                "condition_type": "member_level",
                "condition_value": json.dumps({"level": "GOLD"}),
                "discount_type": DiscountType.PERCENTAGE.value,
                "discount_value": Decimal("15.00"),
                "is_active": True,
                "sort_order": 2,
                "created_by": "system"
            },
            
            # 批量定价规则
            {
                "policy_id": created_policies["BULK_PRICING"],
                "rule_name": "购买10件以上9.5折",
                "condition_type": "quantity_range",
                "condition_value": json.dumps({"min_quantity": 10, "max_quantity": 49}),
                "discount_type": DiscountType.PERCENTAGE.value,
                "discount_value": Decimal("5.00"),
                "is_active": True,
                "sort_order": 1,
                "created_by": "system"
            },
            {
                "policy_id": created_policies["BULK_PRICING"],
                "rule_name": "购买50件以上9折",
                "condition_type": "quantity_range",
                "condition_value": json.dumps({"min_quantity": 50, "max_quantity": 99}),
                "discount_type": DiscountType.PERCENTAGE.value,
                "discount_value": Decimal("10.00"),
                "is_active": True,
                "sort_order": 2,
                "created_by": "system"
            },
            {
                "policy_id": created_policies["BULK_PRICING"],
                "rule_name": "购买100件以上8折",
                "condition_type": "quantity_range",
                "condition_value": json.dumps({"min_quantity": 100}),
                "discount_type": DiscountType.PERCENTAGE.value,
                "discount_value": Decimal("20.00"),
                "is_active": True,
                "sort_order": 3,
                "created_by": "system"
            },
            
            # 促销定价规则
            {
                "policy_id": created_policies["PROMOTION_PRICING"],
                "rule_name": "满减优惠",
                "condition_type": "amount_threshold",
                "condition_value": json.dumps({"min_amount": 100}),
                "discount_type": DiscountType.FIXED_AMOUNT.value,
                "discount_value": Decimal("20.00"),
                "max_discount_amount": Decimal("50.00"),
                "is_active": True,
                "sort_order": 1,
                "created_by": "system"
            }
        ]
        
        for rule_data in rules_data:
            rule = PriceRule(**rule_data)
            session.add(rule)
            logger.info(f"✅ 创建价格规则: {rule.rule_name}")
        
        await session.commit()
        logger.info(f"✅ 共创建 {len(rules_data)} 个价格规则")

    async def init_promotion_activities(self, session: AsyncSession):
        """初始化示例促销活动"""
        logger.info("🎉 初始化示例促销活动...")
        
        now = datetime.utcnow()
        activities_data = [
            {
                "activity_name": "新用户注册专享",
                "activity_code": "NEW_USER_SPECIAL",
                "activity_type": "discount",
                "description": "新用户注册即享全场8.8折优惠",
                "start_time": now,
                "end_time": now + timedelta(days=30),
                "discount_config": json.dumps({
                    "type": "percentage",
                    "value": 12,  # 12%折扣，即8.8折
                    "max_discount": 50
                }),
                "target_config": json.dumps({
                    "apply_to_all": True
                }),
                "condition_config": json.dumps({
                    "user_type": "new_user",
                    "min_amount": 50
                }),
                "max_participants": 1000,
                "max_usage_per_user": 1,
                "status": "active",
                "is_active": True,
                "priority": 100,
                "created_by": "system"
            },
            {
                "activity_name": "双十一狂欢节",
                "activity_code": "DOUBLE_ELEVEN",
                "activity_type": "festival",
                "description": "双十一特惠，全场商品5折起",
                "start_time": now + timedelta(days=30),
                "end_time": now + timedelta(days=32),
                "discount_config": json.dumps({
                    "type": "tiered",
                    "tiers": [
                        {"min_amount": 100, "discount": 20},
                        {"min_amount": 300, "discount": 60},
                        {"min_amount": 500, "discount": 120}
                    ]
                }),
                "target_config": json.dumps({
                    "category_ids": ["1", "2", "3"],
                    "exclude_brands": ["luxury_brand"]
                }),
                "condition_config": json.dumps({
                    "min_amount": 100,
                    "max_discount": 200
                }),
                "max_participants": 10000,
                "max_usage_per_user": 3,
                "status": "scheduled",
                "is_active": False,
                "priority": 500,
                "created_by": "system"
            },
            {
                "activity_name": "买二送一",
                "activity_code": "BUY_TWO_GET_ONE",
                "activity_type": "bundle",
                "description": "指定商品买二送一，第三件免费",
                "start_time": now,
                "end_time": now + timedelta(days=15),
                "discount_config": json.dumps({
                    "type": "buy_x_get_y",
                    "buy_quantity": 2,
                    "get_quantity": 1,
                    "get_discount": 100  # 免费
                }),
                "target_config": json.dumps({
                    "spu_ids": ["SPU001", "SPU002", "SPU003"]
                }),
                "condition_config": json.dumps({
                    "min_quantity": 3
                }),
                "max_participants": 500,
                "max_usage_per_user": 2,
                "status": "active",
                "is_active": True,
                "priority": 300,
                "created_by": "system"
            }
        ]
        
        for activity_data in activities_data:
            activity = PromotionActivity(**activity_data)
            session.add(activity)
            logger.info(f"✅ 创建促销活动: {activity.activity_name}")
        
        await session.commit()
        logger.info(f"✅ 共创建 {len(activities_data)} 个促销活动")

    async def init_sample_sku_prices(self, session: AsyncSession, created_policies: dict):
        """初始化示例SKU价格"""
        logger.info("🏷️ 初始化示例SKU价格...")
        
        # 获取基础定价策略ID
        base_policy_id = created_policies["BASE_PRICING"]
        member_policy_id = created_policies["MEMBER_PRICING"]
        
        sample_prices = [
            {
                "sku_id": "SKU001",
                "policy_id": base_policy_id,
                "original_price": Decimal("99.90"),
                "current_price": Decimal("89.90"),
                "cost_price": Decimal("45.00"),
                "min_price": Decimal("60.00"),
                "max_price": Decimal("120.00"),
                "vip_price": Decimal("79.90"),
                "member_price": Decimal("84.90"),
                "bulk_config": json.dumps({
                    "tiers": [
                        {"min_quantity": 10, "discount_type": "percentage", "discount": 5},
                        {"min_quantity": 50, "discount_type": "percentage", "discount": 10},
                        {"min_quantity": 100, "discount_type": "percentage", "discount": 15}
                    ]
                }),
                "is_active": True,
                "created_by": "system"
            },
            {
                "sku_id": "SKU002",
                "policy_id": base_policy_id,
                "original_price": Decimal("199.90"),
                "current_price": Decimal("179.90"),
                "cost_price": Decimal("90.00"),
                "min_price": Decimal("120.00"),
                "max_price": Decimal("250.00"),
                "vip_price": Decimal("159.90"),
                "member_price": Decimal("169.90"),
                "bulk_config": json.dumps({
                    "tiers": [
                        {"min_quantity": 5, "discount_type": "percentage", "discount": 5},
                        {"min_quantity": 20, "discount_type": "percentage", "discount": 10},
                        {"min_quantity": 50, "discount_type": "percentage", "discount": 15}
                    ]
                }),
                "is_active": True,
                "created_by": "system"
            },
            {
                "sku_id": "SKU003",
                "policy_id": base_policy_id,
                "original_price": Decimal("59.90"),
                "current_price": Decimal("49.90"),
                "cost_price": Decimal("25.00"),
                "min_price": Decimal("35.00"),
                "max_price": Decimal("80.00"),
                "vip_price": Decimal("44.90"),
                "member_price": Decimal("47.90"),
                "bulk_config": json.dumps({
                    "tiers": [
                        {"min_quantity": 20, "discount_type": "percentage", "discount": 8},
                        {"min_quantity": 100, "discount_type": "percentage", "discount": 15},
                        {"min_quantity": 500, "discount_type": "percentage", "discount": 25}
                    ]
                }),
                "is_active": True,
                "created_by": "system"
            }
        ]
        
        for price_data in sample_prices:
            price = SkuPrice(**price_data)
            session.add(price)
            logger.info(f"✅ 创建SKU价格: {price.sku_id} - ¥{price.current_price}")
        
        await session.commit()
        logger.info(f"✅ 共创建 {len(sample_prices)} 个SKU价格")

    async def run_initialization(self):
        """运行完整初始化流程"""
        logger.info("🚀 开始价格体系初始化...")
        logger.info("=" * 60)
        
        start_time = datetime.now()
        
        try:
            # 1. 创建数据表
            await self.create_tables()
            
            async with self.SessionLocal() as session:
                # 2. 初始化价格策略
                created_policies = await self.init_price_policies(session)
                
                # 3. 初始化价格规则
                await self.init_price_rules(session, created_policies)
                
                # 4. 初始化促销活动
                await self.init_promotion_activities(session)
                
                # 5. 初始化示例SKU价格
                await self.init_sample_sku_prices(session, created_policies)
            
            end_time = datetime.now()
            duration = end_time - start_time
            
            logger.info("=" * 60)
            logger.info(f"🎉 价格体系初始化完成！耗时: {duration}")
            logger.info("📊 初始化统计:")
            logger.info(f"   价格策略: {len(created_policies)} 个")
            logger.info(f"   价格规则: 6 个")
            logger.info(f"   促销活动: 3 个")
            logger.info(f"   SKU价格: 3 个")
            logger.info("")
            logger.info("🎯 价格体系功能:")
            logger.info("   ✅ 基础定价策略")
            logger.info("   ✅ 会员差异定价")
            logger.info("   ✅ 批量采购优惠")
            logger.info("   ✅ 促销活动支持")
            logger.info("   ✅ 价格历史跟踪")
            logger.info("   ✅ 动态价格计算")
            
            return True
            
        except Exception as e:
            logger.error(f"❌ 价格体系初始化失败: {e}")
            raise
        
        finally:
            await self.engine.dispose()


async def main():
    """主函数"""
    print("价格体系管理初始化工具")
    print("=" * 60)
    print("⚠️ 注意：此操作会创建价格体系相关数据表并初始化数据")
    print("📋 初始化内容：")
    print("   1. 创建价格体系数据表")
    print("   2. 初始化5个基础价格策略")
    print("   3. 创建6个价格规则")
    print("   4. 创建3个示例促销活动")
    print("   5. 初始化3个SKU价格示例")
    print()
    
    confirm = input("确认继续执行初始化？(y/N): ")
    if confirm.lower() != 'y':
        print("❌ 初始化已取消")
        return
    
    # 使用配置中的数据库URL
    database_url = settings.DATABASE_URL
    if database_url.startswith("postgresql://"):
        database_url = database_url.replace("postgresql://", "postgresql+asyncpg://")
    
    initializer = PricingSystemInitializer(database_url)
    await initializer.run_initialization()


if __name__ == "__main__":
    asyncio.run(main())