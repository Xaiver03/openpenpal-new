"""
OpenPenPal Postcode 地址编码系统 API
基于FastAPI实现的完整后端服务
"""

from typing import List, Optional, Dict, Any
from fastapi import APIRouter, HTTPException, Depends, Query, Path
from pydantic import BaseModel, validator
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_, func, text
import re

from app.core.database import get_db
from app.models.postcode import (
    PostcodeSchool, PostcodeArea, PostcodeBuilding, PostcodeRoom,
    PostcodeCourierPermission, PostcodeFeedback, PostcodeStats
)
from app.schemas.postcode import (
    SchoolCreate, SchoolUpdate, SchoolResponse,
    AreaCreate, AreaUpdate, AreaResponse,
    BuildingCreate, BuildingUpdate, BuildingResponse,
    RoomCreate, RoomUpdate, RoomResponse,
    AddressSearchResult, AddressHierarchy,
    CourierPermissionCreate, CourierPermissionResponse,
    FeedbackCreate, FeedbackResponse, FeedbackReview,
    PostcodeStatsResponse, PostcodeValidation
)
from app.utils.auth import get_current_user

router = APIRouter(prefix="/postcode", tags=["postcode"])

# 临时实现 require_courier_level 函数
def require_courier_level(level: int):
    """要求特定的信使等级"""
    def dependency(current_user: str = Depends(get_current_user)):
        # TODO: 实现实际的信使等级检查逻辑
        # 这里暂时只返回当前用户，应该检查用户是否具有所需的信使等级
        return current_user
    return dependency

# ==================== 编码解析与查询 ====================

@router.get("/{code}", response_model=AddressSearchResult)
async def get_address_by_postcode(
    code: str = Path(..., description="6位Postcode编码", pattern="^[A-Z0-9]{6}$"),
    db: Session = Depends(get_db)
):
    """根据6位编码查询完整地址信息"""
    
    # 解析编码结构
    if len(code) != 6:
        raise HTTPException(status_code=400, detail="编码必须为6位")
    
    school_code = code[:2]
    area_code = code[2:3]
    building_code = code[3:4]
    room_code = code[4:6]
    
    # 查询完整地址信息
    query = db.query(
        PostcodeRoom,
        PostcodeBuilding,
        PostcodeArea,
        PostcodeSchool
    ).join(
        PostcodeBuilding, and_(
            PostcodeRoom.school_code == PostcodeBuilding.school_code,
            PostcodeRoom.area_code == PostcodeBuilding.area_code,
            PostcodeRoom.building_code == PostcodeBuilding.code
        )
    ).join(
        PostcodeArea, and_(
            PostcodeBuilding.school_code == PostcodeArea.school_code,
            PostcodeBuilding.area_code == PostcodeArea.code
        )
    ).join(
        PostcodeSchool, PostcodeArea.school_code == PostcodeSchool.code
    ).filter(
        PostcodeRoom.full_postcode == code
    )
    
    result = query.first()
    if not result:
        raise HTTPException(status_code=404, detail="未找到对应的地址信息")
    
    room, building, area, school = result
    
    # 构建响应数据
    hierarchy = AddressHierarchy(
        school=SchoolResponse.from_orm(school),
        area=AreaResponse.from_orm(area),
        building=BuildingResponse.from_orm(building),
        room=RoomResponse.from_orm(room)
    )
    
    full_address = f"{school.name} {area.name} {building.name} {room.name}"
    
    return AddressSearchResult(
        postcode=code,
        fullAddress=full_address,
        hierarchy=hierarchy,
        matchScore=1.0
    )

@router.get("/search", response_model=List[AddressSearchResult])
async def search_addresses(
    q: str = Query(..., min_length=2, description="搜索关键词"),
    limit: int = Query(10, ge=1, le=50, description="返回结果数量限制"),
    db: Session = Depends(get_db)
):
    """地址模糊搜索"""
    
    # 使用PostgreSQL全文搜索
    search_query = f"%{q}%"
    
    query = db.query(
        PostcodeRoom,
        PostcodeBuilding,
        PostcodeArea,
        PostcodeSchool,
        # 计算匹配分数
        func.greatest(
            func.similarity(PostcodeRoom.name, q),
            func.similarity(PostcodeBuilding.name, q),
            func.similarity(PostcodeArea.name, q),
            func.similarity(PostcodeSchool.name, q),
            func.similarity(PostcodeRoom.full_postcode, q)
        ).label('match_score')
    ).join(
        PostcodeBuilding, and_(
            PostcodeRoom.school_code == PostcodeBuilding.school_code,
            PostcodeRoom.area_code == PostcodeBuilding.area_code,
            PostcodeRoom.building_code == PostcodeBuilding.code
        )
    ).join(
        PostcodeArea, and_(
            PostcodeBuilding.school_code == PostcodeArea.school_code,
            PostcodeBuilding.area_code == PostcodeArea.code
        )
    ).join(
        PostcodeSchool, PostcodeArea.school_code == PostcodeSchool.code
    ).filter(
        or_(
            PostcodeRoom.name.ilike(search_query),
            PostcodeRoom.full_postcode.ilike(search_query),
            PostcodeBuilding.name.ilike(search_query),
            PostcodeArea.name.ilike(search_query),
            PostcodeSchool.name.ilike(search_query),
            PostcodeSchool.full_name.ilike(search_query)
        )
    ).order_by(
        text('match_score DESC'),
        PostcodeRoom.full_postcode
    ).limit(limit)
    
    results = query.all()
    
    search_results = []
    for result in results:
        room, building, area, school, match_score = result
        
        hierarchy = AddressHierarchy(
            school=SchoolResponse.from_orm(school),
            area=AreaResponse.from_orm(area),
            building=BuildingResponse.from_orm(building),
            room=RoomResponse.from_orm(room)
        )
        
        full_address = f"{school.name} {area.name} {building.name} {room.name}"
        
        search_results.append(AddressSearchResult(
            postcode=room.full_postcode,
            fullAddress=full_address,
            hierarchy=hierarchy,
            matchScore=float(match_score or 0.5)
        ))
    
    return search_results

# ==================== 学校管理 ====================

@router.get("/schools", response_model=List[SchoolResponse])
async def get_schools(
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取学校列表"""
    schools = db.query(PostcodeSchool).filter(
        PostcodeSchool.status == 'active'
    ).order_by(PostcodeSchool.name).all()
    
    return [SchoolResponse.from_orm(school) for school in schools]

@router.post("/schools", response_model=SchoolResponse)
async def create_school(
    school_data: SchoolCreate,
    current_user: str = Depends(require_courier_level(4)),
    db: Session = Depends(get_db)
):
    """创建新学校（四级信使权限）"""
    
    # 检查编码是否已存在
    existing = db.query(PostcodeSchool).filter(
        PostcodeSchool.code == school_data.code.upper()
    ).first()
    
    if existing:
        raise HTTPException(status_code=400, detail="学校编码已存在")
    
    school = PostcodeSchool(
        code=school_data.code.upper(),
        name=school_data.name,
        full_name=school_data.full_name,
        managed_by=current_user.get('id', 'unknown')
    )
    
    db.add(school)
    db.commit()
    db.refresh(school)
    
    return SchoolResponse.from_orm(school)

@router.get("/schools/{code}/areas", response_model=List[AreaResponse])
async def get_school_areas(
    code: str = Path(..., description="学校编码"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取学校的片区列表"""
    areas = db.query(PostcodeArea).filter(
        and_(
            PostcodeArea.school_code == code.upper(),
            PostcodeArea.status == 'active'
        )
    ).order_by(PostcodeArea.code).all()
    
    return [AreaResponse.from_orm(area) for area in areas]

@router.post("/schools/{code}/areas", response_model=AreaResponse)
async def create_area(
    area_data: AreaCreate,
    code: str = Path(..., description="学校编码"),
    current_user: str = Depends(require_courier_level(3)),
    db: Session = Depends(get_db)
):
    """创建新片区（三级信使权限）"""
    
    # 验证学校存在
    school = db.query(PostcodeSchool).filter(
        PostcodeSchool.code == code.upper()
    ).first()
    if not school:
        raise HTTPException(status_code=404, detail="学校不存在")
    
    # 检查片区编码是否已存在
    existing = db.query(PostcodeArea).filter(
        and_(
            PostcodeArea.school_code == code.upper(),
            PostcodeArea.code == area_data.code
        )
    ).first()
    
    if existing:
        raise HTTPException(status_code=400, detail="片区编码已存在")
    
    area = PostcodeArea(
        school_code=code.upper(),
        code=area_data.code,
        name=area_data.name,
        description=area_data.description,
        managed_by=current_user.get('id', 'unknown')
    )
    
    db.add(area)
    db.commit()
    db.refresh(area)
    
    return AreaResponse.from_orm(area)

@router.get("/schools/{code}/areas/{area}/buildings", response_model=List[BuildingResponse])
async def get_area_buildings(
    code: str = Path(..., description="学校编码"),
    area: str = Path(..., description="片区编码"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取片区的楼栋列表"""
    buildings = db.query(PostcodeBuilding).filter(
        and_(
            PostcodeBuilding.school_code == code.upper(),
            PostcodeBuilding.area_code == area,
            PostcodeBuilding.status == 'active'
        )
    ).order_by(PostcodeBuilding.code).all()
    
    return [BuildingResponse.from_orm(building) for building in buildings]

@router.post("/schools/{code}/areas/{area}/buildings", response_model=BuildingResponse)
async def create_building(
    building_data: BuildingCreate,
    code: str = Path(..., description="学校编码"),
    area: str = Path(..., description="片区编码"),
    current_user: str = Depends(require_courier_level(2)),
    db: Session = Depends(get_db)
):
    """创建新楼栋（二级信使权限）"""
    
    # 验证片区存在
    area_exists = db.query(PostcodeArea).filter(
        and_(
            PostcodeArea.school_code == code.upper(),
            PostcodeArea.code == area
        )
    ).first()
    if not area_exists:
        raise HTTPException(status_code=404, detail="片区不存在")
    
    # 检查楼栋编码是否已存在
    existing = db.query(PostcodeBuilding).filter(
        and_(
            PostcodeBuilding.school_code == code.upper(),
            PostcodeBuilding.area_code == area,
            PostcodeBuilding.code == building_data.code
        )
    ).first()
    
    if existing:
        raise HTTPException(status_code=400, detail="楼栋编码已存在")
    
    building = PostcodeBuilding(
        school_code=code.upper(),
        area_code=area,
        code=building_data.code,
        name=building_data.name,
        type=building_data.type,
        floors=building_data.floors,
        managed_by=current_user.get('id', 'unknown')
    )
    
    db.add(building)
    db.commit()
    db.refresh(building)
    
    return BuildingResponse.from_orm(building)

@router.get("/schools/{code}/areas/{area}/buildings/{building}/rooms", response_model=List[RoomResponse])
async def get_building_rooms(
    code: str = Path(..., description="学校编码"),
    area: str = Path(..., description="片区编码"),
    building: str = Path(..., description="楼栋编码"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取楼栋的房间列表"""
    rooms = db.query(PostcodeRoom).filter(
        and_(
            PostcodeRoom.school_code == code.upper(),
            PostcodeRoom.area_code == area,
            PostcodeRoom.building_code == building,
            PostcodeRoom.status == 'active'
        )
    ).order_by(PostcodeRoom.code).all()
    
    return [RoomResponse.from_orm(room) for room in rooms]

@router.post("/schools/{code}/areas/{area}/buildings/{building}/rooms", response_model=RoomResponse)
async def create_room(
    room_data: RoomCreate,
    code: str = Path(..., description="学校编码"),
    area: str = Path(..., description="片区编码"),
    building: str = Path(..., description="楼栋编码"),
    current_user: str = Depends(require_courier_level(1)),
    db: Session = Depends(get_db)
):
    """创建新房间（一级信使权限）"""
    
    # 验证楼栋存在
    building_exists = db.query(PostcodeBuilding).filter(
        and_(
            PostcodeBuilding.school_code == code.upper(),
            PostcodeBuilding.area_code == area,
            PostcodeBuilding.code == building
        )
    ).first()
    if not building_exists:
        raise HTTPException(status_code=404, detail="楼栋不存在")
    
    # 检查房间编码是否已存在
    existing = db.query(PostcodeRoom).filter(
        and_(
            PostcodeRoom.school_code == code.upper(),
            PostcodeRoom.area_code == area,
            PostcodeRoom.building_code == building,
            PostcodeRoom.code == room_data.code
        )
    ).first()
    
    if existing:
        raise HTTPException(status_code=400, detail="房间编码已存在")
    
    room = PostcodeRoom(
        school_code=code.upper(),
        area_code=area,
        building_code=building,
        code=room_data.code,
        name=room_data.name,
        type=room_data.type,
        capacity=room_data.capacity,
        floor=room_data.floor,
        managed_by=current_user.get('id', 'unknown')
    )
    
    db.add(room)
    db.commit()
    db.refresh(room)
    
    return RoomResponse.from_orm(room)

# ==================== 权限管理 ====================

@router.get("/permissions/{courier_id}", response_model=CourierPermissionResponse)
async def get_courier_permissions(
    courier_id: str = Path(..., description="信使ID"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取信使的Postcode权限"""
    
    permission = db.query(PostcodeCourierPermission).filter(
        PostcodeCourierPermission.courier_id == courier_id
    ).first()
    
    if not permission:
        raise HTTPException(status_code=404, detail="未找到权限信息")
    
    return CourierPermissionResponse.from_orm(permission)

@router.post("/permissions", response_model=CourierPermissionResponse)
async def create_courier_permission(
    permission_data: CourierPermissionCreate,
    current_user: str = Depends(require_courier_level(4)),
    db: Session = Depends(get_db)
):
    """分配Postcode权限给信使"""
    
    # 检查是否已存在
    existing = db.query(PostcodeCourierPermission).filter(
        PostcodeCourierPermission.courier_id == permission_data.courier_id
    ).first()
    
    if existing:
        raise HTTPException(status_code=400, detail="权限已存在，请使用更新接口")
    
    permission = PostcodeCourierPermission(
        courier_id=permission_data.courier_id,
        level=permission_data.level,
        prefix_patterns=permission_data.prefix_patterns,
        can_manage=permission_data.can_manage,
        can_create=permission_data.can_create,
        can_review=permission_data.can_review
    )
    
    db.add(permission)
    db.commit()
    db.refresh(permission)
    
    return CourierPermissionResponse.from_orm(permission)

# ==================== 统计分析 ====================

@router.get("/stats", response_model=List[PostcodeStatsResponse])
async def get_postcode_stats(
    postcode: Optional[str] = Query(None, description="指定Postcode"),
    limit: int = Query(20, ge=1, le=100, description="返回数量"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取Postcode使用统计"""
    
    query = db.query(PostcodeStats)
    
    if postcode:
        query = query.filter(PostcodeStats.postcode == postcode.upper())
    
    stats = query.order_by(
        PostcodeStats.popularity_score.desc()
    ).limit(limit).all()
    
    return [PostcodeStatsResponse.from_orm(stat) for stat in stats]

@router.get("/stats/popular", response_model=List[PostcodeStatsResponse])
async def get_popular_addresses(
    limit: int = Query(20, ge=1, le=50, description="返回数量"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取热门地址排行"""
    
    stats = db.query(PostcodeStats).order_by(
        PostcodeStats.delivery_count.desc(),
        PostcodeStats.popularity_score.desc()
    ).limit(limit).all()
    
    return [PostcodeStatsResponse.from_orm(stat) for stat in stats]

@router.get("/stats/problematic", response_model=List[PostcodeStatsResponse])
async def get_problematic_addresses(
    limit: int = Query(20, ge=1, le=50, description="返回数量"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取投递失败率高的地址"""
    
    stats = db.query(PostcodeStats).filter(
        PostcodeStats.error_count > 0
    ).order_by(
        (PostcodeStats.error_count / func.greatest(PostcodeStats.delivery_count, 1)).desc()
    ).limit(limit).all()
    
    return [PostcodeStatsResponse.from_orm(stat) for stat in stats]

# ==================== 工具接口 ====================

@router.post("/tools/validate", response_model=PostcodeValidation)
async def validate_postcode(
    code: str = Query(..., description="要验证的Postcode"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """验证Postcode格式和存在性"""
    
    errors = []
    
    # 格式验证
    if not code:
        errors.append("编码不能为空")
    elif len(code) != 6:
        errors.append("编码必须为6位")
    elif not re.match(r'^[A-Z0-9]{6}$', code.upper()):
        errors.append("编码只能包含大写字母和数字")
    
    is_valid = len(errors) == 0
    exists = False
    
    if is_valid:
        # 检查是否存在
        room = db.query(PostcodeRoom).filter(
            PostcodeRoom.full_postcode == code.upper()
        ).first()
        exists = room is not None
        
        if not exists:
            errors.append("地址不存在")
            is_valid = False
    
    return PostcodeValidation(
        code=code.upper(),
        is_valid=is_valid,
        exists=exists,
        errors=errors
    )