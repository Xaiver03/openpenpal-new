#!/usr/bin/env python3
"""
Postcode数据库测试脚本
验证数据库初始化和API集成是否正常工作
"""

import asyncio
import aiohttp
import json
import sys
import time
from typing import Dict, Any, List

# 测试配置
API_BASE_URL = "http://localhost:8001"
GATEWAY_URL = "http://localhost:8000"

# 测试用例
TEST_CASES = {
    "postcode_query": [
        "PKA101",  # 北京大学东区1栋101室
        "PKA102",  # 北京大学东区1栋102室
        "THA101",  # 清华大学紫荆区1栋101室
        "INVALID", # 无效编码
    ],
    "address_search": [
        "北京大学",
        "PKA101",
        "101室",
        "东区",
    ],
    "hierarchical_queries": [
        "/api/v1/postcode/schools",
        "/api/v1/postcode/schools/PK/areas",
        "/api/v1/postcode/schools/PK/areas/A/buildings",
        "/api/v1/postcode/schools/PK/areas/A/buildings/1/rooms",
    ]
}

class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    PURPLE = '\033[0;35m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'

def print_colored(text: str, color: str):
    print(f"{color}{text}{Colors.NC}")

def print_result(success: bool, message: str):
    color = Colors.GREEN if success else Colors.RED
    symbol = "✅" if success else "❌"
    print_colored(f"{symbol} {message}", color)

async def test_api_endpoint(session: aiohttp.ClientSession, url: str, method: str = "GET", data: Dict = None) -> Dict[str, Any]:
    """测试API端点"""
    try:
        headers = {"Content-Type": "application/json"} if data else {}
        
        async with session.request(method, url, json=data, headers=headers, timeout=10) as response:
            content = await response.text()
            
            try:
                result = json.loads(content)
            except json.JSONDecodeError:
                result = {"raw_content": content}
            
            return {
                "success": response.status < 400,
                "status_code": response.status,
                "data": result,
                "url": url
            }
    except asyncio.TimeoutError:
        return {
            "success": False,
            "status_code": 408,
            "data": {"error": "Request timeout"},
            "url": url
        }
    except Exception as e:
        return {
            "success": False,
            "status_code": 500,
            "data": {"error": str(e)},
            "url": url
        }

async def test_postcode_queries(session: aiohttp.ClientSession):
    """测试Postcode查询功能"""
    print_colored("\n🔍 测试Postcode查询功能", Colors.BLUE)
    
    for code in TEST_CASES["postcode_query"]:
        url = f"{API_BASE_URL}/api/v1/postcode/{code}"
        result = await test_api_endpoint(session, url)
        
        if code == "INVALID":
            # 无效编码应该返回错误
            success = not result["success"]
            message = f"无效编码 {code} 正确返回错误"
        else:
            success = result["success"]
            if success:
                data = result["data"].get("data", {})
                postcode = data.get("postcode", "")
                hierarchy = data.get("hierarchy", {})
                school_name = hierarchy.get("school", {}).get("name", "")
                message = f"编码 {code} 查询成功 - {school_name}"
            else:
                message = f"编码 {code} 查询失败 - {result['data']}"
        
        print_result(success, message)

async def test_address_search(session: aiohttp.ClientSession):
    """测试地址搜索功能"""
    print_colored("\n🔎 测试地址搜索功能", Colors.BLUE)
    
    for query in TEST_CASES["address_search"]:
        url = f"{API_BASE_URL}/api/v1/address/search?query={query}&limit=3"
        result = await test_api_endpoint(session, url)
        
        if result["success"]:
            data = result["data"].get("data", {})
            results = data.get("results", [])
            total = len(results)
            message = f"搜索 '{query}' 找到 {total} 个结果"
            
            # 显示前几个结果
            for i, item in enumerate(results[:2]):
                postcode = item.get("postcode", "")
                address = item.get("fullAddress", "")
                print_colored(f"    {i+1}. {postcode} - {address}", Colors.CYAN)
        else:
            message = f"搜索 '{query}' 失败 - {result['data']}"
        
        print_result(result["success"], message)

async def test_hierarchical_apis(session: aiohttp.ClientSession):
    """测试层次化API"""
    print_colored("\n🏗️  测试层次化API", Colors.BLUE)
    
    for endpoint in TEST_CASES["hierarchical_queries"]:
        url = f"{API_BASE_URL}{endpoint}"
        result = await test_api_endpoint(session, url)
        
        if result["success"]:
            data = result["data"].get("data", {})
            items = data.get("items", [])
            total = data.get("total", len(items))
            endpoint_name = endpoint.split("/")[-1]
            message = f"{endpoint_name} API 返回 {total} 条记录"
        else:
            message = f"{endpoint} API 失败 - {result.get('status_code', 'Unknown')}"
        
        print_result(result["success"], message)

async def test_permission_system(session: aiohttp.ClientSession):
    """测试权限系统"""
    print_colored("\n🔐 测试权限系统", Colors.BLUE)
    
    courier_ids = ["courier1", "courier2", "courier3", "courier4"]
    
    for courier_id in courier_ids:
        url = f"{API_BASE_URL}/api/v1/postcode/permissions/{courier_id}"
        result = await test_api_endpoint(session, url)
        
        if result["success"]:
            data = result["data"].get("data", {})
            level = data.get("level", 0)
            patterns = data.get("prefix_patterns", [])
            message = f"{courier_id} 权限查询成功 - Level {level}, 模式: {patterns}"
        else:
            message = f"{courier_id} 权限查询失败"
        
        print_result(result["success"], message)

async def test_statistics(session: aiohttp.ClientSession):
    """测试统计功能"""
    print_colored("\n📊 测试统计功能", Colors.BLUE)
    
    url = f"{API_BASE_URL}/api/v1/postcode/stats/popular?limit=5"
    result = await test_api_endpoint(session, url)
    
    if result["success"]:
        data = result["data"].get("data", {})
        items = data.get("items", [])
        message = f"热门地址统计返回 {len(items)} 条记录"
        
        # 显示前几个热门地址
        for i, item in enumerate(items[:3]):
            postcode = item.get("postcode", "")
            count = item.get("delivery_count", 0)
            score = item.get("popularity_score", 0)
            print_colored(f"    {i+1}. {postcode} - 投递{count}次, 评分{score}", Colors.CYAN)
    else:
        message = f"统计API失败 - {result['data']}"
    
    print_result(result["success"], message)

async def test_validation_system(session: aiohttp.ClientSession):
    """测试验证系统"""
    print_colored("\n✅ 测试验证系统", Colors.BLUE)
    
    test_codes = ["PKA101", "PKA102", "INVALID", "THA101", "BADCODE"]
    url = f"{API_BASE_URL}/api/v1/postcode/validate"
    data = {"codes": test_codes}
    
    result = await test_api_endpoint(session, url, method="POST", data=data)
    
    if result["success"]:
        response_data = result["data"].get("data", {})
        total = response_data.get("total", 0)
        valid = response_data.get("valid", 0)
        invalid = response_data.get("invalid", 0)
        message = f"批量验证 {total} 个编码 - {valid} 有效, {invalid} 无效"
        
        # 显示验证结果
        results = response_data.get("results", [])
        for item in results:
            code = item.get("code", "")
            is_valid = item.get("is_valid", False)
            exists = item.get("exists", False)
            status = "✓" if is_valid else "✗"
            exist_status = "存在" if exists else "不存在"
            print_colored(f"    {status} {code} - {exist_status}", Colors.CYAN)
    else:
        message = f"验证API失败 - {result['data']}"
    
    print_result(result["success"], message)

async def check_service_health():
    """检查服务健康状态"""
    print_colored("\n🏥 检查服务健康状态", Colors.BLUE)
    
    services = [
        ("写信服务", f"{API_BASE_URL}/health"),
        ("API网关", f"{GATEWAY_URL}/health"),
    ]
    
    async with aiohttp.ClientSession() as session:
        for name, url in services:
            result = await test_api_endpoint(session, url)
            
            if result["success"]:
                data = result["data"]
                service_name = data.get("service", name)
                status = data.get("status", "unknown")
                message = f"{name} ({service_name}) - {status}"
            else:
                message = f"{name} 无法连接"
            
            print_result(result["success"], message)

async def main():
    """主测试函数"""
    print_colored("🚀 开始Postcode数据库集成测试", Colors.PURPLE)
    print_colored("=" * 50, Colors.PURPLE)
    
    # 检查服务状态
    await check_service_health()
    
    # 等待服务稳定
    print_colored("\n⏳ 等待服务稳定...", Colors.YELLOW)
    await asyncio.sleep(2)
    
    # 执行测试套件
    async with aiohttp.ClientSession() as session:
        await test_postcode_queries(session)
        await test_address_search(session)
        await test_hierarchical_apis(session)
        await test_permission_system(session)
        await test_statistics(session)
        await test_validation_system(session)
    
    print_colored("\n" + "=" * 50, Colors.PURPLE)
    print_colored("🎉 Postcode数据库集成测试完成", Colors.PURPLE)
    
    print_colored("\n💡 下一步:", Colors.YELLOW)
    print_colored("1. 运行数据库初始化: ./scripts/init-postcode-db.sh", Colors.YELLOW)
    print_colored("2. 重启应用服务以连接数据库", Colors.YELLOW)
    print_colored("3. 使用测试账号登录验证前端功能", Colors.YELLOW)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print_colored("\n⚠️  测试被用户中断", Colors.YELLOW)
        sys.exit(1)
    except Exception as e:
        print_colored(f"\n❌ 测试过程中发生错误: {e}", Colors.RED)
        sys.exit(1)