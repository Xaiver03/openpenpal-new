"""
商品分类管理服务

实现多级分类管理系统的核心功能：
1. 分类树管理 - 支持无限级分类
2. 分类模板系统 - 每个分类可配置属性模板
3. 分类统计和分析
4. 分类导航和面包屑
5. 分类排序和权限管理
"""

from typing import List, Dict, Any, Optional, Tuple, Set
from datetime import datetime
from sqlalchemy import select, and_, or_, func, text
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload, joinedload
import uuid
from collections import defaultdict, deque

from app.models.product_new import ProductCategory, ProductSPU, AttributeType
from app.core.exceptions import BusinessException
from app.core.logger import logger


class CategoryNode:
    """分类树节点"""
    def __init__(self, category: ProductCategory):
        self.category = category
        self.children: List[CategoryNode] = []
        self.parent: Optional[CategoryNode] = None
        self._depth = 0
    
    @property
    def depth(self) -> int:
        """获取节点深度"""
        return self._depth
    
    @depth.setter
    def depth(self, value: int):
        """设置节点深度"""
        self._depth = value
        for child in self.children:
            child.depth = value + 1
    
    def add_child(self, child_node: 'CategoryNode'):
        """添加子节点"""
        child_node.parent = self
        child_node.depth = self._depth + 1
        self.children.append(child_node)
    
    def to_dict(self, include_children: bool = True) -> Dict:
        """转换为字典格式"""
        data = {
            "id": self.category.id,
            "name": self.category.name,
            "description": self.category.description,
            "parent_id": self.category.parent_id,
            "path": self.category.path,
            "level": self.category.level,
            "depth": self.depth,
            "icon": self.category.icon,
            "banner_image": self.category.banner_image,
            "is_active": self.category.is_active,
            "is_visible": self.category.is_visible,
            "sort_order": self.category.sort_order,
            "spu_count": self.category.spu_count,
            "total_sales": self.category.total_sales,
            "attribute_template": self.category.attribute_template,
            "created_at": self.category.created_at.isoformat() if self.category.created_at else None,
            "updated_at": self.category.updated_at.isoformat() if self.category.updated_at else None,
        }
        
        if include_children and self.children:
            data["children"] = [child.to_dict(include_children) for child in self.children]
        
        return data


class CategoryService:
    """商品分类服务"""
    
    def __init__(self, session: AsyncSession):
        self.session = session
        self._category_cache: Dict[str, CategoryNode] = {}
        self._tree_cache: List[CategoryNode] = []
        self._cache_updated = False
    
    def generate_id(self, prefix: str = "CAT") -> str:
        """生成分类ID"""
        return f"{prefix}{uuid.uuid4().hex[:16].upper()}"
    
    async def create_category(self, name: str, parent_id: Optional[str] = None,
                            description: Optional[str] = None, **kwargs) -> ProductCategory:
        """创建新分类"""
        # 验证父分类
        parent_category = None
        level = 0
        path = "/"
        
        if parent_id:
            parent_category = await self.session.get(ProductCategory, parent_id)
            if not parent_category:
                raise BusinessException("父分类不存在")
            
            if not parent_category.is_active:
                raise BusinessException("父分类已禁用，无法在其下创建子分类")
            
            level = parent_category.level + 1
            path = f"{parent_category.path}{parent_id}/"
        
        # 检查同级分类名称重复
        result = await self.session.execute(
            select(func.count(ProductCategory.id))
            .where(
                and_(
                    ProductCategory.parent_id == parent_id,
                    ProductCategory.name == name
                )
            )
        )
        if result.scalar() > 0:
            raise BusinessException("同级分类中已存在相同名称的分类")
        
        # 创建分类
        category_id = self.generate_id()
        path = f"{path}{category_id}/" if path != "/" else f"/{category_id}/"
        
        category = ProductCategory(
            id=category_id,
            name=name,
            description=description,
            parent_id=parent_id,
            path=path,
            level=level,
            icon=kwargs.get("icon"),
            banner_image=kwargs.get("banner_image"),
            is_active=kwargs.get("is_active", True),
            is_visible=kwargs.get("is_visible", True),
            sort_order=kwargs.get("sort_order", 0),
            attribute_template=kwargs.get("attribute_template"),
            seo_title=kwargs.get("seo_title"),
            seo_description=kwargs.get("seo_description"),
            seo_keywords=kwargs.get("seo_keywords"),
            slug=kwargs.get("slug") or self._generate_slug(name)
        )
        
        self.session.add(category)
        await self.session.commit()
        
        # 清除缓存
        self._clear_cache()
        
        logger.info(f"创建分类成功: {name} ({category_id})")
        return category
    
    def _generate_slug(self, name: str) -> str:
        """生成URL slug"""
        import re
        # 简化处理：将中文和特殊字符替换为连字符
        slug = re.sub(r'[^\w\s-]', '', name.lower())
        slug = re.sub(r'[-\s]+', '-', slug)
        return slug.strip('-')
    
    async def update_category(self, category_id: str, **kwargs) -> ProductCategory:
        """更新分类信息"""
        category = await self.session.get(ProductCategory, category_id)
        if not category:
            raise BusinessException("分类不存在")
        
        # 不允许更新的字段
        protected_fields = {"id", "path", "level", "created_at"}
        
        # 更新允许的字段
        for field, value in kwargs.items():
            if field not in protected_fields and hasattr(category, field):
                setattr(category, field, value)
        
        # 特殊处理：更新父分类
        if "parent_id" in kwargs:
            await self._move_category(category, kwargs["parent_id"])
        
        await self.session.commit()
        self._clear_cache()
        
        return category
    
    async def _move_category(self, category: ProductCategory, new_parent_id: Optional[str]):
        """移动分类到新的父分类下"""
        if category.parent_id == new_parent_id:
            return  # 没有变化
        
        # 验证新父分类
        new_parent = None
        new_level = 0
        new_path = "/"
        
        if new_parent_id:
            new_parent = await self.session.get(ProductCategory, new_parent_id)
            if not new_parent:
                raise BusinessException("新父分类不存在")
            
            # 检查是否会形成循环引用
            if await self._would_create_cycle(category.id, new_parent_id):
                raise BusinessException("不能将分类移动到其子分类下")
            
            new_level = new_parent.level + 1
            new_path = f"{new_parent.path}{new_parent_id}/"
        
        # 更新当前分类
        old_path = category.path
        old_level = category.level
        
        category.parent_id = new_parent_id
        category.level = new_level
        category.path = f"{new_path}{category.id}/"
        
        # 更新所有子分类的路径和层级
        await self._update_descendants_path(category.id, old_path, old_level)
    
    async def _would_create_cycle(self, category_id: str, new_parent_id: str) -> bool:
        """检查移动分类是否会创建循环引用"""
        current_id = new_parent_id
        visited = set()
        
        while current_id and current_id not in visited:
            if current_id == category_id:
                return True
            
            visited.add(current_id)
            
            # 获取父分类ID
            result = await self.session.execute(
                select(ProductCategory.parent_id)
                .where(ProductCategory.id == current_id)
            )
            row = result.first()
            current_id = row[0] if row else None
        
        return False
    
    async def _update_descendants_path(self, category_id: str, old_path: str, old_level: int):
        """更新所有子分类的路径和层级"""
        # 获取当前分类的新信息
        category = await self.session.get(ProductCategory, category_id)
        new_path = category.path
        new_level = category.level
        level_diff = new_level - old_level
        
        # 更新所有后代分类
        await self.session.execute(
            text("""
                UPDATE product_categories 
                SET 
                    path = REPLACE(path, :old_path, :new_path),
                    level = level + :level_diff
                WHERE path LIKE :search_pattern AND id != :category_id
            """),
            {
                "old_path": old_path,
                "new_path": new_path,
                "level_diff": level_diff,
                "search_pattern": f"{old_path}%",
                "category_id": category_id
            }
        )
    
    async def delete_category(self, category_id: str, force: bool = False):
        """删除分类"""
        category = await self.session.get(ProductCategory, category_id)
        if not category:
            raise BusinessException("分类不存在")
        
        # 检查是否有子分类
        result = await self.session.execute(
            select(func.count(ProductCategory.id))
            .where(ProductCategory.parent_id == category_id)
        )
        children_count = result.scalar()
        
        if children_count > 0 and not force:
            raise BusinessException("该分类下有子分类，无法删除。请先删除子分类或使用强制删除")
        
        # 检查是否有商品
        result = await self.session.execute(
            select(func.count(ProductSPU.id))
            .where(ProductSPU.category_id == category_id)
        )
        spu_count = result.scalar()
        
        if spu_count > 0 and not force:
            raise BusinessException(f"该分类下有 {spu_count} 个商品，无法删除。请先移动商品或使用强制删除")
        
        if force:
            # 强制删除：删除所有子分类和商品的分类关联
            await self.session.execute(
                text("DELETE FROM product_categories WHERE path LIKE :pattern"),
                {"pattern": f"{category.path}%"}
            )
            
            # 将商品的分类设为空
            await self.session.execute(
                text("UPDATE product_spu SET category_id = NULL WHERE category_id = :id"),
                {"id": category_id}
            )
        
        await self.session.delete(category)
        await self.session.commit()
        
        self._clear_cache()
        logger.info(f"删除分类: {category.name} ({category_id})")
    
    async def get_category(self, category_id: str) -> Optional[ProductCategory]:
        """获取单个分类"""
        return await self.session.get(ProductCategory, category_id)
    
    async def get_category_tree(self, include_inactive: bool = False) -> List[CategoryNode]:
        """获取完整的分类树"""
        if not self._cache_updated:
            await self._build_category_cache(include_inactive)
        
        return self._tree_cache
    
    async def _build_category_cache(self, include_inactive: bool = False):
        """构建分类缓存"""
        # 获取所有分类
        query = select(ProductCategory).order_by(
            ProductCategory.level, 
            ProductCategory.sort_order, 
            ProductCategory.name
        )
        
        if not include_inactive:
            query = query.where(ProductCategory.is_active == True)
        
        result = await self.session.execute(query)
        categories = result.scalars().all()
        
        # 构建节点映射
        self._category_cache = {}
        node_map = {}
        
        for category in categories:
            node = CategoryNode(category)
            self._category_cache[category.id] = node
            node_map[category.id] = node
        
        # 构建树结构
        root_nodes = []
        
        for category in categories:
            node = node_map[category.id]
            
            if category.parent_id and category.parent_id in node_map:
                # 添加到父节点
                parent_node = node_map[category.parent_id]
                parent_node.add_child(node)
            else:
                # 根节点
                node.depth = 0
                root_nodes.append(node)
        
        self._tree_cache = root_nodes
        self._cache_updated = True
    
    def _clear_cache(self):
        """清除缓存"""
        self._category_cache.clear()
        self._tree_cache.clear()
        self._cache_updated = False
    
    async def get_category_path(self, category_id: str) -> List[ProductCategory]:
        """获取分类路径（面包屑导航）"""
        category = await self.session.get(ProductCategory, category_id)
        if not category:
            raise BusinessException("分类不存在")
        
        # 解析路径获取所有父级ID
        path_ids = [id for id in category.path.split('/') if id]
        
        if not path_ids:
            return [category]
        
        # 批量获取所有分类
        result = await self.session.execute(
            select(ProductCategory)
            .where(ProductCategory.id.in_(path_ids))
            .order_by(ProductCategory.level)
        )
        
        return result.scalars().all()
    
    async def get_children(self, parent_id: Optional[str] = None, 
                          include_inactive: bool = False) -> List[ProductCategory]:
        """获取指定分类的直接子分类"""
        query = select(ProductCategory).where(ProductCategory.parent_id == parent_id)
        
        if not include_inactive:
            query = query.where(ProductCategory.is_active == True)
        
        query = query.order_by(ProductCategory.sort_order, ProductCategory.name)
        
        result = await self.session.execute(query)
        return result.scalars().all()
    
    async def get_descendants(self, category_id: str, 
                            include_self: bool = False) -> List[ProductCategory]:
        """获取所有后代分类"""
        category = await self.session.get(ProductCategory, category_id)
        if not category:
            raise BusinessException("分类不存在")
        
        query = select(ProductCategory).where(
            ProductCategory.path.like(f"{category.path}%")
        )
        
        if not include_self:
            query = query.where(ProductCategory.id != category_id)
        
        query = query.order_by(ProductCategory.level, ProductCategory.sort_order)
        
        result = await self.session.execute(query)
        return result.scalars().all()
    
    async def search_categories(self, keyword: str, parent_id: Optional[str] = None) -> List[ProductCategory]:
        """搜索分类"""
        query = select(ProductCategory).where(
            or_(
                ProductCategory.name.like(f"%{keyword}%"),
                ProductCategory.description.like(f"%{keyword}%")
            )
        )
        
        if parent_id:
            parent_category = await self.session.get(ProductCategory, parent_id)
            if parent_category:
                query = query.where(
                    ProductCategory.path.like(f"{parent_category.path}%")
                )
        
        query = query.where(ProductCategory.is_active == True)
        query = query.order_by(ProductCategory.level, ProductCategory.name)
        
        result = await self.session.execute(query)
        return result.scalars().all()
    
    async def update_spu_count(self, category_id: str):
        """更新分类的商品数量"""
        result = await self.session.execute(
            select(func.count(ProductSPU.id))
            .where(ProductSPU.category_id == category_id)
        )
        spu_count = result.scalar()
        
        category = await self.session.get(ProductCategory, category_id)
        if category:
            category.spu_count = spu_count
            await self.session.commit()
        
        return spu_count
    
    async def bulk_update_statistics(self):
        """批量更新所有分类的统计信息"""
        # 更新商品数量
        await self.session.execute(text("""
            UPDATE product_categories 
            SET spu_count = (
                SELECT COUNT(*) 
                FROM product_spu 
                WHERE category_id = product_categories.id
            )
        """))
        
        # 更新销量统计（从SPU汇总）
        await self.session.execute(text("""
            UPDATE product_categories 
            SET total_sales = (
                SELECT COALESCE(SUM(sales_count), 0)
                FROM product_spu 
                WHERE category_id = product_categories.id
            )
        """))
        
        await self.session.commit()
        logger.info("分类统计信息更新完成")
    
    async def reorder_categories(self, category_orders: List[Dict[str, int]]):
        """批量更新分类排序"""
        for item in category_orders:
            category_id = item["id"]
            sort_order = item["sort_order"]
            
            await self.session.execute(
                text("UPDATE product_categories SET sort_order = :order WHERE id = :id"),
                {"order": sort_order, "id": category_id}
            )
        
        await self.session.commit()
        self._clear_cache()
    
    async def get_popular_categories(self, limit: int = 10) -> List[Dict]:
        """获取热门分类（按商品数量和销量排序）"""
        result = await self.session.execute(
            select(
                ProductCategory.id,
                ProductCategory.name,
                ProductCategory.spu_count,
                ProductCategory.total_sales,
                ProductCategory.icon,
                ProductCategory.banner_image
            )
            .where(
                and_(
                    ProductCategory.is_active == True,
                    ProductCategory.is_visible == True,
                    ProductCategory.spu_count > 0
                )
            )
            .order_by(
                ProductCategory.total_sales.desc(),
                ProductCategory.spu_count.desc()
            )
            .limit(limit)
        )
        
        categories = []
        for row in result.fetchall():
            categories.append({
                "id": row.id,
                "name": row.name,
                "spu_count": row.spu_count,
                "total_sales": row.total_sales,
                "icon": row.icon,
                "banner_image": row.banner_image
            })
        
        return categories
    
    async def validate_attribute_template(self, template: Dict) -> bool:
        """验证分类属性模板的格式"""
        if not isinstance(template, dict):
            return False
        
        if "attributes" not in template:
            return False
        
        attributes = template["attributes"]
        if not isinstance(attributes, list):
            return False
        
        # 验证每个属性定义
        for attr in attributes:
            if not isinstance(attr, dict):
                return False
            
            required_fields = ["name", "type"]
            if not all(field in attr for field in required_fields):
                return False
            
            # 验证属性类型
            if attr["type"] not in [AttributeType.BASIC.value, AttributeType.SALE.value, AttributeType.CUSTOM.value]:
                return False
        
        return True