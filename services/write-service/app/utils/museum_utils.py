import random
import string
from typing import Optional, List, Dict, Any, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import desc, and_, or_, func, extract
from datetime import datetime, timedelta
import time

from app.models.museum import (
    MuseumLetter, MuseumFavorite, MuseumRating, TimelineEvent, 
    MuseumCollection, CollectionLetter, MuseumLetterStatus, MuseumEra
)
from app.models.letter import Letter

def generate_museum_letter_id() -> str:
    """生成博物馆信件ID: MS + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    # 避免混淆字符
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"MS{random_part}"

def generate_timeline_event_id() -> str:
    """生成时间线事件ID: TL + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"TL{random_part}"

def generate_collection_id() -> str:
    """生成收藏集ID: CL + 10位随机字符"""
    chars = string.ascii_uppercase + string.digits
    chars = chars.replace('O', '').replace('0', '').replace('I', '').replace('1', '').replace('L', '')
    random_part = ''.join(random.choices(chars, k=10))
    return f"CL{random_part}"

def generate_unique_museum_letter_id(db: Session) -> str:
    """生成唯一的博物馆信件ID"""
    max_attempts = 10
    for _ in range(max_attempts):
        letter_id = generate_museum_letter_id()
        existing = db.query(MuseumLetter).filter(MuseumLetter.id == letter_id).first()
        if not existing:
            return letter_id
    
    # 如果10次都冲突，使用时间戳确保唯一性
    timestamp = int(datetime.now().timestamp() * 1000)
    return f"MS{timestamp % 10000000000:010d}"

def generate_unique_timeline_event_id(db: Session) -> str:
    """生成唯一的时间线事件ID"""
    max_attempts = 10
    for _ in range(max_attempts):
        event_id = generate_timeline_event_id()
        existing = db.query(TimelineEvent).filter(TimelineEvent.id == event_id).first()
        if not existing:
            return event_id
    
    timestamp = int(datetime.now().timestamp() * 1000)
    return f"TL{timestamp % 10000000000:010d}"

def generate_unique_collection_id(db: Session) -> str:
    """生成唯一的收藏集ID"""
    max_attempts = 10
    for _ in range(max_attempts):
        collection_id = generate_collection_id()
        existing = db.query(MuseumCollection).filter(MuseumCollection.id == collection_id).first()
        if not existing:
            return collection_id
    
    timestamp = int(datetime.now().timestamp() * 1000)
    return f"CL{timestamp % 10000000000:010d}"

def create_museum_summary(content: str, max_length: int = 300) -> str:
    """从内容中提取博物馆信件摘要"""
    if not content:
        return ""
    
    # 移除多余的空白字符
    clean_content = ' '.join(content.split())
    
    if len(clean_content) <= max_length:
        return clean_content
    
    # 在合适的位置截断
    summary = clean_content[:max_length]
    
    # 尝试在句号或感叹号处截断
    for delimiter in ['。', '！', '？', '.', '!', '?']:
        pos = summary.rfind(delimiter)
        if pos > max_length * 0.7:
            return summary[:pos + 1]
    
    # 尝试在逗号或空格处截断
    for delimiter in ['，', '、', ',', ' ']:
        pos = summary.rfind(delimiter)
        if pos > max_length * 0.8:
            return summary[:pos] + "..."
    
    return summary + "..."

def get_era_by_date(date: datetime) -> MuseumEra:
    """根据日期自动判断历史时期"""
    if not date:
        return MuseumEra.PRESENT
    
    year = date.year
    
    if year < 1840:
        return MuseumEra.ANCIENT
    elif year < 1919:
        return MuseumEra.MODERN
    elif year < 1949:
        return MuseumEra.CONTEMPORARY
    elif year < 2000:
        return MuseumEra.PRESENT
    else:
        return MuseumEra.DIGITAL

def get_popular_museum_tags(db: Session, limit: int = 30) -> List[str]:
    """获取博物馆热门标签"""
    try:
        # 查询所有已审核的博物馆信件标签
        letters = db.query(MuseumLetter.tags).filter(
            MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value]),
            MuseumLetter.tags.isnot(None),
            MuseumLetter.tags != ""
        ).all()
        
        # 统计标签频率
        tag_count = {}
        for letter in letters:
            if letter.tags:
                tags = [tag.strip() for tag in letter.tags.split(',') if tag.strip()]
                for tag in tags:
                    tag_count[tag] = tag_count.get(tag, 0) + 1
        
        # 按频率排序
        popular_tags = sorted(tag_count.items(), key=lambda x: x[1], reverse=True)
        return [tag for tag, count in popular_tags[:limit]]
        
    except Exception as e:
        print(f"获取博物馆热门标签失败: {e}")
        return []

def get_featured_museum_letters(db: Session, limit: int = 10) -> List[MuseumLetter]:
    """获取精选博物馆信件"""
    try:
        return db.query(MuseumLetter).filter(
            MuseumLetter.status == MuseumLetterStatus.FEATURED.value,
            MuseumLetter.is_featured == True
        ).order_by(
            desc(MuseumLetter.display_order),
            desc(MuseumLetter.view_count)
        ).limit(limit).all()
        
    except Exception as e:
        print(f"获取精选博物馆信件失败: {e}")
        return []

def get_recommended_museum_letters(db: Session, user_id: str, limit: int = 10) -> List[MuseumLetter]:
    """获取推荐博物馆信件"""
    try:
        # 简单推荐算法：基于用户收藏的历史和热门信件
        
        # 获取用户收藏过的信件的时期和分类
        user_preferences = db.query(
            MuseumLetter.era, MuseumLetter.category
        ).join(
            MuseumFavorite, MuseumLetter.id == MuseumFavorite.museum_letter_id
        ).filter(
            MuseumFavorite.user_id == user_id
        ).distinct().all()
        
        preferred_eras = [pref[0] for pref in user_preferences]
        preferred_categories = [pref[1] for pref in user_preferences]
        
        # 构建推荐查询
        query = db.query(MuseumLetter).filter(
            MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value])
        )
        
        # 如果有偏好，优先推荐相似的
        if preferred_eras or preferred_categories:
            query = query.filter(
                or_(
                    MuseumLetter.era.in_(preferred_eras) if preferred_eras else False,
                    MuseumLetter.category.in_(preferred_categories) if preferred_categories else False
                )
            )
        
        # 按热度排序（浏览数 + 收藏数*2 + 评分*10）
        letters = query.order_by(
            desc(MuseumLetter.view_count + MuseumLetter.favorite_count * 2 + MuseumLetter.rating_avg * 10)
        ).limit(limit).all()
        
        # 如果没有足够的推荐结果，补充热门信件
        if len(letters) < limit:
            additional_letters = db.query(MuseumLetter).filter(
                MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value]),
                ~MuseumLetter.id.in_([l.id for l in letters])
            ).order_by(
                desc(MuseumLetter.view_count + MuseumLetter.favorite_count * 2)
            ).limit(limit - len(letters)).all()
            
            letters.extend(additional_letters)
        
        return letters[:limit]
        
    except Exception as e:
        print(f"获取推荐博物馆信件失败: {e}")
        # 降级到简单的热门信件
        return db.query(MuseumLetter).filter(
            MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value])
        ).order_by(desc(MuseumLetter.view_count)).limit(limit).all()

def update_museum_letter_stats(db: Session, letter_id: str, stat_type: str, increment: int = 1):
    """更新博物馆信件统计数据"""
    try:
        letter = db.query(MuseumLetter).filter(MuseumLetter.id == letter_id).first()
        if not letter:
            return False
        
        if stat_type == "view":
            letter.view_count += increment
        elif stat_type == "favorite":
            letter.favorite_count += increment
        elif stat_type == "share":
            letter.share_count += increment
        
        db.commit()
        return True
        
    except Exception as e:
        print(f"更新博物馆信件统计失败: {e}")
        db.rollback()
        return False

def update_museum_letter_rating(db: Session, letter_id: str, new_rating: int, old_rating: Optional[int] = None):
    """更新博物馆信件评分"""
    try:
        letter = db.query(MuseumLetter).filter(MuseumLetter.id == letter_id).first()
        if not letter:
            return False
        
        if old_rating is None:
            # 新评分
            total_score = letter.rating_avg * letter.rating_count + new_rating
            letter.rating_count += 1
            letter.rating_avg = round(total_score / letter.rating_count, 2)
        else:
            # 更新评分
            if letter.rating_count > 0:
                total_score = letter.rating_avg * letter.rating_count - old_rating + new_rating
                letter.rating_avg = round(total_score / letter.rating_count, 2)
        
        db.commit()
        return True
        
    except Exception as e:
        print(f"更新博物馆信件评分失败: {e}")
        db.rollback()
        return False

def get_museum_statistics(db: Session) -> Dict[str, Any]:
    """获取博物馆统计数据"""
    try:
        stats = {}
        
        # 基础统计
        stats["total_letters"] = db.query(MuseumLetter).filter(
            MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value])
        ).count()
        
        stats["total_collections"] = db.query(MuseumCollection).filter(
            MuseumCollection.is_public == True
        ).count()
        
        stats["total_timeline_events"] = db.query(TimelineEvent).count()
        
        # 时期分布
        era_distribution = db.query(
            MuseumLetter.era, func.count(MuseumLetter.id)
        ).filter(
            MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value])
        ).group_by(MuseumLetter.era).all()
        
        stats["era_distribution"] = {era: count for era, count in era_distribution}
        
        # 分类分布
        category_distribution = db.query(
            MuseumLetter.category, func.count(MuseumLetter.id)
        ).filter(
            MuseumLetter.status.in_([MuseumLetterStatus.APPROVED.value, MuseumLetterStatus.FEATURED.value])
        ).group_by(MuseumLetter.category).all()
        
        stats["category_distribution"] = {category: count for category, count in category_distribution}
        
        # 热门标签
        stats["popular_tags"] = get_popular_museum_tags(db, 20)
        
        return stats
        
    except Exception as e:
        print(f"获取博物馆统计数据失败: {e}")
        return {}

def get_timeline_by_date_range(
    db: Session, 
    start_date: datetime, 
    end_date: datetime, 
    limit: int = 100
) -> List[TimelineEvent]:
    """根据日期范围获取时间线事件"""
    try:
        return db.query(TimelineEvent).filter(
            TimelineEvent.event_date >= start_date,
            TimelineEvent.event_date <= end_date
        ).order_by(
            TimelineEvent.event_date.desc(),
            desc(TimelineEvent.importance),
            desc(TimelineEvent.is_featured)
        ).limit(limit).all()
        
    except Exception as e:
        print(f"获取时间线事件失败: {e}")
        return []

def get_timeline_by_era(db: Session, era: MuseumEra, limit: int = 50) -> List[TimelineEvent]:
    """根据历史时期获取时间线事件"""
    try:
        return db.query(TimelineEvent).filter(
            TimelineEvent.era == era.value
        ).order_by(
            TimelineEvent.event_date.desc(),
            desc(TimelineEvent.importance)
        ).limit(limit).all()
        
    except Exception as e:
        print(f"获取时期时间线事件失败: {e}")
        return []

def validate_museum_letter_permissions(letter: MuseumLetter, user_id: str, user_role: str = "user") -> Dict[str, bool]:
    """验证用户对博物馆信件的权限"""
    permissions = {
        "can_view": True,
        "can_edit": False,
        "can_delete": False,
        "can_approve": False,
        "can_feature": False,
        "can_rate": True,
        "can_favorite": True
    }
    
    # 管理员和审核员拥有更多权限
    if user_role in ["admin", "reviewer"]:
        permissions.update({
            "can_edit": True,
            "can_delete": True,
            "can_approve": True,
            "can_feature": True
        })
        return permissions
    
    # 贡献者权限
    if letter.contributor_id == user_id:
        permissions.update({
            "can_edit": letter.status == MuseumLetterStatus.PENDING.value,  # 只能编辑待审核的
            "can_delete": letter.status == MuseumLetterStatus.PENDING.value
        })
    
    # 未审核的信件只有贡献者和管理员可以查看
    if letter.status == MuseumLetterStatus.PENDING.value:
        if letter.contributor_id != user_id and user_role not in ["admin", "reviewer"]:
            permissions["can_view"] = False
    
    # 被拒绝的信件只有贡献者和管理员可以查看
    if letter.status == MuseumLetterStatus.REJECTED.value:
        if letter.contributor_id != user_id and user_role not in ["admin", "reviewer"]:
            permissions["can_view"] = False
    
    return permissions

class MuseumLetterFilter:
    """博物馆信件过滤器"""
    
    @staticmethod
    def apply_filters(query, filters: Dict[str, Any]):
        """应用过滤条件"""
        
        # 状态过滤（默认只显示已审核的）
        if filters.get("status"):
            query = query.filter(MuseumLetter.status == filters["status"])
        else:
            query = query.filter(MuseumLetter.status.in_([
                MuseumLetterStatus.APPROVED.value, 
                MuseumLetterStatus.FEATURED.value
            ]))
        
        # 时期过滤
        if filters.get("era"):
            query = query.filter(MuseumLetter.era == filters["era"])
        
        # 分类过滤
        if filters.get("category"):
            query = query.filter(MuseumLetter.category == filters["category"])
        
        # 标签过滤
        if filters.get("tags"):
            tags = filters["tags"] if isinstance(filters["tags"], list) else [filters["tags"]]
            for tag in tags:
                query = query.filter(MuseumLetter.tags.contains(tag))
        
        # 作者过滤
        if filters.get("author"):
            query = query.filter(MuseumLetter.original_author.contains(filters["author"]))
        
        # 地点过滤
        if filters.get("location"):
            query = query.filter(MuseumLetter.location.contains(filters["location"]))
        
        # 日期范围过滤
        if filters.get("date_from"):
            query = query.filter(MuseumLetter.historical_date >= filters["date_from"])
        
        if filters.get("date_to"):
            query = query.filter(MuseumLetter.historical_date <= filters["date_to"])
        
        # 关键词搜索
        if filters.get("keyword"):
            keyword = f"%{filters['keyword']}%"
            query = query.filter(
                or_(
                    MuseumLetter.title.contains(keyword),
                    MuseumLetter.content.contains(keyword),
                    MuseumLetter.original_author.contains(keyword),
                    MuseumLetter.tags.contains(keyword)
                )
            )
        
        # 精选过滤
        if filters.get("featured"):
            query = query.filter(MuseumLetter.is_featured == True)
        
        return query
    
    @staticmethod
    def apply_sorting(query, sort_by: str = "created_at", order: str = "desc"):
        """应用排序"""
        
        if sort_by == "created_at":
            order_by = MuseumLetter.created_at
        elif sort_by == "historical_date":
            order_by = MuseumLetter.historical_date
        elif sort_by == "view_count":
            order_by = MuseumLetter.view_count
        elif sort_by == "rating":
            order_by = MuseumLetter.rating_avg
        elif sort_by == "favorite_count":
            order_by = MuseumLetter.favorite_count
        elif sort_by == "relevance":
            # 相关性排序：评分*0.4 + 浏览量*0.3 + 收藏数*0.3
            order_by = (MuseumLetter.rating_avg * 0.4 + 
                       MuseumLetter.view_count * 0.3 + 
                       MuseumLetter.favorite_count * 0.3)
        else:
            order_by = MuseumLetter.created_at
        
        if order == "desc":
            query = query.order_by(desc(order_by))
        else:
            query = query.order_by(order_by)
        
        return query