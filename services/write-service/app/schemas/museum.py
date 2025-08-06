from pydantic import BaseModel, Field, validator
from typing import Optional, List
from datetime import datetime
from enum import Enum

class MuseumEra(str, Enum):
    """历史时期枚举"""
    ANCIENT = "ancient"
    MODERN = "modern"
    CONTEMPORARY = "contemporary"
    PRESENT = "present"
    DIGITAL = "digital"

class MuseumLetterStatus(str, Enum):
    """博物馆信件状态枚举"""
    PENDING = "pending"
    APPROVED = "approved"
    FEATURED = "featured"
    ARCHIVED = "archived"
    REJECTED = "rejected"

class SourceType(str, Enum):
    """来源类型枚举"""
    ORIGINAL = "original"      # 原始历史信件
    CONTRIBUTED = "contributed"  # 用户贡献
    DIGITIZED = "digitized"    # 数字化扫描

class EventType(str, Enum):
    """事件类型枚举"""
    LETTER = "letter"
    HISTORICAL = "historical"
    CULTURAL = "cultural"
    PERSONAL = "personal"

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

# 博物馆信件相关Schema
class MuseumLetterCreate(BaseModel):
    """创建博物馆信件请求"""
    title: str = Field(..., min_length=1, max_length=200, description="信件标题")
    content: str = Field(..., min_length=1, description="信件内容")
    summary: Optional[str] = Field(None, max_length=500, description="信件摘要")
    original_author: Optional[str] = Field(None, max_length=100, description="原作者")
    original_recipient: Optional[str] = Field(None, max_length=100, description="原收件人")
    historical_date: Optional[datetime] = Field(None, description="历史日期")
    era: MuseumEra = Field(default=MuseumEra.PRESENT, description="历史时期")
    location: Optional[str] = Field(None, max_length=200, description="地理位置")
    category: str = Field(..., max_length=50, description="信件分类")
    tags: Optional[List[str]] = Field(default=[], description="标签列表")
    language: str = Field(default="zh", description="语言")
    source_type: SourceType = Field(default=SourceType.CONTRIBUTED, description="来源类型")
    source_description: Optional[str] = Field(None, description="来源描述")
    letter_id: Optional[str] = Field(None, description="关联的现代信件ID")
    
    @validator('tags')
    def validate_tags(cls, v):
        if v is None:
            return []
        if len(v) > 15:
            raise ValueError('标签数量不能超过15个')
        for tag in v:
            if len(tag) > 30:
                raise ValueError('单个标签长度不能超过30个字符')
        return v
    
    @validator('content')
    def validate_content(cls, v):
        if len(v) > 100000:
            raise ValueError('内容长度不能超过100000个字符')
        return v

class MuseumLetterUpdate(BaseModel):
    """更新博物馆信件请求"""
    title: Optional[str] = Field(None, min_length=1, max_length=200, description="信件标题")
    content: Optional[str] = Field(None, min_length=1, description="信件内容")
    summary: Optional[str] = Field(None, max_length=500, description="信件摘要")
    original_author: Optional[str] = Field(None, max_length=100, description="原作者")
    original_recipient: Optional[str] = Field(None, max_length=100, description="原收件人")
    historical_date: Optional[datetime] = Field(None, description="历史日期")
    era: Optional[MuseumEra] = Field(None, description="历史时期")
    location: Optional[str] = Field(None, max_length=200, description="地理位置")
    category: Optional[str] = Field(None, max_length=50, description="信件分类")
    tags: Optional[List[str]] = Field(None, description="标签列表")
    source_description: Optional[str] = Field(None, description="来源描述")

class MuseumLetterResponse(BaseModel):
    """博物馆信件响应"""
    id: str
    title: str
    content: Optional[str] = None
    summary: Optional[str] = None
    original_author: Optional[str] = None
    original_recipient: Optional[str] = None
    historical_date: Optional[datetime] = None
    era: str
    location: Optional[str] = None
    category: str
    tags: List[str] = []
    language: str
    source_type: str
    source_description: Optional[str] = None
    contributor_name: Optional[str] = None
    status: str
    is_featured: bool = False
    view_count: int = 0
    favorite_count: int = 0
    share_count: int = 0
    rating_avg: float = 0.0
    rating_count: int = 0
    letter_id: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class MuseumLetterListItem(BaseModel):
    """博物馆信件列表项（不包含完整内容）"""
    id: str
    title: str
    summary: Optional[str] = None
    original_author: Optional[str] = None
    historical_date: Optional[datetime] = None
    era: str
    location: Optional[str] = None
    category: str
    tags: List[str] = []
    contributor_name: Optional[str] = None
    status: str
    is_featured: bool = False
    view_count: int = 0
    favorite_count: int = 0
    rating_avg: float = 0.0
    created_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class MuseumLetterListResponse(BaseModel):
    """博物馆信件列表响应"""
    letters: List[MuseumLetterListItem]
    total: int
    page: int
    pages: int
    has_next: bool
    has_prev: bool

# 收藏相关Schema
class MuseumFavoriteCreate(BaseModel):
    """添加收藏请求"""
    museum_letter_id: str = Field(..., description="博物馆信件ID")
    note: Optional[str] = Field(None, max_length=200, description="收藏备注")

class MuseumFavoriteResponse(BaseModel):
    """收藏响应"""
    museum_letter_id: str
    favorited: bool
    favorite_count: int
    note: Optional[str] = None

# 评分相关Schema
class MuseumRatingCreate(BaseModel):
    """评分请求"""
    rating: int = Field(..., ge=1, le=5, description="评分(1-5)")
    comment: Optional[str] = Field(None, max_length=500, description="评价评论")

class MuseumRatingResponse(BaseModel):
    """评分响应"""
    museum_letter_id: str
    user_id: str
    rating: int
    comment: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

# 时间线相关Schema
class TimelineEventCreate(BaseModel):
    """创建时间线事件请求"""
    title: str = Field(..., min_length=1, max_length=200, description="事件标题")
    description: Optional[str] = Field(None, description="事件描述")
    event_date: datetime = Field(..., description="事件日期")
    era: MuseumEra = Field(..., description="历史时期")
    location: Optional[str] = Field(None, max_length=200, description="事件地点")
    event_type: EventType = Field(..., description="事件类型")
    category: Optional[str] = Field(None, max_length=50, description="事件分类")
    importance: int = Field(default=1, ge=1, le=5, description="重要程度(1-5)")
    museum_letter_id: Optional[str] = Field(None, description="关联博物馆信件ID")
    image_url: Optional[str] = Field(None, description="事件图片URL")
    audio_url: Optional[str] = Field(None, description="事件音频URL")
    video_url: Optional[str] = Field(None, description="事件视频URL")

class TimelineEventResponse(BaseModel):
    """时间线事件响应"""
    id: str
    title: str
    description: Optional[str] = None
    event_date: datetime
    era: str
    location: Optional[str] = None
    event_type: str
    category: Optional[str] = None
    importance: int
    museum_letter_id: Optional[str] = None
    is_featured: bool = False
    image_url: Optional[str] = None
    audio_url: Optional[str] = None
    video_url: Optional[str] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class TimelineResponse(BaseModel):
    """时间线响应"""
    events: List[TimelineEventResponse]
    total: int
    date_range: dict  # {"start_date": "xxx", "end_date": "xxx"}
    eras: List[str]

# 收藏集相关Schema
class MuseumCollectionCreate(BaseModel):
    """创建收藏集请求"""
    name: str = Field(..., min_length=1, max_length=100, description="收藏集名称")
    description: Optional[str] = Field(None, description="收藏集描述")
    theme: Optional[str] = Field(None, max_length=50, description="主题")
    is_public: bool = Field(default=True, description="是否公开")

class MuseumCollectionResponse(BaseModel):
    """收藏集响应"""
    id: str
    name: str
    description: Optional[str] = None
    theme: Optional[str] = None
    creator_id: str
    creator_name: Optional[str] = None
    is_public: bool
    is_featured: bool = False
    letter_count: int = 0
    view_count: int = 0
    follow_count: int = 0
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    
    class Config:
        from_attributes = True

class MuseumCollectionListResponse(BaseModel):
    """收藏集列表响应"""
    collections: List[MuseumCollectionResponse]
    total: int
    page: int
    pages: int

# 统计相关Schema
class MuseumStatsResponse(BaseModel):
    """博物馆统计响应"""
    total_letters: int
    total_collections: int
    total_timeline_events: int
    era_distribution: dict  # {"ancient": 10, "modern": 20, ...}
    category_distribution: dict
    popular_tags: List[str]
    featured_letters: List[MuseumLetterListItem]
    recent_contributions: List[MuseumLetterListItem]

# 搜索相关Schema
class MuseumSearchRequest(BaseModel):
    """博物馆搜索请求"""
    keyword: Optional[str] = Field(None, description="关键词")
    era: Optional[MuseumEra] = Field(None, description="历史时期过滤")
    category: Optional[str] = Field(None, description="分类过滤")
    tags: Optional[List[str]] = Field(None, description="标签过滤")
    author: Optional[str] = Field(None, description="作者过滤")
    location: Optional[str] = Field(None, description="地点过滤")
    date_from: Optional[datetime] = Field(None, description="开始日期")
    date_to: Optional[datetime] = Field(None, description="结束日期")
    sort_by: str = Field(default="created_at", description="排序字段")
    order: str = Field(default="desc", description="排序方向")
    page: int = Field(default=1, ge=1, description="页码")
    limit: int = Field(default=20, ge=1, le=100, description="每页数量")

class MuseumSearchResponse(BaseModel):
    """博物馆搜索响应"""
    letters: List[MuseumLetterListItem]
    total: int
    page: int
    pages: int
    search_time: float  # 搜索耗时(秒)
    filters_applied: dict  # 应用的过滤条件摘要