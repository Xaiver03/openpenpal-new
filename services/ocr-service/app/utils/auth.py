import jwt
from datetime import datetime
from functools import wraps
from flask import request, current_app
from typing import Dict, Optional

from app.utils.response import permission_error_response, error_response


def decode_jwt_token(token: str) -> Optional[Dict]:
    """解码JWT token"""
    try:
        # 移除 'Bearer ' 前缀
        if token.startswith('Bearer '):
            token = token[7:]
        
        payload = jwt.decode(
            token,
            current_app.config['JWT_SECRET'],
            algorithms=['HS256']
        )
        
        # 检查token是否过期
        if 'exp' in payload and payload['exp'] < datetime.utcnow().timestamp():
            return None
            
        return payload
        
    except jwt.ExpiredSignatureError:
        return None
    except jwt.InvalidTokenError:
        return None
    except Exception:
        return None


def get_current_user() -> Optional[Dict]:
    """从请求头中获取当前用户信息"""
    auth_header = request.headers.get('Authorization')
    
    if not auth_header:
        return None
    
    return decode_jwt_token(auth_header)


def jwt_required(f):
    """JWT认证装饰器"""
    @wraps(f)
    def decorated_function(*args, **kwargs):
        auth_header = request.headers.get('Authorization')
        
        if not auth_header:
            return permission_error_response("缺少认证信息"), 403
        
        user_info = decode_jwt_token(auth_header)
        
        if not user_info:
            return permission_error_response("无效的认证信息"), 403
        
        # 将用户信息传递给视图函数
        request.current_user = user_info
        
        return f(*args, **kwargs)
    
    return decorated_function


def admin_required(f):
    """管理员权限装饰器"""
    @wraps(f)
    def decorated_function(*args, **kwargs):
        # 先检查JWT认证
        auth_header = request.headers.get('Authorization')
        
        if not auth_header:
            return permission_error_response("缺少认证信息"), 403
        
        user_info = decode_jwt_token(auth_header)
        
        if not user_info:
            return permission_error_response("无效的认证信息"), 403
        
        # 检查管理员权限
        user_role = user_info.get('role', '')
        if user_role not in ['admin', 'super_admin']:
            return permission_error_response("需要管理员权限"), 403
        
        request.current_user = user_info
        
        return f(*args, **kwargs)
    
    return decorated_function


def validate_user_access(resource_user_id: str) -> bool:
    """验证用户是否有权限访问特定资源"""
    user_info = getattr(request, 'current_user', None)
    
    if not user_info:
        return False
    
    current_user_id = user_info.get('user_id')
    user_role = user_info.get('role', '')
    
    # 管理员可以访问所有资源
    if user_role in ['admin', 'super_admin']:
        return True
    
    # 普通用户只能访问自己的资源
    return current_user_id == resource_user_id