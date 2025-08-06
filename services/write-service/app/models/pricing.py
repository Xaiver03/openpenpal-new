"""
商品价格体系模型

支持多层次价格策略：
1. 基础定价
2. 会员等级定价
3. 批量定价
4. 促销定价
5. 动态定价
"""

from sqlalchemy import Column, Integer, String, DateTime, Boolean, Text, ForeignKey, Index
from sqlalchemy.types import DECIMAL
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from datetime import datetime
from enum import Enum
from typing import Dict, Any, Optional

from app.core.database import Base


class PriceType(Enum):
    """价格类型"""
    BASE = 1  # 基础价格
    MEMBER = 2  # 会员价格
    BULK = 3  # 批量价格
    PROMOTION = 4  # 促销价格
    DYNAMIC = 5  # 动态价格


class DiscountType(Enum):
    """折扣类型"""
    PERCENTAGE = 1  # 百分比折扣
    FIXED_AMOUNT = 2  # 固定金额折扣
    BUY_X_GET_Y = 3  # 买X送Y
    TIERED = 4  # 阶梯折扣


class PricePolicy(Base):
    """价格策略表"""
    __tablename__ = "price_policy"

    policy_id = Column(Integer, primary_key=True, autoincrement=True)
    policy_name = Column(String(100), nullable=False, comment="策略名称")
    policy_code = Column(String(50), unique=True, nullable=False, comment="策略编码")
    description = Column(Text, comment="策略描述")
    
    # 策略配置
    price_type = Column(Integer, default=PriceType.BASE.value, comment="价格类型")
    is_active = Column(Boolean, default=True, comment="是否启用")
    priority = Column(Integer, default=0, comment="优先级")
    
    # 适用范围
    apply_to_all = Column(Boolean, default=False, comment="是否适用所有商品")
    category_ids = Column(Text, comment="适用分类ID列表(JSON)")
    brand_ids = Column(Text, comment="适用品牌ID列表(JSON)")
    spu_ids = Column(Text, comment="适用SPU ID列表(JSON)")
    
    # 时间限制
    start_time = Column(DateTime, comment="生效开始时间")
    end_time = Column(DateTime, comment="生效结束时间")
    
    # 其他条件
    min_quantity = Column(Integer, default=1, comment="最小数量")
    max_quantity = Column(Integer, comment="最大数量")
    member_level_required = Column(String(50), comment="所需会员等级")
    
    # 审计字段
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    created_by = Column(String(50))
    updated_by = Column(String(50))

    # 关系
    price_rules = relationship("PriceRule", back_populates="policy")
    sku_prices = relationship("SkuPrice", back_populates="policy")

    def to_dict(self) -> Dict[str, Any]:
        return {
            "policy_id": self.policy_id,
            "policy_name": self.policy_name,
            "policy_code": self.policy_code,
            "description": self.description,
            "price_type": self.price_type,
            "is_active": self.is_active,
            "priority": self.priority,
            "apply_to_all": self.apply_to_all,
            "category_ids": self.category_ids,
            "brand_ids": self.brand_ids,
            "spu_ids": self.spu_ids,
            "start_time": self.start_time.isoformat() if self.start_time else None,
            "end_time": self.end_time.isoformat() if self.end_time else None,
            "min_quantity": self.min_quantity,
            "max_quantity": self.max_quantity,
            "member_level_required": self.member_level_required,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "created_by": self.created_by,
            "updated_by": self.updated_by
        }


class PriceRule(Base):
    """价格规则表"""
    __tablename__ = "price_rule"

    rule_id = Column(Integer, primary_key=True, autoincrement=True)
    policy_id = Column(Integer, ForeignKey("price_policy.policy_id"), nullable=False)
    rule_name = Column(String(100), nullable=False, comment="规则名称")
    
    # 条件配置
    condition_type = Column(String(50), nullable=False, comment="条件类型")
    condition_value = Column(Text, comment="条件值(JSON)")
    
    # 折扣配置
    discount_type = Column(Integer, default=DiscountType.PERCENTAGE.value, comment="折扣类型")
    discount_value = Column(DECIMAL(10, 4), comment="折扣值")
    max_discount_amount = Column(DECIMAL(10, 2), comment="最大折扣金额")
    
    # 规则状态
    is_active = Column(Boolean, default=True, comment="是否启用")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    
    # 审计字段
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    created_by = Column(String(50))
    updated_by = Column(String(50))

    # 关系
    policy = relationship("PricePolicy", back_populates="price_rules")

    def to_dict(self) -> Dict[str, Any]:
        return {
            "rule_id": self.rule_id,
            "policy_id": self.policy_id,
            "rule_name": self.rule_name,
            "condition_type": self.condition_type,
            "condition_value": self.condition_value,
            "discount_type": self.discount_type,
            "discount_value": float(self.discount_value) if self.discount_value else None,
            "max_discount_amount": float(self.max_discount_amount) if self.max_discount_amount else None,
            "is_active": self.is_active,
            "sort_order": self.sort_order,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "created_by": self.created_by,
            "updated_by": self.updated_by
        }


class SkuPrice(Base):
    """SKU价格表"""
    __tablename__ = "sku_price"

    price_id = Column(Integer, primary_key=True, autoincrement=True)
    sku_id = Column(String(50), nullable=False, comment="SKU ID")
    policy_id = Column(Integer, ForeignKey("price_policy.policy_id"), nullable=False)
    
    # 价格信息
    original_price = Column(DECIMAL(10, 2), nullable=False, comment="原价")
    current_price = Column(DECIMAL(10, 2), nullable=False, comment="当前价格")
    cost_price = Column(DECIMAL(10, 2), comment="成本价")
    
    # 价格区间
    min_price = Column(DECIMAL(10, 2), comment="最低价格")
    max_price = Column(DECIMAL(10, 2), comment="最高价格")
    
    # 会员价格
    vip_price = Column(DECIMAL(10, 2), comment="VIP价格")
    member_price = Column(DECIMAL(10, 2), comment="会员价格")
    
    # 批量价格配置
    bulk_config = Column(Text, comment="批量价格配置(JSON)")
    
    # 时效性
    effective_start = Column(DateTime, comment="生效开始时间")
    effective_end = Column(DateTime, comment="生效结束时间")
    
    # 状态
    is_active = Column(Boolean, default=True, comment="是否启用")
    
    # 审计字段
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    created_by = Column(String(50))
    updated_by = Column(String(50))

    # 关系
    policy = relationship("PricePolicy", back_populates="sku_prices")

    # 索引
    __table_args__ = (
        Index("idx_sku_price_sku_id", "sku_id"),
        Index("idx_sku_price_policy_id", "policy_id"),
        Index("idx_sku_price_active", "is_active"),
        Index("idx_sku_price_effective", "effective_start", "effective_end"),
    )

    def to_dict(self) -> Dict[str, Any]:
        return {
            "price_id": self.price_id,
            "sku_id": self.sku_id,
            "policy_id": self.policy_id,
            "original_price": float(self.original_price) if self.original_price else None,
            "current_price": float(self.current_price) if self.current_price else None,
            "cost_price": float(self.cost_price) if self.cost_price else None,
            "min_price": float(self.min_price) if self.min_price else None,
            "max_price": float(self.max_price) if self.max_price else None,
            "vip_price": float(self.vip_price) if self.vip_price else None,
            "member_price": float(self.member_price) if self.member_price else None,
            "bulk_config": self.bulk_config,
            "effective_start": self.effective_start.isoformat() if self.effective_start else None,
            "effective_end": self.effective_end.isoformat() if self.effective_end else None,
            "is_active": self.is_active,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "created_by": self.created_by,
            "updated_by": self.updated_by
        }


class PriceHistory(Base):
    """价格历史表"""
    __tablename__ = "price_history"

    history_id = Column(Integer, primary_key=True, autoincrement=True)
    sku_id = Column(String(50), nullable=False, comment="SKU ID")
    price_type = Column(String(50), nullable=False, comment="价格类型")
    
    # 价格变动信息
    old_price = Column(DECIMAL(10, 2), comment="变更前价格")
    new_price = Column(DECIMAL(10, 2), comment="变更后价格")
    change_reason = Column(String(200), comment="变更原因")
    change_type = Column(String(50), comment="变更类型")
    
    # 生效信息
    effective_time = Column(DateTime, nullable=False, comment="生效时间")
    operator_id = Column(String(50), comment="操作员ID")
    operator_name = Column(String(100), comment="操作员姓名")
    
    # 审计字段
    created_at = Column(DateTime, default=func.now())

    # 索引
    __table_args__ = (
        Index("idx_price_history_sku_id", "sku_id"),
        Index("idx_price_history_effective_time", "effective_time"),
        Index("idx_price_history_operator", "operator_id"),
    )

    def to_dict(self) -> Dict[str, Any]:
        return {
            "history_id": self.history_id,
            "sku_id": self.sku_id,
            "price_type": self.price_type,
            "old_price": float(self.old_price) if self.old_price else None,
            "new_price": float(self.new_price) if self.new_price else None,
            "change_reason": self.change_reason,
            "change_type": self.change_type,
            "effective_time": self.effective_time.isoformat() if self.effective_time else None,
            "operator_id": self.operator_id,
            "operator_name": self.operator_name,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }


class PromotionActivity(Base):
    """促销活动表"""
    __tablename__ = "promotion_activity"

    activity_id = Column(Integer, primary_key=True, autoincrement=True)
    activity_name = Column(String(100), nullable=False, comment="活动名称")
    activity_code = Column(String(50), unique=True, nullable=False, comment="活动编码")
    activity_type = Column(String(50), nullable=False, comment="活动类型")
    description = Column(Text, comment="活动描述")
    
    # 活动时间
    start_time = Column(DateTime, nullable=False, comment="活动开始时间")
    end_time = Column(DateTime, nullable=False, comment="活动结束时间")
    
    # 活动配置
    discount_config = Column(Text, comment="折扣配置(JSON)")
    target_config = Column(Text, comment="目标商品配置(JSON)")
    condition_config = Column(Text, comment="参与条件配置(JSON)")
    
    # 活动限制
    max_participants = Column(Integer, comment="最大参与人数")
    current_participants = Column(Integer, default=0, comment="当前参与人数")
    max_usage_per_user = Column(Integer, default=1, comment="每人最大使用次数")
    
    # 活动状态
    status = Column(String(20), default="draft", comment="活动状态")
    is_active = Column(Boolean, default=False, comment="是否启用")
    priority = Column(Integer, default=0, comment="优先级")
    
    # 统计信息
    view_count = Column(Integer, default=0, comment="查看次数")
    participate_count = Column(Integer, default=0, comment="参与次数")
    order_count = Column(Integer, default=0, comment="订单数量")
    total_amount = Column(DECIMAL(12, 2), default=0, comment="总交易金额")
    
    # 审计字段
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    created_by = Column(String(50))
    updated_by = Column(String(50))

    # 索引
    __table_args__ = (
        Index("idx_promotion_activity_code", "activity_code"),
        Index("idx_promotion_activity_time", "start_time", "end_time"),
        Index("idx_promotion_activity_status", "status", "is_active"),
    )

    def to_dict(self) -> Dict[str, Any]:
        return {
            "activity_id": self.activity_id,
            "activity_name": self.activity_name,
            "activity_code": self.activity_code,
            "activity_type": self.activity_type,
            "description": self.description,
            "start_time": self.start_time.isoformat() if self.start_time else None,
            "end_time": self.end_time.isoformat() if self.end_time else None,
            "discount_config": self.discount_config,
            "target_config": self.target_config,
            "condition_config": self.condition_config,
            "max_participants": self.max_participants,
            "current_participants": self.current_participants,
            "max_usage_per_user": self.max_usage_per_user,
            "status": self.status,
            "is_active": self.is_active,
            "priority": self.priority,
            "view_count": self.view_count,
            "participate_count": self.participate_count,
            "order_count": self.order_count,
            "total_amount": float(self.total_amount) if self.total_amount else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "created_by": self.created_by,
            "updated_by": self.updated_by
        }