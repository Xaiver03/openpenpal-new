import os
from typing import Optional
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # 服务配置
    app_name: str = "OpenPenPal Write Service"
    version: str = "1.0.0"
    host: str = "0.0.0.0"
    port: int = 8001
    
    # 数据库配置
    database_url: str = os.getenv("DATABASE_URL", "postgresql+psycopg2://user:password@localhost:5432/openpenpal")
    
    # JWT配置 - 使用安全的密钥管理
    jwt_secret: str = None  # 将在初始化时设置
    jwt_algorithm: str = "HS256"
    jwt_access_token_expire_minutes: int = int(os.getenv("JWT_ACCESS_TOKEN_EXPIRE_MINUTES", "30"))  # 缩短到30分钟
    
    # Redis配置 (可选)
    redis_url: Optional[str] = os.getenv("REDIS_URL", "redis://localhost:6379/0")
    
    # WebSocket配置
    websocket_url: str = os.getenv("WEBSOCKET_URL", "ws://localhost:8080/ws")
    
    # 前端地址
    frontend_url: str = os.getenv("FRONTEND_URL", "http://localhost:3000")
    
    # 用户服务地址
    user_service_url: str = os.getenv("USER_SERVICE_URL", "http://localhost:8080/api/v1")
    
    # 安全配置
    enable_rate_limiting: bool = os.getenv("ENABLE_RATE_LIMITING", "true").lower() == "true"
    max_requests_per_minute: int = int(os.getenv("MAX_REQUESTS_PER_MINUTE", "60"))
    enable_https: bool = os.getenv("ENABLE_HTTPS", "false").lower() == "true"
    ssl_keyfile: Optional[str] = os.getenv("SSL_KEYFILE")
    ssl_certfile: Optional[str] = os.getenv("SSL_CERTFILE")
    debug_mode: bool = os.getenv("DEBUG_MODE", "false").lower() == "true"
    
    # 内容安全配置
    enable_xss_protection: bool = os.getenv("ENABLE_XSS_PROTECTION", "true").lower() == "true"
    max_content_length: int = int(os.getenv("MAX_CONTENT_LENGTH", "10000"))  # 最大内容长度
    enable_content_filter: bool = os.getenv("ENABLE_CONTENT_FILTER", "true").lower() == "true"
    
    # 日志配置  
    log_level: str = os.getenv("LOG_LEVEL", "INFO")
    
    class Config:
        env_file = ".env"
        extra = "allow"  # Allow extra fields from environment variables
    
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        # 延迟加载JWT密钥以避免循环导入
        self._init_jwt_secret()
    
    def _init_jwt_secret(self):
        """初始化JWT密钥"""
        from app.utils.security import SecurityManager
        self.jwt_secret = SecurityManager.get_environment_jwt_secret()

settings = Settings()