"""
认证工具模块 - 从app.core.auth导入主要功能
"""

# 从core.auth模块导入主要功能，保持向后兼容
from app.core.auth import (
    get_current_user,
    get_current_user_info, 
    get_current_user_optional,
    check_admin_permission,
    check_permission,
    check_role,
    auth_manager,
    create_test_token
)

# 为了保持API兼容性，创建一些别名
def get_current_active_user(current_user = None):
    """获取当前活跃用户（兼容性函数）"""
    if current_user is None:
        from fastapi import Depends
        from app.core.auth import get_current_user
        return Depends(get_current_user)
    return current_user

def verify_password(plain_password: str, hashed_password: str) -> bool:
    """验证密码"""
    return auth_manager.verify_password(plain_password, hashed_password)

def create_access_token(data: dict, expires_delta=None):
    """创建访问令牌"""
    return auth_manager.create_access_token(data, expires_delta)

# 导出所有需要的函数
__all__ = [
    'get_current_user',
    'get_current_user_info',
    'get_current_user_optional', 
    'get_current_active_user',
    'check_admin_permission',
    'check_permission',
    'check_role',
    'verify_password',
    'create_access_token',
    'create_test_token',
    'auth_manager'
]