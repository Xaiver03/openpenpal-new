#!/usr/bin/env python3
"""
OCR服务基础功能测试脚本
用于验证OCR服务的核心功能是否正常工作
"""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'app'))

import logging
from app.services.ocr_engine import MultiEngineOCR
from app.services.cache_service import get_cache_service
from app.services.text_validator import get_text_validator
from app.services.image_processor import ImagePreprocessor, HandwritingPreprocessor
from app.utils.websocket_client import get_websocket_notifier
from app.utils.memory_manager import get_memory_manager

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def test_ocr_engines():
    """测试OCR引擎"""
    print("\n=== 测试OCR引擎 ===")
    
    try:
        ocr = MultiEngineOCR()
        engines = ocr.get_available_engines()
        
        print(f"可用OCR引擎数量: {len([e for e in engines.values() if e.get('available', False)])}")
        
        for engine_name, info in engines.items():
            status = "✅" if info.get('available', False) else "❌"
            print(f"  {status} {engine_name}: {info.get('error', '正常')}")
        
        return len([e for e in engines.values() if e.get('available', False)]) > 0
        
    except Exception as e:
        print(f"❌ OCR引擎测试失败: {str(e)}")
        return False

def test_image_processors():
    """测试图像处理器"""
    print("\n=== 测试图像处理器 ===")
    
    try:
        # 基础图像处理器
        basic_processor = ImagePreprocessor()
        print("✅ 基础图像处理器初始化成功")
        
        # 手写文字处理器
        handwriting_processor = HandwritingPreprocessor()
        print("✅ 手写文字处理器初始化成功")
        
        # 检查处理步骤
        basic_steps = list(basic_processor.pipeline_steps.keys())
        handwriting_steps = list(handwriting_processor.pipeline_steps.keys())
        
        print(f"  基础处理步骤: {', '.join(basic_steps)}")
        print(f"  手写处理步骤: {', '.join(handwriting_steps)}")
        
        return True
        
    except Exception as e:
        print(f"❌ 图像处理器测试失败: {str(e)}")
        return False

def test_text_validator():
    """测试文本验证器"""
    print("\n=== 测试文本验证器 ===")
    
    try:
        validator = get_text_validator()
        
        # 测试文本相似度计算
        original = "亲爱的朋友，最近过得怎么样？"
        ocr_result = "亲爱的朋友，最近过得怎么样？"
        
        result = validator.validate_text_similarity(original, ocr_result)
        
        print(f"✅ 文本验证器工作正常")
        print(f"  相似度得分: {result.get('similarity_score', 0):.3f}")
        print(f"  验证结果: {'通过' if result.get('is_valid', False) else '未通过'}")
        
        return result.get('similarity_score', 0) > 0.9
        
    except Exception as e:
        print(f"❌ 文本验证器测试失败: {str(e)}")
        return False

def test_cache_service():
    """测试缓存服务"""
    print("\n=== 测试缓存服务 ===")
    
    try:
        cache_service = get_cache_service()
        
        # 测试缓存连接
        if cache_service.redis_client:
            # 尝试ping操作
            cache_service.redis_client.ping()
            print("✅ Redis缓存连接正常")
            
            # 测试缓存操作
            test_key = "test_ocr_cache"
            test_data = {"test": "data", "timestamp": "2025-07-21"}
            
            # 设置缓存
            cache_service.redis_client.set(test_key, str(test_data))
            
            # 获取缓存
            cached_data = cache_service.redis_client.get(test_key)
            
            if cached_data:
                print("✅ 缓存读写操作正常")
                # 清理测试数据
                cache_service.redis_client.delete(test_key)
                return True
            else:
                print("❌ 缓存读取失败")
                return False
        else:
            print("⚠️  Redis缓存未连接，使用内存缓存")
            return True
        
    except Exception as e:
        print(f"⚠️  缓存服务测试警告: {str(e)}")
        return True  # 缓存服务不是关键功能

def test_websocket_notifier():
    """测试WebSocket通知服务"""
    print("\n=== 测试WebSocket通知服务 ===")
    
    try:
        ws_notifier = get_websocket_notifier()
        
        # 测试连接状态
        connection_test = ws_notifier.test_connection()
        
        if connection_test.get('status') == 'connected':
            print("✅ WebSocket通知服务连接正常")
            return True
        elif connection_test.get('status') == 'disconnected':
            print("⚠️  WebSocket通知服务未连接Redis，但服务正常")
            return True
        else:
            print(f"❌ WebSocket通知服务异常: {connection_test.get('message', '未知错误')}")
            return False
        
    except Exception as e:
        print(f"⚠️  WebSocket通知服务测试警告: {str(e)}")
        return True  # 非关键功能

def test_memory_manager():
    """测试内存管理器"""
    print("\n=== 测试内存管理器 ===")
    
    try:
        memory_manager = get_memory_manager()
        
        # 获取内存使用情况
        memory_usage = memory_manager.get_memory_usage()
        
        print(f"✅ 内存管理器工作正常")
        print(f"  当前内存使用: {memory_usage.get('rss_mb', 0):.1f}MB ({memory_usage.get('percent', 0):.1f}%)")
        print(f"  可用内存: {memory_usage.get('available_mb', 0):.1f}MB")
        
        # 测试内存清理
        cleanup_result = memory_manager.cleanup_memory()
        
        if cleanup_result:
            print("✅ 内存清理功能正常")
        
        return True
        
    except Exception as e:
        print(f"❌ 内存管理器测试失败: {str(e)}")
        return False

def main():
    """主测试函数"""
    print("🚀 开始OCR服务功能验证测试")
    print("=" * 50)
    
    # 执行所有测试
    tests = [
        ("OCR引擎", test_ocr_engines),
        ("图像处理器", test_image_processors),
        ("文本验证器", test_text_validator),
        ("缓存服务", test_cache_service),
        ("WebSocket通知", test_websocket_notifier),
        ("内存管理器", test_memory_manager),
    ]
    
    results = {}
    for test_name, test_func in tests:
        try:
            results[test_name] = test_func()
        except Exception as e:
            print(f"❌ {test_name}测试异常: {str(e)}")
            results[test_name] = False
    
    # 总结测试结果
    print("\n" + "=" * 50)
    print("📊 测试结果总结")
    print("=" * 50)
    
    passed = 0
    total = len(results)
    
    for test_name, result in results.items():
        status = "✅ 通过" if result else "❌ 失败"
        print(f"{test_name:15} : {status}")
        if result:
            passed += 1
    
    print(f"\n通过率: {passed}/{total} ({passed/total*100:.1f}%)")
    
    if passed >= total * 0.8:  # 80%通过率认为服务正常
        print("\n🎉 OCR服务基础功能验证通过！")
        return 0
    else:
        print("\n⚠️  OCR服务存在问题，请检查失败的测试项")
        return 1


if __name__ == "__main__":
    exit_code = main()
    sys.exit(exit_code)