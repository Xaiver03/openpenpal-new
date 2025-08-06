import os
from typing import Optional


class Config:
    """OCR服务配置类"""
    
    # Flask配置
    SECRET_KEY = os.getenv('SECRET_KEY', 'ocr-service-secret-key')
    DEBUG = os.getenv('FLASK_DEBUG', 'false').lower() == 'true'
    HOST = os.getenv('HOST', '0.0.0.0')
    PORT = int(os.getenv('PORT', 8004))
    
    # JWT配置
    JWT_SECRET = os.getenv('JWT_SECRET', 'shared-jwt-secret')
    JWT_ALGORITHM = 'HS256'
    JWT_EXPIRATION_HOURS = 24
    
    # Redis配置
    REDIS_HOST = os.getenv('REDIS_HOST', 'localhost')
    REDIS_PORT = int(os.getenv('REDIS_PORT', 6379))
    REDIS_PASSWORD = os.getenv('REDIS_PASSWORD', '')
    REDIS_DB = int(os.getenv('REDIS_DB', 0))
    
    # 文件上传配置
    MAX_FILE_SIZE = int(os.getenv('MAX_FILE_SIZE', 10485760))  # 10MB
    # 使用固定的相对路径，避免只读文件系统问题
    # 获取当前文件所在目录的父目录（services/ocr-service）
    _base_dir = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    UPLOAD_FOLDER = os.getenv('UPLOAD_FOLDER', os.path.join(_base_dir, 'uploads'))
    ALLOWED_EXTENSIONS = {'jpg', 'jpeg', 'png', 'bmp', 'tiff'}
    
    # OCR模型配置
    # 使用固定的相对路径，避免只读文件系统问题
    MODEL_PATH = os.getenv('MODEL_PATH', os.path.join(_base_dir, 'models'))
    DEFAULT_OCR_ENGINE = os.getenv('DEFAULT_OCR_ENGINE', 'paddle')
    ENABLE_GPU = os.getenv('ENABLE_GPU', 'false').lower() == 'true'
    
    # 缓存配置
    CACHE_TTL = int(os.getenv('CACHE_TTL', 86400))  # 24小时
    
    # 性能配置
    MAX_WORKERS = int(os.getenv('MAX_WORKERS', 4))
    TASK_TIMEOUT = int(os.getenv('TASK_TIMEOUT', 300))  # 5分钟
    
    # 内存优化配置
    MAX_IMAGE_SIZE = int(os.getenv('MAX_IMAGE_SIZE', 2048))  # 最大图片尺寸（像素）
    ENABLE_IMAGE_CACHE = os.getenv('ENABLE_IMAGE_CACHE', 'true').lower() == 'true'
    MAX_BATCH_SIZE = int(os.getenv('MAX_BATCH_SIZE', 10))  # 批量处理最大数量
    CLEANUP_TEMP_FILES = os.getenv('CLEANUP_TEMP_FILES', 'true').lower() == 'true'
    
    # OCR引擎优化
    LAZY_LOAD_ENGINES = os.getenv('LAZY_LOAD_ENGINES', 'true').lower() == 'true'
    ENGINE_POOL_SIZE = int(os.getenv('ENGINE_POOL_SIZE', 2))  # 引擎池大小
    
    # 外部服务配置
    WRITE_SERVICE_URL = os.getenv('WRITE_SERVICE_URL', 'http://localhost:8002')
    WEBSOCKET_REDIS_CHANNEL = 'user_notifications'
    
    @classmethod
    def init_app(cls, app):
        """初始化Flask应用配置"""
        app.config['SECRET_KEY'] = cls.SECRET_KEY
        app.config['MAX_CONTENT_LENGTH'] = cls.MAX_FILE_SIZE
        
        # 创建上传目录（忽略只读文件系统错误）
        try:
            os.makedirs(cls.UPLOAD_FOLDER, exist_ok=True)
            os.makedirs(cls.MODEL_PATH, exist_ok=True)
        except OSError as e:
            # 忽略只读文件系统错误，使用临时目录
            import tempfile
            if e.errno == 30:  # Read-only file system
                cls.UPLOAD_FOLDER = tempfile.gettempdir()
                cls.MODEL_PATH = tempfile.gettempdir()
                print(f"Using temp directory for uploads: {cls.UPLOAD_FOLDER}")
            else:
                raise


class DevelopmentConfig(Config):
    """开发环境配置"""
    DEBUG = True
    REDIS_HOST = 'localhost'


class ProductionConfig(Config):
    """生产环境配置"""
    DEBUG = False


class TestingConfig(Config):
    """测试环境配置"""
    TESTING = True
    REDIS_DB = 1  # 使用不同的Redis数据库


config = {
    'development': DevelopmentConfig,
    'production': ProductionConfig,
    'testing': TestingConfig,
    'default': DevelopmentConfig
}