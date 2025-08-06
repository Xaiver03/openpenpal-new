"""
重构后的商品模型 - SPU+SKU分离设计

SPU (Standard Product Unit): 标准商品单元，代表一个商品的抽象概念
SKU (Stock Keeping Unit): 库存保持单元，代表具体的商品规格和库存
"""

from sqlalchemy import Column, String, Text, DateTime, Boolean, Integer, ForeignKey, Float, Enum, JSON, Index
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship, validates
from datetime import datetime
from enum import Enum as PyEnum
from app.core.database import Base


class ProductStatus(PyEnum):
    """商品状态枚举"""
    DRAFT = "draft"            # 草稿
    ACTIVE = "active"          # 上架
    INACTIVE = "inactive"      # 下架
    OUT_OF_STOCK = "out_of_stock"  # 缺货
    DISCONTINUED = "discontinued"  # 停产


class SkuStatus(PyEnum):
    """SKU状态枚举"""
    ACTIVE = "active"          # 可销售
    INACTIVE = "inactive"      # 不可销售
    OUT_OF_STOCK = "out_of_stock"  # 缺货
    DISCONTINUED = "discontinued"  # 停产


class AttributeType(PyEnum):
    """属性类型枚举"""
    BASIC = "basic"           # 基本属性（非规格属性，如品牌、材质）
    SALE = "sale"             # 销售属性（影响SKU的规格属性，如颜色、尺寸）
    CUSTOM = "custom"         # 自定义属性


class ProductType(PyEnum):
    """商品类型枚举"""
    ENVELOPE = "envelope"      # 信封
    STATIONERY = "stationery"  # 文具
    STAMP = "stamp"            # 邮票
    POSTCARD = "postcard"      # 明信片
    GIFT = "gift"              # 礼品
    DIGITAL = "digital"        # 数字商品


# ==================== SPU模型 ====================

class ProductSPU(Base):
    """商品SPU模型 - 标准商品单元"""
    __tablename__ = "product_spu"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="SPU ID")
    
    # 基础信息
    name = Column(String(200), nullable=False, comment="商品名称")
    subtitle = Column(String(300), comment="商品副标题")
    description = Column(Text, comment="商品详细描述")
    short_description = Column(String(500), comment="商品简介")
    
    # 分类信息
    category_id = Column(String(20), ForeignKey("product_categories.id"), nullable=False, index=True, comment="分类ID")
    product_type = Column(String(20), nullable=False, index=True, comment="商品类型")
    brand_id = Column(String(20), ForeignKey("product_brands.id"), nullable=True, comment="品牌ID")
    tags = Column(JSON, comment="商品标签列表")
    
    # 商品状态
    status = Column(String(20), default=ProductStatus.DRAFT.value, nullable=False, index=True, comment="商品状态")
    is_featured = Column(Boolean, default=False, comment="是否精选")
    is_digital = Column(Boolean, default=False, comment="是否数字商品")
    is_virtual = Column(Boolean, default=False, comment="是否虚拟商品")
    
    # 媒体资源
    main_image = Column(String(500), comment="主图片URL")
    gallery_images = Column(JSON, comment="图片集列表")
    video_url = Column(String(500), comment="视频URL")
    detail_images = Column(JSON, comment="详情图片列表")
    
    # SEO信息
    seo_title = Column(String(200), comment="SEO标题")
    seo_description = Column(String(500), comment="SEO描述")
    seo_keywords = Column(JSON, comment="SEO关键词列表")
    slug = Column(String(200), unique=True, comment="URL slug")
    
    # 统计信息（汇总所有SKU）
    total_stock = Column(Integer, default=0, comment="总库存")
    min_price = Column(Float, comment="最低价格")
    max_price = Column(Float, comment="最高价格")
    view_count = Column(Integer, default=0, comment="浏览次数")
    sales_count = Column(Integer, default=0, comment="销售数量")
    rating_avg = Column(Float, default=0.0, comment="平均评分")
    rating_count = Column(Integer, default=0, comment="评分人数")
    favorite_count = Column(Integer, default=0, comment="收藏次数")
    
    # 创建者信息
    creator_id = Column(String(50), comment="创建者ID")
    creator_name = Column(String(100), comment="创建者姓名")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    published_at = Column(DateTime(timezone=True), comment="发布时间")
    
    # 关联关系
    category = relationship("ProductCategory", back_populates="spus")
    brand = relationship("ProductBrand", back_populates="spus")
    skus = relationship("ProductSKU", back_populates="spu", cascade="all, delete-orphan")
    attributes = relationship("ProductAttribute", back_populates="spu", cascade="all, delete-orphan")
    reviews = relationship("ProductReview", back_populates="spu", cascade="all, delete-orphan")
    favorites = relationship("ProductFavorite", back_populates="spu", cascade="all, delete-orphan")
    
    # 索引
    __table_args__ = (
        Index('idx_spu_category_status', 'category_id', 'status'),
        Index('idx_spu_brand_status', 'brand_id', 'status'),
        Index('idx_spu_created_at', 'created_at'),
    )
    
    def __repr__(self):
        return f"<ProductSPU(id={self.id}, name={self.name})>"
    
    @validates('tags', 'seo_keywords')
    def validate_json_list(self, key, value):
        """验证JSON列表格式"""
        if value is not None and not isinstance(value, list):
            raise ValueError(f"{key} must be a list")
        return value
    
    def to_dict(self, include_skus=False, include_attributes=False):
        """转换为字典格式"""
        data = {
            "id": self.id,
            "name": self.name,
            "subtitle": self.subtitle,
            "description": self.description,
            "short_description": self.short_description,
            "category_id": self.category_id,
            "product_type": self.product_type,
            "brand_id": self.brand_id,
            "tags": self.tags or [],
            "status": self.status,
            "is_featured": self.is_featured,
            "is_digital": self.is_digital,
            "is_virtual": self.is_virtual,
            "main_image": self.main_image,
            "gallery_images": self.gallery_images or [],
            "video_url": self.video_url,
            "detail_images": self.detail_images or [],
            "seo_title": self.seo_title,
            "seo_description": self.seo_description,
            "seo_keywords": self.seo_keywords or [],
            "slug": self.slug,
            "total_stock": self.total_stock,
            "min_price": self.min_price,
            "max_price": self.max_price,
            "view_count": self.view_count,
            "sales_count": self.sales_count,
            "rating_avg": self.rating_avg,
            "rating_count": self.rating_count,
            "favorite_count": self.favorite_count,
            "creator_id": self.creator_id,
            "creator_name": self.creator_name,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "published_at": self.published_at.isoformat() if self.published_at else None,
        }
        
        if include_skus and self.skus:
            data["skus"] = [sku.to_dict() for sku in self.skus]
        
        if include_attributes and self.attributes:
            data["attributes"] = [attr.to_dict() for attr in self.attributes]
        
        return data


# ==================== SKU模型 ====================

class ProductSKU(Base):
    """商品SKU模型 - 库存保持单元"""
    __tablename__ = "product_sku"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="SKU ID")
    
    # 关联SPU
    spu_id = Column(String(20), ForeignKey("product_spu.id"), nullable=False, index=True, comment="SPU ID")
    
    # SKU信息
    sku_code = Column(String(100), unique=True, nullable=False, index=True, comment="SKU编码")
    name = Column(String(200), comment="SKU名称（如：红色-大号）")
    
    # 价格信息
    price = Column(Float, nullable=False, comment="售价")
    original_price = Column(Float, comment="原价")
    cost_price = Column(Float, comment="成本价")
    currency = Column(String(3), default="CNY", comment="货币类型")
    
    # 库存信息
    stock_quantity = Column(Integer, default=0, comment="库存数量")
    available_stock = Column(Integer, default=0, comment="可售库存（库存-预占）")
    reserved_stock = Column(Integer, default=0, comment="预占库存")
    min_stock = Column(Integer, default=0, comment="最低库存警戒线")
    max_quantity_per_order = Column(Integer, default=999, comment="单次订购最大数量")
    
    # SKU状态
    status = Column(String(20), default=SkuStatus.ACTIVE.value, nullable=False, index=True, comment="SKU状态")
    is_default = Column(Boolean, default=False, comment="是否默认SKU")
    
    # 物理属性
    weight = Column(Float, comment="重量(克)")
    volume = Column(Float, comment="体积(立方厘米)")
    dimensions = Column(String(100), comment="尺寸(长x宽x高)")
    
    # 销售属性值 (JSON格式存储，如: {"颜色": "红色", "尺寸": "L"})
    sale_attributes = Column(JSON, comment="销售属性值")
    
    # SKU专属图片
    main_image = Column(String(500), comment="SKU主图片URL")
    images = Column(JSON, comment="SKU图片列表")
    
    # 统计信息
    sales_count = Column(Integer, default=0, comment="销售数量")
    
    # 供应商信息
    supplier_id = Column(String(50), comment="供应商ID")
    supplier_code = Column(String(100), comment="供应商编码")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    spu = relationship("ProductSPU", back_populates="skus")
    stock_records = relationship("StockRecord", back_populates="sku", cascade="all, delete-orphan")
    price_records = relationship("PriceRecord", back_populates="sku", cascade="all, delete-orphan")
    order_items = relationship("OrderItem", back_populates="sku")
    cart_items = relationship("CartItem", back_populates="sku")
    
    # 索引
    __table_args__ = (
        Index('idx_sku_spu_status', 'spu_id', 'status'),
        Index('idx_sku_code', 'sku_code'),
        Index('idx_sku_price', 'price'),
        Index('idx_sku_stock', 'stock_quantity', 'status'),
    )
    
    def __repr__(self):
        return f"<ProductSKU(id={self.id}, sku_code={self.sku_code}, price={self.price})>"
    
    @validates('sale_attributes')
    def validate_sale_attributes(self, key, value):
        """验证销售属性格式"""
        if value is not None and not isinstance(value, dict):
            raise ValueError("sale_attributes must be a dict")
        return value
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "spu_id": self.spu_id,
            "sku_code": self.sku_code,
            "name": self.name,
            "price": self.price,
            "original_price": self.original_price,
            "cost_price": self.cost_price,
            "currency": self.currency,
            "stock_quantity": self.stock_quantity,
            "available_stock": self.available_stock,
            "reserved_stock": self.reserved_stock,
            "min_stock": self.min_stock,
            "max_quantity_per_order": self.max_quantity_per_order,
            "status": self.status,
            "is_default": self.is_default,
            "weight": self.weight,
            "volume": self.volume,
            "dimensions": self.dimensions,
            "sale_attributes": self.sale_attributes or {},
            "main_image": self.main_image,
            "images": self.images or [],
            "sales_count": self.sales_count,
            "supplier_id": self.supplier_id,
            "supplier_code": self.supplier_code,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }


# ==================== 属性模型 ====================

class ProductAttribute(Base):
    """商品属性模型"""
    __tablename__ = "product_attributes"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="属性ID")
    
    # 关联SPU
    spu_id = Column(String(20), ForeignKey("product_spu.id"), nullable=False, index=True, comment="SPU ID")
    
    # 属性信息
    attribute_name = Column(String(100), nullable=False, comment="属性名称")
    attribute_value = Column(Text, comment="属性值")
    attribute_type = Column(String(20), default=AttributeType.BASIC.value, comment="属性类型")
    
    # 属性配置
    is_required = Column(Boolean, default=False, comment="是否必填")
    is_searchable = Column(Boolean, default=False, comment="是否可搜索")
    is_filterable = Column(Boolean, default=False, comment="是否可筛选")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    
    # 属性选项（JSON格式，用于枚举类型属性）
    options = Column(JSON, comment="属性选项列表")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    spu = relationship("ProductSPU", back_populates="attributes")
    
    # 索引
    __table_args__ = (
        Index('idx_attr_spu_name', 'spu_id', 'attribute_name'),
        Index('idx_attr_type_searchable', 'attribute_type', 'is_searchable'),
    )
    
    def __repr__(self):
        return f"<ProductAttribute(id={self.id}, name={self.attribute_name})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "spu_id": self.spu_id,
            "attribute_name": self.attribute_name,
            "attribute_value": self.attribute_value,
            "attribute_type": self.attribute_type,
            "is_required": self.is_required,
            "is_searchable": self.is_searchable,
            "is_filterable": self.is_filterable,
            "sort_order": self.sort_order,
            "options": self.options,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }


# ==================== 分类模型 ====================

class ProductCategory(Base):
    """商品分类模型 - 支持多级分类"""
    __tablename__ = "product_categories"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="分类ID")
    
    # 分类信息
    name = Column(String(100), nullable=False, comment="分类名称")
    description = Column(String(500), comment="分类描述")
    parent_id = Column(String(20), ForeignKey("product_categories.id"), nullable=True, comment="父分类ID")
    
    # 分类路径（便于查询，如：/1/2/3/）
    path = Column(String(500), comment="分类路径")
    level = Column(Integer, default=0, comment="分类层级")
    
    # 展示配置
    icon = Column(String(100), comment="分类图标")
    banner_image = Column(String(500), comment="分类横幅图片")
    is_active = Column(Boolean, default=True, comment="是否启用")
    is_visible = Column(Boolean, default=True, comment="是否显示")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    
    # 分类属性配置（JSON格式，定义该分类下商品的属性模板）
    attribute_template = Column(JSON, comment="属性模板")
    
    # 统计
    spu_count = Column(Integer, default=0, comment="SPU数量")
    total_sales = Column(Integer, default=0, comment="总销量")
    
    # SEO信息
    seo_title = Column(String(200), comment="SEO标题")
    seo_description = Column(String(500), comment="SEO描述")
    seo_keywords = Column(JSON, comment="SEO关键词")
    slug = Column(String(200), comment="URL slug")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    parent = relationship("ProductCategory", remote_side=[id], back_populates="children")
    children = relationship("ProductCategory", back_populates="parent", cascade="all, delete-orphan")
    spus = relationship("ProductSPU", back_populates="category")
    
    # 索引
    __table_args__ = (
        Index('idx_category_parent_sort', 'parent_id', 'sort_order'),
        Index('idx_category_path', 'path'),
        Index('idx_category_level_active', 'level', 'is_active'),
    )
    
    def __repr__(self):
        return f"<ProductCategory(id={self.id}, name={self.name})>"
    
    def to_dict(self, include_children=False):
        """转换为字典格式"""
        data = {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "parent_id": self.parent_id,
            "path": self.path,
            "level": self.level,
            "icon": self.icon,
            "banner_image": self.banner_image,
            "is_active": self.is_active,
            "is_visible": self.is_visible,
            "sort_order": self.sort_order,
            "attribute_template": self.attribute_template,
            "spu_count": self.spu_count,
            "total_sales": self.total_sales,
            "seo_title": self.seo_title,
            "seo_description": self.seo_description,
            "seo_keywords": self.seo_keywords,
            "slug": self.slug,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }
        
        if include_children and self.children:
            data["children"] = [child.to_dict() for child in self.children]
        
        return data


# ==================== 品牌模型 ====================

class ProductBrand(Base):
    """商品品牌模型"""
    __tablename__ = "product_brands"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="品牌ID")
    
    # 品牌信息
    name = Column(String(100), nullable=False, unique=True, comment="品牌名称")
    english_name = Column(String(100), comment="英文名称")
    description = Column(Text, comment="品牌描述")
    
    # 品牌媒体
    logo = Column(String(500), comment="品牌Logo")
    banner_image = Column(String(500), comment="品牌横幅图片")
    
    # 品牌配置
    is_active = Column(Boolean, default=True, comment="是否启用")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    
    # 品牌统计
    spu_count = Column(Integer, default=0, comment="SPU数量")
    total_sales = Column(Integer, default=0, comment="总销量")
    
    # SEO信息
    seo_title = Column(String(200), comment="SEO标题")
    seo_description = Column(String(500), comment="SEO描述")
    slug = Column(String(200), unique=True, comment="URL slug")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    spus = relationship("ProductSPU", back_populates="brand")
    
    def __repr__(self):
        return f"<ProductBrand(id={self.id}, name={self.name})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "name": self.name,
            "english_name": self.english_name,
            "description": self.description,
            "logo": self.logo,
            "banner_image": self.banner_image,
            "is_active": self.is_active,
            "sort_order": self.sort_order,
            "spu_count": self.spu_count,
            "total_sales": self.total_sales,
            "seo_title": self.seo_title,
            "seo_description": self.seo_description,
            "slug": self.slug,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }


# ==================== 库存记录模型 ====================

class StockRecord(Base):
    """库存变更记录模型"""
    __tablename__ = "stock_records"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="记录ID")
    
    # 关联SKU
    sku_id = Column(String(20), ForeignKey("product_sku.id"), nullable=False, index=True, comment="SKU ID")
    
    # 变更信息
    change_type = Column(String(20), nullable=False, comment="变更类型（入库/出库/调整/预占/释放）")
    change_quantity = Column(Integer, nullable=False, comment="变更数量（正数为增加，负数为减少）")
    before_quantity = Column(Integer, nullable=False, comment="变更前库存")
    after_quantity = Column(Integer, nullable=False, comment="变更后库存")
    
    # 关联单据
    reference_type = Column(String(50), comment="关联单据类型（order/purchase/adjust等）")
    reference_id = Column(String(50), comment="关联单据ID")
    
    # 操作信息
    operator_id = Column(String(50), comment="操作人ID")
    operator_name = Column(String(100), comment="操作人姓名")
    remark = Column(String(500), comment="备注")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    
    # 关联关系
    sku = relationship("ProductSKU", back_populates="stock_records")
    
    # 索引
    __table_args__ = (
        Index('idx_stock_sku_created', 'sku_id', 'created_at'),
        Index('idx_stock_reference', 'reference_type', 'reference_id'),
    )
    
    def __repr__(self):
        return f"<StockRecord(id={self.id}, sku_id={self.sku_id}, change={self.change_quantity})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "sku_id": self.sku_id,
            "change_type": self.change_type,
            "change_quantity": self.change_quantity,
            "before_quantity": self.before_quantity,
            "after_quantity": self.after_quantity,
            "reference_type": self.reference_type,
            "reference_id": self.reference_id,
            "operator_id": self.operator_id,
            "operator_name": self.operator_name,
            "remark": self.remark,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }


# ==================== 价格记录模型 ====================

class PriceRecord(Base):
    """价格变更记录模型"""
    __tablename__ = "price_records"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="记录ID")
    
    # 关联SKU
    sku_id = Column(String(20), ForeignKey("product_sku.id"), nullable=False, index=True, comment="SKU ID")
    
    # 价格信息
    old_price = Column(Float, comment="变更前价格")
    new_price = Column(Float, nullable=False, comment="变更后价格")
    old_original_price = Column(Float, comment="变更前原价")
    new_original_price = Column(Float, comment="变更后原价")
    
    # 变更原因
    change_reason = Column(String(200), comment="变更原因")
    
    # 有效期
    effective_time = Column(DateTime(timezone=True), comment="生效时间")
    expire_time = Column(DateTime(timezone=True), comment="过期时间")
    
    # 操作信息
    operator_id = Column(String(50), comment="操作人ID")
    operator_name = Column(String(100), comment="操作人姓名")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    
    # 关联关系
    sku = relationship("ProductSKU", back_populates="price_records")
    
    # 索引
    __table_args__ = (
        Index('idx_price_sku_created', 'sku_id', 'created_at'),
        Index('idx_price_effective', 'effective_time', 'expire_time'),
    )
    
    def __repr__(self):
        return f"<PriceRecord(id={self.id}, sku_id={self.sku_id}, price={self.new_price})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "sku_id": self.sku_id,
            "old_price": self.old_price,
            "new_price": self.new_price,
            "old_original_price": self.old_original_price,
            "new_original_price": self.new_original_price,
            "change_reason": self.change_reason,
            "effective_time": self.effective_time.isoformat() if self.effective_time else None,
            "expire_time": self.expire_time.isoformat() if self.expire_time else None,
            "operator_id": self.operator_id,
            "operator_name": self.operator_name,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }