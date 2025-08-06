from sqlalchemy import Column, String, Text, DateTime, Boolean, Integer
from sqlalchemy.sql import func
from app.core.database import Base


class LetterDraft(Base):
    """信件草稿模型"""
    __tablename__ = "letter_drafts"

    id = Column(String(20), primary_key=True, index=True)
    user_id = Column(String(20), nullable=False, index=True)
    
    # 草稿内容
    title = Column(String(200), nullable=True)  # 草稿标题可以为空
    content = Column(Text, nullable=True)  # 草稿内容可以为空
    
    # 收件人信息（可选）
    recipient_id = Column(String(20), nullable=True)
    recipient_type = Column(String(20), nullable=True)  # friend/stranger/group
    
    # 信纸和信封样式
    paper_style = Column(String(50), default="classic")
    envelope_style = Column(String(50), default="simple")
    
    # 草稿元数据
    draft_type = Column(String(20), default="letter")  # letter/reply
    parent_letter_id = Column(String(20), nullable=True)  # 如果是回复草稿
    
    # 版本控制
    version = Column(Integer, default=1)
    word_count = Column(Integer, default=0)
    character_count = Column(Integer, default=0)
    
    # 自动保存相关
    auto_save_enabled = Column(Boolean, default=True)
    last_edit_time = Column(DateTime, default=func.now(), onupdate=func.now())
    
    # 时间戳
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    
    # 草稿状态
    is_active = Column(Boolean, default=True)  # 是否为活跃草稿
    is_discarded = Column(Boolean, default=False)  # 是否已丢弃
    
    def __repr__(self):
        return f"<LetterDraft(id={self.id}, user_id={self.user_id}, title={self.title})>"


class DraftHistory(Base):
    """草稿历史记录模型 - 用于恢复和版本管理"""
    __tablename__ = "draft_history"
    
    id = Column(String(20), primary_key=True, index=True)
    draft_id = Column(String(20), nullable=False, index=True)
    user_id = Column(String(20), nullable=False, index=True)
    
    # 历史版本内容
    title = Column(String(200), nullable=True)
    content = Column(Text, nullable=True)
    version = Column(Integer, nullable=False)
    
    # 变更信息
    change_summary = Column(String(500), nullable=True)  # 变更摘要
    change_type = Column(String(20), default="auto_save")  # auto_save/manual_save/version_backup
    
    # 统计信息
    word_count = Column(Integer, default=0)
    character_count = Column(Integer, default=0)
    
    # 时间戳
    created_at = Column(DateTime, default=func.now())
    
    def __repr__(self):
        return f"<DraftHistory(id={self.id}, draft_id={self.draft_id}, version={self.version})>"