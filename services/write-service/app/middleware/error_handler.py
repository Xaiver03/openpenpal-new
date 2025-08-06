import traceback
from typing import Dict, Any
from fastapi import Request, HTTPException, status
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware
import logging

logger = logging.getLogger(__name__)

class ErrorHandlerMiddleware(BaseHTTPMiddleware):
    """错误处理中间件 - 清理敏感信息避免泄露"""
    
    def __init__(self, app, debug: bool = False):
        super().__init__(app)
        self.debug = debug
        
        # 敏感信息关键词
        self.sensitive_keywords = {
            'password', 'secret', 'key', 'token', 'jwt', 'auth',
            'database', 'connection', 'config', 'env', 'environment',
            'file', 'path', 'directory', 'server', 'host', 'port'
        }
    
    def _clean_error_message(self, message: str) -> str:
        """清理错误信息中的敏感内容"""
        if not message:
            return "系统内部错误"
        
        message_lower = message.lower()
        
        # 检查是否包含敏感关键词
        for keyword in self.sensitive_keywords:
            if keyword in message_lower:
                return "系统内部错误，请联系管理员"
        
        # 过滤文件路径信息
        if ('/' in message or '\\' in message) and ('file' in message_lower or 'path' in message_lower):
            return "文件处理错误"
        
        # 过滤数据库相关错误
        if any(db_keyword in message_lower for db_keyword in ['connection', 'database', 'sql', 'postgresql']):
            return "数据访问错误"
        
        # 过滤网络相关错误
        if any(net_keyword in message_lower for net_keyword in ['connection', 'timeout', 'refused', 'unreachable']):
            return "网络连接错误"
        
        # 如果非调试模式，进一步简化错误信息
        if not self.debug:
            if len(message) > 100:
                return "系统处理错误"
        
        return message
    
    def _create_error_response(self, error_code: int, message: str, details: Dict[str, Any] = None) -> JSONResponse:
        """创建标准化错误响应"""
        cleaned_message = self._clean_error_message(message)
        
        response_data = {
            "code": error_code,
            "msg": cleaned_message,
            "data": None
        }
        
        # 调试模式下可以包含更多详情
        if self.debug and details:
            response_data["debug_info"] = details
        
        return JSONResponse(
            status_code=error_code,
            content=response_data
        )
    
    async def dispatch(self, request: Request, call_next):
        """处理所有HTTP请求和异常"""
        try:
            response = await call_next(request)
            return response
            
        except HTTPException as exc:
            # FastAPI HTTPException
            return self._create_error_response(
                error_code=exc.status_code,
                message=exc.detail,
                details={"type": "HTTPException"} if self.debug else None
            )
            
        except ValueError as exc:
            # 值错误（通常是输入验证问题）
            logger.warning(f"ValueError in {request.url.path}: {str(exc)}")
            return self._create_error_response(
                error_code=status.HTTP_400_BAD_REQUEST,
                message="请求参数错误",
                details={"type": "ValueError", "original": str(exc)} if self.debug else None
            )
            
        except KeyError as exc:
            # 缺少必要参数
            logger.warning(f"KeyError in {request.url.path}: {str(exc)}")
            return self._create_error_response(
                error_code=status.HTTP_400_BAD_REQUEST,
                message="缺少必要参数",
                details={"type": "KeyError", "missing_key": str(exc)} if self.debug else None
            )
            
        except ConnectionError as exc:
            # 连接错误（数据库、Redis等）
            logger.error(f"ConnectionError in {request.url.path}: {str(exc)}")
            return self._create_error_response(
                error_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                message="服务暂时不可用，请稍后重试",
                details={"type": "ConnectionError"} if self.debug else None
            )
            
        except TimeoutError as exc:
            # 超时错误
            logger.error(f"TimeoutError in {request.url.path}: {str(exc)}")
            return self._create_error_response(
                error_code=status.HTTP_408_REQUEST_TIMEOUT,
                message="请求超时，请稍后重试",
                details={"type": "TimeoutError"} if self.debug else None
            )
            
        except PermissionError as exc:
            # 权限错误
            logger.warning(f"PermissionError in {request.url.path}: {str(exc)}")
            return self._create_error_response(
                error_code=status.HTTP_403_FORBIDDEN,
                message="权限不足",
                details={"type": "PermissionError"} if self.debug else None
            )
            
        except Exception as exc:
            # 所有其他未捕获的异常
            logger.error(f"Unhandled exception in {request.url.path}: {str(exc)}")
            
            # 记录详细错误到日志（仅在服务器端）
            if self.debug:
                logger.error(f"Exception traceback: {traceback.format_exc()}")
            
            return self._create_error_response(
                error_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="系统内部错误",
                details={
                    "type": type(exc).__name__,
                    "traceback": traceback.format_exc() if self.debug else None
                } if self.debug else None
            )

class InputSanitizer:
    """输入内容清理器 - XSS防护"""
    
    @staticmethod
    def sanitize_html(text: str) -> str:
        """清理HTML内容，防止XSS攻击"""
        if not text:
            return text
        
        # 简单的XSS防护 - 转义危险字符
        dangerous_chars = {
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#x27;',
            '&': '&amp;',
            '/': '&#x2F;'
        }
        
        sanitized = text
        for char, escape in dangerous_chars.items():
            sanitized = sanitized.replace(char, escape)
        
        return sanitized
    
    @staticmethod
    def sanitize_sql(text: str) -> str:
        """基础SQL注入防护"""
        if not text:
            return text
        
        # 检测常见SQL注入模式
        dangerous_patterns = [
            'union', 'select', 'insert', 'update', 'delete', 'drop',
            'exec', 'execute', 'sp_', 'xp_', '--', ';', '/*', '*/',
            'script', 'javascript:', 'vbscript:', 'onload', 'onerror'
        ]
        
        text_lower = text.lower()
        for pattern in dangerous_patterns:
            if pattern in text_lower:
                # 记录可疑输入
                logger.warning(f"Potential SQL injection detected: {pattern}")
                # 返回清理后的内容
                return text.replace(pattern, f"[FILTERED:{pattern.upper()}]")
        
        return text
    
    @staticmethod
    def validate_input_length(text: str, max_length: int = 10000) -> str:
        """验证输入长度"""
        if not text:
            return text
        
        if len(text) > max_length:
            raise ValueError(f"输入内容过长，最大允许{max_length}字符")
        
        return text
    
    @classmethod
    def clean_user_input(cls, text: str, max_length: int = 10000) -> str:
        """综合清理用户输入"""
        if not text:
            return text
        
        # 1. 验证长度
        text = cls.validate_input_length(text, max_length)
        
        # 2. SQL注入防护
        text = cls.sanitize_sql(text)
        
        # 3. XSS防护
        text = cls.sanitize_html(text)
        
        return text

def create_error_handler(debug: bool = False) -> ErrorHandlerMiddleware:
    """创建错误处理中间件"""
    return ErrorHandlerMiddleware(app=None, debug=debug)