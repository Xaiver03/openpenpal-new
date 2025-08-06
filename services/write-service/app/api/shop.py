from fastapi import APIRouter, Depends, HTTPException, Query, BackgroundTasks
from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any
from datetime import datetime
import json
import time

from app.core.database import get_db, get_async_session
from app.utils.websocket_client import websocket_publisher
from app.utils.auth import get_current_user, check_admin_permission
from app.models.shop import (
    Product, ProductCategory, Order, OrderItem, Cart, CartItem,
    ProductReview, ProductFavorite, ProductStatus, OrderStatus, PaymentStatus
)
from app.schemas.shop import (
    ProductCreate, ProductUpdate, ProductResponse, ProductListItem, ProductListResponse,
    CartItemAdd, CartItemUpdate, CartResponse, CartItemResponse,
    OrderCreate, OrderResponse, OrderListResponse, OrderStatusUpdate,
    ProductReviewCreate, ProductReviewResponse, ProductReviewListResponse,
    ProductFavoriteCreate, ProductFavoriteResponse,
    ProductSearchRequest, ProductSearchResponse,
    ShopStatsResponse, SuccessResponse, ErrorResponse
)
from app.utils.shop_utils import (
    generate_unique_product_id, generate_unique_order_id, generate_cart_id,
    generate_order_item_id, generate_cart_item_id, generate_review_id,
    calculate_shipping_fee, calculate_tax_fee, validate_coupon_code,
    check_product_stock, update_product_stock, update_product_stats, update_product_rating,
    get_user_cart, calculate_cart_total, get_popular_products, get_featured_products,
    get_recommended_products, get_shop_statistics, validate_order_permissions,
    ProductFilter
)

router = APIRouter(prefix="/api/shop", tags=["shop"])

# ============= 商品相关API =============

@router.get("/products", response_model=ProductListResponse)
async def get_products(
    page: int = Query(1, ge=1, description="页码"),
    limit: int = Query(20, ge=1, le=100, description="每页数量"),
    category: Optional[str] = Query(None, description="分类过滤"),
    product_type: Optional[str] = Query(None, description="商品类型过滤"),
    keyword: Optional[str] = Query(None, description="关键词搜索"),
    sort_by: str = Query("created_at", description="排序字段"),
    order: str = Query("desc", description="排序方向"),
    in_stock_only: bool = Query(True, description="只显示有库存商品"),
    featured_only: bool = Query(False, description="只显示精选商品"),
    min_price: Optional[float] = Query(None, ge=0, description="最低价格"),
    max_price: Optional[float] = Query(None, ge=0, description="最高价格"),
    db: Session = Depends(get_db)
):
    """获取商品列表"""
    try:
        # 构建查询
        query = db.query(Product)
        
        # 应用过滤器
        filters = {
            "category": category,
            "product_type": product_type,
            "keyword": keyword,
            "in_stock_only": in_stock_only,
            "featured_only": featured_only,
            "min_price": min_price,
            "max_price": max_price
        }
        query = ProductFilter.apply_filters(query, filters)
        
        # 应用排序
        query = ProductFilter.apply_sorting(query, sort_by, order)
        
        # 计算总数
        total = query.count()
        
        # 分页
        offset = (page - 1) * limit
        products = query.offset(offset).limit(limit).all()
        
        # 计算分页信息
        pages = (total + limit - 1) // limit
        has_next = page < pages
        has_prev = page > 1
        
        # 转换为响应格式
        product_items = []
        for product in products:
            item_data = product.to_dict(include_description=False)
            product_items.append(ProductListItem(**item_data))
        
        return ProductListResponse(
            products=product_items,
            total=total,
            page=page,
            pages=pages,
            has_next=has_next,
            has_prev=has_prev
        )
        
    except Exception as e:
        print(f"获取商品列表失败: {e}")
        raise HTTPException(status_code=500, detail="获取商品列表失败")

@router.get("/products/{product_id}", response_model=ProductResponse)
async def get_product(
    product_id: str,
    db: Session = Depends(get_db),
    current_user: Optional[str] = Depends(get_current_user)
):
    """获取商品详情"""
    try:
        product = db.query(Product).filter(Product.id == product_id).first()
        if not product:
            raise HTTPException(status_code=404, detail="商品不存在")
        
        # 增加浏览次数
        if current_user:
            update_product_stats(db, product_id, "view")
        
        # 转换为响应格式
        product_data = product.to_dict()
        return ProductResponse(**product_data)
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"获取商品详情失败: {e}")
        raise HTTPException(status_code=500, detail="获取商品详情失败")

@router.post("/products", response_model=SuccessResponse)
async def create_product(
    product_data: ProductCreate,
    admin_user: str = Depends(check_admin_permission),
    db: Session = Depends(get_db)
):
    """创建商品（管理员）"""
    try:
        
        # 生成商品ID
        product_id = generate_unique_product_id(db)
        
        # 创建商品
        product = Product(
            id=product_id,
            name=product_data.name,
            description=product_data.description,
            short_description=product_data.short_description,
            category=product_data.category,
            product_type=product_data.product_type.value,
            tags=','.join(product_data.tags) if product_data.tags else None,
            brand=product_data.brand,
            price=product_data.price,
            original_price=product_data.original_price,
            stock_quantity=product_data.stock_quantity,
            min_stock=product_data.min_stock,
            max_quantity_per_order=product_data.max_quantity_per_order,
            is_digital=product_data.is_digital,
            weight=product_data.weight,
            dimensions=product_data.dimensions,
            color=product_data.color,
            material=product_data.material,
            main_image=product_data.main_image,
            gallery_images=json.dumps(product_data.gallery_images) if product_data.gallery_images else None,
            video_url=product_data.video_url,
            creator_id=admin_user,
            status=ProductStatus.DRAFT.value
        )
        
        db.add(product)
        db.commit()
        db.refresh(product)
        
        return SuccessResponse(
            msg="商品创建成功",
            data={"product_id": product_id}
        )
        
    except Exception as e:
        print(f"创建商品失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="创建商品失败")

@router.put("/products/{product_id}", response_model=SuccessResponse)
async def update_product(
    product_id: str,
    product_data: ProductUpdate,
    admin_user: str = Depends(check_admin_permission),
    db: Session = Depends(get_db)
):
    """更新商品（管理员）"""
    try:
        product = db.query(Product).filter(Product.id == product_id).first()
        if not product:
            raise HTTPException(status_code=404, detail="商品不存在")
        
        # 更新商品信息
        update_data = product_data.dict(exclude_unset=True)
        
        for field, value in update_data.items():
            if field == "tags" and value is not None:
                setattr(product, field, ','.join(value))
            elif field == "gallery_images" and value is not None:
                setattr(product, field, json.dumps(value))
            elif field == "product_type" and value is not None:
                setattr(product, field, value.value)
            else:
                setattr(product, field, value)
        
        db.commit()
        
        return SuccessResponse(msg="商品更新成功")
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"更新商品失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="更新商品失败")

@router.delete("/products/{product_id}", response_model=SuccessResponse)
async def delete_product(
    product_id: str,
    admin_user: str = Depends(check_admin_permission),
    db: Session = Depends(get_db)
):
    """删除商品（管理员）"""
    try:
        product = db.query(Product).filter(Product.id == product_id).first()
        if not product:
            raise HTTPException(status_code=404, detail="商品不存在")
        
        # 检查是否有相关订单（防止删除已售商品）
        existing_orders = db.query(OrderItem).filter(OrderItem.product_id == product_id).first()
        if existing_orders:
            # 如果有订单记录，只标记为已停产而不是删除
            product.status = ProductStatus.DISCONTINUED.value
            db.commit()
            return SuccessResponse(msg="商品已标记为停产（因存在订单记录）")
        
        # 删除相关的购物车项目、收藏、评价
        db.query(CartItem).filter(CartItem.product_id == product_id).delete()
        db.query(ProductFavorite).filter(ProductFavorite.product_id == product_id).delete()
        db.query(ProductReview).filter(ProductReview.product_id == product_id).delete()
        
        # 删除商品
        db.delete(product)
        db.commit()
        
        return SuccessResponse(msg="商品删除成功")
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"删除商品失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="删除商品失败")

@router.patch("/products/{product_id}/status", response_model=SuccessResponse)
async def update_product_status(
    product_id: str,
    status: str,
    admin_user: str = Depends(check_admin_permission),
    db: Session = Depends(get_db)
):
    """更新商品状态（上架/下架）"""
    try:
        product = db.query(Product).filter(Product.id == product_id).first()
        if not product:
            raise HTTPException(status_code=404, detail="商品不存在")
        
        # 验证状态值
        valid_statuses = [s.value for s in ProductStatus]
        if status not in valid_statuses:
            raise HTTPException(status_code=400, detail=f"无效的状态值，支持的状态: {valid_statuses}")
        
        # 特殊逻辑：上架时检查库存
        if status == ProductStatus.ACTIVE.value and product.stock_quantity == 0:
            raise HTTPException(status_code=400, detail="库存为0的商品无法上架")
        
        old_status = product.status
        product.status = status
        
        # 更新发布时间
        if status == ProductStatus.ACTIVE.value and old_status != ProductStatus.ACTIVE.value:
            product.published_at = datetime.now()
        
        db.commit()
        
        status_names = {
            "draft": "草稿",
            "active": "上架",
            "inactive": "下架", 
            "out_of_stock": "缺货",
            "discontinued": "停产"
        }
        
        return SuccessResponse(
            msg=f"商品状态已更新为：{status_names.get(status, status)}",
            data={"old_status": old_status, "new_status": status}
        )
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"更新商品状态失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="更新商品状态失败")

@router.post("/products/batch-import", response_model=SuccessResponse)
async def batch_import_products(
    products_data: List[ProductCreate],
    admin_user: str = Depends(check_admin_permission),
    db: Session = Depends(get_db)
):
    """批量导入商品（管理员）"""
    try:
        
        if len(products_data) > 100:
            raise HTTPException(status_code=400, detail="单次最多导入100个商品")
        
        success_count = 0
        failed_items = []
        
        for i, product_data in enumerate(products_data):
            try:
                # 生成商品ID
                product_id = generate_unique_product_id(db)
                
                # 创建商品
                product = Product(
                    id=product_id,
                    name=product_data.name,
                    description=product_data.description,
                    short_description=product_data.short_description,
                    category=product_data.category,
                    product_type=product_data.product_type.value,
                    tags=','.join(product_data.tags) if product_data.tags else None,
                    brand=product_data.brand,
                    price=product_data.price,
                    original_price=product_data.original_price,
                    stock_quantity=product_data.stock_quantity,
                    min_stock=product_data.min_stock,
                    max_quantity_per_order=product_data.max_quantity_per_order,
                    is_digital=product_data.is_digital,
                    weight=product_data.weight,
                    dimensions=product_data.dimensions,
                    color=product_data.color,
                    material=product_data.material,
                    main_image=product_data.main_image,
                    gallery_images=json.dumps(product_data.gallery_images) if product_data.gallery_images else None,
                    video_url=product_data.video_url,
                    creator_id=admin_user,
                    status=ProductStatus.DRAFT.value
                )
                
                db.add(product)
                success_count += 1
                
            except Exception as e:
                failed_items.append({
                    "index": i + 1,
                    "name": product_data.name,
                    "error": str(e)
                })
        
        try:
            db.commit()
        except Exception as e:
            db.rollback()
            raise HTTPException(status_code=500, detail=f"批量导入失败: {e}")
        
        result_msg = f"批量导入完成：成功 {success_count} 个"
        if failed_items:
            result_msg += f"，失败 {len(failed_items)} 个"
        
        return SuccessResponse(
            msg=result_msg,
            data={
                "success_count": success_count,
                "failed_count": len(failed_items),
                "failed_items": failed_items
            }
        )
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"批量导入商品失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="批量导入商品失败")

# ============= 购物车相关API =============

@router.get("/cart", response_model=CartResponse)
async def get_user_cart(
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取用户购物车"""
    try:
        cart = get_user_cart(db, current_user)
        
        # 获取购物车项目和商品信息
        cart_items = []
        for cart_item in cart.cart_items:
            product = cart_item.product
            if product and product.status == ProductStatus.ACTIVE.value:
                # 计算小计
                subtotal = product.price * cart_item.quantity
                
                # 构建响应数据
                item_data = cart_item.to_dict()
                item_data["product"] = ProductListItem(**product.to_dict(include_description=False))
                item_data["subtotal"] = subtotal
                
                cart_items.append(CartItemResponse(**item_data))
        
        # 计算购物车总计
        cart_totals = calculate_cart_total(db, cart)
        
        return CartResponse(
            id=cart.id,
            user_id=cart.user_id,
            items=cart_items,
            items_count=cart_totals["items_count"],
            total_amount=cart_totals["total_amount"],
            created_at=cart.created_at,
            updated_at=cart.updated_at
        )
        
    except Exception as e:
        print(f"获取购物车失败: {e}")
        raise HTTPException(status_code=500, detail="获取购物车失败")

@router.post("/cart", response_model=SuccessResponse)
async def add_to_cart(
    cart_data: CartItemAdd,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """添加商品到购物车"""
    try:
        # 检查商品存在和库存
        if not check_product_stock(db, cart_data.product_id, cart_data.quantity):
            raise HTTPException(status_code=400, detail="商品库存不足或不可用")
        
        # 获取用户购物车
        cart = get_user_cart(db, current_user)
        
        # 检查是否已存在相同商品
        existing_item = db.query(CartItem).filter(
            CartItem.cart_id == cart.id,
            CartItem.product_id == cart_data.product_id
        ).first()
        
        if existing_item:
            # 更新数量
            new_quantity = existing_item.quantity + cart_data.quantity
            
            # 再次检查库存
            if not check_product_stock(db, cart_data.product_id, new_quantity):
                raise HTTPException(status_code=400, detail="商品库存不足")
            
            existing_item.quantity = new_quantity
            existing_item.product_attributes = json.dumps(cart_data.product_attributes) if cart_data.product_attributes else None
        else:
            # 创建新项目
            cart_item_id = generate_cart_item_id()
            cart_item = CartItem(
                id=cart_item_id,
                cart_id=cart.id,
                product_id=cart_data.product_id,
                quantity=cart_data.quantity,
                product_attributes=json.dumps(cart_data.product_attributes) if cart_data.product_attributes else None
            )
            db.add(cart_item)
        
        db.commit()
        
        return SuccessResponse(msg="商品已添加到购物车")
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"添加购物车失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="添加购物车失败")

@router.put("/cart/{item_id}", response_model=SuccessResponse)
async def update_cart_item(
    item_id: str,
    update_data: CartItemUpdate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """更新购物车商品"""
    try:
        # 获取购物车项目
        cart_item = db.query(CartItem).join(Cart).filter(
            CartItem.id == item_id,
            Cart.user_id == current_user
        ).first()
        
        if not cart_item:
            raise HTTPException(status_code=404, detail="购物车项目不存在")
        
        # 检查库存
        if not check_product_stock(db, cart_item.product_id, update_data.quantity):
            raise HTTPException(status_code=400, detail="商品库存不足")
        
        # 更新数量和属性
        cart_item.quantity = update_data.quantity
        if update_data.product_attributes is not None:
            cart_item.product_attributes = json.dumps(update_data.product_attributes)
        
        db.commit()
        
        return SuccessResponse(msg="购物车更新成功")
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"更新购物车失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="更新购物车失败")

@router.delete("/cart/{item_id}", response_model=SuccessResponse)
async def remove_cart_item(
    item_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """删除购物车商品"""
    try:
        cart_item = db.query(CartItem).join(Cart).filter(
            CartItem.id == item_id,
            Cart.user_id == current_user
        ).first()
        
        if not cart_item:
            raise HTTPException(status_code=404, detail="购物车项目不存在")
        
        db.delete(cart_item)
        db.commit()
        
        return SuccessResponse(msg="商品已从购物车移除")
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"删除购物车项目失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="删除购物车项目失败")

# ============= 订单相关API =============

@router.post("/orders", response_model=SuccessResponse)
async def create_order(
    order_data: OrderCreate,
    background_tasks: BackgroundTasks,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """创建订单"""
    try:
        # 验证商品库存
        for item in order_data.items:
            if not check_product_stock(db, item.product_id, item.quantity):
                product = db.query(Product).filter(Product.id == item.product_id).first()
                product_name = product.name if product else item.product_id
                raise HTTPException(status_code=400, detail=f"商品 {product_name} 库存不足")
        
        # 计算订单金额
        subtotal = 0.0
        order_items_data = []
        
        for item in order_data.items:
            product = db.query(Product).filter(Product.id == item.product_id).first()
            if not product:
                raise HTTPException(status_code=404, detail=f"商品 {item.product_id} 不存在")
            
            item_total = product.price * item.quantity
            subtotal += item_total
            
            order_items_data.append({
                "product": product,
                "quantity": item.quantity,
                "unit_price": product.price,
                "total_price": item_total,
                "product_attributes": item.product_attributes
            })
        
        # 计算运费和税费
        shipping_fee = calculate_shipping_fee(subtotal, order_data.shipping_method or "standard")
        tax_fee = calculate_tax_fee(subtotal)
        
        # 处理优惠券
        discount_amount = 0.0
        coupon_discount = 0.0
        coupon_description = ""
        
        if order_data.coupon_code:
            is_valid, discount, description = validate_coupon_code(order_data.coupon_code, subtotal)
            if is_valid:
                discount_amount = discount
                coupon_discount = discount
                coupon_description = description
        
        # 计算总金额
        total_amount = subtotal + shipping_fee + tax_fee - discount_amount
        
        # 创建订单
        order_id = generate_unique_order_id(db)
        order = Order(
            id=order_id,
            user_id=current_user,
            status=OrderStatus.PENDING.value,
            payment_status=PaymentStatus.PENDING.value,
            subtotal=subtotal,
            shipping_fee=shipping_fee,
            tax_fee=tax_fee,
            discount_amount=discount_amount,
            total_amount=total_amount,
            shipping_name=order_data.shipping_address.name,
            shipping_phone=order_data.shipping_address.phone,
            shipping_address=order_data.shipping_address.address,
            shipping_city=order_data.shipping_address.city,
            shipping_province=order_data.shipping_address.province,
            shipping_postal_code=order_data.shipping_address.postal_code,
            shipping_method=order_data.shipping_method,
            user_note=order_data.user_note,
            coupon_code=order_data.coupon_code,
            coupon_discount=coupon_discount,
            payment_method=order_data.payment_method
        )
        
        db.add(order)
        db.flush()  # 获取订单ID
        
        # 创建订单项目
        for item_data in order_items_data:
            order_item_id = generate_order_item_id()
            order_item = OrderItem(
                id=order_item_id,
                order_id=order.id,
                product_id=item_data["product"].id,
                product_name=item_data["product"].name,
                product_image=item_data["product"].main_image,
                unit_price=item_data["unit_price"],
                quantity=item_data["quantity"],
                total_price=item_data["total_price"],
                product_attributes=json.dumps(item_data["product_attributes"]) if item_data["product_attributes"] else None
            )
            db.add(order_item)
            
            # 减少库存
            update_product_stock(db, item_data["product"].id, -item_data["quantity"])
        
        db.commit()
        
        # 发送通知
        background_tasks.add_task(
            websocket_publisher.send_event,
            "ORDER_CREATED",
            {
                "order_id": order_id,
                "total_amount": total_amount,
                "message": f"订单 {order_id} 创建成功，总金额 ¥{total_amount}"
            },
            current_user
        )
        
        return SuccessResponse(
            msg="订单创建成功",
            data={"order_id": order_id, "total_amount": total_amount}
        )
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"创建订单失败: {e}")
        db.rollback()
        raise HTTPException(status_code=500, detail="创建订单失败")

@router.get("/orders", response_model=OrderListResponse)
async def get_user_orders(
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
    status: Optional[str] = Query(None, description="订单状态过滤"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取用户订单列表"""
    try:
        query = db.query(Order).filter(Order.user_id == current_user)
        
        if status:
            query = query.filter(Order.status == status)
        
        # 计算总数
        total = query.count()
        
        # 分页查询
        offset = (page - 1) * limit
        orders = query.order_by(Order.created_at.desc()).offset(offset).limit(limit).all()
        
        # 转换为响应格式
        order_responses = []
        for order in orders:
            order_data = order.to_dict()
            
            # 获取订单项目
            order_items = []
            for item in order.order_items:
                order_items.append(item.to_dict())
            
            order_data["items"] = order_items
            order_responses.append(OrderResponse(**order_data))
        
        pages = (total + limit - 1) // limit
        
        return OrderListResponse(
            orders=order_responses,
            total=total,
            page=page,
            pages=pages
        )
        
    except Exception as e:
        print(f"获取订单列表失败: {e}")
        raise HTTPException(status_code=500, detail="获取订单列表失败")

@router.get("/orders/{order_id}", response_model=OrderResponse)
async def get_order(
    order_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取订单详情"""
    try:
        order = db.query(Order).filter(Order.id == order_id).first()
        if not order:
            raise HTTPException(status_code=404, detail="订单不存在")
        
        # 验证权限
        permissions = validate_order_permissions(order, current_user)
        if not permissions["can_view"]:
            raise HTTPException(status_code=403, detail="无权查看此订单")
        
        # 转换为响应格式
        order_data = order.to_dict()
        
        # 获取订单项目
        order_items = []
        for item in order.order_items:
            order_items.append(item.to_dict())
        
        order_data["items"] = order_items
        
        return OrderResponse(**order_data)
        
    except HTTPException:
        raise
    except Exception as e:
        print(f"获取订单详情失败: {e}")
        raise HTTPException(status_code=500, detail="获取订单详情失败")

# ============= 推荐和统计API =============

@router.get("/recommendations", response_model=Dict[str, List[ProductListItem]])
async def get_recommendations(
    current_user: Optional[str] = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取推荐商品"""
    try:
        result = {}
        
        # 热门商品
        popular_products = get_popular_products(db, 10)
        result["popular"] = [ProductListItem(**p.to_dict(include_description=False)) for p in popular_products]
        
        # 精选商品
        featured_products = get_featured_products(db, 8)
        result["featured"] = [ProductListItem(**p.to_dict(include_description=False)) for p in featured_products]
        
        # 个性化推荐（需要登录）
        if current_user:
            recommended_products = get_recommended_products(db, current_user, 10)
            result["recommended"] = [ProductListItem(**p.to_dict(include_description=False)) for p in recommended_products]
        
        return result
        
    except Exception as e:
        print(f"获取推荐商品失败: {e}")
        raise HTTPException(status_code=500, detail="获取推荐商品失败")

@router.get("/stats", response_model=ShopStatsResponse)
async def get_shop_stats(
    admin_user: str = Depends(check_admin_permission),
    db: Session = Depends(get_db)
):
    """获取商店统计数据（管理员）"""
    try:
        
        stats = get_shop_statistics(db)
        
        # 获取热门商品
        popular_products = get_popular_products(db, 5)
        stats["popular_products"] = [ProductListItem(**p.to_dict(include_description=False)) for p in popular_products]
        
        # 获取最近订单
        recent_orders = db.query(Order).order_by(Order.created_at.desc()).limit(5).all()
        order_responses = []
        for order in recent_orders:
            order_data = order.to_dict()
            order_data["items"] = [item.to_dict() for item in order.order_items]
            order_responses.append(OrderResponse(**order_data))
        
        stats["recent_orders"] = order_responses
        
        # 销售趋势数据（简化版）
        stats["sales_trend"] = []
        
        return ShopStatsResponse(**stats)
        
    except Exception as e:
        print(f"获取商店统计失败: {e}")
        raise HTTPException(status_code=500, detail="获取商店统计失败")