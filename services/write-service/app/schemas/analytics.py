"""
阅读分析相关的Pydantic模式
"""
from pydantic import BaseModel, Field
from typing import Optional, Dict, List, Any
from datetime import datetime
from enum import Enum


class TimeRangeEnum(str, Enum):
    """时间范围枚举"""
    HOUR = "hour"
    DAY = "day"
    WEEK = "week"
    MONTH = "month"
    QUARTER = "quarter"
    YEAR = "year"
    CUSTOM = "custom"


class AnalyticsRequest(BaseModel):
    """分析请求基础模型"""
    time_range: TimeRangeEnum = Field(default=TimeRangeEnum.DAY, description="时间范围")
    start_date: Optional[datetime] = Field(None, description="开始时间（time_range为custom时必填）")
    end_date: Optional[datetime] = Field(None, description="结束时间（time_range为custom时必填）")
    letter_id: Optional[str] = Field(None, description="特定信件ID（可选）")
    user_id: Optional[str] = Field(None, description="特定用户ID（可选）")


class ReadingStatsResponse(BaseModel):
    """阅读统计响应模型"""
    total_reads: int = Field(description="总阅读次数")
    unique_readers: int = Field(description="独立阅读者数")
    avg_read_duration: float = Field(description="平均阅读时长（秒）")
    complete_read_rate: float = Field(description="完整阅读率")
    device_distribution: Dict[str, int] = Field(description="设备类型分布")
    browser_distribution: Dict[str, int] = Field(description="浏览器分布")
    location_distribution: Dict[str, int] = Field(description="地理位置分布")
    hourly_distribution: Dict[str, int] = Field(description="小时分布")


class LetterAnalyticsResponse(BaseModel):
    """单个信件详细分析响应"""
    letter_id: str = Field(description="信件ID")
    letter_title: str = Field(description="信件标题")
    total_reads: int = Field(description="总阅读次数")
    unique_readers: int = Field(description="独立阅读者数")
    first_read_at: Optional[datetime] = Field(description="首次阅读时间")
    last_read_at: Optional[datetime] = Field(description="最后阅读时间")
    avg_read_duration: float = Field(description="平均阅读时长")
    max_read_duration: int = Field(description="最长阅读时长")
    complete_reads: int = Field(description="完整阅读次数")
    device_stats: Dict[str, int] = Field(description="设备统计")
    browser_stats: Dict[str, int] = Field(description="浏览器统计")
    time_distribution: List[Dict[str, Any]] = Field(description="时间分布")


class UserReadingBehaviorResponse(BaseModel):
    """用户阅读行为分析响应"""
    user_id: str = Field(description="用户ID")
    total_letters_sent: int = Field(description="发送信件总数")
    total_reads_received: int = Field(description="信件被阅读总次数")
    avg_reads_per_letter: float = Field(description="每封信平均阅读次数")
    most_read_letter: Optional[Dict[str, Any]] = Field(description="最受欢迎信件")
    reading_time_stats: Dict[str, Any] = Field(description="阅读时间统计")
    reader_demographics: Dict[str, Any] = Field(description="读者画像")


class TrendAnalysisResponse(BaseModel):
    """趋势分析响应模型"""
    time_series: List[Dict[str, Any]] = Field(description="时间序列数据")
    growth_rate: float = Field(description="增长率")
    peak_hours: List[int] = Field(description="高峰时段")
    peak_days: List[str] = Field(description="高峰日期")
    seasonal_patterns: Dict[str, Any] = Field(description="季节性模式")


class PopularityRankingResponse(BaseModel):
    """热门排行响应模型"""
    top_letters: List[Dict[str, Any]] = Field(description="热门信件排行")
    top_users: List[Dict[str, Any]] = Field(description="活跃用户排行")
    trending_topics: List[Dict[str, Any]] = Field(description="热门话题")


class RealtimeStatsResponse(BaseModel):
    """实时统计响应模型"""
    current_online_readers: int = Field(description="当前在线阅读者")
    reads_last_hour: int = Field(description="过去1小时阅读次数")
    reads_today: int = Field(description="今日阅读次数")
    active_letters: List[Dict[str, Any]] = Field(description="活跃信件列表")
    live_events: List[Dict[str, Any]] = Field(description="实时事件流")


class ReaderGeographyResponse(BaseModel):
    """读者地理分布响应模型"""
    country_distribution: Dict[str, int] = Field(description="国家分布")
    city_distribution: Dict[str, int] = Field(description="城市分布")
    timezone_distribution: Dict[str, int] = Field(description="时区分布")
    heatmap_data: List[Dict[str, Any]] = Field(description="热力图数据")


class ContentAnalysisResponse(BaseModel):
    """内容分析响应模型"""
    word_frequency: Dict[str, int] = Field(description="词频统计")
    sentiment_analysis: Dict[str, Any] = Field(description="情感分析")
    topic_distribution: Dict[str, int] = Field(description="主题分布")
    readability_score: float = Field(description="可读性评分")
    engagement_metrics: Dict[str, Any] = Field(description="参与度指标")


class ComparisonAnalysisRequest(BaseModel):
    """对比分析请求模型"""
    letter_ids: List[str] = Field(description="要对比的信件ID列表", min_items=2, max_items=10)
    metrics: List[str] = Field(description="对比指标列表", default=["reads", "duration", "completion_rate"])


class ComparisonAnalysisResponse(BaseModel):
    """对比分析响应模型"""
    comparison_data: Dict[str, Dict[str, Any]] = Field(description="对比数据")
    insights: List[str] = Field(description="分析洞察")
    recommendations: List[str] = Field(description="改进建议")


class AnalyticsExportRequest(BaseModel):
    """分析数据导出请求"""
    data_type: str = Field(description="数据类型", pattern="^(reading_stats|user_behavior|trends|detailed_logs)$")
    format: str = Field(description="导出格式", default="json", pattern="^(json|csv|excel)$")
    include_raw_data: bool = Field(description="是否包含原始数据", default=False)
    time_range: TimeRangeEnum = Field(default=TimeRangeEnum.MONTH, description="时间范围")
    start_date: Optional[datetime] = Field(None, description="开始时间")
    end_date: Optional[datetime] = Field(None, description="结束时间")


class SuccessResponse(BaseModel):
    """统一成功响应格式"""
    code: int = Field(default=0, description="响应码，0表示成功")
    msg: str = Field(default="success", description="响应消息")
    data: Any = Field(description="响应数据")
    timestamp: datetime = Field(default_factory=datetime.utcnow, description="响应时间")


# 便捷类型别名
ReadingAnalytics = ReadingStatsResponse
LetterAnalytics = LetterAnalyticsResponse
UserAnalytics = UserReadingBehaviorResponse
TrendAnalytics = TrendAnalysisResponse