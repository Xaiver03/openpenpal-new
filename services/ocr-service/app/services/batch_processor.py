import uuid
import time
import logging
import threading
from typing import List, Dict, Any, Optional
from concurrent.futures import ThreadPoolExecutor, as_completed
import json

from app.services.ocr_engine import MultiEngineOCR
from app.services.cache_service import get_cache_service
from app.utils.websocket_client import get_websocket_notifier

logger = logging.getLogger(__name__)


class BatchProcessor:
    """批量OCR处理器"""
    
    def __init__(self):
        self.ocr_engine = None
        self.cache_service = get_cache_service()
        self.ws_notifier = get_websocket_notifier()
        self.processing_batches = {}  # 存储正在处理的批量任务
        self.max_workers = 4  # 并发处理数量
    
    def get_ocr_engine(self):
        """获取OCR引擎实例（懒加载）"""
        if self.ocr_engine is None:
            self.ocr_engine = MultiEngineOCR(use_gpu=False)
        return self.ocr_engine
    
    def start_batch_processing(self, batch_id: str, files: List, settings: Dict, user_id: str) -> Dict:
        """启动批量处理任务"""
        try:
            # 保存文件到临时目录
            file_paths = []
            for i, file in enumerate(files):
                file_path, error = self._save_temp_file(file, batch_id, i)
                if error:
                    logger.error(f"保存文件失败: {error}")
                    continue
                file_paths.append({
                    'path': file_path,
                    'name': file.filename,
                    'index': i
                })
            
            if not file_paths:
                raise ValueError("没有有效的文件可以处理")
            
            # 创建批量任务状态
            batch_status = {
                "batch_id": batch_id,
                "user_id": user_id,
                "total_images": len(file_paths),
                "completed_images": 0,
                "failed_images": 0,
                "progress_percentage": 0,
                "status": "processing",
                "created_at": time.time(),
                "estimated_completion": time.time() + len(file_paths) * 5,  # 预估每张图片5秒
                "current_step": f"开始处理 {len(file_paths)} 张图片",
                "results": [],
                "settings": settings,
                "file_paths": file_paths
            }
            
            # 保存到缓存
            self.processing_batches[batch_id] = batch_status
            self._cache_batch_status(batch_id, batch_status)
            
            # 启动后台处理线程
            processing_thread = threading.Thread(
                target=self._process_batch_async,
                args=(batch_id, file_paths, settings, user_id),
                daemon=True
            )
            processing_thread.start()
            
            # 返回初始状态
            return {
                "batch_id": batch_id,
                "total_images": len(file_paths),
                "estimated_time": f"{len(file_paths) * 5}s",
                "status": "processing",
                "progress_url": f"/api/ocr/tasks/batch/{batch_id}/progress"
            }
            
        except Exception as e:
            logger.error(f"启动批量处理失败: {str(e)}")
            raise
    
    def _save_temp_file(self, file, batch_id: str, index: int) -> tuple:
        """保存临时文件"""
        try:
            import os
            from werkzeug.utils import secure_filename
            from flask import current_app
            
            if not file or file.filename == '':
                return None, "文件为空"
            
            # 检查文件类型
            allowed_extensions = {'jpg', 'jpeg', 'png', 'bmp', 'tiff'}
            if '.' not in file.filename or \
               file.filename.rsplit('.', 1)[1].lower() not in allowed_extensions:
                return None, "不支持的文件格式"
            
            # 生成文件名
            filename = secure_filename(file.filename)
            unique_filename = f"batch_{batch_id}_{index}_{filename}"
            
            # 创建批量处理临时目录
            batch_dir = os.path.join(current_app.config['UPLOAD_FOLDER'], 'batch', batch_id)
            os.makedirs(batch_dir, exist_ok=True)
            
            # 保存文件
            file_path = os.path.join(batch_dir, unique_filename)
            file.save(file_path)
            
            return file_path, None
            
        except Exception as e:
            logger.error(f"保存临时文件失败: {str(e)}")
            return None, str(e)
    
    def _process_batch_async(self, batch_id: str, file_paths: List[Dict], settings: Dict, user_id: str):
        """异步批量处理"""
        try:
            ocr_engine = self.get_ocr_engine()
            batch_status = self.processing_batches.get(batch_id)
            
            if not batch_status:
                logger.error(f"批量任务状态丢失: {batch_id}")
                return
            
            # 解析设置参数
            language = settings.get('language', 'zh')
            enhance = settings.get('enhance', True)
            engine_name = settings.get('engine', 'paddle')
            is_handwriting = settings.get('is_handwriting', False)
            use_voting = settings.get('use_voting', False)
            
            results = []
            
            # 使用线程池并行处理
            with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
                # 提交所有任务
                future_to_file = {}
                for file_info in file_paths:
                    if use_voting:
                        future = executor.submit(
                            ocr_engine.recognize_with_voting,
                            file_info['path'], language, enhance, is_handwriting
                        )
                    else:
                        future = executor.submit(
                            ocr_engine.recognize_single_engine,
                            file_info['path'], engine_name, language, enhance, is_handwriting
                        )
                    future_to_file[future] = file_info
                
                # 处理完成的任务
                for future in as_completed(future_to_file):
                    file_info = future_to_file[future]
                    
                    try:
                        # 获取识别结果
                        ocr_result = future.result()
                        
                        result = {
                            "image_name": file_info['name'],
                            "image_index": file_info['index'],
                            "status": "completed",
                            "text": ocr_result.get('text', ''),
                            "confidence": ocr_result.get('confidence', 0),
                            "processing_time": ocr_result.get('processing_time', 0),
                            "engine": ocr_result.get('engine', engine_name),
                            "word_count": len(ocr_result.get('text', '').replace(' ', '')),
                            "blocks": ocr_result.get('blocks', [])
                        }
                        
                        results.append(result)
                        batch_status['completed_images'] += 1
                        
                        logger.info(f"批量处理完成图片: {file_info['name']}")
                        
                    except Exception as e:
                        logger.error(f"处理图片 {file_info['name']} 失败: {str(e)}")
                        
                        result = {
                            "image_name": file_info['name'],
                            "image_index": file_info['index'],
                            "status": "failed",
                            "text": "",
                            "confidence": 0,
                            "error": str(e)
                        }
                        
                        results.append(result)
                        batch_status['failed_images'] += 1
                    
                    # 更新进度
                    progress = (batch_status['completed_images'] + batch_status['failed_images']) / batch_status['total_images'] * 100
                    batch_status['progress_percentage'] = int(progress)
                    batch_status['current_step'] = f"已处理 {len(results)}/{batch_status['total_images']} 张图片"
                    batch_status['results'] = sorted(results, key=lambda x: x['image_index'])
                    
                    # 缓存更新的状态
                    self._cache_batch_status(batch_id, batch_status)
                    
                    # 推送进度更新
                    self._push_progress_update(batch_id, batch_status, user_id)
            
            # 处理完成
            batch_status['status'] = 'completed'
            batch_status['progress_percentage'] = 100
            batch_status['current_step'] = '所有图片处理完成'
            batch_status['completed_at'] = time.time()
            
            # 计算统计信息
            successful_results = [r for r in results if r['status'] == 'completed']
            batch_status['statistics'] = {
                'total_processing_time': sum(r.get('processing_time', 0) for r in successful_results),
                'average_confidence': sum(r.get('confidence', 0) for r in successful_results) / len(successful_results) if successful_results else 0,
                'total_text_length': sum(len(r.get('text', '')) for r in successful_results),
                'success_rate': len(successful_results) / batch_status['total_images'] * 100
            }
            
            # 最终缓存和推送
            self._cache_batch_status(batch_id, batch_status)
            self._push_completion_update(batch_id, batch_status, user_id)
            
            # 清理临时文件
            self._cleanup_temp_files(batch_id, file_paths)
            
            logger.info(f"批量处理任务完成: {batch_id}, 成功: {len(successful_results)}, 失败: {batch_status['failed_images']}")
            
        except Exception as e:
            logger.error(f"批量处理异常: {str(e)}")
            
            # 更新为失败状态
            batch_status = self.processing_batches.get(batch_id, {})
            batch_status.update({
                'status': 'failed',
                'error': str(e),
                'completed_at': time.time()
            })
            
            self._cache_batch_status(batch_id, batch_status)
            self._push_error_update(batch_id, str(e), user_id)
    
    def _cache_batch_status(self, batch_id: str, status: Dict):
        """缓存批量任务状态"""
        try:
            cache_key = f"batch_status:{batch_id}"
            self.cache_service.redis_client.setex(
                cache_key, 
                3600,  # 1小时过期
                json.dumps(status, default=str)
            )
        except Exception as e:
            logger.error(f"缓存批量任务状态失败: {str(e)}")
    
    def get_batch_progress(self, batch_id: str) -> Optional[Dict]:
        """获取批量任务进度"""
        try:
            # 先从内存中获取
            if batch_id in self.processing_batches:
                return self.processing_batches[batch_id]
            
            # 从缓存中获取
            cache_key = f"batch_status:{batch_id}"
            cached_data = self.cache_service.redis_client.get(cache_key)
            
            if cached_data:
                return json.loads(cached_data)
            
            return None
            
        except Exception as e:
            logger.error(f"获取批量任务进度失败: {str(e)}")
            return None
    
    def _push_progress_update(self, batch_id: str, status: Dict, user_id: str):
        """推送进度更新"""
        try:
            self.ws_notifier.push_batch_progress(user_id, batch_id, status)
            
        except Exception as e:
            logger.error(f"推送进度更新失败: {str(e)}")
    
    def _push_completion_update(self, batch_id: str, status: Dict, user_id: str):
        """推送完成更新"""
        try:
            completion_data = {
                "batch_id": batch_id,
                "total_images": status['total_images'],
                "completed_images": status['completed_images'],
                "failed_images": status['failed_images'],
                "statistics": status.get('statistics', {}),
                "success_rate": status.get('statistics', {}).get('success_rate', 0),
                "status": "completed"
            }
            self.ws_notifier.push_batch_progress(user_id, batch_id, completion_data)
            
        except Exception as e:
            logger.error(f"推送完成更新失败: {str(e)}")
    
    def _push_error_update(self, batch_id: str, error: str, user_id: str):
        """推送错误更新"""
        try:
            error_data = {
                "batch_id": batch_id,
                "status": "failed",
                "error": error,
                "progress_percentage": 0
            }
            self.ws_notifier.push_batch_progress(user_id, batch_id, error_data)
            
        except Exception as e:
            logger.error(f"推送错误更新失败: {str(e)}")
    
    def _cleanup_temp_files(self, batch_id: str, file_paths: List[Dict]):
        """清理临时文件"""
        try:
            import os
            
            for file_info in file_paths:
                try:
                    if os.path.exists(file_info['path']):
                        os.remove(file_info['path'])
                except Exception as e:
                    logger.warning(f"删除临时文件失败: {file_info['path']}, {str(e)}")
            
            # 删除批量处理目录（如果为空）
            try:
                from flask import current_app
                batch_dir = os.path.join(current_app.config['UPLOAD_FOLDER'], 'batch', batch_id)
                if os.path.exists(batch_dir) and not os.listdir(batch_dir):
                    os.rmdir(batch_dir)
            except Exception as e:
                logger.warning(f"删除批量目录失败: {str(e)}")
                
        except Exception as e:
            logger.error(f"清理临时文件异常: {str(e)}")


# 全局批量处理器实例
_batch_processor = None

def get_batch_processor():
    """获取批量处理器实例（单例模式）"""
    global _batch_processor
    if _batch_processor is None:
        _batch_processor = BatchProcessor()
    return _batch_processor