import time
import asyncio
from collections import defaultdict, deque
from typing import Dict, Deque, Tuple
from fastapi import Request, HTTPException, status
from fastapi.responses import JSONResponse
import hashlib

class InMemoryRateLimiter:
    """内存版本的速率限制器"""
    
    def __init__(self, max_requests: int = 60, window_seconds: int = 60):
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        # 存储格式: {client_id: deque(timestamps)}
        self.requests: Dict[str, Deque[float]] = defaultdict(deque)
        self.cleanup_task = None
    
    def _get_client_id(self, request: Request) -> str:
        """获取客户端标识"""
        # 优先使用用户ID（如果已认证）
        user_id = getattr(request.state, 'user_id', None)
        if user_id:
            return f"user:{user_id}"
        
        # 使用IP地址
        client_ip = request.client.host if request.client else "unknown"
        
        # 考虑代理情况
        forwarded_for = request.headers.get("X-Forwarded-For")
        if forwarded_for:
            client_ip = forwarded_for.split(",")[0].strip()
        
        real_ip = request.headers.get("X-Real-IP")
        if real_ip:
            client_ip = real_ip
        
        return f"ip:{client_ip}"
    
    def _cleanup_old_requests(self, client_id: str, current_time: float):
        """清理过期的请求记录"""
        cutoff_time = current_time - self.window_seconds
        requests_deque = self.requests[client_id]
        
        while requests_deque and requests_deque[0] <= cutoff_time:
            requests_deque.popleft()
    
    def is_allowed(self, request: Request) -> Tuple[bool, Dict[str, any]]:
        """
        检查请求是否被允许
        
        Returns:
            Tuple[bool, Dict]: (是否允许, 限制信息)
        """
        client_id = self._get_client_id(request)
        current_time = time.time()
        
        # 清理过期请求
        self._cleanup_old_requests(client_id, current_time)
        
        requests_deque = self.requests[client_id]
        current_requests = len(requests_deque)
        
        # 检查是否超过限制
        if current_requests >= self.max_requests:
            # 计算重置时间
            reset_time = requests_deque[0] + self.window_seconds
            retry_after = max(1, int(reset_time - current_time))
            
            return False, {
                "limit": self.max_requests,
                "remaining": 0,
                "reset": int(reset_time),
                "retry_after": retry_after
            }
        
        # 记录当前请求
        requests_deque.append(current_time)
        
        # 计算剩余请求数
        remaining = self.max_requests - len(requests_deque)
        next_reset = current_time + self.window_seconds
        
        return True, {
            "limit": self.max_requests,
            "remaining": remaining,
            "reset": int(next_reset),
            "retry_after": 0
        }
    
    async def start_cleanup_task(self):
        """启动定期清理任务"""
        async def cleanup_loop():
            while True:
                try:
                    await asyncio.sleep(300)  # 每5分钟清理一次
                    current_time = time.time()
                    cutoff_time = current_time - self.window_seconds * 2  # 清理2倍窗口时间之前的数据
                    
                    # 清理空的或过期的记录
                    clients_to_remove = []
                    for client_id, requests_deque in self.requests.items():
                        # 清理过期请求
                        while requests_deque and requests_deque[0] <= cutoff_time:
                            requests_deque.popleft()
                        
                        # 如果没有请求记录，标记删除
                        if not requests_deque:
                            clients_to_remove.append(client_id)
                    
                    # 删除空的客户端记录
                    for client_id in clients_to_remove:
                        del self.requests[client_id]
                    
                    print(f"🧹 Rate limiter cleanup: {len(clients_to_remove)} clients removed, {len(self.requests)} active")
                    
                except Exception as e:
                    print(f"❌ Rate limiter cleanup error: {e}")
        
        if self.cleanup_task is None or self.cleanup_task.done():
            self.cleanup_task = asyncio.create_task(cleanup_loop())
    
    async def stop_cleanup_task(self):
        """停止清理任务"""
        if self.cleanup_task and not self.cleanup_task.done():
            self.cleanup_task.cancel()
            try:
                await self.cleanup_task
            except asyncio.CancelledError:
                pass

class RateLimitMiddleware:
    """速率限制中间件"""
    
    def __init__(self, max_requests: int = 60, window_seconds: int = 60, enabled: bool = True):
        self.enabled = enabled
        self.limiter = InMemoryRateLimiter(max_requests, window_seconds) if enabled else None
        
        # 白名单路径（不受速率限制）
        self.whitelist_paths = {
            "/health",
            "/docs",
            "/redoc", 
            "/openapi.json"
        }
        
        # 严格限制的路径（更低的限制）
        self.strict_paths = {
            "/api/letters": (10, 60),  # 创建信件：每分钟10次
            "/api/shop/orders": (5, 60),  # 创建订单：每分钟5次
            "/api/plaza/posts": (10, 60),  # 发帖：每分钟10次
        }
    
    async def __call__(self, request: Request, call_next):
        """中间件处理函数"""
        if not self.enabled or not self.limiter:
            return await call_next(request)
        
        # 检查白名单
        if request.url.path in self.whitelist_paths:
            return await call_next(request)
        
        # 检查是否需要严格限制
        limiter = self.limiter
        for path_prefix, (max_req, window) in self.strict_paths.items():
            if request.url.path.startswith(path_prefix) and request.method == "POST":
                limiter = InMemoryRateLimiter(max_req, window)
                break
        
        # 检查速率限制
        allowed, limit_info = limiter.is_allowed(request)
        
        if not allowed:
            # 返回429错误
            return JSONResponse(
                status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                content={
                    "code": 429,
                    "msg": "请求过于频繁，请稍后再试",
                    "data": {
                        "limit": limit_info["limit"],
                        "remaining": limit_info["remaining"],
                        "reset": limit_info["reset"],
                        "retry_after": limit_info["retry_after"]
                    }
                },
                headers={
                    "X-RateLimit-Limit": str(limit_info["limit"]),
                    "X-RateLimit-Remaining": str(limit_info["remaining"]),
                    "X-RateLimit-Reset": str(limit_info["reset"]),
                    "Retry-After": str(limit_info["retry_after"])
                }
            )
        
        # 添加速率限制头部信息
        response = await call_next(request)
        response.headers["X-RateLimit-Limit"] = str(limit_info["limit"])
        response.headers["X-RateLimit-Remaining"] = str(limit_info["remaining"])
        response.headers["X-RateLimit-Reset"] = str(limit_info["reset"])
        
        return response

# 全局速率限制器实例
rate_limiter = None

def create_rate_limiter(max_requests: int = 60, window_seconds: int = 60, enabled: bool = True) -> RateLimitMiddleware:
    """创建速率限制中间件"""
    global rate_limiter
    rate_limiter = RateLimitMiddleware(max_requests, window_seconds, enabled)
    return rate_limiter

def get_rate_limiter() -> RateLimitMiddleware:
    """获取全局速率限制器"""
    return rate_limiter