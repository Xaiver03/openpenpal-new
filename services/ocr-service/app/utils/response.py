from datetime import datetime
from typing import Any, Optional, Dict
from flask import jsonify


def success_response(data: Any = None, msg: str = "success", code: int = 0) -> Dict:
    """统一成功响应格式"""
    return {
        "code": code,
        "msg": msg,
        "data": data,
        "timestamp": datetime.utcnow().isoformat() + "Z"
    }


def error_response(code: int, msg: str, data: Any = None, error_details: Optional[Dict] = None) -> Dict:
    """统一错误响应格式"""
    response = {
        "code": code,
        "msg": msg,
        "data": data,
        "timestamp": datetime.utcnow().isoformat() + "Z"
    }
    
    if error_details:
        response["error"] = error_details
    
    return response


def validation_error_response(msg: str = "参数验证失败", fields: Optional[list] = None) -> Dict:
    """参数验证错误响应"""
    error_details = {
        "type": "validation_error",
        "details": msg
    }
    
    if fields:
        error_details["fields"] = fields
    
    return error_response(1, msg, error_details=error_details)


def permission_error_response(msg: str = "无权限访问") -> Dict:
    """权限错误响应"""
    return error_response(2, msg, error_details={
        "type": "permission_error",
        "details": msg
    })


def not_found_error_response(msg: str = "资源不存在") -> Dict:
    """资源不存在错误响应"""
    return error_response(3, msg, error_details={
        "type": "not_found_error",
        "details": msg
    })


def business_error_response(msg: str, details: str = None) -> Dict:
    """业务逻辑错误响应"""
    error_details = {
        "type": "business_error",
        "details": details or msg
    }
    
    return error_response(4, msg, error_details=error_details)


def internal_error_response(msg: str = "服务内部错误") -> Dict:
    """服务内部错误响应"""
    return error_response(500, msg, error_details={
        "type": "internal_error",
        "details": msg
    })