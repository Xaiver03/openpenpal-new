import logging
from typing import Dict, List, Optional, Tuple, Any
import os
import time
from abc import ABC, abstractmethod

try:
    import cv2
    CV2_AVAILABLE = True
except ImportError:
    CV2_AVAILABLE = False
    logging.warning("OpenCV (cv2) 不可用")

# numpy是必需的，单独导入
try:
    import numpy as np
except ImportError:
    logging.error("NumPy是必需的依赖，请安装: pip install numpy")
    raise

# OCR引擎导入
try:
    import pytesseract
    TESSERACT_AVAILABLE = True
except ImportError:
    TESSERACT_AVAILABLE = False
    logging.warning("Tesseract不可用")

try:
    from paddleocr import PaddleOCR
    PADDLE_AVAILABLE = True
except ImportError:
    PADDLE_AVAILABLE = False
    logging.warning("PaddleOCR不可用")

try:
    import easyocr
    EASYOCR_AVAILABLE = True
except ImportError:
    EASYOCR_AVAILABLE = False
    logging.warning("EasyOCR不可用")

# 根据可用性导入图像处理器
if CV2_AVAILABLE:
    from app.services.image_processor import ImagePreprocessor, HandwritingPreprocessor
else:
    from app.services.image_processor_mock import ImagePreprocessor, HandwritingPreprocessor

# 如果没有OCR库可用，导入模拟引擎
if not any([TESSERACT_AVAILABLE, PADDLE_AVAILABLE, EASYOCR_AVAILABLE]):
    from app.services.ocr_engine_mock import MockOCREngine
    MOCK_ENGINE = True
else:
    MOCK_ENGINE = False

logger = logging.getLogger(__name__)


class OCREngineBase(ABC):
    """OCR引擎基类"""
    
    def __init__(self, name: str):
        self.name = name
        self.is_available = False
    
    @abstractmethod
    def recognize(self, image: np.ndarray, language: str = 'zh') -> Dict:
        """识别文字"""
        pass
    
    @abstractmethod
    def get_supported_languages(self) -> List[str]:
        """获取支持的语言列表"""
        pass


class TesseractEngine(OCREngineBase):
    """Tesseract OCR引擎"""
    
    def __init__(self):
        super().__init__('tesseract')
        self.is_available = TESSERACT_AVAILABLE
        
        if self.is_available:
            # 语言映射
            self.lang_map = {
                'zh': 'chi_sim',
                'en': 'eng',
                'ja': 'jpn'
            }
            
            # 配置参数
            self.config = '--oem 3 --psm 6 -c tessedit_char_whitelist=0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz一二三四五六七八九十百千万亿'
    
    def recognize(self, image: np.ndarray, language: str = 'zh') -> Dict:
        """使用Tesseract识别文字"""
        if not self.is_available:
            raise RuntimeError("Tesseract不可用")
        
        start_time = time.time()
        
        try:
            # 语言转换
            lang = self.lang_map.get(language, 'chi_sim')
            
            # 文本识别
            text = pytesseract.image_to_string(image, lang=lang, config=self.config)
            
            # 获取详细信息
            data = pytesseract.image_to_data(image, lang=lang, config=self.config, output_type=pytesseract.Output.DICT)
            
            # 计算置信度
            confidences = [conf for conf in data['conf'] if conf > 0]
            avg_confidence = np.mean(confidences) / 100 if confidences else 0.0
            
            # 构建文本块信息
            blocks = self._build_blocks(data)
            
            processing_time = time.time() - start_time
            
            return {
                'text': text.strip(),
                'confidence': round(avg_confidence, 3),
                'blocks': blocks,
                'processing_time': round(processing_time, 3),
                'engine': self.name
            }
            
        except Exception as e:
            logger.error(f"Tesseract识别失败: {str(e)}")
            return {
                'text': '',
                'confidence': 0.0,
                'blocks': [],
                'processing_time': time.time() - start_time,
                'engine': self.name,
                'error': str(e)
            }
    
    def _build_blocks(self, data: Dict) -> List[Dict]:
        """构建文本块信息"""
        blocks = []
        
        for i, text in enumerate(data['text']):
            if text.strip() and data['conf'][i] > 0:
                block = {
                    'text': text.strip(),
                    'confidence': round(data['conf'][i] / 100, 3),
                    'bbox': [data['left'][i], data['top'][i], 
                            data['left'][i] + data['width'][i], 
                            data['top'][i] + data['height'][i]],
                    'line': data['block_num'][i]
                }
                blocks.append(block)
        
        return blocks
    
    def get_supported_languages(self) -> List[str]:
        """获取支持的语言"""
        return ['zh', 'en', 'ja']


class PaddleOCREngine(OCREngineBase):
    """PaddleOCR引擎"""
    
    def __init__(self, use_gpu: bool = False):
        super().__init__('paddle_ocr')
        self.is_available = PADDLE_AVAILABLE
        
        if self.is_available:
            try:
                self.ocr = PaddleOCR(
                    use_angle_cls=True, 
                    lang='ch',
                    use_gpu=use_gpu,
                    show_log=False
                )
                logger.info("PaddleOCR初始化成功")
            except Exception as e:
                logger.error(f"PaddleOCR初始化失败: {str(e)}")
                self.is_available = False
    
    def recognize(self, image: np.ndarray, language: str = 'zh') -> Dict:
        """使用PaddleOCR识别文字"""
        if not self.is_available:
            raise RuntimeError("PaddleOCR不可用")
        
        start_time = time.time()
        
        try:
            # PaddleOCR识别
            results = self.ocr.ocr(image, cls=True)
            
            if not results or not results[0]:
                return {
                    'text': '',
                    'confidence': 0.0,
                    'blocks': [],
                    'processing_time': time.time() - start_time,
                    'engine': self.name
                }
            
            # 提取文本和置信度
            full_text = []
            blocks = []
            confidences = []
            
            for line_result in results[0]:
                if len(line_result) >= 2:
                    bbox, (text, confidence) = line_result[0], line_result[1]
                    
                    if text.strip():
                        full_text.append(text.strip())
                        confidences.append(confidence)
                        
                        # 构建块信息
                        block = {
                            'text': text.strip(),
                            'confidence': round(confidence, 3),
                            'bbox': [int(coord) for coord in bbox[0] + bbox[2]],  # [x1,y1,x2,y2]
                            'line': len(blocks) + 1
                        }
                        blocks.append(block)
            
            # 计算平均置信度
            avg_confidence = np.mean(confidences) if confidences else 0.0
            
            processing_time = time.time() - start_time
            
            return {
                'text': '\n'.join(full_text),
                'confidence': round(avg_confidence, 3),
                'blocks': blocks,
                'processing_time': round(processing_time, 3),
                'engine': self.name
            }
            
        except Exception as e:
            logger.error(f"PaddleOCR识别失败: {str(e)}")
            return {
                'text': '',
                'confidence': 0.0,
                'blocks': [],
                'processing_time': time.time() - start_time,
                'engine': self.name,
                'error': str(e)
            }
    
    def get_supported_languages(self) -> List[str]:
        """获取支持的语言"""
        return ['zh', 'en']


class EasyOCREngine(OCREngineBase):
    """EasyOCR引擎"""
    
    def __init__(self, use_gpu: bool = False):
        super().__init__('easy_ocr')
        self.is_available = EASYOCR_AVAILABLE
        
        if self.is_available:
            try:
                self.reader = easyocr.Reader(['ch_sim', 'en'], gpu=use_gpu)
                logger.info("EasyOCR初始化成功")
            except Exception as e:
                logger.error(f"EasyOCR初始化失败: {str(e)}")
                self.is_available = False
    
    def recognize(self, image: np.ndarray, language: str = 'zh') -> Dict:
        """使用EasyOCR识别文字"""
        if not self.is_available:
            raise RuntimeError("EasyOCR不可用")
        
        start_time = time.time()
        
        try:
            # EasyOCR识别
            results = self.reader.readtext(image)
            
            if not results:
                return {
                    'text': '',
                    'confidence': 0.0,
                    'blocks': [],
                    'processing_time': time.time() - start_time,
                    'engine': self.name
                }
            
            # 提取文本和置信度
            full_text = []
            blocks = []
            confidences = []
            
            for i, (bbox, text, confidence) in enumerate(results):
                if text.strip():
                    full_text.append(text.strip())
                    confidences.append(confidence)
                    
                    # 转换bbox格式
                    x_coords = [point[0] for point in bbox]
                    y_coords = [point[1] for point in bbox]
                    
                    block = {
                        'text': text.strip(),
                        'confidence': round(confidence, 3),
                        'bbox': [int(min(x_coords)), int(min(y_coords)), 
                                int(max(x_coords)), int(max(y_coords))],
                        'line': i + 1
                    }
                    blocks.append(block)
            
            # 计算平均置信度
            avg_confidence = np.mean(confidences) if confidences else 0.0
            
            processing_time = time.time() - start_time
            
            return {
                'text': '\n'.join(full_text),
                'confidence': round(avg_confidence, 3),
                'blocks': blocks,
                'processing_time': round(processing_time, 3),
                'engine': self.name
            }
            
        except Exception as e:
            logger.error(f"EasyOCR识别失败: {str(e)}")
            return {
                'text': '',
                'confidence': 0.0,
                'blocks': [],
                'processing_time': time.time() - start_time,
                'engine': self.name,
                'error': str(e)
            }
    
    def get_supported_languages(self) -> List[str]:
        """获取支持的语言"""
        return ['zh', 'en']


class MultiEngineOCR:
    """多引擎OCR识别器"""
    
    def __init__(self, use_gpu: bool = False):
        self.engines = {}
        self.image_processor = ImagePreprocessor()
        self.handwriting_processor = HandwritingPreprocessor()
        
        # 初始化可用的OCR引擎
        if TESSERACT_AVAILABLE:
            self.engines['tesseract'] = TesseractEngine()
        
        if PADDLE_AVAILABLE:
            self.engines['paddle'] = PaddleOCREngine(use_gpu)
        
        if EASYOCR_AVAILABLE:
            self.engines['easyocr'] = EasyOCREngine(use_gpu)
        
        # 如果没有任何OCR引擎可用，使用模拟引擎
        if MOCK_ENGINE and not self.engines:
            self.engines['mock'] = MockOCREngine()
            logger.warning("使用模拟OCR引擎进行开发测试")
        
        logger.info(f"可用OCR引擎: {list(self.engines.keys())}")
    
    def recognize_single_engine(self, image_path: str, engine_name: str = 'paddle', 
                              language: str = 'zh', enhance: bool = True,
                              is_handwriting: bool = False) -> Dict:
        """使用单个引擎识别"""
        
        if engine_name not in self.engines:
            raise ValueError(f"引擎 {engine_name} 不可用")
        
        engine = self.engines[engine_name]
        
        if not engine.is_available:
            raise RuntimeError(f"引擎 {engine_name} 初始化失败")
        
        try:
            # 图像预处理
            if enhance:
                if is_handwriting:
                    processor = self.handwriting_processor
                    # 使用专门的中文手写文字处理流程
                    if language == 'zh':
                        operations = ['denoise', 'deskew', 'stroke_enhance', 'chinese_optimize', 'contrast']
                    else:
                        operations = ['denoise', 'deskew', 'handwriting_enhance', 'contrast']
                else:
                    processor = self.image_processor
                    operations = ['denoise', 'deskew', 'contrast', 'binarize']
                
                processed_image, processing_info = processor.preprocess(image_path, operations)
            else:
                processed_image = cv2.imread(image_path)
                processing_info = {'applied_operations': []}
            
            # OCR识别
            result = engine.recognize(processed_image, language)
            
            # 添加处理信息
            result['preprocessing'] = processing_info
            result['language_requested'] = language
            result['enhancement_applied'] = enhance
            result['is_handwriting_mode'] = is_handwriting
            
            return result
            
        except Exception as e:
            logger.error(f"OCR识别失败: {str(e)}")
            raise
    
    def recognize_with_voting(self, image_path: str, language: str = 'zh', 
                            enhance: bool = True, is_handwriting: bool = False,
                            engines: List[str] = None) -> Dict:
        """多引擎投票识别"""
        
        if engines is None:
            # 默认使用所有可用引擎
            engines = [name for name, engine in self.engines.items() if engine.is_available]
        
        if not engines:
            raise RuntimeError("没有可用的OCR引擎")
        
        # 如果只有一个引擎，直接使用
        if len(engines) == 1:
            return self.recognize_single_engine(image_path, engines[0], language, enhance, is_handwriting)
        
        results = {}
        start_time = time.time()
        
        # 使用每个引擎识别
        for engine_name in engines:
            try:
                if engine_name in self.engines and self.engines[engine_name].is_available:
                    result = self.recognize_single_engine(
                        image_path, engine_name, language, enhance, is_handwriting
                    )
                    results[engine_name] = result
                    logger.info(f"引擎 {engine_name} 识别完成")
            except Exception as e:
                logger.warning(f"引擎 {engine_name} 识别失败: {str(e)}")
                continue
        
        if not results:
            raise RuntimeError("所有OCR引擎都识别失败")
        
        # 投票算法选择最优结果
        best_result = self.vote_best_result(results)
        
        # 添加投票信息
        best_result['voting_info'] = {
            'engines_used': list(results.keys()),
            'total_engines': len(engines),
            'voting_time': round(time.time() - start_time, 3),
            'engine_results': {
                name: {
                    'confidence': res.get('confidence', 0),
                    'text_length': len(res.get('text', '')),
                    'processing_time': res.get('processing_time', 0)
                }
                for name, res in results.items()
            }
        }
        
        return best_result
    
    def vote_best_result(self, results: Dict[str, Dict]) -> Dict:
        """投票算法选择最优结果"""
        
        if not results:
            raise ValueError("没有识别结果可供投票")
        
        if len(results) == 1:
            return list(results.values())[0]
        
        # 计算每个结果的综合得分
        scored_results = []
        
        for engine_name, result in results.items():
            confidence = result.get('confidence', 0)
            text_length = len(result.get('text', ''))
            processing_time = result.get('processing_time', float('inf'))
            
            # 综合得分计算
            # 置信度权重: 0.6, 文本长度权重: 0.3, 速度权重: 0.1
            confidence_score = confidence * 0.6
            length_score = min(text_length / 100, 1.0) * 0.3  # 归一化文本长度
            speed_score = max(0, (10 - processing_time) / 10) * 0.1  # 速度得分
            
            total_score = confidence_score + length_score + speed_score
            
            scored_results.append((total_score, engine_name, result))
        
        # 按得分排序，选择最高分
        scored_results.sort(key=lambda x: x[0], reverse=True)
        
        best_score, best_engine, best_result = scored_results[0]
        
        logger.info(f"投票结果: 选择引擎 {best_engine}，得分: {best_score:.3f}")
        
        # 添加投票信息到结果中
        best_result['voting_score'] = round(best_score, 3)
        best_result['selected_engine'] = best_engine
        
        return best_result
    
    def calculate_text_similarity(self, text1: str, text2: str) -> float:
        """计算文本相似度"""
        try:
            from difflib import SequenceMatcher
            return SequenceMatcher(None, text1, text2).ratio()
        except:
            # 简单的字符匹配相似度
            if not text1 or not text2:
                return 0.0
            
            common_chars = sum(1 for c in text1 if c in text2)
            total_chars = len(set(text1 + text2))
            
            return common_chars / total_chars if total_chars > 0 else 0.0
    
    def get_available_engines(self) -> Dict[str, Dict]:
        """获取可用引擎信息"""
        engine_info = {}
        
        for name, engine in self.engines.items():
            engine_info[name] = {
                'name': name,
                'available': engine.is_available,
                'supported_languages': engine.get_supported_languages() if engine.is_available else []
            }
        
        return engine_info
    
    def batch_recognize(self, image_paths: List[str], engine_name: str = 'paddle',
                       language: str = 'zh', enhance: bool = True,
                       is_handwriting: bool = False) -> List[Dict]:
        """批量识别"""
        results = []
        
        for i, image_path in enumerate(image_paths):
            try:
                logger.info(f"正在处理第 {i+1}/{len(image_paths)} 张图片: {image_path}")
                
                result = self.recognize_single_engine(
                    image_path, engine_name, language, enhance, is_handwriting
                )
                
                result['image_index'] = i
                result['image_path'] = image_path
                results.append(result)
                
            except Exception as e:
                logger.error(f"处理图片 {image_path} 失败: {str(e)}")
                
                error_result = {
                    'image_index': i,
                    'image_path': image_path,
                    'text': '',
                    'confidence': 0.0,
                    'error': str(e),
                    'engine': engine_name
                }
                results.append(error_result)
        
        return results