from sqlalchemy import Column, String, DateTime, Text, Boolean, Integer
from sqlalchemy.sql import func
from sqlalchemy import ForeignKey
from sqlalchemy.orm import relationship
from app.core.database import Base

class ReadLog(Base):
    """信件阅读日志模型"""
    __tablename__ = "read_logs"
    
    # 主键
    id = Column(Integer, primary_key=True, autoincrement=True)
    
    # 关联信件
    letter_id = Column(String(20), ForeignKey("letters.id"), nullable=False, index=True, comment="信件ID")
    
    # 阅读者信息
    reader_ip = Column(String(45), comment="阅读者IP地址")  # 支持IPv6
    reader_user_agent = Column(Text, comment="浏览器User-Agent")
    reader_location = Column(String(200), comment="阅读地点（可选）")
    
    # 阅读详情
    read_duration = Column(Integer, comment="阅读时长（秒）")
    is_complete_read = Column(Boolean, default=True, comment="是否完整阅读")
    
    # 技术信息
    referer = Column(String(500), comment="来源页面")
    device_info = Column(Text, comment="设备信息JSON")
    
    # 时间戳
    read_at = Column(DateTime(timezone=True), server_default=func.now(), comment="阅读时间")
    
    # 关联关系
    letter = relationship("Letter", back_populates="read_logs")
    
    def __repr__(self):
        return f"<ReadLog(id={self.id}, letter_id={self.letter_id}, read_at={self.read_at})>"
    
    def to_dict(self):
        """转换为字典格式"""
        return {
            "id": self.id,
            "letter_id": self.letter_id,
            "reader_ip": self.reader_ip,
            "reader_user_agent": self.reader_user_agent,
            "reader_location": self.reader_location,
            "read_duration": self.read_duration,
            "is_complete_read": self.is_complete_read,
            "referer": self.referer,
            "device_info": self.device_info,
            "read_at": self.read_at.isoformat() if self.read_at else None
        }