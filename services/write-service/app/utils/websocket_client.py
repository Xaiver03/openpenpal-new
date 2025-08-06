import asyncio
import json
import websockets
from typing import Dict, Any, Optional
from datetime import datetime
from app.core.config import settings
import logging

logger = logging.getLogger(__name__)

# 全局WebSocket管理器实例
_websocket_manager: Optional['WebSocketEventPublisher'] = None

def get_websocket_manager() -> 'WebSocketEventPublisher':
    """获取WebSocket管理器的单例实例"""
    global _websocket_manager
    if _websocket_manager is None:
        _websocket_manager = WebSocketEventPublisher()
    return _websocket_manager

class WebSocketEventPublisher:
    """WebSocket事件发布器 - 连接到主WebSocket服务"""
    
    def __init__(self):
        self.websocket_url = settings.websocket_url
        self.connection: Optional[websockets.WebSocketServerProtocol] = None
        self.is_connected = False
    
    async def connect(self) -> bool:
        """
        连接到WebSocket服务
        
        Returns:
            bool: 连接是否成功
        """
        try:
            self.connection = await websockets.connect(self.websocket_url)
            self.is_connected = True
            logger.info(f"Connected to WebSocket server: {self.websocket_url}")
            return True
        except Exception as e:
            logger.error(f"Failed to connect to WebSocket server: {e}")
            self.is_connected = False
            return False
    
    async def disconnect(self):
        """断开WebSocket连接"""
        if self.connection:
            await self.connection.close()
            self.is_connected = False
            logger.info("Disconnected from WebSocket server")
    
    async def send_event(self, event_type: str, data: Dict[str, Any], user_id: Optional[str] = None) -> bool:
        """
        发送事件到WebSocket服务
        
        Args:
            event_type: 事件类型
            data: 事件数据
            user_id: 目标用户ID（可选）
            
        Returns:
            bool: 发送是否成功
        """
        if not self.is_connected or not self.connection:
            logger.warning("WebSocket not connected, attempting to reconnect...")
            if not await self.connect():
                return False
        
        event = {
            "type": event_type,
            "data": data,
            "timestamp": datetime.utcnow().isoformat(),
            "source": "write-service"
        }
        
        if user_id:
            event["target_user"] = user_id
        
        try:
            await self.connection.send(json.dumps(event))
            logger.info(f"Sent WebSocket event: {event_type}")
            return True
        except Exception as e:
            logger.error(f"Failed to send WebSocket event: {e}")
            self.is_connected = False
            return False
    
    async def broadcast_letter_status_update(self, letter_id: str, status: str, sender_id: str, courier_id: Optional[str] = None):
        """
        广播信件状态更新事件
        
        Args:
            letter_id: 信件ID
            status: 新状态
            sender_id: 发送者ID
            courier_id: 信使ID（可选）
        """
        event_data = {
            "letter_id": letter_id,
            "status": status,
            "timestamp": datetime.utcnow().isoformat()
        }
        
        # 发送给发送者
        await self.send_event("LETTER_STATUS_UPDATE", event_data, sender_id)
        
        # 如果有信使，也发送给信使
        if courier_id:
            await self.send_event("LETTER_STATUS_UPDATE", event_data, courier_id)
    
    async def broadcast_letter_created(self, letter_id: str, sender_id: str):
        """
        广播信件创建事件
        
        Args:
            letter_id: 信件ID
            sender_id: 发送者ID
        """
        event_data = {
            "letter_id": letter_id,
            "action": "created",
            "timestamp": datetime.utcnow().isoformat()
        }
        
        await self.send_event("LETTER_CREATED", event_data, sender_id)
    
    async def broadcast_letter_read(self, letter_id: str, sender_id: str, read_count: int):
        """
        广播信件阅读事件
        
        Args:
            letter_id: 信件ID
            sender_id: 发送者ID
            read_count: 阅读次数
        """
        event_data = {
            "letter_id": letter_id,
            "action": "read",
            "read_count": read_count,
            "timestamp": datetime.utcnow().isoformat()
        }
        
        await self.send_event("LETTER_READ", event_data, sender_id)

# 全局WebSocket事件发布器实例
websocket_publisher = WebSocketEventPublisher()

async def init_websocket():
    """初始化WebSocket连接"""
    await websocket_publisher.connect()

async def cleanup_websocket():
    """清理WebSocket连接"""
    await websocket_publisher.disconnect()

# 便捷函数
async def notify_letter_status_update(letter_id: str, status: str, sender_id: str, courier_id: Optional[str] = None):
    """通知信件状态更新"""
    await websocket_publisher.broadcast_letter_status_update(letter_id, status, sender_id, courier_id)

async def notify_letter_created(letter_id: str, sender_id: str):
    """通知信件创建"""
    await websocket_publisher.broadcast_letter_created(letter_id, sender_id)

async def notify_letter_read(letter_id: str, sender_id: str, read_count: int):
    """通知信件被阅读"""
    await websocket_publisher.broadcast_letter_read(letter_id, sender_id, read_count)

async def notify_plaza_activity(activity_type: str, data: Dict[str, Any]):
    """通知广场活动"""
    await websocket_publisher.send_event(f"PLAZA_{activity_type.upper()}", data)

async def notify_batch_operation_progress(job_id: str, progress: int, status: str, user_id: str):
    """通知批量操作进度"""
    await websocket_publisher.send_event(
        "BATCH_OPERATION_PROGRESS",
        {
            "job_id": job_id,
            "progress": progress,
            "status": status
        },
        user_id
    )