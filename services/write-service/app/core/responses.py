"""
API响应工具模块
"""
from typing import Any, Optional, Dict
from fastapi import HTTPException
from fastapi.responses import JSONResponse


def success_response(
    data: Any = None,
    message: str = "Success",
    code: int = 200,
    meta: Optional[Dict] = None
) -> Dict[str, Any]:
    """
    成功响应格式
    
    Args:
        data: 响应数据
        message: 响应消息
        code: 状态码
        meta: 元数据信息
        
    Returns:
        Dict: 格式化的响应数据
    """
    response = {
        "success": True,
        "code": code,
        "message": message,
        "data": data
    }
    
    if meta:
        response["meta"] = meta
        
    return response


def error_response(
    message: str = "Error",
    code: int = 400,
    error_code: Optional[str] = None,
    data: Any = None
) -> Dict[str, Any]:
    """
    错误响应格式
    
    Args:
        message: 错误消息
        code: 状态码
        error_code: 业务错误码
        data: 附加数据
        
    Returns:
        Dict: 格式化的错误响应数据
    """
    response = {
        "success": False,
        "code": code,
        "message": message,
        "data": data
    }
    
    if error_code:
        response["error_code"] = error_code
        
    return response


def paginated_response(
    items: list,
    total: int,
    page: int,
    size: int,
    message: str = "Success"
) -> Dict[str, Any]:
    """
    分页响应格式
    
    Args:
        items: 数据项列表
        total: 总数量
        page: 当前页码
        size: 每页大小
        message: 响应消息
        
    Returns:
        Dict: 格式化的分页响应数据
    """
    total_pages = (total + size - 1) // size  # 向上取整
    
    return success_response(
        data=items,
        message=message,
        meta={
            "pagination": {
                "total": total,
                "page": page,
                "size": size,
                "total_pages": total_pages,
                "has_next": page < total_pages,
                "has_prev": page > 1
            }
        }
    )


class APIException(HTTPException):
    """API异常基类"""
    
    def __init__(
        self,
        status_code: int,
        message: str,
        error_code: Optional[str] = None,
        data: Any = None
    ):
        self.message = message
        self.error_code = error_code
        self.data = data
        super().__init__(status_code=status_code, detail=message)


def create_error_response(
    status_code: int,
    message: str,
    error_code: Optional[str] = None,
    data: Any = None
) -> JSONResponse:
    """
    创建JSON错误响应
    
    Args:
        status_code: HTTP状态码
        message: 错误消息
        error_code: 业务错误码
        data: 附加数据
        
    Returns:
        JSONResponse: JSON错误响应
    """
    content = error_response(
        message=message,
        code=status_code,
        error_code=error_code,
        data=data
    )
    
    return JSONResponse(
        status_code=status_code,
        content=content
    )