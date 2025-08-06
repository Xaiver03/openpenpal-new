"""
价格服务

处理商品定价的核心业务逻辑：
1. 价格策略管理
2. 价格计算引擎
3. 促销活动管理
4. 价格历史跟踪
5. 批量定价操作
"""

import json
import asyncio
from typing import List, Dict, Any, Optional, Tuple
from decimal import Decimal, ROUND_HALF_UP
from datetime import datetime, timedelta
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, update, delete, and_, or_, func
from sqlalchemy.orm import selectinload

from app.models.pricing import (
    PricePolicy, PriceRule, SkuPrice, PriceHistory, PromotionActivity,
    PriceType, DiscountType
)
from app.core.exceptions import BusinessException
from app.core.logger import get_logger
logger = get_logger(__name__)


class PricingService:
    """价格服务"""

    def __init__(self, session: AsyncSession):
        self.session = session

    # ==================== 价格策略管理 ====================

    async def create_price_policy(self, policy_data: Dict[str, Any]) -> PricePolicy:
        """创建价格策略"""
        try:
            # 检查策略编码是否重复
            existing = await self.session.execute(
                select(PricePolicy).where(PricePolicy.policy_code == policy_data["policy_code"])
            )
            if existing.scalar_one_or_none():
                raise BusinessException(f"价格策略编码 {policy_data['policy_code']} 已存在")

            policy = PricePolicy(**policy_data)
            self.session.add(policy)
            await self.session.commit()
            await self.session.refresh(policy)

            logger.info(f"创建价格策略成功: {policy.policy_name}")
            return policy

        except Exception as e:
            await self.session.rollback()
            logger.error(f"创建价格策略失败: {str(e)}")
            raise

    async def update_price_policy(self, policy_id: int, policy_data: Dict[str, Any]) -> PricePolicy:
        """更新价格策略"""
        try:
            policy = await self.session.get(PricePolicy, policy_id)
            if not policy:
                raise BusinessException("价格策略不存在")

            for key, value in policy_data.items():
                if hasattr(policy, key) and value is not None:
                    setattr(policy, key, value)

            policy.updated_at = datetime.utcnow()
            await self.session.commit()
            await self.session.refresh(policy)

            logger.info(f"更新价格策略成功: {policy.policy_name}")
            return policy

        except Exception as e:
            await self.session.rollback()
            logger.error(f"更新价格策略失败: {str(e)}")
            raise

    async def delete_price_policy(self, policy_id: int) -> bool:
        """删除价格策略"""
        try:
            # 检查是否有关联的SKU价格
            sku_prices = await self.session.execute(
                select(SkuPrice).where(SkuPrice.policy_id == policy_id)
            )
            if sku_prices.scalar_one_or_none():
                raise BusinessException("该价格策略下还有SKU价格，无法删除")

            # 删除价格规则
            await self.session.execute(
                delete(PriceRule).where(PriceRule.policy_id == policy_id)
            )

            # 删除价格策略
            await self.session.execute(
                delete(PricePolicy).where(PricePolicy.policy_id == policy_id)
            )

            await self.session.commit()
            logger.info(f"删除价格策略成功: {policy_id}")
            return True

        except Exception as e:
            await self.session.rollback()
            logger.error(f"删除价格策略失败: {str(e)}")
            raise

    async def get_price_policies(self, 
                               price_type: Optional[int] = None,
                               is_active: Optional[bool] = None,
                               page: int = 1, 
                               size: int = 20) -> Tuple[List[PricePolicy], int]:
        """获取价格策略列表"""
        try:
            query = select(PricePolicy)
            
            # 添加过滤条件
            conditions = []
            if price_type is not None:
                conditions.append(PricePolicy.price_type == price_type)
            if is_active is not None:
                conditions.append(PricePolicy.is_active == is_active)
            
            if conditions:
                query = query.where(and_(*conditions))

            # 获取总数
            count_query = select(func.count(PricePolicy.policy_id))
            if conditions:
                count_query = count_query.where(and_(*conditions))
            
            total = await self.session.execute(count_query)
            total_count = total.scalar()

            # 分页查询
            query = query.order_by(PricePolicy.priority.desc(), PricePolicy.created_at.desc())
            query = query.offset((page - 1) * size).limit(size)
            
            result = await self.session.execute(query)
            policies = result.scalars().all()

            return policies, total_count

        except Exception as e:
            logger.error(f"获取价格策略列表失败: {str(e)}")
            raise

    # ==================== 价格规则管理 ====================

    async def create_price_rule(self, rule_data: Dict[str, Any]) -> PriceRule:
        """创建价格规则"""
        try:
            # 验证策略是否存在
            policy = await self.session.get(PricePolicy, rule_data["policy_id"])
            if not policy:
                raise BusinessException("价格策略不存在")

            rule = PriceRule(**rule_data)
            self.session.add(rule)
            await self.session.commit()
            await self.session.refresh(rule)

            logger.info(f"创建价格规则成功: {rule.rule_name}")
            return rule

        except Exception as e:
            await self.session.rollback()
            logger.error(f"创建价格规则失败: {str(e)}")
            raise

    async def get_policy_rules(self, policy_id: int) -> List[PriceRule]:
        """获取策略的价格规则"""
        try:
            result = await self.session.execute(
                select(PriceRule)
                .where(PriceRule.policy_id == policy_id)
                .where(PriceRule.is_active == True)
                .order_by(PriceRule.sort_order)
            )
            return result.scalars().all()

        except Exception as e:
            logger.error(f"获取价格规则失败: {str(e)}")
            raise

    # ==================== SKU价格管理 ====================

    async def set_sku_price(self, price_data: Dict[str, Any]) -> SkuPrice:
        """设置SKU价格"""
        try:
            # 检查是否已存在价格记录
            existing = await self.session.execute(
                select(SkuPrice).where(
                    and_(
                        SkuPrice.sku_id == price_data["sku_id"],
                        SkuPrice.policy_id == price_data["policy_id"]
                    )
                )
            )
            existing_price = existing.scalar_one_or_none()

            if existing_price:
                # 记录价格历史
                await self._record_price_history(
                    price_data["sku_id"],
                    "current_price",
                    existing_price.current_price,
                    price_data.get("current_price"),
                    "价格更新",
                    price_data.get("updated_by")
                )

                # 更新现有价格
                for key, value in price_data.items():
                    if hasattr(existing_price, key) and value is not None:
                        setattr(existing_price, key, value)
                existing_price.updated_at = datetime.utcnow()
                await self.session.commit()
                await self.session.refresh(existing_price)
                return existing_price
            else:
                # 创建新价格记录
                price = SkuPrice(**price_data)
                self.session.add(price)
                await self.session.commit()
                await self.session.refresh(price)

                # 记录价格历史
                await self._record_price_history(
                    price_data["sku_id"],
                    "current_price",
                    None,
                    price_data.get("current_price"),
                    "首次设价",
                    price_data.get("created_by")
                )

                logger.info(f"设置SKU价格成功: {price.sku_id}")
                return price

        except Exception as e:
            await self.session.rollback()
            logger.error(f"设置SKU价格失败: {str(e)}")
            raise

    async def get_sku_price(self, sku_id: str, policy_id: Optional[int] = None) -> Optional[SkuPrice]:
        """获取SKU价格"""
        try:
            query = select(SkuPrice).where(SkuPrice.sku_id == sku_id)
            
            if policy_id:
                query = query.where(SkuPrice.policy_id == policy_id)
            else:
                # 获取默认基础价格
                query = query.join(PricePolicy).where(
                    PricePolicy.price_type == PriceType.BASE.value
                )

            query = query.where(SkuPrice.is_active == True)
            
            result = await self.session.execute(query)
            return result.scalar_one_or_none()

        except Exception as e:
            logger.error(f"获取SKU价格失败: {str(e)}")
            raise

    async def batch_set_sku_prices(self, prices_data: List[Dict[str, Any]]) -> List[SkuPrice]:
        """批量设置SKU价格"""
        try:
            created_prices = []
            
            for price_data in prices_data:
                price = await self.set_sku_price(price_data)
                created_prices.append(price)

            logger.info(f"批量设置SKU价格成功，共 {len(created_prices)} 条记录")
            return created_prices

        except Exception as e:
            logger.error(f"批量设置SKU价格失败: {str(e)}")
            raise

    # ==================== 价格计算引擎 ====================

    async def calculate_price(self, 
                            sku_id: str, 
                            quantity: int = 1,
                            user_id: Optional[str] = None,
                            member_level: Optional[str] = None) -> Dict[str, Any]:
        """价格计算引擎"""
        try:
            # 获取基础价格
            base_price = await self.get_sku_price(sku_id)
            if not base_price:
                raise BusinessException(f"SKU {sku_id} 未设置价格")

            result = {
                "sku_id": sku_id,
                "quantity": quantity,
                "original_price": float(base_price.original_price),
                "base_price": float(base_price.current_price),
                "final_price": float(base_price.current_price),
                "total_amount": float(base_price.current_price) * quantity,
                "discount_amount": 0.0,
                "applied_promotions": []
            }

            # 计算会员价格
            if member_level and base_price.member_price:
                member_price = float(base_price.member_price)
                if member_price < result["final_price"]:
                    discount_amount = result["final_price"] - member_price
                    result["final_price"] = member_price
                    result["discount_amount"] += discount_amount
                    result["applied_promotions"].append({
                        "type": "member",
                        "name": f"{member_level}会员价",
                        "discount_amount": discount_amount
                    })

            # 计算批量价格
            if quantity > 1 and base_price.bulk_config:
                bulk_price = await self._calculate_bulk_price(base_price, quantity)
                if bulk_price < result["final_price"]:
                    discount_amount = result["final_price"] - bulk_price
                    result["final_price"] = bulk_price
                    result["discount_amount"] += discount_amount
                    result["applied_promotions"].append({
                        "type": "bulk",
                        "name": "批量优惠",
                        "discount_amount": discount_amount
                    })

            # 计算促销价格
            promotions = await self._get_applicable_promotions(sku_id, quantity, user_id)
            for promotion in promotions:
                promotion_discount = await self._calculate_promotion_discount(
                    promotion, result["final_price"], quantity
                )
                if promotion_discount > 0:
                    result["final_price"] = max(0, result["final_price"] - promotion_discount)
                    result["discount_amount"] += promotion_discount
                    result["applied_promotions"].append({
                        "type": "promotion",
                        "name": promotion.activity_name,
                        "discount_amount": promotion_discount,
                        "promotion_id": promotion.activity_id
                    })

            # 重新计算总金额
            result["total_amount"] = result["final_price"] * quantity

            return result

        except Exception as e:
            logger.error(f"价格计算失败: {str(e)}")
            raise

    async def _calculate_bulk_price(self, base_price: SkuPrice, quantity: int) -> float:
        """计算批量价格"""
        try:
            if not base_price.bulk_config:
                return float(base_price.current_price)

            bulk_config = json.loads(base_price.bulk_config)
            current_price = float(base_price.current_price)

            # 按数量阶梯查找最佳价格
            best_price = current_price
            for tier in bulk_config.get("tiers", []):
                min_qty = tier.get("min_quantity", 1)
                discount = tier.get("discount", 0)
                
                if quantity >= min_qty:
                    if tier.get("discount_type") == "percentage":
                        tier_price = current_price * (1 - discount / 100)
                    else:  # fixed amount
                        tier_price = current_price - discount
                    
                    best_price = min(best_price, tier_price)

            return max(0, best_price)

        except Exception as e:
            logger.error(f"计算批量价格失败: {str(e)}")
            return float(base_price.current_price)

    async def _get_applicable_promotions(self, 
                                       sku_id: str, 
                                       quantity: int,
                                       user_id: Optional[str] = None) -> List[PromotionActivity]:
        """获取适用的促销活动"""
        try:
            now = datetime.utcnow()
            
            result = await self.session.execute(
                select(PromotionActivity)
                .where(
                    and_(
                        PromotionActivity.is_active == True,
                        PromotionActivity.start_time <= now,
                        PromotionActivity.end_time >= now
                    )
                )
                .order_by(PromotionActivity.priority.desc())
            )
            
            promotions = result.scalars().all()
            applicable_promotions = []

            for promotion in promotions:
                # 检查商品是否在促销范围内
                if await self._is_sku_in_promotion(promotion, sku_id):
                    # 检查数量条件
                    if await self._check_promotion_quantity_condition(promotion, quantity):
                        # 检查用户条件（如果有用户ID）
                        if not user_id or await self._check_user_promotion_condition(promotion, user_id):
                            applicable_promotions.append(promotion)

            return applicable_promotions

        except Exception as e:
            logger.error(f"获取适用促销活动失败: {str(e)}")
            return []

    async def _is_sku_in_promotion(self, promotion: PromotionActivity, sku_id: str) -> bool:
        """检查SKU是否在促销范围内"""
        try:
            if not promotion.target_config:
                return True  # 如果没有配置目标商品，则适用所有商品

            target_config = json.loads(promotion.target_config)
            
            # 检查指定SKU
            if "sku_ids" in target_config:
                return sku_id in target_config["sku_ids"]
            
            # TODO: 检查SPU、分类、品牌等条件
            # 这里可以扩展更复杂的商品匹配逻辑
            
            return True

        except Exception:
            return False

    async def _check_promotion_quantity_condition(self, promotion: PromotionActivity, quantity: int) -> bool:
        """检查促销活动的数量条件"""
        try:
            if not promotion.condition_config:
                return True

            condition_config = json.loads(promotion.condition_config)
            min_quantity = condition_config.get("min_quantity", 1)
            max_quantity = condition_config.get("max_quantity")

            if quantity < min_quantity:
                return False
            
            if max_quantity and quantity > max_quantity:
                return False

            return True

        except Exception:
            return True

    async def _check_user_promotion_condition(self, promotion: PromotionActivity, user_id: str) -> bool:
        """检查用户参与促销的条件"""
        try:
            # TODO: 检查用户参与次数、会员等级等条件
            return True

        except Exception:
            return False

    async def _calculate_promotion_discount(self, 
                                          promotion: PromotionActivity, 
                                          current_price: float,
                                          quantity: int) -> float:
        """计算促销折扣"""
        try:
            if not promotion.discount_config:
                return 0.0

            discount_config = json.loads(promotion.discount_config)
            discount_type = discount_config.get("type", "percentage")
            discount_value = discount_config.get("value", 0)

            if discount_type == "percentage":
                # 百分比折扣
                discount = current_price * (discount_value / 100)
            elif discount_type == "fixed":
                # 固定金额折扣
                discount = discount_value
            else:
                discount = 0.0

            # 检查最大折扣限制
            max_discount = discount_config.get("max_discount")
            if max_discount:
                discount = min(discount, max_discount)

            return max(0, discount)

        except Exception as e:
            logger.error(f"计算促销折扣失败: {str(e)}")
            return 0.0

    # ==================== 价格历史管理 ====================

    async def _record_price_history(self,
                                   sku_id: str,
                                   price_type: str,
                                   old_price: Optional[Decimal],
                                   new_price: Optional[Decimal],
                                   change_reason: str,
                                   operator_id: Optional[str]) -> None:
        """记录价格历史"""
        try:
            if old_price == new_price:
                return  # 价格没有变化，不需要记录

            history = PriceHistory(
                sku_id=sku_id,
                price_type=price_type,
                old_price=old_price,
                new_price=new_price,
                change_reason=change_reason,
                change_type="manual" if operator_id else "system",
                effective_time=datetime.utcnow(),
                operator_id=operator_id
            )

            self.session.add(history)
            # 这里不提交，让调用者决定何时提交

        except Exception as e:
            logger.error(f"记录价格历史失败: {str(e)}")
            # 不抛出异常，避免影响主流程

    async def get_price_history(self, 
                               sku_id: str,
                               start_time: Optional[datetime] = None,
                               end_time: Optional[datetime] = None,
                               page: int = 1,
                               size: int = 20) -> Tuple[List[PriceHistory], int]:
        """获取价格历史"""
        try:
            query = select(PriceHistory).where(PriceHistory.sku_id == sku_id)
            
            # 添加时间过滤
            if start_time:
                query = query.where(PriceHistory.effective_time >= start_time)
            if end_time:
                query = query.where(PriceHistory.effective_time <= end_time)

            # 获取总数
            count_query = select(func.count(PriceHistory.history_id)).where(PriceHistory.sku_id == sku_id)
            if start_time:
                count_query = count_query.where(PriceHistory.effective_time >= start_time)
            if end_time:
                count_query = count_query.where(PriceHistory.effective_time <= end_time)
            
            total = await self.session.execute(count_query)
            total_count = total.scalar()

            # 分页查询
            query = query.order_by(PriceHistory.effective_time.desc())
            query = query.offset((page - 1) * size).limit(size)
            
            result = await self.session.execute(query)
            histories = result.scalars().all()

            return histories, total_count

        except Exception as e:
            logger.error(f"获取价格历史失败: {str(e)}")
            raise

    # ==================== 促销活动管理 ====================

    async def create_promotion_activity(self, activity_data: Dict[str, Any]) -> PromotionActivity:
        """创建促销活动"""
        try:
            # 检查活动编码是否重复
            existing = await self.session.execute(
                select(PromotionActivity).where(
                    PromotionActivity.activity_code == activity_data["activity_code"]
                )
            )
            if existing.scalar_one_or_none():
                raise BusinessException(f"促销活动编码 {activity_data['activity_code']} 已存在")

            activity = PromotionActivity(**activity_data)
            self.session.add(activity)
            await self.session.commit()
            await self.session.refresh(activity)

            logger.info(f"创建促销活动成功: {activity.activity_name}")
            return activity

        except Exception as e:
            await self.session.rollback()
            logger.error(f"创建促销活动失败: {str(e)}")
            raise

    async def get_promotion_activities(self,
                                     status: Optional[str] = None,
                                     is_active: Optional[bool] = None,
                                     page: int = 1,
                                     size: int = 20) -> Tuple[List[PromotionActivity], int]:
        """获取促销活动列表"""
        try:
            query = select(PromotionActivity)
            
            # 添加过滤条件
            conditions = []
            if status:
                conditions.append(PromotionActivity.status == status)
            if is_active is not None:
                conditions.append(PromotionActivity.is_active == is_active)
            
            if conditions:
                query = query.where(and_(*conditions))

            # 获取总数
            count_query = select(func.count(PromotionActivity.activity_id))
            if conditions:
                count_query = count_query.where(and_(*conditions))
            
            total = await self.session.execute(count_query)
            total_count = total.scalar()

            # 分页查询
            query = query.order_by(PromotionActivity.priority.desc(), PromotionActivity.created_at.desc())
            query = query.offset((page - 1) * size).limit(size)
            
            result = await self.session.execute(query)
            activities = result.scalars().all()

            return activities, total_count

        except Exception as e:
            logger.error(f"获取促销活动列表失败: {str(e)}")
            raise