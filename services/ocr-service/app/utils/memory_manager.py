import gc
import os
import psutil
import logging
from typing import Dict, Optional
import cv2
import numpy as np

logger = logging.getLogger(__name__)


class MemoryManager:
    """内存管理器"""
    
    def __init__(self):
        self.process = psutil.Process()
        self.initial_memory = self.get_memory_usage()
        
    def get_memory_usage(self) -> Dict[str, float]:
        """获取当前内存使用情况"""
        try:
            memory_info = self.process.memory_info()
            memory_percent = self.process.memory_percent()
            
            return {
                'rss_mb': memory_info.rss / 1024 / 1024,  # 物理内存 (MB)
                'vms_mb': memory_info.vms / 1024 / 1024,  # 虚拟内存 (MB)
                'percent': memory_percent,                 # 内存使用百分比
                'available_mb': psutil.virtual_memory().available / 1024 / 1024
            }
        except Exception as e:
            logger.warning(f"获取内存使用情况失败: {str(e)}")
            return {'rss_mb': 0, 'vms_mb': 0, 'percent': 0, 'available_mb': 0}
    
    def cleanup_memory(self) -> bool:
        """清理内存"""
        try:
            # 强制垃圾回收
            collected = gc.collect()
            
            # 清理OpenCV缓存
            cv2.destroyAllWindows()
            
            memory_after = self.get_memory_usage()
            
            logger.info(f"内存清理完成: 回收对象{collected}个, "
                       f"内存使用: {memory_after['rss_mb']:.1f}MB "
                       f"({memory_after['percent']:.1f}%)")
            
            return True
            
        except Exception as e:
            logger.error(f"内存清理失败: {str(e)}")
            return False
    
    def check_memory_pressure(self, threshold_percent: float = 80.0) -> bool:
        """检查内存压力"""
        memory_usage = self.get_memory_usage()
        return memory_usage['percent'] > threshold_percent
    
    def optimize_image_for_memory(self, image: np.ndarray, max_size_mb: float = 50.0) -> np.ndarray:
        """优化图像以减少内存使用"""
        try:
            # 计算当前图像内存使用
            current_size_mb = image.nbytes / 1024 / 1024
            
            if current_size_mb <= max_size_mb:
                return image
            
            # 计算需要的缩放比例
            scale = np.sqrt(max_size_mb / current_size_mb)
            
            # 调整图像大小
            height, width = image.shape[:2]
            new_height = int(height * scale)
            new_width = int(width * scale)
            
            optimized = cv2.resize(image, (new_width, new_height), interpolation=cv2.INTER_AREA)
            
            logger.info(f"图像内存优化: {current_size_mb:.1f}MB -> {optimized.nbytes / 1024 / 1024:.1f}MB, "
                       f"尺寸: {width}x{height} -> {new_width}x{new_height}")
            
            return optimized
            
        except Exception as e:
            logger.warning(f"图像内存优化失败: {str(e)}")
            return image
    
    def clean_temp_files(self, temp_dir: str, max_age_hours: int = 2) -> int:
        """清理临时文件"""
        cleaned_count = 0
        
        try:
            import time
            current_time = time.time()
            max_age_seconds = max_age_hours * 3600
            
            if not os.path.exists(temp_dir):
                return 0
            
            for root, dirs, files in os.walk(temp_dir):
                for file in files:
                    file_path = os.path.join(root, file)
                    try:
                        file_age = current_time - os.path.getmtime(file_path)
                        if file_age > max_age_seconds:
                            os.remove(file_path)
                            cleaned_count += 1
                    except Exception as e:
                        logger.warning(f"清理文件失败 {file_path}: {str(e)}")
                        continue
                
                # 清理空目录
                for dir_name in dirs:
                    dir_path = os.path.join(root, dir_name)
                    try:
                        if not os.listdir(dir_path):
                            os.rmdir(dir_path)
                    except Exception:
                        continue
            
            if cleaned_count > 0:
                logger.info(f"清理临时文件: {cleaned_count} 个文件")
            
            return cleaned_count
            
        except Exception as e:
            logger.error(f"清理临时文件失败: {str(e)}")
            return 0
    
    def get_system_info(self) -> Dict:
        """获取系统信息"""
        try:
            cpu_count = psutil.cpu_count()
            cpu_percent = psutil.cpu_percent(interval=1)
            memory = psutil.virtual_memory()
            disk = psutil.disk_usage('/')
            
            return {
                'cpu': {
                    'cores': cpu_count,
                    'usage_percent': cpu_percent
                },
                'memory': {
                    'total_gb': memory.total / 1024**3,
                    'available_gb': memory.available / 1024**3,
                    'usage_percent': memory.percent
                },
                'disk': {
                    'total_gb': disk.total / 1024**3,
                    'free_gb': disk.free / 1024**3,
                    'usage_percent': (disk.used / disk.total) * 100
                },
                'process': self.get_memory_usage()
            }
            
        except Exception as e:
            logger.error(f"获取系统信息失败: {str(e)}")
            return {}


# 全局内存管理器实例
_memory_manager = None

def get_memory_manager() -> MemoryManager:
    """获取内存管理器实例（单例模式）"""
    global _memory_manager
    if _memory_manager is None:
        _memory_manager = MemoryManager()
    return _memory_manager


def cleanup_on_memory_pressure(threshold: float = 80.0) -> bool:
    """在内存压力大时自动清理"""
    memory_manager = get_memory_manager()
    
    if memory_manager.check_memory_pressure(threshold):
        logger.warning(f"检测到内存压力，开始清理...")
        return memory_manager.cleanup_memory()
    
    return False


def monitor_memory_usage():
    """装饰器：监控函数的内存使用"""
    def decorator(func):
        def wrapper(*args, **kwargs):
            memory_manager = get_memory_manager()
            
            # 记录执行前的内存
            memory_before = memory_manager.get_memory_usage()
            
            try:
                result = func(*args, **kwargs)
                
                # 记录执行后的内存
                memory_after = memory_manager.get_memory_usage()
                
                memory_diff = memory_after['rss_mb'] - memory_before['rss_mb']
                
                if abs(memory_diff) > 10:  # 内存变化超过10MB时记录
                    logger.info(f"{func.__name__} 内存变化: {memory_diff:+.1f}MB, "
                              f"当前: {memory_after['rss_mb']:.1f}MB ({memory_after['percent']:.1f}%)")
                
                # 如果内存使用过高，自动清理
                cleanup_on_memory_pressure()
                
                return result
                
            except Exception as e:
                # 异常时也要清理内存
                memory_manager.cleanup_memory()
                raise e
        
        return wrapper
    return decorator