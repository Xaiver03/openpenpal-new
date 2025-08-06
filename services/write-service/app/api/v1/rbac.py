"""
RBAC权限管理API

提供完整的角色权限管理功能：
1. 菜单管理
2. 角色管理
3. 用户管理
4. 权限验证
5. 操作日志
"""

from typing import List, Dict, Any, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from pydantic import BaseModel, Field
from datetime import datetime

from app.core.database import get_async_session
from app.core.auth import get_current_user
from app.core.responses import success_response, error_response
from app.core.exceptions import BusinessException
from app.services.rbac_service import RBACService
from app.models.rbac import BusinessType, MenuType, PermissionType
from app.models.user import User

router = APIRouter(prefix="/rbac", tags=["RBAC权限管理"])


# ==================== 请求模型 ====================

class MenuCreateRequest(BaseModel):
    """创建菜单请求"""
    parent_id: Optional[int] = Field(None, description="父菜单ID")
    menu_name: str = Field(..., description="菜单名称")
    menu_code: str = Field(..., description="菜单编码")
    menu_type: int = Field(MenuType.MENU.value, description="菜单类型")
    biz_type: int = Field(BusinessType.COMMON.value, description="业务类型")
    path: Optional[str] = Field(None, description="路由地址")
    component: Optional[str] = Field(None, description="组件路径")
    redirect: Optional[str] = Field(None, description="重定向地址")
    icon: Optional[str] = Field(None, description="菜单图标")
    order_num: int = Field(0, description="显示顺序")
    is_hidden: bool = Field(False, description="是否隐藏")
    is_cache: bool = Field(False, description="是否缓存")
    is_affix: bool = Field(False, description="是否固定标签")
    permission: Optional[str] = Field(None, description="权限标识")
    http_method: Optional[str] = Field(None, description="HTTP方法")
    api_url: Optional[str] = Field(None, description="API地址")
    status: int = Field(1, description="菜单状态")
    remark: Optional[str] = Field(None, description="备注")
    meta_info: Optional[Dict] = Field(None, description="元信息")


class MenuUpdateRequest(BaseModel):
    """更新菜单请求"""
    parent_id: Optional[int] = None
    menu_name: Optional[str] = None
    menu_code: Optional[str] = None
    menu_type: Optional[int] = None
    biz_type: Optional[int] = None
    path: Optional[str] = None
    component: Optional[str] = None
    redirect: Optional[str] = None
    icon: Optional[str] = None
    order_num: Optional[int] = None
    is_hidden: Optional[bool] = None
    is_cache: Optional[bool] = None
    is_affix: Optional[bool] = None
    permission: Optional[str] = None
    http_method: Optional[str] = None
    api_url: Optional[str] = None
    status: Optional[int] = None
    remark: Optional[str] = None
    meta_info: Optional[Dict] = None


class RoleCreateRequest(BaseModel):
    """创建角色请求"""
    role_name: str = Field(..., description="角色名称")
    role_code: str = Field(..., description="角色编码")
    role_desc: Optional[str] = Field(None, description="角色描述")
    biz_type: int = Field(BusinessType.COMMON.value, description="业务类型")
    is_admin: bool = Field(False, description="是否管理员角色")
    data_scope: int = Field(1, description="数据权限范围")
    status: int = Field(1, description="角色状态")
    sort_order: int = Field(0, description="显示顺序")


class RoleUpdateRequest(BaseModel):
    """更新角色请求"""
    role_name: Optional[str] = None
    role_desc: Optional[str] = None
    biz_type: Optional[int] = None
    is_admin: Optional[bool] = None
    data_scope: Optional[int] = None
    status: Optional[int] = None
    sort_order: Optional[int] = None


class RoleMenuAssignRequest(BaseModel):
    """角色菜单分配请求"""
    menu_ids: List[int] = Field(..., description="菜单ID列表")


class UserCreateRequest(BaseModel):
    """创建用户请求"""
    user_id: str = Field(..., description="用户ID")
    username: str = Field(..., description="用户名")
    password: str = Field(..., description="密码")
    email: str = Field(..., description="邮箱")
    phone: Optional[str] = Field(None, description="手机号")
    real_name: Optional[str] = Field(None, description="真实姓名")
    nickname: Optional[str] = Field(None, description="昵称")
    status: int = Field(1, description="用户状态")
    user_type: int = Field(1, description="用户类型")


class UserRoleAssignRequest(BaseModel):
    """用户角色分配请求"""
    role_ids: List[int] = Field(..., description="角色ID列表")


# ==================== 菜单管理API ====================

@router.post("/menus")
async def create_menu(
    request: MenuCreateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """创建菜单"""
    try:
        service = RBACService(session)
        
        menu_data = request.dict()
        menu_data.update({
            "create_by": current_user.user_id,
            "update_by": current_user.user_id
        })
        
        menu = await service.create_menu(menu_data)
        
        return success_response(
            data=menu.to_dict(),
            message="菜单创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建菜单失败: {str(e)}")


@router.put("/menus/{menu_id}")
async def update_menu(
    menu_id: int,
    request: MenuUpdateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """更新菜单"""
    try:
        service = RBACService(session)
        
        # 过滤非空字段
        menu_data = {k: v for k, v in request.dict().items() if v is not None}
        menu_data["update_by"] = current_user.user_id
        
        menu = await service.update_menu(menu_id, menu_data)
        
        return success_response(
            data=menu.to_dict(),
            message="菜单更新成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"更新菜单失败: {str(e)}")


@router.delete("/menus/{menu_id}")
async def delete_menu(
    menu_id: int,
    force: bool = Query(False, description="是否强制删除"),
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """删除菜单"""
    try:
        service = RBACService(session)
        await service.delete_menu(menu_id, force=force)
        
        return success_response(message="菜单删除成功")
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"删除菜单失败: {str(e)}")


@router.get("/menus/tree")
async def get_menu_tree(
    biz_type: Optional[int] = Query(None, description="业务类型"),
    include_buttons: bool = Query(True, description="是否包含按钮"),
    session: AsyncSession = Depends(get_async_session)
):
    """获取菜单树"""
    try:
        service = RBACService(session)
        menu_tree = await service.get_menu_tree(biz_type, include_buttons)
        
        return success_response(data={
            "tree": menu_tree,
            "biz_type": biz_type,
            "include_buttons": include_buttons
        })
        
    except Exception as e:
        return error_response(message=f"获取菜单树失败: {str(e)}")


@router.get("/users/{user_id}/menus")
async def get_user_menus(
    user_id: str,
    biz_type: int = Query(..., description="业务类型"),
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取用户菜单"""
    try:
        service = RBACService(session)
        menus = await service.get_user_menus(user_id, biz_type)
        
        return success_response(data={
            "user_id": user_id,
            "biz_type": biz_type,
            "menus": menus
        })
        
    except Exception as e:
        return error_response(message=f"获取用户菜单失败: {str(e)}")


# ==================== 角色管理API ====================

@router.post("/roles")
async def create_role(
    request: RoleCreateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """创建角色"""
    try:
        service = RBACService(session)
        
        role_data = request.dict()
        role_data.update({
            "create_by": current_user.user_id,
            "update_by": current_user.user_id
        })
        
        role = await service.create_role(role_data)
        
        return success_response(
            data=role.to_dict(),
            message="角色创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建角色失败: {str(e)}")


@router.post("/roles/{role_id}/menus")
async def assign_role_menus(
    role_id: int,
    request: RoleMenuAssignRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """分配角色菜单权限"""
    try:
        service = RBACService(session)
        await service.assign_role_menus(role_id, request.menu_ids)
        
        return success_response(
            data={
                "role_id": role_id,
                "menu_count": len(request.menu_ids)
            },
            message="角色权限分配成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"分配角色权限失败: {str(e)}")


@router.get("/roles/{role_id}/menus")
async def get_role_menus(
    role_id: int,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取角色菜单权限"""
    try:
        service = RBACService(session)
        menu_ids = await service.get_role_menus(role_id)
        
        return success_response(data={
            "role_id": role_id,
            "menu_ids": menu_ids
        })
        
    except Exception as e:
        return error_response(message=f"获取角色权限失败: {str(e)}")


# ==================== 用户管理API ====================

@router.post("/users")
async def create_user(
    request: UserCreateRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """创建用户"""
    try:
        service = RBACService(session)
        
        user_data = request.dict()
        user_data.update({
            "create_by": current_user.user_id,
            "update_by": current_user.user_id
        })
        
        user = await service.create_user(user_data)
        
        return success_response(
            data=user.to_dict(),
            message="用户创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"创建用户失败: {str(e)}")


@router.post("/users/{user_id}/roles")
async def assign_user_roles(
    user_id: str,
    request: UserRoleAssignRequest,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """分配用户角色"""
    try:
        service = RBACService(session)
        await service.assign_user_roles(user_id, request.role_ids)
        
        return success_response(
            data={
                "user_id": user_id,
                "role_count": len(request.role_ids)
            },
            message="用户角色分配成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"分配用户角色失败: {str(e)}")


@router.get("/users/{user_id}/roles")
async def get_user_roles(
    user_id: str,
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取用户角色"""
    try:
        service = RBACService(session)
        roles = await service.get_user_roles(user_id)
        
        return success_response(data={
            "user_id": user_id,
            "roles": [role.to_dict() for role in roles]
        })
        
    except Exception as e:
        return error_response(message=f"获取用户角色失败: {str(e)}")


# ==================== 权限验证API ====================

@router.get("/permissions/check")
async def check_permission(
    permission: str = Query(..., description="权限标识"),
    http_method: str = Query("GET", description="HTTP方法"),
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """检查权限"""
    try:
        service = RBACService(session)
        has_permission = await service.check_permission(
            current_user.user_id, permission, http_method
        )
        
        return success_response(data={
            "user_id": current_user.user_id,
            "permission": permission,
            "http_method": http_method,
            "has_permission": has_permission
        })
        
    except Exception as e:
        return error_response(message=f"权限检查失败: {str(e)}")


@router.get("/roles/check")
async def check_role(
    role_code: str = Query(..., description="角色编码"),
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """检查角色"""
    try:
        service = RBACService(session)
        has_role = await service.has_role(current_user.user_id, role_code)
        
        return success_response(data={
            "user_id": current_user.user_id,
            "role_code": role_code,
            "has_role": has_role
        })
        
    except Exception as e:
        return error_response(message=f"角色检查失败: {str(e)}")


@router.get("/admin/check")
async def check_admin(
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """检查管理员权限"""
    try:
        service = RBACService(session)
        is_admin = await service.is_admin(current_user.user_id)
        
        return success_response(data={
            "user_id": current_user.user_id,
            "is_admin": is_admin
        })
        
    except Exception as e:
        return error_response(message=f"管理员检查失败: {str(e)}")


# ==================== 操作日志API ====================

@router.get("/logs/operations")
async def get_operation_logs(
    user_id: Optional[str] = Query(None, description="用户ID"),
    start_time: Optional[datetime] = Query(None, description="开始时间"),
    end_time: Optional[datetime] = Query(None, description="结束时间"),
    page: int = Query(1, ge=1, description="页码"),
    size: int = Query(20, ge=1, le=100, description="每页大小"),
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取操作日志"""
    try:
        service = RBACService(session)
        logs, total = await service.get_operation_logs(
            user_id, start_time, end_time, page, size
        )
        
        return success_response(data={
            "logs": [log.to_dict() for log in logs],
            "pagination": {
                "page": page,
                "size": size,
                "total": total,
                "pages": (total + size - 1) // size
            }
        })
        
    except Exception as e:
        return error_response(message=f"获取操作日志失败: {str(e)}")


# ==================== 在线用户API ====================

@router.get("/online/users")
async def get_online_users(
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取在线用户"""
    try:
        service = RBACService(session)
        online_users = await service.get_online_users()
        
        return success_response(data={
            "online_count": len(online_users),
            "users": [
                {
                    "session_id": user.session_id,
                    "user_id": user.user_id,
                    "login_name": user.login_name,
                    "real_name": user.real_name,
                    "ipaddr": user.ipaddr,
                    "login_location": user.login_location,
                    "browser": user.browser,
                    "os": user.os,
                    "status": user.status,
                    "start_timestamp": user.start_timestamp.isoformat() if user.start_timestamp else None,
                    "last_access_time": user.last_access_time.isoformat() if user.last_access_time else None
                }
                for user in online_users
            ]
        })
        
    except Exception as e:
        return error_response(message=f"获取在线用户失败: {str(e)}")


# ==================== 统计信息API ====================

@router.get("/statistics")
async def get_rbac_statistics(
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """获取RBAC统计信息"""
    try:
        service = RBACService(session)
        stats = await service.get_rbac_statistics()
        
        return success_response(data=stats)
        
    except Exception as e:
        return error_response(message=f"获取统计信息失败: {str(e)}")


# ==================== 初始化数据API ====================

@router.post("/init")
async def init_rbac_data(
    session: AsyncSession = Depends(get_async_session),
    current_user: User = Depends(get_current_user)
):
    """初始化RBAC数据"""
    try:
        service = RBACService(session)
        
        # 创建默认角色
        platform_admin_role = await service.create_role({
            "role_name": "平台管理员",
            "role_code": "PLATFORM_ADMIN",
            "role_desc": "平台超级管理员，拥有所有权限",
            "biz_type": BusinessType.PLATFORM.value,
            "is_admin": True,
            "create_by": "system",
            "update_by": "system"
        })
        
        shop_admin_role = await service.create_role({
            "role_name": "商城管理员", 
            "role_code": "SHOP_ADMIN",
            "role_desc": "商城管理员，管理店铺相关功能",
            "biz_type": BusinessType.SHOP.value,
            "is_admin": False,
            "create_by": "system",
            "update_by": "system"
        })
        
        # 创建默认菜单 - 平台管理
        platform_menus = [
            {
                "menu_name": "平台管理",
                "menu_code": "platform",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform",
                "icon": "platform",
                "order_num": 1,
                "create_by": "system"
            },
            {
                "menu_name": "商户管理", 
                "menu_code": "platform:shop",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/shops",
                "component": "platform/shops/index",
                "icon": "shop",
                "order_num": 1,
                "permission": "platform:shop:list",
                "create_by": "system"
            },
            {
                "menu_name": "分类管理",
                "menu_code": "platform:category", 
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.PLATFORM.value,
                "path": "/platform/categories",
                "component": "platform/categories/index",
                "icon": "category",
                "order_num": 2,
                "permission": "platform:category:list",
                "create_by": "system"
            }
        ]
        
        # 创建默认菜单 - 商城管理
        shop_menus = [
            {
                "menu_name": "商品管理",
                "menu_code": "shop:product",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/products",
                "component": "shop/products/index", 
                "icon": "product",
                "order_num": 1,
                "permission": "shop:product:list",
                "create_by": "system"
            },
            {
                "menu_name": "订单管理",
                "menu_code": "shop:order",
                "menu_type": MenuType.MENU.value,
                "biz_type": BusinessType.SHOP.value,
                "path": "/shop/orders",
                "component": "shop/orders/index",
                "icon": "order",
                "order_num": 2,
                "permission": "shop:order:list",
                "create_by": "system"
            }
        ]
        
        created_menus = []
        for menu_data in platform_menus + shop_menus:
            menu = await service.create_menu(menu_data)
            created_menus.append(menu)
        
        return success_response(
            data={
                "roles_created": 2,
                "menus_created": len(created_menus),
                "roles": [platform_admin_role.to_dict(), shop_admin_role.to_dict()],
                "menus": [menu.to_dict() for menu in created_menus]
            },
            message="RBAC初始化数据创建成功"
        )
        
    except BusinessException as e:
        return error_response(message=str(e))
    except Exception as e:
        return error_response(message=f"初始化RBAC数据失败: {str(e)}")