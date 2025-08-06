from fastapi import APIRouter, Depends, HTTPException, status, Query, Request
from sqlalchemy.orm import Session
from typing import Optional, List
from datetime import datetime
import math

from app.core.database import get_db, get_async_session
from app.models.plaza import PlazaPost, PlazaLike, PlazaComment, PlazaCategory, PostStatus
from app.models.letter import Letter
from app.schemas.plaza import (
    PlazaPostCreate, PlazaPostUpdate, PlazaPostResponse, PlazaPostListItem,
    PlazaPostListResponse, PlazaLikeCreate, PlazaLikeResponse,
    PlazaCommentCreate, PlazaCommentResponse, PlazaCommentListResponse,
    PlazaCategoryResponse, PlazaCategoryListResponse,
    SuccessResponse, ErrorResponse
)
from app.utils.auth import get_current_user, get_current_user_info
from app.utils.user_service import get_user_nickname
from app.utils.plaza_utils import (
    generate_unique_post_id, generate_unique_comment_id, create_excerpt,
    get_popular_tags, get_recommended_posts, update_post_stats,
    validate_post_permissions, PlazaPostFilter
)
from app.utils.websocket_client import notify_plaza_activity

router = APIRouter()

@router.get("/posts", response_model=SuccessResponse, summary="获取广场帖子列表")
async def get_plaza_posts(
    category: Optional[str] = Query(None, description="分类过滤"),
    tags: Optional[str] = Query(None, description="标签过滤（逗号分隔）"),
    author_id: Optional[str] = Query(None, description="作者ID过滤"),
    keyword: Optional[str] = Query(None, description="关键词搜索"),
    sort_by: str = Query("created_at", description="排序字段 (created_at/hot/like_count/view_count)"),
    order: str = Query("desc", description="排序方向 (asc/desc)"),
    page: int = Query(1, ge=1, description="页码"),
    limit: int = Query(20, ge=1, le=100, description="每页数量"),
    db: Session = Depends(get_db)
):
    """
    获取广场帖子列表
    
    支持分类、标签、作者、关键词过滤，以及多种排序方式
    """
    try:
        # 构建基础查询
        query = db.query(PlazaPost)
        
        # 应用过滤条件
        filters = {
            "category": category,
            "tags": tags.split(',') if tags else None,
            "author_id": author_id,
            "keyword": keyword,
            "status": PostStatus.PUBLISHED.value  # 只显示已发布的帖子
        }
        
        query = PlazaPostFilter.apply_filters(query, filters)
        query = PlazaPostFilter.apply_sorting(query, sort_by, order)
        
        # 分页
        total = query.count()
        posts = query.offset((page - 1) * limit).limit(limit).all()
        
        # 转换为响应格式
        post_items = []
        for post in posts:
            post_dict = post.to_dict(include_content=False)
            post_items.append(PlazaPostListItem(**post_dict))
        
        # 分页信息
        pages = math.ceil(total / limit)
        has_next = page < pages
        has_prev = page > 1
        
        response_data = PlazaPostListResponse(
            posts=post_items,
            total=total,
            page=page,
            pages=pages,
            has_next=has_next,
            has_prev=has_prev
        )
        
        return SuccessResponse(data=response_data.dict())
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取帖子列表失败: {str(e)}"
        )

@router.post("/posts", response_model=SuccessResponse, summary="创建广场帖子")
async def create_plaza_post(
    post_data: PlazaPostCreate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    创建新的广场帖子
    
    可以基于现有信件创建，也可以独立创建
    """
    try:
        # 生成唯一ID
        post_id = generate_unique_post_id(db)
        
        # 获取用户昵称
        author_nickname = await get_user_nickname(current_user)
        
        # 创建摘要
        excerpt = post_data.excerpt or create_excerpt(post_data.content)
        
        # 验证关联的信件（如果提供）
        if post_data.letter_id:
            letter = db.query(Letter).filter(
                Letter.id == post_data.letter_id,
                Letter.sender_id == current_user
            ).first()
            if not letter:
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="关联的信件不存在或无权限访问"
                )
        
        # 创建帖子
        new_post = PlazaPost(
            id=post_id,
            title=post_data.title,
            content=post_data.content,
            excerpt=excerpt,
            author_id=current_user,
            author_nickname=author_nickname,
            category=post_data.category.value,
            tags=','.join(post_data.tags) if post_data.tags else None,
            status=PostStatus.PUBLISHED.value,
            allow_comments=post_data.allow_comments,
            anonymous=post_data.anonymous,
            letter_id=post_data.letter_id,
            published_at=datetime.utcnow()
        )
        
        db.add(new_post)
        db.commit()
        db.refresh(new_post)
        
        # 发送WebSocket通知
        try:
            await notify_plaza_activity("post_created", {
                "post_id": new_post.id,
                "title": new_post.title,
                "author": new_post.author_nickname,
                "category": new_post.category
            })
        except Exception as e:
            print(f"WebSocket notification failed: {e}")
        
        return SuccessResponse(
            data={
                "post_id": new_post.id,
                "title": new_post.title,
                "status": new_post.status,
                "created_at": new_post.created_at.isoformat(),
                "published_at": new_post.published_at.isoformat()
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"创建帖子失败: {str(e)}"
        )

@router.get("/posts/{post_id}", response_model=SuccessResponse, summary="获取帖子详情")
async def get_plaza_post(
    post_id: str,
    current_user: Optional[str] = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    获取广场帖子详情
    
    自动增加浏览量，检查用户权限
    """
    try:
        # 查询帖子
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        
        if not post:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"帖子 {post_id} 不存在"
            )
        
        # 检查权限
        user_role = "user"  # 可以从JWT中获取
        permissions = validate_post_permissions(post, current_user or "", user_role)
        
        if not permissions["can_view"]:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="无权限查看此帖子"
            )
        
        # 增加浏览量
        if current_user:
            update_post_stats(db, post_id, "view", 1)
        
        # 转换为响应格式
        post_dict = post.to_dict(include_content=True)
        post_response = PlazaPostResponse(**post_dict)
        
        return SuccessResponse(data=post_response.dict())
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取帖子详情失败: {str(e)}"
        )

@router.put("/posts/{post_id}", response_model=SuccessResponse, summary="更新帖子")
async def update_plaza_post(
    post_id: str,
    post_update: PlazaPostUpdate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    更新广场帖子
    
    只有作者可以编辑自己的帖子
    """
    try:
        # 查询帖子
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        
        if not post:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"帖子 {post_id} 不存在"
            )
        
        # 权限检查
        if post.author_id != current_user:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="无权限编辑此帖子"
            )
        
        # 更新字段
        updated_fields = []
        
        if post_update.title is not None:
            post.title = post_update.title
            updated_fields.append("title")
        
        if post_update.content is not None:
            post.content = post_update.content
            # 重新生成摘要
            post.excerpt = post_update.excerpt or create_excerpt(post_update.content)
            updated_fields.append("content")
        
        if post_update.category is not None:
            post.category = post_update.category.value
            updated_fields.append("category")
        
        if post_update.tags is not None:
            post.tags = ','.join(post_update.tags) if post_update.tags else None
            updated_fields.append("tags")
        
        if post_update.allow_comments is not None:
            post.allow_comments = post_update.allow_comments
            updated_fields.append("allow_comments")
        
        if not updated_fields:
            return SuccessResponse(
                data={
                    "post_id": post.id,
                    "message": "没有字段需要更新",
                    "current_data": post.to_dict()
                }
            )
        
        # 保存更改
        db.commit()
        db.refresh(post)
        
        return SuccessResponse(
            data={
                "post_id": post.id,
                "updated_fields": updated_fields,
                "updated_data": post.to_dict(),
                "updated_at": post.updated_at.isoformat()
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"更新帖子失败: {str(e)}"
        )

@router.delete("/posts/{post_id}", response_model=SuccessResponse, summary="删除帖子")
async def delete_plaza_post(
    post_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    删除广场帖子
    
    只有作者可以删除自己的帖子
    """
    try:
        # 查询帖子
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        
        if not post:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"帖子 {post_id} 不存在"
            )
        
        # 权限检查
        if post.author_id != current_user:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="无权限删除此帖子"
            )
        
        # 删除帖子（级联删除相关的点赞和评论）
        db.delete(post)
        db.commit()
        
        return SuccessResponse(
            data={
                "post_id": post_id,
                "message": "帖子已删除",
                "deleted_at": datetime.utcnow().isoformat()
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"删除帖子失败: {str(e)}"
        )

@router.post("/posts/{post_id}/like", response_model=SuccessResponse, summary="点赞/取消点赞")
async def toggle_post_like(
    post_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    切换帖子点赞状态
    
    如果已点赞则取消，未点赞则添加
    """
    try:
        # 检查帖子是否存在
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        if not post:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"帖子 {post_id} 不存在"
            )
        
        # 检查是否已点赞
        existing_like = db.query(PlazaLike).filter(
            PlazaLike.post_id == post_id,
            PlazaLike.user_id == current_user
        ).first()
        
        if existing_like:
            # 取消点赞
            db.delete(existing_like)
            post.like_count = max(0, post.like_count - 1)
            liked = False
            action = "unliked"
        else:
            # 添加点赞
            new_like = PlazaLike(
                post_id=post_id,
                user_id=current_user
            )
            db.add(new_like)
            post.like_count += 1
            liked = True
            action = "liked"
        
        db.commit()
        
        # 发送WebSocket通知
        try:
            await notify_plaza_activity("post_liked", {
                "post_id": post_id,
                "action": action,
                "like_count": post.like_count,
                "user_id": current_user
            })
        except Exception as e:
            print(f"WebSocket notification failed: {e}")
        
        return SuccessResponse(
            data={
                "post_id": post_id,
                "liked": liked,
                "like_count": post.like_count,
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
            detail=f"点赞操作失败: {str(e)}"
        )

@router.get("/posts/{post_id}/comments", response_model=SuccessResponse, summary="获取帖子评论")
async def get_post_comments(
    post_id: str,
    page: int = Query(1, ge=1, description="页码"),
    limit: int = Query(20, ge=1, le=100, description="每页数量"),
    db: Session = Depends(get_db)
):
    """
    获取帖子评论列表
    
    支持分页，返回层级结构的评论
    """
    try:
        # 检查帖子是否存在
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        if not post:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"帖子 {post_id} 不存在"
            )
        
        # 查询顶级评论（非回复）
        query = db.query(PlazaComment).filter(
            PlazaComment.post_id == post_id,
            PlazaComment.parent_id.is_(None),
            PlazaComment.is_deleted == False
        ).order_by(PlazaComment.created_at.desc())
        
        total = query.count()
        comments = query.offset((page - 1) * limit).limit(limit).all()
        
        # 为每个评论加载回复
        comment_responses = []
        for comment in comments:
            comment_dict = comment.to_dict()
            
            # 加载回复
            replies = db.query(PlazaComment).filter(
                PlazaComment.parent_id == comment.id,
                PlazaComment.is_deleted == False
            ).order_by(PlazaComment.created_at.asc()).all()
            
            comment_dict["replies"] = [reply.to_dict() for reply in replies]
            comment_responses.append(PlazaCommentResponse(**comment_dict))
        
        # 分页信息
        pages = math.ceil(total / limit)
        
        response_data = PlazaCommentListResponse(
            comments=comment_responses,
            total=total,
            page=page,
            pages=pages
        )
        
        return SuccessResponse(data=response_data.dict())
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取评论失败: {str(e)}"
        )

@router.post("/posts/{post_id}/comments", response_model=SuccessResponse, summary="添加评论")
async def create_post_comment(
    post_id: str,
    comment_data: PlazaCommentCreate,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    为帖子添加评论
    
    支持回复其他评论
    """
    try:
        # 检查帖子是否存在且允许评论
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        if not post:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"帖子 {post_id} 不存在"
            )
        
        if not post.allow_comments:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="此帖子不允许评论"
            )
        
        # 检查父评论是否存在（如果是回复）
        if comment_data.parent_id:
            parent_comment = db.query(PlazaComment).filter(
                PlazaComment.id == comment_data.parent_id,
                PlazaComment.post_id == post_id,
                PlazaComment.is_deleted == False
            ).first()
            if not parent_comment:
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="父评论不存在"
                )
        
        # 获取用户昵称
        user_nickname = await get_user_nickname(current_user)
        
        # 创建评论
        comment_id = generate_unique_comment_id(db)
        new_comment = PlazaComment(
            id=comment_id,
            post_id=post_id,
            user_id=current_user,
            user_nickname=user_nickname,
            content=comment_data.content,
            parent_id=comment_data.parent_id,
            reply_to_user=comment_data.reply_to_user
        )
        
        db.add(new_comment)
        
        # 更新帖子评论数
        post.comment_count += 1
        
        db.commit()
        db.refresh(new_comment)
        
        # 发送WebSocket通知
        try:
            await notify_plaza_activity("comment_created", {
                "post_id": post_id,
                "comment_id": new_comment.id,
                "content": new_comment.content[:100] + "..." if len(new_comment.content) > 100 else new_comment.content,
                "user_nickname": new_comment.user_nickname,
                "is_reply": bool(comment_data.parent_id)
            })
        except Exception as e:
            print(f"WebSocket notification failed: {e}")
        
        return SuccessResponse(
            data={
                "comment_id": new_comment.id,
                "post_id": post_id,
                "content": new_comment.content,
                "created_at": new_comment.created_at.isoformat(),
                "comment_count": post.comment_count
            }
        )
        
    except HTTPException:
        db.rollback()
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"添加评论失败: {str(e)}"
        )

@router.get("/categories", response_model=SuccessResponse, summary="获取分类列表")
async def get_plaza_categories(db: Session = Depends(get_db)):
    """
    获取广场分类列表
    
    返回所有启用的分类，按排序顺序
    """
    try:
        categories = db.query(PlazaCategory).filter(
            PlazaCategory.is_active == True
        ).order_by(PlazaCategory.sort_order.asc()).all()
        
        category_responses = [
            PlazaCategoryResponse(**category.to_dict()) 
            for category in categories
        ]
        
        response_data = PlazaCategoryListResponse(categories=category_responses)
        
        return SuccessResponse(data=response_data.dict())
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取分类列表失败: {str(e)}"
        )

@router.get("/tags/popular", response_model=SuccessResponse, summary="获取热门标签")
async def get_popular_tags(
    limit: int = Query(20, ge=1, le=50, description="返回数量"),
    db: Session = Depends(get_db)
):
    """
    获取热门标签列表
    
    基于标签使用频率排序
    """
    try:
        popular_tags = get_popular_tags(db, limit)
        
        return SuccessResponse(
            data={
                "tags": popular_tags,
                "count": len(popular_tags)
            }
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取热门标签失败: {str(e)}"
        )

@router.get("/recommendations", response_model=SuccessResponse, summary="获取推荐帖子")
async def get_recommended_posts_api(
    current_user: str = Depends(get_current_user),
    limit: int = Query(10, ge=1, le=50, description="推荐数量"),
    db: Session = Depends(get_db)
):
    """
    获取个性化推荐帖子
    
    基于用户行为和内容热度推荐
    """
    try:
        recommended_posts = get_recommended_posts(db, current_user, limit)
        
        post_items = []
        for post in recommended_posts:
            post_dict = post.to_dict(include_content=False)
            post_items.append(PlazaPostListItem(**post_dict))
        
        return SuccessResponse(
            data={
                "posts": [item.dict() for item in post_items],
                "count": len(post_items)
            }
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取推荐帖子失败: {str(e)}"
        )