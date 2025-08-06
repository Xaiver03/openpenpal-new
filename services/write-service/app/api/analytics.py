"""
高级数据分析API
支持用户行为分析、内容热度统计、业务KPI监控等功能
"""
import logging
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any, Tuple
from pydantic import BaseModel
from datetime import datetime, timedelta
from enum import Enum

from app.core.database import get_db, get_async_session
from app.utils.auth import get_current_user, get_current_user_optional
from app.utils.analytics_service import get_analytics_service
from app.utils.advanced_analytics import (
    get_advanced_analytics, AdvancedAnalytics,
    MetricType, TimeGranularity, AnalyticsEvent, track_user_action
)
from app.core.responses import success_response, error_response
from app.core.logger import get_logger
from app.schemas.analytics import (
    AnalyticsRequest, ReadingStatsResponse, LetterAnalyticsResponse,
    UserReadingBehaviorResponse, TrendAnalysisResponse, PopularityRankingResponse,
    RealtimeStatsResponse, ComparisonAnalysisRequest, ComparisonAnalysisResponse,
    AnalyticsExportRequest, TimeRangeEnum, SuccessResponse
)

router = APIRouter()
logger = get_logger(__name__)


# 高级分析API模型定义
class TrackEventRequest(BaseModel):
    """事件跟踪请求"""
    event_type: str
    object_id: Optional[str] = None
    object_type: Optional[str] = None
    properties: Optional[Dict[str, Any]] = None
    session_id: Optional[str] = None


class UserBehaviorAnalysisRequest(BaseModel):
    """用户行为分析请求"""
    user_id: Optional[str] = None  # 不指定则分析当前用户
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    include_details: Optional[bool] = False


class ContentPopularityRequest(BaseModel):
    """内容热度分析请求"""
    content_type: str
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    limit: Optional[int] = 100


class BusinessKPIRequest(BaseModel):
    """业务KPI请求"""
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    granularity: Optional[TimeGranularity] = TimeGranularity.DAY


# 高级分析API端点

@router.post("/track", summary="跟踪用户行为事件")
async def track_event(
    request: TrackEventRequest,
    current_user: dict = Depends(get_current_user),
    analytics: AdvancedAnalytics = Depends(get_advanced_analytics)
):
    """
    跟踪用户行为事件
    
    - **event_type**: 事件类型（如：letter_create, letter_view, letter_like等）
    - **object_id**: 相关对象ID（如信件ID、帖子ID等）
    - **object_type**: 对象类型（如letter, post等）
    - **properties**: 事件属性
    - **session_id**: 会话ID
    """
    try:
        user_id = current_user.get('user_id')
        
        event = AnalyticsEvent(
            event_type=request.event_type,
            user_id=user_id,
            object_id=request.object_id,
            object_type=request.object_type,
            properties=request.properties or {},
            timestamp=datetime.utcnow(),
            session_id=request.session_id
        )
        
        success = await analytics.track_event(event)
        
        if success:
            logger.info(f"Event tracked: {request.event_type} by user {user_id}")
            return success_response(
                data={"event_tracked": True},
                message="事件跟踪成功"
            )
        else:
            return error_response(message="事件跟踪失败")
            
    except Exception as e:
        logger.error(f"Track event error: {str(e)}")
        return error_response(message=f"跟踪事件失败: {str(e)}")


@router.post("/user-behavior", summary="获取用户行为分析")
async def get_user_behavior_analysis(
    request: UserBehaviorAnalysisRequest,
    current_user: dict = Depends(get_current_user),
    analytics: AdvancedAnalytics = Depends(get_advanced_analytics)
):
    """
    获取用户行为分析报告
    
    - **user_id**: 目标用户ID（不指定则分析当前用户）
    - **start_date**: 分析开始时间
    - **end_date**: 分析结束时间
    - **include_details**: 是否包含详细事件数据
    """
    try:
        # 权限检查：只能查看自己的数据，除非是管理员
        target_user_id = request.user_id or current_user.get('user_id')
        current_user_id = current_user.get('user_id')
        user_role = current_user.get('role', 'user')
        
        if target_user_id != current_user_id and user_role not in ['admin', 'super_admin']:
            raise HTTPException(status_code=403, detail="权限不足，只能查看自己的行为分析")
        
        # 设置时间范围
        time_range = None
        if request.start_date and request.end_date:
            time_range = (request.start_date, request.end_date)
        
        result = await analytics.get_user_behavior_analysis(
            user_id=target_user_id,
            time_range=time_range,
            include_details=request.include_details
        )
        
        logger.info(f"User behavior analysis requested by {current_user_id} for user {target_user_id}")
        
        return success_response(
            data=result,
            message="用户行为分析获取成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"User behavior analysis error: {str(e)}")
        return error_response(message=f"获取用户行为分析失败: {str(e)}")


@router.post("/content-popularity", summary="获取内容热度分析")
async def get_content_popularity_analysis(
    request: ContentPopularityRequest,
    current_user: dict = Depends(get_current_user),
    analytics: AdvancedAnalytics = Depends(get_advanced_analytics)
):
    """
    获取内容热度分析
    
    - **content_type**: 内容类型（letter, post, product等）
    - **start_date**: 分析开始时间
    - **end_date**: 分析结束时间  
    - **limit**: 返回数量限制
    """
    try:
        # 设置时间范围
        time_range = None
        if request.start_date and request.end_date:
            time_range = (request.start_date, request.end_date)
        
        result = await analytics.get_content_popularity_analysis(
            content_type=request.content_type,
            time_range=time_range,
            limit=request.limit
        )
        
        logger.info(f"Content popularity analysis requested by {current_user.get('user_id')} for type {request.content_type}")
        
        return success_response(
            data=result,
            message="内容热度分析获取成功"
        )
        
    except Exception as e:
        logger.error(f"Content popularity analysis error: {str(e)}")
        return error_response(message=f"获取内容热度分析失败: {str(e)}")


@router.post("/business-kpi", summary="获取业务KPI仪表板")
async def get_business_kpi_dashboard(
    request: BusinessKPIRequest,
    current_user: dict = Depends(get_current_user),
    analytics: AdvancedAnalytics = Depends(get_advanced_analytics)
):
    """
    获取业务KPI仪表板数据（管理员权限）
    
    - **start_date**: 分析开始时间
    - **end_date**: 分析结束时间
    - **granularity**: 时间粒度（hour, day, week, month等）
    """
    try:
        # 权限检查
        user_role = current_user.get('role', 'user')
        if user_role not in ['admin', 'super_admin', 'coordinator']:
            raise HTTPException(status_code=403, detail="权限不足，只有管理员可以查看业务KPI")
        
        # 设置时间范围
        time_range = None
        if request.start_date and request.end_date:
            time_range = (request.start_date, request.end_date)
        
        result = await analytics.get_business_kpi_dashboard(
            time_range=time_range,
            granularity=request.granularity
        )
        
        logger.info(f"Business KPI dashboard requested by {current_user.get('user_id')}")
        
        return success_response(
            data=result,
            message="业务KPI数据获取成功"
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Business KPI dashboard error: {str(e)}")
        return error_response(message=f"获取业务KPI失败: {str(e)}")


@router.get("/realtime-metrics", summary="获取实时指标")
async def get_realtime_metrics(
    current_user: dict = Depends(get_current_user),
    analytics: AdvancedAnalytics = Depends(get_advanced_analytics)
):
    """
    获取实时指标数据
    
    包括：
    - 活跃用户数
    - 每分钟事件数
    - 热门事件类型
    - 系统健康状态
    - 实时内容活动
    """
    try:
        result = await analytics.get_realtime_metrics()
        
        return success_response(
            data=result,
            message="实时指标获取成功"
        )
        
    except Exception as e:
        logger.error(f"Realtime metrics error: {str(e)}")
        return error_response(message=f"获取实时指标失败: {str(e)}")


@router.get("/metrics/types", summary="获取支持的指标类型")
async def get_supported_metric_types():
    """获取系统支持的所有指标类型"""
    try:
        metric_types = {
            "event_types": {
                "user_login": {"category": "auth", "weight": 1, "description": "用户登录"},
                "user_logout": {"category": "auth", "weight": 1, "description": "用户登出"},
                "letter_create": {"category": "content", "weight": 3, "description": "创建信件"},
                "letter_view": {"category": "engagement", "weight": 1, "description": "查看信件"},
                "letter_like": {"category": "engagement", "weight": 2, "description": "点赞信件"},
                "letter_share": {"category": "engagement", "weight": 3, "description": "分享信件"},
                "comment_create": {"category": "social", "weight": 2, "description": "创建评论"},
                "post_create": {"category": "social", "weight": 3, "description": "创建帖子"},
                "order_create": {"category": "commerce", "weight": 5, "description": "创建订单"},
                "search_query": {"category": "discovery", "weight": 1, "description": "搜索查询"},
                "file_upload": {"category": "content", "weight": 2, "description": "文件上传"},
                "notification_click": {"category": "engagement", "weight": 1, "description": "点击通知"}
            },
            "metric_categories": [
                "USER_BEHAVIOR",
                "CONTENT_POPULARITY", 
                "BUSINESS_KPI",
                "PERFORMANCE",
                "ENGAGEMENT"
            ],
            "time_granularities": [
                "HOUR",
                "DAY",
                "WEEK", 
                "MONTH",
                "QUARTER",
                "YEAR"
            ]
        }
        
        return success_response(
            data=metric_types,
            message="支持的指标类型获取成功"
        )
        
    except Exception as e:
        logger.error(f"Get metric types error: {str(e)}")
        return error_response(message=f"获取指标类型失败: {str(e)}")


# 便捷的事件跟踪端点
@router.post("/track/letter-action", summary="跟踪信件相关操作")
async def track_letter_action(
    letter_id: str,
    action: str,  # view, like, share, comment
    properties: Optional[Dict[str, Any]] = None,
    current_user: dict = Depends(get_current_user)
):
    """
    便捷的信件操作跟踪
    
    - **letter_id**: 信件ID
    - **action**: 操作类型（view, like, share, comment）
    - **properties**: 额外属性
    """
    try:
        user_id = current_user.get('user_id')
        
        await track_user_action(
            user_id=user_id,
            action=f"letter_{action}",
            object_id=letter_id,
            object_type="letter",
            properties=properties
        )
        
        logger.info(f"Letter action tracked: {action} on {letter_id} by user {user_id}")
        
        return success_response(
            data={"tracked": True},
            message=f"信件{action}操作跟踪成功"
        )
        
    except Exception as e:
        logger.error(f"Track letter action error: {str(e)}")
        return error_response(message=f"跟踪信件操作失败: {str(e)}")


@router.post("/track/user-session", summary="跟踪用户会话")
async def track_user_session(
    session_action: str,  # login, logout, activity
    session_id: Optional[str] = None,
    properties: Optional[Dict[str, Any]] = None,
    current_user: dict = Depends(get_current_user)
):
    """
    跟踪用户会话活动
    
    - **session_action**: 会话操作（login, logout, activity）
    - **session_id**: 会话ID
    - **properties**: 会话属性
    """
    try:
        user_id = current_user.get('user_id')
        
        await track_user_action(
            user_id=user_id,
            action=f"user_{session_action}",
            properties=properties,
            session_id=session_id
        )
        
        logger.info(f"User session tracked: {session_action} by user {user_id}")
        
        return success_response(
            data={"tracked": True},
            message=f"用户{session_action}活动跟踪成功"
        )
        
    except Exception as e:
        logger.error(f"Track user session error: {str(e)}")
        return error_response(message=f"跟踪用户会话失败: {str(e)}")


# 原有的基础分析API（保持兼容性）

@router.get("/reading-stats", response_model=SuccessResponse, summary="获取阅读统计")
async def get_reading_statistics(
    time_range: TimeRangeEnum = Query(default=TimeRangeEnum.DAY, description="时间范围"),
    start_date: Optional[datetime] = Query(None, description="开始时间（time_range为custom时必填）"),
    end_date: Optional[datetime] = Query(None, description="结束时间（time_range为custom时必填）"),
    letter_id: Optional[str] = Query(None, description="特定信件ID"),
    user_id: Optional[str] = Query(None, description="特定用户ID"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取阅读统计数据"""
    try:
        analytics_service = get_analytics_service()
        
        stats = analytics_service.get_reading_stats(
            db=db,
            time_range=time_range,
            start_date=start_date,
            end_date=end_date,
            letter_id=letter_id,
            user_id=user_id
        )
        
        return SuccessResponse(
            msg="获取阅读统计成功",
            data=stats
        )
        
    except Exception as e:
        logger.error(f"Failed to get reading statistics: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取阅读统计失败: {str(e)}"
        )


@router.get("/letter/{letter_id}/analytics", response_model=SuccessResponse, summary="获取信件详细分析")
async def get_letter_analytics(
    letter_id: str,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取单个信件的详细分析数据"""
    try:
        analytics_service = get_analytics_service()
        
        analytics = analytics_service.get_letter_analytics(
            db=db,
            letter_id=letter_id
        )
        
        return SuccessResponse(
            msg="获取信件分析成功",
            data=analytics
        )
        
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Failed to get letter analytics for {letter_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取信件分析失败: {str(e)}"
        )


@router.get("/user/{user_id}/behavior", response_model=SuccessResponse, summary="获取用户阅读行为分析")
async def get_user_reading_behavior(
    user_id: str,
    time_range: TimeRangeEnum = Query(default=TimeRangeEnum.MONTH, description="时间范围"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取用户阅读行为分析"""
    try:
        analytics_service = get_analytics_service()
        
        behavior = analytics_service.get_user_reading_behavior(
            db=db,
            user_id=user_id,
            time_range=time_range
        )
        
        return SuccessResponse(
            msg="获取用户行为分析成功",
            data=behavior
        )
        
    except Exception as e:
        logger.error(f"Failed to get user behavior for {user_id}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取用户行为分析失败: {str(e)}"
        )


@router.get("/trends", response_model=SuccessResponse, summary="获取趋势分析")
async def get_trend_analysis(
    time_range: TimeRangeEnum = Query(default=TimeRangeEnum.MONTH, description="时间范围"),
    start_date: Optional[datetime] = Query(None, description="开始时间"),
    end_date: Optional[datetime] = Query(None, description="结束时间"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取趋势分析数据"""
    try:
        analytics_service = get_analytics_service()
        
        trends = analytics_service.get_trend_analysis(
            db=db,
            time_range=time_range,
            start_date=start_date,
            end_date=end_date
        )
        
        return SuccessResponse(
            msg="获取趋势分析成功",
            data=trends
        )
        
    except Exception as e:
        logger.error(f"Failed to get trend analysis: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取趋势分析失败: {str(e)}"
        )


@router.get("/popular", response_model=SuccessResponse, summary="获取热门内容排行")
async def get_popular_content(
    limit: int = Query(default=10, ge=1, le=50, description="返回数量限制"),
    time_range: TimeRangeEnum = Query(default=TimeRangeEnum.WEEK, description="时间范围"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取热门内容排行"""
    try:
        analytics_service = get_analytics_service()
        
        popular = analytics_service.get_popular_content(
            db=db,
            limit=limit,
            time_range=time_range
        )
        
        return SuccessResponse(
            msg="获取热门内容成功",
            data=popular
        )
        
    except Exception as e:
        logger.error(f"Failed to get popular content: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取热门内容失败: {str(e)}"
        )


@router.get("/realtime", response_model=SuccessResponse, summary="获取实时统计")
async def get_realtime_statistics(
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取实时统计数据"""
    try:
        analytics_service = get_analytics_service()
        
        realtime_stats = analytics_service.get_realtime_stats(db=db)
        
        return SuccessResponse(
            msg="获取实时统计成功",
            data=realtime_stats
        )
        
    except Exception as e:
        logger.error(f"Failed to get realtime statistics: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取实时统计失败: {str(e)}"
        )


@router.post("/compare", response_model=SuccessResponse, summary="信件对比分析")
async def compare_letters(
    request: ComparisonAnalysisRequest,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """进行信件对比分析"""
    try:
        analytics_service = get_analytics_service()
        
        # 获取每个信件的分析数据
        comparison_data = {}
        for letter_id in request.letter_ids:
            letter_analytics = analytics_service.get_letter_analytics(db, letter_id)
            comparison_data[letter_id] = letter_analytics
        
        # 生成对比洞察
        insights = []
        recommendations = []
        
        # 找出阅读次数最多的信件
        max_reads_letter = max(comparison_data.items(), key=lambda x: x[1]['total_reads'])
        insights.append(f"信件 {max_reads_letter[0]} 获得了最多的阅读次数：{max_reads_letter[1]['total_reads']} 次")
        
        # 找出平均阅读时长最长的信件
        max_duration_letter = max(comparison_data.items(), key=lambda x: x[1]['avg_read_duration'])
        insights.append(f"信件 {max_duration_letter[0]} 有最长的平均阅读时长：{max_duration_letter[1]['avg_read_duration']} 秒")
        
        # 生成建议
        recommendations.append("考虑分析高阅读量信件的内容特点，应用到其他信件中")
        recommendations.append("关注读者的阅读时长，适当调整内容长度和结构")
        
        return SuccessResponse(
            msg="对比分析完成",
            data={
                "comparison_data": comparison_data,
                "insights": insights,
                "recommendations": recommendations
            }
        )
        
    except Exception as e:
        logger.error(f"Failed to compare letters: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"对比分析失败: {str(e)}"
        )


@router.get("/dashboard", response_model=SuccessResponse, summary="分析仪表板数据")
async def get_analytics_dashboard(
    time_range: TimeRangeEnum = Query(default=TimeRangeEnum.WEEK, description="时间范围"),
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """获取分析仪表板的综合数据"""
    try:
        analytics_service = get_analytics_service()
        
        # 获取多个维度的数据
        reading_stats = analytics_service.get_reading_stats(db, time_range)
        trend_analysis = analytics_service.get_trend_analysis(db, time_range)
        popular_content = analytics_service.get_popular_content(db, limit=5, time_range=time_range)
        realtime_stats = analytics_service.get_realtime_stats(db)
        
        dashboard_data = {
            "overview": {
                "total_reads": reading_stats["total_reads"],
                "unique_readers": reading_stats["unique_readers"],
                "avg_read_duration": reading_stats["avg_read_duration"],
                "complete_read_rate": reading_stats["complete_read_rate"]
            },
            "trends": trend_analysis,
            "popular_content": popular_content,
            "realtime": realtime_stats,
            "device_distribution": reading_stats["device_distribution"],
            "time_distribution": reading_stats["hourly_distribution"]
        }
        
        return SuccessResponse(
            msg="获取仪表板数据成功",
            data=dashboard_data
        )
        
    except Exception as e:
        logger.error(f"Failed to get dashboard data: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取仪表板数据失败: {str(e)}"
        )


@router.post("/export", response_model=SuccessResponse, summary="导出分析数据")
async def export_analytics_data(
    request: AnalyticsExportRequest,
    current_user: str = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """导出分析数据"""
    try:
        analytics_service = get_analytics_service()
        
        # 根据数据类型获取相应数据
        if request.data_type == "reading_stats":
            data = analytics_service.get_reading_stats(
                db, request.time_range, request.start_date, request.end_date
            )
        elif request.data_type == "trends":
            data = analytics_service.get_trend_analysis(
                db, request.time_range, request.start_date, request.end_date
            )
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="不支持的数据类型"
            )
        
        # TODO: 实现具体的导出功能（JSON/CSV/Excel）
        # 这里先返回JSON格式数据
        
        return SuccessResponse(
            msg="数据导出成功",
            data={
                "export_format": request.format,
                "data": data,
                "generated_at": datetime.utcnow().isoformat()
            }
        )
        
    except Exception as e:
        logger.error(f"Failed to export data: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"导出数据失败: {str(e)}"
        )


@router.get("/health", summary="分析服务健康检查")
async def analytics_health_check(db: Session = Depends(get_db)):
    """分析服务健康检查"""
    try:
        # 检查数据库连接
        db.execute("SELECT 1")
        
        # 检查缓存连接
        analytics_service = get_analytics_service()
        cache_status = "ok" if analytics_service.cache else "unavailable"
        
        # 检查高级分析服务
        try:
            advanced_analytics = await get_advanced_analytics()
            advanced_status = "ok" if advanced_analytics else "unavailable"
        except:
            advanced_status = "unavailable"
        
        return SuccessResponse(
            msg="分析服务运行正常",
            data={
                "status": "healthy",
                "database": "connected",
                "cache": cache_status,
                "advanced_analytics": advanced_status,
                "timestamp": datetime.utcnow().isoformat()
            }
        )
        
    except Exception as e:
        logger.error(f"Analytics health check failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="分析服务不可用"
        )