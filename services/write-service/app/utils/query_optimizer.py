from sqlalchemy.orm import Session, joinedload, selectinload
from sqlalchemy import func, and_, or_, desc, asc
from typing import List, Optional, Dict, Any
from app.models.letter import Letter, LetterStatus
from app.models.read_log import ReadLog

class LetterQueryOptimizer:
    """信件查询优化器"""
    
    @staticmethod
    def get_user_letters_optimized(
        db: Session,
        user_id: str,
        status_filter: Optional[LetterStatus] = None,
        page: int = 1,
        limit: int = 10,
        include_read_stats: bool = False
    ) -> Dict[str, Any]:
        """
        优化的用户信件列表查询
        
        Args:
            db: 数据库会话
            user_id: 用户ID
            status_filter: 状态过滤
            page: 页码
            limit: 每页数量
            include_read_stats: 是否包含阅读统计
            
        Returns:
            Dict[str, Any]: 查询结果
        """
        # 构建基础查询，使用优化的索引
        query = db.query(Letter).filter(Letter.sender_id == user_id)
        
        # 状态过滤 - 利用复合索引 idx_letters_sender_status
        if status_filter:
            query = query.filter(Letter.status == status_filter)
        
        # 总数查询优化：使用索引覆盖扫描
        if status_filter:
            # 使用复合索引计数
            total_query = db.query(func.count(Letter.id)).filter(
                and_(Letter.sender_id == user_id, Letter.status == status_filter)
            )
        else:
            total_query = db.query(func.count(Letter.id)).filter(Letter.sender_id == user_id)
        
        total = total_query.scalar()
        
        # 分页查询 - 利用复合索引 idx_letters_sender_created
        offset = (page - 1) * limit
        letters_query = query.order_by(desc(Letter.created_at)).offset(offset).limit(limit)
        
        # 如果需要阅读统计，使用预加载避免N+1问题
        if include_read_stats:
            letters_query = letters_query.options(selectinload(Letter.read_logs))
        
        letters = letters_query.all()
        
        # 构建结果
        result = {
            "letters": [letter.to_dict() for letter in letters],
            "total": total,
            "page": page,
            "limit": limit,
            "pages": (total + limit - 1) // limit
        }
        
        # 如果需要阅读统计，批量计算
        if include_read_stats:
            result["read_stats"] = {
                letter.id: {
                    "total_reads": len(letter.read_logs),
                    "unique_readers": len(set(log.reader_ip for log in letter.read_logs if log.reader_ip))
                }
                for letter in letters
            }
        
        return result
    
    @staticmethod
    def get_letters_by_status_optimized(
        db: Session,
        status: LetterStatus,
        limit: int = 100,
        include_sender_info: bool = False
    ) -> List[Letter]:
        """
        按状态获取信件的优化查询（用于管理员仪表板）
        
        Args:
            db: 数据库会话
            status: 信件状态
            limit: 限制数量
            include_sender_info: 是否包含发送者信息
            
        Returns:
            List[Letter]: 信件列表
        """
        # 利用复合索引 idx_letters_status_created
        query = db.query(Letter).filter(Letter.status == status)
        query = query.order_by(desc(Letter.created_at)).limit(limit)
        
        letters = query.all()
        
        # 如果需要发送者信息，可以在这里批量获取用户信息
        # 避免在模板中进行N+1查询
        
        return letters
    
    @staticmethod
    def search_letters_optimized(
        db: Session,
        search_text: str,
        user_id: Optional[str] = None,
        limit: int = 50
    ) -> List[Letter]:
        """
        优化的信件搜索（使用全文搜索索引）
        
        Args:
            db: 数据库会话
            search_text: 搜索文本
            user_id: 用户ID（可选，限制搜索范围）
            limit: 限制数量
            
        Returns:
            List[Letter]: 搜索结果
        """
        # 使用PostgreSQL的全文搜索功能
        search_vector = func.to_tsvector('simple', Letter.title)
        search_query = func.plainto_tsquery('simple', search_text)
        
        query = db.query(Letter).filter(search_vector.match(search_query))
        
        if user_id:
            query = query.filter(Letter.sender_id == user_id)
        
        # 按相关性排序
        query = query.order_by(
            func.ts_rank(search_vector, search_query).desc(),
            desc(Letter.created_at)
        ).limit(limit)
        
        return query.all()
    
    @staticmethod
    def get_hot_letters_optimized(
        db: Session,
        limit: int = 20,
        days: int = 7
    ) -> List[Dict[str, Any]]:
        """
        获取热门信件（按阅读量排序）
        
        Args:
            db: 数据库会话
            limit: 限制数量
            days: 统计天数
            
        Returns:
            List[Dict[str, Any]]: 热门信件列表
        """
        # 使用子查询优化，避免大表JOIN
        from datetime import datetime, timedelta
        
        cutoff_date = datetime.utcnow() - timedelta(days=days)
        
        # 子查询：统计每个信件的阅读次数
        read_counts = db.query(
            ReadLog.letter_id,
            func.count(ReadLog.id).label('read_count'),
            func.count(func.distinct(ReadLog.reader_ip)).label('unique_readers')
        ).filter(
            ReadLog.read_at >= cutoff_date
        ).group_by(ReadLog.letter_id).subquery()
        
        # 主查询：获取信件信息和阅读统计
        query = db.query(
            Letter,
            read_counts.c.read_count,
            read_counts.c.unique_readers
        ).join(
            read_counts, Letter.id == read_counts.c.letter_id
        ).filter(
            Letter.status.in_([LetterStatus.DELIVERED, LetterStatus.IN_TRANSIT])
        ).order_by(
            desc(read_counts.c.read_count),
            desc(Letter.created_at)
        ).limit(limit)
        
        results = []
        for letter, read_count, unique_readers in query.all():
            result = letter.to_dict()
            result.update({
                "read_count": read_count,
                "unique_readers": unique_readers,
                "popularity_score": read_count * 0.7 + unique_readers * 0.3
            })
            results.append(result)
        
        return results
    
    @staticmethod
    def get_letter_analytics_optimized(
        db: Session,
        letter_id: str
    ) -> Dict[str, Any]:
        """
        获取信件的详细分析数据（优化版）
        
        Args:
            db: 数据库会话
            letter_id: 信件ID
            
        Returns:
            Dict[str, Any]: 分析数据
        """
        # 使用单次查询获取所有统计数据
        stats_query = db.query(
            func.count(ReadLog.id).label('total_reads'),
            func.count(func.distinct(ReadLog.reader_ip)).label('unique_ips'),
            func.avg(ReadLog.read_duration).label('avg_duration'),
            func.sum(
                func.case([(ReadLog.is_complete_read == True, 1)], else_=0)
            ).label('complete_reads'),
            func.min(ReadLog.read_at).label('first_read'),
            func.max(ReadLog.read_at).label('last_read')
        ).filter(ReadLog.letter_id == letter_id)
        
        result = stats_query.first()
        
        if not result or result.total_reads == 0:
            return {
                "total_reads": 0,
                "unique_readers": 0,
                "average_duration": 0,
                "completion_rate": 0,
                "first_read": None,
                "last_read": None
            }
        
        return {
            "total_reads": result.total_reads or 0,
            "unique_readers": result.unique_ips or 0,
            "average_duration": float(result.avg_duration or 0),
            "completion_rate": (result.complete_reads or 0) / result.total_reads,
            "first_read": result.first_read.isoformat() if result.first_read else None,
            "last_read": result.last_read.isoformat() if result.last_read else None
        }

# 便捷函数
def get_optimized_user_letters(
    db: Session,
    user_id: str,
    status_filter: Optional[LetterStatus] = None,
    page: int = 1,
    limit: int = 10
) -> Dict[str, Any]:
    """获取用户信件列表的便捷函数"""
    return LetterQueryOptimizer.get_user_letters_optimized(
        db, user_id, status_filter, page, limit
    )