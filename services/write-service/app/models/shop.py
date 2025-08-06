from sqlalchemy import Column, String, Text, DateTime, Boolean, Integer, ForeignKey, Float, Enum
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
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

class ProductType(PyEnum):
    """商品类型枚举"""
    ENVELOPE = "envelope"      # 信封
    STATIONERY = "stationery"  # 文具
    STAMP = "stamp"            # 邮票
    POSTCARD = "postcard"      # 明信片
    GIFT = "gift"              # 礼品
    DIGITAL = "digital"        # 数字商品

class OrderStatus(PyEnum):
    """订单状态枚举"""
    PENDING = "pending"        # 待付款
    PAID = "paid"              # 已付款
    PROCESSING = "processing"  # 处理中
    SHIPPED = "shipped"        # 已发货
    DELIVERED = "delivered"    # 已送达
    CANCELLED = "cancelled"    # 已取消
    REFUNDED = "refunded"      # 已退款

class PaymentStatus(PyEnum):
    """支付状态枚举"""
    PENDING = "pending"        # 待支付
    SUCCESS = "success"        # 支付成功
    FAILED = "failed"          # 支付失败
    REFUNDED = "refunded"      # 已退款
    CANCELLED = "cancelled"    # 已取消

class Product(Base):
    """商品模型"""
    __tablename__ = "shop_products"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="商品ID")
    
    # 基础信息
    name = Column(String(200), nullable=False, comment="商品名称")
    description = Column(Text, comment="商品描述")
    short_description = Column(String(500), comment="商品简介")
    
    # 分类信息
    category = Column(String(50), nullable=False, index=True, comment="商品分类")
    product_type = Column(String(20), nullable=False, index=True, comment="商品类型")
    tags = Column(String(300), comment="商品标签(逗号分隔)")
    brand = Column(String(100), comment="品牌")
    
    # 价格信息
    price = Column(Float, nullable=False, comment="价格")
    original_price = Column(Float, comment="原价")
    cost_price = Column(Float, comment="成本价")
    currency = Column(String(3), default="CNY", comment="货币类型")
    
    # 库存信息
    stock_quantity = Column(Integer, default=0, comment="库存数量")
    min_stock = Column(Integer, default=0, comment="最低库存警戒线")
    max_quantity_per_order = Column(Integer, default=999, comment="单次订购最大数量")
    
    # 商品状态
    status = Column(String(20), default=ProductStatus.DRAFT.value, nullable=False, index=True, comment="商品状态")
    is_featured = Column(Boolean, default=False, comment="是否精选")
    is_digital = Column(Boolean, default=False, comment="是否数字商品")
    
    # 商品属性
    weight = Column(Float, comment="重量(克)")
    dimensions = Column(String(100), comment="尺寸(长x宽x高)")
    color = Column(String(50), comment="颜色")
    material = Column(String(100), comment="材质")
    
    # 媒体资源
    main_image = Column(String(500), comment="主图片URL")
    gallery_images = Column(Text, comment="图片集(JSON格式)")
    video_url = Column(String(500), comment="视频URL")
    
    # SEO信息
    seo_title = Column(String(200), comment="SEO标题")
    seo_description = Column(String(500), comment="SEO描述")
    seo_keywords = Column(String(300), comment="SEO关键词")
    
    # 统计信息
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
    order_items = relationship("OrderItem", back_populates="product", cascade="all, delete-orphan")
    cart_items = relationship("CartItem", back_populates="product", cascade="all, delete-orphan")
    product_reviews = relationship("ProductReview", back_populates="product", cascade="all, delete-orphan")
    product_favorites = relationship("ProductFavorite", back_populates="product", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Product(id={self.id}, name={self.name}, price={self.price})>"
    
    def to_dict(self, include_description=True):
        """转换为字典格式"""
        data = {
            "id": self.id,
            "name": self.name,
            "short_description": self.short_description,
            "category": self.category,
            "product_type": self.product_type,
            "tags": self.tags.split(',') if self.tags else [],
            "brand": self.brand,
            "price": self.price,
            "original_price": self.original_price,
            "currency": self.currency,
            "stock_quantity": self.stock_quantity,
            "max_quantity_per_order": self.max_quantity_per_order,
            "status": self.status,
            "is_featured": self.is_featured,
            "is_digital": self.is_digital,
            "weight": self.weight,
            "dimensions": self.dimensions,
            "color": self.color,
            "material": self.material,
            "main_image": self.main_image,
            "gallery_images": self.gallery_images,
            "video_url": self.video_url,
            "view_count": self.view_count,
            "sales_count": self.sales_count,
            "rating_avg": self.rating_avg,
            "rating_count": self.rating_count,
            "favorite_count": self.favorite_count,
            "creator_name": self.creator_name,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "published_at": self.published_at.isoformat() if self.published_at else None
        }
        
        if include_description:
            data["description"] = self.description
            
        return data

class ProductCategory(Base):
    """商品分类模型"""
    __tablename__ = "shop_categories"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="分类ID")
    
    # 分类信息
    name = Column(String(100), nullable=False, comment="分类名称")
    description = Column(String(500), comment="分类描述")
    parent_id = Column(String(20), ForeignKey("shop_categories.id"), nullable=True, comment="父分类ID")
    
    # 展示配置
    icon = Column(String(100), comment="分类图标")
    banner_image = Column(String(500), comment="分类横幅图片")
    is_active = Column(Boolean, default=True, comment="是否启用")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    
    # 统计
    product_count = Column(Integer, default=0, comment="商品数量")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    parent = relationship("ProductCategory", remote_side=[id], back_populates="children")
    children = relationship("ProductCategory", back_populates="parent", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<ProductCategory(id={self.id}, name={self.name})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "parent_id": self.parent_id,
            "icon": self.icon,
            "banner_image": self.banner_image,
            "is_active": self.is_active,
            "sort_order": self.sort_order,
            "product_count": self.product_count,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class Order(Base):
    """订单模型"""
    __tablename__ = "shop_orders"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="订单ID")
    
    # 用户信息
    user_id = Column(String(50), nullable=False, index=True, comment="用户ID")
    user_name = Column(String(100), comment="用户姓名")
    user_email = Column(String(200), comment="用户邮箱")
    user_phone = Column(String(20), comment="用户电话")
    
    # 订单状态
    status = Column(String(20), default=OrderStatus.PENDING.value, nullable=False, index=True, comment="订单状态")
    payment_status = Column(String(20), default=PaymentStatus.PENDING.value, nullable=False, comment="支付状态")
    
    # 金额信息
    subtotal = Column(Float, nullable=False, comment="商品小计")
    shipping_fee = Column(Float, default=0.0, comment="运费")
    tax_fee = Column(Float, default=0.0, comment="税费")
    discount_amount = Column(Float, default=0.0, comment="优惠金额")
    total_amount = Column(Float, nullable=False, comment="总金额")
    currency = Column(String(3), default="CNY", comment="货币类型")
    
    # 收货信息
    shipping_name = Column(String(100), comment="收货人姓名")
    shipping_phone = Column(String(20), comment="收货人电话")
    shipping_address = Column(Text, comment="收货地址")
    shipping_city = Column(String(100), comment="收货城市")
    shipping_province = Column(String(100), comment="收货省份")
    shipping_postal_code = Column(String(20), comment="邮政编码")
    shipping_method = Column(String(50), comment="配送方式")
    
    # 订单备注
    user_note = Column(String(500), comment="用户备注")
    admin_note = Column(String(500), comment="管理员备注")
    
    # 优惠信息
    coupon_code = Column(String(50), comment="优惠券代码")
    coupon_discount = Column(Float, default=0.0, comment="优惠券折扣金额")
    
    # 支付信息
    payment_method = Column(String(50), comment="支付方式")
    payment_transaction_id = Column(String(100), comment="支付交易号")
    paid_at = Column(DateTime(timezone=True), comment="支付时间")
    
    # 物流信息
    tracking_number = Column(String(100), comment="物流单号")
    shipping_company = Column(String(100), comment="物流公司")
    shipped_at = Column(DateTime(timezone=True), comment="发货时间")
    delivered_at = Column(DateTime(timezone=True), comment="签收时间")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    order_items = relationship("OrderItem", back_populates="order", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Order(id={self.id}, user_id={self.user_id}, total={self.total_amount})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "user_id": self.user_id,
            "user_name": self.user_name,
            "user_email": self.user_email,
            "user_phone": self.user_phone,
            "status": self.status,
            "payment_status": self.payment_status,
            "subtotal": self.subtotal,
            "shipping_fee": self.shipping_fee,
            "tax_fee": self.tax_fee,
            "discount_amount": self.discount_amount,
            "total_amount": self.total_amount,
            "currency": self.currency,
            "shipping_name": self.shipping_name,
            "shipping_phone": self.shipping_phone,
            "shipping_address": self.shipping_address,
            "shipping_city": self.shipping_city,
            "shipping_province": self.shipping_province,
            "shipping_postal_code": self.shipping_postal_code,
            "shipping_method": self.shipping_method,
            "user_note": self.user_note,
            "admin_note": self.admin_note,
            "coupon_code": self.coupon_code,
            "coupon_discount": self.coupon_discount,
            "payment_method": self.payment_method,
            "payment_transaction_id": self.payment_transaction_id,
            "paid_at": self.paid_at.isoformat() if self.paid_at else None,
            "tracking_number": self.tracking_number,
            "shipping_company": self.shipping_company,
            "shipped_at": self.shipped_at.isoformat() if self.shipped_at else None,
            "delivered_at": self.delivered_at.isoformat() if self.delivered_at else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class OrderItem(Base):
    """订单商品项模型"""
    __tablename__ = "shop_order_items"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="订单项ID")
    
    # 关联信息
    order_id = Column(String(20), ForeignKey("shop_orders.id"), nullable=False, index=True, comment="订单ID")
    product_id = Column(String(20), ForeignKey("shop_products.id"), nullable=False, comment="商品ID")
    
    # 商品信息快照
    product_name = Column(String(200), nullable=False, comment="商品名称")
    product_image = Column(String(500), comment="商品图片")
    product_sku = Column(String(100), comment="商品SKU")
    
    # 价格和数量
    unit_price = Column(Float, nullable=False, comment="单价")
    quantity = Column(Integer, nullable=False, comment="数量")
    total_price = Column(Float, nullable=False, comment="小计")
    
    # 商品属性
    product_attributes = Column(Text, comment="商品属性(JSON格式)")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    
    # 关联关系
    order = relationship("Order", back_populates="order_items")
    product = relationship("Product", back_populates="order_items")
    
    def __repr__(self):
        return f"<OrderItem(id={self.id}, product={self.product_name}, qty={self.quantity})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "order_id": self.order_id,
            "product_id": self.product_id,
            "product_name": self.product_name,
            "product_image": self.product_image,
            "product_sku": self.product_sku,
            "unit_price": self.unit_price,
            "quantity": self.quantity,
            "total_price": self.total_price,
            "product_attributes": self.product_attributes,
            "created_at": self.created_at.isoformat() if self.created_at else None
        }

class Cart(Base):
    """购物车模型"""
    __tablename__ = "shop_carts"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="购物车ID")
    
    # 用户信息
    user_id = Column(String(50), nullable=False, unique=True, index=True, comment="用户ID")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    cart_items = relationship("CartItem", back_populates="cart", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Cart(id={self.id}, user_id={self.user_id})>"

class CartItem(Base):
    """购物车商品项模型"""
    __tablename__ = "shop_cart_items"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="购物车项ID")
    
    # 关联信息
    cart_id = Column(String(20), ForeignKey("shop_carts.id"), nullable=False, index=True, comment="购物车ID")
    product_id = Column(String(20), ForeignKey("shop_products.id"), nullable=False, comment="商品ID")
    
    # 数量和属性
    quantity = Column(Integer, nullable=False, comment="数量")
    product_attributes = Column(Text, comment="商品属性(JSON格式)")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="添加时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    cart = relationship("Cart", back_populates="cart_items")
    product = relationship("Product", back_populates="cart_items")
    
    def __repr__(self):
        return f"<CartItem(id={self.id}, product_id={self.product_id}, qty={self.quantity})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "cart_id": self.cart_id,
            "product_id": self.product_id,
            "quantity": self.quantity,
            "product_attributes": self.product_attributes,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class ProductReview(Base):
    """商品评价模型"""
    __tablename__ = "shop_product_reviews"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="评价ID")
    
    # 关联信息
    product_id = Column(String(20), ForeignKey("shop_products.id"), nullable=False, index=True, comment="商品ID")
    user_id = Column(String(50), nullable=False, comment="用户ID")
    user_name = Column(String(100), comment="用户姓名")
    order_id = Column(String(20), ForeignKey("shop_orders.id"), nullable=True, comment="订单ID")
    
    # 评价内容
    rating = Column(Integer, nullable=False, comment="评分(1-5)")
    title = Column(String(200), comment="评价标题")
    content = Column(Text, comment="评价内容")
    
    # 评价图片
    images = Column(Text, comment="评价图片(JSON格式)")
    
    # 商家回复
    reply_content = Column(Text, comment="商家回复")
    reply_at = Column(DateTime(timezone=True), comment="回复时间")
    
    # 状态
    is_anonymous = Column(Boolean, default=False, comment="是否匿名")
    is_verified = Column(Boolean, default=False, comment="是否已验证购买")
    is_hidden = Column(Boolean, default=False, comment="是否隐藏")
    
    # 统计
    helpful_count = Column(Integer, default=0, comment="有用数")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    product = relationship("Product", back_populates="product_reviews")
    order = relationship("Order")
    
    def __repr__(self):
        return f"<ProductReview(id={self.id}, product_id={self.product_id}, rating={self.rating})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "product_id": self.product_id,
            "user_id": self.user_id if not self.is_anonymous else None,
            "user_name": self.user_name if not self.is_anonymous else "匿名用户",
            "order_id": self.order_id,
            "rating": self.rating,
            "title": self.title,
            "content": self.content,
            "images": self.images,
            "reply_content": self.reply_content,
            "reply_at": self.reply_at.isoformat() if self.reply_at else None,
            "is_anonymous": self.is_anonymous,
            "is_verified": self.is_verified,
            "is_hidden": self.is_hidden,
            "helpful_count": self.helpful_count,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class ProductFavorite(Base):
    """商品收藏模型"""
    __tablename__ = "shop_product_favorites"
    
    # 联合主键
    product_id = Column(String(20), ForeignKey("shop_products.id"), primary_key=True, comment="商品ID")
    user_id = Column(String(50), primary_key=True, comment="用户ID")
    
    # 收藏备注
    note = Column(String(200), comment="收藏备注")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="收藏时间")
    
    # 关联关系
    product = relationship("Product", back_populates="product_favorites")
    
    def __repr__(self):
        return f"<ProductFavorite(product_id={self.product_id}, user_id={self.user_id})>"