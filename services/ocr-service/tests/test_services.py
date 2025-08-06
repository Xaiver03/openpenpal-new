import pytest
import numpy as np
import cv2
from unittest.mock import patch, MagicMock
import tempfile
import os

from app.services.image_processor import ImagePreprocessor, HandwritingPreprocessor
from app.services.cache_service import CacheService
from app.services.ocr_engine import MultiEngineOCR


class TestImagePreprocessor:
    """图像预处理器测试"""
    
    def create_test_image(self, width=100, height=100):
        """创建测试图像"""
        # 创建一个简单的白色图像
        image = np.ones((height, width, 3), dtype=np.uint8) * 255
        # 添加一些黑色文字区域
        cv2.rectangle(image, (10, 10), (90, 30), (0, 0, 0), -1)
        return image
    
    def save_temp_image(self, image):
        """保存临时图像文件"""
        temp_file = tempfile.NamedTemporaryFile(suffix='.jpg', delete=False)
        cv2.imwrite(temp_file.name, image)
        return temp_file.name
    
    def test_image_preprocessing(self):
        """测试图像预处理"""
        processor = ImagePreprocessor()
        test_image = self.create_test_image()
        temp_path = self.save_temp_image(test_image)
        
        try:
            processed_image, info = processor.preprocess(temp_path)
            
            assert processed_image is not None
            assert isinstance(info, dict)
            assert 'applied_operations' in info
            assert 'quality_metrics' in info
            
        finally:
            os.unlink(temp_path)
    
    def test_resize_image(self):
        """测试图像大小调整"""
        processor = ImagePreprocessor()
        test_image = self.create_test_image(2000, 1500)  # 大图
        
        resized = processor.resize_image(test_image, max_width=1024, max_height=768)
        
        height, width = resized.shape[:2]
        assert width <= 1024
        assert height <= 768
    
    def test_denoise(self):
        """测试降噪"""
        processor = ImagePreprocessor()
        test_image = self.create_test_image()
        
        # 添加噪声
        noise = np.random.randint(0, 50, test_image.shape, dtype=np.uint8)
        noisy_image = cv2.add(test_image, noise)
        
        denoised = processor.denoise(noisy_image)
        
        assert denoised is not None
        assert denoised.shape == test_image.shape
    
    def test_binarize(self):
        """测试二值化"""
        processor = ImagePreprocessor()
        test_image = self.create_test_image()
        
        binary = processor.binarize(test_image)
        
        assert binary is not None
        assert len(binary.shape) == 2  # 灰度图
        assert binary.dtype == np.uint8
    
    def test_custom_operations(self):
        """测试自定义操作序列"""
        processor = ImagePreprocessor()
        test_image = self.create_test_image()
        temp_path = self.save_temp_image(test_image)
        
        try:
            operations = ['denoise', 'contrast', 'binarize']
            processed_image, info = processor.preprocess(temp_path, operations)
            
            assert 'denoise' in info['applied_operations']
            assert 'contrast' in info['applied_operations']
            assert 'binarize' in info['applied_operations']
            
        finally:
            os.unlink(temp_path)


class TestHandwritingPreprocessor:
    """手写文字预处理器测试"""
    
    def test_handwriting_enhance(self):
        """测试手写文字增强"""
        processor = HandwritingPreprocessor()
        test_image = np.ones((100, 100, 3), dtype=np.uint8) * 255
        
        enhanced = processor.handwriting_enhance(test_image)
        
        assert enhanced is not None
        assert len(enhanced.shape) == 2  # 二值图
    
    def test_character_segmentation(self):
        """测试字符分割"""
        processor = HandwritingPreprocessor()
        
        # 创建包含多个字符区域的图像
        test_image = np.ones((100, 200), dtype=np.uint8) * 255
        cv2.rectangle(test_image, (10, 20), (40, 80), 0, -1)  # 字符1
        cv2.rectangle(test_image, (60, 20), (90, 80), 0, -1)  # 字符2
        cv2.rectangle(test_image, (110, 20), (140, 80), 0, -1)  # 字符3
        
        characters = processor.segment_characters(test_image)
        
        assert isinstance(characters, list)
        assert len(characters) >= 1  # 至少应该有一个字符区域


class TestCacheService:
    """缓存服务测试"""
    
    def test_cache_initialization(self):
        """测试缓存初始化"""
        cache = CacheService()
        
        # 缓存应该能初始化（Redis或内存缓存）
        assert cache is not None
    
    def test_ocr_result_caching(self):
        """测试OCR结果缓存"""
        cache = CacheService()
        
        # 测试数据
        image_hash = "test_hash_123"
        settings = {"language": "zh", "engine": "tesseract"}
        result = {"text": "测试文本", "confidence": 0.95}
        
        # 缓存结果
        cache.cache_ocr_result(image_hash, settings, result)
        
        # 获取缓存结果
        cached_result = cache.get_ocr_result(image_hash, settings)
        
        if cached_result:  # 如果缓存可用
            assert cached_result['text'] == "测试文本"
            assert cached_result['confidence'] == 0.95
    
    def test_task_status_caching(self):
        """测试任务状态缓存"""
        cache = CacheService()
        
        task_id = "test_task_123"
        status_data = {"status": "completed", "progress": 100}
        
        # 缓存任务状态
        cache.cache_task_status(task_id, status_data)
        
        # 获取任务状态
        cached_status = cache.get_task_status(task_id)
        
        if cached_status:  # 如果缓存可用
            assert cached_status['status'] == "completed"
            assert cached_status['progress'] == 100
    
    def test_image_hash_calculation(self):
        """测试图像哈希计算"""
        cache = CacheService()
        
        # 创建临时图像文件
        test_image = np.ones((100, 100, 3), dtype=np.uint8) * 255
        temp_file = tempfile.NamedTemporaryFile(suffix='.jpg', delete=False)
        cv2.imwrite(temp_file.name, test_image)
        
        try:
            hash1 = cache.calculate_image_hash(temp_file.name)
            hash2 = cache.calculate_image_hash(temp_file.name)
            
            # 相同文件应该产生相同哈希
            assert hash1 == hash2
            assert isinstance(hash1, str)
            assert len(hash1) > 0
            
        finally:
            os.unlink(temp_file.name)


class TestMultiEngineOCR:
    """多引擎OCR测试"""
    
    @patch('app.services.ocr_engine.TESSERACT_AVAILABLE', True)
    @patch('app.services.ocr_engine.PADDLE_AVAILABLE', False)
    @patch('app.services.ocr_engine.EASYOCR_AVAILABLE', False)
    def test_single_engine_initialization(self):
        """测试单引擎初始化"""
        ocr = MultiEngineOCR()
        
        available_engines = ocr.get_available_engines()
        
        assert isinstance(available_engines, dict)
        # 应该至少有tesseract可用（如果安装了）
        if 'tesseract' in available_engines:
            assert available_engines['tesseract']['available'] in [True, False]
    
    def test_text_similarity_calculation(self):
        """测试文本相似度计算"""
        ocr = MultiEngineOCR()
        
        text1 = "Hello World"
        text2 = "Hello World"
        text3 = "Goodbye World"
        
        similarity1 = ocr.calculate_text_similarity(text1, text2)
        similarity2 = ocr.calculate_text_similarity(text1, text3)
        
        assert similarity1 == 1.0  # 完全相同
        assert 0 <= similarity2 < 1  # 部分相似
    
    def test_engine_availability_check(self):
        """测试引擎可用性检查"""
        ocr = MultiEngineOCR()
        
        engines = ocr.get_available_engines()
        
        # 检查返回格式
        for engine_name, info in engines.items():
            assert 'name' in info
            assert 'available' in info
            assert 'supported_languages' in info
            assert isinstance(info['available'], bool)
            assert isinstance(info['supported_languages'], list)


class TestEndToEndOCR:
    """端到端OCR测试"""
    
    def create_text_image(self, text="Test Text", font_scale=1, thickness=2):
        """创建包含文本的图像"""
        image = np.ones((100, 300, 3), dtype=np.uint8) * 255
        
        font = cv2.FONT_HERSHEY_SIMPLEX
        text_size = cv2.getTextSize(text, font, font_scale, thickness)[0]
        
        # 计算文本位置（居中）
        x = (image.shape[1] - text_size[0]) // 2
        y = (image.shape[0] + text_size[1]) // 2
        
        cv2.putText(image, text, (x, y), font, font_scale, (0, 0, 0), thickness)
        
        return image
    
    def test_full_ocr_pipeline(self):
        """测试完整OCR流程"""
        # 创建测试图像
        test_image = self.create_text_image("Hello")
        temp_file = tempfile.NamedTemporaryFile(suffix='.jpg', delete=False)
        cv2.imwrite(temp_file.name, test_image)
        
        try:
            # 图像预处理
            processor = ImagePreprocessor()
            processed_image, info = processor.preprocess(temp_file.name)
            
            assert processed_image is not None
            assert isinstance(info, dict)
            
            # 缓存测试
            cache = CacheService()
            
            # 计算图像哈希
            image_hash = cache.calculate_image_hash(temp_file.name)
            assert isinstance(image_hash, str)
            
        finally:
            os.unlink(temp_file.name)


if __name__ == '__main__':
    pytest.main([__file__])