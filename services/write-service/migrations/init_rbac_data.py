#!/usr/bin/env python3
"""
RBAC系统初始化数据脚本

初始化商城管理后台的角色权限数据：
1. 创建基础角色
2. 初始化菜单结构
3. 分配角色权限
4. 创建默认管理员用户
"""

import asyncio
import logging
from datetime import datetime
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
import uuid

from app.core.config import settings
from app.models.rbac import (
    SysMenu, SysRole, SysUser, SysRoleMenu, SysUserRole,
    BusinessType, MenuType, PermissionType
)
from app.core.database import Base

# 设置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class RBACInitializer:
    """RBAC系统初始化工具"""
    
    def __init__(self, database_url: str):
        self.engine = create_async_engine(database_url)
        self.SessionLocal = sessionmaker(
            bind=self.engine,
            class_=AsyncSession,
            expire_on_commit=False
        )
    
    def generate_id(self, prefix: str = "") -> str:
        """生成唯一ID"""
        return f"{prefix}{uuid.uuid4().hex[:16].upper()}"
    
    async def create_tables(self):
        """创建RBAC相关表"""
        logger.info("🏗️ 创建RBAC数据表...")
        
        async with self.engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)
        
        logger.info("✅ RBAC数据表创建完成")
    
    async def init_roles(self, session: AsyncSession):
        """初始化基础角色"""
        logger.info("👥 初始化基础角色...")
        
        roles_data = [
            {
                "role_name": "平台超级管理员",
                "role_code": "PLATFORM_SUPER_ADMIN", 
                "role_desc": "平台超级管理员，拥有所有权限",
                "biz_type": BusinessType.PLATFORM.value,
                "is_admin": True,
                "status": 1,
                "sort_order": 1,
                "create_by": "system",
                "update_by": "system"
            },
            {
                "role_name": "平台管理员",
                "role_code": "PLATFORM_ADMIN",
                "role_desc": "平台管理员，管理商户和平台配置",
                "biz_type": BusinessType.PLATFORM.value,
                "is_admin": False,
                "status": 1,
                "sort_order": 2,
                "create_by": "system",
                "update_by": "system"
            },
            {
                "role_name": "商城管理员",
                "role_code": "SHOP_ADMIN",
                "role_desc": "商城管理员，管理店铺所有功能",
                "biz_type": BusinessType.SHOP.value,
                "is_admin": False,
                "status": 1,
                "sort_order": 10,
                "create_by": "system",
                "update_by": "system"
            },
            {
                "role_name": "商城运营",
                "role_code": "SHOP_OPERATOR",
                "role_desc": "商城运营人员，管理商品和订单",
                "biz_type": BusinessType.SHOP.value,
                "is_admin": False,
                "status": 1,
                "sort_order": 11,
                "create_by": "system",
                "update_by": "system"
            },
            {
                "role_name": "客服人员",
                "role_code": "CUSTOMER_SERVICE",
                "role_desc": "客服人员，处理订单和售后",
                "biz_type": BusinessType.COMMON.value,
                "is_admin": False,
                "status": 1,
                "sort_order": 20,
                "create_by": "system",
                "update_by": "system"
            },
            {
                "role_name": "财务人员",
                "role_code": "FINANCE_STAFF",
                "role_desc": "财务人员，管理财务和结算",
                "biz_type": BusinessType.COMMON.value,
                "is_admin": False,
                "status": 1,
                "sort_order": 30,
                "create_by": "system",
                "update_by": "system"
            }
        ]
        
        created_roles = {}
        for role_data in roles_data:
            role = SysRole(**role_data)
            session.add(role)
            await session.flush()  # 获取ID
            created_roles[role_data["role_code"]] = role.role_id
            logger.info(f"✅ 创建角色: {role.role_name} ({role.role_id})")
        
        await session.commit()
        logger.info(f"✅ 共创建 {len(roles_data)} 个角色")
        return created_roles
    
    async def init_menus(self, session: AsyncSession):
        """初始化菜单结构"""
        logger.info("📋 初始化菜单结构...")
        
        menus_data = [
            # ==================== 平台管理菜单 ====================
            {
                "parent_id": None,
                "menu_name": "平台管理",
                "menu_code": "platform",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform",
                "component": "Layout",
                "icon": "platform",
                "order_num": 1,
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "platform",  # 临时使用编码，稍后替换为ID
                "menu_name": "系统管理",
                "menu_code": "platform:system",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/system",
                "component": "platform/system/index",
                "icon": "system",
                "order_num": 1,
                "permission": "platform:system:view",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "platform:system",
                "menu_name": "用户管理",
                "menu_code": "platform:system:user",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/system/users",
                "component": "platform/system/users/index",
                "icon": "user",
                "order_num": 1,
                "permission": "platform:system:user:list",
                "http_method": "GET",
                "api_url": "/api/v1/rbac/users",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "platform:system",
                "menu_name": "角色管理",
                "menu_code": "platform:system:role",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/system/roles",
                "component": "platform/system/roles/index",
                "icon": "role",
                "order_num": 2,
                "permission": "platform:system:role:list",
                "http_method": "GET",
                "api_url": "/api/v1/rbac/roles",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "platform:system",
                "menu_name": "菜单管理",
                "menu_code": "platform:system:menu",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/system/menus",
                "component": "platform/system/menus/index",
                "icon": "menu",
                "order_num": 3,
                "permission": "platform:system:menu:list",
                "http_method": "GET",
                "api_url": "/api/v1/rbac/menus",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "platform",
                "menu_name": "商户管理",
                "menu_code": "platform:shop",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/shops",
                "component": "platform/shops/index",
                "icon": "shop",
                "order_num": 2,
                "permission": "platform:shop:list",
                "http_method": "GET",
                "api_url": "/api/v1/shops",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "platform",
                "menu_name": "分类管理",
                "menu_code": "platform:category",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/categories",
                "component": "platform/categories/index",
                "icon": "category",
                "order_num": 3,
                "permission": "platform:category:list",
                "http_method": "GET",
                "api_url": "/api/v1/categories",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "platform",
                "menu_name": "数据统计",
                "menu_code": "platform:dashboard",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/dashboard",
                "component": "platform/dashboard/index",
                "icon": "dashboard",
                "order_num": 4,
                "permission": "platform:dashboard:view",
                "status": 1,
                "create_by": "system"
            },
            
            # ==================== 商城管理菜单 ====================
            {
                "parent_id": None,
                "menu_name": "商城管理",
                "menu_code": "shop",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop",
                "component": "Layout",
                "icon": "shop-manage",
                "order_num": 1,
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "shop",
                "menu_name": "商品管理",
                "menu_code": "shop:product",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/products",
                "component": "shop/products/index",
                "icon": "product",
                "order_num": 1,
                "permission": "shop:product:list",
                "http_method": "GET",
                "api_url": "/api/spu",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "shop:product",
                "menu_name": "SPU管理",
                "menu_code": "shop:product:spu",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/products/spu",
                "component": "shop/products/spu/index",
                "icon": "spu",
                "order_num": 1,
                "permission": "shop:product:spu:list",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "shop:product",
                "menu_name": "SKU管理",
                "menu_code": "shop:product:sku",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/products/sku",
                "component": "shop/products/sku/index",
                "icon": "sku",
                "order_num": 2,
                "permission": "shop:product:sku:list",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "shop",
                "menu_name": "订单管理",
                "menu_code": "shop:order",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/orders",
                "component": "shop/orders/index",
                "icon": "order",
                "order_num": 2,
                "permission": "shop:order:list",
                "http_method": "GET",
                "api_url": "/api/orders",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "shop",
                "menu_name": "库存管理",
                "menu_code": "shop:inventory",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/inventory",
                "component": "shop/inventory/index",
                "icon": "inventory",
                "order_num": 3,
                "permission": "shop:inventory:list",
                "status": 1,
                "create_by": "system"
            },
            {
                "parent_id": "shop",
                "menu_name": "营销工具",
                "menu_code": "shop:marketing",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/marketing",
                "component": "shop/marketing/index",
                "icon": "marketing",
                "order_num": 4,
                "permission": "shop:marketing:view",
                "status": 1,
                "create_by": "system"
            },
            
            # ==================== 通用功能菜单 ====================
            {
                "parent_id": None,
                "menu_name": "个人中心",
                "menu_code": "profile",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.COMMON.value,
                "path": "/profile",
                "component": "profile/index",
                "icon": "user-profile",
                "order_num": 100,
                "permission": "profile:view",
                "status": 1,
                "create_by": "system"
            }
        ]
        
        # 第一轮：创建所有菜单
        created_menus = {}
        for menu_data in menus_data:
            # 暂时移除parent_id，稍后处理层级关系
            parent_code = menu_data.pop("parent_id", None)
            menu = SysMenu(**menu_data)
            session.add(menu)
            await session.flush()
            created_menus[menu_data["menu_code"]] = {
                "menu_id": menu.menu_id,
                "parent_code": parent_code
            }
            logger.info(f"✅ 创建菜单: {menu.menu_name} ({menu.menu_id})")
        
        await session.commit()
        
        # 第二轮：设置父子关系
        logger.info("🔗 设置菜单父子关系...")
        for menu_code, info in created_menus.items():
            if info["parent_code"]:
                parent_info = created_menus.get(info["parent_code"])
                if parent_info:
                    menu = await session.get(SysMenu, info["menu_id"])
                    if menu:
                        menu.parent_id = parent_info["menu_id"]
        
        await session.commit()
        logger.info(f"✅ 共创建 {len(menus_data)} 个菜单")
        return created_menus
    
    async def assign_role_permissions(self, session: AsyncSession, 
                                    created_roles: dict, created_menus: dict):
        """分配角色权限"""
        logger.info("🔐 分配角色权限...")
        
        # 平台超级管理员：所有平台权限
        platform_menu_ids = [
            info["menu_id"] for code, info in created_menus.items()
            if "platform:" in code or code == "platform"
        ]
        
        await self._assign_menus_to_role(
            session,
            created_roles["PLATFORM_SUPER_ADMIN"],
            platform_menu_ids,
            "平台超级管理员"
        )
        
        # 平台管理员：平台管理权限（不包含系统管理）
        platform_admin_menu_ids = [
            info["menu_id"] for code, info in created_menus.items()
            if code in ["platform", "platform:shop", "platform:category", "platform:dashboard"]
        ]
        
        await self._assign_menus_to_role(
            session,
            created_roles["PLATFORM_ADMIN"],
            platform_admin_menu_ids,
            "平台管理员"
        )
        
        # 商城管理员：所有商城权限
        shop_menu_ids = [
            info["menu_id"] for code, info in created_menus.items()
            if "shop:" in code or code == "shop"
        ]
        
        await self._assign_menus_to_role(
            session,
            created_roles["SHOP_ADMIN"],
            shop_menu_ids,
            "商城管理员"
        )
        
        # 商城运营：商品和订单管理
        shop_operator_menu_ids = [
            info["menu_id"] for code, info in created_menus.items()
            if code in ["shop", "shop:product", "shop:product:spu", 
                       "shop:product:sku", "shop:order"]
        ]
        
        await self._assign_menus_to_role(
            session,
            created_roles["SHOP_OPERATOR"], 
            shop_operator_menu_ids,
            "商城运营"
        )
        
        # 通用权限：个人中心
        common_menu_ids = [
            info["menu_id"] for code, info in created_menus.items()
            if code == "profile"
        ]
        
        for role_code in ["CUSTOMER_SERVICE", "FINANCE_STAFF"]:
            await self._assign_menus_to_role(
                session,
                created_roles[role_code],
                common_menu_ids,
                role_code
            )
        
        await session.commit()
        logger.info("✅ 角色权限分配完成")
    
    async def _assign_menus_to_role(self, session: AsyncSession, role_id: int, 
                                   menu_ids: list, role_name: str):
        """为角色分配菜单权限"""
        for menu_id in menu_ids:
            role_menu = SysRoleMenu(
                role_id=role_id,
                menu_id=menu_id,
                permission_type=PermissionType.READ.value,
                create_by="system"
            )
            session.add(role_menu)
        
        logger.info(f"✅ 为 {role_name} 分配 {len(menu_ids)} 个菜单权限")
    
    async def create_admin_user(self, session: AsyncSession, created_roles: dict):
        """创建默认管理员用户"""
        logger.info("👤 创建默认管理员用户...")
        
        import secrets
        import hashlib
        
        # 创建平台超级管理员
        admin_user_id = self.generate_id("ADM")
        salt = secrets.token_hex(16)
        password = "admin123"  # 默认密码，首次登录需要修改
        password_hash = hashlib.pbkdf2_hmac('sha256', 
                                          password.encode('utf-8'),
                                          salt.encode('utf-8'),
                                          100000).hex()
        
        admin_user = SysUser(
            user_id=admin_user_id,
            username="admin",
            email="admin@openpenpal.com",
            real_name="系统管理员",
            nickname="超级管理员",
            status=1,
            user_type=1,
            password_hash=password_hash,
            salt=salt,
            password_update_time=datetime.utcnow(),
            create_by="system",
            update_by="system"
        )
        
        session.add(admin_user)
        await session.flush()
        
        # 分配超级管理员角色
        user_role = SysUserRole(
            user_id=admin_user_id,
            role_id=created_roles["PLATFORM_SUPER_ADMIN"],
            create_by="system"
        )
        session.add(user_role)
        
        await session.commit()
        
        logger.info(f"✅ 创建管理员用户: admin (密码: {password})")
        logger.warning("⚠️ 请首次登录后立即修改默认密码！")
        
        return admin_user_id
    
    async def run_initialization(self):
        """运行完整初始化流程"""
        logger.info("🚀 开始RBAC系统初始化...")
        logger.info("=" * 60)
        
        start_time = datetime.now()
        
        try:
            # 1. 创建数据表
            await self.create_tables()
            
            async with self.SessionLocal() as session:
                # 2. 初始化角色
                created_roles = await self.init_roles(session)
                
                # 3. 初始化菜单
                created_menus = await self.init_menus(session)
                
                # 4. 分配角色权限
                await self.assign_role_permissions(session, created_roles, created_menus)
                
                # 5. 创建管理员用户
                admin_user_id = await self.create_admin_user(session, created_roles)
            
            end_time = datetime.now()
            duration = end_time - start_time
            
            logger.info("=" * 60)
            logger.info(f"🎉 RBAC系统初始化完成！耗时: {duration}")
            logger.info("📊 初始化统计:")
            logger.info(f"   角色数量: {len(created_roles)}")
            logger.info(f"   菜单数量: {len(created_menus)}")
            logger.info(f"   管理员用户ID: {admin_user_id}")
            logger.info("🔑 默认登录信息:")
            logger.info("   用户名: admin")
            logger.info("   密码: admin123")
            logger.info("⚠️ 请首次登录后立即修改密码！")
            
            return True
            
        except Exception as e:
            logger.error(f"❌ RBAC系统初始化失败: {e}")
            raise
        
        finally:
            await self.engine.dispose()


async def main():
    """主函数"""
    print("RBAC权限管理系统初始化工具")
    print("=" * 60)
    print("⚠️ 注意：此操作会创建RBAC相关数据表并初始化数据")
    print("📋 初始化内容：")
    print("   1. 创建RBAC数据表")
    print("   2. 初始化6个基础角色")
    print("   3. 创建完整菜单结构")
    print("   4. 分配角色权限")
    print("   5. 创建默认管理员账户")
    print()
    
    confirm = input("确认继续执行初始化？(y/N): ")
    if confirm.lower() != 'y':
        print("❌ 初始化已取消")
        return
    
    # 使用配置中的数据库URL
    database_url = settings.DATABASE_URL
    if database_url.startswith("postgresql://"):
        database_url = database_url.replace("postgresql://", "postgresql+asyncpg://")
    
    initializer = RBACInitializer(database_url)
    await initializer.run_initialization()


if __name__ == "__main__":
    asyncio.run(main())