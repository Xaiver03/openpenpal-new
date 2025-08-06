from flask import Blueprint, request, current_app
import time
from app.utils.response import (
    success_response, 
    validation_error_response, 
    not_found_error_response,
    internal_error_response,
    permission_error_response
)
from app.utils.auth import jwt_required, get_current_user, validate_user_access
from app.services.cache_service import get_cache_service

tasks_bp = Blueprint('tasks', __name__)


@tasks_bp.route('/<task_id>', methods=['GET'])
@jwt_required
def get_task_status(task_id):
    """获取OCR任务状态"""
    try:
        user_info = get_current_user()
        user_id = user_info.get('user_id')
        
        if not task_id.startswith('ocr_task_'):
            return not_found_error_response("任务不存在"), 404
        
        try:
            # 获取缓存服务
            cache_service = get_cache_service()
            
            # 尝试从缓存获取任务状态
            cached_task = cache_service.get_task_status(task_id)
            
            if cached_task:
                return success_response(cached_task)
            
            # 如果缓存中没有，返回默认的进行中状态
            # 在实际应用中，这里应该查询数据库
            mock_task = {
                "task_id": task_id,
                "status": "processing",
                "progress": 50,
                "created_at": "2025-07-20T11:58:30Z",
                "completed_at": None,
                "result": None,
                "error": None,
                "message": "任务正在处理中..."
            }
            
            return success_response(mock_task)
            
        except Exception as e:
            current_app.logger.error(f"获取任务状态异常: {str(e)}")
            # 降级返回模拟数据
            mock_task = {
                "task_id": task_id,
                "status": "unknown",
                "progress": 0,
                "error": f"获取状态失败: {str(e)}"
            }
            return success_response(mock_task)
        
    except Exception as e:
        current_app.logger.error(f"获取任务状态失败: {str(e)}")
        return internal_error_response("获取任务状态失败"), 500


@tasks_bp.route('/', methods=['GET'])
@jwt_required
def get_user_tasks():
    """获取用户的OCR任务列表"""
    try:
        user_info = get_current_user()
        user_id = user_info.get('user_id')
        
        # 获取查询参数
        page = int(request.args.get('page', 1))
        limit = int(request.args.get('limit', 10))
        status = request.args.get('status')  # processing, completed, failed
        
        # 限制每页数量
        limit = min(limit, 50)
        
        # TODO: 从数据库中查询用户任务
        # 这里使用模拟数据
        
        mock_tasks = [
            {
                "task_id": f"ocr_task_{i}",
                "status": "completed" if i % 2 == 0 else "processing",
                "created_at": "2025-07-20T12:00:00Z",
                "processing_time": 2.3 if i % 2 == 0 else None,
                "confidence": 0.85 if i % 2 == 0 else None
            }
            for i in range(1, 26)  # 模拟25个任务
        ]
        
        # 状态过滤
        if status:
            mock_tasks = [task for task in mock_tasks if task['status'] == status]
        
        # 分页
        total = len(mock_tasks)
        start = (page - 1) * limit
        end = start + limit
        items = mock_tasks[start:end]
        
        result = {
            "items": items,
            "pagination": {
                "page": page,
                "limit": limit,
                "total": total,
                "pages": (total + limit - 1) // limit,
                "has_next": end < total,
                "has_prev": page > 1
            }
        }
        
        return success_response(result)
        
    except Exception as e:
        current_app.logger.error(f"获取任务列表失败: {str(e)}")
        return internal_error_response("获取任务列表失败"), 500


@tasks_bp.route('/<task_id>', methods=['DELETE'])
@jwt_required
def delete_task(task_id):
    """删除OCR任务"""
    try:
        user_info = get_current_user()
        user_id = user_info.get('user_id')
        
        # TODO: 验证任务所有权并删除
        # 这里模拟删除操作
        
        if not task_id.startswith('ocr_task_'):
            return not_found_error_response("任务不存在"), 404
        
        # 模拟删除成功
        return success_response({"task_id": task_id}, "任务删除成功")
        
    except Exception as e:
        current_app.logger.error(f"删除任务失败: {str(e)}")
        return internal_error_response("删除任务失败"), 500


@tasks_bp.route('/batch/<batch_id>/progress', methods=['GET'])
@jwt_required
def get_batch_progress(batch_id):
    """获取批量任务进度"""
    try:
        user_info = get_current_user()
        user_id = user_info.get('user_id')
        
        if not batch_id.startswith('batch_'):
            return not_found_error_response("批量任务不存在"), 404
        
        # 使用批量处理器获取进度
        from app.services.batch_processor import get_batch_processor
        batch_processor = get_batch_processor()
        
        progress_data = batch_processor.get_batch_progress(batch_id)
        
        if not progress_data:
            return not_found_error_response("批量任务不存在或已过期"), 404
        
        # 验证用户权限 - 只有任务创建者可以查看
        if progress_data.get('user_id') != user_id:
            return permission_error_response("无权限访问此批量任务"), 403
        
        # 构建响应数据
        response_data = {
            "batch_id": progress_data['batch_id'],
            "total_images": progress_data['total_images'],
            "completed_images": progress_data['completed_images'],
            "failed_images": progress_data['failed_images'],
            "progress_percentage": progress_data['progress_percentage'],
            "status": progress_data['status'],
            "current_step": progress_data['current_step'],
            "results": progress_data.get('results', []),
            "created_at": progress_data.get('created_at'),
            "completed_at": progress_data.get('completed_at'),
            "statistics": progress_data.get('statistics', {}),
            "error": progress_data.get('error')
        }
        
        # 计算预估剩余时间
        if progress_data['status'] == 'processing' and progress_data['progress_percentage'] > 0:
            elapsed_time = time.time() - progress_data.get('created_at', time.time())
            remaining_images = progress_data['total_images'] - progress_data['completed_images'] - progress_data['failed_images']
            if remaining_images > 0 and progress_data['completed_images'] > 0:
                avg_time_per_image = elapsed_time / (progress_data['completed_images'] + progress_data['failed_images'])
                estimated_remaining = remaining_images * avg_time_per_image
                response_data['estimated_time_remaining'] = f"{int(estimated_remaining)}s"
            else:
                response_data['estimated_time_remaining'] = "计算中..."
        else:
            response_data['estimated_time_remaining'] = None
        
        return success_response(response_data)
        
    except Exception as e:
        current_app.logger.error(f"获取批量任务进度失败: {str(e)}")
        return internal_error_response("获取批量任务进度失败"), 500