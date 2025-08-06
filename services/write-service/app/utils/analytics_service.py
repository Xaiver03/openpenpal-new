"""
阅读分析服务模块
"""
import json
import logging
from typing import Dict, List, Any, Optional, Tuple
from datetime import datetime, timedelta, timezone
from sqlalchemy.orm import Session
from sqlalchemy import func, and_, or_, desc, asc, text
from collections import defaultdict, Counter
import pandas as pd

from app.models.read_log import ReadLog
from app.models.letter import Letter
from app.schemas.analytics import TimeRangeEnum
from app.utils.cache_manager import get_cache_manager

logger = logging.getLogger(__name__)


class AnalyticsService:
    """阅读分析服务类"""
    
    def __init__(self):
        self.cache = get_cache_manager()
        self.cache_ttl = 300  # 5分钟缓存
    
    def get_time_range_bounds(
        self, 
        time_range: TimeRangeEnum, 
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None
    ) -> Tuple[datetime, datetime]:
        """
        获取时间范围的起始和结束时间
        
        Args:
            time_range: 时间范围类型
            start_date: 自定义开始时间
            end_date: 自定义结束时间
            
        Returns:
            Tuple[datetime, datetime]: (开始时间, 结束时间)
        """
        now = datetime.utcnow()
        
        if time_range == TimeRangeEnum.CUSTOM:
            if not start_date or not end_date:
                raise ValueError("Custom time range requires both start_date and end_date")
            return start_date, end_date
        elif time_range == TimeRangeEnum.HOUR:
            return now - timedelta(hours=1), now
        elif time_range == TimeRangeEnum.DAY:
            return now - timedelta(days=1), now
        elif time_range == TimeRangeEnum.WEEK:
            return now - timedelta(weeks=1), now
        elif time_range == TimeRangeEnum.MONTH:
            return now - timedelta(days=30), now
        elif time_range == TimeRangeEnum.QUARTER:
            return now - timedelta(days=90), now
        elif time_range == TimeRangeEnum.YEAR:
            return now - timedelta(days=365), now
        else:
            return now - timedelta(days=1), now
    
    def get_reading_stats(
        self,
        db: Session,
        time_range: TimeRangeEnum = TimeRangeEnum.DAY,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        letter_id: Optional[str] = None,
        user_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        获取阅读统计数据
        
        Args:
            db: 数据库会话
            time_range: 时间范围
            start_date: 开始时间
            end_date: 结束时间  
            letter_id: 特定信件ID
            user_id: 特定用户ID
            
        Returns:
            Dict[str, Any]: 统计数据
        """
        # 生成缓存键
        cache_key = f"reading_stats:{time_range}:{start_date}:{end_date}:{letter_id}:{user_id}"
        
        # 尝试从缓存获取
        cached_result = self.cache.get(cache_key)
        if cached_result:
            return cached_result
        
        # 获取时间范围
        start_time, end_time = self.get_time_range_bounds(time_range, start_date, end_date)
        
        # 构建查询条件
        query = db.query(ReadLog).filter(
            and_(
                ReadLog.read_at >= start_time,
                ReadLog.read_at <= end_time
            )
        )
        
        # 添加额外过滤条件
        if letter_id:
            query = query.filter(ReadLog.letter_id == letter_id)
        
        if user_id:
            # 通过信件关联查询特定用户的信件阅读情况
            query = query.join(Letter).filter(Letter.sender_id == user_id)
        
        # 执行查询
        read_logs = query.all()
        
        # 统计计算
        total_reads = len(read_logs)
        unique_readers = len(set(log.reader_ip for log in read_logs))
        
        # 阅读时长统计
        valid_durations = [log.read_duration for log in read_logs if log.read_duration and log.read_duration > 0]
        avg_read_duration = sum(valid_durations) / len(valid_durations) if valid_durations else 0
        
        # 完整阅读率
        complete_reads = sum(1 for log in read_logs if log.is_complete_read)
        complete_read_rate = complete_reads / total_reads if total_reads > 0 else 0
        
        # 设备和浏览器分布
        device_distribution = defaultdict(int)
        browser_distribution = defaultdict(int)
        location_distribution = defaultdict(int)
        hourly_distribution = defaultdict(int)
        
        for log in read_logs:
            # 解析设备信息
            if log.device_info:
                try:
                    device_data = json.loads(log.device_info)
                    device_type = device_data.get('device_type', 'unknown')
                    browser = device_data.get('browser', 'unknown')
                    
                    device_distribution[device_type] += 1
                    browser_distribution[browser] += 1
                except:
                    device_distribution['unknown'] += 1
                    browser_distribution['unknown'] += 1
            else:
                device_distribution['unknown'] += 1
                browser_distribution['unknown'] += 1
            
            # 地理位置统计
            if log.reader_location:
                location_distribution[log.reader_location] += 1
            else:
                location_distribution['未知'] += 1
            
            # 小时分布统计
            hour = log.read_at.hour
            hourly_distribution[str(hour)] += 1
        
        result = {
            "total_reads": total_reads,
            "unique_readers": unique_readers,
            "avg_read_duration": round(avg_read_duration, 2),
            "complete_read_rate": round(complete_read_rate, 4),
            "device_distribution": dict(device_distribution),
            "browser_distribution": dict(browser_distribution),
            "location_distribution": dict(location_distribution),
            "hourly_distribution": dict(hourly_distribution)
        }
        
        # 缓存结果
        self.cache.set(cache_key, result, self.cache_ttl)
        
        return result
    
    def get_letter_analytics(
        self,
        db: Session,
        letter_id: str
    ) -> Dict[str, Any]:
        """
        获取单个信件的详细分析
        
        Args:
            db: 数据库会话
            letter_id: 信件ID
            
        Returns:
            Dict[str, Any]: 信件分析数据
        """
        cache_key = f"letter_analytics:{letter_id}"
        
        # 尝试从缓存获取
        cached_result = self.cache.get(cache_key)
        if cached_result:
            return cached_result
        
        # 获取信件信息
        letter = db.query(Letter).filter(Letter.id == letter_id).first()
        if not letter:
            raise ValueError(f"Letter {letter_id} not found")
        
        # 获取所有阅读日志
        read_logs = db.query(ReadLog).filter(ReadLog.letter_id == letter_id).order_by(ReadLog.read_at).all()
        
        if not read_logs:
            return {
                "letter_id": letter_id,
                "letter_title": letter.title,
                "total_reads": 0,
                "unique_readers": 0,
                "first_read_at": None,
                "last_read_at": None,
                "avg_read_duration": 0,
                "max_read_duration": 0,
                "complete_reads": 0,
                "device_stats": {},
                "browser_stats": {},
                "time_distribution": []
            }
        
        # 基础统计
        total_reads = len(read_logs)
        unique_readers = len(set(log.reader_ip for log in read_logs))
        first_read_at = read_logs[0].read_at
        last_read_at = read_logs[-1].read_at
        
        # 阅读时长分析
        valid_durations = [log.read_duration for log in read_logs if log.read_duration and log.read_duration > 0]
        avg_read_duration = sum(valid_durations) / len(valid_durations) if valid_durations else 0
        max_read_duration = max(valid_durations) if valid_durations else 0
        
        # 完整阅读统计
        complete_reads = sum(1 for log in read_logs if log.is_complete_read)
        
        # 设备和浏览器统计
        device_stats = defaultdict(int)
        browser_stats = defaultdict(int)
        
        for log in read_logs:
            if log.device_info:
                try:
                    device_data = json.loads(log.device_info)
                    device_stats[device_data.get('device_type', 'unknown')] += 1
                    browser_stats[device_data.get('browser', 'unknown')] += 1
                except:
                    device_stats['unknown'] += 1
                    browser_stats['unknown'] += 1
        
        # 时间分布（按小时）
        time_distribution = []
        hourly_counts = defaultdict(int)
        
        for log in read_logs:
            hour = log.read_at.hour
            hourly_counts[hour] += 1
        
        for hour in range(24):
            time_distribution.append({
                "hour": hour,
                "count": hourly_counts[hour],
                "label": f"{hour:02d}:00"
            })
        
        result = {
            "letter_id": letter_id,
            "letter_title": letter.title,
            "total_reads": total_reads,
            "unique_readers": unique_readers,
            "first_read_at": first_read_at.isoformat() if first_read_at else None,
            "last_read_at": last_read_at.isoformat() if last_read_at else None,
            "avg_read_duration": round(avg_read_duration, 2),
            "max_read_duration": max_read_duration,
            "complete_reads": complete_reads,
            "device_stats": dict(device_stats),
            "browser_stats": dict(browser_stats),
            "time_distribution": time_distribution
        }
        
        # 缓存结果
        self.cache.set(cache_key, result, self.cache_ttl)
        
        return result
    
    def get_user_reading_behavior(
        self,
        db: Session,
        user_id: str,
        time_range: TimeRangeEnum = TimeRangeEnum.MONTH
    ) -> Dict[str, Any]:
        """
        获取用户阅读行为分析
        
        Args:
            db: 数据库会话
            user_id: 用户ID
            time_range: 时间范围
            
        Returns:
            Dict[str, Any]: 用户行为分析数据
        """
        cache_key = f"user_behavior:{user_id}:{time_range}"
        
        # 尝试从缓存获取
        cached_result = self.cache.get(cache_key)
        if cached_result:
            return cached_result
        
        # 获取时间范围
        start_time, end_time = self.get_time_range_bounds(time_range)
        
        # 获取用户发送的信件
        user_letters = db.query(Letter).filter(
            and_(
                Letter.sender_id == user_id,
                Letter.created_at >= start_time,
                Letter.created_at <= end_time
            )
        ).all()
        
        if not user_letters:
            return {
                "user_id": user_id,
                "total_letters_sent": 0,
                "total_reads_received": 0,
                "avg_reads_per_letter": 0,
                "most_read_letter": None,
                "reading_time_stats": {},
                "reader_demographics": {}
            }
        
        letter_ids = [letter.id for letter in user_letters]
        
        # 获取这些信件的所有阅读日志
        read_logs = db.query(ReadLog).filter(
            and_(
                ReadLog.letter_id.in_(letter_ids),
                ReadLog.read_at >= start_time,
                ReadLog.read_at <= end_time
            )
        ).all()
        
        # 统计分析
        total_letters_sent = len(user_letters)
        total_reads_received = len(read_logs)
        avg_reads_per_letter = total_reads_received / total_letters_sent if total_letters_sent > 0 else 0
        
        # 找出最受欢迎的信件
        letter_read_counts = defaultdict(int)
        for log in read_logs:
            letter_read_counts[log.letter_id] += 1
        
        most_read_letter = None
        if letter_read_counts:
            most_read_letter_id = max(letter_read_counts.items(), key=lambda x: x[1])[0]
            most_read_letter_obj = next(letter for letter in user_letters if letter.id == most_read_letter_id)
            most_read_letter = {
                "letter_id": most_read_letter_id,
                "title": most_read_letter_obj.title,
                "read_count": letter_read_counts[most_read_letter_id]
            }
        
        # 阅读时间统计
        valid_durations = [log.read_duration for log in read_logs if log.read_duration and log.read_duration > 0]
        reading_time_stats = {
            "avg_duration": sum(valid_durations) / len(valid_durations) if valid_durations else 0,
            "max_duration": max(valid_durations) if valid_durations else 0,
            "min_duration": min(valid_durations) if valid_durations else 0,
            "total_reading_time": sum(valid_durations) if valid_durations else 0
        }
        
        # 读者画像分析
        device_types = defaultdict(int)
        browsers = defaultdict(int)
        unique_readers = set()
        
        for log in read_logs:
            unique_readers.add(log.reader_ip)
            
            if log.device_info:
                try:
                    device_data = json.loads(log.device_info)
                    device_types[device_data.get('device_type', 'unknown')] += 1
                    browsers[device_data.get('browser', 'unknown')] += 1
                except:
                    device_types['unknown'] += 1
                    browsers['unknown'] += 1
        
        reader_demographics = {
            "unique_readers": len(unique_readers),
            "device_preferences": dict(device_types),
            "browser_preferences": dict(browsers)
        }
        
        result = {
            "user_id": user_id,
            "total_letters_sent": total_letters_sent,
            "total_reads_received": total_reads_received,
            "avg_reads_per_letter": round(avg_reads_per_letter, 2),
            "most_read_letter": most_read_letter,
            "reading_time_stats": {
                "avg_duration": round(reading_time_stats["avg_duration"], 2),
                "max_duration": reading_time_stats["max_duration"],
                "min_duration": reading_time_stats["min_duration"],
                "total_reading_time": round(reading_time_stats["total_reading_time"], 2)
            },
            "reader_demographics": reader_demographics
        }
        
        # 缓存结果
        self.cache.set(cache_key, result, self.cache_ttl)
        
        return result
    
    def get_trend_analysis(
        self,
        db: Session,
        time_range: TimeRangeEnum = TimeRangeEnum.MONTH,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None
    ) -> Dict[str, Any]:
        """
        获取趋势分析数据
        
        Args:
            db: 数据库会话
            time_range: 时间范围
            start_date: 开始时间
            end_date: 结束时间
            
        Returns:
            Dict[str, Any]: 趋势分析数据
        """
        cache_key = f"trend_analysis:{time_range}:{start_date}:{end_date}"
        
        # 尝试从缓存获取
        cached_result = self.cache.get(cache_key)
        if cached_result:
            return cached_result
        
        # 获取时间范围
        start_time, end_time = self.get_time_range_bounds(time_range, start_date, end_date)
        
        # 获取阅读日志
        read_logs = db.query(ReadLog).filter(
            and_(
                ReadLog.read_at >= start_time,
                ReadLog.read_at <= end_time
            )
        ).order_by(ReadLog.read_at).all()
        
        if not read_logs:
            return {
                "time_series": [],
                "growth_rate": 0,
                "peak_hours": [],
                "peak_days": [],
                "seasonal_patterns": {}
            }
        
        # 生成时间序列数据
        time_series = []
        
        # 根据时间范围决定分组粒度
        if time_range in [TimeRangeEnum.HOUR, TimeRangeEnum.DAY]:
            # 按小时分组
            hourly_counts = defaultdict(int)
            for log in read_logs:
                hour_key = log.read_at.strftime('%Y-%m-%d %H:00')
                hourly_counts[hour_key] += 1
            
            for hour_key in sorted(hourly_counts.keys()):
                time_series.append({
                    "time": hour_key,
                    "count": hourly_counts[hour_key],
                    "timestamp": datetime.fromisoformat(hour_key.replace(' ', 'T')).isoformat()
                })
        
        elif time_range in [TimeRangeEnum.WEEK, TimeRangeEnum.MONTH]:
            # 按天分组
            daily_counts = defaultdict(int)
            for log in read_logs:
                day_key = log.read_at.strftime('%Y-%m-%d')
                daily_counts[day_key] += 1
            
            for day_key in sorted(daily_counts.keys()):
                time_series.append({
                    "time": day_key,
                    "count": daily_counts[day_key],
                    "timestamp": datetime.fromisoformat(day_key).isoformat()
                })
        
        else:
            # 按月分组
            monthly_counts = defaultdict(int)
            for log in read_logs:
                month_key = log.read_at.strftime('%Y-%m')
                monthly_counts[month_key] += 1
            
            for month_key in sorted(monthly_counts.keys()):
                time_series.append({
                    "time": month_key + "-01",
                    "count": monthly_counts[month_key],
                    "timestamp": datetime.fromisoformat(month_key + "-01").isoformat()
                })
        
        # 计算增长率
        growth_rate = 0
        if len(time_series) >= 2:
            first_count = time_series[0]["count"]
            last_count = time_series[-1]["count"]
            if first_count > 0:
                growth_rate = ((last_count - first_count) / first_count) * 100
        
        # 找出高峰时段
        hour_counts = defaultdict(int)
        day_counts = defaultdict(int)
        
        for log in read_logs:
            hour_counts[log.read_at.hour] += 1
            day_counts[log.read_at.strftime('%A')] += 1
        
        # 排序找出前3个高峰时段
        peak_hours = sorted(hour_counts.items(), key=lambda x: x[1], reverse=True)[:3]
        peak_hours = [hour for hour, _ in peak_hours]
        
        peak_days = sorted(day_counts.items(), key=lambda x: x[1], reverse=True)[:3]
        peak_days = [day for day, _ in peak_days]
        
        # 季节性模式分析
        seasonal_patterns = {
            "hourly": dict(hour_counts),
            "daily": dict(day_counts),
            "monthly": defaultdict(int)
        }
        
        for log in read_logs:
            month = log.read_at.month
            seasonal_patterns["monthly"][month] += 1
        
        seasonal_patterns["monthly"] = dict(seasonal_patterns["monthly"])
        
        result = {
            "time_series": time_series,
            "growth_rate": round(growth_rate, 2),
            "peak_hours": peak_hours,
            "peak_days": peak_days,
            "seasonal_patterns": seasonal_patterns
        }
        
        # 缓存结果
        self.cache.set(cache_key, result, self.cache_ttl)
        
        return result
    
    def get_popular_content(
        self,
        db: Session,
        limit: int = 10,
        time_range: TimeRangeEnum = TimeRangeEnum.WEEK
    ) -> Dict[str, Any]:
        """
        获取热门内容排行
        
        Args:
            db: 数据库会话
            limit: 返回数量限制
            time_range: 时间范围
            
        Returns:
            Dict[str, Any]: 热门内容数据
        """
        cache_key = f"popular_content:{limit}:{time_range}"
        
        # 尝试从缓存获取
        cached_result = self.cache.get(cache_key)
        if cached_result:
            return cached_result
        
        # 获取时间范围
        start_time, end_time = self.get_time_range_bounds(time_range)
        
        # 查询热门信件
        letter_stats = db.query(
            ReadLog.letter_id,
            func.count(ReadLog.id).label('read_count'),
            func.count(func.distinct(ReadLog.reader_ip)).label('unique_readers'),
            func.avg(ReadLog.read_duration).label('avg_duration')
        ).filter(
            and_(
                ReadLog.read_at >= start_time,
                ReadLog.read_at <= end_time
            )
        ).group_by(ReadLog.letter_id).order_by(desc('read_count')).limit(limit).all()
        
        # 获取信件详情
        top_letters = []
        for stat in letter_stats:
            letter = db.query(Letter).filter(Letter.id == stat.letter_id).first()
            if letter:
                top_letters.append({
                    "letter_id": letter.id,
                    "title": letter.title,
                    "sender_id": letter.sender_id,
                    "read_count": stat.read_count,
                    "unique_readers": stat.unique_readers,
                    "avg_duration": round(float(stat.avg_duration) if stat.avg_duration else 0, 2),
                    "created_at": letter.created_at.isoformat()
                })
        
        # 查询活跃用户
        user_stats = db.query(
            Letter.sender_id,
            func.count(func.distinct(Letter.id)).label('letters_count'),
            func.count(ReadLog.id).label('total_reads')
        ).join(ReadLog, Letter.id == ReadLog.letter_id).filter(
            and_(
                ReadLog.read_at >= start_time,
                ReadLog.read_at <= end_time
            )
        ).group_by(Letter.sender_id).order_by(desc('total_reads')).limit(limit).all()
        
        top_users = []
        for stat in user_stats:
            top_users.append({
                "user_id": stat.sender_id,
                "letters_count": stat.letters_count,
                "total_reads": stat.total_reads,
                "avg_reads_per_letter": round(stat.total_reads / stat.letters_count, 2)
            })
        
        result = {
            "top_letters": top_letters,
            "top_users": top_users,
            "trending_topics": []  # 可以后续添加话题分析
        }
        
        # 缓存结果
        self.cache.set(cache_key, result, self.cache_ttl)
        
        return result
    
    def get_realtime_stats(self, db: Session) -> Dict[str, Any]:
        """
        获取实时统计数据
        
        Args:
            db: 数据库会话
            
        Returns:
            Dict[str, Any]: 实时统计数据
        """
        now = datetime.utcnow()
        one_hour_ago = now - timedelta(hours=1)
        today_start = now.replace(hour=0, minute=0, second=0, microsecond=0)
        
        # 过去1小时的阅读数
        reads_last_hour = db.query(ReadLog).filter(ReadLog.read_at >= one_hour_ago).count()
        
        # 今日阅读数
        reads_today = db.query(ReadLog).filter(ReadLog.read_at >= today_start).count()
        
        # 活跃信件（过去1小时有阅读的信件）
        active_letter_ids = db.query(ReadLog.letter_id).filter(
            ReadLog.read_at >= one_hour_ago
        ).distinct().all()
        
        active_letters = []
        for (letter_id,) in active_letter_ids[:10]:  # 限制返回数量
            letter = db.query(Letter).filter(Letter.id == letter_id).first()
            if letter:
                recent_reads = db.query(ReadLog).filter(
                    and_(
                        ReadLog.letter_id == letter_id,
                        ReadLog.read_at >= one_hour_ago
                    )
                ).count()
                
                active_letters.append({
                    "letter_id": letter.id,
                    "title": letter.title,
                    "recent_reads": recent_reads
                })
        
        # 实时事件流（最近的阅读事件）
        recent_events = db.query(ReadLog).order_by(desc(ReadLog.read_at)).limit(5).all()
        
        live_events = []
        for event in recent_events:
            letter = db.query(Letter).filter(Letter.id == event.letter_id).first()
            if letter:
                live_events.append({
                    "event_type": "letter_read",
                    "letter_id": event.letter_id,
                    "letter_title": letter.title,
                    "read_at": event.read_at.isoformat(),
                    "duration": event.read_duration,
                    "complete": event.is_complete_read
                })
        
        return {
            "current_online_readers": 0,  # 需要实时WebSocket连接数据
            "reads_last_hour": reads_last_hour,
            "reads_today": reads_today,
            "active_letters": active_letters,
            "live_events": live_events
        }


# 全局服务实例
analytics_service = AnalyticsService()


def get_analytics_service() -> AnalyticsService:
    """获取分析服务实例"""
    return analytics_service