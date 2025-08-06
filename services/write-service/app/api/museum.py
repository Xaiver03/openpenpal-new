from fastapi import APIRouter, Depends, HTTPException, status, Query, Request
from sqlalchemy.orm import Session
from typing import Optional, List
from datetime import datetime
import math
import time

from app.core.database import get_db, get_async_session
from app.models.museum import (
    MuseumLetter, MuseumFavorite, MuseumRating, TimelineEvent, 
    MuseumCollection, CollectionLetter, MuseumLetterStatus, MuseumEra
)
from app.models.letter import Letter
from app.schemas.museum import (
    MuseumLetterCreate, MuseumLetterUpdate, MuseumLetterResponse, MuseumLetterListItem,
    MuseumLetterListResponse, MuseumFavoriteCreate, MuseumFavoriteResponse,
    MuseumRatingCreate, MuseumRatingResponse, TimelineEventCreate, TimelineEventResponse,
    TimelineResponse, MuseumCollectionCreate, MuseumCollectionResponse, MuseumCollectionListResponse,
    MuseumStatsResponse, MuseumSearchRequest, MuseumSearchResponse,
    SuccessResponse, ErrorResponse
)
from app.utils.auth import get_current_user, get_current_user_info
from app.utils.user_service import get_user_nickname
from app.utils.museum_utils import (
    generate_unique_museum_letter_id, generate_unique_timeline_event_id, generate_unique_collection_id,
    create_museum_summary, get_era_by_date, get_popular_museum_tags, get_featured_museum_letters,
    get_recommended_museum_letters, update_museum_letter_stats, update_museum_letter_rating,
    get_museum_statistics, get_timeline_by_date_range, get_timeline_by_era,
    validate_museum_letter_permissions, MuseumLetterFilter
)
from app.utils.websocket_client import notify_plaza_activity

router = APIRouter()

@router.get("/letters", response_model=SuccessResponse, summary="获取博物馆信件列表")
async def get_museum_letters(
    era: Optional[str] = Query(None, description="历史时期过滤"),
    category: Optional[str] = Query(None, description="分类过滤"),
    tags: Optional[str] = Query(None, description="标签过滤（逗号分隔）"),
    author: Optional[str] = Query(None, description="作者过滤"),
    location: Optional[str] = Query(None, description="地点过滤"),
    keyword: Optional[str] = Query(None, description="关键词搜索"),
    featured: Optional[bool] = Query(None, description="只显示精选"),
    sort_by: str = Query("created_at", description="排序字段 (created_at/historical_date/view_count/rating/relevance)"),
    order: str = Query("desc", description="排序方向 (asc/desc)"),
    page: int = Query(1, ge=1, description="页码"),
    limit: int = Query(20, ge=1, le=100, description="每页数量"),
    db: Session = Depends(get_db)
):
    """
    获取博物馆信件列表
    
    支持多种过滤和排序方式，只显示已审核通过的信件
    """
    try:
        start_time = time.time()
        
        # 构建基础查询
        query = db.query(MuseumLetter)
        
        # 应用过滤条件
        filters = {
            "era": era,
            "category": category,
            "tags": tags.split(',') if tags else None,
            "author": author,
            "location": location,
            "keyword": keyword,
            "featured": featured
        }
        
        query = MuseumLetterFilter.apply_filters(query, filters)
        query = MuseumLetterFilter.apply_sorting(query, sort_by, order)
        
        # 分页
        total = query.count()
        letters = query.offset((page - 1) * limit).limit(limit).all()
        
        # 转换为响应格式
        letter_items = []
        for letter in letters:
            letter_dict = letter.to_dict(include_content=False)
            letter_items.append(MuseumLetterListItem(**letter_dict))
        
        # 分页信息
        pages = math.ceil(total / limit)
        has_next = page < pages
        has_prev = page > 1
        
        response_data = MuseumLetterListResponse(
            letters=letter_items,
            total=total,
            page=page,
            pages=pages,
            has_next=has_next,
            has_prev=has_prev
        )
        
        search_time = time.time() - start_time
        
        return SuccessResponse(data={
            **response_data.dict(),
            "search_time": round(search_time, 3)
        })
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取博物馆信件列表失败: {str(e)}"
        )

@router.post("/letters", response_model=SuccessResponse, summary="贡献博物馆信件")
async def create_museum_letter(
    letter_data: MuseumLetterCreate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    贡献新的博物馆信件
    
    用户可以贡献历史信件或将现代信件转换为博物馆展品
    """
    try:
        # 生成唯一ID
        letter_id = generate_unique_museum_letter_id(db)
        
        # 获取用户昵称
        contributor_name = await get_user_nickname(current_user)
        
        # 创建摘要
        summary = letter_data.summary or create_museum_summary(letter_data.content)
        
        # 自动判断历史时期（如果没有提供）
        era = letter_data.era
        if letter_data.historical_date and not letter_data.era:
            era = get_era_by_date(letter_data.historical_date)
        
        # 验证关联的现代信件（如果提供）
        if letter_data.letter_id:
            letter = db.query(Letter).filter(
                Letter.id == letter_data.letter_id,
                Letter.sender_id == current_user
            ).first()
            if not letter:
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="关联的信件不存在或无权限访问"
                )
        
        # 创建博物馆信件
        new_letter = MuseumLetter(
            id=letter_id,
            title=letter_data.title,
            content=letter_data.content,
            summary=summary,
            original_author=letter_data.original_author,
            original_recipient=letter_data.original_recipient,
            historical_date=letter_data.historical_date,
            era=era.value,
            location=letter_data.location,
            category=letter_data.category,
            tags=','.join(letter_data.tags) if letter_data.tags else None,
            language=letter_data.language,
            source_type=letter_data.source_type.value,
            source_description=letter_data.source_description,
            contributor_id=current_user,
            contributor_name=contributor_name,
            letter_id=letter_data.letter_id,
            status=MuseumLetterStatus.PENDING.value
        )
        
        db.add(new_letter)
        db.commit()
        db.refresh(new_letter)
        
        # 发送WebSocket通知
        try:
            await notify_plaza_activity("museum_contribution", {
                "letter_id": new_letter.id,
                "title": new_letter.title,
                "contributor": new_letter.contributor_name,
                "era": new_letter.era
            })
        except Exception as e:
            print(f"WebSocket notification failed: {e}")
        
        return SuccessResponse(
            data={
                "letter_id": new_letter.id,
                "title": new_letter.title,
                "status": new_letter.status,
                "era": new_letter.era,
                "created_at": new_letter.created_at.isoformat(),
                "message": "信件已提交，等待审核"
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"贡献博物馆信件失败: {str(e)}"
        )

@router.get("/letters/{letter_id}", response_model=SuccessResponse, summary="获取博物馆信件详情")
async def get_museum_letter(
    letter_id: str,
    current_user: Optional[str] = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    获取博物馆信件详情
    
    自动增加浏览量，检查用户权限
    """
    try:
        # 查询信件
        letter = db.query(MuseumLetter).filter(MuseumLetter.id == letter_id).first()
        
        if not letter:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"博物馆信件 {letter_id} 不存在"
            )
        
        # 检查权限
        user_role = "user"  # 可以从JWT中获取
        permissions = validate_museum_letter_permissions(letter, current_user or "", user_role)
        
        if not permissions["can_view"]:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="无权限查看此信件"
            )
        
        # 增加浏览量
        if current_user:
            update_museum_letter_stats(db, letter_id, "view", 1)
        
        # 转换为响应格式
        letter_dict = letter.to_dict(include_content=True)
        letter_response = MuseumLetterResponse(**letter_dict)
        
        return SuccessResponse(data=letter_response.dict())
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取博物馆信件详情失败: {str(e)}"
        )

@router.put("/letters/{letter_id}", response_model=SuccessResponse, summary="更新博物馆信件")
async def update_museum_letter(
    letter_id: str,
    letter_update: MuseumLetterUpdate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    更新博物馆信件
    
    只有贡献者可以编辑待审核的信件，管理员可以编辑所有信件
    """
    try:
        # 查询信件
        letter = db.query(MuseumLetter).filter(MuseumLetter.id == letter_id).first()
        
        if not letter:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"博物馆信件 {letter_id} 不存在"
            )
        
        # 权限检查
        user_role = "user"  # 可以从JWT中获取
        permissions = validate_museum_letter_permissions(letter, current_user, user_role)
        
        if not permissions["can_edit"]:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="无权限编辑此信件"
            )
        
        # 更新字段
        updated_fields = []
        
        if letter_update.title is not None:
            letter.title = letter_update.title
            updated_fields.append("title")
        
        if letter_update.content is not None:
            letter.content = letter_update.content
            # 重新生成摘要
            letter.summary = letter_update.summary or create_museum_summary(letter_update.content)
            updated_fields.append("content")
        
        if letter_update.original_author is not None:
            letter.original_author = letter_update.original_author
            updated_fields.append("original_author")
        
        if letter_update.original_recipient is not None:
            letter.original_recipient = letter_update.original_recipient
            updated_fields.append("original_recipient")
        
        if letter_update.historical_date is not None:
            letter.historical_date = letter_update.historical_date
            # 重新判断历史时期
            if not letter_update.era:
                letter.era = get_era_by_date(letter_update.historical_date).value
            updated_fields.append("historical_date")
        
        if letter_update.era is not None:
            letter.era = letter_update.era.value
            updated_fields.append("era")
        
        if letter_update.location is not None:
            letter.location = letter_update.location
            updated_fields.append("location")
        
        if letter_update.category is not None:
            letter.category = letter_update.category
            updated_fields.append("category")
        
        if letter_update.tags is not None:
            letter.tags = ','.join(letter_update.tags) if letter_update.tags else None
            updated_fields.append("tags")
        
        if letter_update.source_description is not None:
            letter.source_description = letter_update.source_description
            updated_fields.append("source_description")
        
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
        
        return SuccessResponse(
            data={
                "letter_id": letter.id,
                "updated_fields": updated_fields,
                "updated_data": letter.to_dict(),
                "updated_at": letter.updated_at.isoformat()
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"更新博物馆信件失败: {str(e)}"
        )

@router.post("/letters/{letter_id}/favorite", response_model=SuccessResponse, summary="收藏/取消收藏")
async def toggle_museum_letter_favorite(
    letter_id: str,
    favorite_data: MuseumFavoriteCreate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    切换博物馆信件收藏状态
    
    如果已收藏则取消，未收藏则添加
    """
    try:
        # 检查信件是否存在
        letter = db.query(MuseumLetter).filter(MuseumLetter.id == letter_id).first()
        if not letter:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"博物馆信件 {letter_id} 不存在"
            )
        
        # 检查是否已收藏
        existing_favorite = db.query(MuseumFavorite).filter(
            MuseumFavorite.museum_letter_id == letter_id,
            MuseumFavorite.user_id == current_user
        ).first()
        
        if existing_favorite:
            # 取消收藏
            db.delete(existing_favorite)
            letter.favorite_count = max(0, letter.favorite_count - 1)
            favorited = False
            action = "unfavorited"
        else:
            # 添加收藏
            new_favorite = MuseumFavorite(
                museum_letter_id=letter_id,
                user_id=current_user,
                note=favorite_data.note
            )
            db.add(new_favorite)
            letter.favorite_count += 1
            favorited = True
            action = "favorited"
        
        db.commit()
        
        # 发送WebSocket通知
        try:
            await notify_plaza_activity("museum_favorite", {
                "letter_id": letter_id,
                "action": action,
                "favorite_count": letter.favorite_count,
                "user_id": current_user
            })
        except Exception as e:
            print(f"WebSocket notification failed: {e}")
        
        return SuccessResponse(
            data={
                "letter_id": letter_id,
                "favorited": favorited,
                "favorite_count": letter.favorite_count,
                "action": action
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"收藏操作失败: {str(e)}"
        )

@router.post("/letters/{letter_id}/rating", response_model=SuccessResponse, summary="评分博物馆信件")
async def rate_museum_letter(
    letter_id: str,
    rating_data: MuseumRatingCreate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    为博物馆信件评分
    
    支持1-5星评分和评价评论
    """
    try:
        # 检查信件是否存在
        letter = db.query(MuseumLetter).filter(MuseumLetter.id == letter_id).first()
        if not letter:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"博物馆信件 {letter_id} 不存在"
            )
        
        # 检查是否已评分
        existing_rating = db.query(MuseumRating).filter(
            MuseumRating.museum_letter_id == letter_id,
            MuseumRating.user_id == current_user
        ).first()
        
        old_rating = None
        if existing_rating:
            # 更新评分
            old_rating = existing_rating.rating
            existing_rating.rating = rating_data.rating
            existing_rating.comment = rating_data.comment
            db.commit()
            
            # 更新信件平均评分
            update_museum_letter_rating(db, letter_id, rating_data.rating, old_rating)
            
            action = "updated"
            rating_record = existing_rating
        else:
            # 新评分
            new_rating = MuseumRating(
                museum_letter_id=letter_id,
                user_id=current_user,
                rating=rating_data.rating,
                comment=rating_data.comment
            )
            db.add(new_rating)
            db.commit()
            db.refresh(new_rating)
            
            # 更新信件平均评分
            update_museum_letter_rating(db, letter_id, rating_data.rating)
            
            action = "created"
            rating_record = new_rating
        
        # 获取更新后的信件数据
        db.refresh(letter)
        
        return SuccessResponse(
            data={
                "letter_id": letter_id,
                "rating": rating_record.rating,
                "comment": rating_record.comment,
                "action": action,
                "letter_rating_avg": letter.rating_avg,
                "letter_rating_count": letter.rating_count,
                "created_at": rating_record.created_at.isoformat(),
                "updated_at": rating_record.updated_at.isoformat()
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"评分操作失败: {str(e)}"
        )

@router.get("/timeline", response_model=SuccessResponse, summary="获取历史时间线")
async def get_museum_timeline(
    era: Optional[str] = Query(None, description="历史时期过滤"),
    start_date: Optional[datetime] = Query(None, description="开始日期"),
    end_date: Optional[datetime] = Query(None, description="结束日期"),
    event_type: Optional[str] = Query(None, description="事件类型过滤"),
    limit: int = Query(100, ge=1, le=500, description="事件数量"),
    db: Session = Depends(get_db)
):
    """
    获取历史时间线
    
    支持按时期、日期范围、事件类型过滤
    """
    try:
        events = []
        
        if era:
            # 按历史时期查询
            era_enum = MuseumEra(era)
            events = get_timeline_by_era(db, era_enum, limit)
        elif start_date and end_date:
            # 按日期范围查询
            events = get_timeline_by_date_range(db, start_date, end_date, limit)
        else:
            # 查询所有事件
            query = db.query(TimelineEvent)
            
            if event_type:
                query = query.filter(TimelineEvent.event_type == event_type)
            
            events = query.order_by(
                TimelineEvent.event_date.desc(),
                TimelineEvent.importance.desc()
            ).limit(limit).all()
        
        # 转换为响应格式
        event_responses = [
            TimelineEventResponse(**event.to_dict()) 
            for event in events
        ]
        
        # 计算日期范围
        if events:
            dates = [event.event_date for event in events if event.event_date]
            date_range = {
                "start_date": min(dates).isoformat() if dates else None,
                "end_date": max(dates).isoformat() if dates else None
            }
        else:
            date_range = {"start_date": None, "end_date": None}
        
        # 获取涉及的历史时期
        eras = list(set(event.era for event in events))
        
        response_data = TimelineResponse(
            events=event_responses,
            total=len(event_responses),
            date_range=date_range,
            eras=eras
        )
        
        return SuccessResponse(data=response_data.dict())
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取时间线失败: {str(e)}"
        )

@router.get("/collections", response_model=SuccessResponse, summary="获取收藏集列表")
async def get_museum_collections(
    theme: Optional[str] = Query(None, description="主题过滤"),
    creator_id: Optional[str] = Query(None, description="创建者过滤"),
    featured: Optional[bool] = Query(None, description="只显示精选"),
    page: int = Query(1, ge=1, description="页码"),
    limit: int = Query(20, ge=1, le=100, description="每页数量"),
    db: Session = Depends(get_db)
):
    """
    获取博物馆收藏集列表
    
    只显示公开的收藏集
    """
    try:
        # 构建查询
        query = db.query(MuseumCollection).filter(MuseumCollection.is_public == True)
        
        if theme:
            query = query.filter(MuseumCollection.theme == theme)
        
        if creator_id:
            query = query.filter(MuseumCollection.creator_id == creator_id)
        
        if featured:
            query = query.filter(MuseumCollection.is_featured == True)
        
        # 排序
        query = query.order_by(
            MuseumCollection.is_featured.desc(),
            MuseumCollection.view_count.desc(),
            MuseumCollection.created_at.desc()
        )
        
        # 分页
        total = query.count()
        collections = query.offset((page - 1) * limit).limit(limit).all()
        
        # 转换为响应格式
        collection_responses = [
            MuseumCollectionResponse(**collection.to_dict()) 
            for collection in collections
        ]
        
        pages = math.ceil(total / limit)
        
        response_data = MuseumCollectionListResponse(
            collections=collection_responses,
            total=total,
            page=page,
            pages=pages
        )
        
        return SuccessResponse(data=response_data.dict())
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取收藏集列表失败: {str(e)}"
        )

@router.get("/stats", response_model=SuccessResponse, summary="获取博物馆统计数据")
async def get_museum_stats(db: Session = Depends(get_db)):
    """
    获取博物馆统计数据
    
    包括信件数量、分类分布、热门标签等
    """
    try:
        stats = get_museum_statistics(db)
        
        # 获取精选信件
        featured_letters = get_featured_museum_letters(db, 5)
        featured_letter_items = [
            MuseumLetterListItem(**letter.to_dict(include_content=False))
            for letter in featured_letters
        ]
        
        # 获取最近贡献
        recent_letters = db.query(MuseumLetter).filter(
            MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value])
        ).order_by(MuseumLetter.created_at.desc()).limit(5).all()
        
        recent_letter_items = [
            MuseumLetterListItem(**letter.to_dict(include_content=False))
            for letter in recent_letters
        ]
        
        stats_response = MuseumStatsResponse(
            total_letters=stats.get("total_letters", 0),
            total_collections=stats.get("total_collections", 0),
            total_timeline_events=stats.get("total_timeline_events", 0),
            era_distribution=stats.get("era_distribution", {}),
            category_distribution=stats.get("category_distribution", {}),
            popular_tags=stats.get("popular_tags", []),
            featured_letters=featured_letter_items,
            recent_contributions=recent_letter_items
        )
        
        return SuccessResponse(data=stats_response.dict())
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取博物馆统计数据失败: {str(e)}"
        )

@router.get("/recommendations", response_model=SuccessResponse, summary="获取推荐信件")
async def get_museum_recommendations(
    current_user: str = Depends(get_current_user),
    limit: int = Query(10, ge=1, le=50, description="推荐数量"),
    db: Session = Depends(get_db)
):
    """
    获取个性化推荐博物馆信件
    
    基于用户收藏历史和内容热度推荐
    """
    try:
        recommended_letters = get_recommended_museum_letters(db, current_user, limit)
        
        letter_items = []
        for letter in recommended_letters:
            letter_dict = letter.to_dict(include_content=False)
            letter_items.append(MuseumLetterListItem(**letter_dict))
        
        return SuccessResponse(
            data={
                "letters": [item.dict() for item in letter_items],
                "count": len(letter_items)
            }
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取推荐信件失败: {str(e)}"
        )

@router.get("/search", response_model=SuccessResponse, summary="高级搜索博物馆信件")
async def search_museum_letters(
    request: MuseumSearchRequest = Depends(),
    db: Session = Depends(get_db)
):
    """
    高级搜索博物馆信件
    
    支持复合条件搜索和多种排序方式
    """
    try:
        start_time = time.time()
        
        # 构建查询
        query = db.query(MuseumLetter)
        
        # 应用搜索过滤条件
        filters = {
            "keyword": request.keyword,
            "era": request.era.value if request.era else None,
            "category": request.category,
            "tags": request.tags,
            "author": request.author,
            "location": request.location,
            "date_from": request.date_from,
            "date_to": request.date_to
        }
        
        query = MuseumLetterFilter.apply_filters(query, filters)
        query = MuseumLetterFilter.apply_sorting(query, request.sort_by, request.order)
        
        # 分页
        total = query.count()
        letters = query.offset((request.page - 1) * request.limit).limit(request.limit).all()
        
        # 转换为响应格式
        letter_items = []
        for letter in letters:
            letter_dict = letter.to_dict(include_content=False)
            letter_items.append(MuseumLetterListItem(**letter_dict))
        
        # 计算搜索时间
        search_time = time.time() - start_time
        
        # 统计应用的过滤条件
        applied_filters = {k: v for k, v in filters.items() if v is not None}
        
        pages = math.ceil(total / request.limit)
        
        response_data = MuseumSearchResponse(
            letters=letter_items,
            total=total,
            page=request.page,
            pages=pages,
            search_time=round(search_time, 3),
            filters_applied=applied_filters
        )
        
        return SuccessResponse(data=response_data.dict())
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"搜索博物馆信件失败: {str(e)}"
        )