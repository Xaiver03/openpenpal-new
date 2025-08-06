"""
RBAC (Role-Based Access Control) 角色权限管理模型

基于mall4cloud的RBAC设计，支持：
1. 多层级菜单管理
2. 细粒度权限控制
3. 角色权限分配
4. 用户角色映射
5. 业务类型分离（商城管理vs平台管理）
"""

from sqlalchemy import Column, String, Text, DateTime, Boolean, Integer, ForeignKey, JSON, Index, UniqueConstraint
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from datetime import datetime
from enum import Enum as PyEnum
from app.core.database import Base


class BusinessType(PyEnum):
    """业务类型枚举"""
    PLATFORM = 1      # 平台管理（超级管理员）
    SHOP = 2          # 商城管理（商户）
    COMMON = 3        # 通用功能


class MenuType(PyEnum):
    """菜单类型枚举"""
    MENU = 1          # 菜单
    BUTTON = 2        # 按钮
    API = 3           # API接口


class PermissionType(PyEnum):
    """权限类型枚举"""
    READ = "read"       # 读取权限
    WRITE = "write"     # 写入权限
    DELETE = "delete"   # 删除权限
    EXECUTE = "execute" # 执行权限


# ==================== 菜单管理 ====================

class SysMenu(Base):
    """系统菜单表"""
    __tablename__ = "sys_menu"
    
    # 主键
    menu_id = Column(Integer, primary_key=True, autoincrement=True, comment="菜单ID")
    
    # 菜单信息
    parent_id = Column(Integer, ForeignKey("sys_menu.menu_id"), nullable=True, comment="父菜单ID")
    menu_name = Column(String(50), nullable=False, comment="菜单名称")
    menu_code = Column(String(100), unique=True, nullable=False, comment="菜单编码")
    menu_type = Column(Integer, default=MenuType.MENU.value, comment="菜单类型(1菜单 2按钮 3接口)")
    biz_type = Column(Integer, default=BusinessType.COMMON.value, comment="业务类型(1平台 2商城 3通用)")
    
    # 路由信息
    path = Column(String(200), comment="路由地址")
    component = Column(String(255), comment="组件路径")
    redirect = Column(String(255), comment="重定向地址")
    
    # 显示信息
    icon = Column(String(100), comment="菜单图标")
    order_num = Column(Integer, default=0, comment="显示顺序")
    is_hidden = Column(Boolean, default=False, comment="是否隐藏")
    is_cache = Column(Boolean, default=False, comment="是否缓存")
    is_affix = Column(Boolean, default=False, comment="是否固定标签")
    
    # 权限信息
    permission = Column(String(200), comment="权限标识")
    http_method = Column(String(10), comment="HTTP方法")
    api_url = Column(String(500), comment="API地址")
    
    # 状态信息
    status = Column(Integer, default=1, comment="菜单状态(0停用 1正常)")
    
    # 扩展信息
    meta_info = Column(JSON, comment="元信息(JSON格式)")
    remark = Column(String(500), comment="备注")
    
    # 时间戳
    create_time = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    update_time = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    create_by = Column(String(50), comment="创建者")
    update_by = Column(String(50), comment="更新者")
    
    # 关联关系
    parent = relationship("SysMenu", remote_side=[menu_id], back_populates="children")
    children = relationship("SysMenu", back_populates="parent", cascade="all, delete-orphan")
    role_menus = relationship("SysRoleMenu", back_populates="menu", cascade="all, delete-orphan")
    
    # 索引
    __table_args__ = (
        Index('idx_menu_parent_id', 'parent_id'),
        Index('idx_menu_biz_type', 'biz_type'),
        Index('idx_menu_status', 'status'),
    )
    
    def __repr__(self):
        return f"<SysMenu(menu_id={self.menu_id}, menu_name={self.menu_name})>"
    
    def to_dict(self, include_children=False):
        """转换为字典格式"""
        data = {
            "menu_id": self.menu_id,
            "parent_id": self.parent_id,
            "menu_name": self.menu_name,
            "menu_code": self.menu_code,
            "menu_type": self.menu_type,
            "biz_type": self.biz_type,
            "path": self.path,
            "component": self.component,
            "redirect": self.redirect,
            "icon": self.icon,
            "order_num": self.order_num,
            "is_hidden": self.is_hidden,
            "is_cache": self.is_cache,
            "is_affix": self.is_affix,
            "permission": self.permission,
            "http_method": self.http_method,
            "api_url": self.api_url,
            "status": self.status,
            "meta_info": self.meta_info,
            "remark": self.remark,
            "create_time": self.create_time.isoformat() if self.create_time else None,
            "update_time": self.update_time.isoformat() if self.update_time else None,
            "create_by": self.create_by,
            "update_by": self.update_by
        }
        
        if include_children and self.children:
            data["children"] = [child.to_dict(include_children) for child in self.children]
        
        return data


# ==================== 角色管理 ====================

class SysRole(Base):
    """系统角色表"""
    __tablename__ = "sys_role"
    
    # 主键
    role_id = Column(Integer, primary_key=True, autoincrement=True, comment="角色ID")
    
    # 角色信息
    role_name = Column(String(30), nullable=False, comment="角色名称")
    role_code = Column(String(100), unique=True, nullable=False, comment="角色编码")
    role_desc = Column(String(500), comment="角色描述")
    biz_type = Column(Integer, default=BusinessType.COMMON.value, comment="业务类型")
    
    # 角色配置
    is_admin = Column(Boolean, default=False, comment="是否管理员角色")
    data_scope = Column(Integer, default=1, comment="数据权限范围")
    
    # 状态信息
    status = Column(Integer, default=1, comment="角色状态(0停用 1正常)")
    sort_order = Column(Integer, default=0, comment="显示顺序")
    
    # 时间戳
    create_time = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    update_time = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    create_by = Column(String(50), comment="创建者")
    update_by = Column(String(50), comment="更新者")
    
    # 关联关系
    role_menus = relationship("SysRoleMenu", back_populates="role", cascade="all, delete-orphan")
    user_roles = relationship("SysUserRole", back_populates="role", cascade="all, delete-orphan")
    
    # 索引
    __table_args__ = (
        Index('idx_role_biz_type', 'biz_type'),
        Index('idx_role_status', 'status'),
    )
    
    def __repr__(self):
        return f"<SysRole(role_id={self.role_id}, role_name={self.role_name})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "role_id": self.role_id,
            "role_name": self.role_name,
            "role_code": self.role_code,
            "role_desc": self.role_desc,
            "biz_type": self.biz_type,
            "is_admin": self.is_admin,
            "data_scope": self.data_scope,
            "status": self.status,
            "sort_order": self.sort_order,
            "create_time": self.create_time.isoformat() if self.create_time else None,
            "update_time": self.update_time.isoformat() if self.update_time else None,
            "create_by": self.create_by,
            "update_by": self.update_by
        }


# ==================== 角色菜单关联 ====================

class SysRoleMenu(Base):
    """角色菜单关联表"""
    __tablename__ = "sys_role_menu"
    
    # 联合主键
    role_id = Column(Integer, ForeignKey("sys_role.role_id"), primary_key=True, comment="角色ID")
    menu_id = Column(Integer, ForeignKey("sys_menu.menu_id"), primary_key=True, comment="菜单ID")
    
    # 权限类型
    permission_type = Column(String(20), default=PermissionType.READ.value, comment="权限类型")
    
    # 时间戳
    create_time = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    create_by = Column(String(50), comment="创建者")
    
    # 关联关系
    role = relationship("SysRole", back_populates="role_menus")
    menu = relationship("SysMenu", back_populates="role_menus")
    
    def __repr__(self):
        return f"<SysRoleMenu(role_id={self.role_id}, menu_id={self.menu_id})>"


# ==================== 用户表扩展 ====================

class SysUser(Base):
    """系统用户表（扩展原用户表）"""
    __tablename__ = "sys_users"
    
    # 主键
    user_id = Column(String(50), primary_key=True, comment="用户ID")
    
    # 基础信息
    username = Column(String(50), unique=True, nullable=False, comment="用户名")
    email = Column(String(100), unique=True, nullable=False, comment="邮箱")
    phone = Column(String(20), comment="手机号")
    real_name = Column(String(50), comment="真实姓名")
    nickname = Column(String(50), comment="昵称")
    
    # 用户状态
    status = Column(Integer, default=1, comment="用户状态(0停用 1正常)")
    user_type = Column(Integer, default=1, comment="用户类型(1系统用户 2商户用户)")
    
    # 最后登录
    last_login_time = Column(DateTime(timezone=True), comment="最后登录时间")
    last_login_ip = Column(String(45), comment="最后登录IP")
    
    # 密码信息
    password_hash = Column(String(255), comment="密码哈希")
    salt = Column(String(50), comment="密码盐值")
    password_update_time = Column(DateTime(timezone=True), comment="密码更新时间")
    
    # 账号安全
    login_fail_count = Column(Integer, default=0, comment="登录失败次数")
    lock_time = Column(DateTime(timezone=True), comment="锁定时间")
    
    # 时间戳
    create_time = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    update_time = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), comment="更新时间")
    create_by = Column(String(50), comment="创建者")
    update_by = Column(String(50), comment="更新者")
    
    # 关联关系
    user_roles = relationship("SysUserRole", back_populates="user", cascade="all, delete-orphan")
    
    # 索引
    __table_args__ = (
        Index('idx_user_status', 'status'),
        Index('idx_user_type', 'user_type'),
        UniqueConstraint('username', name='uk_user_username'),
        UniqueConstraint('email', name='uk_user_email'),
    )
    
    def __repr__(self):
        return f"<SysUser(user_id={self.user_id}, username={self.username})>"
    
    def to_dict(self, include_sensitive=False):
        """转换为字典格式"""
        data = {
            "user_id": self.user_id,
            "username": self.username,
            "email": self.email,
            "phone": self.phone,
            "real_name": self.real_name,
            "nickname": self.nickname,
            "status": self.status,
            "user_type": self.user_type,
            "last_login_time": self.last_login_time.isoformat() if self.last_login_time else None,
            "last_login_ip": self.last_login_ip,
            "create_time": self.create_time.isoformat() if self.create_time else None,
            "update_time": self.update_time.isoformat() if self.update_time else None,
            "create_by": self.create_by,
            "update_by": self.update_by
        }
        
        if include_sensitive:
            data.update({
                "password_update_time": self.password_update_time.isoformat() if self.password_update_time else None,
                "login_fail_count": self.login_fail_count,
                "lock_time": self.lock_time.isoformat() if self.lock_time else None,
            })
        
        return data


# ==================== 用户角色关联 ====================

class SysUserRole(Base):
    """用户角色关联表"""
    __tablename__ = "sys_user_role"
    
    # 联合主键
    user_id = Column(String(50), ForeignKey("sys_users.user_id"), primary_key=True, comment="用户ID")
    role_id = Column(Integer, ForeignKey("sys_role.role_id"), primary_key=True, comment="角色ID")
    
    # 时间戳
    create_time = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
    create_by = Column(String(50), comment="创建者")
    
    # 关联关系
    user = relationship("SysUser", back_populates="user_roles")
    role = relationship("SysRole", back_populates="user_roles")
    
    def __repr__(self):
        return f"<SysUserRole(user_id={self.user_id}, role_id={self.role_id})>"


# ==================== 权限操作日志 ====================

class SysOperLog(Base):
    """系统操作日志表"""
    __tablename__ = "sys_oper_log"
    
    # 主键
    oper_id = Column(Integer, primary_key=True, autoincrement=True, comment="日志主键")
    
    # 操作信息
    title = Column(String(50), comment="模块标题")
    business_type = Column(Integer, comment="业务类型")
    method = Column(String(100), comment="方法名称")
    request_method = Column(String(10), comment="请求方式")
    
    # 操作人员
    oper_name = Column(String(50), comment="操作人员")
    oper_user_id = Column(String(50), comment="操作用户ID")
    
    # 请求信息
    oper_url = Column(String(255), comment="请求URL")
    oper_ip = Column(String(45), comment="主机地址")
    oper_location = Column(String(255), comment="操作地点")
    oper_param = Column(Text, comment="请求参数")
    
    # 响应信息
    json_result = Column(Text, comment="返回参数")
    status = Column(Integer, default=0, comment="操作状态(0正常 1异常)")
    error_msg = Column(String(2000), comment="错误消息")
    
    # 时间信息
    oper_time = Column(DateTime(timezone=True), server_default=func.now(), comment="操作时间")
    cost_time = Column(Integer, comment="消耗时间(毫秒)")
    
    # 索引
    __table_args__ = (
        Index('idx_oper_time', 'oper_time'),
        Index('idx_oper_user', 'oper_user_id'),
        Index('idx_oper_status', 'status'),
    )
    
    def __repr__(self):
        return f"<SysOperLog(oper_id={self.oper_id}, title={self.title})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "oper_id": self.oper_id,
            "title": self.title,
            "business_type": self.business_type,
            "method": self.method,
            "request_method": self.request_method,
            "oper_name": self.oper_name,
            "oper_user_id": self.oper_user_id,
            "oper_url": self.oper_url,
            "oper_ip": self.oper_ip,
            "oper_location": self.oper_location,
            "oper_param": self.oper_param,
            "json_result": self.json_result,
            "status": self.status,
            "error_msg": self.error_msg,
            "oper_time": self.oper_time.isoformat() if self.oper_time else None,
            "cost_time": self.cost_time
        }


# ==================== 在线用户表 ====================

class SysUserOnline(Base):
    """在线用户表"""
    __tablename__ = "sys_user_online"
    
    # 主键
    session_id = Column(String(50), primary_key=True, comment="会话编号")
    
    # 用户信息
    user_id = Column(String(50), comment="用户ID")
    login_name = Column(String(50), comment="登录账号")
    real_name = Column(String(50), comment="用户姓名")
    
    # 会话信息
    ipaddr = Column(String(45), comment="登录IP地址")
    login_location = Column(String(255), comment="登录地点")
    browser = Column(String(50), comment="浏览器类型")
    os = Column(String(50), comment="操作系统")
    
    # 时间信息
    status = Column(String(10), default="on_line", comment="在线状态")
    start_timestamp = Column(DateTime(timezone=True), comment="session创建时间")
    last_access_time = Column(DateTime(timezone=True), comment="session最后访问时间")
    expire_time = Column(Integer, comment="超时时间(分钟)")
    
    # 索引
    __table_args__ = (
        Index('idx_online_user', 'user_id'),
        Index('idx_online_status', 'status'),
    )
    
    def __repr__(self):
        return f"<SysUserOnline(session_id={self.session_id}, login_name={self.login_name})>"