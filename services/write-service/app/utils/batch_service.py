"""
批量操作服务模块
"""
import uuid
import logging
from typing import List, Dict, Any, Optional, Tuple, Type
from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_
import asyncio
import json
from concurrent.futures import ThreadPoolExecutor

from app.models.letter import Letter, LetterStatus
from app.models.shop import Product, ProductStatus
from app.models.plaza import PlazaPost, PostStatus
from app.models.museum import MuseumLetter, MuseumLetterStatus
from app.models.draft import LetterDraft
from app.schemas.batch import (
    BatchOperationEnum, BatchTargetEnum, BatchOperationRequest,
    BatchOperationResult, BatchOperationResponse, BatchJobStatus
)
from app.utils.cache_manager import get_cache_manager
from app.utils.websocket_client import notify_batch_operation_progress

logger = logging.getLogger(__name__)


class BatchOperationService:
    """批量操作服务类"""
    
    def __init__(self):
        self.cache = get_cache_manager()
        self.active_jobs: Dict[str, BatchJobStatus] = {}
        self.executor = ThreadPoolExecutor(max_workers=4)
        
        # 目标模型映射
        self.target_models = {
            BatchTargetEnum.LETTERS: Letter,
            BatchTargetEnum.PRODUCTS: Product,
            BatchTargetEnum.PLAZA_POSTS: PlazaPost,
            BatchTargetEnum.MUSEUM_ITEMS: MuseumLetter,
            BatchTargetEnum.DRAFTS: LetterDraft
        }
    
    def get_target_model(self, target_type: BatchTargetEnum) -> Type:
        """获取目标模型类"""
        return self.target_models.get(target_type)
    
    async def execute_batch_operation(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> BatchOperationResponse:
        """
        执行批量操作
        
        Args:
            request: 批量操作请求
            db: 数据库会话
            current_user: 当前用户ID
            
        Returns:
            BatchOperationResponse: 批量操作响应
        """
        # 生成操作ID
        operation_id = str(uuid.uuid4())
        started_at = datetime.utcnow()
        
        logger.info(f"Starting batch operation {operation_id}: {request.operation} on {len(request.target_ids)} {request.target_type}")
        
        # 创建作业状态
        job_status = BatchJobStatus(
            job_id=operation_id,
            status="running",
            progress=0.0,
            created_at=started_at,
            updated_at=started_at
        )
        self.active_jobs[operation_id] = job_status
        
        try:
            # 根据操作类型分发
            if request.operation == BatchOperationEnum.DELETE:
                results = await self._batch_delete(request, db, current_user)
            elif request.operation == BatchOperationEnum.UPDATE:
                results = await self._batch_update(request, db, current_user)
            elif request.operation == BatchOperationEnum.STATUS_UPDATE:
                results = await self._batch_status_update(request, db, current_user)
            elif request.operation == BatchOperationEnum.EXPORT:
                results = await self._batch_export(request, db, current_user)
            elif request.operation == BatchOperationEnum.ARCHIVE:
                results = await self._batch_archive(request, db, current_user)
            elif request.operation == BatchOperationEnum.RESTORE:
                results = await self._batch_restore(request, db, current_user)
            elif request.operation == BatchOperationEnum.BULK_CREATE:
                results = await self._batch_create(request, db, current_user)
            else:
                raise ValueError(f"Unsupported operation: {request.operation}")
            
            # 计算统计信息
            success_count = sum(1 for r in results if r.success)
            failure_count = len(results) - success_count
            
            completed_at = datetime.utcnow()
            duration_ms = int((completed_at - started_at).total_seconds() * 1000)
            
            # 更新作业状态
            job_status.status = "completed"
            job_status.progress = 100.0
            job_status.updated_at = completed_at
            
            response = BatchOperationResponse(
                operation_id=operation_id,
                operation=request.operation,
                target_type=request.target_type,
                total_count=len(request.target_ids),
                success_count=success_count,
                failure_count=failure_count,
                results=results,
                started_at=started_at,
                completed_at=completed_at,
                duration_ms=duration_ms,
                dry_run=request.dry_run
            )
            
            logger.info(f"Batch operation {operation_id} completed: {success_count}/{len(results)} successful")
            
            # 发送WebSocket通知
            await notify_batch_operation_progress(
                current_user,
                operation_id,
                {
                    "status": "completed",
                    "progress": 100,
                    "success_count": success_count,
                    "failure_count": failure_count
                }
            )
            
            return response
            
        except Exception as e:
            logger.error(f"Batch operation {operation_id} failed: {e}")
            
            # 更新作业状态为失败
            job_status.status = "failed"
            job_status.error_message = str(e)
            job_status.updated_at = datetime.utcnow()
            
            # 清理作业状态
            if operation_id in self.active_jobs:
                del self.active_jobs[operation_id]
            
            raise
    
    async def _batch_delete(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> List[BatchOperationResult]:
        """批量删除操作"""
        results = []
        model_class = self.get_target_model(request.target_type)
        
        if not model_class:
            raise ValueError(f"Unsupported target type: {request.target_type}")
        
        soft_delete = request.operation_data.get('soft_delete', True) if request.operation_data else True
        
        for i, target_id in enumerate(request.target_ids):
            try:
                # 查找目标资源
                target = db.query(model_class).filter(model_class.id == target_id).first()
                
                if not target:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=False,
                        message=f"{request.target_type} not found"
                    ))
                    continue
                
                # 检查权限（对于需要权限验证的资源）
                if hasattr(target, 'sender_id') or hasattr(target, 'user_id'):
                    owner_id = getattr(target, 'sender_id', None) or getattr(target, 'user_id', None)
                    if owner_id != current_user:
                        results.append(BatchOperationResult(
                            target_id=target_id,
                            success=False,
                            message="Permission denied",
                            error_code="PERMISSION_DENIED"
                        ))
                        continue
                
                if not request.dry_run:
                    if soft_delete and hasattr(target, 'is_deleted'):
                        # 软删除
                        target.is_deleted = True
                        target.deleted_at = datetime.utcnow()
                        if hasattr(target, 'updated_at'):
                            target.updated_at = datetime.utcnow()
                    else:
                        # 硬删除
                        db.delete(target)
                    
                    db.commit()
                
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=True,
                    message="Deleted successfully"
                ))
                
                # 更新进度
                progress = (i + 1) / len(request.target_ids) * 100
                await self._update_progress(request.operation, progress, current_user)
                
            except Exception as e:
                logger.error(f"Failed to delete {target_id}: {e}")
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=False,
                    message=str(e),
                    error_code="DELETE_FAILED"
                ))
                db.rollback()
        
        return results
    
    async def _batch_update(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> List[BatchOperationResult]:
        """批量更新操作"""
        results = []
        model_class = self.get_target_model(request.target_type)
        
        if not model_class:
            raise ValueError(f"Unsupported target type: {request.target_type}")
        
        update_data = request.operation_data or {}
        
        for i, target_id in enumerate(request.target_ids):
            try:
                # 查找目标资源
                target = db.query(model_class).filter(model_class.id == target_id).first()
                
                if not target:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=False,
                        message=f"{request.target_type} not found"
                    ))
                    continue
                
                # 检查权限
                if hasattr(target, 'sender_id') or hasattr(target, 'user_id'):
                    owner_id = getattr(target, 'sender_id', None) or getattr(target, 'user_id', None)
                    if owner_id != current_user:
                        results.append(BatchOperationResult(
                            target_id=target_id,
                            success=False,
                            message="Permission denied",
                            error_code="PERMISSION_DENIED"
                        ))
                        continue
                
                if not request.dry_run:
                    # 更新字段
                    updated_fields = []
                    for field, value in update_data.items():
                        if hasattr(target, field):
                            setattr(target, field, value)
                            updated_fields.append(field)
                    
                    # 更新时间戳
                    if hasattr(target, 'updated_at'):
                        target.updated_at = datetime.utcnow()
                    
                    db.commit()
                    
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=True,
                        message="Updated successfully",
                        data={"updated_fields": updated_fields}
                    ))
                else:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=True,
                        message="Update validated (dry run)"
                    ))
                
                # 更新进度
                progress = (i + 1) / len(request.target_ids) * 100
                await self._update_progress(request.operation, progress, current_user)
                
            except Exception as e:
                logger.error(f"Failed to update {target_id}: {e}")
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=False,
                    message=str(e),
                    error_code="UPDATE_FAILED"
                ))
                db.rollback()
        
        return results
    
    async def _batch_status_update(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> List[BatchOperationResult]:
        """批量状态更新操作"""
        results = []
        model_class = self.get_target_model(request.target_type)
        
        if not model_class:
            raise ValueError(f"Unsupported target type: {request.target_type}")
        
        new_status = request.operation_data.get('new_status')
        force = request.operation_data.get('force', False)
        
        if not new_status:
            raise ValueError("new_status is required for status update operation")
        
        for i, target_id in enumerate(request.target_ids):
            try:
                # 查找目标资源
                target = db.query(model_class).filter(model_class.id == target_id).first()
                
                if not target:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=False,
                        message=f"{request.target_type} not found"
                    ))
                    continue
                
                # 检查权限
                if hasattr(target, 'sender_id') or hasattr(target, 'user_id'):
                    owner_id = getattr(target, 'sender_id', None) or getattr(target, 'user_id', None)
                    if owner_id != current_user:
                        results.append(BatchOperationResult(
                            target_id=target_id,
                            success=False,
                            message="Permission denied",
                            error_code="PERMISSION_DENIED"
                        ))
                        continue
                
                if not request.dry_run:
                    # 状态验证（如果不是强制更新）
                    if not force and hasattr(target, 'status'):
                        # 这里可以添加状态转换验证逻辑
                        pass
                    
                    # 更新状态
                    old_status = getattr(target, 'status', None)
                    target.status = new_status
                    
                    if hasattr(target, 'updated_at'):
                        target.updated_at = datetime.utcnow()
                    
                    db.commit()
                    
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=True,
                        message="Status updated successfully",
                        data={
                            "old_status": old_status,
                            "new_status": new_status
                        }
                    ))
                else:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=True,
                        message="Status update validated (dry run)"
                    ))
                
                # 更新进度
                progress = (i + 1) / len(request.target_ids) * 100
                await self._update_progress(request.operation, progress, current_user)
                
            except Exception as e:
                logger.error(f"Failed to update status for {target_id}: {e}")
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=False,
                    message=str(e),
                    error_code="STATUS_UPDATE_FAILED"
                ))
                db.rollback()
        
        return results
    
    async def _batch_export(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> List[BatchOperationResult]:
        """批量导出操作"""
        results = []
        model_class = self.get_target_model(request.target_type)
        
        if not model_class:
            raise ValueError(f"Unsupported target type: {request.target_type}")
        
        export_format = request.operation_data.get('export_format', 'json')
        include_fields = request.operation_data.get('include_fields')
        exclude_fields = request.operation_data.get('exclude_fields', [])
        
        exported_data = []
        
        for i, target_id in enumerate(request.target_ids):
            try:
                # 查找目标资源
                target = db.query(model_class).filter(model_class.id == target_id).first()
                
                if not target:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=False,
                        message=f"{request.target_type} not found"
                    ))
                    continue
                
                # 检查权限
                if hasattr(target, 'sender_id') or hasattr(target, 'user_id'):
                    owner_id = getattr(target, 'sender_id', None) or getattr(target, 'user_id', None)
                    if owner_id != current_user:
                        results.append(BatchOperationResult(
                            target_id=target_id,
                            success=False,
                            message="Permission denied",
                            error_code="PERMISSION_DENIED"
                        ))
                        continue
                
                # 构建导出数据
                if hasattr(target, 'to_dict'):
                    item_data = target.to_dict()
                else:
                    # 基础的字典转换
                    item_data = {
                        column.name: getattr(target, column.name)
                        for column in target.__table__.columns
                    }
                
                # 过滤字段
                if include_fields:
                    item_data = {k: v for k, v in item_data.items() if k in include_fields}
                if exclude_fields:
                    item_data = {k: v for k, v in item_data.items() if k not in exclude_fields}
                
                exported_data.append(item_data)
                
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=True,
                    message="Exported successfully"
                ))
                
                # 更新进度
                progress = (i + 1) / len(request.target_ids) * 100
                await self._update_progress(request.operation, progress, current_user)
                
            except Exception as e:
                logger.error(f"Failed to export {target_id}: {e}")
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=False,
                    message=str(e),
                    error_code="EXPORT_FAILED"
                ))
        
        # 将导出数据存储到缓存或文件系统
        export_id = str(uuid.uuid4())
        self.cache.set(f"export:{export_id}", {
            "format": export_format,
            "data": exported_data,
            "created_at": datetime.utcnow().isoformat(),
            "created_by": current_user
        }, ttl=3600)  # 1小时过期
        
        # 在结果中添加导出信息
        for result in results:
            if result.success:
                result.data = {"export_id": export_id, "export_url": f"/api/batch/export/{export_id}"}
        
        return results
    
    async def _batch_archive(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> List[BatchOperationResult]:
        """批量归档操作"""
        results = []
        model_class = self.get_target_model(request.target_type)
        
        if not model_class:
            raise ValueError(f"Unsupported target type: {request.target_type}")
        
        for i, target_id in enumerate(request.target_ids):
            try:
                # 查找目标资源
                target = db.query(model_class).filter(model_class.id == target_id).first()
                
                if not target:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=False,
                        message=f"{request.target_type} not found"
                    ))
                    continue
                
                # 检查权限
                if hasattr(target, 'sender_id') or hasattr(target, 'user_id'):
                    owner_id = getattr(target, 'sender_id', None) or getattr(target, 'user_id', None)
                    if owner_id != current_user:
                        results.append(BatchOperationResult(
                            target_id=target_id,
                            success=False,
                            message="Permission denied",
                            error_code="PERMISSION_DENIED"
                        ))
                        continue
                
                if not request.dry_run:
                    # 归档处理
                    if hasattr(target, 'is_archived'):
                        target.is_archived = True
                    if hasattr(target, 'archived_at'):
                        target.archived_at = datetime.utcnow()
                    if hasattr(target, 'updated_at'):
                        target.updated_at = datetime.utcnow()
                    
                    db.commit()
                
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=True,
                    message="Archived successfully"
                ))
                
                # 更新进度
                progress = (i + 1) / len(request.target_ids) * 100
                await self._update_progress(request.operation, progress, current_user)
                
            except Exception as e:
                logger.error(f"Failed to archive {target_id}: {e}")
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=False,
                    message=str(e),
                    error_code="ARCHIVE_FAILED"
                ))
                db.rollback()
        
        return results
    
    async def _batch_restore(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> List[BatchOperationResult]:
        """批量恢复操作"""
        results = []
        model_class = self.get_target_model(request.target_type)
        
        if not model_class:
            raise ValueError(f"Unsupported target type: {request.target_type}")
        
        for i, target_id in enumerate(request.target_ids):
            try:
                # 查找目标资源（包括已删除/归档的）
                query = db.query(model_class).filter(model_class.id == target_id)
                target = query.first()
                
                if not target:
                    results.append(BatchOperationResult(
                        target_id=target_id,
                        success=False,
                        message=f"{request.target_type} not found"
                    ))
                    continue
                
                # 检查权限
                if hasattr(target, 'sender_id') or hasattr(target, 'user_id'):
                    owner_id = getattr(target, 'sender_id', None) or getattr(target, 'user_id', None)
                    if owner_id != current_user:
                        results.append(BatchOperationResult(
                            target_id=target_id,
                            success=False,
                            message="Permission denied",
                            error_code="PERMISSION_DENIED"
                        ))
                        continue
                
                if not request.dry_run:
                    # 恢复处理
                    if hasattr(target, 'is_deleted'):
                        target.is_deleted = False
                        target.deleted_at = None
                    if hasattr(target, 'is_archived'):
                        target.is_archived = False
                        target.archived_at = None
                    if hasattr(target, 'updated_at'):
                        target.updated_at = datetime.utcnow()
                    
                    db.commit()
                
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=True,
                    message="Restored successfully"
                ))
                
                # 更新进度
                progress = (i + 1) / len(request.target_ids) * 100
                await self._update_progress(request.operation, progress, current_user)
                
            except Exception as e:
                logger.error(f"Failed to restore {target_id}: {e}")
                results.append(BatchOperationResult(
                    target_id=target_id,
                    success=False,
                    message=str(e),
                    error_code="RESTORE_FAILED"
                ))
                db.rollback()
        
        return results
    
    async def _batch_create(
        self,
        request: BatchOperationRequest,
        db: Session,
        current_user: str
    ) -> List[BatchOperationResult]:
        """批量创建操作"""
        results = []
        model_class = self.get_target_model(request.target_type)
        
        if not model_class:
            raise ValueError(f"Unsupported target type: {request.target_type}")
        
        items_data = request.operation_data.get('items', [])
        continue_on_error = request.operation_data.get('continue_on_error', True)
        
        for i, item_data in enumerate(items_data):
            try:
                if not request.dry_run:
                    # 创建新对象
                    new_item = model_class(**item_data)
                    
                    # 设置创建者信息
                    if hasattr(new_item, 'sender_id'):
                        new_item.sender_id = current_user
                    elif hasattr(new_item, 'user_id'):
                        new_item.user_id = current_user
                    
                    # 设置时间戳
                    if hasattr(new_item, 'created_at'):
                        new_item.created_at = datetime.utcnow()
                    
                    db.add(new_item)
                    db.commit()
                    db.refresh(new_item)
                    
                    results.append(BatchOperationResult(
                        target_id=str(new_item.id),
                        success=True,
                        message="Created successfully",
                        data={"created_id": str(new_item.id)}
                    ))
                else:
                    results.append(BatchOperationResult(
                        target_id=f"item_{i}",
                        success=True,
                        message="Creation validated (dry run)"
                    ))
                
                # 更新进度
                progress = (i + 1) / len(items_data) * 100
                await self._update_progress(request.operation, progress, current_user)
                
            except Exception as e:
                logger.error(f"Failed to create item {i}: {e}")
                results.append(BatchOperationResult(
                    target_id=f"item_{i}",
                    success=False,
                    message=str(e),
                    error_code="CREATE_FAILED"
                ))
                
                if not continue_on_error:
                    break
                    
                db.rollback()
        
        return results
    
    async def _update_progress(self, operation: str, progress: float, user_id: str):
        """更新操作进度"""
        try:
            await notify_batch_operation_progress(user_id, operation, {"progress": progress})
        except Exception as e:
            logger.warning(f"Failed to send progress update: {e}")
    
    def get_job_status(self, job_id: str) -> Optional[BatchJobStatus]:
        """获取作业状态"""
        return self.active_jobs.get(job_id)
    
    def cancel_job(self, job_id: str) -> bool:
        """取消作业"""
        if job_id in self.active_jobs:
            job = self.active_jobs[job_id]
            job.status = "cancelled"
            job.updated_at = datetime.utcnow()
            return True
        return False
    
    def cleanup_completed_jobs(self, older_than_hours: int = 24):
        """清理已完成的作业"""
        cutoff_time = datetime.utcnow() - timedelta(hours=older_than_hours)
        to_remove = []
        
        for job_id, job in self.active_jobs.items():
            if job.status in ["completed", "failed", "cancelled"] and job.updated_at < cutoff_time:
                to_remove.append(job_id)
        
        for job_id in to_remove:
            del self.active_jobs[job_id]
        
        logger.info(f"Cleaned up {len(to_remove)} completed batch jobs")


# 全局服务实例
batch_service = BatchOperationService()


def get_batch_service() -> BatchOperationService:
    """获取批量操作服务实例"""
    return batch_service