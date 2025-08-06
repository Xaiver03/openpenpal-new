from flask import Flask
from flask_cors import CORS
import os

from app.core.config import config
from app.api import ocr_bp, tasks_bp, health_bp


def create_app(config_name=None):
    """Flask应用工厂函数"""
    
    # 获取配置环境
    if config_name is None:
        config_name = os.getenv('FLASK_ENV', 'development')
    
    app = Flask(__name__)
    
    # 加载配置
    app.config.from_object(config[config_name])
    config[config_name].init_app(app)
    
    # 启用CORS
    CORS(app, origins=["http://localhost:3000", "http://localhost:8080"])
    
    # 注册蓝图
    app.register_blueprint(health_bp)
    app.register_blueprint(ocr_bp, url_prefix='/api/ocr')
    app.register_blueprint(tasks_bp, url_prefix='/api/ocr/tasks')
    
    # 错误处理
    @app.errorhandler(404)
    def not_found(error):
        from app.utils.response import not_found_error_response
        return not_found_error_response("接口不存在"), 404
    
    @app.errorhandler(500)
    def internal_error(error):
        from app.utils.response import internal_error_response
        return internal_error_response("服务器内部错误"), 500
    
    @app.errorhandler(413)
    def too_large(error):
        from app.utils.response import validation_error_response
        return validation_error_response("文件大小超出限制"), 413
    
    return app


# 为gunicorn创建app实例
app = create_app()


if __name__ == '__main__':
    app.run(
        host=app.config.get('HOST', '0.0.0.0'),
        port=app.config.get('PORT', 8004),
        debug=app.config.get('DEBUG', False)
    )