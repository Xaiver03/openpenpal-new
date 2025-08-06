import cv2
import numpy as np
from PIL import Image, ImageEnhance
import logging
from typing import Tuple, List, Optional, Dict
import os

logger = logging.getLogger(__name__)


class ImagePreprocessor:
    """图像预处理类"""
    
    def __init__(self):
        self.pipeline_steps = {
            'resize': self.resize_image,
            'denoise': self.denoise,
            'deskew': self.deskew,
            'contrast': self.enhance_contrast,
            'brightness': self.adjust_brightness,
            'binarize': self.binarize,
            'sharpen': self.sharpen
        }
    
    def preprocess(self, image_path: str, operations: List[str] = None) -> Tuple[np.ndarray, Dict]:
        """
        图像预处理主流程
        
        Args:
            image_path: 图像文件路径
            operations: 要执行的操作列表
            
        Returns:
            处理后的图像和处理信息
        """
        try:
            # 读取图像
            if not os.path.exists(image_path):
                raise FileNotFoundError(f"图像文件不存在: {image_path}")
            
            image = cv2.imread(image_path)
            if image is None:
                raise ValueError(f"无法读取图像文件: {image_path}")
            
            original_image = image.copy()
            processing_info = {
                'original_size': image.shape[:2],
                'applied_operations': [],
                'quality_metrics': {}
            }
            
            # 默认处理流程
            if operations is None:
                operations = ['denoise', 'deskew', 'contrast', 'binarize']
            
            # 执行处理步骤
            for operation in operations:
                if operation in self.pipeline_steps:
                    try:
                        image = self.pipeline_steps[operation](image)
                        processing_info['applied_operations'].append(operation)
                        logger.info(f"已执行图像处理: {operation}")
                    except Exception as e:
                        logger.warning(f"图像处理步骤 {operation} 失败: {str(e)}")
                        continue
                else:
                    logger.warning(f"未知的处理操作: {operation}")
            
            # 计算质量指标
            processing_info['quality_metrics'] = self.calculate_quality_metrics(
                original_image, image
            )
            
            return image, processing_info
            
        except Exception as e:
            logger.error(f"图像预处理失败: {str(e)}")
            raise
    
    def resize_image(self, image: np.ndarray, max_width: int = None, max_height: int = None) -> np.ndarray:
        """调整图像大小"""
        # 从配置获取默认值
        from flask import current_app
        if max_width is None:
            max_width = current_app.config.get('MAX_IMAGE_SIZE', 2048)
        if max_height is None:
            max_height = current_app.config.get('MAX_IMAGE_SIZE', 2048)
        
        height, width = image.shape[:2]
        
        # 如果图像已经在限制范围内，不需要调整
        if width <= max_width and height <= max_height:
            return image
        
        # 计算缩放比例
        scale = min(max_width / width, max_height / height)
        new_width = int(width * scale)
        new_height = int(height * scale)
        
        # 使用高质量插值
        resized = cv2.resize(image, (new_width, new_height), interpolation=cv2.INTER_LANCZOS4)
        
        logger.info(f"图像大小调整: {width}x{height} -> {new_width}x{new_height}")
        return resized
    
    def denoise(self, image: np.ndarray) -> np.ndarray:
        """降噪处理"""
        if len(image.shape) == 3:
            # 彩色图像降噪
            denoised = cv2.fastNlMeansDenoisingColored(image, None, 10, 10, 7, 21)
        else:
            # 灰度图像降噪
            denoised = cv2.fastNlMeansDenoising(image, None, 10, 7, 21)
        
        return denoised
    
    def deskew(self, image: np.ndarray) -> np.ndarray:
        """倾斜矫正"""
        try:
            # 转换为灰度图
            if len(image.shape) == 3:
                gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
            else:
                gray = image.copy()
            
            # 边缘检测
            edges = cv2.Canny(gray, 50, 150, apertureSize=3)
            
            # 霍夫变换检测直线
            lines = cv2.HoughLines(edges, 1, np.pi/180, threshold=100)
            
            if lines is not None and len(lines) > 0:
                # 计算倾斜角度
                angle = self.calculate_skew_angle(lines)
                
                # 如果倾斜角度较小，进行矫正
                if abs(angle) > 0.5:  # 大于0.5度才矫正
                    corrected = self.rotate_image(image, angle)
                    logger.info(f"倾斜矫正角度: {angle:.2f}度")
                    return corrected
            
            return image
            
        except Exception as e:
            logger.warning(f"倾斜矫正失败: {str(e)}")
            return image
    
    def calculate_skew_angle(self, lines: np.ndarray) -> float:
        """计算倾斜角度"""
        angles = []
        
        for line in lines:
            rho, theta = line[0]
            angle = theta * 180 / np.pi - 90
            angles.append(angle)
        
        # 使用角度的中位数作为倾斜角度
        if angles:
            return np.median(angles)
        return 0.0
    
    def rotate_image(self, image: np.ndarray, angle: float) -> np.ndarray:
        """旋转图像"""
        height, width = image.shape[:2]
        center = (width // 2, height // 2)
        
        # 计算旋转矩阵
        rotation_matrix = cv2.getRotationMatrix2D(center, angle, 1.0)
        
        # 计算新的边界框
        cos = np.abs(rotation_matrix[0, 0])
        sin = np.abs(rotation_matrix[0, 1])
        new_width = int((height * sin) + (width * cos))
        new_height = int((height * cos) + (width * sin))
        
        # 调整旋转矩阵的平移部分
        rotation_matrix[0, 2] += (new_width / 2) - center[0]
        rotation_matrix[1, 2] += (new_height / 2) - center[1]
        
        # 执行旋转
        rotated = cv2.warpAffine(image, rotation_matrix, (new_width, new_height), 
                                flags=cv2.INTER_CUBIC, borderMode=cv2.BORDER_REPLICATE)
        
        return rotated
    
    def enhance_contrast(self, image: np.ndarray, alpha: float = 1.2) -> np.ndarray:
        """增强对比度"""
        try:
            # 使用CLAHE (自适应直方图均衡化)
            if len(image.shape) == 3:
                # 彩色图像：转换到LAB颜色空间，只处理L通道
                lab = cv2.cvtColor(image, cv2.COLOR_BGR2LAB)
                clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
                lab[:, :, 0] = clahe.apply(lab[:, :, 0])
                enhanced = cv2.cvtColor(lab, cv2.COLOR_LAB2BGR)
            else:
                # 灰度图像
                clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
                enhanced = clahe.apply(image)
            
            return enhanced
            
        except Exception as e:
            logger.warning(f"对比度增强失败: {str(e)}")
            # 降级到简单的对比度调整
            return cv2.convertScaleAbs(image, alpha=alpha, beta=0)
    
    def adjust_brightness(self, image: np.ndarray, beta: int = 30) -> np.ndarray:
        """调整亮度"""
        return cv2.convertScaleAbs(image, alpha=1.0, beta=beta)
    
    def binarize(self, image: np.ndarray) -> np.ndarray:
        """二值化处理"""
        # 转换为灰度图
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image.copy()
        
        # 使用自适应阈值
        binary = cv2.adaptiveThreshold(
            gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
            cv2.THRESH_BINARY, 11, 2
        )
        
        # 形态学操作去除噪点
        kernel = np.ones((2, 2), np.uint8)
        cleaned = cv2.morphologyEx(binary, cv2.MORPH_CLOSE, kernel)
        
        return cleaned
    
    def sharpen(self, image: np.ndarray) -> np.ndarray:
        """锐化处理"""
        kernel = np.array([[-1, -1, -1],
                          [-1,  9, -1],
                          [-1, -1, -1]])
        
        sharpened = cv2.filter2D(image, -1, kernel)
        return sharpened
    
    def calculate_quality_metrics(self, original: np.ndarray, processed: np.ndarray) -> Dict:
        """计算图像质量指标"""
        try:
            # 转换为灰度图进行比较
            if len(original.shape) == 3:
                orig_gray = cv2.cvtColor(original, cv2.COLOR_BGR2GRAY)
            else:
                orig_gray = original
            
            if len(processed.shape) == 3:
                proc_gray = cv2.cvtColor(processed, cv2.COLOR_BGR2GRAY)
            else:
                proc_gray = processed
            
            # 调整大小以便比较
            if orig_gray.shape != proc_gray.shape:
                proc_gray = cv2.resize(proc_gray, (orig_gray.shape[1], orig_gray.shape[0]))
            
            # 计算PSNR (峰值信噪比)
            mse = np.mean((orig_gray.astype(float) - proc_gray.astype(float)) ** 2)
            if mse == 0:
                psnr = float('inf')
            else:
                psnr = 20 * np.log10(255.0 / np.sqrt(mse))
            
            # 计算对比度改善
            orig_contrast = orig_gray.std()
            proc_contrast = proc_gray.std()
            contrast_improvement = (proc_contrast - orig_contrast) / orig_contrast if orig_contrast > 0 else 0
            
            return {
                'psnr': round(psnr, 2),
                'contrast_improvement': round(contrast_improvement, 3),
                'noise_reduction': 0.65,  # 模拟值，实际需要更复杂的计算
                'overall_quality': min(max((psnr / 30.0), 0), 1)  # 归一化到0-1
            }
            
        except Exception as e:
            logger.warning(f"质量指标计算失败: {str(e)}")
            return {
                'psnr': 0,
                'contrast_improvement': 0,
                'noise_reduction': 0,
                'overall_quality': 0.5
            }
    
    def save_processed_image(self, image: np.ndarray, output_path: str) -> bool:
        """保存处理后的图像"""
        try:
            # 确保输出目录存在
            os.makedirs(os.path.dirname(output_path), exist_ok=True)
            
            # 保存图像
            success = cv2.imwrite(output_path, image)
            
            if success:
                logger.info(f"处理后的图像已保存: {output_path}")
                return True
            else:
                logger.error(f"保存图像失败: {output_path}")
                return False
                
        except Exception as e:
            logger.error(f"保存图像异常: {str(e)}")
            return False


class HandwritingPreprocessor(ImagePreprocessor):
    """专门针对手写文字的图像预处理器"""
    
    def __init__(self):
        super().__init__()
        # 添加手写文字特有的处理步骤
        self.pipeline_steps.update({
            'handwriting_enhance': self.handwriting_enhance,
            'character_segment': self.segment_characters,
            'chinese_optimize': self.chinese_handwriting_optimize,
            'stroke_enhance': self.enhance_strokes
        })
    
    def handwriting_enhance(self, image: np.ndarray) -> np.ndarray:
        """手写文字专用增强"""
        # 转换为灰度图
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image.copy()
        
        # 适应性阈值二值化 - 对手写文字效果更好
        binary = cv2.adaptiveThreshold(
            gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
            cv2.THRESH_BINARY, 15, 10
        )
        
        # 形态学操作 - 专门针对手写文字
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (3, 3))
        
        # 闭运算连接笔画
        closed = cv2.morphologyEx(binary, cv2.MORPH_CLOSE, kernel)
        
        # 开运算去除噪点
        opened = cv2.morphologyEx(closed, cv2.MORPH_OPEN, kernel)
        
        return opened
    
    def segment_characters(self, image: np.ndarray) -> List[np.ndarray]:
        """字符分割 - 为逐字符识别做准备"""
        try:
            # 转换为二值图像
            if len(image.shape) == 3:
                gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
                binary = cv2.threshold(gray, 127, 255, cv2.THRESH_BINARY)[1]
            else:
                binary = image.copy()
            
            # 查找轮廓
            contours, _ = cv2.findContours(binary, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
            
            # 过滤和排序轮廓
            character_contours = []
            for contour in contours:
                x, y, w, h = cv2.boundingRect(contour)
                # 过滤太小的区域
                if w > 10 and h > 10:
                    character_contours.append((x, y, w, h))
            
            # 按x坐标排序（从左到右）
            character_contours.sort(key=lambda c: c[0])
            
            # 提取字符图像
            characters = []
            for x, y, w, h in character_contours:
                char_img = binary[y:y+h, x:x+w]
                characters.append(char_img)
            
            logger.info(f"分割出 {len(characters)} 个字符区域")
            return characters
            
        except Exception as e:
            logger.warning(f"字符分割失败: {str(e)}")
            return [image]  # 返回原图像作为单个字符
    
    def chinese_handwriting_optimize(self, image: np.ndarray) -> np.ndarray:
        """中文手写文字专项优化"""
        try:
            # 转换为灰度图
            if len(image.shape) == 3:
                gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
            else:
                gray = image.copy()
            
            # 中文手写文字特点：笔画连接、结构复杂
            # 使用更大的自适应阈值窗口
            binary = cv2.adaptiveThreshold(
                gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
                cv2.THRESH_BINARY, 21, 15
            )
            
            # 专门针对中文字符的形态学操作
            # 使用矩形核来保持汉字的方形结构
            rect_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (3, 3))
            
            # 闭运算：连接断开的笔画
            closed = cv2.morphologyEx(binary, cv2.MORPH_CLOSE, rect_kernel, iterations=2)
            
            # 开运算：去除细小噪点但保持笔画完整性
            ellipse_kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (2, 2))
            opened = cv2.morphologyEx(closed, cv2.MORPH_OPEN, ellipse_kernel)
            
            # 中文字符笔画加粗（适度）
            dilate_kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (2, 2))
            thickened = cv2.dilate(opened, dilate_kernel, iterations=1)
            
            logger.info("中文手写文字优化完成")
            return thickened
            
        except Exception as e:
            logger.warning(f"中文手写优化失败: {str(e)}")
            return image
    
    def enhance_strokes(self, image: np.ndarray) -> np.ndarray:
        """笔画增强处理"""
        try:
            # 转换为灰度图
            if len(image.shape) == 3:
                gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
            else:
                gray = image.copy()
            
            # 使用高斯滤波先平滑图像
            blurred = cv2.GaussianBlur(gray, (3, 3), 0)
            
            # 使用锐化核增强笔画边缘
            sharpen_kernel = np.array([
                [0, -1, 0],
                [-1, 5, -1],
                [0, -1, 0]
            ], dtype=np.float32)
            
            sharpened = cv2.filter2D(blurred, -1, sharpen_kernel)
            
            # 自适应直方图均衡化增强对比度
            clahe = cv2.createCLAHE(clipLimit=3.0, tileGridSize=(8, 8))
            enhanced = clahe.apply(sharpened)
            
            # 使用双边滤波平滑噪声但保持边缘
            bilateral = cv2.bilateralFilter(enhanced, 9, 75, 75)
            
            logger.info("笔画增强处理完成")
            return bilateral
            
        except Exception as e:
            logger.warning(f"笔画增强失败: {str(e)}")
            return image
    
    def preprocess_for_chinese_handwriting(self, image_path: str) -> Tuple[np.ndarray, Dict]:
        """专门用于中文手写文字的预处理流程"""
        operations = [
            'denoise',           # 降噪
            'deskew',           # 倾斜矫正
            'stroke_enhance',   # 笔画增强
            'chinese_optimize', # 中文优化
            'contrast'          # 对比度增强
        ]
        
        return self.preprocess(image_path, operations)