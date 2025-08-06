"""
简化的OCR引擎，用于开发和测试
"""
import base64
import json
import numpy as np
import time
from typing import Dict, Any, List, Optional

class MockOCREngine:
    """模拟OCR引擎"""
    
    def __init__(self):
        self.name = "mock"
        self.is_available = True
    
    def recognize(self, image: np.ndarray, language: str = 'zh') -> Dict:
        """识别文字（模拟）"""
        start_time = time.time()
        
        # 返回模拟数据
        return {
            'text': "这是一封模拟的手写信件内容。\n亲爱的朋友，\n很高兴给你写信...",
            'confidence': 0.95,
            'blocks': [
                {
                    'text': "这是一封模拟的手写信件内容。",
                    'confidence': 0.96,
                    'box': [[10, 10], [200, 10], [200, 50], [10, 50]]
                }
            ],
            'processing_time': round(time.time() - start_time, 3),
            'engine': self.name
        }
    
    def get_supported_languages(self) -> List[str]:
        """获取支持的语言"""
        return ['zh', 'en']
        
    def extract_text(self, image_path: str, lang: str = 'ch') -> Dict[str, Any]:
        """模拟文本提取"""
        # 返回模拟数据
        return {
            "text": "这是一封模拟的手写信件内容。\n亲爱的朋友，\n很高兴给你写信...",
            "confidence": 0.95,
            "regions": [
                {
                    "text": "这是一封模拟的手写信件内容。",
                    "confidence": 0.96,
                    "bbox": [[10, 10], [200, 10], [200, 50], [10, 50]]
                }
            ],
            "language": lang,
            "engine": self.engine_name
        }
    
    def process_batch(self, image_paths: List[str], lang: str = 'ch') -> List[Dict[str, Any]]:
        """批量处理"""
        return [self.extract_text(path, lang) for path in image_paths]
    
    def detect_language(self, image_path: str) -> str:
        """语言检测"""
        return 'ch'
    
    def enhance_image(self, image_path: str) -> str:
        """图像增强"""
        return image_path