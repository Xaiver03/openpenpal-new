from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
from contextlib import asynccontextmanager
import os
from app.core.config import settings
from app.api.letters import router as letters_router
from app.api.plaza import router as plaza_router
from app.api.museum import router as museum_router
from app.api.shop import router as shop_router
from app.api.drafts import router as drafts_router
from app.api.analytics import router as analytics_router
from app.api.batch import router as batch_router
from app.api.upload import router as upload_router
from app.api.notifications import router as notifications_router
from app.api.postcode import router as postcode_router
from app.utils.cache_manager import init_cache, cleanup_cache
from app.utils.websocket_client import init_websocket, cleanup_websocket
from app.middleware.rate_limiter import create_rate_limiter
from app.middleware.error_handler import ErrorHandlerMiddleware, create_error_handler
from app.utils.token_blacklist import get_token_blacklist

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时初始化
    try:
        await init_cache()
        await init_websocket()
        
        # 启动令牌黑名单清理任务
        blacklist = get_token_blacklist()
        if hasattr(blacklist, 'start_cleanup'):
            await blacklist.start_cleanup()
        
        # 启动速率限制器清理任务
        rate_limiter = getattr(app.state, 'rate_limiter', None)
        if rate_limiter and hasattr(rate_limiter, 'limiter') and rate_limiter.limiter:
            await rate_limiter.limiter.start_cleanup_task()
        
        print("✅ All services initialized (Cache, WebSocket, Security)")
    except Exception as e:
        print(f"⚠️  Failed to initialize services: {e}")
    
    yield
    
    # 关闭时清理
    try:
        await cleanup_cache()
        await cleanup_websocket()
        
        # 停止安全相关清理任务
        blacklist = get_token_blacklist()
        if hasattr(blacklist, 'stop_cleanup'):
            await blacklist.stop_cleanup()
        
        rate_limiter = getattr(app.state, 'rate_limiter', None)
        if rate_limiter and hasattr(rate_limiter, 'limiter') and rate_limiter.limiter:
            await rate_limiter.limiter.stop_cleanup_task()
        
        print("✅ All services cleaned up")
    except Exception as e:
        print(f"⚠️  Failed to cleanup services: {e}")

app = FastAPI(
    title="OpenPenPal Write Service",
    description="信件创建和管理服务",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan
)

# 安全中间件配置 - 按优先级顺序添加

# 1. 错误处理中间件（最先执行，最后返回）
error_handler = create_error_handler(debug=settings.debug_mode)
app.add_middleware(ErrorHandlerMiddleware, debug=settings.debug_mode)

# 2. 速率限制中间件
rate_limit_middleware = create_rate_limiter(
    max_requests=settings.max_requests_per_minute,
    window_seconds=60,
    enabled=settings.enable_rate_limiting
)
app.middleware("http")(rate_limit_middleware)
app.state.rate_limiter = rate_limit_middleware

# 3. CORS配置
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", settings.frontend_url],  # 前端地址
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 注册路由
app.include_router(letters_router, prefix="/api/letters", tags=["letters"])
app.include_router(plaza_router, prefix="/api/plaza", tags=["plaza"])
app.include_router(museum_router, prefix="/api/museum", tags=["museum"])
app.include_router(shop_router, tags=["shop"])
app.include_router(drafts_router, prefix="/api", tags=["drafts"])
app.include_router(analytics_router, prefix="/api/analytics", tags=["analytics"])
app.include_router(batch_router, prefix="/api/batch", tags=["batch"])
app.include_router(upload_router, tags=["upload"])
app.include_router(notifications_router, tags=["notifications"])
app.include_router(postcode_router, prefix="/api/v1", tags=["postcode"])

# 地址搜索兼容性路由
from fastapi import APIRouter
address_compat_router = APIRouter()

@address_compat_router.get("/search")
async def address_search_compat(*args, **kwargs):
    """地址搜索兼容性路由，转发到postcode搜索"""
    from app.api.postcode import search_addresses
    return await search_addresses(*args, **kwargs)

app.include_router(address_compat_router, prefix="/api/v1/address", tags=["address-compat"])

# 新增商品管理相关路由
from app.api.v1.categories import router as categories_router
from app.api.v1.product_attributes import router as attributes_router
from app.api.v1.rbac import router as rbac_router
from app.api.v1.pricing import router as pricing_router
from app.api.test_admin import router as test_router
app.include_router(categories_router, prefix="/api/v1", tags=["categories"])
app.include_router(attributes_router, prefix="/api/v1", tags=["attributes"])
app.include_router(rbac_router, prefix="/api/v1", tags=["rbac"])
app.include_router(pricing_router, prefix="/api/v1", tags=["pricing"])
app.include_router(test_router, prefix="/api/v1", tags=["test"])

# 挂载静态文件
static_dir = os.path.join(os.path.dirname(os.path.dirname(__file__)), "static")
if os.path.exists(static_dir):
    app.mount("/static", StaticFiles(directory=static_dir), name="static")

@app.get("/admin")
async def admin_panel():
    """商城管理后台入口"""
    static_file = os.path.join(static_dir, "admin.html")
    if os.path.exists(static_file):
        return FileResponse(static_file)
    else:
        return {"message": "Admin panel not found", "redirect": "/docs"}

@app.get("/health")
async def health_check():
    """健康检查端点"""
    from app.utils.security import SecurityManager
    
    # 检查JWT密钥安全性
    jwt_secure = SecurityManager.validate_jwt_secret(settings.jwt_secret)
    
    # 获取安全状态
    blacklist = get_token_blacklist()
    blacklist_count = blacklist.get_blacklist_count() if hasattr(blacklist, 'get_blacklist_count') else 0
    
    # 安全状态总览
    security_status = {
        "jwt_secure": jwt_secure,
        "rate_limiting": settings.enable_rate_limiting,
        "https_enabled": settings.enable_https,
        "xss_protection": settings.enable_xss_protection,
        "content_filter": settings.enable_content_filter,
        "error_handler": True,  # 已启用错误处理中间件
        "blacklist_tokens": blacklist_count,
        "debug_mode": settings.debug_mode
    }
    
    # 计算安全评分
    security_score = sum([
        jwt_secure,
        settings.enable_rate_limiting,
        settings.enable_https,
        settings.enable_xss_protection,
        settings.enable_content_filter,
        not settings.debug_mode  # 生产环境应该关闭调试模式
    ]) / 6 * 100
    
    return {
        "code": 0,
        "msg": "Write service is healthy",
        "data": {
            "service": "write-service",
            "version": "1.0.0", 
            "status": "running",
            "security": security_status,
            "security_score": f"{security_score:.1f}%",
            "timestamp": f"{__import__('datetime').datetime.utcnow().isoformat()}Z"
        }
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)