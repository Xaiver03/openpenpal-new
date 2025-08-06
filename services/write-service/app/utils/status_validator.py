from typing import Dict, List, Optional
from app.models.letter import LetterStatus
from fastapi import HTTPException, status

class LetterStatusValidator:
    """信件状态转换验证器"""
    
    # 定义合法的状态转换映射
    VALID_TRANSITIONS: Dict[LetterStatus, List[LetterStatus]] = {
        LetterStatus.DRAFT: [
            LetterStatus.GENERATED,  # 草稿 → 已生成二维码
        ],
        LetterStatus.GENERATED: [
            LetterStatus.COLLECTED,  # 已生成 → 已收取
            LetterStatus.DRAFT,      # 允许回退到草稿（取消发送）
        ],
        LetterStatus.COLLECTED: [
            LetterStatus.IN_TRANSIT,  # 已收取 → 投递中
            LetterStatus.FAILED,      # 已收取 → 投递失败（收取后发现问题）
        ],
        LetterStatus.IN_TRANSIT: [
            LetterStatus.DELIVERED,   # 投递中 → 已投递
            LetterStatus.FAILED,      # 投递中 → 投递失败
        ],
        LetterStatus.DELIVERED: [
            # 已投递是终态，通常不允许转换
            # 如果需要支持"重新投递"等场景，可以添加
        ],
        LetterStatus.FAILED: [
            LetterStatus.COLLECTED,   # 失败 → 重新收取（重试投递）
            LetterStatus.IN_TRANSIT,  # 失败 → 重新投递
        ],
    }
    
    # 角色权限：定义哪些角色可以进行哪些状态转换
    ROLE_PERMISSIONS: Dict[str, List[LetterStatus]] = {
        "user": [
            LetterStatus.GENERATED,  # 用户可以发送信件（草稿→生成）
        ],
        "courier": [
            LetterStatus.COLLECTED,   # 信使可以收取信件
            LetterStatus.IN_TRANSIT,  # 信使可以开始投递
            LetterStatus.DELIVERED,   # 信使可以标记已投递
            LetterStatus.FAILED,      # 信使可以标记失败
        ],
        "admin": [
            # 管理员可以进行所有状态转换
            LetterStatus.DRAFT,
            LetterStatus.GENERATED,
            LetterStatus.COLLECTED,
            LetterStatus.IN_TRANSIT,
            LetterStatus.DELIVERED,
            LetterStatus.FAILED,
        ],
    }
    
    @classmethod
    def validate_transition(
        cls, 
        current_status: LetterStatus, 
        new_status: LetterStatus,
        user_role: str = "user"
    ) -> bool:
        """
        验证状态转换是否合法
        
        Args:
            current_status: 当前状态
            new_status: 目标状态
            user_role: 用户角色
            
        Returns:
            bool: 是否允许转换
        """
        # 1. 检查状态转换逻辑是否合法
        if current_status == new_status:
            return True  # 同状态转换总是合法的
        
        allowed_transitions = cls.VALID_TRANSITIONS.get(current_status, [])
        if new_status not in allowed_transitions:
            return False
        
        # 2. 检查用户角色权限
        role_permissions = cls.ROLE_PERMISSIONS.get(user_role, [])
        if new_status not in role_permissions:
            return False
        
        return True
    
    @classmethod
    def validate_transition_or_raise(
        cls,
        current_status: LetterStatus,
        new_status: LetterStatus,
        user_role: str = "user"
    ):
        """
        验证状态转换，如果不合法则抛出异常
        
        Args:
            current_status: 当前状态
            new_status: 目标状态
            user_role: 用户角色
            
        Raises:
            HTTPException: 状态转换不合法时抛出
        """
        if not cls.validate_transition(current_status, new_status, user_role):
            # 构造详细的错误信息
            error_msg = cls._build_error_message(current_status, new_status, user_role)
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=error_msg
            )
    
    @classmethod
    def get_available_transitions(
        cls,
        current_status: LetterStatus,
        user_role: str = "user"
    ) -> List[LetterStatus]:
        """
        获取当前状态下可用的转换选项
        
        Args:
            current_status: 当前状态
            user_role: 用户角色
            
        Returns:
            List[LetterStatus]: 可用的状态转换列表
        """
        allowed_transitions = cls.VALID_TRANSITIONS.get(current_status, [])
        role_permissions = cls.ROLE_PERMISSIONS.get(user_role, [])
        
        # 返回既符合转换逻辑又符合角色权限的状态
        return [s for s in allowed_transitions if s in role_permissions]
    
    @classmethod
    def _build_error_message(
        cls,
        current_status: LetterStatus,
        new_status: LetterStatus,
        user_role: str
    ) -> str:
        """构造详细的错误信息"""
        status_names = {
            LetterStatus.DRAFT: "草稿",
            LetterStatus.GENERATED: "已生成",
            LetterStatus.COLLECTED: "已收取",
            LetterStatus.IN_TRANSIT: "投递中",
            LetterStatus.DELIVERED: "已投递",
            LetterStatus.FAILED: "投递失败",
        }
        
        current_name = status_names.get(current_status, current_status.value)
        new_name = status_names.get(new_status, new_status.value)
        
        # 检查是转换逻辑问题还是权限问题
        allowed_transitions = cls.VALID_TRANSITIONS.get(current_status, [])
        role_permissions = cls.ROLE_PERMISSIONS.get(user_role, [])
        
        if new_status not in allowed_transitions:
            available = [status_names.get(s, s.value) for s in allowed_transitions]
            return f"无法从'{current_name}'状态转换到'{new_name}'状态。可用转换: {', '.join(available)}"
        
        if new_status not in role_permissions:
            available = [status_names.get(s, s.value) for s in role_permissions]
            return f"角色'{user_role}'无权限转换到'{new_name}'状态。可用状态: {', '.join(available)}"
        
        return f"状态转换验证失败: {current_name} → {new_name}"

# 便捷函数
def validate_letter_status_transition(
    current_status: LetterStatus,
    new_status: LetterStatus,
    user_role: str = "user"
) -> bool:
    """验证信件状态转换的便捷函数"""
    return LetterStatusValidator.validate_transition(current_status, new_status, user_role)

def validate_letter_status_transition_or_raise(
    current_status: LetterStatus,
    new_status: LetterStatus,
    user_role: str = "user"
):
    """验证信件状态转换，失败时抛出异常的便捷函数"""
    LetterStatusValidator.validate_transition_or_raise(current_status, new_status, user_role)