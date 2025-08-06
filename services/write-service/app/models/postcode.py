"""
OpenPenPal Postcode 数据模型
基于SQLAlchemy的数据库模型定义
"""

from sqlalchemy import Column, String, Integer, Text, TIMESTAMP, Boolean, DECIMAL, ARRAY
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from sqlalchemy import ForeignKey, CheckConstraint, UniqueConstraint, ForeignKeyConstraint

from app.core.database import Base


class PostcodeSchool(Base):
    """学校站点表"""
    __tablename__ = "postcode_schools"

    id = Column(UUID(as_uuid=True), primary_key=True, server_default=func.gen_random_uuid())
    code = Column(String(2), unique=True, nullable=False, index=True)  # 2位学校编码
    name = Column(String(100), nullable=False)
    full_name = Column(String(200), nullable=False)
    status = Column(String(20), default='active')
    managed_by = Column(String(100))  # 四级信使ID
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now())

    # 关系
    areas = relationship("PostcodeArea", back_populates="school", cascade="all, delete-orphan")

    __table_args__ = (
        CheckConstraint("status IN ('active', 'inactive')", name='check_school_status'),
    )


class PostcodeArea(Base):
    """片区表"""
    __tablename__ = "postcode_areas"

    id = Column(UUID(as_uuid=True), primary_key=True, server_default=func.gen_random_uuid())
    school_code = Column(String(2), ForeignKey('postcode_schools.code', ondelete='CASCADE'), nullable=False)
    code = Column(String(1), nullable=False)  # 1位片区编码
    name = Column(String(100), nullable=False)
    description = Column(Text)
    status = Column(String(20), default='active')
    managed_by = Column(String(100))  # 三级信使ID
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now())

    # 关系
    school = relationship("PostcodeSchool", back_populates="areas")
    buildings = relationship("PostcodeBuilding", back_populates="area", cascade="all, delete-orphan")

    __table_args__ = (
        UniqueConstraint('school_code', 'code', name='uq_area_school_code'),
        CheckConstraint("status IN ('active', 'inactive')", name='check_area_status'),
    )


class PostcodeBuilding(Base):
    """楼栋表"""
    __tablename__ = "postcode_buildings"

    id = Column(UUID(as_uuid=True), primary_key=True, server_default=func.gen_random_uuid())
    school_code = Column(String(2), nullable=False)
    area_code = Column(String(1), nullable=False)
    code = Column(String(1), nullable=False)  # 1位楼栋编码
    name = Column(String(100), nullable=False)
    type = Column(String(20), default='dormitory')
    floors = Column(Integer)
    status = Column(String(20), default='active')
    managed_by = Column(String(100))  # 二级信使ID
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now())

    # 关系
    area = relationship("PostcodeArea", back_populates="buildings")
    rooms = relationship("PostcodeRoom", back_populates="building", cascade="all, delete-orphan")

    __table_args__ = (
        ForeignKeyConstraint(['school_code', 'area_code'], ['postcode_areas.school_code', 'postcode_areas.code'], ondelete='CASCADE'),
        UniqueConstraint('school_code', 'area_code', 'code', name='uq_building_codes'),
        CheckConstraint("status IN ('active', 'inactive')", name='check_building_status'),
        CheckConstraint("type IN ('dormitory', 'teaching', 'office', 'other')", name='check_building_type'),
    )


class PostcodeRoom(Base):
    """房间表"""
    __tablename__ = "postcode_rooms"

    id = Column(UUID(as_uuid=True), primary_key=True, server_default=func.gen_random_uuid())
    school_code = Column(String(2), nullable=False)
    area_code = Column(String(1), nullable=False)
    building_code = Column(String(1), nullable=False)
    code = Column(String(2), nullable=False)  # 2位房间编码
    name = Column(String(100), nullable=False)
    type = Column(String(20), default='dormitory')
    capacity = Column(Integer)
    floor = Column(Integer)
    # PostgreSQL computed column - 6位完整编码
    full_postcode = Column(String(6), nullable=False, index=True)
    status = Column(String(20), default='active')
    managed_by = Column(String(100))  # 一级信使ID
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now())

    # 关系
    building = relationship("PostcodeBuilding", back_populates="rooms")

    __table_args__ = (
        ForeignKeyConstraint(['school_code', 'area_code', 'building_code'], 
                  ['postcode_buildings.school_code', 'postcode_buildings.area_code', 'postcode_buildings.code'], 
                  ondelete='CASCADE'),
        UniqueConstraint('school_code', 'area_code', 'building_code', 'code', name='uq_room_codes'),
        CheckConstraint("status IN ('active', 'inactive')", name='check_room_status'),
        CheckConstraint("type IN ('dormitory', 'classroom', 'office', 'other')", name='check_room_type'),
    )

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        # 自动生成完整编码
        if self.school_code and self.area_code and self.building_code and self.code:
            self.full_postcode = f"{self.school_code}{self.area_code}{self.building_code}{self.code}"


class PostcodeCourierPermission(Base):
    """信使Postcode权限表"""
    __tablename__ = "postcode_courier_permissions"

    id = Column(UUID(as_uuid=True), primary_key=True, server_default=func.gen_random_uuid())
    courier_id = Column(String(100), unique=True, nullable=False, index=True)
    level = Column(Integer, nullable=False)
    prefix_patterns = Column(ARRAY(String), nullable=False)  # 权限前缀数组
    can_manage = Column(Boolean, default=False)
    can_create = Column(Boolean, default=False)
    can_review = Column(Boolean, default=False)
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now())

    __table_args__ = (
        CheckConstraint("level IN (1, 2, 3, 4)", name='check_courier_level'),
    )


class PostcodeFeedback(Base):
    """地址反馈表"""
    __tablename__ = "postcode_feedbacks"

    id = Column(UUID(as_uuid=True), primary_key=True, server_default=func.gen_random_uuid())
    type = Column(String(20), nullable=False)
    postcode = Column(String(6))
    description = Column(Text, nullable=False)
    suggested_school_code = Column(String(2))
    suggested_area_code = Column(String(1))
    suggested_building_code = Column(String(1))
    suggested_room_code = Column(String(2))
    suggested_name = Column(String(200))
    submitted_by = Column(String(100), nullable=False, index=True)
    submitter_type = Column(String(20), default='user')
    status = Column(String(20), default='pending', index=True)
    reviewed_by = Column(String(100))
    review_notes = Column(Text)
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now())

    __table_args__ = (
        CheckConstraint("type IN ('new_address', 'error_report', 'delivery_failed')", name='check_feedback_type'),
        CheckConstraint("submitter_type IN ('user', 'courier')", name='check_submitter_type'),
        CheckConstraint("status IN ('pending', 'approved', 'rejected')", name='check_feedback_status'),
    )


class PostcodeStats(Base):
    """Postcode使用统计表"""
    __tablename__ = "postcode_stats"

    id = Column(UUID(as_uuid=True), primary_key=True, server_default=func.gen_random_uuid())
    postcode = Column(String(6), unique=True, nullable=False, index=True)
    delivery_count = Column(Integer, default=0)
    error_count = Column(Integer, default=0)
    last_used = Column(TIMESTAMP(timezone=True), server_default=func.now())
    popularity_score = Column(DECIMAL(5, 2), default=0.0, index=True)
    created_at = Column(TIMESTAMP(timezone=True), server_default=func.now())
    updated_at = Column(TIMESTAMP(timezone=True), server_default=func.now(), onupdate=func.now())