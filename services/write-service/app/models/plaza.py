from sqlalchemy import Column, String, Text, DateTime, Boolean, Integer, ForeignKey
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from datetime import datetime
from enum import Enum as PyEnum
from app.core.database import Base

class PostCategory(PyEnum):
    """帖子分类枚举"""
    LETTERS = "letters"         # 信件作品
    POETRY = "poetry"          # 诗歌
    PROSE = "prose"            # 散文
    STORIES = "stories"        # 故事
    THOUGHTS = "thoughts"      # 感想
    OTHERS = "others"          # 其他

class PostStatus(PyEnum):
    """帖子状态枚举"""
    DRAFT = "draft"            # 草稿
    PUBLISHED = "published"    # 已发布
    FEATURED = "featured"      # 精选
    HIDDEN = "hidden"          # 隐藏

class PlazaPost(Base):
    """写作广场帖子模型"""
    __tablename__ = "plaza_posts"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="帖子ID")
    
    # 基础信息
    title = Column(String(200), nullable=False, comment="帖子标题")
    content = Column(Text, nullable=False, comment="帖子内容")
    excerpt = Column(String(500), comment="摘要")
    
    # 作者信息
    author_id = Column(String(50), nullable=False, index=True, comment="作者用户ID")
    author_nickname = Column(String(100), comment="作者昵称")
    
    # 分类和标签
    category = Column(String(20), nullable=False, index=True, comment="帖子分类")
    tags = Column(String(200), comment="标签(逗号分隔)")
    
    # 状态
    status = Column(String(20), default=PostStatus.PUBLISHED.value, nullable=False, index=True, comment="帖子状态")
    
    # 配置选项
    allow_comments = Column(Boolean, default=True, comment="是否允许评论")
    anonymous = Column(Boolean, default=False, comment="是否匿名发布")
    
    # 统计信息
    view_count = Column(Integer, default=0, comment="浏览次数")
    like_count = Column(Integer, default=0, comment="点赞次数")
    comment_count = Column(Integer, default=0, comment="评论次数")
    favorite_count = Column(Integer, default=0, comment="收藏次数")
    
    # 关联的信件ID（如果是基于信件创建的帖子）
    letter_id = Column(String(20), ForeignKey("letters.id"), nullable=True, comment="关联信件ID")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    published_at = Column(DateTime(timezone=True), comment="发布时间")
    
    # 关联关系
    letter = relationship("Letter", back_populates="plaza_posts")
    likes = relationship("PlazaLike", back_populates="post", cascade="all, delete-orphan")
    comments = relationship("PlazaComment", back_populates="post", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<PlazaPost(id={self.id}, title={self.title}, author={self.author_nickname})>"
    
    def to_dict(self, include_content=True):
        """转换为字典格式"""
        data = {
            "id": self.id,
            "title": self.title,
            "excerpt": self.excerpt,
            "author_id": self.author_id if not self.anonymous else None,
            "author_nickname": self.author_nickname if not self.anonymous else "匿名用户",
            "category": self.category,
            "tags": self.tags.split(',') if self.tags else [],
            "status": self.status,
            "allow_comments": self.allow_comments,
            "anonymous": self.anonymous,
            "view_count": self.view_count,
            "like_count": self.like_count,
            "comment_count": self.comment_count,
            "favorite_count": self.favorite_count,
            "letter_id": self.letter_id,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "published_at": self.published_at.isoformat() if self.published_at else None
        }
        
        if include_content:
            data["content"] = self.content
            
        return data

class PlazaLike(Base):
    """帖子点赞记录"""
    __tablename__ = "plaza_likes"
    
    # 联合主键
    post_id = Column(String(20), ForeignKey("plaza_posts.id"), primary_key=True, comment="帖子ID")
    user_id = Column(String(50), primary_key=True, comment="用户ID")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="点赞时间")
    
    # 关联关系
    post = relationship("PlazaPost", back_populates="likes")
    
    def __repr__(self):
        return f"<PlazaLike(post_id={self.post_id}, user_id={self.user_id})>"

class PlazaComment(Base):
    """帖子评论"""
    __tablename__ = "plaza_comments"
    
    # 主键
    id = Column(String(20), primary_key=True, index=True, comment="评论ID")
    
    # 关联信息
    post_id = Column(String(20), ForeignKey("plaza_posts.id"), nullable=False, index=True, comment="帖子ID")
    user_id = Column(String(50), nullable=False, comment="评论用户ID")
    user_nickname = Column(String(100), comment="评论用户昵称")
    
    # 评论内容
    content = Column(Text, nullable=False, comment="评论内容")
    
    # 回复相关
    parent_id = Column(String(20), ForeignKey("plaza_comments.id"), nullable=True, comment="父评论ID")
    reply_to_user = Column(String(100), comment="回复的用户昵称")
    
    # 状态
    is_deleted = Column(Boolean, default=False, comment="是否已删除")
    
    # 统计
    like_count = Column(Integer, default=0, comment="点赞次数")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    # 关联关系
    post = relationship("PlazaPost", back_populates="comments")
    parent = relationship("PlazaComment", remote_side=[id], back_populates="replies")
    replies = relationship("PlazaComment", back_populates="parent", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<PlazaComment(id={self.id}, post_id={self.post_id}, user={self.user_nickname})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "post_id": self.post_id,
            "user_id": self.user_id,
            "user_nickname": self.user_nickname,
            "content": self.content if not self.is_deleted else "[评论已删除]",
            "parent_id": self.parent_id,
            "reply_to_user": self.reply_to_user,
            "is_deleted": self.is_deleted,
            "like_count": self.like_count,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }

class PlazaCategory(Base):
    """广场分类配置"""
    __tablename__ = "plaza_categories"
    
    # 主键
    id = Column(String(20), primary_key=True, comment="分类ID")
    
    # 分类信息
    name = Column(String(50), nullable=False, comment="分类名称")
    description = Column(String(200), comment="分类描述")
    icon = Column(String(50), comment="分类图标")
    color = Column(String(20), comment="分类颜色")
    
    # 配置
    is_active = Column(Boolean, default=True, comment="是否启用")
    sort_order = Column(Integer, default=0, comment="排序顺序")
    
    # 统计
    post_count = Column(Integer, default=0, comment="帖子数量")
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    
    def __repr__(self):
        return f"<PlazaCategory(id={self.id}, name={self.name})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "icon": self.icon,
            "color": self.color,
            "is_active": self.is_active,
            "sort_order": self.sort_order,
            "post_count": self.post_count,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None
        }