import re
import logging
from typing import Dict, List, Tuple, Optional
from difflib import SequenceMatcher
import jieba
import numpy as np
from collections import Counter

logger = logging.getLogger(__name__)


class TextValidator:
    """文本验证和相似度计算服务"""
    
    def __init__(self):
        self.sensitive_words = self._load_sensitive_words()
        self.stop_words = self._load_stop_words()
        
        # 初始化jieba分词
        try:
            jieba.initialize()
        except:
            pass
    
    def validate_text_similarity(self, original_text: str, ocr_text: str, 
                                validation_rules: Dict = None) -> Dict:
        """验证OCR文本与原始文本的相似度"""
        try:
            # 默认验证规则
            if validation_rules is None:
                validation_rules = {
                    'min_similarity': 0.8,
                    'check_sensitive_words': True,
                    'max_length': 5000
                }
            
            # 基础验证
            if not original_text or not ocr_text:
                return {
                    'is_valid': False,
                    'similarity_score': 0.0,
                    'issues': ['原始文本或OCR文本为空'],
                    'suggestions': [],
                    'content_analysis': self._analyze_content(ocr_text if ocr_text else original_text)
                }
            
            # 长度检查
            max_length = validation_rules.get('max_length', 5000)
            if len(ocr_text) > max_length:
                return {
                    'is_valid': False,
                    'similarity_score': 0.0,
                    'issues': [f'文本长度超出限制 ({len(ocr_text)} > {max_length})'],
                    'suggestions': [],
                    'content_analysis': self._analyze_content(ocr_text)
                }
            
            # 计算相似度
            similarity_score = self._calculate_comprehensive_similarity(original_text, ocr_text)
            
            # 检测问题和生成建议
            issues = []
            suggestions = self._generate_suggestions(original_text, ocr_text)
            
            # 敏感词检查
            if validation_rules.get('check_sensitive_words', True):
                sensitive_issues = self._check_sensitive_words(ocr_text)
                issues.extend(sensitive_issues)
            
            # 相似度阈值检查
            min_similarity = validation_rules.get('min_similarity', 0.8)
            is_valid = similarity_score >= min_similarity and len(issues) == 0
            
            if similarity_score < min_similarity:
                issues.append(f'文本相似度过低 ({similarity_score:.3f} < {min_similarity})')
            
            # 内容分析
            content_analysis = self._analyze_content(ocr_text)
            
            result = {
                'is_valid': is_valid,
                'similarity_score': round(similarity_score, 3),
                'issues': issues,
                'suggestions': suggestions,
                'content_analysis': content_analysis,
                'validation_details': {
                    'character_similarity': self._calculate_character_similarity(original_text, ocr_text),
                    'word_similarity': self._calculate_word_similarity(original_text, ocr_text),
                    'structure_similarity': self._calculate_structure_similarity(original_text, ocr_text),
                    'length_ratio': len(ocr_text) / len(original_text) if original_text else 0
                }
            }
            
            return result
            
        except Exception as e:
            logger.error(f"文本验证失败: {str(e)}")
            return {
                'is_valid': False,
                'similarity_score': 0.0,
                'issues': [f'验证过程出现错误: {str(e)}'],
                'suggestions': [],
                'content_analysis': {'error': str(e)}
            }
    
    def _calculate_comprehensive_similarity(self, text1: str, text2: str) -> float:
        """计算综合相似度"""
        try:
            # 1. 字符级相似度 (权重: 0.3)
            char_sim = self._calculate_character_similarity(text1, text2)
            
            # 2. 词级相似度 (权重: 0.4)
            word_sim = self._calculate_word_similarity(text1, text2)
            
            # 3. 结构相似度 (权重: 0.2)
            struct_sim = self._calculate_structure_similarity(text1, text2)
            
            # 4. 语义相似度 (权重: 0.1)
            semantic_sim = self._calculate_semantic_similarity(text1, text2)
            
            # 综合计算
            comprehensive_similarity = (
                char_sim * 0.3 + 
                word_sim * 0.4 + 
                struct_sim * 0.2 + 
                semantic_sim * 0.1
            )
            
            return min(1.0, max(0.0, comprehensive_similarity))
            
        except Exception as e:
            logger.warning(f"综合相似度计算失败: {str(e)}")
            # 降级到简单的字符相似度
            return SequenceMatcher(None, text1, text2).ratio()
    
    def _calculate_character_similarity(self, text1: str, text2: str) -> float:
        """计算字符级相似度"""
        if not text1 or not text2:
            return 0.0
        
        return SequenceMatcher(None, text1, text2).ratio()
    
    def _calculate_word_similarity(self, text1: str, text2: str) -> float:
        """计算词级相似度"""
        try:
            # 中文分词
            words1 = set(jieba.cut(text1))
            words2 = set(jieba.cut(text2))
            
            # 去除停用词
            words1 = words1 - self.stop_words
            words2 = words2 - self.stop_words
            
            if not words1 or not words2:
                return 0.0
            
            # 计算Jaccard相似度
            intersection = len(words1.intersection(words2))
            union = len(words1.union(words2))
            
            if union == 0:
                return 0.0
            
            return intersection / union
            
        except Exception as e:
            logger.warning(f"词级相似度计算失败: {str(e)}")
            # 降级到简单的分词
            words1 = set(text1.split())
            words2 = set(text2.split())
            
            if not words1 or not words2:
                return 0.0
            
            intersection = len(words1.intersection(words2))
            union = len(words1.union(words2))
            
            return intersection / union if union > 0 else 0.0
    
    def _calculate_structure_similarity(self, text1: str, text2: str) -> float:
        """计算结构相似度（段落、句子数量等）"""
        try:
            # 段落数量比较
            paragraphs1 = [p.strip() for p in text1.split('\n') if p.strip()]
            paragraphs2 = [p.strip() for p in text2.split('\n') if p.strip()]
            
            para_ratio = min(len(paragraphs1), len(paragraphs2)) / max(len(paragraphs1), len(paragraphs2), 1)
            
            # 句子数量比较
            sentences1 = re.split('[。！？.!?]', text1)
            sentences2 = re.split('[。！？.!?]', text2)
            
            sent_ratio = min(len(sentences1), len(sentences2)) / max(len(sentences1), len(sentences2), 1)
            
            # 长度比例
            length_ratio = min(len(text1), len(text2)) / max(len(text1), len(text2), 1)
            
            # 综合结构相似度
            structure_sim = (para_ratio * 0.3 + sent_ratio * 0.4 + length_ratio * 0.3)
            
            return structure_sim
            
        except Exception as e:
            logger.warning(f"结构相似度计算失败: {str(e)}")
            # 降级到长度比例
            return min(len(text1), len(text2)) / max(len(text1), len(text2), 1)
    
    def _calculate_semantic_similarity(self, text1: str, text2: str) -> float:
        """计算语义相似度（简化版本，可以后续集成词向量模型）"""
        try:
            # 提取关键词
            keywords1 = self._extract_keywords(text1)
            keywords2 = self._extract_keywords(text2)
            
            if not keywords1 or not keywords2:
                return 0.0
            
            # 计算关键词重叠度
            common_keywords = len(set(keywords1).intersection(set(keywords2)))
            total_keywords = len(set(keywords1).union(set(keywords2)))
            
            if total_keywords == 0:
                return 0.0
            
            return common_keywords / total_keywords
            
        except Exception as e:
            logger.warning(f"语义相似度计算失败: {str(e)}")
            return 0.5  # 默认中等相似度
    
    def _extract_keywords(self, text: str, top_n: int = 10) -> List[str]:
        """提取关键词"""
        try:
            words = jieba.cut(text)
            words = [w for w in words if w not in self.stop_words and len(w) > 1]
            
            # 词频统计
            word_freq = Counter(words)
            
            # 返回频率最高的关键词
            return [word for word, freq in word_freq.most_common(top_n)]
            
        except Exception as e:
            logger.warning(f"关键词提取失败: {str(e)}")
            return []
    
    def _generate_suggestions(self, original_text: str, ocr_text: str) -> List[Dict]:
        """生成文本纠错建议"""
        suggestions = []
        
        try:
            # 使用SequenceMatcher找到差异
            matcher = SequenceMatcher(None, original_text, ocr_text)
            
            for tag, i1, i2, j1, j2 in matcher.get_opcodes():
                if tag == 'replace':
                    original_part = original_text[i1:i2]
                    ocr_part = ocr_text[j1:j2]
                    
                    if len(original_part) > 0 and len(ocr_part) > 0:
                        suggestion = {
                            'type': 'correction',
                            'original': ocr_part,
                            'suggested': original_part,
                            'confidence': 0.9,  # 基于原始文本的建议置信度高
                            'position': j1,
                            'reason': 'OCR识别错误'
                        }
                        suggestions.append(suggestion)
                
                elif tag == 'delete':
                    # OCR中多出的字符
                    extra_part = ocr_text[j1:j2]
                    if len(extra_part.strip()) > 0:
                        suggestion = {
                            'type': 'deletion',
                            'original': extra_part,
                            'suggested': '',
                            'confidence': 0.8,
                            'position': j1,
                            'reason': 'OCR识别多余字符'
                        }
                        suggestions.append(suggestion)
                
                elif tag == 'insert':
                    # OCR中缺失的字符
                    missing_part = original_text[i1:i2]
                    if len(missing_part.strip()) > 0:
                        suggestion = {
                            'type': 'insertion',
                            'original': '',
                            'suggested': missing_part,
                            'confidence': 0.85,
                            'position': j1,
                            'reason': 'OCR识别遗漏字符'
                        }
                        suggestions.append(suggestion)
            
            # 限制建议数量
            return suggestions[:10]
            
        except Exception as e:
            logger.warning(f"生成建议失败: {str(e)}")
            return []
    
    def _check_sensitive_words(self, text: str) -> List[str]:
        """检查敏感词"""
        issues = []
        
        try:
            text_lower = text.lower()
            
            for word in self.sensitive_words:
                if word in text_lower:
                    issues.append(f'包含敏感词汇: {word}')
            
            return issues
            
        except Exception as e:
            logger.warning(f"敏感词检查失败: {str(e)}")
            return []
    
    def _analyze_content(self, text: str) -> Dict:
        """分析文本内容"""
        try:
            analysis = {
                'word_count': len(text.replace(' ', '')),  # 中文字符数
                'paragraph_count': len([p for p in text.split('\n') if p.strip()]),
                'sentence_count': len(re.split('[。！？.!?]', text)),
                'language': self._detect_language(text),
                'contains_sensitive': len(self._check_sensitive_words(text)) > 0,
                'sentiment': self._analyze_sentiment(text),
                'readability': self._calculate_readability(text)
            }
            
            return analysis
            
        except Exception as e:
            logger.warning(f"内容分析失败: {str(e)}")
            return {
                'word_count': len(text),
                'paragraph_count': 1,
                'sentence_count': 1,
                'language': 'zh',
                'contains_sensitive': False,
                'sentiment': 'neutral',
                'readability': 0.5
            }
    
    def _detect_language(self, text: str) -> str:
        """检测语言"""
        # 简单的语言检测
        chinese_chars = len(re.findall(r'[\u4e00-\u9fff]', text))
        english_chars = len(re.findall(r'[a-zA-Z]', text))
        
        if chinese_chars > english_chars:
            return 'zh'
        elif english_chars > chinese_chars:
            return 'en'
        else:
            return 'mixed'
    
    def _analyze_sentiment(self, text: str) -> str:
        """简单的情感分析"""
        # 简化的情感词典
        positive_words = {'好', '棒', '优秀', '喜欢', '开心', '快乐', '满意', '感谢', '谢谢', '爱'}
        negative_words = {'坏', '差', '糟糕', '讨厌', '难过', '痛苦', '生气', '愤怒', '失望', '抱怨'}
        
        text_words = set(jieba.cut(text))
        
        positive_count = len(text_words.intersection(positive_words))
        negative_count = len(text_words.intersection(negative_words))
        
        if positive_count > negative_count:
            return 'positive'
        elif negative_count > positive_count:
            return 'negative'
        else:
            return 'neutral'
    
    def _calculate_readability(self, text: str) -> float:
        """计算可读性得分"""
        try:
            # 简化的可读性计算
            sentences = re.split('[。！？.!?]', text)
            sentences = [s.strip() for s in sentences if s.strip()]
            
            if not sentences:
                return 0.0
            
            # 平均句长
            avg_sentence_length = sum(len(s) for s in sentences) / len(sentences)
            
            # 基于句长的可读性评分（句长适中时可读性较高）
            if 10 <= avg_sentence_length <= 30:
                readability = 0.8
            elif 5 <= avg_sentence_length < 10 or 30 < avg_sentence_length <= 50:
                readability = 0.6
            else:
                readability = 0.4
            
            return readability
            
        except Exception as e:
            logger.warning(f"可读性计算失败: {str(e)}")
            return 0.5
    
    def _load_sensitive_words(self) -> set:
        """加载敏感词库"""
        # 简化的敏感词列表，实际应用中应从文件加载
        return {
            '政治', '反动', '暴力', '恐怖', '色情', '赌博', 
            '毒品', '诈骗', '病毒', '木马', '黑客'
        }
    
    def _load_stop_words(self) -> set:
        """加载停用词"""
        # 简化的停用词列表
        return {
            '的', '了', '在', '是', '我', '有', '和', '就', '不', '人', '都', 
            '一', '一个', '上', '也', '很', '到', '说', '要', '去', '你', 
            '会', '着', '没有', '看', '好', '自己', '这', '那', '他', '她',
            '它', '们', '这个', '那个', '什么', '怎么', '为什么', '哪里'
        }


# 全局文本验证器实例
_text_validator = None

def get_text_validator():
    """获取文本验证器实例（单例模式）"""
    global _text_validator
    if _text_validator is None:
        _text_validator = TextValidator()
    return _text_validator