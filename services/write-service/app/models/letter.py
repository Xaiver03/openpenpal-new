from sqlalchemy import Column, String, Text, DateTime, Enum, Boolean, Integer
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from datetime import datetime
from enum import Enum as PyEnum
from app.core.database import Base

class LetterStatus(PyEnum):
    """信件状态枚举"""
    DRAFT = "draft"           # 草稿
    GENERATED = "generated"   # 已生成二维码
    COLLECTED = "collected"   # 已收取  
    IN_TRANSIT = "in_transit" # 投递中
    DELIVERED = "delivered"   # 已投递
    FAILED = "failed"         # 投递失败

class Priority(PyEnum):
    """优先级枚举"""
    NORMAL = "normal"         # 普通
    URGENT = "urgent"         # 紧急

class Letter(Base):
    """信件数据模型"""
    __tablename__ = "letters"
    
    # 主键 - 信件编号 (OP + 10位随机字符)
    id = Column(String(20), primary_key=True, index=True)
    
    # 基础信息
    title = Column(String(200), nullable=False, comment="信件标题")
    content = Column(Text, nullable=False, comment="信件内容")
    
    # 发送者信息
    sender_id = Column(String(50), nullable=False, index=True, comment="发送者用户ID")
    sender_nickname = Column(String(100), comment="发送者昵称")
    
    # 接收者信息
    receiver_hint = Column(String(200), comment="接收者提示信息")
    
    # 状态和优先级
    status = Column(Enum(LetterStatus), default=LetterStatus.DRAFT, nullable=False, index=True, comment="信件状态")
    priority = Column(Enum(Priority), default=Priority.NORMAL, nullable=False, comment="优先级")
    
    # 配置选项
    anonymous = Column(Boolean, default=False, nullable=False, comment="是否匿名")
    
    # 投递说明
    delivery_instructions = Column(Text, comment="投递说明")
    
    # 统计信息
    read_count = Column(Integer, default=0, comment="阅读次数")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    read_logs = relationship("ReadLog", back_populates="letter", cascade="all, delete-orphan")
    plaza_posts = relationship("PlazaPost", back_populates="letter", cascade="all, delete-orphan")
    museum_letters = relationship("MuseumLetter", back_populates="letter", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Letter(id={self.id}, title={self.title}, status={self.status})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "title": self.title,
            "content": self.content,
            "sender_id": self.sender_id,
            "sender_nickname": self.sender_nickname,
            "receiver_hint": self.receiver_hint,
            "status": self.status.value if self.status else None,
            "priority": self.priority.value if self.priority else None,
            "anonymous": self.anonymous,
            "delivery_instructions": self.delivery_instructions,
            "read_count": self.read_count,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }