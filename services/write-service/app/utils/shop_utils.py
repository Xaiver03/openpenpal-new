import random
import string
from typing import Optional, List, Dict, Any, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import desc, and_, or_, func
from datetime import datetime, timedelta
import json

from app.models.shop import (
    Product, ProductCategory, Order, OrderItem, Cart, CartItem,
    ProductReview, ProductFavorite, ProductStatus, OrderStatus, PaymentStatus
)

def generate_product_id() -> str:
    """生成商品ID: PD + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    # 避免混淆字符
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"PD{random_part}"

def generate_order_id() -> str:
    """生成订单ID: OD + 时间戳 + 4位随机"""
    timestamp = int(datetime.now().timestamp() * 1000) % 1000000000  # 取后9位
    random_part = ''.join(random.choices(string.digits, k=4))
    return f"OD{timestamp}{random_part}"

def generate_cart_id() -> str:
    """生成购物车ID: CT + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"CT{random_part}"

def generate_order_item_id() -> str:
    """生成订单项ID: OI + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"OI{random_part}"

def generate_cart_item_id() -> str:
    """生成购物车项ID: CI + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"CI{random_part}"

def generate_review_id() -> str:
    """生成评价ID: RV + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"RV{random_part}"

def generate_unique_product_id(db: Session) -> str:
    """生成唯一的商品ID"""
    max_attempts = 10
    for _ in range(max_attempts):
        product_id = generate_product_id()
        existing = db.query(Product).filter(Product.id == product_id).first()
        if not existing:
            return product_id
    
    # 如果10次都冲突，使用时间戳确保唯一性
    timestamp = int(datetime.now().timestamp() * 1000)
    return f"PD{timestamp % 10000000000:010d}"

def generate_unique_order_id(db: Session) -> str:
    """生成唯一的订单ID"""
    max_attempts = 10
    for _ in range(max_attempts):
        order_id = generate_order_id()
        existing = db.query(Order).filter(Order.id == order_id).first()
        if not existing:
            return order_id
    
    # 备用方案
    timestamp = int(datetime.now().timestamp() * 1000)
    return f"OD{timestamp}"

def calculate_shipping_fee(total_amount: float, shipping_method: str = "standard") -> float:
    """计算运费"""
    # 简单的运费计算逻辑
    if total_amount >= 99.0:  # 满99免邮
        return 0.0
    
    shipping_rates = {
        "standard": 8.0,    # 标准快递
        "express": 15.0,    # 加急快递
        "same_day": 25.0,   # 同城当日达
        "pickup": 0.0       # 自提
    }
    
    return shipping_rates.get(shipping_method, 8.0)

def calculate_tax_fee(subtotal: float, tax_rate: float = 0.0) -> float:
    """计算税费"""
    # 简化处理，暂时不收税
    return subtotal * tax_rate

def validate_coupon_code(coupon_code: str, subtotal: float) -> Tuple[bool, float, str]:
    """验证优惠券代码"""
    # 简单的优惠券验证逻辑
    coupons = {
        "WELCOME10": {"type": "percentage", "value": 0.1, "min_amount": 0, "description": "新用户9折优惠"},
        "SAVE20": {"type": "fixed", "value": 20.0, "min_amount": 100, "description": "满100减20"},
        "VIP15": {"type": "percentage", "value": 0.15, "min_amount": 200, "description": "VIP用户85折"}
    }
    
    if coupon_code not in coupons:
        return False, 0.0, "优惠券不存在"
    
    coupon = coupons[coupon_code]
    
    if subtotal < coupon["min_amount"]:
        return False, 0.0, f"订单金额需满{coupon['min_amount']}元"
    
    if coupon["type"] == "percentage":
        discount = subtotal * coupon["value"]
    else:
        discount = coupon["value"]
    
    # 优惠不能超过商品金额
    discount = min(discount, subtotal)
    
    return True, discount, coupon["description"]

def check_product_stock(db: Session, product_id: str, quantity: int) -> bool:
    """检查商品库存"""
    product = db.query(Product).filter(Product.id == product_id).first()
    if not product:
        return False
    
    if product.status != ProductStatus.ACTIVE.value:
        return False
    
    return product.stock_quantity >= quantity

def update_product_stock(db: Session, product_id: str, quantity_change: int) -> bool:
    """更新商品库存"""
    try:
        product = db.query(Product).filter(Product.id == product_id).first()
        if not product:
            return False
        
        new_stock = product.stock_quantity + quantity_change
        if new_stock < 0:
            return False
        
        product.stock_quantity = new_stock
        
        # 检查是否需要更新商品状态
        if new_stock == 0 and product.status == ProductStatus.ACTIVE.value:
            product.status = ProductStatus.OUT_OF_STOCK.value
        elif new_stock > 0 and product.status == ProductStatus.OUT_OF_STOCK.value:
            product.status = ProductStatus.ACTIVE.value
        
        db.commit()
        return True
        
    except Exception as e:
        print(f"更新商品库存失败: {e}")
        db.rollback()
        return False

def update_product_stats(db: Session, product_id: str, stat_type: str, increment: int = 1):
    """更新商品统计数据"""
    try:
        product = db.query(Product).filter(Product.id == product_id).first()
        if not product:
            return False
        
        if stat_type == "view":
            product.view_count += increment
        elif stat_type == "sales":
            product.sales_count += increment
        elif stat_type == "favorite":
            product.favorite_count += increment
        
        db.commit()
        return True
        
    except Exception as e:
        print(f"更新商品统计失败: {e}")
        db.rollback()
        return False

def update_product_rating(db: Session, product_id: str, new_rating: int, old_rating: Optional[int] = None):
    """更新商品评分"""
    try:
        product = db.query(Product).filter(Product.id == product_id).first()
        if not product:
            return False
        
        if old_rating is None:
            # 新评分
            total_score = product.rating_avg * product.rating_count + new_rating
            product.rating_count += 1
            product.rating_avg = round(total_score / product.rating_count, 2)
        else:
            # 更新评分
            if product.rating_count > 0:
                total_score = product.rating_avg * product.rating_count - old_rating + new_rating
                product.rating_avg = round(total_score / product.rating_count, 2)
        
        db.commit()
        return True
        
    except Exception as e:
        print(f"更新商品评分失败: {e}")
        db.rollback()
        return False

def get_user_cart(db: Session, user_id: str) -> Optional[Cart]:
    """获取用户购物车"""
    cart = db.query(Cart).filter(Cart.user_id == user_id).first()
    if not cart:
        # 创建新购物车
        cart_id = generate_cart_id()
        cart = Cart(id=cart_id, user_id=user_id)
        db.add(cart)
        db.commit()
        db.refresh(cart)
    
    return cart

def calculate_cart_total(db: Session, cart: Cart) -> Dict[str, Any]:
    """计算购物车总价"""
    total_amount = 0.0
    items_count = 0
    
    cart_items_with_products = db.query(CartItem, Product).join(
        Product, CartItem.product_id == Product.id
    ).filter(
        CartItem.cart_id == cart.id,
        Product.status == ProductStatus.ACTIVE.value
    ).all()
    
    for cart_item, product in cart_items_with_products:
        subtotal = product.price * cart_item.quantity
        total_amount += subtotal
        items_count += cart_item.quantity
    
    return {
        "items_count": items_count,
        "total_amount": round(total_amount, 2)
    }

def get_popular_products(db: Session, limit: int = 10) -> List[Product]:
    """获取热门商品"""
    try:
        return db.query(Product).filter(
            Product.status == ProductStatus.ACTIVE.value
        ).order_by(
            desc(Product.sales_count * 0.5 + Product.view_count * 0.3 + Product.rating_avg * 0.2)
        ).limit(limit).all()
        
    except Exception as e:
        print(f"获取热门商品失败: {e}")
        return []

def get_featured_products(db: Session, limit: int = 8) -> List[Product]:
    """获取精选商品"""
    try:
        return db.query(Product).filter(
            Product.status == ProductStatus.ACTIVE.value,
            Product.is_featured == True
        ).order_by(desc(Product.created_at)).limit(limit).all()
        
    except Exception as e:
        print(f"获取精选商品失败: {e}")
        return []

def get_recommended_products(db: Session, user_id: str, limit: int = 10) -> List[Product]:
    """获取推荐商品"""
    try:
        # 简单推荐算法：基于用户购买历史和收藏
        
        # 获取用户购买过的商品分类
        purchased_categories = db.query(Product.category).join(
            OrderItem, Product.id == OrderItem.product_id
        ).join(
            Order, OrderItem.order_id == Order.id
        ).filter(
            Order.user_id == user_id,
            Order.status.in_([OrderStatus.DELIVERED.value, OrderStatus.SHIPPED.value])
        ).distinct().all()
        
        # 获取用户收藏的商品分类
        favorited_categories = db.query(Product.category).join(
            ProductFavorite, Product.id == ProductFavorite.product_id
        ).filter(
            ProductFavorite.user_id == user_id
        ).distinct().all()
        
        preferred_categories = list(set([cat[0] for cat in purchased_categories + favorited_categories]))
        
        # 构建推荐查询
        query = db.query(Product).filter(
            Product.status == ProductStatus.ACTIVE.value
        )
        
        if preferred_categories:
            query = query.filter(Product.category.in_(preferred_categories))
        
        products = query.order_by(
            desc(Product.rating_avg * 0.4 + Product.sales_count * 0.3 + Product.view_count * 0.3)
        ).limit(limit).all()
        
        # 如果推荐不足，补充热门商品
        if len(products) < limit:
            additional_products = get_popular_products(db, limit - len(products))
            products.extend([p for p in additional_products if p.id not in [pr.id for pr in products]])
        
        return products[:limit]
        
    except Exception as e:
        print(f"获取推荐商品失败: {e}")
        return get_popular_products(db, limit)

def get_shop_statistics(db: Session) -> Dict[str, Any]:
    """获取商店统计数据"""
    try:
        stats = {}
        
        # 基础统计
        stats["total_products"] = db.query(Product).filter(
            Product.status == ProductStatus.ACTIVE.value
        ).count()
        
        stats["total_orders"] = db.query(Order).count()
        
        # 计算总销售额
        total_sales = db.query(func.sum(Order.total_amount)).filter(
            Order.payment_status == PaymentStatus.SUCCESS.value
        ).scalar() or 0.0
        stats["total_sales"] = round(total_sales, 2)
        
        # 客户数量
        stats["total_customers"] = db.query(Order.user_id).distinct().count()
        
        # 分类分布
        category_distribution = db.query(
            Product.category, func.count(Product.id)
        ).filter(
            Product.status == ProductStatus.ACTIVE.value
        ).group_by(Product.category).all()
        
        stats["category_distribution"] = {category: count for category, count in category_distribution}
        
        return stats
        
    except Exception as e:
        print(f"获取商店统计数据失败: {e}")
        return {}

def validate_order_permissions(order: Order, user_id: str, user_role: str = "user") -> Dict[str, bool]:
    """验证用户对订单的权限"""
    permissions = {
        "can_view": False,
        "can_edit": False,
        "can_cancel": False,
        "can_refund": False,
        "can_ship": False,
        "can_manage": False
    }
    
    # 管理员拥有所有权限
    if user_role == "admin":
        permissions.update({
            "can_view": True,
            "can_edit": True,
            "can_cancel": True,
            "can_refund": True,
            "can_ship": True,
            "can_manage": True
        })
        return permissions
    
    # 用户只能查看和操作自己的订单
    if order.user_id == user_id:
        permissions["can_view"] = True
        
        # 根据订单状态确定可执行操作
        if order.status == OrderStatus.PENDING.value:
            permissions["can_edit"] = True
            permissions["can_cancel"] = True
        elif order.status in [OrderStatus.PAID.value, OrderStatus.PROCESSING.value]:
            permissions["can_cancel"] = True
    
    return permissions

class ProductFilter:
    """商品过滤器"""
    
    @staticmethod
    def apply_filters(query, filters: Dict[str, Any]):
        """应用过滤条件"""
        
        # 状态过滤（默认只显示上架商品）
        if filters.get("status"):
            query = query.filter(Product.status == filters["status"])
        else:
            query = query.filter(Product.status == ProductStatus.ACTIVE.value)
        
        # 分类过滤
        if filters.get("category"):
            query = query.filter(Product.category == filters["category"])
        
        # 商品类型过滤
        if filters.get("product_type"):
            query = query.filter(Product.product_type == filters["product_type"])
        
        # 品牌过滤
        if filters.get("brand"):
            query = query.filter(Product.brand == filters["brand"])
        
        # 标签过滤
        if filters.get("tags"):
            tags = filters["tags"] if isinstance(filters["tags"], list) else [filters["tags"]]
            for tag in tags:
                query = query.filter(Product.tags.contains(tag))
        
        # 价格范围过滤
        if filters.get("min_price"):
            query = query.filter(Product.price >= filters["min_price"])
        
        if filters.get("max_price"):
            query = query.filter(Product.price <= filters["max_price"])
        
        # 库存过滤
        if filters.get("in_stock_only"):
            query = query.filter(Product.stock_quantity > 0)
        
        # 精选过滤
        if filters.get("featured_only"):
            query = query.filter(Product.is_featured == True)
        
        # 关键词搜索
        if filters.get("keyword"):
            keyword = f"%{filters['keyword']}%"
            query = query.filter(
                or_(
                    Product.name.contains(keyword),
                    Product.description.contains(keyword),
                    Product.short_description.contains(keyword),
                    Product.tags.contains(keyword)
                )
            )
        
        return query
    
    @staticmethod
    def apply_sorting(query, sort_by: str = "created_at", order: str = "desc"):
        """应用排序"""
        
        if sort_by == "created_at":
            order_by = Product.created_at
        elif sort_by == "price":
            order_by = Product.price
        elif sort_by == "sales":
            order_by = Product.sales_count
        elif sort_by == "rating":
            order_by = Product.rating_avg
        elif sort_by == "views":
            order_by = Product.view_count
        elif sort_by == "name":
            order_by = Product.name
        elif sort_by == "popularity":
            # 热度排序：销量*0.4 + 评分*0.3 + 浏览量*0.3
            order_by = Product.sales_count * 0.4 + Product.rating_avg * 0.3 + Product.view_count * 0.3
        else:
            order_by = Product.created_at
        
        if order == "desc":
            query = query.order_by(desc(order_by))
        else:
            query = query.order_by(order_by)
        
        return query