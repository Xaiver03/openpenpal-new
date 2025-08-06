import asyncio
from datetime import datetime, timedelta
from typing import Set, Optional
import hashlib

class TokenBlacklist:
    """JWT令牌黑名单管理器"""
    
    def __init__(self):
        self._blacklisted_tokens: Set[str] = set()
        self._cleanup_task: Optional[asyncio.Task] = None
        self._start_cleanup_task()
    
    def _hash_token(self, token: str) -> str:
        """对令牌进行哈希以节省内存"""
        return hashlib.sha256(token.encode()).hexdigest()
    
    def add_token(self, token: str, expires_at: Optional[datetime] = None) -> None:
        """
        将令牌添加到黑名单
        
        Args:
            token: JWT令牌
            expires_at: 令牌过期时间（用于自动清理）
        """
        token_hash = self._hash_token(token)
        self._blacklisted_tokens.add(token_hash)
        print(f"🚫 Token added to blacklist: {token_hash[:16]}...")
    
    def is_blacklisted(self, token: str) -> bool:
        """
        检查令牌是否在黑名单中
        
        Args:
            token: JWT令牌
            
        Returns:
            bool: 是否被列入黑名单
        """
        token_hash = self._hash_token(token)
        return token_hash in self._blacklisted_tokens
    
    def remove_token(self, token: str) -> bool:
        """
        从黑名单中移除令牌
        
        Args:
            token: JWT令牌
            
        Returns:
            bool: 是否成功移除
        """
        token_hash = self._hash_token(token)
        if token_hash in self._blacklisted_tokens:
            self._blacklisted_tokens.remove(token_hash)
            return True
        return False
    
    def clear_all(self) -> None:
        """清空所有黑名单令牌"""
        self._blacklisted_tokens.clear()
        print("🧹 All tokens cleared from blacklist")
    
    def get_blacklist_count(self) -> int:
        """获取黑名单令牌数量"""
        return len(self._blacklisted_tokens)
    
    def _start_cleanup_task(self) -> None:
        """启动定期清理任务"""
        async def cleanup_expired_tokens():
            while True:
                try:
                    # 每小时清理一次过期令牌
                    await asyncio.sleep(3600)
                    # 简单清理：由于令牌默认30分钟过期，每小时清理一次已足够
                    # 在实际生产环境中，应该记录令牌的过期时间
                    if len(self._blacklisted_tokens) > 10000:  # 如果黑名单过大
                        self._blacklisted_tokens.clear()
                        print("🧹 Blacklist cleared due to size limit")
                except Exception as e:
                    print(f"❌ Error in blacklist cleanup: {e}")
        
        # 不在这里启动任务，而是让主应用程序管理
        pass
    
    async def start_cleanup(self) -> None:
        """启动清理任务（由应用程序调用）"""
        if self._cleanup_task is None or self._cleanup_task.done():
            async def cleanup_loop():
                while True:
                    try:
                        await asyncio.sleep(3600)  # 每小时清理
                        # 简化的清理逻辑
                        if len(self._blacklisted_tokens) > 5000:
                            old_size = len(self._blacklisted_tokens)
                            # 清空一半最老的令牌（简化处理）
                            tokens_list = list(self._blacklisted_tokens)
                            self._blacklisted_tokens = set(tokens_list[len(tokens_list)//2:])
                            print(f"🧹 Cleaned blacklist: {old_size} -> {len(self._blacklisted_tokens)}")
                    except Exception as e:
                        print(f"❌ Blacklist cleanup error: {e}")
            
            self._cleanup_task = asyncio.create_task(cleanup_loop())
    
    async def stop_cleanup(self) -> None:
        """停止清理任务"""
        if self._cleanup_task and not self._cleanup_task.done():
            self._cleanup_task.cancel()
            try:
                await self._cleanup_task
            except asyncio.CancelledError:
                pass
            self._cleanup_task = None

# 全局黑名单实例
token_blacklist = TokenBlacklist()

# 高级黑名单管理（支持Redis）
class RedisTokenBlacklist:
    """基于Redis的令牌黑名单（生产环境推荐）"""
    
    def __init__(self, redis_client=None):
        self.redis = redis_client
        self.key_prefix = "jwt_blacklist:"
    
    async def add_token(self, token: str, expires_at: Optional[datetime] = None) -> None:
        """将令牌添加到Redis黑名单"""
        if not self.redis:
            return
        
        token_hash = hashlib.sha256(token.encode()).hexdigest()
        key = f"{self.key_prefix}{token_hash}"
        
        # 设置过期时间
        if expires_at:
            ttl = int((expires_at - datetime.utcnow()).total_seconds())
            if ttl > 0:
                await self.redis.setex(key, ttl, "1")
        else:
            # 默认2小时过期
            await self.redis.setex(key, 7200, "1")
    
    async def is_blacklisted(self, token: str) -> bool:
        """检查令牌是否在Redis黑名单中"""
        if not self.redis:
            return False
        
        token_hash = hashlib.sha256(token.encode()).hexdigest()
        key = f"{self.key_prefix}{token_hash}"
        
        result = await self.redis.get(key)
        return result is not None
    
    async def remove_token(self, token: str) -> bool:
        """从Redis黑名单中移除令牌"""
        if not self.redis:
            return False
        
        token_hash = hashlib.sha256(token.encode()).hexdigest()
        key = f"{self.key_prefix}{token_hash}"
        
        result = await self.redis.delete(key)
        return result > 0

# 根据配置选择黑名单实现
def get_token_blacklist():
    """获取令牌黑名单实例"""
    # 如果有Redis配置，优先使用Redis黑名单
    try:
        from app.core.config import settings
        if settings.redis_url and settings.redis_url != "redis://localhost:6379/0":
            # 在实际使用中，这里应该初始化Redis客户端
            # return RedisTokenBlacklist(redis_client)
            pass
    except:
        pass
    
    # 否则使用内存黑名单
    return token_blacklist