"""
商品属性管理服务

实现商品属性系统的核心功能：
1. 属性模板管理 - 分类级别的属性定义
2. SPU属性管理 - 基本属性和销售属性
3. SKU属性生成 - 根据销售属性生成SKU组合
4. 属性验证和继承
"""

from typing import List, Dict, Any, Optional, Set, Tuple
from datetime import datetime
from sqlalchemy import select, and_, or_, func
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
import json
import itertools
from collections import defaultdict

from app.models.product_new import (
    ProductSPU, ProductSKU, ProductAttribute, ProductCategory,
    AttributeType, ProductStatus, SkuStatus
)
from app.core.exceptions import BusinessException
from app.core.logger import logger


class AttributeTemplate:
    """属性模板定义"""
    def __init__(self, name: str, attribute_type: str, required: bool = False,
                 options: List[str] = None, default_value: str = None,
                 validation_rules: Dict = None):
        self.name = name
        self.attribute_type = attribute_type
        self.required = required
        self.options = options or []
        self.default_value = default_value
        self.validation_rules = validation_rules or {}
        self.is_searchable = False
        self.is_filterable = False
        self.sort_order = 0


class ProductAttributeService:
    """商品属性服务"""
    
    def __init__(self, session: AsyncSession):
        self.session = session
        
        # 预定义的属性模板（可以从数据库加载）
        self.attribute_templates = {
            "envelope": [
                AttributeTemplate("材质", AttributeType.BASIC.value, True, 
                                ["牛皮纸", "珠光纸", "艺术纸", "硫酸纸"]),
                AttributeTemplate("尺寸", AttributeType.SALE.value, True,
                                ["C6(114x162mm)", "DL(110x220mm)", "C5(162x229mm)"]),
                AttributeTemplate("颜色", AttributeType.SALE.value, True,
                                ["白色", "米色", "棕色", "黑色", "红色", "蓝色"]),
                AttributeTemplate("封口方式", AttributeType.BASIC.value, False,
                                ["自粘", "糊口", "绳扣", "火漆"])
            ],
            "stationery": [
                AttributeTemplate("品牌", AttributeType.BASIC.value, False),
                AttributeTemplate("材质", AttributeType.BASIC.value, True),
                AttributeTemplate("颜色", AttributeType.SALE.value, True),
                AttributeTemplate("规格", AttributeType.SALE.value, True),
                AttributeTemplate("产地", AttributeType.BASIC.value, False)
            ],
            "stamp": [
                AttributeTemplate("国家/地区", AttributeType.BASIC.value, True),
                AttributeTemplate("年份", AttributeType.BASIC.value, True),
                AttributeTemplate("主题", AttributeType.BASIC.value, True),
                AttributeTemplate("面值", AttributeType.BASIC.value, True),
                AttributeTemplate("品相", AttributeType.SALE.value, True,
                                ["全新", "盖销", "信销", "旧票"]),
                AttributeTemplate("版式", AttributeType.SALE.value, False,
                                ["单枚", "四方连", "整版", "小型张"])
            ]
        }
    
    async def get_category_templates(self, category_id: str) -> List[AttributeTemplate]:
        """获取分类的属性模板"""
        # 获取分类信息
        category = await self.session.get(ProductCategory, category_id)
        if not category:
            raise BusinessException("分类不存在")
        
        # 如果分类有自定义模板，使用自定义模板
        if category.attribute_template:
            return self._parse_attribute_template(category.attribute_template)
        
        # 否则根据分类名称匹配预定义模板
        for key, templates in self.attribute_templates.items():
            if key.lower() in category.name.lower():
                return templates
        
        # 返回默认模板
        return self._get_default_templates()
    
    def _parse_attribute_template(self, template_json: Dict) -> List[AttributeTemplate]:
        """解析JSON格式的属性模板"""
        templates = []
        for attr_def in template_json.get("attributes", []):
            template = AttributeTemplate(
                name=attr_def["name"],
                attribute_type=attr_def.get("type", AttributeType.BASIC.value),
                required=attr_def.get("required", False),
                options=attr_def.get("options", []),
                default_value=attr_def.get("default"),
                validation_rules=attr_def.get("validation", {})
            )
            template.is_searchable = attr_def.get("searchable", False)
            template.is_filterable = attr_def.get("filterable", False)
            template.sort_order = attr_def.get("sort_order", 0)
            templates.append(template)
        
        return templates
    
    def _get_default_templates(self) -> List[AttributeTemplate]:
        """获取默认属性模板"""
        return [
            AttributeTemplate("材质", AttributeType.BASIC.value, False),
            AttributeTemplate("颜色", AttributeType.SALE.value, False),
            AttributeTemplate("尺寸", AttributeType.SALE.value, False),
            AttributeTemplate("重量", AttributeType.BASIC.value, False),
            AttributeTemplate("产地", AttributeType.BASIC.value, False)
        ]
    
    async def create_spu_attributes(self, spu_id: str, attributes: List[Dict]) -> List[ProductAttribute]:
        """为SPU创建属性"""
        created_attributes = []
        
        # 获取SPU信息
        spu = await self.session.get(ProductSPU, spu_id)
        if not spu:
            raise BusinessException("商品不存在")
        
        # 获取分类模板
        templates = await self.get_category_templates(spu.category_id)
        template_dict = {t.name: t for t in templates}
        
        # 验证并创建属性
        for attr_data in attributes:
            attr_name = attr_data.get("name")
            attr_value = attr_data.get("value")
            attr_type = attr_data.get("type", AttributeType.BASIC.value)
            
            # 验证属性
            template = template_dict.get(attr_name)
            if template:
                # 使用模板定义
                attr_type = template.attribute_type
                
                # 验证选项值
                if template.options and attr_value not in template.options:
                    raise BusinessException(f"属性 {attr_name} 的值 {attr_value} 不在允许的选项中")
                
                # 验证必填
                if template.required and not attr_value:
                    raise BusinessException(f"属性 {attr_name} 为必填项")
            
            # 创建属性
            attribute = ProductAttribute(
                spu_id=spu_id,
                attribute_name=attr_name,
                attribute_value=attr_value,
                attribute_type=attr_type,
                is_required=template.required if template else False,
                is_searchable=template.is_searchable if template else False,
                is_filterable=template.is_filterable if template else False,
                sort_order=template.sort_order if template else 0,
                options=template.options if template else None
            )
            
            self.session.add(attribute)
            created_attributes.append(attribute)
        
        await self.session.commit()
        return created_attributes
    
    async def generate_sku_combinations(self, spu_id: str) -> List[Dict]:
        """根据销售属性生成SKU组合"""
        # 获取所有销售属性
        result = await self.session.execute(
            select(ProductAttribute)
            .where(
                and_(
                    ProductAttribute.spu_id == spu_id,
                    ProductAttribute.attribute_type == AttributeType.SALE.value
                )
            )
            .order_by(ProductAttribute.sort_order)
        )
        sale_attributes = result.scalars().all()
        
        if not sale_attributes:
            # 没有销售属性，返回单个默认SKU
            return [{"sale_attributes": {}, "name": "默认规格"}]
        
        # 构建属性组合
        attribute_groups = {}
        for attr in sale_attributes:
            if attr.attribute_name not in attribute_groups:
                attribute_groups[attr.attribute_name] = []
            
            # 如果有选项列表，使用选项列表
            if attr.options:
                attribute_groups[attr.attribute_name] = attr.options
            # 否则解析属性值（可能是逗号分隔的多个值）
            else:
                values = [v.strip() for v in attr.attribute_value.split(',')]
                attribute_groups[attr.attribute_name].extend(values)
        
        # 生成所有组合
        combinations = []
        attr_names = list(attribute_groups.keys())
        attr_values = [attribute_groups[name] for name in attr_names]
        
        for combination in itertools.product(*attr_values):
            sale_attrs = dict(zip(attr_names, combination))
            
            # 生成SKU名称
            name_parts = [f"{v}" for k, v in sale_attrs.items()]
            sku_name = "-".join(name_parts)
            
            combinations.append({
                "sale_attributes": sale_attrs,
                "name": sku_name
            })
        
        return combinations
    
    async def create_skus_from_attributes(self, spu_id: str, sku_data: List[Dict]) -> List[ProductSKU]:
        """根据属性组合创建SKU"""
        # 获取SPU信息
        spu = await self.session.get(ProductSPU, spu_id)
        if not spu:
            raise BusinessException("商品不存在")
        
        created_skus = []
        
        for idx, data in enumerate(sku_data):
            # 生成SKU编码
            sku_code = self._generate_sku_code(spu_id, data["sale_attributes"], idx)
            
            # 创建SKU
            sku = ProductSKU(
                spu_id=spu_id,
                sku_code=sku_code,
                name=data["name"],
                price=data.get("price", spu.min_price or 0),
                original_price=data.get("original_price"),
                cost_price=data.get("cost_price"),
                stock_quantity=data.get("stock_quantity", 0),
                available_stock=data.get("stock_quantity", 0),
                status=SkuStatus.ACTIVE.value,
                is_default=(idx == 0),  # 第一个设为默认
                sale_attributes=data["sale_attributes"],
                weight=data.get("weight"),
                volume=data.get("volume"),
                dimensions=data.get("dimensions"),
                main_image=data.get("image") or spu.main_image,
                max_quantity_per_order=data.get("max_quantity_per_order", 999)
            )
            
            self.session.add(sku)
            created_skus.append(sku)
        
        # 更新SPU的价格范围和库存
        await self._update_spu_statistics(spu_id)
        
        await self.session.commit()
        return created_skus
    
    def _generate_sku_code(self, spu_id: str, attributes: Dict, index: int) -> str:
        """生成SKU编码"""
        # 基础编码：SPU_ID的后8位
        base_code = spu_id[-8:].upper()
        
        # 属性编码：每个属性取首字母
        attr_code = ""
        for key, value in sorted(attributes.items()):
            if value:
                # 取值的首字符（如果是中文，取拼音首字母）
                if value[0].encode('utf-8').isalpha():
                    attr_code += value[0].upper()
                else:
                    # 简化处理，用序号代替
                    attr_code += str(index % 10)
        
        # 组合编码
        return f"SKU-{base_code}-{attr_code}-{index:03d}"
    
    async def _update_spu_statistics(self, spu_id: str):
        """更新SPU的统计信息"""
        # 计算价格范围
        result = await self.session.execute(
            select(
                func.min(ProductSKU.price).label('min_price'),
                func.max(ProductSKU.price).label('max_price'),
                func.sum(ProductSKU.stock_quantity).label('total_stock')
            )
            .where(
                and_(
                    ProductSKU.spu_id == spu_id,
                    ProductSKU.status == SkuStatus.ACTIVE.value
                )
            )
        )
        stats = result.first()
        
        # 更新SPU
        spu = await self.session.get(ProductSPU, spu_id)
        if spu and stats:
            spu.min_price = stats.min_price or 0
            spu.max_price = stats.max_price or 0
            spu.total_stock = stats.total_stock or 0
    
    async def get_spu_attributes(self, spu_id: str) -> Dict[str, List[ProductAttribute]]:
        """获取SPU的所有属性，按类型分组"""
        result = await self.session.execute(
            select(ProductAttribute)
            .where(ProductAttribute.spu_id == spu_id)
            .order_by(ProductAttribute.attribute_type, ProductAttribute.sort_order)
        )
        attributes = result.scalars().all()
        
        # 按类型分组
        grouped = defaultdict(list)
        for attr in attributes:
            grouped[attr.attribute_type].append(attr)
        
        return dict(grouped)
    
    async def update_attribute(self, attribute_id: str, data: Dict) -> ProductAttribute:
        """更新属性"""
        attribute = await self.session.get(ProductAttribute, attribute_id)
        if not attribute:
            raise BusinessException("属性不存在")
        
        # 更新允许的字段
        updatable_fields = ["attribute_value", "is_searchable", "is_filterable", "sort_order"]
        for field in updatable_fields:
            if field in data:
                setattr(attribute, field, data[field])
        
        await self.session.commit()
        return attribute
    
    async def delete_attribute(self, attribute_id: str):
        """删除属性"""
        attribute = await self.session.get(ProductAttribute, attribute_id)
        if not attribute:
            raise BusinessException("属性不存在")
        
        # 检查是否为销售属性
        if attribute.attribute_type == AttributeType.SALE.value:
            # 检查是否有SKU使用此属性
            result = await self.session.execute(
                select(func.count(ProductSKU.id))
                .where(ProductSKU.spu_id == attribute.spu_id)
            )
            sku_count = result.scalar()
            
            if sku_count > 0:
                raise BusinessException("该销售属性已被SKU使用，无法删除")
        
        await self.session.delete(attribute)
        await self.session.commit()
    
    async def validate_sku_attributes(self, spu_id: str, sale_attributes: Dict) -> bool:
        """验证SKU的销售属性是否合法"""
        # 获取SPU的所有销售属性
        result = await self.session.execute(
            select(ProductAttribute)
            .where(
                and_(
                    ProductAttribute.spu_id == spu_id,
                    ProductAttribute.attribute_type == AttributeType.SALE.value
                )
            )
        )
        spu_sale_attributes = result.scalars().all()
        
        # 构建允许的属性和值
        allowed_attributes = {}
        for attr in spu_sale_attributes:
            if attr.options:
                allowed_attributes[attr.attribute_name] = attr.options
            else:
                # 解析逗号分隔的值
                values = [v.strip() for v in attr.attribute_value.split(',')]
                allowed_attributes[attr.attribute_name] = values
        
        # 验证SKU属性
        for attr_name, attr_value in sale_attributes.items():
            if attr_name not in allowed_attributes:
                logger.warning(f"SKU包含未定义的销售属性: {attr_name}")
                return False
            
            if attr_value not in allowed_attributes[attr_name]:
                logger.warning(f"SKU属性值不在允许范围内: {attr_name}={attr_value}")
                return False
        
        return True
    
    async def get_filterable_attributes(self, category_id: Optional[str] = None) -> Dict[str, List[str]]:
        """获取可用于筛选的属性和值"""
        query = select(ProductAttribute).where(
            ProductAttribute.is_filterable == True
        )
        
        if category_id:
            # 按分类筛选
            query = query.join(ProductSPU).where(
                ProductSPU.category_id == category_id
            )
        
        result = await self.session.execute(query)
        attributes = result.scalars().all()
        
        # 整理属性和值
        filter_options = defaultdict(set)
        for attr in attributes:
            if attr.options:
                filter_options[attr.attribute_name].update(attr.options)
            else:
                values = [v.strip() for v in attr.attribute_value.split(',')]
                filter_options[attr.attribute_name].update(values)
        
        # 转换为列表
        return {name: sorted(list(values)) for name, values in filter_options.items()}
    
    async def search_by_attributes(self, attribute_filters: Dict[str, List[str]], 
                                  category_id: Optional[str] = None) -> List[str]:
        """根据属性搜索SPU"""
        # 构建查询
        query = select(ProductSPU.id).distinct()
        
        # 分类过滤
        if category_id:
            query = query.where(ProductSPU.category_id == category_id)
        
        # 属性过滤
        for attr_name, attr_values in attribute_filters.items():
            if attr_values:
                # 创建子查询
                subquery = (
                    select(ProductAttribute.spu_id)
                    .where(
                        and_(
                            ProductAttribute.attribute_name == attr_name,
                            ProductAttribute.attribute_value.in_(attr_values)
                        )
                    )
                )
                query = query.where(ProductSPU.id.in_(subquery))
        
        result = await self.session.execute(query)
        return [row[0] for row in result.fetchall()]