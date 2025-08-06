from pydantic import BaseModel, Field, validator
from typing import Optional, List
from datetime import datetime
from app.models.letter import LetterStatus, Priority

# 统一响应格式的基础模式
class BaseResponse(BaseModel):
    """统一API响应格式"""
    code: int = Field(description="响应代码，0表示成功")
    msg: str = Field(description="响应消息")
    timestamp: str = Field(default_factory=lambda: datetime.utcnow().isoformat(), description="响应时间戳")

class SuccessResponse(BaseResponse):
    """成功响应格式"""
    code: int = 0
    msg: str = "success"
    data: Optional[dict] = None

class ErrorResponse(BaseResponse):
    """错误响应格式"""
    error: Optional[dict] = None

# 信件相关的Pydantic模式
class LetterCreate(BaseModel):
    """创建信件的请求模式"""
    title: str = Field(..., min_length=1, max_length=200, description="信件标题")
    content: str = Field(..., min_length=1, description="信件内容")
    receiver_hint: Optional[str] = Field(None, max_length=200, description="接收者提示信息")
    anonymous: bool = Field(False, description="是否匿名")
    priority: Priority = Field(Priority.NORMAL, description="优先级")
    delivery_instructions: Optional[str] = Field(None, description="投递说明")
    
    @validator('title')
    def title_must_not_be_empty(cls, v):
        if not v or not v.strip():
            raise ValueError('标题不能为空')
        return v.strip()
    
    @validator('content')
    def content_must_not_be_empty(cls, v):
        if not v or not v.strip():
            raise ValueError('内容不能为空')
        return v.strip()

class LetterUpdate(BaseModel):
    """更新信件的请求模式"""
    title: Optional[str] = Field(None, min_length=1, max_length=200, description="信件标题")
    content: Optional[str] = Field(None, min_length=1, description="信件内容")
    receiver_hint: Optional[str] = Field(None, max_length=200, description="接收者提示信息")
    delivery_instructions: Optional[str] = Field(None, description="投递说明")

class LetterStatusUpdate(BaseModel):
    """更新信件状态的请求模式"""
    status: LetterStatus = Field(..., description="新状态")
    location: Optional[str] = Field(None, max_length=200, description="当前位置")
    note: Optional[str] = Field(None, max_length=500, description="状态更新备注")
    
    @validator('status')
    def validate_status_transition(cls, v):
        # 这里可以添加状态转换的验证逻辑
        allowed_statuses = [status.value for status in LetterStatus]
        if v.value not in allowed_statuses:
            raise ValueError(f'无效的状态: {v}')
        return v

class LetterResponse(BaseModel):
    """信件响应模式"""
    id: str = Field(..., description="信件编号")
    title: str = Field(..., description="信件标题")
    content: str = Field(..., description="信件内容")
    sender_id: str = Field(..., description="发送者ID")
    sender_nickname: Optional[str] = Field(None, description="发送者昵称")
    receiver_hint: Optional[str] = Field(None, description="接收者提示")
    status: str = Field(..., description="信件状态")
    priority: str = Field(..., description="优先级")
    anonymous: bool = Field(..., description="是否匿名")
    delivery_instructions: Optional[str] = Field(None, description="投递说明")
    read_count: int = Field(0, description="阅读次数")
    created_at: str = Field(..., description="创建时间")
    updated_at: str = Field(..., description="更新时间")
    
    class Config:
        from_attributes = True

class LetterListResponse(BaseModel):
    """信件列表响应模式"""
    letters: List[LetterResponse] = Field(..., description="信件列表")
    total: int = Field(..., description="总数量")
    page: int = Field(..., description="当前页码")
    pages: int = Field(..., description="总页数")
    limit: int = Field(..., description="每页数量")

class LetterCreateResponse(BaseModel):
    """创建信件成功响应模式"""
    letter_id: str = Field(..., description="信件编号")
    status: str = Field(..., description="信件状态")
    created_at: str = Field(..., description="创建时间")

class LetterStatusUpdateResponse(BaseModel):
    """状态更新成功响应模式"""
    letter_id: str = Field(..., description="信件编号")
    status: str = Field(..., description="新状态")
    updated_at: str = Field(..., description="更新时间")