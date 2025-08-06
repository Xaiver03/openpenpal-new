from pydantic import BaseModel, Field, validator
from typing import Optional, List
from datetime import datetime
from enum import Enum

class PostCategory(str, Enum):
    """帖子分类枚举"""
    LETTERS = "letters"
    POETRY = "poetry"
    PROSE = "prose"
    STORIES = "stories"
    THOUGHTS = "thoughts"
    OTHERS = "others"

class PostStatus(str, Enum):
    """帖子状态枚举"""
    DRAFT = "draft"
    PUBLISHED = "published"
    FEATURED = "featured"
    HIDDEN = "hidden"

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

# 广场帖子相关Schema
class PlazaPostCreate(BaseModel):
    """创建广场帖子请求"""
    title: str = Field(..., min_length=1, max_length=200, description="帖子标题")
    content: str = Field(..., min_length=1, description="帖子内容")
    excerpt: Optional[str] = Field(None, max_length=500, description="摘要")
    category: PostCategory = Field(default=PostCategory.OTHERS, description="帖子分类")
    tags: Optional[List[str]] = Field(default=[], description="标签列表")
    allow_comments: bool = Field(default=True, description="是否允许评论")
    anonymous: bool = Field(default=False, description="是否匿名发布")
    letter_id: Optional[str] = Field(None, description="关联的信件ID")
    
    @validator('tags')
    def validate_tags(cls, v):
        if v is None:
            return []
        if len(v) > 10:
            raise ValueError('标签数量不能超过10个')
        for tag in v:
            if len(tag) > 20:
                raise ValueError('单个标签长度不能超过20个字符')
        return v
    
    @validator('content')
    def validate_content(cls, v):
        if len(v) > 50000:
            raise ValueError('内容长度不能超过50000个字符')
        return v

class PlazaPostUpdate(BaseModel):
    """更新广场帖子请求"""
    title: Optional[str] = Field(None, min_length=1, max_length=200, description="帖子标题")
    content: Optional[str] = Field(None, min_length=1, description="帖子内容")
    excerpt: Optional[str] = Field(None, max_length=500, description="摘要")
    category: Optional[PostCategory] = Field(None, description="帖子分类")
    tags: Optional[List[str]] = Field(None, description="标签列表")
    allow_comments: Optional[bool] = Field(None, description="是否允许评论")
    
    @validator('tags')
    def validate_tags(cls, v):
        if v is None:
            return v
        if len(v) > 10:
            raise ValueError('标签数量不能超过10个')
        for tag in v:
            if len(tag) > 20:
                raise ValueError('单个标签长度不能超过20个字符')
        return v

class PlazaPostResponse(BaseModel):
    """广场帖子响应"""
    id: str
    title: str
    content: Optional[str] = None
    excerpt: Optional[str] = None
    author_id: Optional[str] = None
    author_nickname: str
    category: str
    tags: List[str] = []
    status: str
    allow_comments: bool
    anonymous: bool
    view_count: int = 0
    like_count: int = 0
    comment_count: int = 0
    favorite_count: int = 0
    letter_id: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    published_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class PlazaPostListItem(BaseModel):
    """广场帖子列表项（不包含完整内容）"""
    id: str
    title: str
    excerpt: Optional[str] = None
    author_nickname: str
    category: str
    tags: List[str] = []
    status: str
    anonymous: bool
    view_count: int = 0
    like_count: int = 0
    comment_count: int = 0
    favorite_count: int = 0
    created_at: Optional[datetime] = None
    published_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class PlazaPostListResponse(BaseModel):
    """广场帖子列表响应"""
    posts: List[PlazaPostListItem]
    total: int
    page: int
    pages: int
    has_next: bool
    has_prev: bool

# 点赞相关Schema
class PlazaLikeCreate(BaseModel):
    """点赞请求"""
    post_id: str = Field(..., description="帖子ID")

class PlazaLikeResponse(BaseModel):
    """点赞响应"""
    post_id: str
    liked: bool
    like_count: int

# 评论相关Schema
class PlazaCommentCreate(BaseModel):
    """创建评论请求"""
    content: str = Field(..., min_length=1, max_length=1000, description="评论内容")
    parent_id: Optional[str] = Field(None, description="父评论ID")
    reply_to_user: Optional[str] = Field(None, description="回复的用户昵称")

class PlazaCommentResponse(BaseModel):
    """评论响应"""
    id: str
    post_id: str
    user_id: str
    user_nickname: str
    content: str
    parent_id: Optional[str] = None
    reply_to_user: Optional[str] = None
    is_deleted: bool = False
    like_count: int = 0
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    replies: List['PlazaCommentResponse'] = []
    
    class Config:
        from_attributes = True

class PlazaCommentListResponse(BaseModel):
    """评论列表响应"""
    comments: List[PlazaCommentResponse]
    total: int
    page: int
    pages: int

# 分类相关Schema
class PlazaCategoryResponse(BaseModel):
    """分类响应"""
    id: str
    name: str
    description: Optional[str] = None
    icon: Optional[str] = None
    color: Optional[str] = None
    is_active: bool = True
    sort_order: int = 0
    post_count: int = 0
    
    class Config:
        from_attributes = True

class PlazaCategoryListResponse(BaseModel):
    """分类列表响应"""
    categories: List[PlazaCategoryResponse]

# 更新递归模型引用
PlazaCommentResponse.model_rebuild()