"""
商品属性管理API

提供商品属性系统的完整API接口：
1. 属性模板管理
2. SPU属性CRUD
3. SKU自动生成
4. 属性搜索和筛选
"""

from typing import List, Dict, Any, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from pydantic import BaseModel, Field

from app.core.database import get_async_session
from app.core.auth import get_current_user
from app.core.responses import success_response, error_response
from app.core.exceptions import BusinessException
from app.services.product_attribute_service import ProductAttributeService
from app.models.product_new import AttributeType, ProductAttribute
from app.models.user import User

router = APIRouter(prefix="/product-attributes", tags=["商品属性"])


# ==================== 请求模型 ====================

class AttributeCreateRequest(BaseModel):
    """创建属性请求"""
    name: str = Field(..., description="属性名称")
    value: str = Field(..., description="属性值")
    type: str = Field(AttributeType.BASIC.value, description="属性类型")
    required: bool = Field(False, description="是否必填")
    searchable: bool = Field(False, description="是否可搜索")
    filterable: bool = Field(False, description="是否可筛选")
    sort_order: int = Field(0, description="排序顺序")
    options: Optional[List[str]] = Field(None, description="可选值列表")


class AttributeUpdateRequest(BaseModel):
    """更新属性请求"""
    value: Optional[str] = Field(None, description="属性值")
    searchable: Optional[bool] = Field(None, description="是否可搜索")
    filterable: Optional[bool] = Field(None, description="是否可筛选")
    sort_order: Optional[int] = Field(None, description="排序顺序")


class SKUGenerateRequest(BaseModel):
    """SKU生成请求"""
    auto_pricing: bool = Field(False, description="是否自动定价")
    base_price: Optional[float] = Field(None, description="基础价格")
    price_rules: Optional[Dict[str, float]] = Field(None, description="价格规则")
    stock_quantity: int = Field(0, description="默认库存")


class SKUBatchCreateRequest(BaseModel):
    """批量创建SKU请求"""
    skus: List[Dict[str, Any]] = Field(..., description="SKU列表")
    
    class Config:
        schema_extra = {
            "example": {
                "skus": [
                    {
                        "name": "红色-A4",
                        "price": 25.0,
                        "stock_quantity": 100,
                        "sale_attributes": {"color": "红色", "size": "A4"}
                    },
                    {
                        "name": "蓝色-A5", 
                        "price": 20.0,
                        "stock_quantity": 50,
                        "sale_attributes": {"color": "蓝色", "size": "A5"}
                    }
                ]
            }
        }


class AttributeFilterRequest(BaseModel):
    """属性筛选请求"""
    attributes: Dict[str, List[str]] = Field(..., description="属性筛选条件")
    category_id: Optional[str] = Field(None, description="分类ID")
    
    class Config:
        schema_extra = {
            "example": {
                "attributes": {
                    "颜色": ["红色", "蓝色"],
                    "尺寸": ["A4", "A5"]
                },
                "category_id": "CAT001"
            }
        }


# ==================== 属性模板API ====================

@router.get("/templates/{category_id}")
async def get_category_attribute_templates(
    category_id: str,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取分类的属性模板"""
    try:
        service = ProductAttributeService(session)
        templates = await service.get_category_templates(category_id)
        
        # 转换为响应格式
        template_data = []
        for template in templates:
            template_data.append({
                "name": template.name,
                "type": template.attribute_type,
                "required": template.required,
                "options": template.options,
                "default_value": template.default_value,
                "validation_rules": template.validation_rules,
                "searchable": template.is_searchable,
                "filterable": template.is_filterable,
                "sort_order": template.sort_order
            })
        
        return success_response(data={
            "category_id": category_id,
            "templates": template_data
        })
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"获取属性模板失败: {str(e)}")


# ==================== SPU属性API ====================

@router.get("/spu/{spu_id}")
async def get_spu_attributes(
    spu_id: str,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取SPU的所有属性"""
    try:
        service = ProductAttributeService(session)
        attributes = await service.get_spu_attributes(spu_id)
        
        # 转换格式
        result = {}
        for attr_type, attr_list in attributes.items():
            result[attr_type] = [
                {
                    "id": attr.id,
                    "name": attr.attribute_name,
                    "value": attr.attribute_value,
                    "type": attr.attribute_type,
                    "required": attr.is_required,
                    "searchable": attr.is_searchable,
                    "filterable": attr.is_filterable,
                    "sort_order": attr.sort_order,
                    "options": attr.options,
                    "created_at": attr.created_at.isoformat() if attr.created_at else None
                }
                for attr in attr_list
            ]
        
        return success_response(data={
            "spu_id": spu_id,
            "attributes": result
        })
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"获取SPU属性失败: {str(e)}")


@router.post("/spu/{spu_id}")
async def create_spu_attributes(
    spu_id: str,
    attributes: List[AttributeCreateRequest],
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """为SPU创建属性"""
    try:
        service = ProductAttributeService(session)
        
        # 转换请求数据
        attr_data = []
        for attr in attributes:
            attr_data.append(attr.dict())
        
        created_attributes = await service.create_spu_attributes(spu_id, attr_data)
        
        return success_response(
            data={
                "spu_id": spu_id,
                "created_count": len(created_attributes),
                "attributes": [
                    {
                        "id": attr.id,
                        "name": attr.attribute_name,
                        "value": attr.attribute_value,
                        "type": attr.attribute_type
                    }
                    for attr in created_attributes
                ]
            },
            message=f"成功创建 {len(created_attributes)} 个属性"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建SPU属性失败: {str(e)}")


@router.put("/{attribute_id}")
async def update_attribute(
    attribute_id: str,
    data: AttributeUpdateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """更新属性"""
    try:
        service = ProductAttributeService(session)
        
        # 过滤非空字段
        update_data = {k: v for k, v in data.dict().items() if v is not None}
        
        updated_attribute = await service.update_attribute(attribute_id, update_data)
        
        return success_response(
            data={
                "id": updated_attribute.id,
                "name": updated_attribute.attribute_name,
                "value": updated_attribute.attribute_value,
                "type": updated_attribute.attribute_type,
                "updated_at": updated_attribute.updated_at.isoformat()
            },
            message="属性更新成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"更新属性失败: {str(e)}")


@router.delete("/{attribute_id}")
async def delete_attribute(
    attribute_id: str,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """删除属性"""
    try:
        service = ProductAttributeService(session)
        await service.delete_attribute(attribute_id)
        
        return success_response(message="属性删除成功")
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"删除属性失败: {str(e)}")


# ==================== SKU生成API ====================

@router.get("/spu/{spu_id}/sku-combinations")
async def get_sku_combinations(
    spu_id: str,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取根据销售属性生成的SKU组合"""
    try:
        service = ProductAttributeService(session)
        combinations = await service.generate_sku_combinations(spu_id)
        
        return success_response(data={
            "spu_id": spu_id,
            "combination_count": len(combinations),
            "combinations": combinations
        })
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"生成SKU组合失败: {str(e)}")


@router.post("/spu/{spu_id}/generate-skus")
async def generate_skus_from_attributes(
    spu_id: str,
    request: SKUGenerateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """根据销售属性自动生成SKU"""
    try:
        service = ProductAttributeService(session)
        
        # 获取属性组合
        combinations = await service.generate_sku_combinations(spu_id)
        
        # 生成SKU数据
        sku_data = []
        for idx, combo in enumerate(combinations):
            price = request.base_price or 0
            
            # 应用价格规则
            if request.price_rules:
                for attr_name, attr_value in combo["sale_attributes"].items():
                    rule_key = f"{attr_name}:{attr_value}"
                    if rule_key in request.price_rules:
                        price += request.price_rules[rule_key]
            
            sku_data.append({
                "name": combo["name"],
                "price": price,
                "stock_quantity": request.stock_quantity,
                "sale_attributes": combo["sale_attributes"]
            })
        
        # 创建SKU
        created_skus = await service.create_skus_from_attributes(spu_id, sku_data)
        
        return success_response(
            data={
                "spu_id": spu_id,
                "created_count": len(created_skus),
                "skus": [
                    {
                        "id": sku.id,
                        "sku_code": sku.sku_code,
                        "name": sku.name,
                        "price": sku.price,
                        "sale_attributes": sku.sale_attributes
                    }
                    for sku in created_skus
                ]
            },
            message=f"成功生成 {len(created_skus)} 个SKU"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"生成SKU失败: {str(e)}")


@router.post("/spu/{spu_id}/create-skus")
async def create_skus_batch(
    spu_id: str,
    request: SKUBatchCreateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """批量创建SKU"""
    try:
        service = ProductAttributeService(session)
        
        # 验证每个SKU的属性
        for sku_data in request.skus:
            if "sale_attributes" in sku_data:
                valid = await service.validate_sku_attributes(
                    spu_id, sku_data["sale_attributes"]
                )
                if not valid:
                    raise BusinessException(f"SKU属性验证失败: {sku_data['name']}")
        
        # 创建SKU
        created_skus = await service.create_skus_from_attributes(spu_id, request.skus)
        
        return success_response(
            data={
                "spu_id": spu_id,
                "created_count": len(created_skus),
                "skus": [
                    {
                        "id": sku.id,
                        "sku_code": sku.sku_code,
                        "name": sku.name,
                        "price": sku.price,
                        "stock_quantity": sku.stock_quantity,
                        "sale_attributes": sku.sale_attributes
                    }
                    for sku in created_skus
                ]
            },
            message=f"成功创建 {len(created_skus)} 个SKU"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"批量创建SKU失败: {str(e)}")


# ==================== 属性筛选API ====================

@router.get("/filter-options")
async def get_filter_options(
    category_id: Optional[str] = Query(None, description="分类ID"),
    session: AsyncSession = Depends(get_async_session)
):
    """获取可用于筛选的属性选项"""
    try:
        service = ProductAttributeService(session)
        filter_options = await service.get_filterable_attributes(category_id)
        
        return success_response(data={
            "category_id": category_id,
            "filter_options": filter_options
        })
        
    except Exception as e:
        return error_response(message=f"获取筛选选项失败: {str(e)}")


@router.post("/search")
async def search_by_attributes(
    request: AttributeFilterRequest,
    session: AsyncSession = Depends(get_async_session)
):
    """根据属性搜索商品"""
    try:
        service = ProductAttributeService(session)
        spu_ids = await service.search_by_attributes(
            request.attributes, 
            request.category_id
        )
        
        return success_response(data={
            "filter_conditions": request.attributes,
            "category_id": request.category_id,
            "result_count": len(spu_ids),
            "spu_ids": spu_ids
        })
        
    except Exception as e:
        return error_response(message=f"属性搜索失败: {str(e)}")


# ==================== 属性验证API ====================

@router.post("/validate-sku")
async def validate_sku_attributes(
    spu_id: str,
    sale_attributes: Dict[str, str],
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """验证SKU的销售属性是否合法"""
    try:
        service = ProductAttributeService(session)
        is_valid = await service.validate_sku_attributes(spu_id, sale_attributes)
        
        return success_response(data={
            "spu_id": spu_id,
            "sale_attributes": sale_attributes,
            "valid": is_valid
        })
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"属性验证失败: {str(e)}")


# ==================== 统计API ====================

@router.get("/statistics")
async def get_attribute_statistics(
    category_id: Optional[str] = Query(None, description="分类ID"),
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取属性使用统计"""
    try:
        from sqlalchemy import func, select
        from app.models.product_new import ProductSPU
        
        # 统计各类型属性数量
        query = select(
            ProductAttribute.attribute_type,
            func.count(ProductAttribute.id).label('count')
        ).group_by(ProductAttribute.attribute_type)
        
        if category_id:
            query = query.join(ProductSPU).where(ProductSPU.category_id == category_id)
        
        result = await session.execute(query)
        type_stats = {row.attribute_type: row.count for row in result.fetchall()}
        
        # 统计最常用的属性名
        name_query = select(
            ProductAttribute.attribute_name,
            func.count(ProductAttribute.id).label('count')
        ).group_by(ProductAttribute.attribute_name).order_by(func.count(ProductAttribute.id).desc()).limit(10)
        
        if category_id:
            name_query = name_query.join(ProductSPU).where(ProductSPU.category_id == category_id)
        
        result = await session.execute(name_query)
        name_stats = [{"name": row.attribute_name, "count": row.count} for row in result.fetchall()]
        
        return success_response(data={
            "category_id": category_id,
            "type_statistics": type_stats,
            "popular_attributes": name_stats
        })
        
    except Exception as e:
        return error_response(message=f"获取统计信息失败: {str(e)}")