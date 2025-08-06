"""
批量操作相关的Pydantic模式
"""
from pydantic import BaseModel, Field, validator
from typing import List, Dict, Any, Optional, Union
from datetime import datetime
from enum import Enum


class BatchOperationEnum(str, Enum):
    """批量操作类型枚举"""
    DELETE = "delete"
    UPDATE = "update"
    STATUS_UPDATE = "status_update"
    EXPORT = "export"
    ARCHIVE = "archive"
    RESTORE = "restore"
    BULK_CREATE = "bulk_create"


class BatchTargetEnum(str, Enum):
    """批量操作目标类型枚举"""
    LETTERS = "letters"
    PRODUCTS = "products"
    ORDERS = "orders"
    USERS = "users"
    DRAFTS = "drafts"
    PLAZA_POSTS = "plaza_posts"
    MUSEUM_ITEMS = "museum_items"


class BatchOperationRequest(BaseModel):
    """批量操作请求基础模型"""
    operation: BatchOperationEnum = Field(description="操作类型")
    target_type: BatchTargetEnum = Field(description="目标资源类型")
    target_ids: List[str] = Field(description="目标资源ID列表", min_items=1, max_items=1000)
    operation_data: Optional[Dict[str, Any]] = Field(None, description="操作相关数据")
    dry_run: bool = Field(default=False, description="是否为试运行（不执行实际操作）")
    
    @validator('target_ids')
    def validate_target_ids(cls, v):
        if len(v) > 1000:
            raise ValueError("单次批量操作最多支持1000个目标")
        if len(set(v)) != len(v):
            raise ValueError("目标ID列表中包含重复项")
        return v


class BatchLetterOperationRequest(BatchOperationRequest):
    """信件批量操作请求"""
    target_type: BatchTargetEnum = Field(default=BatchTargetEnum.LETTERS, description="固定为letters")


class BatchProductOperationRequest(BatchOperationRequest):
    """商品批量操作请求"""
    target_type: BatchTargetEnum = Field(default=BatchTargetEnum.PRODUCTS, description="固定为products")


class BatchOrderOperationRequest(BatchOperationRequest):
    """订单批量操作请求"""
    target_type: BatchTargetEnum = Field(default=BatchTargetEnum.ORDERS, description="固定为orders")


class BatchOperationResult(BaseModel):
    """单个操作结果"""
    target_id: str = Field(description="目标资源ID")
    success: bool = Field(description="操作是否成功")
    message: str = Field(description="操作结果消息")
    error_code: Optional[str] = Field(None, description="错误码（如有）")
    data: Optional[Dict[str, Any]] = Field(None, description="操作结果数据")


class BatchOperationResponse(BaseModel):
    """批量操作响应模型"""
    operation_id: str = Field(description="批量操作ID")
    operation: BatchOperationEnum = Field(description="操作类型")
    target_type: BatchTargetEnum = Field(description="目标资源类型")
    total_count: int = Field(description="总操作数量")
    success_count: int = Field(description="成功操作数量")
    failure_count: int = Field(description="失败操作数量")
    results: List[BatchOperationResult] = Field(description="详细操作结果列表")
    started_at: datetime = Field(description="操作开始时间")
    completed_at: Optional[datetime] = Field(None, description="操作完成时间")
    duration_ms: Optional[int] = Field(None, description="操作耗时（毫秒）")
    dry_run: bool = Field(description="是否为试运行")


class BatchStatusUpdateRequest(BaseModel):
    """批量状态更新请求"""
    target_ids: List[str] = Field(description="目标ID列表")
    new_status: str = Field(description="新状态")
    reason: Optional[str] = Field(None, description="状态变更原因")
    force: bool = Field(default=False, description="是否强制更新（跳过状态转换验证）")


class BatchDeleteRequest(BaseModel):
    """批量删除请求"""
    target_ids: List[str] = Field(description="目标ID列表")
    soft_delete: bool = Field(default=True, description="是否软删除")
    delete_reason: Optional[str] = Field(None, description="删除原因")


class BatchCreateRequest(BaseModel):
    """批量创建请求"""
    items: List[Dict[str, Any]] = Field(description="要创建的项目数据列表", min_items=1, max_items=100)
    skip_validation: bool = Field(default=False, description="是否跳过数据验证")
    continue_on_error: bool = Field(default=True, description="遇到错误时是否继续处理其他项目")


class BatchExportRequest(BaseModel):
    """批量导出请求"""
    target_ids: Optional[List[str]] = Field(None, description="指定导出的ID列表（为空时导出所有）")
    export_format: str = Field(default="json", description="导出格式", pattern="^(json|csv|excel)$")
    include_fields: Optional[List[str]] = Field(None, description="包含的字段列表")
    exclude_fields: Optional[List[str]] = Field(None, description="排除的字段列表")
    filters: Optional[Dict[str, Any]] = Field(None, description="导出过滤条件")


class BatchArchiveRequest(BaseModel):
    """批量归档请求"""
    target_ids: List[str] = Field(description="目标ID列表")
    archive_reason: Optional[str] = Field(None, description="归档原因")
    archive_location: Optional[str] = Field(None, description="归档位置")


class BatchJobStatus(BaseModel):
    """批量作业状态"""
    job_id: str = Field(description="作业ID")
    status: str = Field(description="作业状态", pattern="^(pending|running|completed|failed|cancelled)$")
    progress: float = Field(description="进度百分比", ge=0, le=100)
    current_item: Optional[str] = Field(None, description="当前处理项目")
    estimated_remaining: Optional[int] = Field(None, description="预计剩余时间（秒）")
    created_at: datetime = Field(description="作业创建时间")
    updated_at: datetime = Field(description="最后更新时间")
    error_message: Optional[str] = Field(None, description="错误信息")


class BatchLetterUpdate(BaseModel):
    """批量信件更新数据"""
    title: Optional[str] = Field(None, description="标题")
    content: Optional[str] = Field(None, description="内容")
    receiver_hint: Optional[str] = Field(None, description="收件人提示")
    priority: Optional[str] = Field(None, description="优先级")
    delivery_instructions: Optional[str] = Field(None, description="投递说明")


class BatchProductUpdate(BaseModel):
    """批量商品更新数据"""
    name: Optional[str] = Field(None, description="商品名称")
    description: Optional[str] = Field(None, description="商品描述")
    price: Optional[float] = Field(None, description="价格")
    stock: Optional[int] = Field(None, description="库存")
    category_id: Optional[str] = Field(None, description="分类ID")
    status: Optional[str] = Field(None, description="状态")


class BatchDraftUpdate(BaseModel):
    """批量草稿更新数据"""
    title: Optional[str] = Field(None, description="标题")
    content: Optional[str] = Field(None, description="内容")
    is_active: Optional[bool] = Field(None, description="是否激活")


# 便捷类型别名
BatchRequest = BatchOperationRequest
BatchResponse = BatchOperationResponse
BatchResult = BatchOperationResult
BatchJob = BatchJobStatus


class SuccessResponse(BaseModel):
    """统一成功响应格式"""
    code: int = Field(default=0, description="响应码，0表示成功")
    msg: str = Field(default="success", description="响应消息")
    data: Any = Field(description="响应数据")
    timestamp: datetime = Field(default_factory=datetime.utcnow, description="响应时间")