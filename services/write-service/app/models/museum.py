from sqlalchemy import Column, String, Text, DateTime, Boolean, Integer, ForeignKey, Float
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from datetime import datetime
from enum import Enum as PyEnum
from app.core.database import Base

class MuseumLetterStatus(PyEnum):
    """博物馆信件状态枚举"""
    PENDING = "pending"        # 待审核
    APPROVED = "approved"      # 已通过
    FEATURED = "featured"      # 精选展示
    ARCHIVED = "archived"      # 已归档
    REJECTED = "rejected"      # 已拒绝

class MuseumEra(PyEnum):
    """历史时期枚举"""
    ANCIENT = "ancient"        # 古代 (1840年前)
    MODERN = "modern"          # 近代 (1840-1919)
    CONTEMPORARY = "contemporary"  # 现代 (1919-1949)
    PRESENT = "present"        # 当代 (1949至今)
    DIGITAL = "digital"        # 数字时代 (2000至今)

class MuseumLetter(Base):
    """博物馆信件模型"""
    __tablename__ = "museum_letters"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="博物馆信件ID")
    
    # 基础信息
    title = Column(String(200), nullable=False, comment="信件标题")
    content = Column(Text, nullable=False, comment="信件内容")
    summary = Column(String(500), comment="信件摘要")
    
    # 历史信息
    original_author = Column(String(100), comment="原作者")
    original_recipient = Column(String(100), comment="原收件人")
    historical_date = Column(DateTime, comment="历史日期")
    era = Column(String(20), nullable=False, index=True, comment="历史时期")
    location = Column(String(200), comment="地理位置")
    
    # 分类信息
    category = Column(String(50), nullable=False, index=True, comment="信件分类")
    tags = Column(String(300), comment="标签(逗号分隔)")
    language = Column(String(10), default="zh", comment="语言")
    
    # 来源信息
    source_type = Column(String(20), nullable=False, comment="来源类型")  # original/contributed/digitized
    source_description = Column(Text, comment="来源描述")
    contributor_id = Column(String(50), comment="贡献者ID")
    contributor_name = Column(String(100), comment="贡献者姓名")
    
    # 状态和审核
    status = Column(String(20), default=MuseumLetterStatus.PENDING.value, nullable=False, index=True, comment="状态")
    reviewer_id = Column(String(50), comment="审核员ID")
    review_note = Column(Text, comment="审核备注")
    reviewed_at = Column(DateTime, comment="审核时间")
    
    # 展示配置
    is_featured = Column(Boolean, default=False, comment="是否精选")
    display_order = Column(Integer, default=0, comment="展示顺序")
    featured_until = Column(DateTime, comment="精选截止时间")
    
    # 统计信息
    view_count = Column(Integer, default=0, comment="浏览次数")
    favorite_count = Column(Integer, default=0, comment="收藏次数")
    share_count = Column(Integer, default=0, comment="分享次数")
    rating_avg = Column(Float, default=0.0, comment="平均评分")
    rating_count = Column(Integer, default=0, comment="评分人数")
    
    # 关联信件ID（如果基于现代信件创建）
    letter_id = Column(String(20), ForeignKey("letters.id"), nullable=True, comment="关联信件ID")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    letter = relationship("Letter", back_populates="museum_letters")
    favorites = relationship("MuseumFavorite", back_populates="museum_letter", cascade="all, delete-orphan")
    ratings = relationship("MuseumRating", back_populates="museum_letter", cascade="all, delete-orphan")
    timeline_events = relationship("TimelineEvent", back_populates="museum_letter", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<MuseumLetter(id={self.id}, title={self.title}, era={self.era})>"
    
    def to_dict(self, include_content=True):
        """转换为字典格式"""
        data = {
            "id": self.id,
            "title": self.title,
            "summary": self.summary,
            "original_author": self.original_author,
            "original_recipient": self.original_recipient,
            "historical_date": self.historical_date.isoformat() if self.historical_date else None,
            "era": self.era,
            "location": self.location,
            "category": self.category,
            "tags": self.tags.split(',') if self.tags else [],
            "language": self.language,
            "source_type": self.source_type,
            "source_description": self.source_description,
            "contributor_name": self.contributor_name,
            "status": self.status,
            "is_featured": self.is_featured,
            "display_order": self.display_order,
            "view_count": self.view_count,
            "favorite_count": self.favorite_count,
            "share_count": self.share_count,
            "rating_avg": self.rating_avg,
            "rating_count": self.rating_count,
            "letter_id": self.letter_id,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }
        
        if include_content:
            data["content"] = self.content
            
        return data

class MuseumFavorite(Base):
    """博物馆信件收藏记录"""
    __tablename__ = "museum_favorites"
    
    # 联合主键
    museum_letter_id = Column(String(20), ForeignKey("museum_letters.id"), primary_key=True, comment="博物馆信件ID")
    user_id = Column(String(50), primary_key=True, comment="用户ID")
    
    # 收藏备注
    note = Column(String(200), comment="收藏备注")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="收藏时间")
    
    # 关联关系
    museum_letter = relationship("MuseumLetter", back_populates="favorites")
    
    def __repr__(self):
        return f"<MuseumFavorite(museum_letter_id={self.museum_letter_id}, user_id={self.user_id})>"

class MuseumRating(Base):
    """博物馆信件评分记录"""
    __tablename__ = "museum_ratings"
    
    # 联合主键
    museum_letter_id = Column(String(20), ForeignKey("museum_letters.id"), primary_key=True, comment="博物馆信件ID")
    user_id = Column(String(50), primary_key=True, comment="用户ID")
    
    # 评分信息
    rating = Column(Integer, nullable=False, comment="评分(1-5)")
    comment = Column(String(500), comment="评价评论")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="评分时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    museum_letter = relationship("MuseumLetter", back_populates="ratings")
    
    def __repr__(self):
        return f"<MuseumRating(museum_letter_id={self.museum_letter_id}, user_id={self.user_id}, rating={self.rating})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "museum_letter_id": self.museum_letter_id,
            "user_id": self.user_id,
            "rating": self.rating,
            "comment": self.comment,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class TimelineEvent(Base):
    """时间线事件模型"""
    __tablename__ = "timeline_events"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="事件ID")
    
    # 事件信息
    title = Column(String(200), nullable=False, comment="事件标题")
    description = Column(Text, comment="事件描述")
    event_date = Column(DateTime, nullable=False, index=True, comment="事件日期")
    era = Column(String(20), nullable=False, index=True, comment="历史时期")
    location = Column(String(200), comment="事件地点")
    
    # 事件类型
    event_type = Column(String(30), nullable=False, index=True, comment="事件类型")  # letter/historical/cultural/personal
    category = Column(String(50), comment="事件分类")
    importance = Column(Integer, default=1, comment="重要程度(1-5)")
    
    # 关联信件
    museum_letter_id = Column(String(20), ForeignKey("museum_letters.id"), nullable=True, comment="关联博物馆信件ID")
    
    # 展示配置
    is_featured = Column(Boolean, default=False, comment="是否在时间线突出显示")
    display_order = Column(Integer, default=0, comment="同日期事件的显示顺序")
    
    # 媒体资源
    image_url = Column(String(500), comment="事件图片URL")
    audio_url = Column(String(500), comment="事件音频URL")
    video_url = Column(String(500), comment="事件视频URL")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    museum_letter = relationship("MuseumLetter", back_populates="timeline_events")
    
    def __repr__(self):
        return f"<TimelineEvent(id={self.id}, title={self.title}, date={self.event_date})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "title": self.title,
            "description": self.description,
            "event_date": self.event_date.isoformat() if self.event_date else None,
            "era": self.era,
            "location": self.location,
            "event_type": self.event_type,
            "category": self.category,
            "importance": self.importance,
            "museum_letter_id": self.museum_letter_id,
            "is_featured": self.is_featured,
            "display_order": self.display_order,
            "image_url": self.image_url,
            "audio_url": self.audio_url,
            "video_url": self.video_url,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class MuseumCollection(Base):
    """博物馆收藏集合"""
    __tablename__ = "museum_collections"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="收藏集ID")
    
    # 基础信息
    name = Column(String(100), nullable=False, comment="收藏集名称")
    description = Column(Text, comment="收藏集描述")
    theme = Column(String(50), comment="主题")
    
    # 创建者信息
    creator_id = Column(String(50), nullable=False, comment="创建者ID")
    creator_name = Column(String(100), comment="创建者姓名")
    
    # 配置
    is_public = Column(Boolean, default=True, comment="是否公开")
    is_featured = Column(Boolean, default=False, comment="是否精选")
    
    # 统计
    letter_count = Column(Integer, default=0, comment="信件数量")
    view_count = Column(Integer, default=0, comment="浏览次数")
    follow_count = Column(Integer, default=0, comment="关注数量")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    def __repr__(self):
        return f"<MuseumCollection(id={self.id}, name={self.name})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "theme": self.theme,
            "creator_id": self.creator_id,
            "creator_name": self.creator_name,
            "is_public": self.is_public,
            "is_featured": self.is_featured,
            "letter_count": self.letter_count,
            "view_count": self.view_count,
            "follow_count": self.follow_count,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class CollectionLetter(Base):
    """收藏集与信件的关联关系"""
    __tablename__ = "collection_letters"
    
    # 联合主键
    collection_id = Column(String(20), ForeignKey("museum_collections.id"), primary_key=True, comment="收藏集ID")
    museum_letter_id = Column(String(20), ForeignKey("museum_letters.id"), primary_key=True, comment="博物馆信件ID")
    
    # 关联信息
    added_by = Column(String(50), comment="添加者ID")
    note = Column(String(200), comment="添加备注")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="添加时间")
    
    def __repr__(self):
        return f"<CollectionLetter(collection_id={self.collection_id}, museum_letter_id={self.museum_letter_id})>"