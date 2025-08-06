"""
商品分类管理API

提供多级分类管理系统的完整API接口：
1. 分类CRUD操作
2. 分类树管理
3. 分类统计分析
4. 分类模板配置
"""

from typing import List, Dict, Any, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from pydantic import BaseModel, Field, validator

from app.core.database import get_async_session
from app.core.responses import success_response, error_response
from app.core.exceptions import BusinessException
from app.utils.auth import get_current_user
from app.services.category_service import CategoryService
from app.models.product_new import ProductCategory, AttributeType
from app.models.user import User

router = APIRouter(prefix="/categories", tags=["商品分类"])


# ==================== 请求模型 ====================

class CategoryCreateRequest(BaseModel):
    """创建分类请求"""
    name: str = Field(..., min_length=1, max_length=100, description="分类名称")
    parent_id: Optional[str] = Field(None, description="父分类ID")
    description: Optional[str] = Field(None, max_length=500, description="分类描述")
    icon: Optional[str] = Field(None, description="分类图标")
    banner_image: Optional[str] = Field(None, description="横幅图片")
    is_active: bool = Field(True, description="是否启用")
    is_visible: bool = Field(True, description="是否显示")
    sort_order: int = Field(0, description="排序顺序")
    seo_title: Optional[str] = Field(None, description="SEO标题")
    seo_description: Optional[str] = Field(None, description="SEO描述")
    seo_keywords: Optional[List[str]] = Field(None, description="SEO关键词")
    slug: Optional[str] = Field(None, description="URL slug")
    attribute_template: Optional[Dict] = Field(None, description="属性模板")
    
    class Config:
        schema_extra = {
            "example": {
                "name": "信封",
                "parent_id": None,
                "description": "各种类型的信封产品",
                "icon": "envelope-icon",
                "is_active": True,
                "is_visible": True,
                "sort_order": 1,
                "seo_title": "信封商品 - OpenPenPal",
                "seo_description": "精选各类信封，书写你的心意",
                "attribute_template": {
                    "attributes": [
                        {
                            "name": "材质",
                            "type": "basic",
                            "required": True,
                            "options": ["牛皮纸", "珠光纸", "艺术纸"]
                        }
                    ]
                }
            }
        }


class CategoryUpdateRequest(BaseModel):
    """更新分类请求"""
    name: Optional[str] = Field(None, min_length=1, max_length=100, description="分类名称")
    parent_id: Optional[str] = Field(None, description="父分类ID")
    description: Optional[str] = Field(None, max_length=500, description="分类描述")
    icon: Optional[str] = Field(None, description="分类图标")
    banner_image: Optional[str] = Field(None, description="横幅图片")
    is_active: Optional[bool] = Field(None, description="是否启用")
    is_visible: Optional[bool] = Field(None, description="是否显示")
    sort_order: Optional[int] = Field(None, description="排序顺序")
    seo_title: Optional[str] = Field(None, description="SEO标题")
    seo_description: Optional[str] = Field(None, description="SEO描述")
    seo_keywords: Optional[List[str]] = Field(None, description="SEO关键词")
    slug: Optional[str] = Field(None, description="URL slug")
    attribute_template: Optional[Dict] = Field(None, description="属性模板")


class CategoryReorderRequest(BaseModel):
    """分类重排序请求"""
    categories: List[Dict[str, Any]] = Field(..., description="分类排序列表")
    
    class Config:
        schema_extra = {
            "example": {
                "categories": [
                    {"id": "CAT001", "sort_order": 1},
                    {"id": "CAT002", "sort_order": 2},
                    {"id": "CAT003", "sort_order": 3}
                ]
            }
        }


class AttributeTemplateRequest(BaseModel):
    """属性模板设置请求"""
    template: Dict = Field(..., description="属性模板配置")
    
    @validator('template')
    def validate_template(cls, v):
        """验证模板格式"""
        if not isinstance(v, dict) or "attributes" not in v:
            raise ValueError("模板格式不正确")
        
        attributes = v["attributes"]
        if not isinstance(attributes, list):
            raise ValueError("属性列表格式不正确")
        
        for attr in attributes:
            if not isinstance(attr, dict):
                raise ValueError("属性定义格式不正确")
            
            if "name" not in attr or "type" not in attr:
                raise ValueError("属性必须包含name和type字段")
            
            if attr["type"] not in ["basic", "sale", "custom"]:
                raise ValueError("属性类型不正确")
        
        return v
    
    class Config:
        schema_extra = {
            "example": {
                "template": {
                    "attributes": [
                        {
                            "name": "材质",
                            "type": "basic",
                            "required": True,
                            "options": ["牛皮纸", "珠光纸", "艺术纸"],
                            "searchable": True,
                            "filterable": True,
                            "sort_order": 1
                        },
                        {
                            "name": "颜色",
                            "type": "sale", 
                            "required": True,
                            "options": ["白色", "米色", "棕色"],
                            "searchable": False,
                            "filterable": True,
                            "sort_order": 2
                        }
                    ]
                }
            }
        }


# ==================== 分类基础API ====================

@router.post("", response_model=dict)
async def create_category(
    request: CategoryCreateRequest,
    db: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """创建新分类"""
    try:
        service = CategoryService(db)
        
        # 验证属性模板格式
        if request.attribute_template:
            is_valid = await service.validate_attribute_template(request.attribute_template)
            if not is_valid:
                raise BusinessException("属性模板格式不正确")
        
        category = await service.create_category(
            name=request.name,
            parent_id=request.parent_id,
            description=request.description,
            icon=request.icon,
            banner_image=request.banner_image,
            is_active=request.is_active,
            is_visible=request.is_visible,
            sort_order=request.sort_order,
            seo_title=request.seo_title,
            seo_description=request.seo_description,
            seo_keywords=request.seo_keywords,
            slug=request.slug,
            attribute_template=request.attribute_template
        )
        
        return success_response(
            data=category.to_dict(),
            message="分类创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建分类失败: {str(e)}")


@router.get("/{category_id}")
async def get_category(
    category_id: str,
    db: AsyncSession = Depends(get_async_session)
):
    """获取单个分类详情"""
    try:
        service = CategoryService(db)
        category = await service.get_category(category_id)
        
        if not category:
            return error_response(message="分类不存在", code=404)
        
        return success_response(data=category.to_dict())
        
    except Exception as e:
        return error_response(message=f"获取分类失败: {str(e)}")


@router.put("/{category_id}")
async def update_category(
    category_id: str,
    request: CategoryUpdateRequest,
    db: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """更新分类信息"""
    try:
        service = CategoryService(db)
        
        # 过滤非空字段
        update_data = {k: v for k, v in request.dict().items() if v is not None}
        
        # 验证属性模板
        if "attribute_template" in update_data and update_data["attribute_template"]:
            is_valid = await service.validate_attribute_template(update_data["attribute_template"])
            if not is_valid:
                raise BusinessException("属性模板格式不正确")
        
        category = await service.update_category(category_id, **update_data)
        
        return success_response(
            data=category.to_dict(),
            message="分类更新成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"更新分类失败: {str(e)}")


@router.delete("/{category_id}")
async def delete_category(
    category_id: str,
    force: bool = Query(False, description="是否强制删除"),
    db: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """删除分类"""
    try:
        service = CategoryService(db)
        await service.delete_category(category_id, force=force)
        
        return success_response(message="分类删除成功")
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"删除分类失败: {str(e)}")


# ==================== 分类树管理API ====================

@router.get("/tree/full")
async def get_category_tree(
    include_inactive: bool = Query(False, description="是否包含已禁用的分类"),
    db: AsyncSession = Depends(get_async_session)
):
    """获取完整的分类树"""
    try:
        service = CategoryService(db)
        tree_nodes = await service.get_category_tree(include_inactive)
        
        # 转换为响应格式
        tree_data = [node.to_dict() for node in tree_nodes]
        
        return success_response(data={
            "tree": tree_data,
            "total_nodes": sum(1 for _ in service._category_cache.values())
        })
        
    except Exception as e:
        return error_response(message=f"获取分类树失败: {str(e)}")


@router.get("/{category_id}/path")
async def get_category_path(
    category_id: str,
    db: AsyncSession = Depends(get_async_session)
):
    """获取分类路径（面包屑导航）"""
    try:
        service = CategoryService(db)
        path_categories = await service.get_category_path(category_id)
        
        path_data = [
            {
                "id": cat.id,
                "name": cat.name,
                "level": cat.level,
                "slug": cat.slug
            }
            for cat in path_categories
        ]
        
        return success_response(data={
            "category_id": category_id,
            "path": path_data
        })
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"获取分类路径失败: {str(e)}")


@router.get("/{category_id}/children")
async def get_category_children(
    category_id: str,
    include_inactive: bool = Query(False, description="是否包含已禁用的分类"),
    db: AsyncSession = Depends(get_async_session)
):
    """获取分类的直接子分类"""
    try:
        service = CategoryService(db)
        
        # 处理根分类查询
        parent_id = category_id if category_id != "root" else None
        
        children = await service.get_children(parent_id, include_inactive)
        
        children_data = [
            {
                "id": child.id,
                "name": child.name,
                "description": child.description,
                "level": child.level,
                "icon": child.icon,
                "is_active": child.is_active,
                "is_visible": child.is_visible,
                "spu_count": child.spu_count,
                "sort_order": child.sort_order
            }
            for child in children
        ]
        
        return success_response(data={
            "parent_id": parent_id,
            "children": children_data
        })
        
    except Exception as e:
        return error_response(message=f"获取子分类失败: {str(e)}")


@router.get("/{category_id}/descendants")
async def get_category_descendants(
    category_id: str,
    include_self: bool = Query(False, description="是否包含自身"),
    db: AsyncSession = Depends(get_async_session)
):
    """获取分类的所有后代分类"""
    try:
        service = CategoryService(db)
        descendants = await service.get_descendants(category_id, include_self)
        
        descendants_data = [cat.to_dict() for cat in descendants]
        
        return success_response(data={
            "category_id": category_id,
            "include_self": include_self,
            "descendants": descendants_data
        })
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"获取后代分类失败: {str(e)}")


# ==================== 分类搜索API ====================

@router.get("/search")
async def search_categories(
    keyword: str = Query(..., description="搜索关键词"),
    parent_id: Optional[str] = Query(None, description="父分类ID，限制搜索范围"),
    db: AsyncSession = Depends(get_async_session)
):
    """搜索分类"""
    try:
        service = CategoryService(db)
        categories = await service.search_categories(keyword, parent_id)
        
        search_results = [
            {
                "id": cat.id,
                "name": cat.name,
                "description": cat.description,
                "path": cat.path,
                "level": cat.level,
                "spu_count": cat.spu_count
            }
            for cat in categories
        ]
        
        return success_response(data={
            "keyword": keyword,
            "parent_id": parent_id,
            "result_count": len(search_results),
            "results": search_results
        })
        
    except Exception as e:
        return error_response(message=f"搜索分类失败: {str(e)}")


# ==================== 分类统计API ====================

@router.get("/{category_id}/statistics")
async def get_category_statistics(
    category_id: str,
    db: AsyncSession = Depends(get_async_session)
):
    """获取分类统计信息"""
    try:
        service = CategoryService(db)
        
        # 获取分类信息
        category = await service.get_category(category_id)
        if not category:
            return error_response(message="分类不存在", code=404)
        
        # 获取后代分类
        descendants = await service.get_descendants(category_id, include_self=True)
        
        # 计算统计信息
        total_spu = sum(cat.spu_count for cat in descendants)
        total_sales = sum(cat.total_sales or 0 for cat in descendants)
        child_count = len([cat for cat in descendants if cat.parent_id == category_id])
        descendant_count = len(descendants) - 1  # 不包含自身
        
        return success_response(data={
            "category_id": category_id,
            "category_name": category.name,
            "statistics": {
                "spu_count": category.spu_count,
                "total_spu_including_descendants": total_spu,
                "total_sales": total_sales,
                "direct_children_count": child_count,
                "all_descendants_count": descendant_count,
                "level": category.level
            }
        })
        
    except Exception as e:
        return error_response(message=f"获取分类统计失败: {str(e)}")


@router.get("/analytics/popular")
async def get_popular_categories(
    limit: int = Query(10, ge=1, le=50, description="返回数量限制"),
    db: AsyncSession = Depends(get_async_session)
):
    """获取热门分类"""
    try:
        service = CategoryService(db)
        popular_categories = await service.get_popular_categories(limit)
        
        return success_response(data={
            "limit": limit,
            "categories": popular_categories
        })
        
    except Exception as e:
        return error_response(message=f"获取热门分类失败: {str(e)}")


# ==================== 分类管理API ====================

@router.post("/reorder")
async def reorder_categories(
    request: CategoryReorderRequest,
    db: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """批量更新分类排序"""
    try:
        service = CategoryService(db)
        await service.reorder_categories(request.categories)
        
        return success_response(
            data={"updated_count": len(request.categories)},
            message="分类排序更新成功"
        )
        
    except Exception as e:
        return error_response(message=f"更新分类排序失败: {str(e)}")


@router.post("/statistics/refresh")
async def refresh_category_statistics(
    db: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """刷新所有分类的统计信息"""
    try:
        service = CategoryService(db)
        await service.bulk_update_statistics()
        
        return success_response(message="分类统计信息刷新成功")
        
    except Exception as e:
        return error_response(message=f"刷新统计信息失败: {str(e)}")


@router.post("/{category_id}/statistics/update")
async def update_category_spu_count(
    category_id: str,
    db: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """更新单个分类的商品数量统计"""
    try:
        service = CategoryService(db)
        spu_count = await service.update_spu_count(category_id)
        
        return success_response(
            data={"category_id": category_id, "spu_count": spu_count},
            message="分类统计更新成功"
        )
        
    except Exception as e:
        return error_response(message=f"更新分类统计失败: {str(e)}")


# ==================== 属性模板API ====================

@router.get("/{category_id}/template")
async def get_category_template(
    category_id: str,
    db: AsyncSession = Depends(get_async_session)
):
    """获取分类的属性模板"""
    try:
        service = CategoryService(db)
        category = await service.get_category(category_id)
        
        if not category:
            return error_response(message="分类不存在", code=404)
        
        return success_response(data={
            "category_id": category_id,
            "category_name": category.name,
            "attribute_template": category.attribute_template
        })
        
    except Exception as e:
        return error_response(message=f"获取分类模板失败: {str(e)}")


@router.put("/{category_id}/template")
async def update_category_template(
    category_id: str,
    request: AttributeTemplateRequest,
    db: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """更新分类的属性模板"""
    try:
        service = CategoryService(db)
        
        # 验证模板格式
        is_valid = await service.validate_attribute_template(request.template)
        if not is_valid:
            raise BusinessException("属性模板格式不正确")
        
        category = await service.update_category(
            category_id, 
            attribute_template=request.template
        )
        
        return success_response(
            data={
                "category_id": category_id,
                "attribute_template": category.attribute_template
            },
            message="属性模板更新成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"更新属性模板失败: {str(e)}")