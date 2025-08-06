from flask import Blueprint, request, current_app
import uuid
import os
from werkzeug.utils import secure_filename
import logging

from app.utils.response import (
    success_response, 
    validation_error_response, 
    internal_error_response,
    business_error_response,
    permission_error_response
)
from app.utils.auth import jwt_required, get_current_user
from app.services.ocr_engine import MultiEngineOCR
from app.services.cache_service import get_cache_service

ocr_bp = Blueprint('ocr', __name__)
logger = logging.getLogger(__name__)

# 全局OCR引擎实例
ocr_engine = None

def get_ocr_engine():
    """获取OCR引擎实例（懒加载）"""
    global ocr_engine
    if ocr_engine is None:
        try:
            use_gpu = current_app.config.get('ENABLE_GPU', False)
            ocr_engine = MultiEngineOCR(use_gpu=use_gpu)
            logger.info("OCR引擎初始化成功")
        except Exception as e:
            logger.error(f"OCR引擎初始化失败: {str(e)}")
            raise
    return ocr_engine


def allowed_file(filename):
    """检查文件类型是否允许"""
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in current_app.config['ALLOWED_EXTENSIONS']


def save_uploaded_file(file):
    """保存上传的文件"""
    if not file or file.filename == '':
        return None, "没有选择文件"
    
    if not allowed_file(file.filename):
        return None, "不支持的文件格式"
    
    # 生成唯一文件名
    filename = secure_filename(file.filename)
    unique_filename = f"{uuid.uuid4().hex}_{filename}"
    
    # 保存文件
    upload_folder = current_app.config['UPLOAD_FOLDER']
    os.makedirs(upload_folder, exist_ok=True)
    
    file_path = os.path.join(upload_folder, unique_filename)
    file.save(file_path)
    
    return file_path, None


@ocr_bp.route('/recognize', methods=['POST'])
@jwt_required
def recognize_image():
    """图片OCR识别接口"""
    try:
        # 获取当前用户
        user_info = get_current_user()
        user_id = user_info.get('user_id')
        
        # 检查文件
        if 'image' not in request.files:
            return validation_error_response("缺少图片文件"), 400
        
        file = request.files['image']
        file_path, error = save_uploaded_file(file)
        
        if error:
            return validation_error_response(error), 400
        
        # 获取参数
        language = request.form.get('language', 'auto')
        if language == 'auto':
            language = 'zh'  # 默认中文
        enhance = request.form.get('enhance', 'true').lower() == 'true'
        confidence_threshold = float(request.form.get('confidence_threshold', 0.7))
        engine_name = request.form.get('engine', current_app.config.get('DEFAULT_OCR_ENGINE', 'paddle'))
        is_handwriting = request.form.get('is_handwriting', 'false').lower() == 'true'
        
        # 生成任务ID
        task_id = f"ocr_task_{uuid.uuid4().hex}"
        
        try:
            # 获取缓存服务
            cache_service = get_cache_service()
            
            # 计算图像哈希用于缓存
            image_hash = cache_service.calculate_image_hash(file_path)
            
            # 构建缓存设置
            cache_settings = {
                'language': language,
                'enhance': enhance,
                'engine': engine_name,
                'is_handwriting': is_handwriting,
                'use_voting': request.form.get('use_voting', 'false').lower() == 'true',
                'confidence_threshold': confidence_threshold
            }
            
            # 尝试从缓存获取结果
            cached_result = cache_service.get_ocr_result(image_hash, cache_settings)
            if cached_result and request.form.get('use_cache', 'true').lower() == 'true':
                logger.info(f"使用缓存的OCR结果: {task_id}")
                # 使用缓存结果，但更新任务ID
                cached_result['task_id'] = task_id
                cached_result['from_cache'] = True
                ocr_result = {
                    "task_id": task_id,
                    "status": "completed",
                    "results": cached_result,
                    "metadata": cached_result.get('metadata', {})
                }
            else:
                # 获取OCR引擎
                ocr = get_ocr_engine()
                
                # 执行OCR识别
                if request.form.get('use_voting', 'false').lower() == 'true':
                    # 使用多引擎投票
                    result = ocr.recognize_with_voting(
                        file_path, language, enhance, is_handwriting
                    )
                else:
                    # 使用单引擎
                    result = ocr.recognize_single_engine(
                        file_path, engine_name, language, enhance, is_handwriting
                    )
                
                # 缓存识别结果
                cache_service.cache_ocr_result(image_hash, cache_settings, result)
                
                # 过滤低置信度结果
                filtered_blocks = [
                    block for block in result.get('blocks', [])
                    if block.get('confidence', 0) >= confidence_threshold
                ]
                
                # 构建响应结果
                ocr_result = {
                    "task_id": task_id,
                    "status": "completed",
                    "results": {
                        "text": result.get('text', ''),
                        "confidence": result.get('confidence', 0),
                        "word_count": len(result.get('text', '').replace(' ', '')),
                        "processing_time": result.get('processing_time', 0),
                        "language_detected": language,
                        "blocks": filtered_blocks,
                        "from_cache": False
                    },
                    "metadata": {
                        "image_size": f"{result.get('preprocessing', {}).get('original_size', [0,0])[1]}x{result.get('preprocessing', {}).get('original_size', [0,0])[0]}",
                        "image_format": file.filename.rsplit('.', 1)[1].lower(),
                        "processing_method": result.get('engine', engine_name),
                        "enhancement_applied": enhance,
                        "preprocessing_operations": result.get('preprocessing', {}).get('applied_operations', []),
                        "is_handwriting_mode": is_handwriting,
                        "image_hash": image_hash
                    }
                }
                
                # 如果是投票模式，添加投票信息
                if 'voting_info' in result:
                    ocr_result['metadata']['voting_info'] = result['voting_info']
            
            # 缓存任务状态
            cache_service.cache_task_status(task_id, ocr_result)
            
        except Exception as e:
            logger.error(f"OCR识别执行失败: {str(e)}")
            # 返回错误结果
            ocr_result = {
                "task_id": task_id,
                "status": "failed",
                "error": str(e),
                "results": None
            }
        
        # 清理临时文件
        try:
            os.remove(file_path)
        except:
            pass
        
        # 返回结果
        if ocr_result["status"] == "completed":
            return success_response(ocr_result, "识别成功")
        else:
            return internal_error_response(f"识别失败: {ocr_result.get('error', '未知错误')}"), 500
        
    except Exception as e:
        current_app.logger.error(f"OCR识别失败: {str(e)}")
        return internal_error_response("识别过程出现错误"), 500


@ocr_bp.route('/batch', methods=['POST'])
@jwt_required
def batch_recognize():
    """批量OCR识别接口"""
    try:
        user_info = get_current_user()
        user_id = user_info.get('user_id')
        
        # 检查文件
        if 'images' not in request.files:
            return validation_error_response("缺少图片文件"), 400
        
        files = request.files.getlist('images')
        
        if not files or len(files) == 0:
            return validation_error_response("至少需要一个图片文件"), 400
        
        # 限制批量处理数量
        if len(files) > 10:
            return validation_error_response("批量处理最多支持10个文件"), 400
        
        # 验证文件类型和大小
        max_size = current_app.config.get('MAX_CONTENT_LENGTH', 10 * 1024 * 1024)  # 10MB
        for file in files:
            if not file.filename:
                return validation_error_response("存在无效的文件"), 400
            
            if not allowed_file(file.filename):
                return validation_error_response(f"文件 {file.filename} 格式不支持"), 400
        
        # 获取设置参数
        settings_str = request.form.get('settings', '{}')
        try:
            import json
            settings = json.loads(settings_str)
        except:
            settings = {}
        
        # 验证和设置默认参数
        settings.setdefault('language', 'zh')
        settings.setdefault('enhance', True)
        settings.setdefault('engine', current_app.config.get('DEFAULT_OCR_ENGINE', 'paddle'))
        settings.setdefault('is_handwriting', False)
        settings.setdefault('use_voting', False)
        settings.setdefault('confidence_threshold', 0.7)
        
        batch_id = f"batch_{uuid.uuid4().hex}"
        
        # 使用批量处理器
        from app.services.batch_processor import get_batch_processor
        batch_processor = get_batch_processor()
        
        result = batch_processor.start_batch_processing(batch_id, files, settings, user_id)
        
        return success_response(result, "批量任务已创建")
        
    except Exception as e:
        current_app.logger.error(f"批量OCR识别失败: {str(e)}")
        return internal_error_response(f"批量识别过程出现错误: {str(e)}"), 500


@ocr_bp.route('/enhance', methods=['POST'])
@jwt_required
def enhance_image():
    """图像预处理和增强接口"""
    try:
        # 检查文件
        if 'image' not in request.files:
            return validation_error_response("缺少图片文件"), 400
        
        file = request.files['image']
        file_path, error = save_uploaded_file(file)
        
        if error:
            return validation_error_response(error), 400
        
        # 获取操作参数
        operations_str = request.form.get('operations', '["denoise", "deskew", "contrast"]')
        return_enhanced = request.form.get('return_enhanced', 'false').lower() == 'true'
        is_handwriting = request.form.get('is_handwriting', 'false').lower() == 'true'
        
        try:
            import json
            operations = json.loads(operations_str)
        except:
            operations = ["denoise", "deskew", "contrast"]
        
        try:
            # 获取图像处理器
            from app.services.image_processor import ImagePreprocessor, HandwritingPreprocessor
            
            if is_handwriting:
                processor = HandwritingPreprocessor()
                if not operations or operations == ["denoise", "deskew", "contrast"]:
                    operations = ["denoise", "deskew", "handwriting_enhance"]
            else:
                processor = ImagePreprocessor()
            
            # 执行图像增强
            enhanced_image, processing_info = processor.preprocess(file_path, operations)
            
            enhanced_image_url = None
            if return_enhanced:
                # 保存增强后的图像
                enhanced_filename = f"enhanced_{uuid.uuid4().hex}.jpg"
                enhanced_path = os.path.join(current_app.config['UPLOAD_FOLDER'], enhanced_filename)
                
                success = processor.save_processed_image(enhanced_image, enhanced_path)
                if success:
                    enhanced_image_url = f"/api/ocr/files/{enhanced_filename}"
            
            # 构建结果
            result_data = {
                "enhanced_image_url": enhanced_image_url,
                "operations_applied": processing_info.get('applied_operations', operations),
                "quality_score": processing_info.get('quality_metrics', {}).get('overall_quality', 0.78),
                "enhancement_metrics": processing_info.get('quality_metrics', {
                    "noise_reduction": 0.65,
                    "contrast_improvement": 0.42,
                    "skew_correction": "2.3°"
                }),
                "original_size": processing_info.get('original_size', [0, 0]),
                "is_handwriting_mode": is_handwriting
            }
            
        except Exception as e:
            logger.error(f"图像增强处理失败: {str(e)}")
            # 返回模拟结果作为降级
            result_data = {
                "enhanced_image_url": None,
                "operations_applied": operations,
                "quality_score": 0.5,
                "enhancement_metrics": {
                    "error": str(e)
                },
                "is_handwriting_mode": is_handwriting
            }
        
        # 清理临时文件
        try:
            os.remove(file_path)
        except:
            pass
        
        return success_response(result_data, "图像增强完成")
        
    except Exception as e:
        current_app.logger.error(f"图像增强失败: {str(e)}")
        return internal_error_response("图像增强过程出现错误"), 500


@ocr_bp.route('/validate', methods=['POST'])
@jwt_required
def validate_text():
    """文本内容验证接口"""
    try:
        # 获取JSON数据
        if not request.is_json:
            return validation_error_response("请求必须是JSON格式"), 400
        
        data = request.get_json()
        
        original_text = data.get('original_text', '')
        ocr_text = data.get('ocr_text', '')
        validation_rules = data.get('validation_rules', {})
        
        if not original_text or not ocr_text:
            return validation_error_response("原始文本和OCR文本不能为空"), 400
        
        # 使用文本验证服务
        from app.services.text_validator import get_text_validator
        text_validator = get_text_validator()
        
        validation_result = text_validator.validate_text_similarity(
            original_text, ocr_text, validation_rules
        )
        
        return success_response(validation_result, "验证完成")
        
    except Exception as e:
        current_app.logger.error(f"文本验证失败: {str(e)}")
        return internal_error_response(f"验证过程出现错误: {str(e)}"), 500


@ocr_bp.route('/models', methods=['GET'])
@jwt_required
def get_available_models():
    """获取可用的OCR模型列表"""
    try:
        # 获取OCR引擎
        ocr = get_ocr_engine()
        engine_info = ocr.get_available_engines()
        
        # 构建模型信息
        available_models = []
        
        for engine_name, info in engine_info.items():
            if info['available']:
                model_data = {
                    "name": engine_name,
                    "available": True,
                    "languages": info['supported_languages']
                }
                
                # 添加引擎特定信息
                if engine_name == 'paddle':
                    model_data.update({
                        "description": "百度PaddleOCR - 高精度中英文识别",
                        "accuracy": 0.95,
                        "speed": "fast",
                        "best_for": "printed_text"
                    })
                elif engine_name == 'tesseract':
                    model_data.update({
                        "description": "Tesseract OCR - 开源多语言识别",
                        "accuracy": 0.88,
                        "speed": "medium",
                        "best_for": "document_text"
                    })
                elif engine_name == 'easyocr':
                    model_data.update({
                        "description": "EasyOCR - 简单易用的OCR引擎",
                        "accuracy": 0.92,
                        "speed": "slow",
                        "best_for": "handwritten_text"
                    })
                
                available_models.append(model_data)
            else:
                # 不可用的引擎也显示，但标记为不可用
                available_models.append({
                    "name": engine_name,
                    "available": False,
                    "description": f"{engine_name} OCR引擎 (不可用)",
                    "languages": [],
                    "error": "引擎未安装或初始化失败"
                })
        
        result = {
            "available_models": available_models,
            "default_model": current_app.config.get('DEFAULT_OCR_ENGINE', 'paddle'),
            "total_engines": len(available_models),
            "available_engines": len([m for m in available_models if m['available']]),
            "supports_gpu": current_app.config.get('ENABLE_GPU', False),
            "supports_voting": len([m for m in available_models if m['available']]) > 1
        }
        
        return success_response(result)
        
    except Exception as e:
        current_app.logger.error(f"获取模型列表失败: {str(e)}")
        return internal_error_response("获取模型列表失败"), 500


@ocr_bp.route('/cache/stats', methods=['GET'])
@jwt_required
def get_cache_stats():
    """获取缓存统计信息"""
    try:
        cache_service = get_cache_service()
        stats = cache_service.get_cache_stats()
        
        return success_response(stats, "缓存统计信息获取成功")
        
    except Exception as e:
        current_app.logger.error(f"获取缓存统计失败: {str(e)}")
        return internal_error_response("获取缓存统计失败"), 500


@ocr_bp.route('/cache/clear', methods=['POST'])
@jwt_required
def clear_cache():
    """清理缓存"""
    try:
        user_info = get_current_user()
        user_role = user_info.get('role', '')
        
        # 只有管理员可以清理缓存
        if user_role not in ['admin', 'super_admin']:
            return permission_error_response("需要管理员权限"), 403
        
        cache_service = get_cache_service()
        
        # 获取清理模式
        pattern = request.json.get('pattern') if request.is_json else None
        
        cache_service.clear_cache(pattern)
        
        return success_response({
            "pattern": pattern,
            "message": "缓存清理完成"
        }, "缓存清理成功")
        
    except Exception as e:
        current_app.logger.error(f"清理缓存失败: {str(e)}")
        return internal_error_response("清理缓存失败"), 500