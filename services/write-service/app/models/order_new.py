"""
重构后的订单模型 - 配合SPU+SKU设计

支持新的商品架构，订单项直接关联SKU而不是SPU
"""

from sqlalchemy import Column, String, Text, DateTime, Boolean, Integer, ForeignKey, Float, Enum, JSON, Index
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship, validates
from datetime import datetime
from enum import Enum as PyEnum
from app.core.database import Base


class OrderStatus(PyEnum):
    """订单状态枚举"""
    PENDING = "pending"        # 待付款
    PAID = "paid"              # 已付款
    PROCESSING = "processing"  # 处理中
    PACKED = "packed"          # 已打包
    SHIPPED = "shipped"        # 已发货
    DELIVERED = "delivered"    # 已送达
    COMPLETED = "completed"    # 已完成
    CANCELLED = "cancelled"    # 已取消
    REFUNDED = "refunded"      # 已退款


class PaymentStatus(PyEnum):
    """支付状态枚举"""
    PENDING = "pending"        # 待支付
    PROCESSING = "processing"  # 支付处理中
    SUCCESS = "success"        # 支付成功
    FAILED = "failed"          # 支付失败
    REFUNDED = "refunded"      # 已退款
    PARTIAL_REFUND = "partial_refund"  # 部分退款
    CANCELLED = "cancelled"    # 已取消


class ShippingStatus(PyEnum):
    """物流状态枚举"""
    PENDING = "pending"        # 待发货
    PACKED = "packed"          # 已打包
    SHIPPED = "shipped"        # 已发货
    IN_TRANSIT = "in_transit"  # 运输中
    OUT_FOR_DELIVERY = "out_for_delivery"  # 派送中
    DELIVERED = "delivered"    # 已送达
    FAILED_DELIVERY = "failed_delivery"    # 投递失败
    RETURNED = "returned"      # 已退回


class RefundStatus(PyEnum):
    """退款状态枚举"""
    NONE = "none"              # 无退款
    REQUESTED = "requested"    # 已申请
    PROCESSING = "processing"  # 处理中
    APPROVED = "approved"      # 已批准
    REJECTED = "rejected"      # 已拒绝
    COMPLETED = "completed"    # 已完成


# ==================== 订单主表 ====================

class Order(Base):
    """订单模型"""
    __tablename__ = "orders"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="订单ID")
    order_no = Column(String(32), unique=True, nullable=False, index=True, comment="订单编号")
    
    # 用户信息
    user_id = Column(String(50), nullable=False, index=True, comment="用户ID")
    user_name = Column(String(100), comment="用户姓名")
    user_email = Column(String(200), comment="用户邮箱")
    user_phone = Column(String(20), comment="用户电话")
    
    # 订单状态
    status = Column(String(20), default=OrderStatus.PENDING.value, nullable=False, index=True, comment="订单状态")
    payment_status = Column(String(20), default=PaymentStatus.PENDING.value, nullable=False, index=True, comment="支付状态")
    shipping_status = Column(String(20), default=ShippingStatus.PENDING.value, nullable=False, comment="物流状态")
    refund_status = Column(String(20), default=RefundStatus.NONE.value, nullable=False, comment="退款状态")
    
    # 金额信息
    item_count = Column(Integer, nullable=False, comment="商品总件数")
    subtotal = Column(Float, nullable=False, comment="商品小计")
    shipping_fee = Column(Float, default=0.0, comment="运费")
    tax_fee = Column(Float, default=0.0, comment="税费")
    discount_amount = Column(Float, default=0.0, comment="优惠金额")
    coupon_discount = Column(Float, default=0.0, comment="优惠券折扣")
    total_amount = Column(Float, nullable=False, comment="订单总金额")
    actual_payment = Column(Float, comment="实际支付金额")
    currency = Column(String(3), default="CNY", comment="货币类型")
    
    # 收货信息
    shipping_info = Column(JSON, comment="收货信息JSON")
    shipping_method = Column(String(50), comment="配送方式")
    shipping_fee_template = Column(String(50), comment="运费模板")
    
    # 订单备注
    user_note = Column(String(500), comment="用户备注")
    admin_note = Column(String(500), comment="管理员备注")
    internal_note = Column(String(500), comment="内部备注")
    
    # 优惠信息
    promotion_info = Column(JSON, comment="促销信息JSON")
    coupon_info = Column(JSON, comment="优惠券信息JSON")
    
    # 支付信息
    payment_method = Column(String(50), comment="支付方式")
    payment_info = Column(JSON, comment="支付信息JSON")
    
    # 物流信息
    shipping_info_detail = Column(JSON, comment="物流详情JSON")
    
    # 订单来源
    source = Column(String(50), default="web", comment="订单来源")
    platform = Column(String(50), comment="平台标识")
    
    # 重要时间节点
    paid_at = Column(DateTime(timezone=True), comment="支付时间")
    shipped_at = Column(DateTime(timezone=True), comment="发货时间")
    delivered_at = Column(DateTime(timezone=True), comment="签收时间")
    completed_at = Column(DateTime(timezone=True), comment="完成时间")
    cancelled_at = Column(DateTime(timezone=True), comment="取消时间")
    
    # 自动确认收货时间
    auto_confirm_time = Column(DateTime(timezone=True), comment="自动确认收货时间")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    order_items = relationship("OrderItem", back_populates="order", cascade="all, delete-orphan")
    order_logs = relationship("OrderLog", back_populates="order", cascade="all, delete-orphan")
    payments = relationship("Payment", back_populates="order", cascade="all, delete-orphan")
    refunds = relationship("Refund", back_populates="order", cascade="all, delete-orphan")
    
    # 索引
    __table_args__ = (
        Index('idx_order_user_created', 'user_id', 'created_at'),
        Index('idx_order_status_created', 'status', 'created_at'),
        Index('idx_order_payment_status', 'payment_status'),
        Index('idx_order_no', 'order_no'),
    )
    
    def __repr__(self):
        return f"<Order(id={self.id}, order_no={self.order_no}, total={self.total_amount})>"
    
    @validates('shipping_info', 'promotion_info', 'coupon_info', 'payment_info', 'shipping_info_detail')
    def validate_json_fields(self, key, value):
        """验证JSON字段格式"""
        if value is not None and not isinstance(value, (dict, list)):
            raise ValueError(f"{key} must be valid JSON")
        return value
    
    def to_dict(self, include_items=False, include_logs=False):
        """转换为字典格式"""
        data = {
            "id": self.id,
            "order_no": self.order_no,
            "user_id": self.user_id,
            "user_name": self.user_name,
            "user_email": self.user_email,
            "user_phone": self.user_phone,
            "status": self.status,
            "payment_status": self.payment_status,
            "shipping_status": self.shipping_status,
            "refund_status": self.refund_status,
            "item_count": self.item_count,
            "subtotal": self.subtotal,
            "shipping_fee": self.shipping_fee,
            "tax_fee": self.tax_fee,
            "discount_amount": self.discount_amount,
            "coupon_discount": self.coupon_discount,
            "total_amount": self.total_amount,
            "actual_payment": self.actual_payment,
            "currency": self.currency,
            "shipping_info": self.shipping_info,
            "shipping_method": self.shipping_method,
            "user_note": self.user_note,
            "admin_note": self.admin_note,
            "promotion_info": self.promotion_info,
            "coupon_info": self.coupon_info,
            "payment_method": self.payment_method,
            "payment_info": self.payment_info,
            "shipping_info_detail": self.shipping_info_detail,
            "source": self.source,
            "platform": self.platform,
            "paid_at": self.paid_at.isoformat() if self.paid_at else None,
            "shipped_at": self.shipped_at.isoformat() if self.shipped_at else None,
            "delivered_at": self.delivered_at.isoformat() if self.delivered_at else None,
            "completed_at": self.completed_at.isoformat() if self.completed_at else None,
            "cancelled_at": self.cancelled_at.isoformat() if self.cancelled_at else None,
            "auto_confirm_time": self.auto_confirm_time.isoformat() if self.auto_confirm_time else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }
        
        if include_items and self.order_items:
            data["order_items"] = [item.to_dict() for item in self.order_items]
        
        if include_logs and self.order_logs:
            data["order_logs"] = [log.to_dict() for log in self.order_logs]
        
        return data


# ==================== 订单商品项 ====================

class OrderItem(Base):
    """订单商品项模型"""
    __tablename__ = "order_items"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="订单项ID")
    
    # 关联信息
    order_id = Column(String(20), ForeignKey("orders.id"), nullable=False, index=True, comment="订单ID")
    spu_id = Column(String(20), ForeignKey("product_spu.id"), nullable=False, comment="SPU ID")
    sku_id = Column(String(20), ForeignKey("product_sku.id"), nullable=False, index=True, comment="SKU ID")
    
    # 商品信息快照（下单时的商品信息）
    spu_name = Column(String(200), nullable=False, comment="SPU名称")
    sku_name = Column(String(200), comment="SKU名称")
    sku_code = Column(String(100), comment="SKU编码")
    spu_image = Column(String(500), comment="SPU主图")
    sku_image = Column(String(500), comment="SKU图片")
    
    # 商品属性快照
    sale_attributes = Column(JSON, comment="销售属性快照")
    basic_attributes = Column(JSON, comment="基本属性快照")
    
    # 价格和数量
    unit_price = Column(Float, nullable=False, comment="单价")
    original_price = Column(Float, comment="原价")
    quantity = Column(Integer, nullable=False, comment="购买数量")
    total_price = Column(Float, nullable=False, comment="小计金额")
    
    # 优惠信息
    item_discount = Column(Float, default=0.0, comment="单品优惠金额")
    promotion_info = Column(JSON, comment="促销信息")
    
    # 商品状态
    item_status = Column(String(20), comment="商品项状态")
    
    # 退款信息
    refund_quantity = Column(Integer, default=0, comment="退款数量")
    refund_amount = Column(Float, default=0.0, comment="退款金额")
    
    # 物理属性（用于计算运费）
    weight = Column(Float, comment="重量")
    volume = Column(Float, comment="体积")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    
    # 关联关系
    order = relationship("Order", back_populates="order_items")
    spu = relationship("ProductSPU")
    sku = relationship("ProductSKU")
    
    # 索引
    __table_args__ = (
        Index('idx_order_item_order', 'order_id'),
        Index('idx_order_item_sku', 'sku_id'),
        Index('idx_order_item_spu', 'spu_id'),
    )
    
    def __repr__(self):
        return f"<OrderItem(id={self.id}, sku_code={self.sku_code}, qty={self.quantity})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "order_id": self.order_id,
            "spu_id": self.spu_id,
            "sku_id": self.sku_id,
            "spu_name": self.spu_name,
            "sku_name": self.sku_name,
            "sku_code": self.sku_code,
            "spu_image": self.spu_image,
            "sku_image": self.sku_image,
            "sale_attributes": self.sale_attributes,
            "basic_attributes": self.basic_attributes,
            "unit_price": self.unit_price,
            "original_price": self.original_price,
            "quantity": self.quantity,
            "total_price": self.total_price,
            "item_discount": self.item_discount,
            "promotion_info": self.promotion_info,
            "item_status": self.item_status,
            "refund_quantity": self.refund_quantity,
            "refund_amount": self.refund_amount,
            "weight": self.weight,
            "volume": self.volume,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }


# ==================== 购物车 ====================

class Cart(Base):
    """购物车模型"""
    __tablename__ = "carts"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="购物车ID")
    
    # 用户信息
    user_id = Column(String(50), nullable=False, unique=True, index=True, comment="用户ID")
    
    # 购物车配置
    total_items = Column(Integer, default=0, comment="商品总件数")
    total_amount = Column(Float, default=0.0, comment="商品总金额")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    cart_items = relationship("CartItem", back_populates="cart", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Cart(id={self.id}, user_id={self.user_id}, items={self.total_items})>"
    
    def to_dict(self, include_items=False):
        """转换为字典格式"""
        data = {
            "id": self.id,
            "user_id": self.user_id,
            "total_items": self.total_items,
            "total_amount": self.total_amount,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }
        
        if include_items and self.cart_items:
            data["cart_items"] = [item.to_dict() for item in self.cart_items]
        
        return data


class CartItem(Base):
    """购物车商品项模型"""
    __tablename__ = "cart_items"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="购物车项ID")
    
    # 关联信息
    cart_id = Column(String(20), ForeignKey("carts.id"), nullable=False, index=True, comment="购物车ID")
    spu_id = Column(String(20), ForeignKey("product_spu.id"), nullable=False, comment="SPU ID")
    sku_id = Column(String(20), ForeignKey("product_sku.id"), nullable=False, index=True, comment="SKU ID")
    
    # 数量和属性
    quantity = Column(Integer, nullable=False, comment="数量")
    
    # 选中状态
    is_selected = Column(Boolean, default=True, comment="是否选中（用于结算）")
    
    # 价格快照（防止价格波动影响用户体验）
    snapshot_price = Column(Float, comment="价格快照")
    snapshot_time = Column(DateTime(timezone=True), comment="快照时间")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="添加时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    cart = relationship("Cart", back_populates="cart_items")
    spu = relationship("ProductSPU")
    sku = relationship("ProductSKU")
    
    # 索引
    __table_args__ = (
        Index('idx_cart_item_cart', 'cart_id'),
        Index('idx_cart_item_sku', 'sku_id'),
        Index('idx_cart_item_updated', 'updated_at'),
    )
    
    def __repr__(self):
        return f"<CartItem(id={self.id}, sku_id={self.sku_id}, qty={self.quantity})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "cart_id": self.cart_id,
            "spu_id": self.spu_id,
            "sku_id": self.sku_id,
            "quantity": self.quantity,
            "is_selected": self.is_selected,
            "snapshot_price": self.snapshot_price,
            "snapshot_time": self.snapshot_time.isoformat() if self.snapshot_time else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }


# ==================== 订单日志 ====================

class OrderLog(Base):
    """订单日志模型"""
    __tablename__ = "order_logs"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="日志ID")
    
    # 关联订单
    order_id = Column(String(20), ForeignKey("orders.id"), nullable=False, index=True, comment="订单ID")
    
    # 日志信息
    action_type = Column(String(50), nullable=False, comment="操作类型")
    action_desc = Column(String(200), comment="操作描述")
    old_value = Column(Text, comment="原值")
    new_value = Column(Text, comment="新值")
    
    # 操作人信息
    operator_type = Column(String(20), comment="操作人类型（user/admin/system）")
    operator_id = Column(String(50), comment="操作人ID")
    operator_name = Column(String(100), comment="操作人姓名")
    
    # 客户端信息
    ip_address = Column(String(45), comment="IP地址")
    user_agent = Column(String(500), comment="用户代理")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="操作时间")
    
    # 关联关系
    order = relationship("Order", back_populates="order_logs")
    
    # 索引
    __table_args__ = (
        Index('idx_order_log_order_created', 'order_id', 'created_at'),
        Index('idx_order_log_action_type', 'action_type'),
    )
    
    def __repr__(self):
        return f"<OrderLog(id={self.id}, order_id={self.order_id}, action={self.action_type})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "order_id": self.order_id,
            "action_type": self.action_type,
            "action_desc": self.action_desc,
            "old_value": self.old_value,
            "new_value": self.new_value,
            "operator_type": self.operator_type,
            "operator_id": self.operator_id,
            "operator_name": self.operator_name,
            "ip_address": self.ip_address,
            "user_agent": self.user_agent,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }


# ==================== 支付记录 ====================

class Payment(Base):
    """支付记录模型"""
    __tablename__ = "payments"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="支付记录ID")
    
    # 关联订单
    order_id = Column(String(20), ForeignKey("orders.id"), nullable=False, index=True, comment="订单ID")
    
    # 支付信息
    payment_no = Column(String(32), unique=True, nullable=False, comment="支付流水号")
    payment_method = Column(String(50), nullable=False, comment="支付方式")
    payment_channel = Column(String(50), comment="支付渠道")
    
    # 金额信息
    payment_amount = Column(Float, nullable=False, comment="支付金额")
    actual_amount = Column(Float, comment="实际到账金额")
    currency = Column(String(3), default="CNY", comment="货币类型")
    
    # 支付状态
    status = Column(String(20), default=PaymentStatus.PENDING.value, nullable=False, comment="支付状态")
    
    # 第三方支付信息
    third_party_no = Column(String(100), comment="第三方支付单号")
    third_party_response = Column(JSON, comment="第三方支付响应")
    
    # 支付时间
    paid_at = Column(DateTime(timezone=True), comment="支付完成时间")
    notify_at = Column(DateTime(timezone=True), comment="异步通知时间")
    
    # 失败信息
    fail_reason = Column(String(200), comment="失败原因")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    order = relationship("Order", back_populates="payments")
    
    # 索引
    __table_args__ = (
        Index('idx_payment_order', 'order_id'),
        Index('idx_payment_no', 'payment_no'),
        Index('idx_payment_third_party', 'third_party_no'),
        Index('idx_payment_status', 'status'),
    )
    
    def __repr__(self):
        return f"<Payment(id={self.id}, order_id={self.order_id}, amount={self.payment_amount})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "order_id": self.order_id,
            "payment_no": self.payment_no,
            "payment_method": self.payment_method,
            "payment_channel": self.payment_channel,
            "payment_amount": self.payment_amount,
            "actual_amount": self.actual_amount,
            "currency": self.currency,
            "status": self.status,
            "third_party_no": self.third_party_no,
            "third_party_response": self.third_party_response,
            "paid_at": self.paid_at.isoformat() if self.paid_at else None,
            "notify_at": self.notify_at.isoformat() if self.notify_at else None,
            "fail_reason": self.fail_reason,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }


# ==================== 退款记录 ====================

class Refund(Base):
    """退款记录模型"""
    __tablename__ = "refunds"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="退款记录ID")
    
    # 关联订单
    order_id = Column(String(20), ForeignKey("orders.id"), nullable=False, index=True, comment="订单ID")
    
    # 退款信息
    refund_no = Column(String(32), unique=True, nullable=False, comment="退款单号")
    refund_type = Column(String(20), nullable=False, comment="退款类型（全额退款/部分退款）")
    refund_reason = Column(String(200), comment="退款原因")
    
    # 金额信息
    refund_amount = Column(Float, nullable=False, comment="退款金额")
    actual_refund = Column(Float, comment="实际退款金额")
    
    # 退款状态
    status = Column(String(20), default=RefundStatus.REQUESTED.value, nullable=False, comment="退款状态")
    
    # 第三方退款信息
    third_party_refund_no = Column(String(100), comment="第三方退款单号")
    
    # 处理信息
    approved_by = Column(String(50), comment="审批人ID")
    approved_at = Column(DateTime(timezone=True), comment="审批时间")
    approved_note = Column(String(500), comment="审批备注")
    
    # 退款时间
    refunded_at = Column(DateTime(timezone=True), comment="退款完成时间")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    order = relationship("Order", back_populates="refunds")
    
    # 索引
    __table_args__ = (
        Index('idx_refund_order', 'order_id'),
        Index('idx_refund_no', 'refund_no'),
        Index('idx_refund_status', 'status'),
    )
    
    def __repr__(self):
        return f"<Refund(id={self.id}, order_id={self.order_id}, amount={self.refund_amount})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "order_id": self.order_id,
            "refund_no": self.refund_no,
            "refund_type": self.refund_type,
            "refund_reason": self.refund_reason,
            "refund_amount": self.refund_amount,
            "actual_refund": self.actual_refund,
            "status": self.status,
            "third_party_refund_no": self.third_party_refund_no,
            "approved_by": self.approved_by,
            "approved_at": self.approved_at.isoformat() if self.approved_at else None,
            "approved_note": self.approved_note,
            "refunded_at": self.refunded_at.isoformat() if self.refunded_at else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }