from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from enum import Enum


class DraftType(str, Enum):
    """草稿类型枚举"""
    letter = "letter"
    reply = "reply"


class RecipientType(str, Enum):
    """收件人类型枚举"""
    friend = "friend"
    stranger = "stranger"
    group = "group"


class ChangeType(str, Enum):
    """变更类型枚举"""
    auto_save = "auto_save"
    manual_save = "manual_save"
    version_backup = "version_backup"


# 草稿基础模型
class DraftBase(BaseModel):
    title: Optional[str] = Field(None, max_length=200, description="草稿标题")
    content: Optional[str] = Field(None, description="草稿内容")
    recipient_id: Optional[str] = Field(None, max_length=20, description="收件人ID")
    recipient_type: Optional[RecipientType] = Field(None, description="收件人类型")
    paper_style: str = Field("classic", max_length=50, description="信纸样式")
    envelope_style: str = Field("simple", max_length=50, description="信封样式")
    draft_type: DraftType = Field(DraftType.letter, description="草稿类型")
    parent_letter_id: Optional[str] = Field(None, max_length=20, description="父信件ID（回复时）")
    auto_save_enabled: bool = Field(True, description="是否启用自动保存")


# 创建草稿请求
class DraftCreate(DraftBase):
    pass


# 更新草稿请求
class DraftUpdate(BaseModel):
    title: Optional[str] = Field(None, max_length=200)
    content: Optional[str] = Field(None)
    recipient_id: Optional[str] = Field(None, max_length=20)
    recipient_type: Optional[RecipientType] = None
    paper_style: Optional[str] = Field(None, max_length=50)
    envelope_style: Optional[str] = Field(None, max_length=50)
    auto_save_enabled: Optional[bool] = None


# 草稿响应模型
class DraftResponse(DraftBase):
    id: str
    user_id: str
    version: int
    word_count: int
    character_count: int
    last_edit_time: datetime
    created_at: datetime
    updated_at: datetime
    is_active: bool
    is_discarded: bool

    class Config:
        from_attributes = True


# 草稿列表项
class DraftListItem(BaseModel):
    id: str
    title: Optional[str]
    draft_type: DraftType
    recipient_id: Optional[str]
    version: int
    word_count: int
    last_edit_time: datetime
    created_at: datetime
    is_active: bool

    class Config:
        from_attributes = True


# 草稿历史记录
class DraftHistoryBase(BaseModel):
    title: Optional[str]
    content: Optional[str]
    change_summary: Optional[str]
    change_type: ChangeType = ChangeType.auto_save


class DraftHistoryResponse(DraftHistoryBase):
    id: str
    draft_id: str
    user_id: str
    version: int
    word_count: int
    character_count: int
    created_at: datetime

    class Config:
        from_attributes = True


# 自动保存请求
class AutoSaveRequest(BaseModel):
    content: Optional[str] = Field(None, description="当前内容")
    title: Optional[str] = Field(None, max_length=200, description="当前标题")
    cursor_position: Optional[int] = Field(None, description="光标位置")
    selection_start: Optional[int] = Field(None, description="选择开始位置")
    selection_end: Optional[int] = Field(None, description="选择结束位置")


# 草稿统计
class DraftStats(BaseModel):
    total_drafts: int
    active_drafts: int
    discarded_drafts: int
    total_words: int
    total_characters: int
    oldest_draft: Optional[datetime]
    newest_draft: Optional[datetime]


# 批量操作请求
class DraftBatchOperation(BaseModel):
    draft_ids: List[str] = Field(..., description="草稿ID列表")
    operation: str = Field(..., description="操作类型: delete/discard/activate")


# 草稿搜索请求
class DraftSearchRequest(BaseModel):
    keyword: Optional[str] = Field(None, description="搜索关键词")
    draft_type: Optional[DraftType] = Field(None, description="草稿类型")
    recipient_type: Optional[RecipientType] = Field(None, description="收件人类型")
    date_from: Optional[datetime] = Field(None, description="开始日期")
    date_to: Optional[datetime] = Field(None, description="结束日期")
    is_active: Optional[bool] = Field(None, description="是否活跃")
    limit: int = Field(20, ge=1, le=100, description="返回数量限制")
    offset: int = Field(0, ge=0, description="偏移量")


# 标准响应模型
class SuccessResponse(BaseModel):
    code: int = 0
    msg: str = "操作成功"
    data: Optional[dict] = None