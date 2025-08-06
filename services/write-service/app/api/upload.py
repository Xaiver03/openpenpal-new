"""
文件上传API
支持普通上传、分片上传、图片压缩等功能
"""

from fastapi import APIRouter, UploadFile, File, Form, HTTPException, Depends, BackgroundTasks
from typing import List, Optional
import uuid

from app.utils.file_upload_service import get_file_upload_service, FileUploadService
from app.utils.auth import get_current_user
from app.core.responses import success_response, error_response
from app.core.logger import get_logger

router = APIRouter(prefix="/api/upload", tags=["文件上传"])
logger = get_logger(__name__)


@router.post("/single", summary="单文件上传")
async def upload_single_file(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    compress_images: bool = Form(True),
    use_cdn: Optional[bool] = Form(None),
    current_user: dict = Depends(get_current_user),
    upload_service: FileUploadService = Depends(get_file_upload_service)
):
    """
    上传单个文件
    
    - **file**: 要上传的文件
    - **compress_images**: 是否压缩图片（默认True）
    - **use_cdn**: 是否使用CDN（None为自动判断）
    """
    try:
        file_info = await upload_service.upload_file(
            file=file,
            compress_images=compress_images,
            use_cdn=use_cdn
        )
        
        logger.info(f"User {current_user.get('user_id')} uploaded file: {file_info['id']}")
        
        return success_response(
            data=file_info,
            message="文件上传成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"File upload error: {str(e)}")
        return error_response(message=f"文件上传失败: {str(e)}")


@router.post("/multiple", summary="多文件上传")
async def upload_multiple_files(
    background_tasks: BackgroundTasks,
    files: List[UploadFile] = File(...),
    compress_images: bool = Form(True),
    use_cdn: Optional[bool] = Form(None),
    current_user: dict = Depends(get_current_user),
    upload_service: FileUploadService = Depends(get_file_upload_service)
):
    """
    批量上传多个文件
    
    - **files**: 要上传的文件列表
    - **compress_images**: 是否压缩图片
    - **use_cdn**: 是否使用CDN
    """
    try:
        if len(files) > 10:  # 限制单次最多上传10个文件
            raise HTTPException(status_code=400, detail="单次最多上传10个文件")
        
        upload_results = []
        failed_files = []
        
        for file in files:
            try:
                file_info = await upload_service.upload_file(
                    file=file,
                    compress_images=compress_images,
                    use_cdn=use_cdn
                )
                upload_results.append(file_info)
                
            except Exception as e:
                failed_files.append({
                    'filename': file.filename,
                    'error': str(e)
                })
                logger.error(f"Failed to upload {file.filename}: {str(e)}")
        
        logger.info(f"User {current_user.get('user_id')} uploaded {len(upload_results)} files")
        
        return success_response(
            data={
                'uploaded_files': upload_results,
                'failed_files': failed_files,
                'total_count': len(files),
                'success_count': len(upload_results),
                'failed_count': len(failed_files)
            },
            message=f"批量上传完成，成功{len(upload_results)}个，失败{len(failed_files)}个"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Multiple file upload error: {str(e)}")
        return error_response(message=f"批量上传失败: {str(e)}")


@router.post("/chunk/init", summary="初始化分片上传")
async def init_chunk_upload(
    filename: str = Form(...),
    file_size: int = Form(...),
    chunk_size: int = Form(5 * 1024 * 1024),  # 默认5MB
    current_user: dict = Depends(get_current_user)
):
    """
    初始化分片上传
    
    - **filename**: 文件名
    - **file_size**: 文件总大小
    - **chunk_size**: 分片大小
    """
    try:
        # 验证参数
        max_file_size = 500 * 1024 * 1024  # 最大500MB
        if file_size > max_file_size:
            raise HTTPException(status_code=400, detail="文件过大")
        
        if chunk_size > 10 * 1024 * 1024:  # 最大10MB分片
            raise HTTPException(status_code=400, detail="分片过大")
        
        # 生成上传ID
        chunk_id = str(uuid.uuid4())
        total_chunks = (file_size + chunk_size - 1) // chunk_size
        
        chunk_info = {
            'chunk_id': chunk_id,
            'filename': filename,
            'file_size': file_size,
            'chunk_size': chunk_size,
            'total_chunks': total_chunks,
            'uploaded_chunks': 0,
            'user_id': current_user.get('user_id')
        }
        
        logger.info(f"Chunk upload initialized: {chunk_id} for user {current_user.get('user_id')}")
        
        return success_response(
            data=chunk_info,
            message="分片上传初始化成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Chunk upload init error: {str(e)}")
        return error_response(message=f"分片上传初始化失败: {str(e)}")


@router.post("/chunk/upload", summary="上传文件分片")
async def upload_chunk(
    chunk_id: str = Form(...),
    chunk_index: int = Form(...),
    total_chunks: int = Form(...),
    filename: str = Form(...),
    chunk: UploadFile = File(...),
    current_user: dict = Depends(get_current_user),
    upload_service: FileUploadService = Depends(get_file_upload_service)
):
    """
    上传文件分片
    
    - **chunk_id**: 分片组ID
    - **chunk_index**: 当前分片索引（从0开始）
    - **total_chunks**: 总分片数
    - **filename**: 文件名
    - **chunk**: 分片数据
    """
    try:
        # 读取分片数据
        chunk_data = await chunk.read()
        
        # 上传分片
        result = await upload_service.upload_chunks(
            chunk_id=chunk_id,
            chunk_index=chunk_index,
            total_chunks=total_chunks,
            chunk_data=chunk_data,
            filename=filename
        )
        
        logger.info(f"Chunk uploaded: {chunk_id}/{chunk_index} by user {current_user.get('user_id')}")
        
        return success_response(
            data=result,
            message="分片上传成功" if result['status'] == 'uploading' else "文件合并完成"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Chunk upload error: {str(e)}")
        return error_response(message=f"分片上传失败: {str(e)}")


@router.get("/info/{file_hash}", summary="获取文件信息")
async def get_file_info(
    file_hash: str,
    current_user: dict = Depends(get_current_user),
    upload_service: FileUploadService = Depends(get_file_upload_service)
):
    """
    获取文件信息
    
    - **file_hash**: 文件哈希值
    """
    try:
        file_info = await upload_service.get_file_info(file_hash)
        
        if not file_info:
            raise HTTPException(status_code=404, detail="文件不存在")
        
        return success_response(
            data=file_info,
            message="获取文件信息成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get file info error: {str(e)}")
        return error_response(message=f"获取文件信息失败: {str(e)}")


@router.delete("/{file_hash}", summary="删除文件")
async def delete_file(
    file_hash: str,
    current_user: dict = Depends(get_current_user),
    upload_service: FileUploadService = Depends(get_file_upload_service)
):
    """
    删除文件
    
    - **file_hash**: 文件哈希值
    """
    try:
        success = await upload_service.delete_file(file_hash)
        
        if not success:
            raise HTTPException(status_code=404, detail="文件不存在")
        
        logger.info(f"File deleted: {file_hash} by user {current_user.get('user_id')}")
        
        return success_response(
            message="文件删除成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Delete file error: {str(e)}")
        return error_response(message=f"删除文件失败: {str(e)}")


@router.get("/stats", summary="上传统计信息")
async def get_upload_stats(
    current_user: dict = Depends(get_current_user)
):
    """
    获取用户上传统计信息
    """
    try:
        # TODO: 实现用户上传统计
        stats = {
            'total_files': 0,
            'total_size': 0,
            'recent_uploads': []
        }
        
        return success_response(
            data=stats,
            message="获取统计信息成功"
        )
        
    except Exception as e:
        logger.error(f"Get upload stats error: {str(e)}")
        return error_response(message=f"获取统计信息失败: {str(e)}")