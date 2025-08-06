import random
import string
from typing import Optional, List, Dict, Any
from sqlalchemy.orm import Session
from sqlalchemy import desc, and_, or_, func
from datetime import datetime, timedelta

from app.models.plaza import PlazaPost, PlazaLike, PlazaComment, PlazaCategory, PostStatus
from app.models.letter import Letter
from app.utils.cache_manager import LetterCacheService

def generate_post_id() -> str:
    """生成广场帖子ID: PZ + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    # 避免混淆字符
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"PZ{random_part}"

def generate_comment_id() -> str:
    """生成评论ID: CM + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    # 避免混淆字符
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"CM{random_part}"

def generate_unique_post_id(db: Session) -> str:
    """生成唯一的帖子ID"""
    max_attempts = 10
    for _ in range(max_attempts):
        post_id = generate_post_id()
        existing = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        if not existing:
            return post_id
    
    # 如果10次都冲突，使用时间戳确保唯一性
    timestamp = int(datetime.now().timestamp() * 1000)
    return f"PZ{timestamp % 10000000000:010d}"

def generate_unique_comment_id(db: Session) -> str:
    """生成唯一的评论ID"""
    max_attempts = 10
    for _ in range(max_attempts):
        comment_id = generate_comment_id()
        existing = db.query(PlazaComment).filter(PlazaComment.id == comment_id).first()
        if not existing:
            return comment_id
    
    # 如果10次都冲突，使用时间戳确保唯一性
    timestamp = int(datetime.now().timestamp() * 1000)
    return f"CM{timestamp % 10000000000:010d}"

def create_excerpt(content: str, max_length: int = 200) -> str:
    """从内容中提取摘要"""
    if not content:
        return ""
    
    # 移除多余的空白字符
    clean_content = ' '.join(content.split())
    
    if len(clean_content) <= max_length:
        return clean_content
    
    # 在合适的位置截断
    excerpt = clean_content[:max_length]
    
    # 尝试在句号或换行处截断
    for delimiter in ['。', '！', '？', '\n', '.', '!', '?']:
        pos = excerpt.rfind(delimiter)
        if pos > max_length * 0.7:  # 至少保留70%的长度
            return excerpt[:pos + 1]
    
    # 尝试在空格处截断
    space_pos = excerpt.rfind(' ')
    if space_pos > max_length * 0.8:
        return excerpt[:space_pos] + "..."
    
    return excerpt + "..."

def get_popular_tags(db: Session, limit: int = 20) -> List[str]:
    """获取热门标签"""
    try:
        # 从缓存获取
        cache_key = "popular_tags"
        cached_tags = None  # await LetterCacheService.get_cached_data(cache_key)
        
        if cached_tags:
            return cached_tags
        
        # 查询所有标签
        posts = db.query(PlazaPost.tags).filter(
            PlazaPost.status == PostStatus.PUBLISHED.value,
            PlazaPost.tags.isnot(None),
            PlazaPost.tags != ""
        ).all()
        
        # 统计标签频率
        tag_count = {}
        for post in posts:
            if post.tags:
                tags = [tag.strip() for tag in post.tags.split(',') if tag.strip()]
                for tag in tags:
                    tag_count[tag] = tag_count.get(tag, 0) + 1
        
        # 按频率排序
        popular_tags = sorted(tag_count.items(), key=lambda x: x[1], reverse=True)
        result = [tag for tag, count in popular_tags[:limit]]
        
        # 缓存结果（缓存1小时）
        # await LetterCacheService.cache_data(cache_key, result, expire=3600)
        
        return result
        
    except Exception as e:
        print(f"获取热门标签失败: {e}")
        return []

def get_recommended_posts(db: Session, user_id: str, limit: int = 10) -> List[PlazaPost]:
    """获取推荐帖子"""
    try:
        # 简单的推荐算法：基于用户的点赞历史和热门帖子
        
        # 获取用户点赞过的帖子分类
        user_liked_categories = db.query(PlazaPost.category).join(
            PlazaLike, PlazaPost.id == PlazaLike.post_id
        ).filter(
            PlazaLike.user_id == user_id,
            PlazaPost.status == PostStatus.PUBLISHED.value
        ).distinct().all()
        
        liked_categories = [cat[0] for cat in user_liked_categories]
        
        # 构建推荐查询
        query = db.query(PlazaPost).filter(
            PlazaPost.status == PostStatus.PUBLISHED.value,
            PlazaPost.author_id != user_id  # 不推荐自己的帖子
        )
        
        # 如果有喜欢的分类，优先推荐相同分类
        if liked_categories:
            query = query.filter(PlazaPost.category.in_(liked_categories))
        
        # 按热度排序（点赞数 + 评论数 + 浏览数的加权）
        posts = query.order_by(
            desc(PlazaPost.like_count * 3 + PlazaPost.comment_count * 2 + PlazaPost.view_count)
        ).limit(limit).all()
        
        # 如果没有足够的推荐结果，补充热门帖子
        if len(posts) < limit:
            additional_posts = db.query(PlazaPost).filter(
                PlazaPost.status == PostStatus.PUBLISHED.value,
                PlazaPost.author_id != user_id,
                ~PlazaPost.id.in_([p.id for p in posts])
            ).order_by(
                desc(PlazaPost.like_count * 3 + PlazaPost.comment_count * 2 + PlazaPost.view_count)
            ).limit(limit - len(posts)).all()
            
            posts.extend(additional_posts)
        
        return posts[:limit]
        
    except Exception as e:
        print(f"获取推荐帖子失败: {e}")
        # 降级到简单的热门帖子
        return db.query(PlazaPost).filter(
            PlazaPost.status == PostStatus.PUBLISHED.value
        ).order_by(desc(PlazaPost.like_count)).limit(limit).all()

def update_post_stats(db: Session, post_id: str, stat_type: str, increment: int = 1):
    """更新帖子统计数据"""
    try:
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        if not post:
            return False
        
        if stat_type == "view":
            post.view_count += increment
        elif stat_type == "like":
            post.like_count += increment
        elif stat_type == "comment":
            post.comment_count += increment
        elif stat_type == "favorite":
            post.favorite_count += increment
        
        db.commit()
        return True
        
    except Exception as e:
        print(f"更新帖子统计失败: {e}")
        db.rollback()
        return False

def get_post_analytics(db: Session, post_id: str) -> Dict[str, Any]:
    """获取帖子分析数据"""
    try:
        post = db.query(PlazaPost).filter(PlazaPost.id == post_id).first()
        if not post:
            return {}
        
        # 基础统计
        analytics = {
            "post_id": post_id,
            "view_count": post.view_count,
            "like_count": post.like_count,
            "comment_count": post.comment_count,
            "favorite_count": post.favorite_count,
            "engagement_rate": 0,
            "daily_views": [],
            "daily_likes": [],
            "top_comments": []
        }
        
        # 计算参与度（点赞+评论）/ 浏览量
        if post.view_count > 0:
            analytics["engagement_rate"] = round(
                (post.like_count + post.comment_count) / post.view_count * 100, 2
            )
        
        # 获取最近7天的数据趋势（这里简化处理）
        # 实际应该有专门的统计表记录每日数据
        
        # 获取热门评论
        top_comments = db.query(PlazaComment).filter(
            PlazaComment.post_id == post_id,
            PlazaComment.is_deleted == False
        ).order_by(desc(PlazaComment.like_count)).limit(5).all()
        
        analytics["top_comments"] = [
            {
                "id": comment.id,
                "content": comment.content[:100] + "..." if len(comment.content) > 100 else comment.content,
                "user_nickname": comment.user_nickname,
                "like_count": comment.like_count,
                "created_at": comment.created_at.isoformat() if comment.created_at else None
            }
            for comment in top_comments
        ]
        
        return analytics
        
    except Exception as e:
        print(f"获取帖子分析数据失败: {e}")
        return {}

def validate_post_permissions(post: PlazaPost, user_id: str, user_role: str = "user") -> Dict[str, bool]:
    """验证用户对帖子的权限"""
    permissions = {
        "can_view": True,
        "can_edit": False,
        "can_delete": False,
        "can_comment": True,
        "can_like": True,
        "can_manage": False
    }
    
    # 管理员拥有所有权限
    if user_role == "admin":
        permissions.update({
            "can_edit": True,
            "can_delete": True,
            "can_manage": True
        })
        return permissions
    
    # 作者权限
    if post.author_id == user_id:
        permissions.update({
            "can_edit": True,
            "can_delete": True
        })
    
    # 隐藏帖子只有作者和管理员可以查看
    if post.status == PostStatus.HIDDEN.value and post.author_id != user_id and user_role != "admin":
        permissions["can_view"] = False
    
    # 草稿只有作者可以查看
    if post.status == PostStatus.DRAFT.value and post.author_id != user_id:
        permissions["can_view"] = False
    
    # 不允许评论的帖子
    if not post.allow_comments:
        permissions["can_comment"] = False
    
    return permissions

class PlazaPostFilter:
    """广场帖子过滤器"""
    
    @staticmethod
    def apply_filters(query, filters: Dict[str, Any]):
        """应用过滤条件"""
        
        # 分类过滤
        if filters.get("category"):
            query = query.filter(PlazaPost.category == filters["category"])
        
        # 标签过滤
        if filters.get("tags"):
            tags = filters["tags"] if isinstance(filters["tags"], list) else [filters["tags"]]
            for tag in tags:
                query = query.filter(PlazaPost.tags.contains(tag))
        
        # 状态过滤
        if filters.get("status"):
            query = query.filter(PlazaPost.status == filters["status"])
        else:
            # 默认只显示已发布的帖子
            query = query.filter(PlazaPost.status == PostStatus.PUBLISHED.value)
        
        # 作者过滤
        if filters.get("author_id"):
            query = query.filter(PlazaPost.author_id == filters["author_id"])
        
        # 时间范围过滤
        if filters.get("date_from"):
            query = query.filter(PlazaPost.created_at >= filters["date_from"])
        
        if filters.get("date_to"):
            query = query.filter(PlazaPost.created_at <= filters["date_to"])
        
        # 关键词搜索
        if filters.get("keyword"):
            keyword = f"%{filters['keyword']}%"
            query = query.filter(
                or_(
                    PlazaPost.title.contains(keyword),
                    PlazaPost.content.contains(keyword),
                    PlazaPost.tags.contains(keyword)
                )
            )
        
        return query
    
    @staticmethod
    def apply_sorting(query, sort_by: str = "created_at", order: str = "desc"):
        """应用排序"""
        
        if sort_by == "created_at":
            order_by = PlazaPost.created_at
        elif sort_by == "updated_at":
            order_by = PlazaPost.updated_at
        elif sort_by == "view_count":
            order_by = PlazaPost.view_count
        elif sort_by == "like_count":
            order_by = PlazaPost.like_count
        elif sort_by == "comment_count":
            order_by = PlazaPost.comment_count
        elif sort_by == "hot":
            # 热度排序：点赞*3 + 评论*2 + 浏览*1
            order_by = PlazaPost.like_count * 3 + PlazaPost.comment_count * 2 + PlazaPost.view_count
        else:
            order_by = PlazaPost.created_at
        
        if order == "desc":
            query = query.order_by(desc(order_by))
        else:
            query = query.order_by(order_by)
        
        return query