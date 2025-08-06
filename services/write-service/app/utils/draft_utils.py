"""
草稿工具模块 - 处理草稿相关的业务逻辑
"""
import re
import asyncio
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple
from sqlalchemy.orm import Session
from app.models.draft import LetterDraft, DraftHistory
from app.utils.id_generator import generate_id


class DraftManager:
    """草稿管理器"""
    
    @staticmethod
    def calculate_content_stats(content: str) -> Tuple[int, int]:
        """
        计算内容统计信息
        
        Args:
            content: 文本内容
            
        Returns:
            Tuple[int, int]: (字数, 字符数)
        """
        if not content:
            return 0, 0
        
        # 字符数（包含空格和标点）
        character_count = len(content)
        
        # 字数统计（中英文混合）
        # 移除多余空格
        cleaned_content = re.sub(r'\s+', ' ', content.strip())
        
        # 分离中文和英文
        chinese_chars = re.findall(r'[\u4e00-\u9fff]', cleaned_content)
        english_words = re.findall(r'[a-zA-Z]+', cleaned_content)
        
        # 中文字符数 + 英文单词数
        word_count = len(chinese_chars) + len(english_words)
        
        return word_count, character_count
    
    @staticmethod
    def detect_content_changes(old_content: str, new_content: str) -> Dict[str, any]:
        """
        检测内容变化
        
        Args:
            old_content: 旧内容
            new_content: 新内容
            
        Returns:
            Dict: 变化信息
        """
        if not old_content:
            old_content = ""
        if not new_content:
            new_content = ""
        
        old_words, old_chars = DraftManager.calculate_content_stats(old_content)
        new_words, new_chars = DraftManager.calculate_content_stats(new_content)
        
        # 计算差异
        word_diff = new_words - old_words
        char_diff = new_chars - old_chars
        
        # 简单的变化类型检测
        change_type = "minor"
        if abs(word_diff) > 50:
            change_type = "major"
        elif abs(word_diff) > 10:
            change_type = "moderate"
        
        return {
            "word_diff": word_diff,
            "char_diff": char_diff,
            "change_type": change_type,
            "old_stats": {"words": old_words, "chars": old_chars},
            "new_stats": {"words": new_words, "chars": new_chars}
        }
    
    @staticmethod
    def generate_change_summary(old_content: str, new_content: str) -> str:
        """
        生成变更摘要
        
        Args:
            old_content: 旧内容
            new_content: 新内容
            
        Returns:
            str: 变更摘要
        """
        changes = DraftManager.detect_content_changes(old_content, new_content)
        
        summary_parts = []
        
        if changes["word_diff"] > 0:
            summary_parts.append(f"新增{changes['word_diff']}字")
        elif changes["word_diff"] < 0:
            summary_parts.append(f"删除{abs(changes['word_diff'])}字")
        
        if changes["change_type"] == "major":
            summary_parts.append("大幅修改")
        elif changes["change_type"] == "moderate":
            summary_parts.append("适度修改")
        else:
            summary_parts.append("微调")
        
        return "，".join(summary_parts) if summary_parts else "内容未变化"
    
    @staticmethod
    def should_create_history_backup(draft: LetterDraft, new_content: str) -> bool:
        """
        判断是否应该创建历史备份
        
        Args:
            draft: 当前草稿
            new_content: 新内容
            
        Returns:
            bool: 是否需要备份
        """
        if not draft.content and not new_content:
            return False
        
        changes = DraftManager.detect_content_changes(draft.content or "", new_content or "")
        
        # 主要变化需要备份
        if changes["change_type"] == "major":
            return True
        
        # 定期备份（每个版本）
        if draft.version % 5 == 0:  # 每5个版本备份一次
            return True
        
        # 时间间隔备份（超过1小时）
        if draft.updated_at and datetime.utcnow() - draft.updated_at > timedelta(hours=1):
            return True
        
        return False


class AutoSaveManager:
    """自动保存管理器"""
    
    def __init__(self):
        self._save_tasks: Dict[str, asyncio.Task] = {}
        self._save_intervals: Dict[str, int] = {}  # 用户自定义保存间隔
    
    def schedule_auto_save(
        self, 
        user_id: str, 
        draft_id: str, 
        content: str, 
        title: str = None,
        delay_seconds: int = 30
    ) -> None:
        """
        调度自动保存任务
        
        Args:
            user_id: 用户ID
            draft_id: 草稿ID
            content: 内容
            title: 标题
            delay_seconds: 延迟秒数
        """
        task_key = f"{user_id}:{draft_id}"
        
        # 取消之前的任务
        if task_key in self._save_tasks:
            self._save_tasks[task_key].cancel()
        
        # 创建新的延迟保存任务
        async def delayed_save():
            try:
                await asyncio.sleep(delay_seconds)
                # 这里应该调用实际的保存逻辑
                print(f"自动保存草稿 {draft_id} for user {user_id}")
                # await self._perform_save(user_id, draft_id, content, title)
            except asyncio.CancelledError:
                pass  # 任务被取消
            except Exception as e:
                print(f"自动保存失败: {e}")
        
        self._save_tasks[task_key] = asyncio.create_task(delayed_save())
    
    def cancel_auto_save(self, user_id: str, draft_id: str) -> None:
        """
        取消自动保存任务
        
        Args:
            user_id: 用户ID
            draft_id: 草稿ID
        """
        task_key = f"{user_id}:{draft_id}"
        if task_key in self._save_tasks:
            self._save_tasks[task_key].cancel()
            del self._save_tasks[task_key]
    
    def set_save_interval(self, user_id: str, interval_seconds: int) -> None:
        """
        设置用户的自动保存间隔
        
        Args:
            user_id: 用户ID
            interval_seconds: 间隔秒数
        """
        self._save_intervals[user_id] = max(10, min(300, interval_seconds))  # 限制在10-300秒之间
    
    def get_save_interval(self, user_id: str) -> int:
        """
        获取用户的自动保存间隔
        
        Args:
            user_id: 用户ID
            
        Returns:
            int: 间隔秒数
        """
        return self._save_intervals.get(user_id, 30)  # 默认30秒


class DraftCleanupService:
    """草稿清理服务"""
    
    @staticmethod
    def get_expired_drafts(db: Session, days: int = 30) -> List[LetterDraft]:
        """
        获取过期的草稿
        
        Args:
            db: 数据库会话
            days: 天数阈值
            
        Returns:
            List[LetterDraft]: 过期草稿列表
        """
        cutoff_date = datetime.utcnow() - timedelta(days=days)
        
        return db.query(LetterDraft).filter(
            LetterDraft.updated_at < cutoff_date,
            LetterDraft.is_discarded == True
        ).all()
    
    @staticmethod
    def cleanup_old_history(db: Session, days: int = 90) -> int:
        """
        清理旧的历史记录
        
        Args:
            db: 数据库会话
            days: 天数阈值
            
        Returns:
            int: 清理的记录数
        """
        cutoff_date = datetime.utcnow() - timedelta(days=days)
        
        # 保留每个草稿最新的几个版本
        subquery = db.query(DraftHistory.draft_id).distinct()
        
        deleted_count = 0
        for draft_id_row in subquery:
            draft_id = draft_id_row[0]
            
            # 获取该草稿的历史记录，按版本倒序
            histories = db.query(DraftHistory).filter(
                DraftHistory.draft_id == draft_id,
                DraftHistory.created_at < cutoff_date
            ).order_by(DraftHistory.version.desc()).all()
            
            # 保留最新的5个版本，删除其余的
            if len(histories) > 5:
                to_delete = histories[5:]
                for history in to_delete:
                    db.delete(history)
                    deleted_count += 1
        
        db.commit()
        return deleted_count


# 全局自动保存管理器实例
auto_save_manager = AutoSaveManager()