from fastapi import APIRouter, Depends, HTTPException, status, Query, Request
from sqlalchemy.orm import Session
from typing import Optional, List
from datetime import datetime
import math

from app.core.database import get_db, get_async_session
from app.models.letter import Letter, LetterStatus
from app.schemas.letter import (
    LetterCreate, LetterResponse, LetterCreateResponse, 
    LetterStatusUpdate, LetterStatusUpdateResponse, LetterUpdate,
    LetterListResponse, SuccessResponse, ErrorResponse
)
from app.utils.code_generator import generate_unique_letter_code
from app.utils.auth import get_current_user, get_current_user_info
from app.utils.websocket_client import notify_letter_created, notify_letter_status_update, notify_letter_read
from app.utils.user_service import get_user_nickname
from app.utils.status_validator import validate_letter_status_transition_or_raise
from app.utils.read_logger import log_letter_read, ReadLogManager
from app.utils.query_optimizer import get_optimized_user_letters, LetterQueryOptimizer
from app.utils.cache_manager import LetterCacheService

router = APIRouter()

@router.post("", response_model=SuccessResponse, summary="创建信件")
async def create_letter(
    letter_data: LetterCreate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    创建新的信件
    
    - **title**: 信件标题（必填，1-200字符）
    - **content**: 信件内容（必填）
    - **receiver_hint**: 接收者提示信息（可选）
    - **anonymous**: 是否匿名（默认false）
    - **priority**: 优先级（normal/urgent，默认normal）
    - **delivery_instructions**: 投递说明（可选）
    """
    try:
        # 生成唯一编号
        letter_id = generate_unique_letter_code(db)
        
        # 获取真实的用户昵称
        sender_nickname = await get_user_nickname(current_user)
        
        # 创建信件对象
        new_letter = Letter(
            id=letter_id,
            title=letter_data.title,
            content=letter_data.content,
            sender_id=current_user,
            sender_nickname=sender_nickname,
            receiver_hint=letter_data.receiver_hint,
            anonymous=letter_data.anonymous,
            priority=letter_data.priority,
            delivery_instructions=letter_data.delivery_instructions,
            status=LetterStatus.DRAFT
        )
        
        # 保存到数据库
        db.add(new_letter)
        db.commit()
        db.refresh(new_letter)
        
        # 发送WebSocket通知
        try:
            await notify_letter_created(new_letter.id, current_user)
        except Exception as e:
            # WebSocket通知失败不应该影响主业务
            print(f"WebSocket notification failed: {e}")
        
        # 返回创建结果
        return SuccessResponse(
            data={
                "letter_id": new_letter.id,
                "status": new_letter.status.value,
                "created_at": new_letter.created_at.isoformat()
            }
        )
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"创建信件失败: {str(e)}"
        )

@router.get("/{letter_id}", response_model=SuccessResponse, summary="获取信件详情")
async def get_letter(
    letter_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    根据信件ID获取信件详情
    
    - **letter_id**: 信件编号
    
    只有信件发送者可以查看完整内容
    """
    # 尝试从缓存获取
    cached_letter = await LetterCacheService.get_cached_letter_detail(letter_id)
    if cached_letter and cached_letter.get("sender_id") == current_user:
        return SuccessResponse(data=cached_letter)
    
    # 查询信件
    letter = db.query(Letter).filter(Letter.id == letter_id).first()
    
    if not letter:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"信件 {letter_id} 不存在"
        )
    
    # 权限检查：只有发送者可以查看
    if letter.sender_id != current_user:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="无权限查看此信件"
        )
    
    # 转换为响应格式
    letter_dict = letter.to_dict()
    
    # 缓存结果
    try:
        await LetterCacheService.cache_letter_detail(letter_id, letter_dict)
    except Exception as e:
        print(f"Cache error: {e}")
    
    return SuccessResponse(data=letter_dict)

@router.put("/{letter_id}/status", response_model=SuccessResponse, summary="更新信件状态")
async def update_letter_status(
    letter_id: str,
    status_data: LetterStatusUpdate,
    current_user_info: dict = Depends(get_current_user_info),
    db: Session = Depends(get_db)
):
    """
    更新信件状态
    
    - **letter_id**: 信件编号
    - **status**: 新状态
    - **location**: 当前位置（可选）
    - **note**: 状态更新备注（可选）
    
    状态流转：draft → generated → collected → in_transit → delivered/failed
    """
    current_user = current_user_info.get("user_id")
    user_role = current_user_info.get("role", "user")
    
    # 查询信件
    letter = db.query(Letter).filter(Letter.id == letter_id).first()
    
    if not letter:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"信件 {letter_id} 不存在"
        )
    
    # 权限检查：用户只能更新自己的信件状态，除非是信使或管理员
    if user_role == "user" and letter.sender_id != current_user:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="无权限更新此信件状态"
        )
    
    # 状态转换验证
    validate_letter_status_transition_or_raise(
        current_status=letter.status,
        new_status=status_data.status,
        user_role=user_role
    )
    
    try:
        # 更新状态
        old_status = letter.status
        letter.status = status_data.status
        
        db.commit()
        db.refresh(letter)
        
        # 发送WebSocket事件通知
        try:
            await notify_letter_status_update(letter_id, status_data.status.value, letter.sender_id)
        except Exception as e:
            # WebSocket通知失败不应该影响主业务
            print(f"WebSocket notification failed: {e}")
        
        return SuccessResponse(
            data={
                "letter_id": letter.id,
                "old_status": old_status.value,
                "new_status": letter.status.value,
                "updated_at": letter.updated_at.isoformat()
            }
        )
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"更新状态失败: {str(e)}"
        )

@router.get("/user/{user_id}", response_model=SuccessResponse, summary="获取用户信件列表")
async def get_user_letters(
    user_id: str,
    current_user: str = Depends(get_current_user),
    status_filter: Optional[str] = Query(None, description="状态过滤 (draft/generated/collected/in_transit/delivered/failed)"),
    page: int = Query(1, ge=1, description="页码"),
    limit: int = Query(10, ge=1, le=100, description="每页数量"),
    db: Session = Depends(get_db)
):
    """
    获取指定用户的信件列表
    
    - **user_id**: 用户ID  
    - **status**: 状态过滤（可选）
    - **page**: 页码（默认1）
    - **limit**: 每页数量（默认10，最大100）
    
    只能查看自己的信件列表
    """
    # 权限检查：只能查看自己的信件
    if user_id != current_user:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="无权限查看此用户的信件"
        )
    
    # 使用优化的查询
    try:
        status_enum = None
        if status_filter:
            try:
                status_enum = LetterStatus(status_filter)
            except ValueError:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=f"无效的状态过滤器: {status_filter}"
                )
        
        # 尝试从缓存获取
        cached_result = await LetterCacheService.get_cached_user_letters(
            user_id, status_filter, page
        )
        if cached_result:
            return SuccessResponse(data=cached_result)
        
        # 使用优化查询
        result = get_optimized_user_letters(
            db=db,
            user_id=user_id,
            status_filter=status_enum,
            page=page,
            limit=limit
        )
        
        # 缓存结果
        try:
            await LetterCacheService.cache_user_letters(
                user_id, status_filter, page, result
            )
        except Exception as e:
            print(f"Cache error: {e}")
        
        return SuccessResponse(data=result)
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"查询信件列表失败: {str(e)}"
        )

@router.get("/read/{code}", response_model=SuccessResponse, summary="通过编号读取信件")
async def read_letter_by_code(
    code: str,
    request: Request,
    read_duration: Optional[int] = Query(None, description="阅读时长（秒）"),
    is_complete: Optional[bool] = Query(True, description="是否完整阅读"),
    db: Session = Depends(get_db)
):
    """
    通过信件编号读取信件内容（公开接口，不需要认证）
    
    - **code**: 信件编号
    
    此接口供接收者通过扫码/输入编号来阅读信件
    """
    # 查询信件
    letter = db.query(Letter).filter(Letter.id == code).first()
    
    if not letter:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"信件编号 {code} 不存在"
        )
    
    # 检查信件状态（只有已投递的信件才能阅读）
    if letter.status not in [LetterStatus.DELIVERED, LetterStatus.IN_TRANSIT]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="信件尚未投递，无法阅读"
        )
    
    try:
        # 记录详细的阅读日志
        try:
            await log_letter_read(
                db=db,
                letter_id=letter.id,
                request=request,
                read_duration=read_duration,
                is_complete_read=is_complete
            )
        except Exception as e:
            # 日志记录失败不应该影响主业务
            print(f"Read log recording failed: {e}")
        
        # 增加阅读次数
        letter.read_count += 1
        db.commit()
        
        # 发送WebSocket通知给发送者
        try:
            await notify_letter_read(letter.id, letter.sender_id, letter.read_count)
        except Exception as e:
            # WebSocket通知失败不应该影响主业务
            print(f"WebSocket notification failed: {e}")
        
        # 准备返回数据（隐藏敏感信息）
        letter_data = {
            "id": letter.id,
            "title": letter.title,
            "content": letter.content,
            "sender_nickname": letter.sender_nickname if not letter.anonymous else "匿名用户",
            "receiver_hint": letter.receiver_hint,
            "created_at": letter.created_at.isoformat(),
            "read_count": letter.read_count,
            "anonymous": letter.anonymous
        }
        
        return SuccessResponse(data=letter_data)
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"读取信件失败: {str(e)}"
        )

@router.get("/{letter_id}/stats", response_model=SuccessResponse, summary="获取信件阅读统计")
async def get_letter_read_stats(
    letter_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    获取信件的详细阅读统计
    
    - **letter_id**: 信件编号
    
    只有信件发送者可以查看阅读统计
    """
    # 查询信件
    letter = db.query(Letter).filter(Letter.id == letter_id).first()
    
    if not letter:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"信件 {letter_id} 不存在"
        )
    
    # 权限检查：只有发送者可以查看统计
    if letter.sender_id != current_user:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="无权限查看此信件的阅读统计"
        )
    
    # 使用优化的统计查询
    stats = LetterQueryOptimizer.get_letter_analytics_optimized(db, letter_id)
    
    return SuccessResponse(data=stats)

@router.put("/{letter_id}", response_model=SuccessResponse, summary="编辑信件内容")
async def edit_letter(
    letter_id: str,
    letter_update: LetterUpdate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    编辑信件内容（只有草稿状态的信件可以编辑）
    
    - **letter_id**: 信件编号
    - **title**: 新标题（可选）
    - **content**: 新内容（可选）
    - **receiver_hint**: 新接收者提示（可选）
    - **delivery_instructions**: 新投递说明（可选）
    
    只能编辑处于草稿状态的信件，且只有发送者可以编辑
    """
    # 查询信件
    letter = db.query(Letter).filter(Letter.id == letter_id).first()
    
    if not letter:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"信件 {letter_id} 不存在"
        )
    
    # 权限检查：只有发送者可以编辑
    if letter.sender_id != current_user:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="无权限编辑此信件"
        )
    
    # 状态检查：只能编辑草稿状态的信件
    if letter.status != LetterStatus.DRAFT:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"只能编辑草稿状态的信件，当前状态：{letter.status.value}"
        )
    
    try:
        # 记录原始内容（用于审计）
        original_data = {
            "title": letter.title,
            "content": letter.content,
            "receiver_hint": letter.receiver_hint,
            "delivery_instructions": letter.delivery_instructions
        }
        
        # 更新字段（只更新提供的字段）
        updated_fields = []
        if letter_update.title is not None:
            letter.title = letter_update.title
            updated_fields.append("title")
        
        if letter_update.content is not None:
            letter.content = letter_update.content
            updated_fields.append("content")
        
        if letter_update.receiver_hint is not None:
            letter.receiver_hint = letter_update.receiver_hint
            updated_fields.append("receiver_hint")
        
        if letter_update.delivery_instructions is not None:
            letter.delivery_instructions = letter_update.delivery_instructions
            updated_fields.append("delivery_instructions")
        
        # 如果没有任何更新，返回当前数据
        if not updated_fields:
            return SuccessResponse(
                data={
                    "letter_id": letter.id,
                    "message": "没有字段需要更新",
                    "current_data": letter.to_dict()
                }
            )
        
        # 保存更改
        db.commit()
        db.refresh(letter)
        
        # TODO: 记录编辑历史（可选功能）
        # await log_letter_edit(letter_id, current_user, original_data, updated_fields)
        
        return SuccessResponse(
            data={
                "letter_id": letter.id,
                "updated_fields": updated_fields,
                "updated_data": letter.to_dict(),
                "updated_at": letter.updated_at.isoformat()
            }
        )
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"编辑信件失败: {str(e)}"
        )

@router.delete("/{letter_id}", response_model=SuccessResponse, summary="删除/撤回信件")
async def delete_letter(
    letter_id: str,
    force: bool = Query(False, description="是否强制删除（管理员权限）"),
    current_user_info: dict = Depends(get_current_user_info),
    db: Session = Depends(get_db)
):
    """
    删除或撤回信件
    
    - **letter_id**: 信件编号
    - **force**: 是否强制删除（管理员权限）
    
    删除规则：
    - 草稿状态：可以直接删除
    - 已生成状态：可以撤回（将状态改回草稿并标记撤回）
    - 其他状态：只有管理员可以强制删除
    """
    current_user = current_user_info.get("user_id")
    user_role = current_user_info.get("role", "user")
    
    # 查询信件
    letter = db.query(Letter).filter(Letter.id == letter_id).first()
    
    if not letter:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"信件 {letter_id} 不存在"
        )
    
    # 权限检查：只有发送者或管理员可以删除
    if letter.sender_id != current_user and user_role != "admin":
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="无权限删除此信件"
        )
    
    try:
        # 根据状态和权限决定删除策略
        if letter.status == LetterStatus.DRAFT:
            # 草稿状态：直接删除
            db.delete(letter)
            action = "deleted"
            message = "草稿信件已删除"
            
        elif letter.status == LetterStatus.GENERATED:
            # 已生成状态：可以撤回
            if force and user_role == "admin":
                # 管理员强制删除
                db.delete(letter)
                action = "force_deleted"
                message = "信件已强制删除"
            else:
                # 普通用户撤回：改回草稿状态
                letter.status = LetterStatus.DRAFT
                letter.title = f"[已撤回] {letter.title}"
                action = "recalled"
                message = "信件已撤回，改为草稿状态"
                
        elif letter.status in [LetterStatus.COLLECTED, LetterStatus.IN_TRANSIT]:
            # 已收取或投递中：只有管理员可以强制删除
            if force and user_role == "admin":
                db.delete(letter)
                action = "force_deleted"
                message = "信件已强制删除"
            else:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=f"信件已被收取或正在投递中，无法删除。当前状态：{letter.status.value}"
                )
                
        elif letter.status in [LetterStatus.DELIVERED, LetterStatus.FAILED]:
            # 已投递或失败：只有管理员可以强制删除
            if force and user_role == "admin":
                db.delete(letter)
                action = "force_deleted"
                message = "信件已强制删除"
            else:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=f"信件已投递完成，无法删除。当前状态：{letter.status.value}"
                )
        
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"未知的信件状态：{letter.status.value}"
            )
        
        db.commit()
        
        # 发送WebSocket通知（如果是撤回操作）
        if action == "recalled":
            try:
                await notify_letter_status_update(letter_id, letter.status.value, current_user)
            except Exception as e:
                print(f"WebSocket notification failed: {e}")
        
        return SuccessResponse(
            data={
                "letter_id": letter_id,
                "action": action,
                "message": message,
                "performed_by": current_user,
                "performed_at": datetime.utcnow().isoformat()
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"删除信件失败: {str(e)}"
        )