"""
业务异常定义模块
"""


class BusinessException(Exception):
    """业务逻辑异常基类"""
    
    def __init__(self, message: str, code: str = None, status_code: int = 400):
        self.message = message
        self.code = code
        self.status_code = status_code
        super().__init__(self.message)


class ValidationException(BusinessException):
    """数据验证异常"""
    
    def __init__(self, message: str, field: str = None):
        self.field = field
        super().__init__(message, "VALIDATION_ERROR", 400)


class PermissionException(BusinessException):
    """权限异常"""
    
    def __init__(self, message: str = "Permission denied"):
        super().__init__(message, "PERMISSION_DENIED", 403)


class ResourceNotFoundException(BusinessException):
    """资源未找到异常"""
    
    def __init__(self, message: str = "Resource not found", resource_type: str = None):
        self.resource_type = resource_type
        super().__init__(message, "RESOURCE_NOT_FOUND", 404)


class DuplicateResourceException(BusinessException):
    """资源重复异常"""
    
    def __init__(self, message: str = "Resource already exists", resource_type: str = None):
        self.resource_type = resource_type
        super().__init__(message, "DUPLICATE_RESOURCE", 409)


class ServiceException(BusinessException):
    """服务异常"""
    
    def __init__(self, message: str = "Service error", service_name: str = None):
        self.service_name = service_name
        super().__init__(message, "SERVICE_ERROR", 500)


class DatabaseException(BusinessException):
    """数据库异常"""
    
    def __init__(self, message: str = "Database error"):
        super().__init__(message, "DATABASE_ERROR", 500)


class ExternalServiceException(BusinessException):
    """外部服务异常"""
    
    def __init__(self, message: str = "External service error", service_name: str = None):
        self.service_name = service_name
        super().__init__(message, "EXTERNAL_SERVICE_ERROR", 502)


class RateLimitException(BusinessException):
    """频率限制异常"""
    
    def __init__(self, message: str = "Rate limit exceeded", retry_after: int = None):
        self.retry_after = retry_after
        super().__init__(message, "RATE_LIMIT_EXCEEDED", 429)