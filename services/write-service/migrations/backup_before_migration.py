#!/usr/bin/env python3
"""
迁移前数据库备份脚本

在执行SPU+SKU模型迁移前，备份所有相关的原始数据
确保在迁移失败时可以完全恢复到原始状态

备份策略：
1. 创建旧表的完整副本
2. 导出数据到文件
3. 验证备份完整性
"""

import asyncio
import logging
import json
import os
from typing import Dict, List
from sqlalchemy import create_engine, text
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from datetime import datetime

from app.core.config import settings

# 设置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class PreMigrationBackup:
    """迁移前备份工具"""
    
    def __init__(self, database_url: str):
        self.engine = create_async_engine(database_url)
        self.SessionLocal = sessionmaker(
            bind=self.engine, 
            class_=AsyncSession, 
            expire_on_commit=False
        )
        self.timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        self.backup_dir = f"backups/migration_backup_{self.timestamp}"
    
    def create_backup_directory(self):
        """创建备份目录"""
        os.makedirs(self.backup_dir, exist_ok=True)
        logger.info(f"📁 备份目录: {self.backup_dir}")
    
    async def backup_table_structure(self, session: AsyncSession, table_name: str):
        """备份表结构"""
        logger.info(f"🏗️ 备份表结构: {table_name}")
        
        # 获取表结构
        result = await session.execute(text(f"""
            SELECT column_name, data_type, is_nullable, column_default
            FROM information_schema.columns
            WHERE table_name = '{table_name}'
            ORDER BY ordinal_position
        """))
        
        columns = []
        for row in result.fetchall():
            columns.append({
                'name': row.column_name,
                'type': row.data_type,
                'nullable': row.is_nullable,
                'default': row.column_default
            })
        
        # 保存结构信息
        structure_file = os.path.join(self.backup_dir, f"{table_name}_structure.json")
        with open(structure_file, 'w', encoding='utf-8') as f:
            json.dump({
                'table_name': table_name,
                'columns': columns,
                'backup_time': self.timestamp
            }, f, indent=2, ensure_ascii=False, default=str)
        
        return len(columns)
    
    async def backup_table_data(self, session: AsyncSession, table_name: str):
        """备份表数据"""
        logger.info(f"💾 备份表数据: {table_name}")
        
        try:
            # 获取所有数据
            result = await session.execute(text(f"SELECT * FROM {table_name}"))
            rows = result.fetchall()
            columns = result.keys()
            
            # 转换为字典格式
            data = []
            for row in rows:
                row_dict = {}
                for i, col in enumerate(columns):
                    value = row[i]
                    # 处理特殊类型
                    if isinstance(value, datetime):
                        value = value.isoformat()
                    row_dict[col] = value
                data.append(row_dict)
            
            # 保存数据
            data_file = os.path.join(self.backup_dir, f"{table_name}_data.json")
            with open(data_file, 'w', encoding='utf-8') as f:
                json.dump({
                    'table_name': table_name,
                    'row_count': len(data),
                    'columns': list(columns),
                    'data': data,
                    'backup_time': self.timestamp
                }, f, indent=2, ensure_ascii=False, default=str)
            
            return len(data)
        
        except Exception as e:
            logger.warning(f"⚠️ 备份表 {table_name} 数据失败: {e}")
            return 0
    
    async def create_table_backup_copy(self, table_name: str):
        """创建表的副本"""
        backup_table_name = f"{table_name}_backup_{self.timestamp}"
        
        try:
            async with self.engine.begin() as conn:
                # 创建表副本
                await conn.execute(text(f"""
                    CREATE TABLE {backup_table_name} AS 
                    SELECT * FROM {table_name}
                """))
                
                logger.info(f"✅ 创建表副本: {table_name} → {backup_table_name}")
                return backup_table_name
        
        except Exception as e:
            logger.error(f"❌ 创建表副本失败 {table_name}: {e}")
            return None
    
    async def backup_indexes_and_constraints(self, session: AsyncSession, table_name: str):
        """备份索引和约束信息"""
        logger.info(f"🔗 备份索引和约束: {table_name}")
        
        # 获取索引信息
        indexes_result = await session.execute(text(f"""
            SELECT indexname, indexdef 
            FROM pg_indexes 
            WHERE tablename = '{table_name}'
        """))
        indexes = [{'name': row.indexname, 'definition': row.indexdef} 
                  for row in indexes_result.fetchall()]
        
        # 获取约束信息
        constraints_result = await session.execute(text(f"""
            SELECT conname, pg_get_constraintdef(c.oid) as definition
            FROM pg_constraint c
            JOIN pg_class t ON c.conrelid = t.oid
            WHERE t.relname = '{table_name}'
        """))
        constraints = [{'name': row.conname, 'definition': row.definition}
                      for row in constraints_result.fetchall()]
        
        # 保存索引和约束信息
        metadata_file = os.path.join(self.backup_dir, f"{table_name}_metadata.json")
        with open(metadata_file, 'w', encoding='utf-8') as f:
            json.dump({
                'table_name': table_name,
                'indexes': indexes,
                'constraints': constraints,
                'backup_time': self.timestamp
            }, f, indent=2, ensure_ascii=False)
        
        return len(indexes) + len(constraints)
    
    async def verify_backup(self, session: AsyncSession, table_name: str):
        """验证备份完整性"""
        try:
            # 检查原表行数
            result = await session.execute(text(f"SELECT COUNT(*) FROM {table_name}"))
            original_count = result.fetchone()[0]
            
            # 检查备份表行数
            backup_table = f"{table_name}_backup_{self.timestamp}"
            result = await session.execute(text(f"SELECT COUNT(*) FROM {backup_table}"))
            backup_count = result.fetchone()[0]
            
            # 检查数据文件
            data_file = os.path.join(self.backup_dir, f"{table_name}_data.json")
            if os.path.exists(data_file):
                with open(data_file, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    file_count = data['row_count']
            else:
                file_count = 0
            
            success = (original_count == backup_count == file_count)
            
            if success:
                logger.info(f"✅ 备份验证成功: {table_name} ({original_count} 行)")
            else:
                logger.warning(f"⚠️ 备份验证失败: {table_name}")
                logger.warning(f"   原表: {original_count}, 备份表: {backup_count}, 文件: {file_count}")
            
            return success
        
        except Exception as e:
            logger.error(f"❌ 备份验证失败 {table_name}: {e}")
            return False
    
    async def create_backup_manifest(self, backup_info: Dict):
        """创建备份清单"""
        manifest = {
            'backup_time': self.timestamp,
            'backup_directory': self.backup_dir,
            'database_url': settings.DATABASE_URL.split('@')[1] if '@' in settings.DATABASE_URL else 'masked',
            'tables_backed_up': backup_info,
            'total_tables': len(backup_info),
            'total_rows': sum(info.get('row_count', 0) for info in backup_info.values()),
            'backup_files': []
        }
        
        # 列出所有备份文件
        for root, dirs, files in os.walk(self.backup_dir):
            for file in files:
                file_path = os.path.join(root, file)
                manifest['backup_files'].append({
                    'name': file,
                    'path': file_path,
                    'size': os.path.getsize(file_path)
                })
        
        # 保存清单
        manifest_file = os.path.join(self.backup_dir, 'backup_manifest.json')
        with open(manifest_file, 'w', encoding='utf-8') as f:
            json.dump(manifest, f, indent=2, ensure_ascii=False, default=str)
        
        logger.info(f"📋 备份清单已创建: {manifest_file}")
        return manifest_file
    
    async def run_backup(self):
        """运行完整备份流程"""
        logger.info("💾 开始迁移前数据库备份...")
        logger.info("=" * 60)
        
        start_time = datetime.now()
        self.create_backup_directory()
        
        # 需要备份的表
        tables_to_backup = [
            'shop_products', 
            'shop_categories', 
            'shop_orders', 
            'shop_order_items',
            'shop_carts',
            'shop_cart_items', 
            'shop_product_reviews',
            'shop_product_favorites'
        ]
        
        backup_info = {}
        successful_backups = 0
        
        try:
            async with self.SessionLocal() as session:
                for table_name in tables_to_backup:
                    logger.info(f"\n📦 备份表: {table_name}")
                    
                    try:
                        # 备份表结构
                        column_count = await self.backup_table_structure(session, table_name)
                        
                        # 备份表数据
                        row_count = await self.backup_table_data(session, table_name)
                        
                        # 创建表副本
                        backup_table = await self.create_table_backup_copy(table_name)
                        
                        # 备份索引和约束
                        metadata_count = await self.backup_indexes_and_constraints(session, table_name)
                        
                        # 验证备份
                        verification_success = await self.verify_backup(session, table_name)
                        
                        backup_info[table_name] = {
                            'row_count': row_count,
                            'column_count': column_count,
                            'metadata_count': metadata_count,
                            'backup_table': backup_table,
                            'verification_success': verification_success,
                            'backup_time': datetime.now().isoformat()
                        }
                        
                        if verification_success:
                            successful_backups += 1
                            logger.info(f"✅ {table_name} 备份成功")
                        else:
                            logger.warning(f"⚠️ {table_name} 备份验证失败")
                    
                    except Exception as e:
                        logger.error(f"❌ 备份表 {table_name} 失败: {e}")
                        backup_info[table_name] = {
                            'error': str(e),
                            'backup_time': datetime.now().isoformat()
                        }
            
            # 创建备份清单
            manifest_file = await self.create_backup_manifest(backup_info)
            
            end_time = datetime.now()
            duration = end_time - start_time
            
            logger.info("\n" + "=" * 60)
            logger.info("📊 备份结果汇总:")
            logger.info(f"   总表数: {len(tables_to_backup)}")
            logger.info(f"   成功备份: {successful_backups}")
            logger.info(f"   失败备份: {len(tables_to_backup) - successful_backups}")
            logger.info(f"   总行数: {sum(info.get('row_count', 0) for info in backup_info.values())}")
            logger.info(f"   备份目录: {self.backup_dir}")
            logger.info(f"   耗时: {duration}")
            
            if successful_backups == len(tables_to_backup):
                logger.info("🎉 所有表备份成功！可以安全执行迁移")
                return True
            else:
                logger.warning("⚠️ 部分表备份失败，建议检查问题后重新备份")
                return False
        
        except Exception as e:
            logger.error(f"❌ 备份过程失败: {e}")
            return False
        
        finally:
            await self.engine.dispose()

async def main():
    """主函数"""
    print("迁移前数据库备份工具")
    print("=" * 60)
    print("📋 本工具将备份以下数据：")
    print("   - 所有商品相关表的完整结构和数据")
    print("   - 所有订单相关表的完整结构和数据")
    print("   - 索引、约束等元数据信息")
    print("   - 创建表副本用于快速恢复")
    print()
    
    # 使用配置中的数据库URL
    database_url = settings.DATABASE_URL
    if database_url.startswith("postgresql://"):
        database_url = database_url.replace("postgresql://", "postgresql+asyncpg://")
    
    backup_tool = PreMigrationBackup(database_url)
    success = await backup_tool.run_backup()
    
    if success:
        print("\n✅ 备份完成！现在可以安全执行迁移")
        print(f"💡 备份位置: {backup_tool.backup_dir}")
        print("💡 如需恢复，请使用 rollback_spu_sku.py 脚本")
    else:
        print("\n❌ 备份未完全成功，请检查错误后重新运行")
        print("⚠️ 建议在备份成功后再执行迁移")

if __name__ == "__main__":
    asyncio.run(main())