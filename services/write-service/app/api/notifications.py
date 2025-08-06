"""
通知管理API
支持通知发送、获取、偏好设置管理等功能
"""

from fastapi import APIRouter, Depends, HTTPException, Query
from typing import List, Optional, Dict, Any
from pydantic import BaseModel
from datetime import datetime

from app.utils.notification_service import (
    get_notification_service, NotificationService,
    NotificationType, NotificationChannel, NotificationPriority
)
from app.utils.auth import get_current_user
from app.core.responses import success_response, error_response
from app.core.logger import get_logger

router = APIRouter(prefix="/api/notifications", tags=["通知管理"])
logger = get_logger(__name__)


class SendNotificationRequest(BaseModel):
    """发送通知请求"""
    user_id: Optional[str] = None  # 如果不指定，发送给当前用户
    notification_type: NotificationType
    channels: List[NotificationChannel]
    data: Dict[str, Any]
    priority: NotificationPriority = NotificationPriority.NORMAL
    template_data: Optional[Dict[str, Any]] = None


class BulkNotificationRequest(BaseModel):
    """批量发送通知请求"""
    user_ids: List[str]
    notification_type: NotificationType
    channels: List[NotificationChannel]
    data: Dict[str, Any]
    priority: NotificationPriority = NotificationPriority.NORMAL
    template_data: Optional[Dict[str, Any]] = None


class NotificationPreferencesRequest(BaseModel):
    """通知偏好设置请求"""
    websocket_enabled: bool = True
    email_enabled: bool = True
    push_enabled: bool = True
    notification_types: Dict[str, bool]


@router.post("/send", summary="发送单个通知")
async def send_notification(
    request: SendNotificationRequest,
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """
    发送通知到指定用户
    
    - **user_id**: 目标用户ID，不指定则发送给当前用户
    - **notification_type**: 通知类型
    - **channels**: 发送渠道列表
    - **data**: 通知数据
    - **priority**: 优先级
    - **template_data**: 模板数据
    """
    try:
        target_user_id = request.user_id or current_user.get('user_id')
        
        result = await notification_service.send_notification(
            user_id=target_user_id,
            notification_type=request.notification_type,
            channels=request.channels,
            data=request.data,
            priority=request.priority,
            template_data=request.template_data
        )
        
        logger.info(f"Notification sent by user {current_user.get('user_id')} to user {target_user_id}")
        
        return success_response(
            data=result,
            message="通知发送成功"
        )
        
    except Exception as e:
        logger.error(f"Send notification error: {str(e)}")
        return error_response(message=f"发送通知失败: {str(e)}")


@router.post("/send/bulk", summary="批量发送通知")
async def send_bulk_notification(
    request: BulkNotificationRequest,
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """
    批量发送通知
    
    - **user_ids**: 目标用户ID列表
    - **notification_type**: 通知类型
    - **channels**: 发送渠道列表
    - **data**: 通知数据
    - **priority**: 优先级
    - **template_data**: 模板数据
    """
    try:
        # 检查权限（只有管理员可以批量发送）
        user_role = current_user.get('role', 'user')
        if user_role not in ['admin', 'super_admin']:
            raise HTTPException(status_code=403, detail="权限不足，只有管理员可以批量发送通知")
        
        if len(request.user_ids) > 1000:  # 限制批量发送数量
            raise HTTPException(status_code=400, detail="单次最多发送给1000个用户")
        
        result = await notification_service.send_bulk_notification(
            user_ids=request.user_ids,
            notification_type=request.notification_type,
            channels=request.channels,
            data=request.data,
            priority=request.priority,
            template_data=request.template_data
        )
        
        logger.info(f"Bulk notification sent by admin {current_user.get('user_id')} to {len(request.user_ids)} users")
        
        return success_response(
            data=result,
            message=f"批量通知发送完成，成功{result['success_count']}个，失败{result['failed_count']}个"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Bulk notification error: {str(e)}")
        return error_response(message=f"批量发送失败: {str(e)}")


@router.get("/list", summary="获取用户通知列表")
async def get_user_notifications(
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """
    获取当前用户的通知列表
    
    - **limit**: 每页数量（1-100）
    - **offset**: 偏移量
    """
    try:
        user_id = current_user.get('user_id')
        
        result = await notification_service.get_user_notifications(
            user_id=user_id,
            limit=limit,
            offset=offset
        )
        
        return success_response(
            data=result,
            message="获取通知列表成功"
        )
        
    except Exception as e:
        logger.error(f"Get notifications error: {str(e)}")
        return error_response(message=f"获取通知列表失败: {str(e)}")


@router.get("/preferences", summary="获取通知偏好设置")
async def get_notification_preferences(
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """获取当前用户的通知偏好设置"""
    try:
        user_id = current_user.get('user_id')
        
        preferences = await notification_service._get_user_notification_preferences(user_id)
        
        return success_response(
            data=preferences,
            message="获取偏好设置成功"
        )
        
    except Exception as e:
        logger.error(f"Get preferences error: {str(e)}")
        return error_response(message=f"获取偏好设置失败: {str(e)}")


@router.put("/preferences", summary="更新通知偏好设置")
async def update_notification_preferences(
    request: NotificationPreferencesRequest,
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """
    更新用户通知偏好设置
    
    - **websocket_enabled**: 是否启用WebSocket通知
    - **email_enabled**: 是否启用邮件通知
    - **push_enabled**: 是否启用推送通知
    - **notification_types**: 各类通知的开关设置
    """
    try:
        user_id = current_user.get('user_id')
        
        # 保存偏好设置到缓存
        if notification_service.cache_manager:
            await notification_service.cache_manager.set(
                f"user:preferences:{user_id}",
                request.dict(),
                expire=86400 * 30  # 缓存30天
            )
        
        logger.info(f"User {user_id} updated notification preferences")
        
        return success_response(
            data=request.dict(),
            message="偏好设置更新成功"
        )
        
    except Exception as e:
        logger.error(f"Update preferences error: {str(e)}")
        return error_response(message=f"更新偏好设置失败: {str(e)}")


@router.get("/stats", summary="获取通知统计信息")
async def get_notification_stats(
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """获取通知统计信息（仅管理员可访问）"""
    try:
        # 检查权限
        user_role = current_user.get('role', 'user')
        if user_role not in ['admin', 'super_admin']:
            raise HTTPException(status_code=403, detail="权限不足")
        
        # TODO: 实现通知统计逻辑
        stats = {
            'total_notifications_sent': 0,
            'notifications_by_type': {},
            'notifications_by_channel': {},
            'delivery_success_rate': 0.0,
            'recent_notifications': []
        }
        
        return success_response(
            data=stats,
            message="获取统计信息成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Get notification stats error: {str(e)}")
        return error_response(message=f"获取统计信息失败: {str(e)}")


@router.post("/test", summary="测试通知发送")
async def test_notification(
    notification_type: NotificationType = NotificationType.SYSTEM_ANNOUNCEMENT,
    channels: List[NotificationChannel] = [NotificationChannel.WEBSOCKET],
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """
    测试通知发送功能
    
    - **notification_type**: 通知类型
    - **channels**: 发送渠道
    """
    try:
        user_id = current_user.get('user_id')
        
        test_data = {
            'test_message': '这是一条测试通知',
            'sent_by': user_id,
            'timestamp': datetime.utcnow().isoformat()
        }
        
        result = await notification_service.send_notification(
            user_id=user_id,
            notification_type=notification_type,
            channels=channels,
            data=test_data,
            priority=NotificationPriority.LOW,
            template_data={'announcement_content': '测试通知内容'}
        )
        
        logger.info(f"Test notification sent to user {user_id}")
        
        return success_response(
            data=result,
            message="测试通知发送成功"
        )
        
    except Exception as e:
        logger.error(f"Test notification error: {str(e)}")
        return error_response(message=f"测试通知发送失败: {str(e)}")


# 系统通知相关端点
@router.post("/system/announcement", summary="发送系统公告")
async def send_system_announcement(
    title: str,
    content: str,
    target_users: Optional[List[str]] = None,  # None表示发送给所有用户
    channels: List[NotificationChannel] = [NotificationChannel.WEBSOCKET, NotificationChannel.EMAIL],
    current_user: dict = Depends(get_current_user),
    notification_service: NotificationService = Depends(get_notification_service)
):
    """
    发送系统公告（仅管理员可用）
    
    - **title**: 公告标题
    - **content**: 公告内容
    - **target_users**: 目标用户列表（不指定则发送给所有在线用户）
    - **channels**: 发送渠道
    """
    try:
        # 检查权限
        user_role = current_user.get('role', 'user')
        if user_role not in ['admin', 'super_admin']:
            raise HTTPException(status_code=403, detail="权限不足，只有管理员可以发送系统公告")
        
        announcement_data = {
            'title': title,
            'content': content,
            'sent_by': current_user.get('user_id'),
            'timestamp': datetime.utcnow().isoformat()
        }
        
        template_data = {
            'announcement_content': content
        }
        
        if target_users:
            # 发送给指定用户
            result = await notification_service.send_bulk_notification(
                user_ids=target_users,
                notification_type=NotificationType.SYSTEM_ANNOUNCEMENT,
                channels=channels,
                data=announcement_data,
                priority=NotificationPriority.HIGH,
                template_data=template_data
            )
        else:
            # TODO: 发送给所有在线用户
            result = {
                'message': '系统公告功能需要配合用户管理系统实现全员推送'
            }
        
        logger.info(f"System announcement sent by admin {current_user.get('user_id')}")
        
        return success_response(
            data=result,
            message="系统公告发送成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"System announcement error: {str(e)}")
        return error_response(message=f"发送系统公告失败: {str(e)}")