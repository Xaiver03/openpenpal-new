#!/usr/bin/env python3
"""
商城后台管理系统 - 综合测试脚本
"""
import requests
import json
import sys
from datetime import datetime

BASE_URL = "http://localhost:8001"

def test_service_health():
    """测试服务健康状态"""
    print("🏥 Testing service health...")
    try:
        response = requests.get(f"{BASE_URL}/health")
        if response.status_code == 200:
            data = response.json()
            print(f"✅ Service is healthy")
            print(f"   Security Score: {data['data']['security_score']}")
            print(f"   Status: {data['data']['status']}")
            return True
        else:
            print(f"❌ Health check failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Health check error: {e}")
        return False

def test_api_docs():
    """测试API文档访问"""
    print("\n📚 Testing API documentation...")
    try:
        response = requests.get(f"{BASE_URL}/docs")
        if response.status_code == 200:
            print("✅ API documentation accessible")
            print(f"   Swagger UI: {BASE_URL}/docs")
            print(f"   ReDoc: {BASE_URL}/redoc")
            return True
        else:
            print(f"❌ API docs failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ API docs error: {e}")
        return False

def test_categories_api():
    """测试分类管理API"""
    print("\n📂 Testing Categories API...")
    try:
        response = requests.get(f"{BASE_URL}/api/v1/test/categories")
        if response.status_code == 200:
            data = response.json()
            print("✅ Categories API working")
            print(f"   Categories found: {data['data']['total_nodes']}")
            print(f"   Sample category: {data['data']['tree'][0]['name']}")
            return True
        else:
            print(f"❌ Categories API failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Categories API error: {e}")
        return False

def test_rbac_api():
    """测试RBAC权限API"""
    print("\n🔐 Testing RBAC API...")
    try:
        response = requests.get(f"{BASE_URL}/api/v1/test/rbac")
        if response.status_code == 200:
            data = response.json()
            print("✅ RBAC API working")
            print(f"   Total users: {data['data']['user_total']}")
            print(f"   Active users: {data['data']['user_active']}")
            print(f"   Total roles: {data['data']['role_total']}")
            print(f"   Online users: {data['data']['online_users']}")
            return True
        else:
            print(f"❌ RBAC API failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ RBAC API error: {e}")
        return False

def test_pricing_api():
    """测试价格管理API"""
    print("\n💰 Testing Pricing API...")
    try:
        response = requests.get(f"{BASE_URL}/api/v1/test/pricing")
        if response.status_code == 200:
            data = response.json()
            print("✅ Pricing API working")
            print(f"   Policies found: {data['data']['total']}")
            print(f"   Sample policy: {data['data']['policies'][0]['policy_name']}")
            return True
        else:
            print(f"❌ Pricing API failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Pricing API error: {e}")
        return False

def test_all_apis():
    """运行所有API测试"""
    print("🚀 OpenPenPal Mall Admin System - Comprehensive Test")
    print("=" * 60)
    print(f"Test started at: {datetime.now()}")
    print(f"Service URL: {BASE_URL}")
    print("=" * 60)
    
    tests = [
        ("Service Health", test_service_health),
        ("API Documentation", test_api_docs), 
        ("Categories Management", test_categories_api),
        ("RBAC Permissions", test_rbac_api),
        ("Pricing Management", test_pricing_api)
    ]
    
    results = []
    for test_name, test_func in tests:
        result = test_func()
        results.append((test_name, result))
    
    print("\n" + "=" * 60)
    print("📊 TEST SUMMARY")
    print("=" * 60)
    
    passed = 0
    failed = 0
    
    for test_name, result in results:
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"{status} - {test_name}")
        if result:
            passed += 1
        else:
            failed += 1
    
    print(f"\nTotal Tests: {len(results)}")
    print(f"Passed: {passed}")
    print(f"Failed: {failed}")
    print(f"Success Rate: {(passed/len(results)*100):.1f}%")
    
    if failed == 0:
        print("\n🎉 ALL TESTS PASSED! Mall Admin System is ready!")
    else:
        print(f"\n⚠️  {failed} test(s) failed. Please check the service status.")
    
    return failed == 0

def show_quick_start_info():
    """显示快速开始信息"""
    print("\n" + "=" * 60)
    print("🎯 QUICK START INFORMATION")
    print("=" * 60)
    print(f"Service URL: {BASE_URL}")
    print(f"API Docs: {BASE_URL}/docs")
    print(f"Health Check: {BASE_URL}/health")
    print("\n📋 Available Test Endpoints:")
    print(f"  Categories: {BASE_URL}/api/v1/test/categories")
    print(f"  RBAC Stats: {BASE_URL}/api/v1/test/rbac") 
    print(f"  Pricing: {BASE_URL}/api/v1/test/pricing")
    print("\n📚 Documentation:")
    print("  - MALL_ADMIN_API.md - Complete API documentation")
    print("  - MALL_ADMIN_QUICK_START.md - Quick start guide")
    print("\n🚀 Next Steps:")
    print("  1. Develop frontend admin interface")
    print("  2. Implement real database operations")
    print("  3. Add complete CRUD functionality")
    print("  4. Integrate with user service")

if __name__ == "__main__":
    try:
        success = test_all_apis()
        show_quick_start_info()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        print("\n⚠️  Test interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n💥 Unexpected error: {e}")
        sys.exit(1)