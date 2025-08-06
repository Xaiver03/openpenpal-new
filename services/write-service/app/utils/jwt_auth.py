import jwt
from datetime import datetime, timedelta
from typing import Optional, Dict, Any
from fastapi import HTTPException, status
from app.core.config import settings

class JWTAuth:
    """JWT认证工具类"""
    
    @staticmethod
    def create_access_token(data: Dict[str, Any], expires_delta: Optional[timedelta] = None) -> str:
        """
        创建访问令牌
        
        Args:
            data: 要编码的数据
            expires_delta: 过期时间间隔
            
        Returns:
            str: JWT token
        """
        to_encode = data.copy()
        
        if expires_delta:
            expire = datetime.utcnow() + expires_delta
        else:
            # 使用配置的过期时间，默认30分钟以提高安全性
            expire = datetime.utcnow() + timedelta(minutes=settings.jwt_access_token_expire_minutes)
        
        to_encode.update({"exp": expire})
        encoded_jwt = jwt.encode(to_encode, settings.jwt_secret, algorithm=settings.jwt_algorithm)
        return encoded_jwt
    
    @staticmethod
    def verify_token(token: str) -> Dict[str, Any]:
        """
        验证JWT令牌
        
        Args:
            token: JWT token
            
        Returns:
            Dict[str, Any]: 解码后的payload
            
        Raises:
            HTTPException: 令牌无效或过期
        """
        try:
            payload = jwt.decode(token, settings.jwt_secret, algorithms=[settings.jwt_algorithm])
            return payload
        except jwt.ExpiredSignatureError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="令牌已过期",
                headers={"WWW-Authenticate": "Bearer"},
            )
        except jwt.JWTError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="无效的令牌",
                headers={"WWW-Authenticate": "Bearer"},
            )
    
    @staticmethod
    def extract_user_id(token: str) -> str:
        """
        从JWT令牌中提取用户ID
        
        Args:
            token: JWT token
            
        Returns:
            str: 用户ID
        """
        payload = JWTAuth.verify_token(token)
        user_id = payload.get("sub")
        
        if not user_id:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="令牌中缺少用户信息",
                headers={"WWW-Authenticate": "Bearer"},
            )
        
        return user_id
    
    @staticmethod
    def extract_user_info(token: str) -> Dict[str, Any]:
        """
        从JWT令牌中提取用户信息
        
        Args:
            token: JWT token
            
        Returns:
            Dict[str, Any]: 用户信息
        """
        payload = JWTAuth.verify_token(token)
        
        user_info = {
            "user_id": payload.get("sub"),
            "username": payload.get("username"),
            "nickname": payload.get("nickname"),
            "role": payload.get("role", "user"),
            "school_code": payload.get("school_code"),
            "exp": payload.get("exp")
        }
        
        if not user_info["user_id"]:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="令牌中缺少用户信息",
                headers={"WWW-Authenticate": "Bearer"},
            )
        
        return user_info