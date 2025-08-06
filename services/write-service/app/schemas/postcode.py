"""
OpenPenPal Postcode Pydantic Schemas
API数据验证和序列化模型
"""

from pydantic import BaseModel, validator, Field
from typing import List, Optional, Dict, Any
from datetime import datetime
from enum import Enum


class StatusEnum(str, Enum):
    active = "active"
    inactive = "inactive"


class BuildingTypeEnum(str, Enum):
    dormitory = "dormitory"
    teaching = "teaching"
    office = "office"
    other = "other"


class RoomTypeEnum(str, Enum):
    dormitory = "dormitory"
    classroom = "classroom"
    office = "office"
    other = "other"


class FeedbackTypeEnum(str, Enum):
    new_address = "new_address"
    error_report = "error_report"
    delivery_failed = "delivery_failed"


class SubmitterTypeEnum(str, Enum):
    user = "user"
    courier = "courier"


class FeedbackStatusEnum(str, Enum):
    pending = "pending"
    approved = "approved"
    rejected = "rejected"


# ==================== 基础数据模型 ====================

class SchoolBase(BaseModel):
    code: str = Field(..., min_length=2, max_length=2, description="2位学校编码")
    name: str = Field(..., min_length=1, max_length=100, description="学校名称")
    full_name: str = Field(..., min_length=1, max_length=200, description="学校全名")
    status: StatusEnum = StatusEnum.active
    managed_by: Optional[str] = Field(None, max_length=100, description="管理信使ID")

    @validator('code')
    def code_must_be_uppercase_alphanumeric(cls, v):
        if not v.isalnum() or not v.isupper():
            raise ValueError('学校编码必须为大写字母和数字')
        return v


class SchoolCreate(SchoolBase):
    pass


class SchoolUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    full_name: Optional[str] = Field(None, min_length=1, max_length=200)
    status: Optional[StatusEnum] = None
    managed_by: Optional[str] = Field(None, max_length=100)


class SchoolResponse(SchoolBase):
    id: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AreaBase(BaseModel):
    code: str = Field(..., min_length=1, max_length=1, description="1位片区编码")
    name: str = Field(..., min_length=1, max_length=100, description="片区名称")
    description: Optional[str] = Field(None, description="片区描述")
    status: StatusEnum = StatusEnum.active
    managed_by: Optional[str] = Field(None, max_length=100, description="管理信使ID")

    @validator('code')
    def code_must_be_alphanumeric(cls, v):
        if not v.isalnum():
            raise ValueError('片区编码必须为字母或数字')
        return v.upper()


class AreaCreate(AreaBase):
    pass


class AreaUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    description: Optional[str] = None
    status: Optional[StatusEnum] = None
    managed_by: Optional[str] = Field(None, max_length=100)


class AreaResponse(AreaBase):
    id: str
    school_code: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class BuildingBase(BaseModel):
    code: str = Field(..., min_length=1, max_length=1, description="1位楼栋编码")
    name: str = Field(..., min_length=1, max_length=100, description="楼栋名称")
    type: BuildingTypeEnum = BuildingTypeEnum.dormitory
    floors: Optional[int] = Field(None, ge=1, le=100, description="楼层数")
    status: StatusEnum = StatusEnum.active
    managed_by: Optional[str] = Field(None, max_length=100, description="管理信使ID")

    @validator('code')
    def code_must_be_alphanumeric(cls, v):
        if not v.isalnum():
            raise ValueError('楼栋编码必须为字母或数字')
        return v.upper()


class BuildingCreate(BuildingBase):
    pass


class BuildingUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    type: Optional[BuildingTypeEnum] = None
    floors: Optional[int] = Field(None, ge=1, le=100)
    status: Optional[StatusEnum] = None
    managed_by: Optional[str] = Field(None, max_length=100)


class BuildingResponse(BuildingBase):
    id: str
    school_code: str
    area_code: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RoomBase(BaseModel):
    code: str = Field(..., min_length=2, max_length=2, description="2位房间编码")
    name: str = Field(..., min_length=1, max_length=100, description="房间名称")
    type: RoomTypeEnum = RoomTypeEnum.dormitory
    capacity: Optional[int] = Field(None, ge=1, le=1000, description="容纳人数")
    floor: Optional[int] = Field(None, ge=1, le=100, description="楼层")
    status: StatusEnum = StatusEnum.active
    managed_by: Optional[str] = Field(None, max_length=100, description="管理信使ID")

    @validator('code')
    def code_must_be_alphanumeric(cls, v):
        if not v.replace(' ', '').isalnum():
            raise ValueError('房间编码必须为字母、数字或空格')
        return v.upper()


class RoomCreate(RoomBase):
    pass


class RoomUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    type: Optional[RoomTypeEnum] = None
    capacity: Optional[int] = Field(None, ge=1, le=1000)
    floor: Optional[int] = Field(None, ge=1, le=100)
    status: Optional[StatusEnum] = None
    managed_by: Optional[str] = Field(None, max_length=100)


class RoomResponse(RoomBase):
    id: str
    school_code: str
    area_code: str
    building_code: str
    full_postcode: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# ==================== 复合数据模型 ====================

class AddressHierarchy(BaseModel):
    """地址层级结构"""
    school: SchoolResponse
    area: Optional[AreaResponse] = None
    building: Optional[BuildingResponse] = None
    room: Optional[RoomResponse] = None


class AddressSearchResult(BaseModel):
    """地址搜索结果"""
    postcode: str = Field(..., min_length=6, max_length=6, description="6位Postcode")
    fullAddress: str = Field(..., description="完整地址描述")
    hierarchy: AddressHierarchy = Field(..., description="地址层级结构")
    matchScore: float = Field(..., ge=0.0, le=1.0, description="匹配分数")


class PostcodeStructure(BaseModel):
    """Postcode编码结构"""
    school: str = Field(..., min_length=2, max_length=2, description="学校编码")
    area: str = Field(..., min_length=1, max_length=1, description="片区编码")
    building: str = Field(..., min_length=1, max_length=1, description="楼栋编码")
    room: str = Field(..., min_length=2, max_length=2, description="房间编码")
    fullCode: str = Field(..., min_length=6, max_length=6, description="完整编码")


# ==================== 权限管理 ====================

class CourierPermissionBase(BaseModel):
    courier_id: str = Field(..., min_length=1, max_length=100, description="信使ID")
    level: int = Field(..., ge=1, le=4, description="信使等级")
    prefix_patterns: List[str] = Field(..., min_items=1, description="权限前缀列表")
    can_manage: bool = Field(False, description="是否可以管理")
    can_create: bool = Field(False, description="是否可以创建")
    can_review: bool = Field(False, description="是否可以审核")


class CourierPermissionCreate(CourierPermissionBase):
    pass


class CourierPermissionUpdate(BaseModel):
    level: Optional[int] = Field(None, ge=1, le=4)
    prefix_patterns: Optional[List[str]] = Field(None, min_items=1)
    can_manage: Optional[bool] = None
    can_create: Optional[bool] = None
    can_review: Optional[bool] = None


class CourierPermissionResponse(CourierPermissionBase):
    id: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# ==================== 反馈管理 ====================

class FeedbackBase(BaseModel):
    type: FeedbackTypeEnum = Field(..., description="反馈类型")
    postcode: Optional[str] = Field(None, min_length=6, max_length=6, description="相关Postcode")
    description: str = Field(..., min_length=1, description="问题描述")
    suggested_school_code: Optional[str] = Field(None, min_length=2, max_length=2)
    suggested_area_code: Optional[str] = Field(None, min_length=1, max_length=1)
    suggested_building_code: Optional[str] = Field(None, min_length=1, max_length=1)
    suggested_room_code: Optional[str] = Field(None, min_length=2, max_length=2)
    suggested_name: Optional[str] = Field(None, min_length=1, max_length=200)
    submitter_type: SubmitterTypeEnum = SubmitterTypeEnum.user


class FeedbackCreate(FeedbackBase):
    pass


class FeedbackUpdate(BaseModel):
    description: Optional[str] = Field(None, min_length=1)
    status: Optional[FeedbackStatusEnum] = None


class FeedbackReview(BaseModel):
    action: str = Field(..., pattern="^(approve|reject)$", description="审核动作")
    notes: Optional[str] = Field(None, description="审核备注")


class FeedbackResponse(FeedbackBase):
    id: str
    submitted_by: str
    status: FeedbackStatusEnum
    reviewed_by: Optional[str] = None
    review_notes: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# ==================== 统计分析 ====================

class PostcodeStatsBase(BaseModel):
    postcode: str = Field(..., min_length=6, max_length=6, description="Postcode编码")
    delivery_count: int = Field(0, ge=0, description="投递次数")
    error_count: int = Field(0, ge=0, description="错误次数")
    last_used: datetime = Field(..., description="最后使用时间")
    popularity_score: float = Field(0.0, ge=0.0, le=100.0, description="热门度分数")


class PostcodeStatsCreate(PostcodeStatsBase):
    pass


class PostcodeStatsUpdate(BaseModel):
    delivery_count: Optional[int] = Field(None, ge=0)
    error_count: Optional[int] = Field(None, ge=0)
    popularity_score: Optional[float] = Field(None, ge=0.0, le=100.0)


class PostcodeStatsResponse(PostcodeStatsBase):
    id: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# ==================== 工具接口 ====================

class PostcodeValidation(BaseModel):
    """Postcode验证结果"""
    code: str = Field(..., description="验证的编码")
    is_valid: bool = Field(..., description="是否有效")
    exists: bool = Field(..., description="是否存在")
    errors: List[str] = Field(default_factory=list, description="错误信息列表")


class BatchImportResult(BaseModel):
    """批量导入结果"""
    imported: int = Field(..., ge=0, description="成功导入数量")
    failed: int = Field(..., ge=0, description="失败数量")
    errors: List[Dict[str, Any]] = Field(default_factory=list, description="错误详情")


class BatchValidationResult(BaseModel):
    """批量验证结果"""
    total: int = Field(..., ge=0, description="总数量")
    valid: int = Field(..., ge=0, description="有效数量")
    invalid: int = Field(..., ge=0, description="无效数量")
    results: List[PostcodeValidation] = Field(default_factory=list, description="详细结果")