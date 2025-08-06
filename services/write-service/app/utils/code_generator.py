import random
import string
import secrets
from typing import Set
from sqlalchemy.orm import Session
from app.models.letter import Letter

# 字符集配置 - 排除容易混淆的字符
CHARSET = string.ascii_uppercase + string.digits
EXCLUDED_CHARS = {'0', 'O', 'I', '1', 'L'}  # 排除容易混淆的字符
SAFE_CHARSET = ''.join(char for char in CHARSET if char not in EXCLUDED_CHARS)

class LetterCodeGenerator:
    """信件编号生成器"""
    
    PREFIX = "OP"  # 固定前缀
    CODE_LENGTH = 10  # 随机部分长度
    TOTAL_LENGTH = 12  # 总长度 (OP + 10位)
    
    @classmethod
    def generate_code(cls) -> str:
        """
        生成唯一信件编号
        格式: OP + 10位随机字符
        例如: OP2K5F7H9M3B
        """
        random_part = ''.join(secrets.choice(SAFE_CHARSET) for _ in range(cls.CODE_LENGTH))
        return f"{cls.PREFIX}{random_part}"
    
    @classmethod
    def generate_unique_code(cls, db: Session, max_attempts: int = 100) -> str:
        """
        生成唯一的信件编号（确保数据库中不重复）
        
        Args:
            db: 数据库会话
            max_attempts: 最大尝试次数
            
        Returns:
            str: 唯一的信件编号
            
        Raises:
            RuntimeError: 超过最大尝试次数仍未生成唯一编号
        """
        for attempt in range(max_attempts):
            code = cls.generate_code()
            
            # 检查数据库中是否已存在
            existing_letter = db.query(Letter).filter(Letter.id == code).first()
            if not existing_letter:
                return code
        
        raise RuntimeError(f"Failed to generate unique code after {max_attempts} attempts")
    
    @classmethod
    def is_valid_code(cls, code: str) -> bool:
        """
        验证信件编号格式是否正确
        
        Args:
            code: 要验证的编号
            
        Returns:
            bool: 是否为有效格式
        """
        if not code or len(code) != cls.TOTAL_LENGTH:
            return False
        
        if not code.startswith(cls.PREFIX):
            return False
        
        random_part = code[len(cls.PREFIX):]
        if len(random_part) != cls.CODE_LENGTH:
            return False
        
        # 检查随机部分是否只包含安全字符集
        return all(char in cls.SAFE_CHARSET for char in random_part)
    
    @classmethod
    def extract_info_from_code(cls, code: str) -> dict:
        """
        从编号中提取信息
        
        Args:
            code: 信件编号
            
        Returns:
            dict: 编号信息
        """
        if not cls.is_valid_code(code):
            return {
                "valid": False,
                "prefix": None,
                "random_part": None,
                "length": len(code) if code else 0
            }
        
        return {
            "valid": True,
            "prefix": cls.PREFIX,
            "random_part": code[len(cls.PREFIX):],
            "length": len(code),
            "total_length": cls.TOTAL_LENGTH
        }
    
    @classmethod
    def batch_generate_codes(cls, count: int, db: Session) -> list[str]:
        """
        批量生成唯一编号
        
        Args:
            count: 需要生成的数量
            db: 数据库会话
            
        Returns:
            list[str]: 唯一编号列表
        """
        codes = []
        generated_codes: Set[str] = set()
        
        for _ in range(count):
            attempts = 0
            max_attempts = 100
            
            while attempts < max_attempts:
                code = cls.generate_code()
                
                # 检查本批次中是否重复
                if code in generated_codes:
                    attempts += 1
                    continue
                
                # 检查数据库中是否存在
                existing_letter = db.query(Letter).filter(Letter.id == code).first()
                if existing_letter:
                    attempts += 1
                    continue
                
                # 找到唯一编号
                codes.append(code)
                generated_codes.add(code)
                break
            else:
                raise RuntimeError(f"Failed to generate unique code #{len(codes) + 1} after {max_attempts} attempts")
        
        return codes

# 便捷函数
def generate_letter_code() -> str:
    """生成信件编号的便捷函数"""
    return LetterCodeGenerator.generate_code()

def generate_unique_letter_code(db: Session) -> str:
    """生成唯一信件编号的便捷函数"""
    return LetterCodeGenerator.generate_unique_code(db)

def validate_letter_code(code: str) -> bool:
    """验证信件编号的便捷函数"""
    return LetterCodeGenerator.is_valid_code(code)