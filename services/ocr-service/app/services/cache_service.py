import redis
import json
import hashlib
import logging
from typing import Optional, Dict, Any
from datetime import datetime, timedelta
import pickle
import os

from app.core.config import Config

logger = logging.getLogger(__name__)


class CacheService:
    """Redis缓存服务"""
    
    def __init__(self, config: Config = None):
        self.config = config or Config()
        self.redis_client = None
        self.is_available = False
        self._connect()
    
    def _connect(self):
        """连接Redis"""
        try:
            self.redis_client = redis.Redis(
                host=self.config.REDIS_HOST,
                port=self.config.REDIS_PORT,
                password=self.config.REDIS_PASSWORD if self.config.REDIS_PASSWORD else None,
                db=self.config.REDIS_DB,
                decode_responses=False,  # 处理二进制数据
                socket_connect_timeout=5,
                socket_timeout=5,
                retry_on_timeout=True
            )
            
            # 测试连接
            self.redis_client.ping()
            self.is_available = True
            logger.info("Redis缓存服务连接成功")
            
        except Exception as e:
            logger.warning(f"Redis连接失败，将使用内存缓存: {str(e)}")
            self.is_available = False
            self._init_memory_cache()
    
    def _init_memory_cache(self):
        """初始化内存缓存作为降级方案"""
        self._memory_cache = {}
        self._cache_timestamps = {}
        logger.info("使用内存缓存作为降级方案")
    
    def _generate_cache_key(self, prefix: str, data: Dict) -> str:
        """生成缓存键"""
        # 对输入数据进行哈希，确保相同输入得到相同缓存键
        data_str = json.dumps(data, sort_keys=True, ensure_ascii=False)
        data_hash = hashlib.md5(data_str.encode('utf-8')).hexdigest()
        return f"{prefix}:{data_hash}"
    
    def get_ocr_result(self, image_hash: str, settings: Dict) -> Optional[Dict]:
        """获取OCR结果缓存"""
        try:
            cache_key = self._generate_cache_key("ocr_result", {
                "image_hash": image_hash,
                "settings": settings
            })
            
            if self.is_available:
                # 从Redis获取
                cached_data = self.redis_client.get(cache_key)
                if cached_data:
                    result = pickle.loads(cached_data)
                    logger.info(f"从Redis缓存命中OCR结果: {cache_key}")
                    return result
            else:
                # 从内存缓存获取
                if cache_key in self._memory_cache:
                    timestamp = self._cache_timestamps.get(cache_key)
                    if timestamp and datetime.now() - timestamp < timedelta(seconds=self.config.CACHE_TTL):
                        logger.info(f"从内存缓存命中OCR结果: {cache_key}")
                        return self._memory_cache[cache_key]
                    else:
                        # 过期，删除
                        self._memory_cache.pop(cache_key, None)
                        self._cache_timestamps.pop(cache_key, None)
            
            return None
            
        except Exception as e:
            logger.warning(f"获取OCR缓存失败: {str(e)}")
            return None
    
    def cache_ocr_result(self, image_hash: str, settings: Dict, result: Dict):
        """缓存OCR识别结果"""
        try:
            cache_key = self._generate_cache_key("ocr_result", {
                "image_hash": image_hash,
                "settings": settings
            })
            
            # 添加缓存时间戳
            result_with_timestamp = {
                **result,
                "cached_at": datetime.utcnow().isoformat() + "Z",
                "cache_key": cache_key
            }
            
            if self.is_available:
                # 存储到Redis
                self.redis_client.setex(
                    cache_key,
                    self.config.CACHE_TTL,
                    pickle.dumps(result_with_timestamp)
                )
                logger.info(f"OCR结果已缓存到Redis: {cache_key}")
            else:
                # 存储到内存缓存
                self._memory_cache[cache_key] = result_with_timestamp
                self._cache_timestamps[cache_key] = datetime.now()
                
                # 简单的内存缓存清理（保持最多1000个条目）
                if len(self._memory_cache) > 1000:
                    self._cleanup_memory_cache()
                
                logger.info(f"OCR结果已缓存到内存: {cache_key}")
                
        except Exception as e:
            logger.warning(f"缓存OCR结果失败: {str(e)}")
    
    def _cleanup_memory_cache(self):
        """清理内存缓存"""
        try:
            current_time = datetime.now()
            expired_keys = []
            
            # 找出过期的键
            for key, timestamp in self._cache_timestamps.items():
                if current_time - timestamp > timedelta(seconds=self.config.CACHE_TTL):
                    expired_keys.append(key)
            
            # 删除过期的缓存
            for key in expired_keys:
                self._memory_cache.pop(key, None)
                self._cache_timestamps.pop(key, None)
            
            # 如果还是太多，删除最老的一半
            if len(self._memory_cache) > 500:
                sorted_items = sorted(
                    self._cache_timestamps.items(),
                    key=lambda x: x[1]
                )
                
                keys_to_remove = [item[0] for item in sorted_items[:len(sorted_items)//2]]
                for key in keys_to_remove:
                    self._memory_cache.pop(key, None)
                    self._cache_timestamps.pop(key, None)
            
            logger.info(f"内存缓存清理完成，剩余条目: {len(self._memory_cache)}")
            
        except Exception as e:
            logger.error(f"内存缓存清理失败: {str(e)}")
    
    def calculate_image_hash(self, image_path: str) -> str:
        """计算图像文件的哈希值"""
        try:
            with open(image_path, 'rb') as f:
                file_content = f.read()
                return hashlib.sha256(file_content).hexdigest()
        except Exception as e:
            logger.error(f"计算图像哈希失败: {str(e)}")
            # 使用文件路径和修改时间作为降级方案
            try:
                stat = os.stat(image_path)
                fallback_data = f"{image_path}_{stat.st_size}_{stat.st_mtime}"
                return hashlib.md5(fallback_data.encode()).hexdigest()
            except:
                return hashlib.md5(image_path.encode()).hexdigest()
    
    def get_task_status(self, task_id: str) -> Optional[Dict]:
        """获取任务状态"""
        try:
            cache_key = f"task_status:{task_id}"
            
            if self.is_available:
                cached_data = self.redis_client.get(cache_key)
                if cached_data:
                    return json.loads(cached_data.decode('utf-8'))
            else:
                if cache_key in self._memory_cache:
                    timestamp = self._cache_timestamps.get(cache_key)
                    if timestamp and datetime.now() - timestamp < timedelta(minutes=30):
                        return self._memory_cache[cache_key]
            
            return None
            
        except Exception as e:
            logger.warning(f"获取任务状态缓存失败: {str(e)}")
            return None
    
    def cache_task_status(self, task_id: str, status_data: Dict, ttl: int = 1800):
        """缓存任务状态 (默认30分钟)"""
        try:
            cache_key = f"task_status:{task_id}"
            
            if self.is_available:
                self.redis_client.setex(
                    cache_key,
                    ttl,
                    json.dumps(status_data, ensure_ascii=False)
                )
            else:
                self._memory_cache[cache_key] = status_data
                self._cache_timestamps[cache_key] = datetime.now()
                
        except Exception as e:
            logger.warning(f"缓存任务状态失败: {str(e)}")
    
    def update_task_progress(self, task_id: str, progress_data: Dict):
        """更新任务进度"""
        try:
            cache_key = f"task_progress:{task_id}"
            
            if self.is_available:
                # 设置较短的TTL，因为进度信息变化频繁
                self.redis_client.setex(
                    cache_key,
                    300,  # 5分钟
                    json.dumps(progress_data, ensure_ascii=False)
                )
            else:
                self._memory_cache[cache_key] = progress_data
                self._cache_timestamps[cache_key] = datetime.now()
                
        except Exception as e:
            logger.warning(f"更新任务进度失败: {str(e)}")
    
    def get_task_progress(self, task_id: str) -> Optional[Dict]:
        """获取任务进度"""
        try:
            cache_key = f"task_progress:{task_id}"
            
            if self.is_available:
                cached_data = self.redis_client.get(cache_key)
                if cached_data:
                    return json.loads(cached_data.decode('utf-8'))
            else:
                if cache_key in self._memory_cache:
                    timestamp = self._cache_timestamps.get(cache_key)
                    if timestamp and datetime.now() - timestamp < timedelta(minutes=5):
                        return self._memory_cache[cache_key]
            
            return None
            
        except Exception as e:
            logger.warning(f"获取任务进度失败: {str(e)}")
            return None
    
    def cache_batch_result(self, batch_id: str, results: list, ttl: int = 3600):
        """缓存批量处理结果"""
        try:
            cache_key = f"batch_result:{batch_id}"
            
            if self.is_available:
                self.redis_client.setex(
                    cache_key,
                    ttl,
                    pickle.dumps(results)
                )
            else:
                self._memory_cache[cache_key] = results
                self._cache_timestamps[cache_key] = datetime.now()
                
        except Exception as e:
            logger.warning(f"缓存批量结果失败: {str(e)}")
    
    def get_batch_result(self, batch_id: str) -> Optional[list]:
        """获取批量处理结果"""
        try:
            cache_key = f"batch_result:{batch_id}"
            
            if self.is_available:
                cached_data = self.redis_client.get(cache_key)
                if cached_data:
                    return pickle.loads(cached_data)
            else:
                if cache_key in self._memory_cache:
                    timestamp = self._cache_timestamps.get(cache_key)
                    if timestamp and datetime.now() - timestamp < timedelta(hours=1):
                        return self._memory_cache[cache_key]
            
            return None
            
        except Exception as e:
            logger.warning(f"获取批量结果失败: {str(e)}")
            return None
    
    def clear_cache(self, pattern: str = None):
        """清理缓存"""
        try:
            if self.is_available:
                if pattern:
                    # 清理匹配模式的键
                    keys = self.redis_client.keys(pattern)
                    if keys:
                        self.redis_client.delete(*keys)
                        logger.info(f"清理了 {len(keys)} 个匹配 {pattern} 的缓存键")
                else:
                    # 清理所有OCR相关缓存
                    keys = self.redis_client.keys("ocr_*")
                    keys.extend(self.redis_client.keys("task_*"))
                    keys.extend(self.redis_client.keys("batch_*"))
                    if keys:
                        self.redis_client.delete(*keys)
                        logger.info(f"清理了 {len(keys)} 个OCR缓存键")
            else:
                # 清理内存缓存
                if pattern:
                    keys_to_remove = [key for key in self._memory_cache.keys() if pattern in key]
                else:
                    keys_to_remove = list(self._memory_cache.keys())
                
                for key in keys_to_remove:
                    self._memory_cache.pop(key, None)
                    self._cache_timestamps.pop(key, None)
                
                logger.info(f"清理了 {len(keys_to_remove)} 个内存缓存条目")
                
        except Exception as e:
            logger.error(f"清理缓存失败: {str(e)}")
    
    def get_cache_stats(self) -> Dict:
        """获取缓存统计信息"""
        try:
            if self.is_available:
                info = self.redis_client.info()
                return {
                    "cache_type": "redis",
                    "connected": True,
                    "used_memory": info.get('used_memory_human', 'unknown'),
                    "total_keys": info.get('db0', {}).get('keys', 0) if 'db0' in info else 0,
                    "hits": info.get('keyspace_hits', 0),
                    "misses": info.get('keyspace_misses', 0)
                }
            else:
                return {
                    "cache_type": "memory",
                    "connected": False,
                    "total_keys": len(self._memory_cache),
                    "memory_entries": len(self._memory_cache)
                }
                
        except Exception as e:
            logger.error(f"获取缓存统计失败: {str(e)}")
            return {
                "cache_type": "unknown",
                "connected": False,
                "error": str(e)
            }


# 全局缓存服务实例
_cache_service = None

def get_cache_service() -> CacheService:
    """获取缓存服务实例（单例模式）"""
    global _cache_service
    if _cache_service is None:
        _cache_service = CacheService()
    return _cache_service