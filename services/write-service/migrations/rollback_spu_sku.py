#!/usr/bin/env python3
"""
SPU+SKU模型回滚脚本

用于在需要时将数据从新的SPU+SKU模型回滚到旧的商品模型
仅在迁移出现问题时使用！

运行前请确保：
1. 旧表结构仍然存在
2. 已备份新模型数据
3. 了解回滚的数据损失风险
"""

import asyncio
import logging
from typing import Dict, List
from sqlalchemy import create_engine, text
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from datetime import datetime
import json

from app.core.config import settings

# 设置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class SPUSKURollback:
    """SPU+SKU模型回滚工具"""
    
    def __init__(self, database_url: str):
        self.engine = create_async_engine(database_url)
        self.SessionLocal = sessionmaker(
            bind=self.engine, 
            class_=AsyncSession, 
            expire_on_commit=False
        )
    
    async def backup_new_tables(self):
        """备份新表数据"""
        logger.info("💾 备份新表数据...")
        
        backup_tables = [
            'product_spu', 'product_sku', 'product_categories', 
            'product_brands', 'product_attributes', 'stock_records',
            'price_records', 'orders', 'order_items'
        ]
        
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        
        async with self.engine.begin() as conn:
            for table in backup_tables:
                backup_name = f"{table}_backup_{timestamp}"
                try:
                    await conn.execute(text(f"""
                        CREATE TABLE {backup_name} AS 
                        SELECT * FROM {table}
                    """))
                    logger.info(f"✅ 备份表 {table} → {backup_name}")
                except Exception as e:
                    logger.warning(f"⚠️ 备份表 {table} 失败: {e}")
        
        logger.info(f"✅ 新表数据备份完成，后缀: _{timestamp}")
        return timestamp
    
    async def restore_old_product_data(self, session: AsyncSession):
        """从SPU+SKU恢复到旧商品模型"""
        logger.info("🔄 从SPU+SKU恢复到旧商品模型...")
        
        # 检查是否存在旧表备份
        result = await session.execute(text("""
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_name = 'shop_products_backup'
            )
        """))
        has_backup = result.fetchone()[0]
        
        if has_backup:
            # 恢复旧商品数据
            logger.info("📦 恢复旧商品数据...")
            await session.execute(text("""
                INSERT INTO shop_products 
                SELECT * FROM shop_products_backup
                ON CONFLICT (id) DO NOTHING
            """))
            
            await session.commit()
            logger.info("✅ 旧商品数据恢复完成")
        else:
            logger.warning("⚠️ 未找到旧商品数据备份，无法恢复")
    
    async def restore_old_order_data(self, session: AsyncSession):
        """恢复旧订单数据"""
        logger.info("🛒 恢复旧订单数据...")
        
        # 检查备份表
        result = await session.execute(text("""
            SELECT EXISTS (
                SELECT FROM information_schema.tables 
                WHERE table_name IN ('shop_orders_backup', 'shop_order_items_backup')
            )
        """))
        has_backup = result.fetchone()[0]
        
        if has_backup:
            # 恢复订单数据
            await session.execute(text("""
                INSERT INTO shop_orders 
                SELECT * FROM shop_orders_backup
                ON CONFLICT (id) DO NOTHING
            """))
            
            # 恢复订单项数据
            await session.execute(text("""
                INSERT INTO shop_order_items 
                SELECT * FROM shop_order_items_backup
                ON CONFLICT (id) DO NOTHING
            """))
            
            await session.commit()
            logger.info("✅ 旧订单数据恢复完成")
        else:
            logger.warning("⚠️ 未找到旧订单数据备份，无法恢复")
    
    async def drop_new_tables(self):
        """删除新表（危险操作！）"""
        logger.warning("⚠️ 准备删除新表结构...")
        
        drop_tables = [
            'order_items', 'orders', 'payments', 'refunds', 'order_logs',
            'cart_items', 'carts', 'stock_records', 'price_records', 
            'product_attributes', 'product_sku', 'product_spu', 
            'product_brands', 'product_categories'
        ]
        
        async with self.engine.begin() as conn:
            for table in drop_tables:
                try:
                    await conn.execute(text(f"DROP TABLE IF EXISTS {table} CASCADE"))
                    logger.info(f"🗑️ 删除表 {table}")
                except Exception as e:
                    logger.warning(f"⚠️ 删除表 {table} 失败: {e}")
        
        logger.warning("⚠️ 新表结构已删除")
    
    async def verify_rollback(self, session: AsyncSession):
        """验证回滚结果"""
        logger.info("🔍 验证回滚结果...")
        
        # 检查旧表数据
        old_table_stats = {}
        old_tables = ['shop_products', 'shop_orders', 'shop_order_items', 'shop_categories']
        
        for table in old_tables:
            try:
                result = await session.execute(text(f"SELECT COUNT(*) FROM {table}"))
                count = result.fetchone()[0]
                old_table_stats[table] = count
            except Exception as e:
                old_table_stats[table] = f"Error: {e}"
        
        logger.info("📊 旧表数据统计:")
        for table, count in old_table_stats.items():
            logger.info(f"   {table}: {count}")
        
        # 检查新表是否已删除
        result = await session.execute(text("""
            SELECT table_name FROM information_schema.tables 
            WHERE table_name IN ('product_spu', 'product_sku', 'orders', 'order_items')
        """))
        remaining_tables = [row[0] for row in result.fetchall()]
        
        if remaining_tables:
            logger.warning(f"⚠️ 仍有新表存在: {', '.join(remaining_tables)}")
        else:
            logger.info("✅ 所有新表已清除")
    
    async def run_rollback(self):
        """运行完整回滚流程"""
        logger.warning("🚨 开始SPU+SKU模型回滚...")
        logger.warning("=" * 60)
        logger.warning("⚠️ 警告：此操作将删除所有新模型数据！")
        logger.warning("⚠️ 警告：请确保已备份重要数据！")
        logger.warning("=" * 60)
        
        start_time = datetime.now()
        
        try:
            # 1. 备份新表数据
            backup_suffix = await self.backup_new_tables()
            
            async with self.SessionLocal() as session:
                # 2. 恢复旧数据
                await self.restore_old_product_data(session)
                await self.restore_old_order_data(session)
            
            # 3. 删除新表
            await self.drop_new_tables()
            
            async with self.SessionLocal() as session:
                # 4. 验证回滚结果
                await self.verify_rollback(session)
            
            end_time = datetime.now()
            duration = end_time - start_time
            
            logger.info("=" * 60)
            logger.info(f"🔙 回滚完成！耗时: {duration}")
            logger.info(f"💾 新表数据已备份，后缀: _{backup_suffix}")
            logger.info("✅ 系统已回滚到旧模型")
            logger.info("💡 请测试系统功能，确保回滚成功")
            
        except Exception as e:
            logger.error(f"❌ 回滚失败: {e}")
            raise
        
        finally:
            await self.engine.dispose()

async def main():
    """主函数"""
    print("SPU+SKU模型回滚工具")
    print("=" * 60)
    print("🚨 危险操作警告！")
    print("⚠️  此工具将删除所有新的SPU+SKU模型数据！")
    print("⚠️  此操作不可逆，请确保：")
    print("   1. 已备份所有重要数据")
    print("   2. 旧表结构和数据仍然存在") 
    print("   3. 确实需要回滚到旧模型")
    print()
    
    confirm1 = input("确认了解风险并继续？(y/N): ")
    if confirm1.lower() != 'y':
        print("❌ 回滚已取消")
        return
    
    confirm2 = input("再次确认执行回滚？输入 'ROLLBACK' 继续: ")
    if confirm2 != 'ROLLBACK':
        print("❌ 回滚已取消")
        return
    
    # 使用配置中的数据库URL
    database_url = settings.DATABASE_URL
    if database_url.startswith("postgresql://"):
        database_url = database_url.replace("postgresql://", "postgresql+asyncpg://")
    
    rollback_tool = SPUSKURollback(database_url)
    await rollback_tool.run_rollback()

if __name__ == "__main__":
    asyncio.run(main())