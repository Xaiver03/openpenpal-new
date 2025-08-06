from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from app.core.config import settings

# 创建数据库引擎
engine = create_engine(
    settings.database_url,
    echo=True,  # 开发环境下显示SQL
    pool_pre_ping=True
)

# 创建会话工厂
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# 创建基础模型类
Base = declarative_base()

def get_db():
    """获取数据库会话的依赖注入函数"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

def create_tables():
    """创建所有数据表"""
    Base.metadata.create_all(bind=engine)

# 异步数据库设置
def get_async_database_url(url: str) -> str:
    """转换数据库URL为异步驱动URL"""
    if url.startswith("postgresql+psycopg2://"):
        return url.replace("postgresql+psycopg2://", "postgresql+asyncpg://")
    elif url.startswith("postgresql://"):
        return url.replace("postgresql://", "postgresql+asyncpg://")
    elif url.startswith("sqlite://"):
        return url.replace("sqlite://", "sqlite+aiosqlite://")
    return url

async_engine = create_async_engine(
    get_async_database_url(settings.database_url),
    echo=True,
    pool_pre_ping=True
)

AsyncSessionLocal = async_sessionmaker(
    async_engine,
    class_=AsyncSession,
    expire_on_commit=False
)

async def get_async_session() -> AsyncSession:
    """获取异步数据库会话的依赖注入函数"""
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()