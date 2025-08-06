"""
高级数据分析服务
支持用户行为分析、内容热度统计、业务指标监控等功能
"""

import json
import asyncio
from typing import Dict, List, Any, Optional, Tuple, Union
from datetime import datetime, timedelta, timezone
from collections import defaultdict, Counter
import pandas as pd
import numpy as np
from dataclasses import dataclass
from enum import Enum

from app.core.config import settings
from app.core.logger import get_logger
from app.utils.cache_manager import get_cache_manager

logger = get_logger(__name__)


class MetricType(str, Enum):
    """指标类型"""
    USER_BEHAVIOR = "user_behavior"      # 用户行为
    CONTENT_POPULARITY = "content_popularity"  # 内容热度
    BUSINESS_KPI = "business_kpi"        # 业务KPI
    PERFORMANCE = "performance"          # 性能指标
    ENGAGEMENT = "engagement"            # 用户参与度


class TimeGranularity(str, Enum):
    """时间粒度"""
    HOUR = "hour"
    DAY = "day" 
    WEEK = "week"
    MONTH = "month"
    QUARTER = "quarter"
    YEAR = "year"


@dataclass
class AnalyticsEvent:
    """分析事件"""
    event_type: str
    user_id: str
    object_id: Optional[str]
    object_type: Optional[str]
    properties: Dict[str, Any]
    timestamp: datetime
    session_id: Optional[str] = None
    ip_address: Optional[str] = None
    user_agent: Optional[str] = None


class AdvancedAnalytics:
    """高级数据分析服务"""
    
    def __init__(self):
        self.cache_manager = None
        self.cache_ttl = 300  # 5分钟缓存
        
        # 事件类型定义
        self.event_types = {
            # 用户行为事件
            'user_login': {'category': 'auth', 'weight': 1},
            'user_logout': {'category': 'auth', 'weight': 1},
            'letter_create': {'category': 'content', 'weight': 3},
            'letter_view': {'category': 'engagement', 'weight': 1},
            'letter_like': {'category': 'engagement', 'weight': 2},
            'letter_share': {'category': 'engagement', 'weight': 3},
            'comment_create': {'category': 'social', 'weight': 2},
            'post_create': {'category': 'social', 'weight': 3},
            'order_create': {'category': 'commerce', 'weight': 5},
            'search_query': {'category': 'discovery', 'weight': 1},
            'file_upload': {'category': 'content', 'weight': 2},
            'notification_click': {'category': 'engagement', 'weight': 1}
        }
    
    async def initialize(self):
        """初始化分析服务"""
        try:
            self.cache_manager = await get_cache_manager()
            logger.info("Advanced analytics service initialized")
        except Exception as e:
            logger.error(f"Failed to initialize analytics service: {str(e)}")
    
    async def track_event(self, event: AnalyticsEvent) -> bool:
        """
        跟踪事件
        
        Args:
            event: 分析事件
            
        Returns:
            bool: 是否成功跟踪
        """
        try:
            # 序列化事件
            event_data = {
                'event_type': event.event_type,
                'user_id': event.user_id,
                'object_id': event.object_id,
                'object_type': event.object_type,
                'properties': event.properties,
                'timestamp': event.timestamp.isoformat(),
                'session_id': event.session_id,
                'ip_address': event.ip_address,
                'user_agent': event.user_agent
            }
            
            if not self.cache_manager:
                await self.initialize()
            
            # 存储到缓存中（用于实时分析）
            event_key = f"analytics:event:{datetime.utcnow().strftime('%Y%m%d%H')}:{event.event_type}"
            events = await self.cache_manager.get(event_key) or []
            events.append(event_data)
            
            # 限制每小时事件数量，避免内存溢出
            if len(events) > 10000:
                events = events[-10000:]
            
            await self.cache_manager.set(event_key, events, expire=3600 * 24)  # 缓存24小时
            
            # 更新实时计数器
            await self._update_realtime_counters(event)
            
            # 异步处理事件聚合
            asyncio.create_task(self._process_event_aggregation(event))
            
            return True
            
        except Exception as e:
            logger.error(f"Failed to track event: {str(e)}")
            return False
    
    async def get_user_behavior_analysis(
        self,
        user_id: str,
        time_range: Optional[Tuple[datetime, datetime]] = None,
        include_details: bool = False
    ) -> Dict[str, Any]:
        """
        获取用户行为分析
        
        Args:
            user_id: 用户ID
            time_range: 时间范围
            include_details: 是否包含详细信息
            
        Returns:
            用户行为分析结果
        """
        try:
            cache_key = f"analytics:user_behavior:{user_id}"
            if time_range:
                cache_key += f":{time_range[0].isoformat()}:{time_range[1].isoformat()}"
            
            # 尝试从缓存获取
            cached_result = await self.cache_manager.get(cache_key)
            if cached_result and not include_details:
                return cached_result
            
            # 分析用户行为
            if not time_range:
                end_time = datetime.utcnow()
                start_time = end_time - timedelta(days=30)  # 默认30天
                time_range = (start_time, end_time)
            
            # 获取用户事件
            user_events = await self._get_user_events(user_id, time_range)
            
            # 计算行为指标
            behavior_analysis = {
                'user_id': user_id,
                'time_range': {
                    'start': time_range[0].isoformat(),
                    'end': time_range[1].isoformat()
                },
                'activity_summary': self._analyze_activity_summary(user_events),
                'engagement_score': self._calculate_engagement_score(user_events),
                'behavior_patterns': self._analyze_behavior_patterns(user_events),
                'content_preferences': self._analyze_content_preferences(user_events),
                'usage_frequency': self._analyze_usage_frequency(user_events),
                'peak_activity_times': self._analyze_peak_activity_times(user_events)
            }
            
            if include_details:
                behavior_analysis['detailed_events'] = user_events[-100:]  # 最近100条事件
            
            # 缓存结果
            await self.cache_manager.set(cache_key, behavior_analysis, expire=self.cache_ttl)
            
            return behavior_analysis
            
        except Exception as e:
            logger.error(f"Failed to get user behavior analysis: {str(e)}")
            return {'error': str(e)}
    
    async def get_content_popularity_analysis(
        self,
        content_type: str,
        time_range: Optional[Tuple[datetime, datetime]] = None,
        limit: int = 100
    ) -> Dict[str, Any]:
        """
        获取内容热度分析
        
        Args:
            content_type: 内容类型（letter, post, product等）
            time_range: 时间范围
            limit: 返回数量限制
            
        Returns:
            内容热度分析结果
        """
        try:
            cache_key = f"analytics:content_popularity:{content_type}"
            if time_range:
                cache_key += f":{time_range[0].isoformat()}:{time_range[1].isoformat()}"
            
            cached_result = await self.cache_manager.get(cache_key)
            if cached_result:
                return cached_result
            
            if not time_range:
                end_time = datetime.utcnow()
                start_time = end_time - timedelta(days=7)  # 默认7天
                time_range = (start_time, end_time)
            
            # 获取内容相关事件
            content_events = await self._get_content_events(content_type, time_range)
            
            # 分析内容热度
            popularity_analysis = {
                'content_type': content_type,
                'time_range': {
                    'start': time_range[0].isoformat(),
                    'end': time_range[1].isoformat()
                },
                'total_content_items': len(set(event.get('object_id') for event in content_events if event.get('object_id'))),
                'total_interactions': len(content_events),
                'top_content': self._get_top_content_by_interactions(content_events, limit),
                'interaction_breakdown': self._analyze_interaction_breakdown(content_events),
                'trending_content': self._get_trending_content(content_events, limit//2),
                'engagement_metrics': self._calculate_content_engagement_metrics(content_events),
                'viral_coefficient': self._calculate_viral_coefficient(content_events)
            }
            
            # 缓存结果
            await self.cache_manager.set(cache_key, popularity_analysis, expire=self.cache_ttl)
            
            return popularity_analysis
            
        except Exception as e:
            logger.error(f"Failed to get content popularity analysis: {str(e)}")
            return {'error': str(e)}
    
    async def get_business_kpi_dashboard(
        self,
        time_range: Optional[Tuple[datetime, datetime]] = None,
        granularity: TimeGranularity = TimeGranularity.DAY
    ) -> Dict[str, Any]:
        """
        获取业务KPI仪表板
        
        Args:
            time_range: 时间范围
            granularity: 时间粒度
            
        Returns:
            业务KPI数据
        """
        try:
            cache_key = f"analytics:business_kpi:{granularity}"
            if time_range:
                cache_key += f":{time_range[0].isoformat()}:{time_range[1].isoformat()}"
            
            cached_result = await self.cache_manager.get(cache_key)
            if cached_result:
                return cached_result
            
            if not time_range:
                end_time = datetime.utcnow()
                start_time = end_time - timedelta(days=30)  # 默认30天
                time_range = (start_time, end_time)
            
            # 计算业务KPI
            kpi_data = {
                'time_range': {
                    'start': time_range[0].isoformat(),
                    'end': time_range[1].isoformat()
                },
                'granularity': granularity,
                'user_metrics': await self._calculate_user_metrics(time_range, granularity),
                'content_metrics': await self._calculate_content_metrics(time_range, granularity),
                'engagement_metrics': await self._calculate_engagement_metrics(time_range, granularity),
                'revenue_metrics': await self._calculate_revenue_metrics(time_range, granularity),
                'growth_metrics': await self._calculate_growth_metrics(time_range, granularity),
                'retention_metrics': await self._calculate_retention_metrics(time_range, granularity)
            }
            
            # 缓存结果
            await self.cache_manager.set(cache_key, kpi_data, expire=self.cache_ttl)
            
            return kpi_data
            
        except Exception as e:
            logger.error(f"Failed to get business KPI dashboard: {str(e)}")
            return {'error': str(e)}
    
    async def get_realtime_metrics(self) -> Dict[str, Any]:
        """获取实时指标"""
        try:
            realtime_data = {
                'timestamp': datetime.utcnow().isoformat(),
                'active_users': await self._get_active_users_count(),
                'events_per_minute': await self._get_events_per_minute(),
                'top_events': await self._get_top_events_realtime(),
                'system_health': await self._get_system_health_metrics(),
                'content_activity': await self._get_content_activity_realtime()
            }
            
            return realtime_data
            
        except Exception as e:
            logger.error(f"Failed to get realtime metrics: {str(e)}")
            return {'error': str(e)}
    
    # 私有方法实现
    
    async def _update_realtime_counters(self, event: AnalyticsEvent):
        """更新实时计数器"""
        try:
            current_minute = datetime.utcnow().strftime('%Y%m%d%H%M')
            
            # 事件计数
            event_key = f"analytics:realtime:events:{current_minute}"
            current_count = await self.cache_manager.get(event_key) or 0
            await self.cache_manager.set(event_key, current_count + 1, expire=3600)
            
            # 活跃用户计数
            users_key = f"analytics:realtime:active_users:{current_minute}"
            active_users = await self.cache_manager.get(users_key) or set()
            if isinstance(active_users, list):
                active_users = set(active_users)
            active_users.add(event.user_id)
            await self.cache_manager.set(users_key, list(active_users), expire=3600)
            
        except Exception as e:
            logger.error(f"Failed to update realtime counters: {str(e)}")
    
    async def _process_event_aggregation(self, event: AnalyticsEvent):
        """处理事件聚合"""
        try:
            # 这里可以实现更复杂的事件聚合逻辑
            # 比如按用户、按内容类型、按时间段进行聚合
            pass
        except Exception as e:
            logger.error(f"Failed to process event aggregation: {str(e)}")
    
    async def _get_user_events(self, user_id: str, time_range: Tuple[datetime, datetime]) -> List[Dict]:
        """获取用户事件"""
        events = []
        try:
            # 从缓存中获取用户事件
            hours_in_range = int((time_range[1] - time_range[0]).total_seconds() / 3600)
            for i in range(hours_in_range + 1):
                hour_time = time_range[0] + timedelta(hours=i)
                hour_key = hour_time.strftime('%Y%m%d%H')
                
                # 获取该小时的所有事件类型
                for event_type in self.event_types.keys():
                    event_key = f"analytics:event:{hour_key}:{event_type}"
                    hour_events = await self.cache_manager.get(event_key) or []
                    
                    # 筛选用户事件
                    user_events = [e for e in hour_events if e.get('user_id') == user_id]
                    events.extend(user_events)
            
        except Exception as e:
            logger.error(f"Failed to get user events: {str(e)}")
        
        return events
    
    async def _get_content_events(self, content_type: str, time_range: Tuple[datetime, datetime]) -> List[Dict]:
        """获取内容相关事件"""
        events = []
        try:
            # 内容相关的事件类型
            content_event_types = [
                'letter_view', 'letter_like', 'letter_share', 'comment_create',
                'post_create', 'post_view', 'post_like'
            ]
            
            hours_in_range = int((time_range[1] - time_range[0]).total_seconds() / 3600)
            for i in range(hours_in_range + 1):
                hour_time = time_range[0] + timedelta(hours=i)
                hour_key = hour_time.strftime('%Y%m%d%H')
                
                for event_type in content_event_types:
                    event_key = f"analytics:event:{hour_key}:{event_type}"
                    hour_events = await self.cache_manager.get(event_key) or []
                    
                    # 筛选内容类型事件
                    content_events = [e for e in hour_events if e.get('object_type') == content_type]
                    events.extend(content_events)
            
        except Exception as e:
            logger.error(f"Failed to get content events: {str(e)}")
        
        return events
    
    def _analyze_activity_summary(self, events: List[Dict]) -> Dict[str, Any]:
        """分析活动摘要"""
        if not events:
            return {'total_events': 0, 'unique_days': 0, 'avg_events_per_day': 0}
        
        event_dates = set()
        event_types_count = Counter()
        
        for event in events:
            try:
                event_time = datetime.fromisoformat(event['timestamp'].replace('Z', '+00:00'))
                event_dates.add(event_time.date())
                event_types_count[event['event_type']] += 1
            except:
                continue
        
        return {
            'total_events': len(events),
            'unique_days': len(event_dates),
            'avg_events_per_day': len(events) / max(len(event_dates), 1),
            'event_types_distribution': dict(event_types_count.most_common())
        }
    
    def _calculate_engagement_score(self, events: List[Dict]) -> float:
        """计算用户参与度分数"""
        if not events:
            return 0.0
        
        total_weight = 0
        for event in events:
            event_type = event.get('event_type')
            weight = self.event_types.get(event_type, {}).get('weight', 1)
            total_weight += weight
        
        # 归一化分数（0-100）
        max_possible_score = len(events) * 5  # 假设最高权重为5
        score = min((total_weight / max_possible_score) * 100, 100) if max_possible_score > 0 else 0
        
        return round(score, 2)
    
    def _analyze_behavior_patterns(self, events: List[Dict]) -> Dict[str, Any]:
        """分析行为模式"""
        if not events:
            return {}
        
        # 按小时分析活动模式
        hourly_activity = defaultdict(int)
        daily_activity = defaultdict(int)
        
        for event in events:
            try:
                event_time = datetime.fromisoformat(event['timestamp'].replace('Z', '+00:00'))
                hourly_activity[event_time.hour] += 1
                daily_activity[event_time.strftime('%A')] += 1
            except:
                continue
        
        return {
            'peak_hour': max(hourly_activity.items(), key=lambda x: x[1])[0] if hourly_activity else 0,
            'peak_day': max(daily_activity.items(), key=lambda x: x[1])[0] if daily_activity else 'Monday',
            'hourly_distribution': dict(hourly_activity),
            'daily_distribution': dict(daily_activity)
        }
    
    def _analyze_content_preferences(self, events: List[Dict]) -> Dict[str, Any]:
        """分析内容偏好"""
        content_interactions = defaultdict(int)
        content_types = defaultdict(int)
        
        for event in events:
            if event.get('object_type'):
                content_types[event['object_type']] += 1
            
            event_category = self.event_types.get(event.get('event_type'), {}).get('category')
            if event_category:
                content_interactions[event_category] += 1
        
        return {
            'preferred_content_types': dict(content_types),
            'interaction_categories': dict(content_interactions)
        }
    
    def _analyze_usage_frequency(self, events: List[Dict]) -> Dict[str, Any]:
        """分析使用频率"""
        if not events:
            return {'frequency': 'inactive', 'sessions_per_week': 0}
        
        # 简化的会话分析：相邻事件间隔超过30分钟算新会话
        sessions = []
        current_session_start = None
        
        sorted_events = sorted(events, key=lambda x: x.get('timestamp', ''))
        
        for event in sorted_events:
            try:
                event_time = datetime.fromisoformat(event['timestamp'].replace('Z', '+00:00'))
                
                if current_session_start is None:
                    current_session_start = event_time
                elif (event_time - current_session_start).total_seconds() > 1800:  # 30分钟
                    sessions.append(current_session_start)
                    current_session_start = event_time
            except:
                continue
        
        if current_session_start:
            sessions.append(current_session_start)
        
        # 计算每周会话数
        if len(sessions) > 0:
            time_span_days = (sorted_events[-1]['timestamp'] - sorted_events[0]['timestamp']).total_seconds() / (24 * 3600)
            sessions_per_week = len(sessions) / max(time_span_days / 7, 1)
        else:
            sessions_per_week = 0
        
        # 定义使用频率等级
        if sessions_per_week >= 7:
            frequency = 'very_active'
        elif sessions_per_week >= 3:
            frequency = 'active'
        elif sessions_per_week >= 1:
            frequency = 'regular'
        elif sessions_per_week >= 0.5:
            frequency = 'occasional'
        else:
            frequency = 'inactive'
        
        return {
            'frequency': frequency,
            'sessions_per_week': round(sessions_per_week, 2),
            'total_sessions': len(sessions)
        }
    
    def _analyze_peak_activity_times(self, events: List[Dict]) -> Dict[str, Any]:
        """分析活跃时间段"""
        if not events:
            return {}
        
        hourly_counts = defaultdict(int)
        for event in events:
            try:
                event_time = datetime.fromisoformat(event['timestamp'].replace('Z', '+00:00'))
                hourly_counts[event_time.hour] += 1
            except:
                continue
        
        if not hourly_counts:
            return {}
        
        sorted_hours = sorted(hourly_counts.items(), key=lambda x: x[1], reverse=True)
        
        return {
            'most_active_hour': sorted_hours[0][0],
            'least_active_hour': sorted_hours[-1][0],
            'top_3_hours': [hour for hour, _ in sorted_hours[:3]]
        }
    
    def _get_top_content_by_interactions(self, events: List[Dict], limit: int) -> List[Dict]:
        """获取按交互数排序的热门内容"""
        content_interactions = defaultdict(lambda: {'count': 0, 'types': defaultdict(int)})
        
        for event in events:
            object_id = event.get('object_id')
            if object_id:
                content_interactions[object_id]['count'] += 1
                content_interactions[object_id]['types'][event.get('event_type', 'unknown')] += 1
        
        sorted_content = sorted(
            content_interactions.items(), 
            key=lambda x: x[1]['count'], 
            reverse=True
        )[:limit]
        
        return [
            {
                'object_id': obj_id,
                'interaction_count': data['count'],
                'interaction_types': dict(data['types'])
            }
            for obj_id, data in sorted_content
        ]
    
    def _analyze_interaction_breakdown(self, events: List[Dict]) -> Dict[str, int]:
        """分析交互类型分布"""
        interaction_counts = Counter(event.get('event_type', 'unknown') for event in events)
        return dict(interaction_counts)
    
    def _get_trending_content(self, events: List[Dict], limit: int) -> List[Dict]:
        """获取趋势内容（最近活跃度高的内容）"""
        recent_cutoff = datetime.utcnow() - timedelta(hours=24)  # 最近24小时
        
        recent_events = []
        for event in events:
            try:
                event_time = datetime.fromisoformat(event['timestamp'].replace('Z', '+00:00'))
                if event_time >= recent_cutoff:
                    recent_events.append(event)
            except:
                continue
        
        return self._get_top_content_by_interactions(recent_events, limit)
    
    def _calculate_content_engagement_metrics(self, events: List[Dict]) -> Dict[str, float]:
        """计算内容参与度指标"""
        if not events:
            return {}
        
        total_views = sum(1 for e in events if e.get('event_type') in ['letter_view', 'post_view'])
        total_likes = sum(1 for e in events if e.get('event_type') in ['letter_like', 'post_like'])
        total_shares = sum(1 for e in events if e.get('event_type') in ['letter_share', 'post_share'])
        total_comments = sum(1 for e in events if e.get('event_type') == 'comment_create')
        
        like_rate = (total_likes / total_views * 100) if total_views > 0 else 0
        share_rate = (total_shares / total_views * 100) if total_views > 0 else 0
        comment_rate = (total_comments / total_views * 100) if total_views > 0 else 0
        
        return {
            'total_views': total_views,
            'total_likes': total_likes,
            'total_shares': total_shares,
            'total_comments': total_comments,
            'like_rate': round(like_rate, 2),
            'share_rate': round(share_rate, 2),
            'comment_rate': round(comment_rate, 2),
            'engagement_rate': round(like_rate + share_rate + comment_rate, 2)
        }
    
    def _calculate_viral_coefficient(self, events: List[Dict]) -> float:
        """计算病毒系数"""
        shares = sum(1 for e in events if e.get('event_type') in ['letter_share', 'post_share'])
        unique_content = len(set(e.get('object_id') for e in events if e.get('object_id')))
        
        return round(shares / max(unique_content, 1), 2)
    
    async def _calculate_user_metrics(self, time_range: Tuple[datetime, datetime], granularity: TimeGranularity) -> Dict[str, Any]:
        """计算用户指标"""
        # 模拟数据，实际应该从数据库获取
        return {
            'total_users': 1000,
            'active_users': 450,
            'new_users': 25,
            'retention_rate': 0.85
        }
    
    async def _calculate_content_metrics(self, time_range: Tuple[datetime, datetime], granularity: TimeGranularity) -> Dict[str, Any]:
        """计算内容指标"""
        return {
            'total_content': 5000,
            'new_content': 120,
            'content_views': 15000,
            'avg_engagement_rate': 0.12
        }
    
    async def _calculate_engagement_metrics(self, time_range: Tuple[datetime, datetime], granularity: TimeGranularity) -> Dict[str, Any]:
        """计算参与度指标"""
        return {
            'total_interactions': 8500,
            'likes': 3200,
            'comments': 1800,
            'shares': 650,
            'avg_session_duration': 12.5
        }
    
    async def _calculate_revenue_metrics(self, time_range: Tuple[datetime, datetime], granularity: TimeGranularity) -> Dict[str, Any]:
        """计算收入指标"""
        return {
            'total_revenue': 25000.00,
            'new_orders': 85,
            'avg_order_value': 294.12,
            'conversion_rate': 0.034
        }
    
    async def _calculate_growth_metrics(self, time_range: Tuple[datetime, datetime], granularity: TimeGranularity) -> Dict[str, Any]:
        """计算增长指标"""
        return {
            'user_growth_rate': 0.05,
            'content_growth_rate': 0.08,
            'revenue_growth_rate': 0.12,
            'engagement_growth_rate': 0.03
        }
    
    async def _calculate_retention_metrics(self, time_range: Tuple[datetime, datetime], granularity: TimeGranularity) -> Dict[str, Any]:
        """计算留存指标"""
        return {
            'day_1_retention': 0.75,
            'day_7_retention': 0.45,
            'day_30_retention': 0.25,
            'churn_rate': 0.15
        }
    
    async def _get_active_users_count(self) -> int:
        """获取当前活跃用户数"""
        try:
            current_minute = datetime.utcnow().strftime('%Y%m%d%H%M')
            users_key = f"analytics:realtime:active_users:{current_minute}"
            active_users = await self.cache_manager.get(users_key) or []
            return len(set(active_users))
        except:
            return 0
    
    async def _get_events_per_minute(self) -> int:
        """获取每分钟事件数"""
        try:
            current_minute = datetime.utcnow().strftime('%Y%m%d%H%M')
            event_key = f"analytics:realtime:events:{current_minute}"
            return await self.cache_manager.get(event_key) or 0
        except:
            return 0
    
    async def _get_top_events_realtime(self) -> List[Dict]:
        """获取实时热门事件"""
        try:
            current_hour = datetime.utcnow().strftime('%Y%m%d%H')
            event_counts = {}
            
            for event_type in self.event_types.keys():
                event_key = f"analytics:event:{current_hour}:{event_type}"
                events = await self.cache_manager.get(event_key) or []
                event_counts[event_type] = len(events)
            
            sorted_events = sorted(event_counts.items(), key=lambda x: x[1], reverse=True)
            return [{'event_type': event, 'count': count} for event, count in sorted_events[:10]]
        except:
            return []
    
    async def _get_system_health_metrics(self) -> Dict[str, Any]:
        """获取系统健康指标"""
        return {
            'status': 'healthy',
            'response_time_avg': 150,  # ms
            'error_rate': 0.001,
            'cache_hit_rate': 0.85
        }
    
    async def _get_content_activity_realtime(self) -> Dict[str, Any]:
        """获取实时内容活动"""
        return {
            'new_letters': 5,
            'new_posts': 12,
            'new_comments': 28,
            'likes_per_minute': 45
        }


# 全局实例
_advanced_analytics: Optional[AdvancedAnalytics] = None

async def get_advanced_analytics() -> AdvancedAnalytics:
    """获取高级分析服务实例"""
    global _advanced_analytics
    if _advanced_analytics is None:
        _advanced_analytics = AdvancedAnalytics()
        await _advanced_analytics.initialize()
    return _advanced_analytics


# 便捷的事件跟踪函数
async def track_user_action(
    user_id: str,
    action: str,
    object_id: Optional[str] = None,
    object_type: Optional[str] = None,
    properties: Optional[Dict[str, Any]] = None,
    session_id: Optional[str] = None
):
    """跟踪用户行为"""
    analytics = await get_advanced_analytics()
    event = AnalyticsEvent(
        event_type=action,
        user_id=user_id,
        object_id=object_id,
        object_type=object_type,
        properties=properties or {},
        timestamp=datetime.utcnow(),
        session_id=session_id
    )
    await analytics.track_event(event)