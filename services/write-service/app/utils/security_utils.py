"""
安全工具模块 - 提供各种安全相关的工具函数
"""
import re
import html
import urllib.parse
from typing import List, Optional
import bleach
from app.middleware.error_handler import InputSanitizer


class XSSProtection:
    """XSS防护工具类"""
    
    # 允许的HTML标签（用于富文本编辑器）
    ALLOWED_TAGS = [
        'p', 'br', 'strong', 'b', 'em', 'i', 'u', 's', 'del',
        'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
        'ul', 'ol', 'li', 'blockquote',
        'a', 'img',
        'code', 'pre'
    ]
    
    # 允许的HTML属性
    ALLOWED_ATTRIBUTES = {
        'a': ['href', 'title'],
        'img': ['src', 'alt', 'title', 'width', 'height'],
        'blockquote': ['cite'],
        '*': ['class']
    }
    
    # 允许的协议
    ALLOWED_PROTOCOLS = ['http', 'https', 'mailto']
    
    @classmethod
    def clean_html(cls, content: str, allow_tags: bool = False) -> str:
        """
        清理HTML内容，防止XSS攻击
        
        Args:
            content: 要清理的内容
            allow_tags: 是否允许安全的HTML标签
            
        Returns:
            str: 清理后的内容
        """
        if not content:
            return content
        
        if allow_tags:
            # 使用bleach库清理HTML，保留安全标签
            try:
                import bleach
                return bleach.clean(
                    content,
                    tags=cls.ALLOWED_TAGS,
                    attributes=cls.ALLOWED_ATTRIBUTES,
                    protocols=cls.ALLOWED_PROTOCOLS,
                    strip=True
                )
            except ImportError:
                # 如果没有bleach库，回退到简单清理
                return cls._simple_html_escape(content)
        else:
            # 完全转义HTML
            return cls._simple_html_escape(content)
    
    @staticmethod
    def _simple_html_escape(content: str) -> str:
        """简单的HTML转义"""
        return html.escape(content, quote=True)
    
    @staticmethod
    def clean_url(url: str) -> str:
        """清理URL，防止javascript:等危险协议"""
        if not url:
            return url
        
        # 移除危险的协议
        dangerous_protocols = [
            'javascript:', 'vbscript:', 'data:', 'file:',
            'ftp:', 'gopher:', 'ldap:', 'mailto:'
        ]
        
        url_lower = url.lower().strip()
        for protocol in dangerous_protocols:
            if url_lower.startswith(protocol):
                return '#'  # 替换为安全的锚点
        
        # URL编码清理
        try:
            parsed = urllib.parse.urlparse(url)
            if parsed.scheme and parsed.scheme not in ['http', 'https']:
                return '#'
            return url
        except:
            return '#'
    
    @staticmethod
    def validate_file_upload(filename: str, allowed_extensions: List[str] = None) -> bool:
        """
        验证上传文件的安全性
        
        Args:
            filename: 文件名
            allowed_extensions: 允许的文件扩展名列表
            
        Returns:
            bool: 是否安全
        """
        if not filename:
            return False
        
        # 默认允许的文件类型
        if allowed_extensions is None:
            allowed_extensions = ['.jpg', '.jpeg', '.png', '.gif', '.pdf', '.txt', '.docx']
        
        # 检查文件扩展名
        filename_lower = filename.lower()
        if not any(filename_lower.endswith(ext) for ext in allowed_extensions):
            return False
        
        # 检查危险字符
        dangerous_chars = ['..', '/', '\\', ':', '*', '?', '"', '<', '>', '|']
        if any(char in filename for char in dangerous_chars):
            return False
        
        # 检查文件名长度
        if len(filename) > 255:
            return False
        
        return True


class ContentFilter:
    """内容过滤器 - 检测和过滤不当内容"""
    
    # 敏感词库（示例，实际应用中应该从配置文件或数据库加载）
    SENSITIVE_WORDS = [
        # 这里应该包含实际的敏感词列表
        # 为了示例，这里只放几个占位符
        '测试敏感词1', '测试敏感词2'
    ]
    
    @classmethod
    def filter_sensitive_words(cls, content: str, replace_char: str = '*') -> str:
        """
        过滤敏感词
        
        Args:
            content: 要过滤的内容
            replace_char: 替换字符
            
        Returns:
            str: 过滤后的内容
        """
        if not content:
            return content
        
        filtered_content = content
        for word in cls.SENSITIVE_WORDS:
            if word in filtered_content:
                replacement = replace_char * len(word)
                filtered_content = filtered_content.replace(word, replacement)
        
        return filtered_content
    
    @staticmethod
    def detect_spam(content: str) -> bool:
        """
        简单的垃圾内容检测
        
        Args:
            content: 要检测的内容
            
        Returns:
            bool: 是否为垃圾内容
        """
        if not content:
            return False
        
        # 检测重复字符
        if re.search(r'(.)\1{10,}', content):  # 同一字符重复超过10次
            return True
        
        # 检测过多的大写字母
        uppercase_ratio = sum(1 for c in content if c.isupper()) / len(content)
        if uppercase_ratio > 0.8 and len(content) > 20:
            return True
        
        # 检测过多的特殊字符
        special_chars = sum(1 for c in content if not c.isalnum() and not c.isspace())
        if special_chars / len(content) > 0.5:
            return True
        
        return False


class ValidationUtils:
    """验证工具类"""
    
    @staticmethod
    def validate_email(email: str) -> bool:
        """验证邮箱格式"""
        if not email:
            return False
        
        pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        return re.match(pattern, email) is not None
    
    @staticmethod
    def validate_phone(phone: str) -> bool:
        """验证手机号格式（中国大陆）"""
        if not phone:
            return False
        
        pattern = r'^1[3-9]\d{9}$'
        return re.match(pattern, phone) is not None
    
    @staticmethod
    def validate_password_strength(password: str) -> dict:
        """
        验证密码强度
        
        Returns:
            dict: 包含强度信息的字典
        """
        if not password:
            return {"valid": False, "message": "密码不能为空"}
        
        issues = []
        
        if len(password) < 8:
            issues.append("密码长度至少8位")
        
        if not re.search(r'[a-z]', password):
            issues.append("需要包含小写字母")
        
        if not re.search(r'[A-Z]', password):
            issues.append("需要包含大写字母")
        
        if not re.search(r'\d', password):
            issues.append("需要包含数字")
        
        if not re.search(r'[!@#$%^&*(),.?":{}|<>]', password):
            issues.append("需要包含特殊字符")
        
        return {
            "valid": len(issues) == 0,
            "message": "密码强度符合要求" if len(issues) == 0 else "；".join(issues)
        }


def secure_content_processing(content: str, content_type: str = "text") -> str:
    """
    安全的内容处理函数
    
    Args:
        content: 要处理的内容
        content_type: 内容类型 (text, html, url)
        
    Returns:
        str: 处理后的安全内容
    """
    if not content:
        return content
    
    # 1. 基础输入清理
    content = InputSanitizer.clean_user_input(content)
    
    # 2. 根据内容类型进行特定处理
    if content_type == "html":
        content = XSSProtection.clean_html(content, allow_tags=True)
    elif content_type == "text":
        content = XSSProtection.clean_html(content, allow_tags=False)
    elif content_type == "url":
        content = XSSProtection.clean_url(content)
    
    # 3. 敏感词过滤
    content = ContentFilter.filter_sensitive_words(content)
    
    # 4. 垃圾内容检测
    if ContentFilter.detect_spam(content):
        raise ValueError("检测到垃圾内容，请重新编辑")
    
    return content