"""
测试管理后台API端点
"""
from fastapi import APIRouter, Depends
from typing import Dict, Any
from app.core.auth import get_current_user
from app.core.responses import success_response, error_response

router = APIRouter(prefix="/test", tags=["测试"])

@router.get("/auth")
async def test_auth(current_user: Dict = Depends(get_current_user)):
    """测试认证功能"""
    return success_response(
        data={
            "authenticated": True,
            "user_info": current_user,
            "message": "认证成功"
        },
        message="Authentication test successful"
    )

@router.get("/categories")
async def test_categories():
    """测试分类API - 不需要数据库"""
    mock_categories = [
        {
            "id": "CAT001", 
            "name": "文具用品",
            "parent_id": None,
            "children": [
                {
                    "id": "CAT002",
                    "name": "笔类",
                    "parent_id": "CAT001"
                }
            ]
        }
    ]
    
    return success_response(
        data={
            "tree": mock_categories,
            "total_nodes": 2
        },
        message="Mock categories data"
    )

@router.get("/rbac")
async def test_rbac():
    """测试RBAC系统 - 模拟数据"""
    mock_rbac_data = {
        "user_total": 10,
        "user_active": 8,
        "role_total": 5,
        "role_active": 4,
        "menu_total": 15,
        "menu_active": 12,
        "online_users": 3
    }
    
    return success_response(
        data=mock_rbac_data,
        message="Mock RBAC statistics"
    )

@router.get("/pricing") 
async def test_pricing():
    """测试价格管理API - 模拟数据"""
    mock_pricing_data = {
        "policies": [
            {
                "policy_id": 1,
                "policy_name": "基础定价",
                "policy_code": "BASE_PRICING",
                "is_active": True
            }
        ],
        "total": 1
    }
    
    return success_response(
        data=mock_pricing_data,
        message="Mock pricing policies"
    )