"""
价格管理API

提供完整的价格体系管理功能：
1. 价格策略管理
2. SKU价格设置
3. 价格计算引擎
4. 促销活动管理
5. 价格历史查询
"""

from typing import List, Dict, Any, Optional
from decimal import Decimal
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from pydantic import BaseModel, Field

from app.core.database import get_async_session
from app.core.auth import get_current_user
from app.core.responses import success_response, error_response
from app.core.exceptions import BusinessException
from app.services.pricing_service import PricingService
from app.models.pricing import PriceType, DiscountType
from app.models.user import User

router = APIRouter(prefix="/pricing", tags=["价格管理"])


# ==================== 请求模型 ====================

class PricePolicyCreateRequest(BaseModel):
    """创建价格策略请求"""
    policy_name: str = Field(..., description="策略名称")
    policy_code: str = Field(..., description="策略编码")
    description: Optional[str] = Field(None, description="策略描述")
    price_type: int = Field(PriceType.BASE.value, description="价格类型")
    is_active: bool = Field(True, description="是否启用")
    priority: int = Field(0, description="优先级")
    apply_to_all: bool = Field(False, description="是否适用所有商品")
    category_ids: Optional[str] = Field(None, description="适用分类ID列表")
    brand_ids: Optional[str] = Field(None, description="适用品牌ID列表")
    spu_ids: Optional[str] = Field(None, description="适用SPU ID列表")
    start_time: Optional[datetime] = Field(None, description="生效开始时间")
    end_time: Optional[datetime] = Field(None, description="生效结束时间")
    min_quantity: int = Field(1, description="最小数量")
    max_quantity: Optional[int] = Field(None, description="最大数量")
    member_level_required: Optional[str] = Field(None, description="所需会员等级")


class PricePolicyUpdateRequest(BaseModel):
    """更新价格策略请求"""
    policy_name: Optional[str] = None
    description: Optional[str] = None
    is_active: Optional[bool] = None
    priority: Optional[int] = None
    apply_to_all: Optional[bool] = None
    category_ids: Optional[str] = None
    brand_ids: Optional[str] = None
    spu_ids: Optional[str] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    min_quantity: Optional[int] = None
    max_quantity: Optional[int] = None
    member_level_required: Optional[str] = None


class PriceRuleCreateRequest(BaseModel):
    """创建价格规则请求"""
    policy_id: int = Field(..., description="价格策略ID")
    rule_name: str = Field(..., description="规则名称")
    condition_type: str = Field(..., description="条件类型")
    condition_value: Optional[str] = Field(None, description="条件值")
    discount_type: int = Field(DiscountType.PERCENTAGE.value, description="折扣类型")
    discount_value: Optional[Decimal] = Field(None, description="折扣值")
    max_discount_amount: Optional[Decimal] = Field(None, description="最大折扣金额")
    is_active: bool = Field(True, description="是否启用")
    sort_order: int = Field(0, description="排序顺序")


class SkuPriceSetRequest(BaseModel):
    """SKU价格设置请求"""
    sku_id: str = Field(..., description="SKU ID")
    policy_id: int = Field(..., description="价格策略ID")
    original_price: Decimal = Field(..., description="原价")
    current_price: Decimal = Field(..., description="当前价格")
    cost_price: Optional[Decimal] = Field(None, description="成本价")
    min_price: Optional[Decimal] = Field(None, description="最低价格")
    max_price: Optional[Decimal] = Field(None, description="最高价格")
    vip_price: Optional[Decimal] = Field(None, description="VIP价格")
    member_price: Optional[Decimal] = Field(None, description="会员价格")
    bulk_config: Optional[str] = Field(None, description="批量价格配置")
    effective_start: Optional[datetime] = Field(None, description="生效开始时间")
    effective_end: Optional[datetime] = Field(None, description="生效结束时间")


class BatchSkuPriceSetRequest(BaseModel):
    """批量SKU价格设置请求"""
    prices: List[SkuPriceSetRequest] = Field(..., description="价格列表")


class PriceCalculateRequest(BaseModel):
    """价格计算请求"""
    sku_id: str = Field(..., description="SKU ID")
    quantity: int = Field(1, ge=1, description="购买数量")
    user_id: Optional[str] = Field(None, description="用户ID")
    member_level: Optional[str] = Field(None, description="会员等级")


class PromotionActivityCreateRequest(BaseModel):
    """创建促销活动请求"""
    activity_name: str = Field(..., description="活动名称")
    activity_code: str = Field(..., description="活动编码")
    activity_type: str = Field(..., description="活动类型")
    description: Optional[str] = Field(None, description="活动描述")
    start_time: datetime = Field(..., description="活动开始时间")
    end_time: datetime = Field(..., description="活动结束时间")
    discount_config: Optional[str] = Field(None, description="折扣配置")
    target_config: Optional[str] = Field(None, description="目标商品配置")
    condition_config: Optional[str] = Field(None, description="参与条件配置")
    max_participants: Optional[int] = Field(None, description="最大参与人数")
    max_usage_per_user: int = Field(1, description="每人最大使用次数")
    priority: int = Field(0, description="优先级")


# ==================== 价格策略管理API ====================

@router.post("/policies")
async def create_price_policy(
    request: PricePolicyCreateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """创建价格策略"""
    try:
        service = PricingService(session)
        
        policy_data = request.dict()
        policy_data.update({
            "created_by": current_user.user_id,
            "updated_by": current_user.user_id
        })
        
        policy = await service.create_price_policy(policy_data)
        
        return success_response(
            data=policy.to_dict(),
            message="价格策略创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建价格策略失败: {str(e)}")


@router.put("/policies/{policy_id}")
async def update_price_policy(
    policy_id: int,
    request: PricePolicyUpdateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """更新价格策略"""
    try:
        service = PricingService(session)
        
        policy_data = {k: v for k, v in request.dict().items() if v is not None}
        policy_data["updated_by"] = current_user.user_id
        
        policy = await service.update_price_policy(policy_id, policy_data)
        
        return success_response(
            data=policy.to_dict(),
            message="价格策略更新成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"更新价格策略失败: {str(e)}")


@router.delete("/policies/{policy_id}")
async def delete_price_policy(
    policy_id: int,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """删除价格策略"""
    try:
        service = PricingService(session)
        await service.delete_price_policy(policy_id)
        
        return success_response(message="价格策略删除成功")
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"删除价格策略失败: {str(e)}")


@router.get("/policies")
async def get_price_policies(
    price_type: Optional[int] = Query(None, description="价格类型"),
    is_active: Optional[bool] = Query(None, description="是否启用"),
    page: int = Query(1, ge=1, description="页码"),
    size: int = Query(20, ge=1, le=100, description="每页大小"),
    session: AsyncSession = Depends(get_async_session)
):
    """获取价格策略列表"""
    try:
        service = PricingService(session)
        policies, total = await service.get_price_policies(price_type, is_active, page, size)
        
        return success_response(data={
            "policies": [policy.to_dict() for policy in policies],
            "pagination": {
                "page": page,
                "size": size,
                "total": total,
                "pages": (total + size - 1) // size
            }
        })
        
    except Exception as e:
        return error_response(message=f"获取价格策略列表失败: {str(e)}")


# ==================== 价格规则管理API ====================

@router.post("/rules")
async def create_price_rule(
    request: PriceRuleCreateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """创建价格规则"""
    try:
        service = PricingService(session)
        
        rule_data = request.dict()
        rule_data.update({
            "created_by": current_user.user_id,
            "updated_by": current_user.user_id
        })
        
        rule = await service.create_price_rule(rule_data)
        
        return success_response(
            data=rule.to_dict(),
            message="价格规则创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建价格规则失败: {str(e)}")


@router.get("/policies/{policy_id}/rules")
async def get_policy_rules(
    policy_id: int,
    session: AsyncSession = Depends(get_async_session)
):
    """获取策略的价格规则"""
    try:
        service = PricingService(session)
        rules = await service.get_policy_rules(policy_id)
        
        return success_response(data={
            "policy_id": policy_id,
            "rules": [rule.to_dict() for rule in rules]
        })
        
    except Exception as e:
        return error_response(message=f"获取价格规则失败: {str(e)}")


# ==================== SKU价格管理API ====================

@router.post("/sku-prices")
async def set_sku_price(
    request: SkuPriceSetRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """设置SKU价格"""
    try:
        service = PricingService(session)
        
        price_data = request.dict()
        price_data.update({
            "created_by": current_user.user_id,
            "updated_by": current_user.user_id
        })
        
        price = await service.set_sku_price(price_data)
        
        return success_response(
            data=price.to_dict(),
            message="SKU价格设置成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"设置SKU价格失败: {str(e)}")


@router.post("/sku-prices/batch")
async def batch_set_sku_prices(
    request: BatchSkuPriceSetRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """批量设置SKU价格"""
    try:
        service = PricingService(session)
        
        prices_data = []
        for price_request in request.prices:
            price_data = price_request.dict()
            price_data.update({
                "created_by": current_user.user_id,
                "updated_by": current_user.user_id
            })
            prices_data.append(price_data)
        
        prices = await service.batch_set_sku_prices(prices_data)
        
        return success_response(
            data={
                "created_count": len(prices),
                "prices": [price.to_dict() for price in prices]
            },
            message="批量设置SKU价格成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"批量设置SKU价格失败: {str(e)}")


@router.get("/sku-prices/{sku_id}")
async def get_sku_price(
    sku_id: str,
    policy_id: Optional[int] = Query(None, description="价格策略ID"),
    session: AsyncSession = Depends(get_async_session)
):
    """获取SKU价格"""
    try:
        service = PricingService(session)
        price = await service.get_sku_price(sku_id, policy_id)
        
        if not price:
            return error_response(message="SKU价格不存在")
        
        return success_response(data=price.to_dict())
        
    except Exception as e:
        return error_response(message=f"获取SKU价格失败: {str(e)}")


# ==================== 价格计算API ====================

@router.post("/calculate")
async def calculate_price(
    request: PriceCalculateRequest,
    session: AsyncSession = Depends(get_async_session)
):
    """价格计算"""
    try:
        service = PricingService(session)
        
        result = await service.calculate_price(
            request.sku_id,
            request.quantity,
            request.user_id,
            request.member_level
        )
        
        return success_response(data=result)
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"价格计算失败: {str(e)}")


@router.post("/calculate/batch")
async def batch_calculate_price(
    requests: List[PriceCalculateRequest],
    session: AsyncSession = Depends(get_async_session)
):
    """批量价格计算"""
    try:
        service = PricingService(session)
        results = []
        
        for request in requests:
            try:
                result = await service.calculate_price(
                    request.sku_id,
                    request.quantity,
                    request.user_id,
                    request.member_level
                )
                results.append(result)
            except Exception as e:
                results.append({
                    "sku_id": request.sku_id,
                    "error": str(e)
                })
        
        return success_response(data={
            "results": results,
            "total_count": len(results)
        })
        
    except Exception as e:
        return error_response(message=f"批量价格计算失败: {str(e)}")


# ==================== 价格历史API ====================

@router.get("/history/{sku_id}")
async def get_price_history(
    sku_id: str,
    start_time: Optional[datetime] = Query(None, description="开始时间"),
    end_time: Optional[datetime] = Query(None, description="结束时间"),
    page: int = Query(1, ge=1, description="页码"),
    size: int = Query(20, ge=1, le=100, description="每页大小"),
    session: AsyncSession = Depends(get_async_session)
):
    """获取价格历史"""
    try:
        service = PricingService(session)
        histories, total = await service.get_price_history(sku_id, start_time, end_time, page, size)
        
        return success_response(data={
            "sku_id": sku_id,
            "histories": [history.to_dict() for history in histories],
            "pagination": {
                "page": page,
                "size": size,
                "total": total,
                "pages": (total + size - 1) // size
            }
        })
        
    except Exception as e:
        return error_response(message=f"获取价格历史失败: {str(e)}")


# ==================== 促销活动管理API ====================

@router.post("/promotions")
async def create_promotion_activity(
    request: PromotionActivityCreateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """创建促销活动"""
    try:
        service = PricingService(session)
        
        activity_data = request.dict()
        activity_data.update({
            "created_by": current_user.user_id,
            "updated_by": current_user.user_id
        })
        
        activity = await service.create_promotion_activity(activity_data)
        
        return success_response(
            data=activity.to_dict(),
            message="促销活动创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建促销活动失败: {str(e)}")


@router.get("/promotions")
async def get_promotion_activities(
    status: Optional[str] = Query(None, description="活动状态"),
    is_active: Optional[bool] = Query(None, description="是否启用"),
    page: int = Query(1, ge=1, description="页码"),
    size: int = Query(20, ge=1, le=100, description="每页大小"),
    session: AsyncSession = Depends(get_async_session)
):
    """获取促销活动列表"""
    try:
        service = PricingService(session)
        activities, total = await service.get_promotion_activities(status, is_active, page, size)
        
        return success_response(data={
            "activities": [activity.to_dict() for activity in activities],
            "pagination": {
                "page": page,
                "size": size,
                "total": total,
                "pages": (total + size - 1) // size
            }
        })
        
    except Exception as e:
        return error_response(message=f"获取促销活动列表失败: {str(e)}")


# ==================== 价格管理工具API ====================

@router.get("/tools/price-suggestion/{sku_id}")
async def get_price_suggestion(
    sku_id: str,
    session: AsyncSession = Depends(get_async_session)
):
    """获取价格建议"""
    try:
        # TODO: 实现价格建议算法
        # 基于历史价格、市场价格、成本等因素给出定价建议
        
        return success_response(data={
            "sku_id": sku_id,
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
        })
        
    except Exception as e:
        return error_response(message=f"获取价格建议失败: {str(e)}")


@router.post("/tools/price-check")
async def check_price_competitiveness(
    sku_id: str,
    proposed_price: Decimal,
    session: AsyncSession = Depends(get_async_session)
):
    """检查价格竞争力"""
    try:
        # TODO: 实现价格竞争力分析
        # 与竞品价格对比，分析定价是否合理
        
        return success_response(data={
            "sku_id": sku_id,
            "proposed_price": float(proposed_price),
            "competitiveness": "moderate",
            "market_position": "middle",
            "recommendations": [
                "价格处于市场中等水平",
                "可考虑适当提高价格以增加利润空间"
            ]
        })
        
    except Exception as e:
        return error_response(message=f"检查价格竞争力失败: {str(e)}")