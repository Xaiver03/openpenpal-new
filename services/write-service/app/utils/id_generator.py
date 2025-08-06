"""ID生成器工具"""
import uuid
import secrets
import string
from typing import Optional

def generate_id(prefix: Optional[str] = None, length: int = 12) -> str:
    """
    生成唯一ID
    
    Args:
        prefix: 前缀（可选）
        length: 长度（不包含前缀）
        
    Returns:
        str: 生成的ID
    """
    chars = string.ascii_uppercase + string.digits
    # 排除容易混淆的字符
    excluded = {'0', 'O', 'I', '1', 'L'}
    safe_chars = ''.join(c for c in chars if c not in excluded)
    
    random_part = ''.join(secrets.choice(safe_chars) for _ in range(length))
    
    if prefix:
        return f"{prefix}{random_part}"
    return random_part

def generate_uuid() -> str:
    """
    生成标准UUID4
    
    Returns:
        str: UUID字符串
    """
    return str(uuid.uuid4())

def generate_short_id(length: int = 8) -> str:
    """
    生成短ID（只包含数字和大写字母）
    
    Args:
        length: ID长度
        
    Returns:
        str: 短ID
    """
    return generate_id(length=length)