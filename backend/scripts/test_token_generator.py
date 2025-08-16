#!/usr/bin/env python3

"""
🔐 OpenPenPal 安全测试令牌生成器 (Python版本)

功能：为Python测试脚本生成安全的JWT令牌，替代硬编码令牌
安全原则：
1. 使用环境变量存储密钥
2. 生成短期有效令牌
3. 包含明确的测试标识
4. 支持不同角色和权限
"""

import jwt
import os
import secrets
import time
from datetime import datetime, timedelta
from typing import Dict, Any, Optional

# 测试环境专用密钥 (绝不在生产环境使用)
TEST_JWT_SECRET = os.getenv(
    'TEST_JWT_SECRET', 
    'test_secret_for_local_development_only_never_use_in_production_' + secrets.token_hex(16)
)

# 预定义测试用户角色
TEST_ROLES = {
    'ADMIN': {
        'role': 'super_admin',
        'permissions': [
            'MANAGE_USERS', 'VIEW_ANALYTICS', 'MODERATE_CONTENT',
            'MANAGE_SCHOOLS', 'MANAGE_EXHIBITIONS', 'SYSTEM_CONFIG',
            'AUDIT_SUBMISSIONS', 'HANDLE_REPORTS'
        ],
        'user_id': f'test-admin-{secrets.token_hex(4)}'
    },
    'USER': {
        'role': 'user',
        'permissions': ['CREATE_LETTER', 'VIEW_LETTERS', 'PARTICIPATE_ACTIVITIES'],
        'user_id': f'test-user-{secrets.token_hex(4)}'
    },
    'COURIER_L1': {
        'role': 'courier_level1',
        'permissions': ['courier_scan_code', 'courier_deliver_letter', 'courier_view_own_tasks'],
        'user_id': f'test-courier-l1-{secrets.token_hex(4)}'
    },
    'COURIER_L2': {
        'role': 'courier_level2',
        'permissions': ['courier_scan_code', 'courier_deliver_letter', 'courier_manage_subordinates'],
        'user_id': f'test-courier-l2-{secrets.token_hex(4)}'
    },
    'COURIER_L3': {
        'role': 'courier_level3',
        'permissions': ['courier_manage_school_zone', 'courier_view_school_analytics'],
        'user_id': f'test-courier-l3-{secrets.token_hex(4)}'
    },
    'COURIER_L4': {
        'role': 'courier_level4',
        'permissions': ['courier_manage_city_operations', 'courier_view_city_analytics'],
        'user_id': f'test-courier-l4-{secrets.token_hex(4)}'
    }
}


def generate_test_token(
    role_type: str = 'USER',
    custom_payload: Optional[Dict[str, Any]] = None,
    expires_in_hours: int = 2
) -> str:
    """
    生成测试JWT令牌
    
    Args:
        role_type: 角色类型 (ADMIN, USER, COURIER_L1-L4)
        custom_payload: 自定义载荷
        expires_in_hours: 过期时间（小时）
        
    Returns:
        JWT令牌字符串
        
    Raises:
        ValueError: 未知的角色类型
    """
    if role_type not in TEST_ROLES:
        raise ValueError(f"Unknown role type: {role_type}. Available: {list(TEST_ROLES.keys())}")
    
    role_config = TEST_ROLES[role_type]
    now = int(time.time())
    
    payload = {
        # 标准字段
        'userId': role_config['user_id'],
        'username': f'test_{role_config["role"]}',
        'role': role_config['role'],
        'permissions': role_config['permissions'],
        
        # JWT标准声明
        'iss': 'openpenpal-test',  # 明确标识为测试环境
        'aud': 'openpenpal-client',
        'iat': now,
        'exp': now + (expires_in_hours * 3600),
        'jti': secrets.token_hex(16),  # 唯一标识符
        
        # 测试环境标识
        'env': 'test',
        'schoolCode': 'TEST01' if 'COURIER' in role_type else 'PKU001',
    }
    
    # 合并自定义载荷
    if custom_payload:
        payload.update(custom_payload)
    
    return jwt.encode(payload, TEST_JWT_SECRET, algorithm='HS256')


def verify_test_token(token: str) -> Dict[str, Any]:
    """
    验证测试令牌
    
    Args:
        token: JWT令牌
        
    Returns:
        解码后的载荷
        
    Raises:
        jwt.InvalidTokenError: 令牌验证失败
    """
    return jwt.decode(token, TEST_JWT_SECRET, algorithms=['HS256'])


def decode_test_token(token: str) -> Dict[str, Any]:
    """
    解码令牌信息（不验证签名）
    
    Args:
        token: JWT令牌
        
    Returns:
        解码后的载荷
    """
    return jwt.decode(token, options={"verify_signature": False})


def generate_long_lived_token(role_type: str = 'ADMIN') -> str:
    """
    生成长期有效的令牌（用于长时间运行的测试）
    
    Args:
        role_type: 角色类型
        
    Returns:
        长期有效令牌
    """
    return generate_test_token(role_type, expires_in_hours=720)  # 30天


def generate_all_test_tokens() -> Dict[str, str]:
    """
    批量生成测试令牌
    
    Returns:
        包含所有角色的令牌字典
    """
    tokens = {}
    for role_type in TEST_ROLES:
        tokens[role_type.lower()] = generate_test_token(role_type)
    return tokens


# 常用的预生成令牌函数（为了方便导入使用）
def get_admin_token() -> str:
    """获取管理员令牌"""
    return generate_test_token('ADMIN')


def get_user_token() -> str:
    """获取普通用户令牌"""
    return generate_test_token('USER')


def get_courier_token(level: int = 1) -> str:
    """获取信使令牌"""
    role_type = f'COURIER_L{level}'
    return generate_test_token(role_type)


if __name__ == '__main__':
    import sys
    
    print('🔐 OpenPenPal 安全测试令牌生成器 (Python版本)\n')
    
    if len(sys.argv) < 2:
        command = 'admin'
    else:
        command = sys.argv[1].lower()
    
    try:
        if command == 'admin':
            print('管理员令牌:')
            print(generate_test_token('ADMIN'))
            
        elif command == 'user':
            print('普通用户令牌:')
            print(generate_test_token('USER'))
            
        elif command == 'courier':
            level = int(sys.argv[2]) if len(sys.argv) > 2 else 1
            courier_type = f'COURIER_L{level}'
            print(f'{level}级信使令牌:')
            print(generate_test_token(courier_type))
            
        elif command == 'all':
            print('所有测试令牌:')
            all_tokens = generate_all_test_tokens()
            for role, token in all_tokens.items():
                print(f'{role.upper()}:')
                print(f'  {token}\n')
                
        elif command == 'long':
            print('长期有效管理员令牌 (30天):')
            print(generate_long_lived_token('ADMIN'))
            
        elif command == 'verify':
            if len(sys.argv) < 3:
                print('❌ 请提供要验证的令牌')
                sys.exit(1)
            token = sys.argv[2]
            decoded = verify_test_token(token)
            print('令牌验证成功:')
            print(json.dumps(decoded, indent=2, default=str))
            
        else:
            print('使用方法:')
            print('  python test_token_generator.py admin       # 生成管理员令牌')
            print('  python test_token_generator.py user        # 生成用户令牌')
            print('  python test_token_generator.py courier 1   # 生成1级信使令牌')
            print('  python test_token_generator.py all         # 生成所有角色令牌')
            print('  python test_token_generator.py long        # 生成长期令牌')
            print('  python test_token_generator.py verify <token> # 验证令牌')
            
    except Exception as error:
        print(f'❌ 错误: {error}')
        sys.exit(1)