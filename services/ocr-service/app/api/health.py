from flask import Blueprint
from app.utils.response import success_response

health_bp = Blueprint('health', __name__)


@health_bp.route('/health', methods=['GET'])
def health_check():
    """健康检查接口"""
    try:
        # 检查OCR引擎状态
        from app.services.ocr_engine import MultiEngineOCR
        from app.services.cache_service import get_cache_service
        from app.utils.websocket_client import get_websocket_notifier
        
        ocr_engine = MultiEngineOCR()
        cache_service = get_cache_service()
        ws_notifier = get_websocket_notifier()
        
        engine_status = ocr_engine.get_available_engines()
        available_engines = [name for name, info in engine_status.items() if info.get('available', False)]
        
        health_data = {
            "service": "ocr-service",
            "status": "healthy",
            "version": "1.0.0",
            "timestamp": "2025-07-21T00:00:00Z",
            "engines": {
                "available": available_engines,
                "total": len(engine_status),
                "details": engine_status
            },
            "cache": {
                "redis_available": cache_service.redis_client is not None,
                "status": "connected" if cache_service.redis_client else "disconnected"
            },
            "websocket": {
                "redis_available": ws_notifier.is_available,
                "status": "connected" if ws_notifier.is_available else "disconnected"
            },
            "features": {
                "handwriting_processing": True,
                "batch_processing": True,
                "text_validation": True,
                "chinese_optimization": True,
                "multi_engine_voting": len(available_engines) > 1
            }
        }
        
        return success_response(health_data)
        
    except Exception as e:
        return success_response({
            "service": "ocr-service",
            "status": "degraded",
            "version": "1.0.0",
            "error": str(e)
        }, status_code=503)


@health_bp.route('/ping', methods=['GET'])
def ping():
    """简单的ping接口"""
    return success_response("pong")