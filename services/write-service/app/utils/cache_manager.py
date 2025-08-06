import json
import asyncio
import logging
from typing import Optional, Any, Dict, List
import redis.asyncio as redis
from datetime import timedelta
from app.core.config import settings

logger = logging.getLogger(__name__)

class CacheManager:
    """Redis缓存管理器"""
    
    def __init__(self):
        self.redis_url = settings.redis_url
        self.redis_pool = None
        self.is_connected = False
    
    async def connect(self):
        """连接到Redis"""
        try:
            self.redis_pool = redis.ConnectionPool.from_url(
                self.redis_url,
                encoding="utf-8",
                decode_responses=True,
                max_connections=20
            )
            
            # 测试连接
            redis_client = redis.Redis(connection_pool=self.redis_pool)
            await redis_client.ping()
            self.is_connected = True
            logger.info("Connected to Redis successfully")
            
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            self.is_connected = False
    
    async def disconnect(self):
        """断开Redis连接"""
        if self.redis_pool:
            await self.redis_pool.disconnect()
            self.is_connected = False
            logger.info("Disconnected from Redis")
    
    async def get_client(self) -> redis.Redis:
        """获取Redis客户端"""
        if not self.is_connected:
            await self.connect()
        
        if not self.is_connected:
            raise Exception("Redis not connected")
        
        return redis.Redis(connection_pool=self.redis_pool)
    
    async def set(
        self,
        key: str,
        value: Any,
        expire: Optional[int] = None,
        prefix: str = "write_service"
    ) -> bool:
        """
        设置缓存值
        
        Args:
            key: 缓存键
            value: 缓存值
            expire: 过期时间（秒）
            prefix: 键前缀
            
        Returns:
            bool: 是否设置成功
        """
        try:
            client = await self.get_client()
            full_key = f"{prefix}:{key}"
            
            # 序列化值
            if isinstance(value, (dict, list)):
                serialized_value = json.dumps(value, ensure_ascii=False)
            else:
                serialized_value = str(value)
            
            if expire:
                result = await client.setex(full_key, expire, serialized_value)
            else:
                result = await client.set(full_key, serialized_value)
            
            return result
        except Exception as e:
            logger.error(f"Cache set error for key {key}: {e}")
            return False
    
    async def get(
        self,
        key: str,
        prefix: str = "write_service",
        default: Any = None
    ) -> Any:
        """
        获取缓存值
        
        Args:
            key: 缓存键
            prefix: 键前缀
            default: 默认值
            
        Returns:
            Any: 缓存值
        """
        try:
            client = await self.get_client()
            full_key = f"{prefix}:{key}"
            
            value = await client.get(full_key)
            if value is None:
                return default
            
            # 尝试反序列化JSON
            try:
                return json.loads(value)
            except json.JSONDecodeError:
                return value
                
        except Exception as e:
            logger.error(f"Cache get error for key {key}: {e}")
            return default
    
    async def delete(self, key: str, prefix: str = "write_service") -> bool:
        """删除缓存"""
        try:
            client = await self.get_client()
            full_key = f"{prefix}:{key}"
            result = await client.delete(full_key)
            return result > 0
        except Exception as e:
            logger.error(f"Cache delete error for key {key}: {e}")
            return False
    
    async def exists(self, key: str, prefix: str = "write_service") -> bool:
        """检查缓存是否存在"""
        try:
            client = await self.get_client()
            full_key = f"{prefix}:{key}"
            result = await client.exists(full_key)
            return result > 0
        except Exception as e:
            logger.error(f"Cache exists error for key {key}: {e}")
            return False
    
    async def increment(
        self,
        key: str,
        amount: int = 1,
        prefix: str = "write_service"
    ) -> int:
        """增加计数器"""
        try:
            client = await self.get_client()
            full_key = f"{prefix}:{key}"
            result = await client.incrby(full_key, amount)
            return result
        except Exception as e:
            logger.error(f"Cache increment error for key {key}: {e}")
            return 0
    
    async def set_hash(
        self,
        key: str,
        mapping: Dict[str, Any],
        prefix: str = "write_service"
    ) -> bool:
        """设置哈希表"""
        try:
            client = await self.get_client()
            full_key = f"{prefix}:{key}"
            
            # 序列化值
            serialized_mapping = {}
            for k, v in mapping.items():
                if isinstance(v, (dict, list)):
                    serialized_mapping[k] = json.dumps(v, ensure_ascii=False)
                else:
                    serialized_mapping[k] = str(v)
            
            result = await client.hset(full_key, mapping=serialized_mapping)
            return result > 0
        except Exception as e:
            logger.error(f"Cache set_hash error for key {key}: {e}")
            return False
    
    async def get_hash(
        self,
        key: str,
        field: Optional[str] = None,
        prefix: str = "write_service"
    ) -> Any:
        """获取哈希表值"""
        try:
            client = await self.get_client()
            full_key = f"{prefix}:{key}"
            
            if field:
                value = await client.hget(full_key, field)
                if value is None:
                    return None
                
                try:
                    return json.loads(value)
                except json.JSONDecodeError:
                    return value
            else:
                values = await client.hgetall(full_key)
                result = {}
                for k, v in values.items():
                    try:
                        result[k] = json.loads(v)
                    except json.JSONDecodeError:
                        result[k] = v
                return result
                
        except Exception as e:
            logger.error(f"Cache get_hash error for key {key}: {e}")
            return None

# 全局缓存管理器实例
cache_manager = CacheManager()

class LetterCacheService:
    """信件缓存服务"""
    
    # 缓存过期时间配置
    CACHE_TTL = {
        "letter_detail": 300,      # 信件详情缓存5分钟
        "letter_list": 60,         # 信件列表缓存1分钟
        "read_stats": 600,         # 阅读统计缓存10分钟
        "user_nickname": 1800,     # 用户昵称缓存30分钟
        "hot_letters": 3600,       # 热门信件缓存1小时
    }
    
    @staticmethod
    async def cache_letter_detail(letter_id: str, letter_data: Dict[str, Any]):
        """缓存信件详情"""
        key = f"letter:detail:{letter_id}"
        await cache_manager.set(
            key, letter_data, 
            expire=LetterCacheService.CACHE_TTL["letter_detail"]
        )
    
    @staticmethod
    async def get_cached_letter_detail(letter_id: str) -> Optional[Dict[str, Any]]:
        """获取缓存的信件详情"""
        key = f"letter:detail:{letter_id}"
        return await cache_manager.get(key)
    
    @staticmethod
    async def cache_user_letters(
        user_id: str,
        status: Optional[str],
        page: int,
        letters_data: Dict[str, Any]
    ):
        """缓存用户信件列表"""
        status_part = status or "all"
        key = f"user:letters:{user_id}:{status_part}:{page}"
        await cache_manager.set(
            key, letters_data,
            expire=LetterCacheService.CACHE_TTL["letter_list"]
        )
    
    @staticmethod
    async def get_cached_user_letters(
        user_id: str,
        status: Optional[str],
        page: int
    ) -> Optional[Dict[str, Any]]:
        """获取缓存的用户信件列表"""
        status_part = status or "all"
        key = f"user:letters:{user_id}:{status_part}:{page}"
        return await cache_manager.get(key)
    
    @staticmethod
    async def cache_read_stats(letter_id: str, stats_data: Dict[str, Any]):
        """缓存阅读统计"""
        key = f"letter:stats:{letter_id}"
        await cache_manager.set(
            key, stats_data,
            expire=LetterCacheService.CACHE_TTL["read_stats"]
        )
    
    @staticmethod
    async def get_cached_read_stats(letter_id: str) -> Optional[Dict[str, Any]]:
        """获取缓存的阅读统计"""
        key = f"letter:stats:{letter_id}"
        return await cache_manager.get(key)
    
    @staticmethod
    async def invalidate_letter_cache(letter_id: str):
        """失效信件相关缓存"""
        keys_to_delete = [
            f"letter:detail:{letter_id}",
            f"letter:stats:{letter_id}"
        ]
        
        for key in keys_to_delete:
            await cache_manager.delete(key)
    
    @staticmethod
    async def invalidate_user_cache(user_id: str):
        """失效用户相关缓存（粗粒度，实际可以更精细）"""
        # 这里简化处理，实际可以使用pattern匹配删除
        try:
            client = await cache_manager.get_client()
            pattern = f"write_service:user:letters:{user_id}:*"
            keys = await client.keys(pattern)
            if keys:
                await client.delete(*keys)
        except Exception as e:
            logger.error(f"Failed to invalidate user cache for {user_id}: {e}")

# 便捷函数
async def init_cache():
    """初始化缓存连接"""
    await cache_manager.connect()

async def cleanup_cache():
    """清理缓存连接"""
    await cache_manager.disconnect()

def get_cache_manager() -> CacheManager:
    """获取缓存管理器实例"""
    return cache_manager