"""
RBAC权限管理服务

基于mall4cloud设计的完整RBAC权限管理系统：
1. 菜单权限管理
2. 角色权限分配
3. 用户角色映射
4. 权限验证和缓存
5. 操作日志记录
"""

from typing import List, Dict, Any, Optional, Set, Tuple
from datetime import datetime, timedelta
from sqlalchemy import select, and_, or_, func, text
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload, joinedload
import json
import hashlib
import secrets
from collections import defaultdict

from app.models.rbac import (
    SysMenu, SysRole, SysUser, SysRoleMenu, SysUserRole, 
    SysOperLog, SysUserOnline, BusinessType, MenuType, PermissionType
)
from app.core.exceptions import BusinessException
from app.core.logger import logger
from app.utils.cache_manager import CacheManager


class RBACService:
    """RBAC权限管理服务"""
    
    def __init__(self, session: AsyncSession):
        self.session = session
        self.cache = CacheManager()
        # 安全的缓存过期时间配置
        self.user_permission_expire = 300   # 用户权限缓存5分钟
        self.role_permission_expire = 600   # 角色权限缓存10分钟
        self.menu_cache_expire = 1800       # 菜单缓存30分钟
        self.cache_expire = self.user_permission_expire  # 向后兼容
    
    # ==================== 菜单管理 ====================
    
    async def create_menu(self, menu_data: Dict) -> SysMenu:
        """创建菜单"""
        # 验证父菜单
        if menu_data.get("parent_id"):
            parent_menu = await self.session.get(SysMenu, menu_data["parent_id"])
            if not parent_menu:
                raise BusinessException("父菜单不存在")
        
        # 验证菜单编码唯一性
        result = await self.session.execute(
            select(func.count(SysMenu.menu_id))
            .where(SysMenu.menu_code == menu_data["menu_code"])
        )
        if result.scalar() > 0:
            raise BusinessException("菜单编码已存在")
        
        menu = SysMenu(**menu_data)
        self.session.add(menu)
        await self.session.commit()
        
        # 清除菜单缓存
        await self._clear_menu_cache()
        
        logger.info(f"创建菜单成功: {menu.menu_name} ({menu.menu_id})")
        return menu
    
    async def update_menu(self, menu_id: int, menu_data: Dict) -> SysMenu:
        """更新菜单"""
        menu = await self.session.get(SysMenu, menu_id)
        if not menu:
            raise BusinessException("菜单不存在")
        
        # 更新字段
        for field, value in menu_data.items():
            if hasattr(menu, field) and field not in ["menu_id", "create_time", "create_by"]:
                setattr(menu, field, value)
        
        await self.session.commit()
        await self._clear_menu_cache()
        
        return menu
    
    async def delete_menu(self, menu_id: int, force: bool = False):
        """删除菜单"""
        menu = await self.session.get(SysMenu, menu_id)
        if not menu:
            raise BusinessException("菜单不存在")
        
        # 检查是否有子菜单
        result = await self.session.execute(
            select(func.count(SysMenu.menu_id))
            .where(SysMenu.parent_id == menu_id)
        )
        children_count = result.scalar()
        
        if children_count > 0 and not force:
            raise BusinessException("该菜单下有子菜单，请先删除子菜单")
        
        # 检查是否有角色关联
        result = await self.session.execute(
            select(func.count(SysRoleMenu.role_id))
            .where(SysRoleMenu.menu_id == menu_id)
        )
        role_count = result.scalar()
        
        if role_count > 0 and not force:
            raise BusinessException("该菜单已被角色使用，请先解除关联")
        
        if force:
            # 强制删除：删除所有子菜单和角色关联
            await self.session.execute(
                text("DELETE FROM sys_menu WHERE parent_id = :menu_id"),
                {"menu_id": menu_id}
            )
            await self.session.execute(
                text("DELETE FROM sys_role_menu WHERE menu_id = :menu_id"),
                {"menu_id": menu_id}
            )
        
        await self.session.delete(menu)
        await self.session.commit()
        
        await self._clear_menu_cache()
        logger.info(f"删除菜单: {menu.menu_name} ({menu_id})")
    
    async def get_menu_tree(self, biz_type: Optional[int] = None, 
                           include_buttons: bool = True) -> List[Dict]:
        """获取菜单树"""
        cache_key = f"menu_tree:{biz_type}:{include_buttons}"
        cached = await self.cache.get(cache_key)
        if cached:
            return json.loads(cached)
        
        # 构建查询
        query = select(SysMenu).where(SysMenu.status == 1)
        
        if biz_type is not None:
            query = query.where(or_(
                SysMenu.biz_type == biz_type,
                SysMenu.biz_type == BusinessType.COMMON.value
            ))
        
        if not include_buttons:
            query = query.where(SysMenu.menu_type != MenuType.BUTTON.value)
        
        query = query.order_by(SysMenu.order_num, SysMenu.menu_id)
        
        result = await self.session.execute(query)
        menus = result.scalars().all()
        
        # 构建树结构
        menu_tree = self._build_menu_tree(menus)
        
        # 缓存结果 - 使用菜单专用过期时间
        await self.cache.set(cache_key, json.dumps(menu_tree, default=str), expire=self.menu_cache_expire)
        
        return menu_tree
    
    def _build_menu_tree(self, menus: List[SysMenu]) -> List[Dict]:
        """构建菜单树"""
        menu_dict = {menu.menu_id: menu.to_dict() for menu in menus}
        tree = []
        
        for menu in menus:
            menu_data = menu_dict[menu.menu_id]
            if menu.parent_id is None:
                tree.append(menu_data)
            else:
                parent = menu_dict.get(menu.parent_id)
                if parent:
                    if "children" not in parent:
                        parent["children"] = []
                    parent["children"].append(menu_data)
        
        return tree
    
    async def get_user_menus(self, user_id: str, biz_type: int) -> List[Dict]:
        """获取用户菜单"""
        cache_key = f"user_menus:{user_id}:{biz_type}"
        cached = await self.cache.get(cache_key)
        if cached:
            return json.loads(cached)
        
        # 查询用户角色的菜单权限
        query = select(SysMenu).distinct().join(
            SysRoleMenu, SysMenu.menu_id == SysRoleMenu.menu_id
        ).join(
            SysRole, SysRoleMenu.role_id == SysRole.role_id
        ).join(
            SysUserRole, SysRole.role_id == SysUserRole.role_id
        ).where(
            and_(
                SysUserRole.user_id == user_id,
                SysRole.status == 1,
                SysMenu.status == 1,
                or_(
                    SysMenu.biz_type == biz_type,
                    SysMenu.biz_type == BusinessType.COMMON.value
                )
            )
        ).order_by(SysMenu.order_num, SysMenu.menu_id)
        
        result = await self.session.execute(query)
        menus = result.scalars().all()
        
        menu_tree = self._build_menu_tree(menus)
        
        # 缓存结果 - 使用菜单专用过期时间
        await self.cache.set(cache_key, json.dumps(menu_tree, default=str), expire=self.menu_cache_expire)
        
        return menu_tree
    
    # ==================== 角色管理 ====================
    
    async def create_role(self, role_data: Dict) -> SysRole:
        """创建角色"""
        # 验证角色编码唯一性
        result = await self.session.execute(
            select(func.count(SysRole.role_id))
            .where(SysRole.role_code == role_data["role_code"])
        )
        if result.scalar() > 0:
            raise BusinessException("角色编码已存在")
        
        role = SysRole(**role_data)
        self.session.add(role)
        await self.session.commit()
        
        logger.info(f"创建角色成功: {role.role_name} ({role.role_id})")
        return role
    
    async def assign_role_menus(self, role_id: int, menu_ids: List[int]):
        """分配角色菜单权限"""
        role = await self.session.get(SysRole, role_id)
        if not role:
            raise BusinessException("角色不存在")
        
        # 删除原有权限
        await self.session.execute(
            text("DELETE FROM sys_role_menu WHERE role_id = :role_id"),
            {"role_id": role_id}
        )
        
        # 添加新权限
        for menu_id in menu_ids:
            role_menu = SysRoleMenu(
                role_id=role_id,
                menu_id=menu_id,
                permission_type=PermissionType.READ.value
            )
            self.session.add(role_menu)
        
        await self.session.commit()
        
        # 清除相关缓存
        await self._clear_role_cache(role_id)
        
        logger.info(f"角色 {role_id} 分配菜单权限成功，共 {len(menu_ids)} 个菜单")
    
    async def get_role_menus(self, role_id: int) -> List[int]:
        """获取角色菜单权限"""
        result = await self.session.execute(
            select(SysRoleMenu.menu_id)
            .where(SysRoleMenu.role_id == role_id)
        )
        return [row[0] for row in result.fetchall()]
    
    # ==================== 用户管理 ====================
    
    async def create_user(self, user_data: Dict) -> SysUser:
        """创建用户"""
        # 生成密码哈希
        if "password" in user_data:
            salt = secrets.token_hex(16)
            password_hash = self._hash_password(user_data["password"], salt)
            user_data.update({
                "password_hash": password_hash,
                "salt": salt,
                "password_update_time": datetime.utcnow()
            })
            del user_data["password"]
        
        user = SysUser(**user_data)
        self.session.add(user)
        await self.session.commit()
        
        logger.info(f"创建用户成功: {user.username} ({user.user_id})")
        return user
    
    async def assign_user_roles(self, user_id: str, role_ids: List[int]):
        """分配用户角色"""
        user = await self.session.get(SysUser, user_id)
        if not user:
            raise BusinessException("用户不存在")
        
        # 删除原有角色
        await self.session.execute(
            text("DELETE FROM sys_user_role WHERE user_id = :user_id"),
            {"user_id": user_id}
        )
        
        # 添加新角色
        for role_id in role_ids:
            user_role = SysUserRole(user_id=user_id, role_id=role_id)
            self.session.add(user_role)
        
        await self.session.commit()
        
        # 清除用户权限缓存
        await self._clear_user_cache(user_id)
        
        logger.info(f"用户 {user_id} 分配角色成功，共 {len(role_ids)} 个角色")
    
    async def get_user_roles(self, user_id: str) -> List[SysRole]:
        """获取用户角色"""
        result = await self.session.execute(
            select(SysRole)
            .join(SysUserRole, SysRole.role_id == SysUserRole.role_id)
            .where(
                and_(
                    SysUserRole.user_id == user_id,
                    SysRole.status == 1
                )
            )
            .order_by(SysRole.sort_order)
        )
        return result.scalars().all()
    
    def _hash_password(self, password: str, salt: str) -> str:
        """密码哈希"""
        return hashlib.pbkdf2_hmac('sha256', 
                                 password.encode('utf-8'), 
                                 salt.encode('utf-8'), 
                                 100000).hex()
    
    async def verify_password(self, user_id: str, password: str) -> bool:
        """验证密码"""
        user = await self.session.get(SysUser, user_id)
        if not user or not user.password_hash or not user.salt:
            return False
        
        password_hash = self._hash_password(password, user.salt)
        return password_hash == user.password_hash
    
    # ==================== 权限验证 ====================
    
    async def check_permission(self, user_id: str, permission: str, 
                              http_method: str = "GET") -> bool:
        """检查用户权限"""
        cache_key = f"user_permissions:{user_id}"
        cached = await self.cache.get(cache_key)
        
        if cached:
            permissions = json.loads(cached)
        else:
            permissions = await self._load_user_permissions(user_id)
            await self.cache.set(cache_key, json.dumps(permissions), expire=self.user_permission_expire)
        
        # 检查权限
        for perm in permissions:
            if (perm["permission"] == permission and 
                (perm["http_method"] is None or perm["http_method"] == http_method)):
                return True
        
        return False
    
    async def _load_user_permissions(self, user_id: str) -> List[Dict]:
        """加载用户权限"""
        result = await self.session.execute(
            select(SysMenu.permission, SysMenu.http_method)
            .distinct()
            .join(SysRoleMenu, SysMenu.menu_id == SysRoleMenu.menu_id)
            .join(SysRole, SysRoleMenu.role_id == SysRole.role_id)
            .join(SysUserRole, SysRole.role_id == SysUserRole.role_id)
            .where(
                and_(
                    SysUserRole.user_id == user_id,
                    SysRole.status == 1,
                    SysMenu.status == 1,
                    SysMenu.permission.is_not(None)
                )
            )
        )
        
        permissions = []
        for row in result.fetchall():
            permissions.append({
                "permission": row.permission,
                "http_method": row.http_method
            })
        
        return permissions
    
    async def has_role(self, user_id: str, role_code: str) -> bool:
        """检查用户是否有指定角色"""
        result = await self.session.execute(
            select(func.count(SysUserRole.user_id))
            .join(SysRole, SysUserRole.role_id == SysRole.role_id)
            .where(
                and_(
                    SysUserRole.user_id == user_id,
                    SysRole.role_code == role_code,
                    SysRole.status == 1
                )
            )
        )
        return result.scalar() > 0
    
    async def is_admin(self, user_id: str) -> bool:
        """检查用户是否为管理员"""
        result = await self.session.execute(
            select(func.count(SysUserRole.user_id))
            .join(SysRole, SysUserRole.role_id == SysRole.role_id)
            .where(
                and_(
                    SysUserRole.user_id == user_id,
                    SysRole.is_admin == True,
                    SysRole.status == 1
                )
            )
        )
        return result.scalar() > 0
    
    # ==================== 操作日志 ====================
    
    async def log_operation(self, log_data: Dict):
        """记录操作日志"""
        log = SysOperLog(**log_data)
        self.session.add(log)
        await self.session.commit()
    
    async def get_operation_logs(self, user_id: Optional[str] = None,
                                start_time: Optional[datetime] = None,
                                end_time: Optional[datetime] = None,
                                page: int = 1, size: int = 20) -> Tuple[List[SysOperLog], int]:
        """获取操作日志"""
        query = select(SysOperLog)
        count_query = select(func.count(SysOperLog.oper_id))
        
        # 构建条件
        conditions = []
        if user_id:
            conditions.append(SysOperLog.oper_user_id == user_id)
        if start_time:
            conditions.append(SysOperLog.oper_time >= start_time)
        if end_time:
            conditions.append(SysOperLog.oper_time <= end_time)
        
        if conditions:
            query = query.where(and_(*conditions))
            count_query = count_query.where(and_(*conditions))
        
        # 获取总数
        total_result = await self.session.execute(count_query)
        total = total_result.scalar()
        
        # 分页查询
        query = query.order_by(SysOperLog.oper_time.desc())
        query = query.offset((page - 1) * size).limit(size)
        
        result = await self.session.execute(query)
        logs = result.scalars().all()
        
        return logs, total
    
    # ==================== 在线用户 ====================
    
    async def user_login(self, user_id: str, session_id: str, login_info: Dict):
        """用户登录"""
        # 更新用户登录信息
        user = await self.session.get(SysUser, user_id)
        if user:
            user.last_login_time = datetime.utcnow()
            user.last_login_ip = login_info.get("ipaddr")
            user.login_fail_count = 0  # 重置失败次数
        
        # 记录在线用户
        online_user = SysUserOnline(
            session_id=session_id,
            user_id=user_id,
            login_name=user.username if user else "",
            real_name=user.real_name if user else "",
            ipaddr=login_info.get("ipaddr"),
            login_location=login_info.get("login_location"),
            browser=login_info.get("browser"),
            os=login_info.get("os"),
            status="on_line",
            start_timestamp=datetime.utcnow(),
            last_access_time=datetime.utcnow(),
            expire_time=login_info.get("expire_time", 30)
        )
        
        self.session.add(online_user)
        await self.session.commit()
        
        # 记录登录日志
        await self.log_operation({
            "title": "用户登录",
            "business_type": 1,
            "method": "user_login",
            "request_method": "POST",
            "oper_name": user.username if user else "",
            "oper_user_id": user_id,
            "oper_url": "/login",
            "oper_ip": login_info.get("ipaddr"),
            "oper_location": login_info.get("login_location"),
            "status": 0
        })
    
    async def user_logout(self, session_id: str):
        """用户登出"""
        # 删除在线用户记录
        await self.session.execute(
            text("DELETE FROM sys_user_online WHERE session_id = :session_id"),
            {"session_id": session_id}
        )
        await self.session.commit()
    
    async def get_online_users(self) -> List[SysUserOnline]:
        """获取在线用户"""
        result = await self.session.execute(
            select(SysUserOnline)
            .where(SysUserOnline.status == "on_line")
            .order_by(SysUserOnline.last_access_time.desc())
        )
        return result.scalars().all()
    
    # ==================== 缓存管理 ====================
    
    async def _clear_menu_cache(self):
        """清除菜单缓存"""
        pattern = "menu_tree:*"
        await self.cache.delete_pattern(pattern)
    
    async def _clear_role_cache(self, role_id: int):
        """清除角色缓存"""
        # 清除该角色相关的用户菜单缓存
        result = await self.session.execute(
            select(SysUserRole.user_id)
            .where(SysUserRole.role_id == role_id)
        )
        user_ids = [row[0] for row in result.fetchall()]
        
        for user_id in user_ids:
            await self._clear_user_cache(user_id)
    
    async def _clear_user_cache(self, user_id: str):
        """清除用户缓存"""
        patterns = [
            f"user_menus:{user_id}:*",
            f"user_permissions:{user_id}"
        ]
        for pattern in patterns:
            await self.cache.delete_pattern(pattern)
    
    # ==================== 数据统计 ====================
    
    async def get_rbac_statistics(self) -> Dict:
        """获取RBAC统计信息"""
        # 用户统计
        user_total = await self.session.execute(select(func.count(SysUser.user_id)))
        user_active = await self.session.execute(
            select(func.count(SysUser.user_id)).where(SysUser.status == 1)
        )
        
        # 角色统计
        role_total = await self.session.execute(select(func.count(SysRole.role_id)))
        role_active = await self.session.execute(
            select(func.count(SysRole.role_id)).where(SysRole.status == 1)
        )
        
        # 菜单统计
        menu_total = await self.session.execute(select(func.count(SysMenu.menu_id)))
        menu_active = await self.session.execute(
            select(func.count(SysMenu.menu_id)).where(SysMenu.status == 1)
        )
        
        # 在线用户统计
        online_users = await self.session.execute(
            select(func.count(SysUserOnline.session_id))
            .where(SysUserOnline.status == "on_line")
        )
        
        return {
            "user_total": user_total.scalar(),
            "user_active": user_active.scalar(),
            "role_total": role_total.scalar(),
            "role_active": role_active.scalar(),
            "menu_total": menu_total.scalar(),
            "menu_active": menu_active.scalar(),
            "online_users": online_users.scalar()
        }