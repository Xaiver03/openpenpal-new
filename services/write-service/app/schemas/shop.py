from pydantic import BaseModel, Field, validator
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum

class ProductType(str, Enum):
    """商品类型枚举"""
    ENVELOPE = "envelope"
    STATIONERY = "stationery"
    STAMP = "stamp"
    POSTCARD = "postcard"
    GIFT = "gift"
    DIGITAL = "digital"

class ProductStatus(str, Enum):
    """商品状态枚举"""
    DRAFT = "draft"
    ACTIVE = "active"
    INACTIVE = "inactive"
    OUT_OF_STOCK = "out_of_stock"
    DISCONTINUED = "discontinued"

class OrderStatus(str, Enum):
    """订单状态枚举"""
    PENDING = "pending"
    PAID = "paid"
    PROCESSING = "processing"
    SHIPPED = "shipped"
    DELIVERED = "delivered"
    CANCELLED = "cancelled"
    REFUNDED = "refunded"

class PaymentStatus(str, Enum):
    """支付状态枚举"""
    PENDING = "pending"
    SUCCESS = "success"
    FAILED = "failed"
    REFUNDED = "refunded"
    CANCELLED = "cancelled"

# 基础响应格式
class SuccessResponse(BaseModel):
    """统一成功响应格式"""
    code: int = 0
    msg: str = "success"
    data: Optional[dict] = None

class ErrorResponse(BaseModel):
    """统一错误响应格式"""
    code: int = Field(gt=0, description="错误码")
    msg: str = Field(description="错误信息")
    data: Optional[dict] = None

# 商品相关Schema
class ProductCreate(BaseModel):
    """创建商品请求"""
    name: str = Field(..., min_length=1, max_length=200, description="商品名称")
    description: Optional[str] = Field(None, description="商品描述")
    short_description: Optional[str] = Field(None, max_length=500, description="商品简介")
    category: str = Field(..., max_length=50, description="商品分类")
    product_type: ProductType = Field(..., description="商品类型")
    tags: Optional[List[str]] = Field(default=[], description="商品标签")
    brand: Optional[str] = Field(None, max_length=100, description="品牌")
    price: float = Field(..., gt=0, description="价格")
    original_price: Optional[float] = Field(None, gt=0, description="原价")
    stock_quantity: int = Field(default=0, ge=0, description="库存数量")
    min_stock: int = Field(default=0, ge=0, description="最低库存")
    max_quantity_per_order: int = Field(default=999, gt=0, description="单次最大购买数量")
    is_digital: bool = Field(default=False, description="是否数字商品")
    weight: Optional[float] = Field(None, gt=0, description="重量(克)")
    dimensions: Optional[str] = Field(None, max_length=100, description="尺寸")
    color: Optional[str] = Field(None, max_length=50, description="颜色")
    material: Optional[str] = Field(None, max_length=100, description="材质")
    main_image: Optional[str] = Field(None, description="主图片URL")
    gallery_images: Optional[List[str]] = Field(default=[], description="图片集")
    video_url: Optional[str] = Field(None, description="视频URL")
    
    @validator('tags')
    def validate_tags(cls, v):
        if v is None:
            return []
        if len(v) > 20:
            raise ValueError('标签数量不能超过20个')
        for tag in v:
            if len(tag) > 30:
                raise ValueError('单个标签长度不能超过30个字符')
        return v
    
    @validator('original_price')
    def validate_original_price(cls, v, values):
        if v is not None and 'price' in values and v <= values['price']:
            raise ValueError('原价必须大于现价')
        return v

class ProductUpdate(BaseModel):
    """更新商品请求"""
    name: Optional[str] = Field(None, min_length=1, max_length=200, description="商品名称")
    description: Optional[str] = Field(None, description="商品描述")
    short_description: Optional[str] = Field(None, max_length=500, description="商品简介")
    category: Optional[str] = Field(None, max_length=50, description="商品分类")
    product_type: Optional[ProductType] = Field(None, description="商品类型")
    tags: Optional[List[str]] = Field(None, description="商品标签")
    brand: Optional[str] = Field(None, max_length=100, description="品牌")
    price: Optional[float] = Field(None, gt=0, description="价格")
    original_price: Optional[float] = Field(None, gt=0, description="原价")
    stock_quantity: Optional[int] = Field(None, ge=0, description="库存数量")
    min_stock: Optional[int] = Field(None, ge=0, description="最低库存")
    max_quantity_per_order: Optional[int] = Field(None, gt=0, description="单次最大购买数量")
    weight: Optional[float] = Field(None, gt=0, description="重量(克)")
    dimensions: Optional[str] = Field(None, max_length=100, description="尺寸")
    color: Optional[str] = Field(None, max_length=50, description="颜色")
    material: Optional[str] = Field(None, max_length=100, description="材质")
    main_image: Optional[str] = Field(None, description="主图片URL")
    gallery_images: Optional[List[str]] = Field(None, description="图片集")
    video_url: Optional[str] = Field(None, description="视频URL")

class ProductResponse(BaseModel):
    """商品响应"""
    id: str
    name: str
    description: Optional[str] = None
    short_description: Optional[str] = None
    category: str
    product_type: str
    tags: List[str] = []
    brand: Optional[str] = None
    price: float
    original_price: Optional[float] = None
    currency: str = "CNY"
    stock_quantity: int
    max_quantity_per_order: int
    status: str
    is_featured: bool = False
    is_digital: bool = False
    weight: Optional[float] = None
    dimensions: Optional[str] = None
    color: Optional[str] = None
    material: Optional[str] = None
    main_image: Optional[str] = None
    gallery_images: Optional[str] = None
    video_url: Optional[str] = None
    view_count: int = 0
    sales_count: int = 0
    rating_avg: float = 0.0
    rating_count: int = 0
    favorite_count: int = 0
    creator_name: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    published_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class ProductListItem(BaseModel):
    """商品列表项（简化版）"""
    id: str
    name: str
    short_description: Optional[str] = None
    category: str
    product_type: str
    tags: List[str] = []
    brand: Optional[str] = None
    price: float
    original_price: Optional[float] = None
    currency: str = "CNY"
    stock_quantity: int
    status: str
    is_featured: bool = False
    main_image: Optional[str] = None
    view_count: int = 0
    sales_count: int = 0
    rating_avg: float = 0.0
    rating_count: int = 0
    created_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class ProductListResponse(BaseModel):
    """商品列表响应"""
    products: List[ProductListItem]
    total: int
    page: int
    pages: int
    has_next: bool
    has_prev: bool

# 购物车相关Schema
class CartItemAdd(BaseModel):
    """添加购物车商品请求"""
    product_id: str = Field(..., description="商品ID")
    quantity: int = Field(..., gt=0, description="数量")
    product_attributes: Optional[Dict[str, Any]] = Field(default={}, description="商品属性")

class CartItemUpdate(BaseModel):
    """更新购物车商品请求"""
    quantity: int = Field(..., gt=0, description="数量")
    product_attributes: Optional[Dict[str, Any]] = Field(None, description="商品属性")

class CartItemResponse(BaseModel):
    """购物车商品项响应"""
    id: str
    cart_id: str
    product_id: str
    quantity: int
    product_attributes: Optional[Dict[str, Any]] = None
    product: Optional[ProductListItem] = None  # 包含商品信息
    subtotal: Optional[float] = None  # 小计
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class CartResponse(BaseModel):
    """购物车响应"""
    id: str
    user_id: str
    items: List[CartItemResponse] = []
    items_count: int = 0
    total_amount: float = 0.0
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None

# 订单相关Schema
class OrderItemCreate(BaseModel):
    """订单商品项创建"""
    product_id: str = Field(..., description="商品ID")
    quantity: int = Field(..., gt=0, description="数量")
    product_attributes: Optional[Dict[str, Any]] = Field(default={}, description="商品属性")

class ShippingAddress(BaseModel):
    """收货地址"""
    name: str = Field(..., min_length=1, max_length=100, description="收货人姓名")
    phone: str = Field(..., min_length=1, max_length=20, description="收货人电话")
    address: str = Field(..., min_length=1, description="详细地址")
    city: str = Field(..., min_length=1, max_length=100, description="城市")
    province: str = Field(..., min_length=1, max_length=100, description="省份")
    postal_code: Optional[str] = Field(None, max_length=20, description="邮政编码")

class OrderCreate(BaseModel):
    """创建订单请求"""
    items: List[OrderItemCreate] = Field(..., min_items=1, description="订单商品")
    shipping_address: ShippingAddress = Field(..., description="收货地址")
    shipping_method: Optional[str] = Field(None, description="配送方式")
    payment_method: Optional[str] = Field(None, description="支付方式")
    user_note: Optional[str] = Field(None, max_length=500, description="用户备注")
    coupon_code: Optional[str] = Field(None, description="优惠券代码")

class OrderItemResponse(BaseModel):
    """订单商品项响应"""
    id: str
    order_id: str
    product_id: str
    product_name: str
    product_image: Optional[str] = None
    product_sku: Optional[str] = None
    unit_price: float
    quantity: int
    total_price: float
    product_attributes: Optional[Dict[str, Any]] = None
    created_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class OrderResponse(BaseModel):
    """订单响应"""
    id: str
    user_id: str
    user_name: Optional[str] = None
    user_email: Optional[str] = None
    user_phone: Optional[str] = None
    status: str
    payment_status: str
    subtotal: float
    shipping_fee: float = 0.0
    tax_fee: float = 0.0
    discount_amount: float = 0.0
    total_amount: float
    currency: str = "CNY"
    shipping_name: Optional[str] = None
    shipping_phone: Optional[str] = None
    shipping_address: Optional[str] = None
    shipping_city: Optional[str] = None
    shipping_province: Optional[str] = None
    shipping_postal_code: Optional[str] = None
    shipping_method: Optional[str] = None
    user_note: Optional[str] = None
    admin_note: Optional[str] = None
    coupon_code: Optional[str] = None
    coupon_discount: float = 0.0
    payment_method: Optional[str] = None
    payment_transaction_id: Optional[str] = None
    paid_at: Optional[datetime] = None
    tracking_number: Optional[str] = None
    shipping_company: Optional[str] = None
    shipped_at: Optional[datetime] = None
    delivered_at: Optional[datetime] = None
    items: List[OrderItemResponse] = []
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class OrderListResponse(BaseModel):
    """订单列表响应"""
    orders: List[OrderResponse]
    total: int
    page: int
    pages: int

class OrderStatusUpdate(BaseModel):
    """订单状态更新"""
    status: OrderStatus = Field(..., description="订单状态")
    admin_note: Optional[str] = Field(None, description="管理员备注")
    tracking_number: Optional[str] = Field(None, description="物流单号")
    shipping_company: Optional[str] = Field(None, description="物流公司")

# 商品评价相关Schema
class ProductReviewCreate(BaseModel):
    """创建商品评价请求"""
    rating: int = Field(..., ge=1, le=5, description="评分(1-5)")
    title: Optional[str] = Field(None, max_length=200, description="评价标题")
    content: Optional[str] = Field(None, description="评价内容")
    images: Optional[List[str]] = Field(default=[], description="评价图片")
    is_anonymous: bool = Field(default=False, description="是否匿名")

class ProductReviewResponse(BaseModel):
    """商品评价响应"""
    id: str
    product_id: str
    user_id: Optional[str] = None
    user_name: str
    order_id: Optional[str] = None
    rating: int
    title: Optional[str] = None
    content: Optional[str] = None
    images: Optional[str] = None
    reply_content: Optional[str] = None
    reply_at: Optional[datetime] = None
    is_anonymous: bool = False
    is_verified: bool = False
    is_hidden: bool = False
    helpful_count: int = 0
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class ProductReviewListResponse(BaseModel):
    """商品评价列表响应"""
    reviews: List[ProductReviewResponse]
    total: int
    page: int
    pages: int
    rating_summary: Dict[str, int] = {}  # 评分分布统计

# 收藏相关Schema
class ProductFavoriteCreate(BaseModel):
    """添加商品收藏请求"""
    product_id: str = Field(..., description="商品ID")
    note: Optional[str] = Field(None, max_length=200, description="收藏备注")

class ProductFavoriteResponse(BaseModel):
    """商品收藏响应"""
    product_id: str
    favorited: bool
    favorite_count: int
    note: Optional[str] = None

# 分类相关Schema
class ProductCategoryResponse(BaseModel):
    """商品分类响应"""
    id: str
    name: str
    description: Optional[str] = None
    parent_id: Optional[str] = None
    icon: Optional[str] = None
    banner_image: Optional[str] = None
    is_active: bool = True
    sort_order: int = 0
    product_count: int = 0
    children: List['ProductCategoryResponse'] = []
    
    class Config:
        from_attributes = True

class ProductCategoryListResponse(BaseModel):
    """商品分类列表响应"""
    categories: List[ProductCategoryResponse]

# 搜索相关Schema
class ProductSearchRequest(BaseModel):
    """商品搜索请求"""
    keyword: Optional[str] = Field(None, description="关键词")
    category: Optional[str] = Field(None, description="分类过滤")
    product_type: Optional[ProductType] = Field(None, description="商品类型过滤")
    tags: Optional[List[str]] = Field(None, description="标签过滤")
    brand: Optional[str] = Field(None, description="品牌过滤")
    min_price: Optional[float] = Field(None, ge=0, description="最低价格")
    max_price: Optional[float] = Field(None, ge=0, description="最高价格")
    in_stock_only: bool = Field(default=True, description="只显示有库存商品")
    featured_only: bool = Field(default=False, description="只显示精选商品")
    sort_by: str = Field(default="created_at", description="排序字段")
    order: str = Field(default="desc", description="排序方向")
    page: int = Field(default=1, ge=1, description="页码")
    limit: int = Field(default=20, ge=1, le=100, description="每页数量")

class ProductSearchResponse(BaseModel):
    """商品搜索响应"""
    products: List[ProductListItem]
    total: int
    page: int
    pages: int
    search_time: float
    filters_applied: Dict[str, Any]

# 统计相关Schema
class ShopStatsResponse(BaseModel):
    """商店统计响应"""
    total_products: int
    total_orders: int
    total_sales: float
    total_customers: int
    category_distribution: Dict[str, int]
    popular_products: List[ProductListItem]
    recent_orders: List[OrderResponse]
    sales_trend: List[Dict[str, Any]]  # 销售趋势数据

# 更新递归模型引用
ProductCategoryResponse.model_rebuild()