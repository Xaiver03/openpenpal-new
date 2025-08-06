"""
增强的通知服务
支持WebSocket实时推送、邮件通知、多种通知类型
"""

import json
import asyncio
from typing import Dict, List, Optional, Any, Union
from enum import Enum
import aiohttp
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email import encoders
from datetime import datetime, timedelta
import uuid

from app.core.config import settings
from app.core.logger import get_logger
from app.utils.cache_manager import get_cache_manager
from app.utils.websocket_client import get_websocket_manager

logger = get_logger(__name__)


class NotificationType(str, Enum):
    """通知类型"""
    LETTER_RECEIVED = "letter_received"           # 收到新信件
    LETTER_DELIVERED = "letter_delivered"         # 信件已投递
    COMMENT_RECEIVED = "comment_received"         # 收到评论
    LIKE_RECEIVED = "like_received"               # 收到点赞
    ORDER_STATUS_CHANGED = "order_status_changed" # 订单状态变更
    SYSTEM_ANNOUNCEMENT = "system_announcement"   # 系统公告
    SECURITY_ALERT = "security_alert"            # 安全警告
    PROMOTION = "promotion"                      # 促销信息


class NotificationChannel(str, Enum):
    """通知渠道"""
    WEBSOCKET = "websocket"    # WebSocket实时推送
    EMAIL = "email"           # 邮件通知
    SMS = "sms"               # 短信通知（预留）
    PUSH = "push"             # 推送通知（预留）


class NotificationPriority(str, Enum):
    """通知优先级"""
    LOW = "low"         # 低优先级
    NORMAL = "normal"   # 普通优先级
    HIGH = "high"       # 高优先级
    URGENT = "urgent"   # 紧急优先级


class NotificationService:
    """通知服务"""
    
    def __init__(self):
        self.websocket_manager = None
        self.cache_manager = None
        
        # 邮件配置
        self.smtp_server = getattr(settings, 'SMTP_SERVER', 'smtp.gmail.com')
        self.smtp_port = getattr(settings, 'SMTP_PORT', 587)
        self.smtp_username = getattr(settings, 'SMTP_USERNAME', '')
        self.smtp_password = getattr(settings, 'SMTP_PASSWORD', '')
        self.smtp_from_email = getattr(settings, 'SMTP_FROM_EMAIL', '')
        
        # 推送通知配置
        self.push_service_url = getattr(settings, 'PUSH_SERVICE_URL', '')
        self.push_api_key = getattr(settings, 'PUSH_API_KEY', '')
        
        # 通知模板
        self.templates = {
            NotificationType.LETTER_RECEIVED: {
                'title': '您收到了一封新信件',
                'body': '来自 {sender_name} 的信件：{letter_title}',
                'email_template': 'letter_received.html'
            },
            NotificationType.LETTER_DELIVERED: {
                'title': '信件投递成功',
                'body': '您的信件《{letter_title}》已成功投递',
                'email_template': 'letter_delivered.html'
            },
            NotificationType.COMMENT_RECEIVED: {
                'title': '收到新评论',
                'body': '{commenter_name} 评论了您的帖子：{comment_content}',
                'email_template': 'comment_received.html'
            },
            NotificationType.ORDER_STATUS_CHANGED: {
                'title': '订单状态更新',
                'body': '您的订单 {order_id} 状态已更新为：{new_status}',
                'email_template': 'order_status_changed.html'
            },
            NotificationType.SYSTEM_ANNOUNCEMENT: {
                'title': '系统公告',
                'body': '{announcement_content}',
                'email_template': 'system_announcement.html'
            }
        }
    
    async def initialize(self):
        """初始化通知服务"""
        try:
            from app.utils.websocket_client import get_websocket_manager
            self.websocket_manager = await get_websocket_manager()
            self.cache_manager = await get_cache_manager()
            logger.info("Notification service initialized")
        except Exception as e:
            logger.error(f"Failed to initialize notification service: {str(e)}")
    
    async def send_notification(
        self,
        user_id: str,
        notification_type: NotificationType,
        channels: List[NotificationChannel],
        data: Dict[str, Any],
        priority: NotificationPriority = NotificationPriority.NORMAL,
        expires_at: Optional[datetime] = None,
        template_data: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        发送通知
        
        Args:
            user_id: 用户ID
            notification_type: 通知类型
            channels: 通知渠道列表
            data: 通知数据
            priority: 优先级
            expires_at: 过期时间
            template_data: 模板数据
            
        Returns:
            发送结果
        """
        try:
            # 生成通知ID
            notification_id = str(uuid.uuid4())
            
            # 获取用户偏好设置
            user_preferences = await self._get_user_notification_preferences(user_id)
            
            # 过滤渠道（根据用户偏好）
            filtered_channels = self._filter_channels_by_preferences(
                channels, notification_type, user_preferences
            )
            
            # 准备通知内容
            notification_content = await self._prepare_notification_content(
                notification_type, data, template_data or {}
            )
            
            # 创建通知记录
            notification_record = {
                'id': notification_id,
                'user_id': user_id,
                'type': notification_type,
                'priority': priority,
                'content': notification_content,
                'data': data,
                'channels': filtered_channels,
                'status': 'pending',
                'created_at': datetime.utcnow().isoformat(),
                'expires_at': expires_at.isoformat() if expires_at else None,
                'delivery_results': {}
            }
            
            # 保存通知记录
            await self._save_notification_record(notification_record)
            
            # 发送通知到各个渠道
            delivery_tasks = []
            for channel in filtered_channels:
                task = asyncio.create_task(
                    self._send_to_channel(channel, user_id, notification_record)
                )
                delivery_tasks.append(task)
            
            # 等待所有渠道发送完成
            if delivery_tasks:
                delivery_results = await asyncio.gather(*delivery_tasks, return_exceptions=True)
                
                # 更新发送结果
                for i, result in enumerate(delivery_results):
                    channel = filtered_channels[i]
                    if isinstance(result, Exception):
                        notification_record['delivery_results'][channel] = {
                            'success': False,
                            'error': str(result),
                            'sent_at': datetime.utcnow().isoformat()
                        }
                    else:
                        notification_record['delivery_results'][channel] = result
            
            # 更新通知状态
            notification_record['status'] = 'sent' if delivery_tasks else 'filtered'
            await self._update_notification_record(notification_record)
            
            logger.info(f"Notification sent: {notification_id} to user {user_id}")
            
            return {
                'notification_id': notification_id,
                'status': notification_record['status'],
                'channels_sent': len(filtered_channels),
                'delivery_results': notification_record['delivery_results']
            }
            
        except Exception as e:
            logger.error(f"Failed to send notification: {str(e)}")
            raise
    
    async def send_bulk_notification(
        self,
        user_ids: List[str],
        notification_type: NotificationType,
        channels: List[NotificationChannel],
        data: Dict[str, Any],
        priority: NotificationPriority = NotificationPriority.NORMAL,
        template_data: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """批量发送通知"""
        try:
            results = []
            
            # 批量发送，限制并发数
            semaphore = asyncio.Semaphore(10)  # 最大并发10
            
            async def send_single(user_id: str):
                async with semaphore:
                    return await self.send_notification(
                        user_id=user_id,
                        notification_type=notification_type,
                        channels=channels,
                        data=data,
                        priority=priority,
                        template_data=template_data
                    )
            
            tasks = [send_single(user_id) for user_id in user_ids]
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # 统计结果
            success_count = sum(1 for r in results if not isinstance(r, Exception))
            failed_count = len(results) - success_count
            
            logger.info(f"Bulk notification completed: {success_count} success, {failed_count} failed")
            
            return {
                'total_users': len(user_ids),
                'success_count': success_count,
                'failed_count': failed_count,
                'results': results
            }
            
        except Exception as e:
            logger.error(f"Bulk notification failed: {str(e)}")
            raise
    
    async def _send_to_channel(
        self, 
        channel: NotificationChannel, 
        user_id: str, 
        notification: Dict[str, Any]
    ) -> Dict[str, Any]:
        """发送到指定渠道"""
        try:
            if channel == NotificationChannel.WEBSOCKET:
                return await self._send_websocket_notification(user_id, notification)
            elif channel == NotificationChannel.EMAIL:
                return await self._send_email_notification(user_id, notification)
            elif channel == NotificationChannel.PUSH:
                return await self._send_push_notification(user_id, notification)
            else:
                return {
                    'success': False,
                    'error': f'Unsupported channel: {channel}',
                    'sent_at': datetime.utcnow().isoformat()
                }
                
        except Exception as e:
            return {
                'success': False,
                'error': str(e),
                'sent_at': datetime.utcnow().isoformat()
            }
    
    async def _send_websocket_notification(self, user_id: str, notification: Dict[str, Any]) -> Dict[str, Any]:
        """发送WebSocket通知"""
        try:
            if not self.websocket_manager:
                raise Exception("WebSocket manager not initialized")
            
            # 构造WebSocket消息
            ws_message = {
                'type': 'notification',
                'notification_id': notification['id'],
                'notification_type': notification['type'],
                'priority': notification['priority'],
                'title': notification['content']['title'],
                'body': notification['content']['body'],
                'data': notification['data'],
                'timestamp': notification['created_at']
            }
            
            # 发送到WebSocket
            success = await self.websocket_manager.send_to_user(user_id, ws_message)
            
            return {
                'success': success,
                'sent_at': datetime.utcnow().isoformat(),
                'channel': 'websocket'
            }
            
        except Exception as e:
            logger.error(f"WebSocket notification failed: {str(e)}")
            raise
    
    async def _send_email_notification(self, user_id: str, notification: Dict[str, Any]) -> Dict[str, Any]:
        """发送邮件通知"""
        try:
            # 获取用户邮箱
            user_email = await self._get_user_email(user_id)
            if not user_email:
                raise Exception("User email not found")
            
            # 创建邮件消息
            msg = MIMEMultipart('alternative')
            msg['Subject'] = notification['content']['title']
            msg['From'] = self.smtp_from_email
            msg['To'] = user_email
            
            # 添加纯文本内容
            text_content = notification['content']['body']
            text_part = MIMEText(text_content, 'plain', 'utf-8')
            msg.attach(text_part)
            
            # 添加HTML内容（如果有模板）
            html_content = await self._render_email_template(
                notification['type'], notification['data'], notification['content']
            )
            if html_content:
                html_part = MIMEText(html_content, 'html', 'utf-8')
                msg.attach(html_part)
            
            # 发送邮件
            async with asyncio.Lock():  # 防止SMTP连接冲突
                with smtplib.SMTP(self.smtp_server, self.smtp_port) as server:
                    server.starttls()
                    if self.smtp_username and self.smtp_password:
                        server.login(self.smtp_username, self.smtp_password)
                    
                    server.send_message(msg)
            
            return {
                'success': True,
                'sent_at': datetime.utcnow().isoformat(),
                'channel': 'email',
                'recipient': user_email
            }
            
        except Exception as e:
            logger.error(f"Email notification failed: {str(e)}")
            raise
    
    async def _send_push_notification(self, user_id: str, notification: Dict[str, Any]) -> Dict[str, Any]:
        """发送推送通知"""
        try:
            if not self.push_service_url or not self.push_api_key:
                raise Exception("Push service not configured")
            
            # 获取用户的推送Token
            push_tokens = await self._get_user_push_tokens(user_id)
            if not push_tokens:
                raise Exception("No push tokens found for user")
            
            # 构造推送消息
            push_payload = {
                'tokens': push_tokens,
                'notification': {
                    'title': notification['content']['title'],
                    'body': notification['content']['body'],
                    'data': notification['data']
                },
                'priority': notification['priority']
            }
            
            # 发送推送请求
            async with aiohttp.ClientSession() as session:
                headers = {
                    'Authorization': f'Bearer {self.push_api_key}',
                    'Content-Type': 'application/json'
                }
                
                async with session.post(
                    self.push_service_url,
                    json=push_payload,
                    headers=headers
                ) as response:
                    response.raise_for_status()
                    result = await response.json()
            
            return {
                'success': True,
                'sent_at': datetime.utcnow().isoformat(),
                'channel': 'push',
                'tokens_count': len(push_tokens),
                'response': result
            }
            
        except Exception as e:
            logger.error(f"Push notification failed: {str(e)}")
            raise
    
    async def _prepare_notification_content(
        self, 
        notification_type: NotificationType, 
        data: Dict[str, Any],
        template_data: Dict[str, Any]
    ) -> Dict[str, str]:
        """准备通知内容"""
        template = self.templates.get(notification_type, {
            'title': '通知',
            'body': '您有一条新通知'
        })
        
        # 合并数据
        merged_data = {**data, **template_data}
        
        # 格式化模板
        try:
            title = template['title'].format(**merged_data)
            body = template['body'].format(**merged_data)
        except KeyError as e:
            logger.warning(f"Template formatting failed: {str(e)}")
            title = template['title']
            body = template['body']
        
        return {
            'title': title,
            'body': body
        }
    
    async def _get_user_notification_preferences(self, user_id: str) -> Dict[str, Any]:
        """获取用户通知偏好设置"""
        try:
            if self.cache_manager:
                preferences = await self.cache_manager.get(f"user:preferences:{user_id}")
                if preferences:
                    return preferences
            
            # 默认偏好设置
            return {
                'websocket_enabled': True,
                'email_enabled': True,
                'push_enabled': True,
                'notification_types': {
                    NotificationType.LETTER_RECEIVED: True,
                    NotificationType.LETTER_DELIVERED: True,
                    NotificationType.COMMENT_RECEIVED: True,
                    NotificationType.ORDER_STATUS_CHANGED: True,
                    NotificationType.SYSTEM_ANNOUNCEMENT: True
                }
            }
            
        except Exception as e:
            logger.error(f"Failed to get user preferences: {str(e)}")
            return {}
    
    def _filter_channels_by_preferences(
        self, 
        channels: List[NotificationChannel], 
        notification_type: NotificationType,
        preferences: Dict[str, Any]
    ) -> List[NotificationChannel]:
        """根据用户偏好过滤通知渠道"""
        filtered_channels = []
        
        # 检查通知类型是否被启用
        type_preferences = preferences.get('notification_types', {})
        if not type_preferences.get(notification_type, True):
            return []
        
        for channel in channels:
            channel_enabled = preferences.get(f"{channel}_enabled", True)
            if channel_enabled:
                filtered_channels.append(channel)
        
        return filtered_channels
    
    async def _get_user_email(self, user_id: str) -> Optional[str]:
        """获取用户邮箱"""
        try:
            if self.cache_manager:
                user_info = await self.cache_manager.get(f"user:info:{user_id}")
                if user_info and 'email' in user_info:
                    return user_info['email']
            
            # TODO: 从数据库查询用户邮箱
            return None
            
        except Exception as e:
            logger.error(f"Failed to get user email: {str(e)}")
            return None
    
    async def _get_user_push_tokens(self, user_id: str) -> List[str]:
        """获取用户推送Token"""
        try:
            if self.cache_manager:
                tokens = await self.cache_manager.get(f"user:push_tokens:{user_id}")
                if tokens:
                    return tokens
            
            return []
            
        except Exception as e:
            logger.error(f"Failed to get push tokens: {str(e)}")
            return []
    
    async def _render_email_template(
        self, 
        notification_type: NotificationType, 
        data: Dict[str, Any],
        content: Dict[str, str]
    ) -> Optional[str]:
        """渲染邮件模板"""
        try:
            template_name = self.templates.get(notification_type, {}).get('email_template')
            if not template_name:
                return None
            
            # 简单的HTML模板
            html_template = f"""
            <!DOCTYPE html>
            <html>
            <head>
                <meta charset="utf-8">
                <title>{content['title']}</title>
                <style>
                    body {{ font-family: Arial, sans-serif; margin: 20px; }}
                    .container {{ max-width: 600px; margin: 0 auto; }}
                    .header {{ background-color: #f8f9fa; padding: 20px; border-radius: 5px; }}
                    .content {{ padding: 20px 0; }}
                    .footer {{ font-size: 12px; color: #6c757d; }}
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="header">
                        <h2>{content['title']}</h2>
                    </div>
                    <div class="content">
                        <p>{content['body']}</p>
                    </div>
                    <div class="footer">
                        <p>此邮件由 OpenPenPal 系统自动发送，请勿回复。</p>
                    </div>
                </div>
            </body>
            </html>
            """
            
            return html_template
            
        except Exception as e:
            logger.error(f"Failed to render email template: {str(e)}")
            return None
    
    async def _save_notification_record(self, notification: Dict[str, Any]) -> None:
        """保存通知记录"""
        try:
            if self.cache_manager:
                # 缓存通知记录（7天过期）
                await self.cache_manager.set(
                    f"notification:{notification['id']}", 
                    notification, 
                    expire=86400 * 7
                )
                
                # 添加到用户通知列表
                user_notifications_key = f"user:notifications:{notification['user_id']}"
                user_notifications = await self.cache_manager.get(user_notifications_key) or []
                user_notifications.insert(0, notification['id'])
                
                # 只保留最近100条
                if len(user_notifications) > 100:
                    user_notifications = user_notifications[:100]
                
                await self.cache_manager.set(user_notifications_key, user_notifications, expire=86400 * 7)
                
        except Exception as e:
            logger.error(f"Failed to save notification record: {str(e)}")
    
    async def _update_notification_record(self, notification: Dict[str, Any]) -> None:
        """更新通知记录"""
        try:
            if self.cache_manager:
                await self.cache_manager.set(
                    f"notification:{notification['id']}", 
                    notification, 
                    expire=86400 * 7
                )
        except Exception as e:
            logger.error(f"Failed to update notification record: {str(e)}")
    
    async def get_user_notifications(
        self, 
        user_id: str, 
        limit: int = 20, 
        offset: int = 0
    ) -> Dict[str, Any]:
        """获取用户通知列表"""
        try:
            if not self.cache_manager:
                return {'notifications': [], 'total': 0}
            
            # 获取用户通知ID列表
            user_notifications_key = f"user:notifications:{user_id}"
            notification_ids = await self.cache_manager.get(user_notifications_key) or []
            
            # 分页
            paginated_ids = notification_ids[offset:offset + limit]
            
            # 获取通知详情
            notifications = []
            for notification_id in paginated_ids:
                notification = await self.cache_manager.get(f"notification:{notification_id}")
                if notification:
                    notifications.append(notification)
            
            return {
                'notifications': notifications,
                'total': len(notification_ids),
                'limit': limit,
                'offset': offset
            }
            
        except Exception as e:
            logger.error(f"Failed to get user notifications: {str(e)}")
            return {'notifications': [], 'total': 0}


# 全局实例
_notification_service: Optional[NotificationService] = None

async def get_notification_service() -> NotificationService:
    """获取通知服务实例"""
    global _notification_service
    if _notification_service is None:
        _notification_service = NotificationService()
        await _notification_service.initialize()
    return _notification_service


# 便捷函数
async def send_letter_notification(user_id: str, letter_data: Dict[str, Any]):
    """发送信件通知"""
    service = await get_notification_service()
    await service.send_notification(
        user_id=user_id,
        notification_type=NotificationType.LETTER_RECEIVED,
        channels=[NotificationChannel.WEBSOCKET, NotificationChannel.EMAIL],
        data=letter_data,
        priority=NotificationPriority.NORMAL
    )

async def send_order_notification(user_id: str, order_data: Dict[str, Any]):
    """发送订单通知"""
    service = await get_notification_service()
    await service.send_notification(
        user_id=user_id,
        notification_type=NotificationType.ORDER_STATUS_CHANGED,
        channels=[NotificationChannel.WEBSOCKET],
        data=order_data,
        priority=NotificationPriority.HIGH
    )