#!/usr/bin/env python3
"""
阅读分析API测试脚本
"""
import requests
import json
from datetime import datetime, timedelta

# 服务配置
BASE_URL = "http://localhost:8001"
TEST_TOKEN = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.test"  # 测试用token

def make_request(method, endpoint, data=None, params=None):
    """发送HTTP请求"""
    url = f"{BASE_URL}{endpoint}"
    headers = {
        "Authorization": f"Bearer {TEST_TOKEN}",
        "Content-Type": "application/json"
    }
    
    try:
        if method == "GET":
            response = requests.get(url, headers=headers, params=params)
        elif method == "POST":
            response = requests.post(url, headers=headers, json=data)
        else:
            raise ValueError(f"Unsupported method: {method}")
        
        print(f"\n{method} {endpoint}")
        print(f"Status: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print(f"Response: {json.dumps(result, indent=2, ensure_ascii=False)}")
            return result
        else:
            print(f"Error: {response.text}")
            return None
            
    except requests.exceptions.ConnectionError:
        print(f"❌ 无法连接到服务器 {BASE_URL}")
        return None
    except Exception as e:
        print(f"❌ 请求失败: {e}")
        return None

def test_analytics_endpoints():
    """测试分析API端点"""
    print("🧪 测试阅读分析API")
    print("=" * 50)
    
    # 1. 测试健康检查
    print("\n1. 测试分析服务健康检查")
    make_request("GET", "/api/analytics/health")
    
    # 2. 测试阅读统计
    print("\n2. 测试阅读统计")
    params = {
        "time_range": "day",
        "letter_id": None,
        "user_id": None
    }
    make_request("GET", "/api/analytics/reading-stats", params=params)
    
    # 3. 测试趋势分析
    print("\n3. 测试趋势分析")
    params = {
        "time_range": "week"
    }
    make_request("GET", "/api/analytics/trends", params=params)
    
    # 4. 测试热门内容
    print("\n4. 测试热门内容")
    params = {
        "limit": 5,
        "time_range": "week"
    }
    make_request("GET", "/api/analytics/popular", params=params)
    
    # 5. 测试实时统计
    print("\n5. 测试实时统计")
    make_request("GET", "/api/analytics/realtime")
    
    # 6. 测试仪表板数据
    print("\n6. 测试仪表板数据")
    params = {
        "time_range": "week"
    }
    make_request("GET", "/api/analytics/dashboard", params=params)
    
    # 7. 测试信件详细分析（需要有效的letter_id）
    print("\n7. 测试信件详细分析")
    make_request("GET", "/api/analytics/letter/OP1K2L3M4N5O/analytics")
    
    # 8. 测试用户行为分析
    print("\n8. 测试用户行为分析")
    params = {
        "time_range": "month"
    }
    make_request("GET", "/api/analytics/user/test_user_123/behavior", params=params)
    
    # 9. 测试对比分析
    print("\n9. 测试信件对比分析")
    data = {
        "letter_ids": ["OP1K2L3M4N5O", "OP2K2L3M4N5P"],
        "metrics": ["reads", "duration", "completion_rate"]
    }
    make_request("POST", "/api/analytics/compare", data=data)
    
    # 10. 测试数据导出
    print("\n10. 测试数据导出")
    data = {
        "data_type": "reading_stats",
        "format": "json",
        "include_raw_data": False,
        "time_range": "week"
    }
    make_request("POST", "/api/analytics/export", data=data)

def test_service_health():
    """测试服务健康状态"""
    print("\n🔍 检查写信服务状态")
    print("=" * 50)
    
    # 检查主服务健康
    make_request("GET", "/health")
    
    # 检查分析服务健康
    make_request("GET", "/api/analytics/health")

if __name__ == "__main__":
    print("🚀 OpenPenPal 阅读分析API测试")
    print("=" * 50)
    
    # 首先检查服务健康状态
    test_service_health()
    
    # 测试分析端点
    test_analytics_endpoints()
    
    print("\n✅ 测试完成")
    print("=" * 50)
    print("📖 API文档: http://localhost:8001/docs")
    print("🔧 ReDoc文档: http://localhost:8001/redoc")