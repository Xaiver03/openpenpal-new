#!/usr/bin/env python3
"""
OpenPenPal Write Service API 测试脚本
"""

import json
import requests
import sys
from datetime import datetime

# 服务配置
BASE_URL = "http://localhost:8001"
API_BASE = f"{BASE_URL}/api/letters"

def test_health_check():
    """测试健康检查接口"""
    print("🔍 Testing health check...")
    try:
        response = requests.get(f"{BASE_URL}/health")
        if response.status_code == 200:
            data = response.json()
            print(f"✅ Health check passed: {data['msg']}")
            return True
        else:
            print(f"❌ Health check failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Health check error: {e}")
        return False

def test_create_letter():
    """测试创建信件接口（需要JWT token）"""
    print("\n📝 Testing create letter...")
    
    # 模拟JWT token（实际应该使用真实token）
    headers = {
        "Authorization": "Bearer test-token-user123",
        "Content-Type": "application/json"
    }
    
    letter_data = {
        "title": f"测试信件 - {datetime.now().strftime('%Y%m%d_%H%M%S')}",
        "content": "这是一封通过API测试创建的信件内容。",
        "receiver_hint": "测试地址 - 北京大学",
        "anonymous": False,
        "priority": "normal",
        "delivery_instructions": "请投递到指定地址"
    }
    
    try:
        response = requests.post(API_BASE, headers=headers, json=letter_data)
        if response.status_code == 200:
            data = response.json()
            if data['code'] == 0:
                letter_id = data['data']['letter_id']
                print(f"✅ Letter created successfully: {letter_id}")
                return letter_id
            else:
                print(f"❌ Letter creation failed: {data['msg']}")
                return None
        else:
            print(f"❌ HTTP error: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"❌ Create letter error: {e}")
        return None

def test_get_letter(letter_id):
    """测试获取信件详情"""
    if not letter_id:
        print("⏭️  Skipping get letter test (no letter_id)")
        return False
    
    print(f"\n📖 Testing get letter: {letter_id}")
    
    headers = {
        "Authorization": "Bearer test-token-user123"
    }
    
    try:
        response = requests.get(f"{API_BASE}/{letter_id}", headers=headers)
        if response.status_code == 200:
            data = response.json()
            if data['code'] == 0:
                print(f"✅ Letter retrieved: {data['data']['title']}")
                return True
            else:
                print(f"❌ Get letter failed: {data['msg']}")
                return False
        else:
            print(f"❌ HTTP error: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"❌ Get letter error: {e}")
        return False

def test_read_letter_by_code():
    """测试通过编号读取信件（公开接口）"""
    print("\n🔍 Testing read letter by code...")
    
    # 使用测试编号
    test_code = "OP1234567890"
    
    try:
        response = requests.get(f"{API_BASE}/read/{test_code}")
        if response.status_code == 200:
            data = response.json()
            if data['code'] == 0:
                print(f"✅ Letter read by code: {data['data']['title']}")
                return True
            else:
                print(f"❌ Read letter failed: {data['msg']}")
                return False
        elif response.status_code == 404:
            print(f"⚠️  Test letter not found (expected for fresh setup)")
            return True
        else:
            print(f"❌ HTTP error: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"❌ Read letter error: {e}")
        return False

def test_api_docs():
    """测试API文档访问"""
    print("\n📚 Testing API documentation...")
    
    try:
        # 测试 Swagger UI
        response = requests.get(f"{BASE_URL}/docs")
        if response.status_code == 200:
            print("✅ Swagger UI accessible")
        else:
            print(f"❌ Swagger UI error: {response.status_code}")
        
        # 测试 ReDoc
        response = requests.get(f"{BASE_URL}/redoc")
        if response.status_code == 200:
            print("✅ ReDoc accessible")
            return True
        else:
            print(f"❌ ReDoc error: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ API docs error: {e}")
        return False

def main():
    """主测试函数"""
    print("🚀 OpenPenPal Write Service API 测试")
    print("=" * 50)
    
    # 测试计数
    total_tests = 0
    passed_tests = 0
    
    # 1. 健康检查
    total_tests += 1
    if test_health_check():
        passed_tests += 1
    
    # 2. 创建信件
    total_tests += 1
    letter_id = test_create_letter()
    if letter_id:
        passed_tests += 1
    
    # 3. 获取信件详情
    total_tests += 1
    if test_get_letter(letter_id):
        passed_tests += 1
    
    # 4. 通过编号读取信件
    total_tests += 1
    if test_read_letter_by_code():
        passed_tests += 1
    
    # 5. API文档
    total_tests += 1
    if test_api_docs():
        passed_tests += 1
    
    # 测试结果
    print("\n" + "=" * 50)
    print(f"📊 测试结果: {passed_tests}/{total_tests} 通过")
    
    if passed_tests == total_tests:
        print("🎉 所有测试通过！")
        return 0
    else:
        print("⚠️  部分测试失败，请检查服务状态")
        return 1

if __name__ == "__main__":
    sys.exit(main())