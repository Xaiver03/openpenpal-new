"""
认证和权限管理模块
"""
from typing import Optional, Dict, List
from fastapi import HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
import jwt
from datetime import datetime, timedelta
import hashlib
from app.core.config import settings
from app.core.logger import get_logger

logger = get_logger(__name__)
security = HTTPBearer()

class AuthManager:
    """认证管理器"""
    
    def __init__(self):
        self.secret_key = settings.jwt_secret
        self.algorithm = settings.jwt_algorithm
        self.access_token_expire_minutes = settings.jwt_access_token_expire_minutes

    def create_access_token(self, data: Dict, expires_delta: Optional[timedelta] = None) -> str:
        """创建访问令牌"""
        to_encode = data.copy()
        if expires_delta:
            expire = datetime.utcnow() + expires_delta
        else:
            expire = datetime.utcnow() + timedelta(minutes=self.access_token_expire_minutes)
        
        to_encode.update({"exp": expire})
        encoded_jwt = jwt.encode(to_encode, self.secret_key, algorithm=self.algorithm)
        return encoded_jwt

    def verify_token(self, token: str) -> Dict:
        """验证令牌"""
        try:
            payload = jwt.decode(token, self.secret_key, algorithms=[self.algorithm])
            return payload
        except jwt.ExpiredSignatureError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Token has expired",
                headers={"WWW-Authenticate": "Bearer"},
            )
        except jwt.JWTError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Could not validate credentials",
                headers={"WWW-Authenticate": "Bearer"},
            )

    def hash_password(self, password: str) -> str:
        """密码哈希"""
        return hashlib.sha256(password.encode()).hexdigest()

    def verify_password(self, plain_password: str, hashed_password: str) -> bool:
        """验证密码"""
        return self.hash_password(plain_password) == hashed_password

# 全局认证管理器实例
auth_manager = AuthManager()

def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)) -> Dict:
    """获取当前用户信息"""
    try:
        payload = auth_manager.verify_token(credentials.credentials)
        user_id = payload.get("user_id")
        if user_id is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid authentication credentials",
                headers={"WWW-Authenticate": "Bearer"},
            )
        
        # 模拟用户信息（实际应该从数据库获取）
        user = {
            "user_id": user_id,
            "username": payload.get("username"),
            "roles": payload.get("roles", []),
            "permissions": payload.get("permissions", [])
        }
        return user
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error getting current user: {e}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )

def get_current_user_info(credentials: HTTPAuthorizationCredentials = Depends(security)) -> Dict:
    """获取当前用户详细信息"""
    return get_current_user(credentials)

def get_current_user_optional(credentials: Optional[HTTPAuthorizationCredentials] = Depends(security)) -> Optional[Dict]:
    """获取当前用户信息（可选）"""
    if not credentials:
        return None
    try:
        return get_current_user(credentials)
    except HTTPException:
        return None

def check_admin_permission(current_user: Dict = Depends(get_current_user)) -> Dict:
    """检查管理员权限"""
    roles = current_user.get("roles", [])
    if not any(role in ["PLATFORM_SUPER_ADMIN", "PLATFORM_ADMIN", "SHOP_ADMIN"] for role in roles):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Admin permission required"
        )
    return current_user

def check_permission(required_permission: str):
    """权限检查装饰器"""
    def permission_checker(current_user: Dict = Depends(get_current_user)) -> Dict:
        permissions = current_user.get("permissions", [])
        if required_permission not in permissions:
            # 检查是否是超级管理员
            roles = current_user.get("roles", [])
            if "PLATFORM_SUPER_ADMIN" not in roles:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail=f"Permission '{required_permission}' required"
                )
        return current_user
    return permission_checker

def check_role(required_role: str):
    """角色检查装饰器"""
    def role_checker(current_user: Dict = Depends(get_current_user)) -> Dict:
        roles = current_user.get("roles", [])
        if required_role not in roles:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Role '{required_role}' required"
            )
        return current_user
    return role_checker

# 创建测试用户Token的辅助函数
def create_test_token(
    user_id: str = "TEST_USER", 
    username: str = "test_admin",
    roles: List[str] = None,
    permissions: List[str] = None
) -> str:
    """创建测试Token"""
    if roles is None:
        roles = ["PLATFORM_SUPER_ADMIN"]
    if permissions is None:
        permissions = [
            "platform:system:user:list",
            "platform:system:role:list", 
            "platform:category:list",
            "shop:product:list"
        ]
    
    token_data = {
        "user_id": user_id,
        "username": username,
        "roles": roles,
        "permissions": permissions
    }
    
    return auth_manager.create_access_token(token_data)