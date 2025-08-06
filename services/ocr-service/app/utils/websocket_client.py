import redis
import json
import logging
from datetime import datetime
from typing import Dict, Optional
from app.core.config import Config

logger = logging.getLogger(__name__)


class WebSocketNotifier:
    """WebSocket通知客户端"""
    
    def __init__(self, config: Config = None):
        self.config = config or Config()
        self.redis_client = None
        self.is_available = False
        self._connect()
    
    def _connect(self):
        """连接Redis用于WebSocket通信"""
        try:
            self.redis_client = redis.Redis(
                host=self.config.REDIS_HOST,
                port=self.config.REDIS_PORT,
                password=self.config.REDIS_PASSWORD if self.config.REDIS_PASSWORD else None,
                db=self.config.REDIS_DB,
                socket_connect_timeout=5,
                socket_timeout=5
            )
            
            # 测试连接
            self.redis_client.ping()
            self.is_available = True
            logger.info("WebSocket通知客户端连接成功")
            
        except Exception as e:
            logger.warning(f"WebSocket通知客户端连接失败: {str(e)}")
            self.is_available = False
    
    def push_ocr_progress(self, user_id: str, task_id: str, progress_data: Dict):
        """推送OCR识别进度"""
        if not self.is_available:
            logger.warning("WebSocket服务不可用，跳过进度推送")
            return
        
        try:
            event = {
                "type": "OCR_PROGRESS_UPDATE",
                "data": {
                    "task_id": task_id,
                    "progress": progress_data.get('percentage', 0),
                    "status": progress_data.get('status', 'processing'),
                    "current_step": progress_data.get('step', ''),
                    "estimated_time_remaining": progress_data.get('eta', 0),
                    "message": progress_data.get('message', '')
                },
                "user_id": user_id,
                "timestamp": datetime.utcnow().isoformat() + "Z"
            }
            
            # 推送到用户专用频道
            channel = f"user:{user_id}:notifications"
            self.redis_client.publish(channel, json.dumps(event, ensure_ascii=False))
            
            # 同时推送到通用OCR频道（如果有全局监听）
            general_channel = "ocr:notifications"
            self.redis_client.publish(general_channel, json.dumps(event, ensure_ascii=False))
            
            logger.info(f"OCR进度推送成功: 用户{user_id}, 任务{task_id}, 进度{progress_data.get('percentage', 0)}%")
            
        except Exception as e:
            logger.error(f"OCR进度推送失败: {str(e)}")
    
    def push_ocr_completion(self, user_id: str, task_id: str, result_data: Dict):
        """推送OCR识别完成事件"""
        if not self.is_available:
            logger.warning("WebSocket服务不可用，跳过完成通知推送")
            return
        
        try:
            # 构建完成事件
            event = {
                "type": "OCR_TASK_COMPLETED",
                "data": {
                    "task_id": task_id,
                    "success": result_data.get('status') == 'completed',
                    "text_preview": self._get_text_preview(result_data.get('results', {}).get('text', '')),
                    "confidence": result_data.get('results', {}).get('confidence', 0),
                    "processing_time": result_data.get('results', {}).get('processing_time', 0),
                    "word_count": result_data.get('results', {}).get('word_count', 0),
                    "language": result_data.get('results', {}).get('language_detected', 'unknown'),
                    "engine": result_data.get('metadata', {}).get('processing_method', 'unknown'),
                    "from_cache": result_data.get('results', {}).get('from_cache', False)
                },
                "user_id": user_id,
                "timestamp": datetime.utcnow().isoformat() + "Z"
            }
            
            # 如果识别失败，添加错误信息
            if not event["data"]["success"]:
                event["data"]["error"] = result_data.get('error', '识别失败')
            
            # 推送到用户频道
            channel = f"user:{user_id}:notifications"
            self.redis_client.publish(channel, json.dumps(event, ensure_ascii=False))
            
            logger.info(f"OCR完成通知推送成功: 用户{user_id}, 任务{task_id}, 成功: {event['data']['success']}")
            
        except Exception as e:
            logger.error(f"OCR完成通知推送失败: {str(e)}")
    
    def push_batch_progress(self, user_id: str, batch_id: str, progress_data: Dict):
        """推送批量OCR进度"""
        if not self.is_available:
            logger.warning("WebSocket服务不可用，跳过批量进度推送")
            return
        
        try:
            event = {
                "type": "OCR_BATCH_PROGRESS",
                "data": {
                    "batch_id": batch_id,
                    "total_images": progress_data.get('total_images', 0),
                    "completed_images": progress_data.get('completed_images', 0),
                    "failed_images": progress_data.get('failed_images', 0),
                    "progress_percentage": progress_data.get('progress_percentage', 0),
                    "current_image": progress_data.get('current_image', ''),
                    "estimated_time_remaining": progress_data.get('eta', ''),
                    "status": progress_data.get('status', 'processing')
                },
                "user_id": user_id,
                "timestamp": datetime.utcnow().isoformat() + "Z"
            }
            
            # 推送到用户频道
            channel = f"user:{user_id}:notifications"
            self.redis_client.publish(channel, json.dumps(event, ensure_ascii=False))
            
            logger.info(f"批量OCR进度推送成功: 用户{user_id}, 批次{batch_id}, 进度{progress_data.get('progress_percentage', 0)}%")
            
        except Exception as e:
            logger.error(f"批量OCR进度推送失败: {str(e)}")
    
    def push_image_enhancement_progress(self, user_id: str, task_id: str, step: str, progress: int):
        """推送图像增强进度"""
        if not self.is_available:
            return
        
        try:
            event = {
                "type": "IMAGE_ENHANCEMENT_PROGRESS",
                "data": {
                    "task_id": task_id,
                    "step": step,
                    "progress": progress,
                    "message": f"正在执行: {step}"
                },
                "user_id": user_id,
                "timestamp": datetime.utcnow().isoformat() + "Z"
            }
            
            channel = f"user:{user_id}:notifications"
            self.redis_client.publish(channel, json.dumps(event, ensure_ascii=False))
            
        except Exception as e:
            logger.error(f"图像增强进度推送失败: {str(e)}")
    
    def push_system_notification(self, user_id: str, notification_type: str, message: str, data: Optional[Dict] = None):
        """推送系统通知"""
        if not self.is_available:
            return
        
        try:
            event = {
                "type": "SYSTEM_NOTIFICATION",
                "data": {
                    "notification_type": notification_type,  # info, warning, error, success
                    "message": message,
                    "additional_data": data or {}
                },
                "user_id": user_id,
                "timestamp": datetime.utcnow().isoformat() + "Z"
            }
            
            channel = f"user:{user_id}:notifications"
            self.redis_client.publish(channel, json.dumps(event, ensure_ascii=False))
            
            logger.info(f"系统通知推送成功: 用户{user_id}, 类型{notification_type}")
            
        except Exception as e:
            logger.error(f"系统通知推送失败: {str(e)}")
    
    def _get_text_preview(self, text: str, max_length: int = 100) -> str:
        """获取文本预览"""
        if not text:
            return ""
        
        # 清理文本
        cleaned_text = text.strip().replace('\n', ' ').replace('\r', ' ')
        
        # 截断并添加省略号
        if len(cleaned_text) > max_length:
            return cleaned_text[:max_length] + "..."
        
        return cleaned_text
    
    def test_connection(self) -> Dict:
        """测试WebSocket连接"""
        try:
            if not self.is_available:
                return {
                    "status": "disconnected",
                    "message": "Redis连接不可用"
                }
            
            # 测试发布功能
            test_channel = "test:notifications"
            test_message = {
                "type": "CONNECTION_TEST",
                "timestamp": datetime.utcnow().isoformat() + "Z"
            }
            
            result = self.redis_client.publish(test_channel, json.dumps(test_message))
            
            return {
                "status": "connected",
                "message": "WebSocket通知服务正常",
                "subscribers": result  # 订阅者数量
            }
            
        except Exception as e:
            logger.error(f"WebSocket连接测试失败: {str(e)}")
            return {
                "status": "error",
                "message": f"连接测试失败: {str(e)}"
            }
    
    def get_active_channels(self) -> list:
        """获取活跃的通知频道"""
        if not self.is_available:
            return []
        
        try:
            # 获取所有以user:开头的频道
            pubsub_channels = self.redis_client.pubsub_channels("user:*:notifications")
            return [channel.decode('utf-8') if isinstance(channel, bytes) else channel 
                   for channel in pubsub_channels]
            
        except Exception as e:
            logger.error(f"获取活跃频道失败: {str(e)}")
            return []


# 全局WebSocket通知器实例
_websocket_notifier = None

def get_websocket_notifier() -> WebSocketNotifier:
    """获取WebSocket通知器实例（单例模式）"""
    global _websocket_notifier
    if _websocket_notifier is None:
        _websocket_notifier = WebSocketNotifier()
    return _websocket_notifier