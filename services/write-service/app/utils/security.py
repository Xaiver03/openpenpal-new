import secrets
import string
import hashlib
import os
from typing import Optional

class SecurityManager:
    """安全管理器 - 负责密钥生成和安全配置"""
    
    @staticmethod
    def generate_jwt_secret(length: int = 64) -> str:
        """
        生成强随机JWT密钥
        
        Args:
            length: 密钥长度，建议至少32字符
            
        Returns:
            str: 强随机密钥
        """
        # 使用密码学安全的随机数生成器
        alphabet = string.ascii_letters + string.digits + "!@#$%^&*"
        return ''.join(secrets.choice(alphabet) for _ in range(length))
    
    @staticmethod
    def generate_secure_key(length: int = 32) -> str:
        """
        生成十六进制格式的安全密钥
        
        Args:
            length: 密钥字节长度
            
        Returns:
            str: 十六进制密钥字符串
        """
        return secrets.token_hex(length)
    
    @staticmethod
    def validate_jwt_secret(secret: str) -> bool:
        """
        验证JWT密钥强度
        
        Args:
            secret: 待验证的密钥
            
        Returns:
            bool: 是否满足安全要求
        """
        if len(secret) < 32:
            return False
        
        # 检查是否包含多种字符类型
        has_upper = any(c.isupper() for c in secret)
        has_lower = any(c.islower() for c in secret)
        has_digit = any(c.isdigit() for c in secret)
        has_special = any(c in "!@#$%^&*" for c in secret)
        
        # 检查是否为常见弱密钥
        weak_patterns = [
            "your-super-secret-jwt-key",
            "development-jwt-secret-key", 
            "secret",
            "password",
            "12345",
            "admin"
        ]
        
        if any(pattern in secret.lower() for pattern in weak_patterns):
            return False
        
        # 至少包含3种字符类型
        char_types = sum([has_upper, has_lower, has_digit, has_special])
        return char_types >= 3
    
    @staticmethod
    def hash_password(password: str, salt: Optional[str] = None) -> tuple[str, str]:
        """
        使用PBKDF2算法哈希密码
        
        Args:
            password: 原始密码
            salt: 盐值，如果为None则自动生成
            
        Returns:
            tuple: (hashed_password, salt)
        """
        if salt is None:
            salt = secrets.token_hex(16)
        
        # 使用PBKDF2进行100000次迭代
        hashed = hashlib.pbkdf2_hmac(
            'sha256', 
            password.encode('utf-8'), 
            salt.encode('utf-8'), 
            100000
        )
        
        return hashed.hex(), salt
    
    @staticmethod
    def verify_password(password: str, hashed_password: str, salt: str) -> bool:
        """
        验证密码
        
        Args:
            password: 待验证的密码
            hashed_password: 存储的哈希密码
            salt: 盐值
            
        Returns:
            bool: 密码是否正确
        """
        test_hash, _ = SecurityManager.hash_password(password, salt)
        return test_hash == hashed_password
    
    @staticmethod
    def get_environment_jwt_secret() -> str:
        """
        从环境变量获取JWT密钥，如果不安全则生成新的
        
        Returns:
            str: 安全的JWT密钥
        """
        secret = os.getenv("JWT_SECRET")
        
        if not secret or not SecurityManager.validate_jwt_secret(secret):
            print("⚠️  Warning: JWT_SECRET is not set or is weak, generating a secure one...")
            new_secret = SecurityManager.generate_jwt_secret()
            print(f"✅ Generated secure JWT secret. Please set JWT_SECRET environment variable to:")
            print(f"   JWT_SECRET={new_secret}")
            return new_secret
        
        return secret

def generate_env_template():
    """生成包含安全密钥的环境变量模板"""
    jwt_secret = SecurityManager.generate_jwt_secret()
    db_password = SecurityManager.generate_secure_key(16)
    redis_password = SecurityManager.generate_secure_key(16)
    
    template = f"""# OpenPenPal Write Service 环境变量配置

# JWT 配置 (重要: 生产环境必须使用强密钥)
JWT_SECRET={jwt_secret}
JWT_ALGORITHM=HS256

# 数据库配置
DATABASE_URL=postgresql://openpenpal:{db_password}@localhost:5432/openpenpal

# Redis 配置
REDIS_URL=redis://:{redis_password}@localhost:6379/0

# WebSocket 配置
WEBSOCKET_URL=ws://localhost:8080/ws

# 服务配置
FRONTEND_URL=http://localhost:3000
USER_SERVICE_URL=http://localhost:8080/api/v1

# 安全配置
ENABLE_RATE_LIMITING=true
MAX_REQUESTS_PER_MINUTE=60
ENABLE_HTTPS=false

# 日志配置
LOG_LEVEL=INFO
"""
    
    return template

if __name__ == "__main__":
    # 生成安全配置文件
    print("🔐 生成安全的环境变量配置...")
    template = generate_env_template()
    
    with open(".env.example", "w", encoding="utf-8") as f:
        f.write(template)
    
    print("✅ 已生成 .env.example 文件")
    print("📝 请复制为 .env 文件并根据实际环境调整配置")