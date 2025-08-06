"""
日志配置模块
"""
import logging
import sys
from typing import Optional
from app.core.config import settings


def setup_logger(
    name: str = __name__,
    level: Optional[str] = None,
    format_string: Optional[str] = None
) -> logging.Logger:
    """
    设置并返回配置好的日志记录器
    
    Args:
        name: 日志记录器名称
        level: 日志级别
        format_string: 自定义格式字符串
        
    Returns:
        logging.Logger: 配置好的日志记录器
    """
    # 创建日志记录器
    logger = logging.getLogger(name)
    
    # 设置日志级别
    log_level = level or settings.log_level
    logger.setLevel(getattr(logging, log_level.upper(), logging.INFO))
    
    # 避免重复添加处理器
    if not logger.handlers:
        # 创建控制台处理器
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(getattr(logging, log_level.upper(), logging.INFO))
        
        # 设置日志格式
        default_format = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        formatter = logging.Formatter(format_string or default_format)
        console_handler.setFormatter(formatter)
        
        # 添加处理器到日志记录器
        logger.addHandler(console_handler)
    
    return logger


# 创建默认的日志记录器
logger = setup_logger("write-service")


def get_logger(name: str) -> logging.Logger:
    """
    获取指定名称的日志记录器
    
    Args:
        name: 日志记录器名称
        
    Returns:
        logging.Logger: 日志记录器实例
    """
    return setup_logger(name)


class LoggerMixin:
    """日志记录器混入类"""
    
    @property
    def logger(self) -> logging.Logger:
        """获取当前类的日志记录器"""
        return get_logger(self.__class__.__name__)


# 预定义的日志记录器
auth_logger = get_logger("auth")
database_logger = get_logger("database")
api_logger = get_logger("api")
service_logger = get_logger("service")
cache_logger = get_logger("cache")
websocket_logger = get_logger("websocket")