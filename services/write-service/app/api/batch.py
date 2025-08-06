"""
批量操作API
"""
import logging
from fastapi import APIRouter, Depends, HTTPException, status, BackgroundTasks, Query
from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any
from datetime import datetime

from app.core.database import get_db, get_async_session
from app.utils.auth import get_current_user, check_admin_permission
from app.utils.batch_service import get_batch_service
from app.schemas.batch import (
    BatchOperationRequest, BatchOperationResponse, BatchJobStatus,
    BatchLetterOperationRequest, BatchProductOperationRequest, BatchOrderOperationRequest,
    BatchStatusUpdateRequest, BatchDeleteRequest, BatchCreateRequest,
    BatchExportRequest, BatchArchiveRequest, BatchOperationEnum, BatchTargetEnum,
    SuccessResponse
)

router = APIRouter()
logger = logging.getLogger(__name__)


@router.post("/execute", response_model=SuccessResponse, summary="执行批量操作")
async def execute_batch_operation(
    request: BatchOperationRequest,
    background_tasks: BackgroundTasks,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    执行批量操作
    
    支持的操作类型:
    - delete: 批量删除
    - update: 批量更新
    - status_update: 批量状态更新
    - export: 批量导出
    - archive: 批量归档
    - restore: 批量恢复
    - bulk_create: 批量创建
    """
    try:
        batch_service = get_batch_service()
        
        # 对于某些操作，检查管理员权限
        if request.operation in [BatchOperationEnum.DELETE, BatchOperationEnum.ARCHIVE]:
            await check_admin_permission(current_user)
        
        # 执行批量操作
        response = await batch_service.execute_batch_operation(
            request=request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg="批量操作执行成功" if not request.dry_run else "批量操作验证成功（试运行）",
            data=response.model_dump()
        )
        
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except PermissionError as e:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Batch operation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量操作执行失败: {str(e)}"
        )


@router.post("/letters/delete", response_model=SuccessResponse, summary="批量删除信件")
async def batch_delete_letters(
    request: BatchDeleteRequest,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量删除信件"""
    try:
        batch_service = get_batch_service()
        
        batch_request = BatchLetterOperationRequest(
            operation=BatchOperationEnum.DELETE,
            target_ids=request.target_ids,
            operation_data={
                "soft_delete": request.soft_delete,
                "delete_reason": request.delete_reason
            },
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量删除完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Batch delete letters failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量删除信件失败: {str(e)}"
        )


@router.post("/letters/status", response_model=SuccessResponse, summary="批量更新信件状态")
async def batch_update_letter_status(
    request: BatchStatusUpdateRequest,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量更新信件状态"""
    try:
        batch_service = get_batch_service()
        
        batch_request = BatchLetterOperationRequest(
            operation=BatchOperationEnum.STATUS_UPDATE,
            target_ids=request.target_ids,
            operation_data={
                "new_status": request.new_status,
                "reason": request.reason,
                "force": request.force
            },
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量状态更新完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Batch update letter status failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量更新信件状态失败: {str(e)}"
        )


@router.post("/products/delete", response_model=SuccessResponse, summary="批量删除商品")
async def batch_delete_products(
    request: BatchDeleteRequest,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量删除商品"""
    try:
        # 检查管理员权限
        await check_admin_permission(current_user)
        
        batch_service = get_batch_service()
        
        batch_request = BatchProductOperationRequest(
            operation=BatchOperationEnum.DELETE,
            target_ids=request.target_ids,
            operation_data={
                "soft_delete": request.soft_delete,
                "delete_reason": request.delete_reason
            },
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量删除商品完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Batch delete products failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量删除商品失败: {str(e)}"
        )


@router.post("/products/status", response_model=SuccessResponse, summary="批量更新商品状态")
async def batch_update_product_status(
    request: BatchStatusUpdateRequest,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量更新商品状态"""
    try:
        # 检查管理员权限
        await check_admin_permission(current_user)
        
        batch_service = get_batch_service()
        
        batch_request = BatchProductOperationRequest(
            operation=BatchOperationEnum.STATUS_UPDATE,
            target_ids=request.target_ids,
            operation_data={
                "new_status": request.new_status,
                "reason": request.reason,
                "force": request.force
            },
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量更新商品状态完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Batch update product status failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量更新商品状态失败: {str(e)}"
        )


@router.post("/archive", response_model=SuccessResponse, summary="批量归档")
async def batch_archive(
    request: BatchArchiveRequest,
    target_type: BatchTargetEnum = Query(description="目标资源类型"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量归档资源"""
    try:
        batch_service = get_batch_service()
        
        batch_request = BatchOperationRequest(
            operation=BatchOperationEnum.ARCHIVE,
            target_type=target_type,
            target_ids=request.target_ids,
            operation_data={
                "archive_reason": request.archive_reason,
                "archive_location": request.archive_location
            },
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量归档完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Batch archive failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量归档失败: {str(e)}"
        )


@router.post("/restore", response_model=SuccessResponse, summary="批量恢复")
async def batch_restore(
    target_ids: List[str],
    target_type: BatchTargetEnum = Query(description="目标资源类型"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量恢复已删除或归档的资源"""
    try:
        batch_service = get_batch_service()
        
        batch_request = BatchOperationRequest(
            operation=BatchOperationEnum.RESTORE,
            target_type=target_type,
            target_ids=target_ids,
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量恢复完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Batch restore failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量恢复失败: {str(e)}"
        )


@router.post("/export", response_model=SuccessResponse, summary="批量导出")
async def batch_export(
    request: BatchExportRequest,
    target_type: BatchTargetEnum = Query(description="目标资源类型"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量导出数据"""
    try:
        batch_service = get_batch_service()
        
        batch_request = BatchOperationRequest(
            operation=BatchOperationEnum.EXPORT,
            target_type=target_type,
            target_ids=request.target_ids or [],
            operation_data={
                "export_format": request.export_format,
                "include_fields": request.include_fields,
                "exclude_fields": request.exclude_fields,
                "filters": request.filters
            },
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量导出完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Batch export failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量导出失败: {str(e)}"
        )


@router.post("/bulk-create", response_model=SuccessResponse, summary="批量创建")
async def bulk_create(
    request: BatchCreateRequest,
    target_type: BatchTargetEnum = Query(description="目标资源类型"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """批量创建资源"""
    try:
        batch_service = get_batch_service()
        
        batch_request = BatchOperationRequest(
            operation=BatchOperationEnum.BULK_CREATE,
            target_type=target_type,
            target_ids=[],  # 创建操作不需要目标ID
            operation_data={
                "items": request.items,
                "skip_validation": request.skip_validation,
                "continue_on_error": request.continue_on_error
            },
            dry_run=False
        )
        
        response = await batch_service.execute_batch_operation(
            request=batch_request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg=f"批量创建完成，成功: {response.success_count}/{response.total_count}",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Bulk create failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"批量创建失败: {str(e)}"
        )


@router.get("/jobs/{job_id}", response_model=SuccessResponse, summary="获取批量作业状态")
async def get_batch_job_status(
    job_id: str,
    current_user: str = Depends(get_current_user)
):
    """获取批量作业的执行状态"""
    try:
        batch_service = get_batch_service()
        job_status = batch_service.get_job_status(job_id)
        
        if not job_status:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="批量作业不存在或已过期"
            )
        
        return SuccessResponse(
            msg="获取作业状态成功",
            data=job_status.model_dump()
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get batch job status failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取作业状态失败: {str(e)}"
        )


@router.post("/jobs/{job_id}/cancel", response_model=SuccessResponse, summary="取消批量作业")
async def cancel_batch_job(
    job_id: str,
    current_user: str = Depends(get_current_user)
):
    """取消正在执行的批量作业"""
    try:
        batch_service = get_batch_service()
        cancelled = batch_service.cancel_job(job_id)
        
        if not cancelled:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="批量作业不存在或无法取消"
            )
        
        return SuccessResponse(
            msg="批量作业已取消",
            data={"job_id": job_id, "cancelled": True}
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Cancel batch job failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"取消作业失败: {str(e)}"
        )


@router.get("/export/{export_id}", summary="下载导出文件")
async def download_export(
    export_id: str,
    current_user: str = Depends(get_current_user)
):
    """下载批量导出生成的文件"""
    try:
        batch_service = get_batch_service()
        export_data = batch_service.cache.get(f"export:{export_id}")
        
        if not export_data:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="导出文件不存在或已过期"
            )
        
        # 检查权限
        if export_data.get("created_by") != current_user:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="无权访问此导出文件"
            )
        
        from fastapi.responses import JSONResponse
        
        return JSONResponse(
            content={
                "export_id": export_id,
                "format": export_data.get("format", "json"),
                "data": export_data.get("data", []),
                "created_at": export_data.get("created_at"),
                "created_by": export_data.get("created_by")
            },
            media_type="application/json"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Download export failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"下载导出文件失败: {str(e)}"
        )


@router.post("/validate", response_model=SuccessResponse, summary="验证批量操作")
async def validate_batch_operation(
    request: BatchOperationRequest,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """验证批量操作（试运行模式）"""
    try:
        batch_service = get_batch_service()
        
        # 设置为试运行模式
        request.dry_run = True
        
        response = await batch_service.execute_batch_operation(
            request=request,
            db=db,
            current_user=current_user
        )
        
        return SuccessResponse(
            msg="批量操作验证完成",
            data=response.model_dump()
        )
        
    except Exception as e:
        logger.error(f"Validate batch operation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"验证批量操作失败: {str(e)}"
        )


@router.get("/health", summary="批量操作服务健康检查")
async def batch_service_health():
    """批量操作服务健康检查"""
    try:
        batch_service = get_batch_service()
        
        # 检查服务状态
        active_jobs_count = len(batch_service.active_jobs)
        
        return SuccessResponse(
            msg="批量操作服务运行正常",
            data={
                "status": "healthy",
                "active_jobs": active_jobs_count,
                "supported_operations": [op.value for op in BatchOperationEnum],
                "supported_targets": [target.value for target in BatchTargetEnum],
                "timestamp": datetime.utcnow().isoformat()
            }
        )
        
    except Exception as e:
        logger.error(f"Batch service health check failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="批量操作服务不可用"
        )


# 管理员专用接口
@router.get("/admin/jobs", response_model=SuccessResponse, summary="获取所有批量作业状态")
async def get_all_batch_jobs(
    current_user: str = Depends(get_current_user)
):
    """获取所有批量作业状态（管理员专用）"""
    try:
        # 检查管理员权限
        await check_admin_permission(current_user)
        
        batch_service = get_batch_service()
        all_jobs = {
            job_id: job.model_dump() 
            for job_id, job in batch_service.active_jobs.items()
        }
        
        return SuccessResponse(
            msg="获取所有作业状态成功",
            data={
                "jobs": all_jobs,
                "total_jobs": len(all_jobs)
            }
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get all batch jobs failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取所有作业状态失败: {str(e)}"
        )


@router.post("/admin/cleanup", response_model=SuccessResponse, summary="清理已完成的作业")
async def cleanup_completed_jobs(
    older_than_hours: int = Query(default=24, ge=1, le=720, description="清理多少小时前的作业"),
    current_user: str = Depends(get_current_user)
):
    """清理已完成的批量作业（管理员专用）"""
    try:
        # 检查管理员权限
        await check_admin_permission(current_user)
        
        batch_service = get_batch_service()
        initial_count = len(batch_service.active_jobs)
        
        batch_service.cleanup_completed_jobs(older_than_hours)
        
        final_count = len(batch_service.active_jobs)
        cleaned_count = initial_count - final_count
        
        return SuccessResponse(
            msg=f"清理完成，清理了 {cleaned_count} 个已完成的作业",
            data={
                "initial_jobs": initial_count,
                "remaining_jobs": final_count,
                "cleaned_jobs": cleaned_count,
                "cleanup_threshold_hours": older_than_hours
            }
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Cleanup completed jobs failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"清理作业失败: {str(e)}"
        )