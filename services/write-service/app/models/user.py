"""
用户模型（基础版本）
注意：实际的用户管理在独立的用户服务中，这里只是为了类型提示
"""
from typing import Optional
from pydantic import BaseModel


class User(BaseModel):
    """用户模型"""
    id: str
    username: Optional[str] = None
    email: Optional[str] = None
    nickname: Optional[str] = None
    is_active: bool = True
    is_admin: bool = False
    
    class Config:
        from_attributes = True