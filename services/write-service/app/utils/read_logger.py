import json
import logging
from typing import Optional, Dict, Any
from fastapi import Request
from sqlalchemy.orm import Session
from app.models.read_log import ReadLog
from app.models.letter import Letter

logger = logging.getLogger(__name__)

class ReadLogManager:
    """阅读日志管理器"""
    
    @staticmethod
    def extract_device_info(request: Request) -> Dict[str, Any]:
        """
        从请求中提取设备信息
        
        Args:
            request: FastAPI请求对象
            
        Returns:
            Dict[str, Any]: 设备信息
        """
        user_agent = request.headers.get("user-agent", "")
        device_info = {
            "user_agent": user_agent,
            "accept": request.headers.get("accept", ""),
            "accept_language": request.headers.get("accept-language", ""),
            "accept_encoding": request.headers.get("accept-encoding", ""),
        }
        
        # 简单的设备类型判断
        user_agent_lower = user_agent.lower()
        if "mobile" in user_agent_lower or "android" in user_agent_lower or "iphone" in user_agent_lower:
            device_info["device_type"] = "mobile"
        elif "tablet" in user_agent_lower or "ipad" in user_agent_lower:
            device_info["device_type"] = "tablet"
        else:
            device_info["device_type"] = "desktop"
        
        # 浏览器检测
        if "chrome" in user_agent_lower:
            device_info["browser"] = "chrome"
        elif "firefox" in user_agent_lower:
            device_info["browser"] = "firefox"
        elif "safari" in user_agent_lower:
            device_info["browser"] = "safari"
        elif "edge" in user_agent_lower:
            device_info["browser"] = "edge"
        else:
            device_info["browser"] = "unknown"
        
        return device_info
    
    @staticmethod
    def get_client_ip(request: Request) -> str:
        """
        获取客户端真实IP地址
        
        Args:
            request: FastAPI请求对象
            
        Returns:
            str: 客户端IP地址
        """
        # 检查代理头部
        forwarded_for = request.headers.get("x-forwarded-for")
        if forwarded_for:
            # 取第一个IP（最原始的客户端IP）
            return forwarded_for.split(",")[0].strip()
        
        real_ip = request.headers.get("x-real-ip")
        if real_ip:
            return real_ip
        
        # 返回直连IP
        return request.client.host if request.client else "unknown"
    
    @classmethod
    async def log_read_event(
        cls,
        db: Session,
        letter_id: str,
        request: Request,
        read_duration: Optional[int] = None,
        is_complete_read: bool = True,
        reader_location: Optional[str] = None
    ) -> ReadLog:
        """
        记录阅读事件
        
        Args:
            db: 数据库会话
            letter_id: 信件ID
            request: 请求对象
            read_duration: 阅读时长（秒）
            is_complete_read: 是否完整阅读
            reader_location: 阅读地点
            
        Returns:
            ReadLog: 创建的阅读日志记录
        """
        try:
            # 提取请求信息
            client_ip = cls.get_client_ip(request)
            device_info = cls.extract_device_info(request)
            user_agent = request.headers.get("user-agent", "")
            referer = request.headers.get("referer", "")
            
            # 创建阅读日志
            read_log = ReadLog(
                letter_id=letter_id,
                reader_ip=client_ip,
                reader_user_agent=user_agent,
                reader_location=reader_location,
                read_duration=read_duration,
                is_complete_read=is_complete_read,
                referer=referer,
                device_info=json.dumps(device_info, ensure_ascii=False)
            )
            
            db.add(read_log)
            db.commit()
            db.refresh(read_log)
            
            logger.info(f"Logged read event for letter {letter_id} from IP {client_ip}")
            return read_log
            
        except Exception as e:
            logger.error(f"Failed to log read event for letter {letter_id}: {e}")
            db.rollback()
            raise
    
    @classmethod
    def get_letter_read_stats(cls, db: Session, letter_id: str) -> Dict[str, Any]:
        """
        获取信件阅读统计
        
        Args:
            db: 数据库会话
            letter_id: 信件ID
            
        Returns:
            Dict[str, Any]: 阅读统计信息
        """
        # 获取信件
        letter = db.query(Letter).filter(Letter.id == letter_id).first()
        if not letter:
            return {}
        
        # 获取所有阅读日志
        read_logs = db.query(ReadLog).filter(ReadLog.letter_id == letter_id).all()
        
        if not read_logs:
            return {
                "total_reads": 0,
                "unique_ips": 0,
                "device_types": {},
                "browsers": {},
                "average_duration": 0,
                "complete_reads": 0
            }
        
        # 统计分析
        unique_ips = set()
        device_types = {}
        browsers = {}
        total_duration = 0
        complete_reads = 0
        
        for log in read_logs:
            # 统计唯一IP
            unique_ips.add(log.reader_ip)
            
            # 统计完整阅读
            if log.is_complete_read:
                complete_reads += 1
            
            # 统计阅读时长
            if log.read_duration:
                total_duration += log.read_duration
            
            # 统计设备信息
            if log.device_info:
                try:
                    device_data = json.loads(log.device_info)
                    device_type = device_data.get("device_type", "unknown")
                    browser = device_data.get("browser", "unknown")
                    
                    device_types[device_type] = device_types.get(device_type, 0) + 1
                    browsers[browser] = browsers.get(browser, 0) + 1
                except:
                    pass
        
        return {
            "total_reads": len(read_logs),
            "unique_ips": len(unique_ips),
            "device_types": device_types,
            "browsers": browsers,
            "average_duration": total_duration / len(read_logs) if read_logs else 0,
            "complete_reads": complete_reads,
            "completion_rate": complete_reads / len(read_logs) if read_logs else 0
        }

# 便捷函数
async def log_letter_read(
    db: Session,
    letter_id: str,
    request: Request,
    read_duration: Optional[int] = None,
    is_complete_read: bool = True
) -> ReadLog:
    """记录信件阅读的便捷函数"""
    return await ReadLogManager.log_read_event(
        db, letter_id, request, read_duration, is_complete_read
    )