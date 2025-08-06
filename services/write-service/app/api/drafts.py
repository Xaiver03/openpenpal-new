"""
草稿管理API
"""
from fastapi import APIRouter, Depends, HTTPException, status, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime, timedelta

from app.core.database import get_db, get_async_session
from app.utils.auth import get_current_user, get_current_user_optional
from app.models.draft import LetterDraft, DraftHistory
from app.schemas.draft import (
    DraftCreate, DraftUpdate, DraftResponse, DraftListItem,
    DraftHistoryResponse, AutoSaveRequest, DraftStats,
    DraftBatchOperation, DraftSearchRequest, SuccessResponse
)
from app.utils.id_generator import generate_id
from app.utils.draft_utils import DraftManager, auto_save_manager, DraftCleanupService
from app.middleware.error_handler import InputSanitizer
from app.utils.security_utils import secure_content_processing

router = APIRouter()


@router.post("/drafts", response_model=SuccessResponse)
async def create_draft(
    draft_data: DraftCreate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """创建新草稿"""
    try:
        # 内容安全处理
        if draft_data.content:
            draft_data.content = secure_content_processing(draft_data.content, "text")
        if draft_data.title:
            draft_data.title = secure_content_processing(draft_data.title, "text")
        
        # 计算内容统计
        word_count, character_count = DraftManager.calculate_content_stats(draft_data.content or "")
        
        # 创建草稿
        draft = LetterDraft(
            id=generate_id(),
            user_id=current_user,
            title=draft_data.title,
            content=draft_data.content,
            recipient_id=draft_data.recipient_id,
            recipient_type=draft_data.recipient_type,
            paper_style=draft_data.paper_style,
            envelope_style=draft_data.envelope_style,
            draft_type=draft_data.draft_type,
            parent_letter_id=draft_data.parent_letter_id,
            auto_save_enabled=draft_data.auto_save_enabled,
            word_count=word_count,
            character_count=character_count
        )
        
        db.add(draft)
        db.commit()
        db.refresh(draft)
        
        return SuccessResponse(
            msg="草稿创建成功",
            data={"draft_id": draft.id}
        )
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"创建草稿失败: {str(e)}"
        )


@router.get("/drafts", response_model=SuccessResponse)
async def get_user_drafts(
    skip: int = 0,
    limit: int = 20,
    is_active: Optional[bool] = None,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取用户草稿列表"""
    query = db.query(LetterDraft).filter(LetterDraft.user_id == current_user)
    
    if is_active is not None:
        query = query.filter(LetterDraft.is_active == is_active)
    
    # 按最后编辑时间倒序
    drafts = query.order_by(LetterDraft.last_edit_time.desc()).offset(skip).limit(limit).all()
    
    # 转换为列表项格式
    draft_items = [
        DraftListItem(
            id=draft.id,
            title=draft.title,
            draft_type=draft.draft_type,
            recipient_id=draft.recipient_id,
            version=draft.version,
            word_count=draft.word_count,
            last_edit_time=draft.last_edit_time,
            created_at=draft.created_at,
            is_active=draft.is_active
        )
        for draft in drafts
    ]
    
    return SuccessResponse(
        msg="获取草稿列表成功",
        data={
            "drafts": [item.model_dump() for item in draft_items],
            "total": len(draft_items)
        }
    )


@router.get("/drafts/{draft_id}", response_model=SuccessResponse)
async def get_draft_detail(
    draft_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取草稿详情"""
    draft = db.query(LetterDraft).filter(
        LetterDraft.id == draft_id,
        LetterDraft.user_id == current_user
    ).first()
    
    if not draft:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="草稿不存在"
        )
    
    return SuccessResponse(
        msg="获取草稿详情成功",
        data=DraftResponse.model_validate(draft).model_dump()
    )


@router.put("/drafts/{draft_id}", response_model=SuccessResponse)
async def update_draft(
    draft_id: str,
    draft_data: DraftUpdate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """更新草稿"""
    draft = db.query(LetterDraft).filter(
        LetterDraft.id == draft_id,
        LetterDraft.user_id == current_user
    ).first()
    
    if not draft:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="草稿不存在"
        )
    
    try:
        # 保存旧内容用于变化检测
        old_content = draft.content
        
        # 内容安全处理
        if draft_data.content is not None:
            draft_data.content = secure_content_processing(draft_data.content, "text")
        if draft_data.title is not None:
            draft_data.title = secure_content_processing(draft_data.title, "text")
        
        # 检查是否需要创建历史备份
        if DraftManager.should_create_history_backup(draft, draft_data.content or ""):
            history = DraftHistory(
                id=generate_id(),
                draft_id=draft_id,
                user_id=current_user,
                title=draft.title,
                content=draft.content,
                version=draft.version,
                change_summary=DraftManager.generate_change_summary(old_content, draft_data.content or ""),
                change_type="version_backup",
                word_count=draft.word_count,
                character_count=draft.character_count
            )
            db.add(history)
        
        # 更新草稿字段
        update_data = draft_data.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(draft, field, value)
        
        # 重新计算统计信息
        if draft_data.content is not None:
            word_count, character_count = DraftManager.calculate_content_stats(draft_data.content)
            draft.word_count = word_count
            draft.character_count = character_count
        
        # 更新版本和时间
        draft.version += 1
        draft.last_edit_time = datetime.utcnow()
        
        db.commit()
        db.refresh(draft)
        
        return SuccessResponse(
            msg="草稿更新成功",
            data={"version": draft.version}
        )
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"更新草稿失败: {str(e)}"
        )


@router.post("/drafts/{draft_id}/auto-save", response_model=SuccessResponse)
async def auto_save_draft(
    draft_id: str,
    save_data: AutoSaveRequest,
    background_tasks: BackgroundTasks,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """自动保存草稿"""
    draft = db.query(LetterDraft).filter(
        LetterDraft.id == draft_id,
        LetterDraft.user_id == current_user
    ).first()
    
    if not draft:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="草稿不存在"
        )
    
    if not draft.auto_save_enabled:
        return SuccessResponse(
            msg="自动保存已禁用",
            data={"saved": False}
        )
    
    try:
        # 内容安全处理
        if save_data.content is not None:
            save_data.content = secure_content_processing(save_data.content, "text")
        if save_data.title is not None:
            save_data.title = secure_content_processing(save_data.title, "text")
        
        # 检查是否有实际变化
        content_changed = save_data.content is not None and save_data.content != draft.content
        title_changed = save_data.title is not None and save_data.title != draft.title
        
        if not content_changed and not title_changed:
            return SuccessResponse(
                msg="内容无变化，跳过保存",
                data={"saved": False}
            )
        
        # 创建自动保存历史记录
        if content_changed:
            history = DraftHistory(
                id=generate_id(),
                draft_id=draft_id,
                user_id=current_user,
                title=draft.title,
                content=draft.content,
                version=draft.version,
                change_summary="自动保存",
                change_type="auto_save",
                word_count=draft.word_count,
                character_count=draft.character_count
            )
            db.add(history)
        
        # 更新草稿
        if save_data.content is not None:
            draft.content = save_data.content
            word_count, character_count = DraftManager.calculate_content_stats(save_data.content)
            draft.word_count = word_count
            draft.character_count = character_count
        
        if save_data.title is not None:
            draft.title = save_data.title
        
        draft.last_edit_time = datetime.utcnow()
        
        db.commit()
        
        # 安排下次自动保存
        if draft.auto_save_enabled:
            interval = auto_save_manager.get_save_interval(current_user)
            auto_save_manager.schedule_auto_save(
                current_user, 
                draft_id, 
                draft.content or "", 
                draft.title,
                interval
            )
        
        return SuccessResponse(
            msg="自动保存成功",
            data={
                "saved": True,
                "version": draft.version,
                "word_count": draft.word_count,
                "save_time": draft.last_edit_time.isoformat()
            }
        )
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"自动保存失败: {str(e)}"
        )


@router.get("/drafts/{draft_id}/history", response_model=SuccessResponse)
async def get_draft_history(
    draft_id: str,
    limit: int = 10,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取草稿历史记录"""
    # 验证草稿所有权
    draft = db.query(LetterDraft).filter(
        LetterDraft.id == draft_id,
        LetterDraft.user_id == current_user
    ).first()
    
    if not draft:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="草稿不存在"
        )
    
    # 获取历史记录
    histories = db.query(DraftHistory).filter(
        DraftHistory.draft_id == draft_id
    ).order_by(DraftHistory.version.desc()).limit(limit).all()
    
    history_items = [
        DraftHistoryResponse.model_validate(history).model_dump()
        for history in histories
    ]
    
    return SuccessResponse(
        msg="获取历史记录成功",
        data={
            "histories": history_items,
            "total": len(history_items)
        }
    )


@router.post("/drafts/{draft_id}/restore/{version}", response_model=SuccessResponse)
async def restore_draft_version(
    draft_id: str,
    version: int,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """恢复草稿到指定版本"""
    # 验证草稿所有权
    draft = db.query(LetterDraft).filter(
        LetterDraft.id == draft_id,
        LetterDraft.user_id == current_user
    ).first()
    
    if not draft:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="草稿不存在"
        )
    
    # 查找历史版本
    history = db.query(DraftHistory).filter(
        DraftHistory.draft_id == draft_id,
        DraftHistory.version == version
    ).first()
    
    if not history:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="历史版本不存在"
        )
    
    try:
        # 备份当前版本
        current_backup = DraftHistory(
            id=generate_id(),
            draft_id=draft_id,
            user_id=current_user,
            title=draft.title,
            content=draft.content,
            version=draft.version,
            change_summary=f"恢复前备份（版本{draft.version}）",
            change_type="version_backup",
            word_count=draft.word_count,
            character_count=draft.character_count
        )
        db.add(current_backup)
        
        # 恢复到历史版本
        draft.title = history.title
        draft.content = history.content
        draft.word_count = history.word_count
        draft.character_count = history.character_count
        draft.version += 1
        draft.last_edit_time = datetime.utcnow()
        
        db.commit()
        
        return SuccessResponse(
            msg=f"成功恢复到版本 {version}",
            data={
                "new_version": draft.version,
                "restored_from": version
            }
        )
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"恢复版本失败: {str(e)}"
        )


@router.delete("/drafts/{draft_id}", response_model=SuccessResponse)
async def delete_draft(
    draft_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """删除草稿"""
    draft = db.query(LetterDraft).filter(
        LetterDraft.id == draft_id,
        LetterDraft.user_id == current_user
    ).first()
    
    if not draft:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="草稿不存在"
        )
    
    try:
        # 取消自动保存任务
        auto_save_manager.cancel_auto_save(current_user, draft_id)
        
        # 删除历史记录
        db.query(DraftHistory).filter(DraftHistory.draft_id == draft_id).delete()
        
        # 删除草稿
        db.delete(draft)
        db.commit()
        
        return SuccessResponse(msg="草稿删除成功")
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"删除草稿失败: {str(e)}"
        )


@router.post("/drafts/{draft_id}/discard", response_model=SuccessResponse)
async def discard_draft(
    draft_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """丢弃草稿（软删除）"""
    draft = db.query(LetterDraft).filter(
        LetterDraft.id == draft_id,
        LetterDraft.user_id == current_user
    ).first()
    
    if not draft:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="草稿不存在"
        )
    
    try:
        # 取消自动保存
        auto_save_manager.cancel_auto_save(current_user, draft_id)
        
        # 标记为已丢弃
        draft.is_discarded = True
        draft.is_active = False
        
        db.commit()
        
        return SuccessResponse(msg="草稿已丢弃")
        
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"丢弃草稿失败: {str(e)}"
        )


@router.get("/drafts/stats", response_model=SuccessResponse)
async def get_draft_stats(
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取草稿统计信息"""
    # 基础统计
    total_drafts = db.query(LetterDraft).filter(LetterDraft.user_id == current_user).count()
    active_drafts = db.query(LetterDraft).filter(
        LetterDraft.user_id == current_user,
        LetterDraft.is_active == True
    ).count()
    discarded_drafts = db.query(LetterDraft).filter(
        LetterDraft.user_id == current_user,
        LetterDraft.is_discarded == True
    ).count()
    
    # 内容统计
    from sqlalchemy import func
    word_stats = db.query(
        func.sum(LetterDraft.word_count).label('total_words'),
        func.sum(LetterDraft.character_count).label('total_characters')
    ).filter(
        LetterDraft.user_id == current_user,
        LetterDraft.is_active == True
    ).first()
    
    # 时间统计
    time_stats = db.query(
        func.min(LetterDraft.created_at).label('oldest'),
        func.max(LetterDraft.created_at).label('newest')
    ).filter(LetterDraft.user_id == current_user).first()
    
    stats = DraftStats(
        total_drafts=total_drafts,
        active_drafts=active_drafts,
        discarded_drafts=discarded_drafts,
        total_words=word_stats.total_words or 0,
        total_characters=word_stats.total_characters or 0,
        oldest_draft=time_stats.oldest,
        newest_draft=time_stats.newest
    )
    
    return SuccessResponse(
        msg="获取统计信息成功",
        data=stats.model_dump()
    )