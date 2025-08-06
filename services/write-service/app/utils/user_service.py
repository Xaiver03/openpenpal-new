import asyncio
import aiohttp
import logging
from typing import Optional, Dict, Any
from app.core.config import settings

logger = logging.getLogger(__name__)

class UserServiceClient:
    """用户服务客户端 - 用于获取用户信息"""
    
    def __init__(self):
        # 用户服务的基础URL（可以从环境变量配置）
        self.base_url = getattr(settings, 'user_service_url', 'http://localhost:8080/api/v1')
        self.timeout = aiohttp.ClientTimeout(total=5)  # 5秒超时
    
    async def get_user_info(self, user_id: str, jwt_token: Optional[str] = None) -> Dict[str, Any]:
        """
        获取用户基本信息
        
        Args:
            user_id: 用户ID
            jwt_token: JWT token（可选，用于认证）
            
        Returns:
            Dict[str, Any]: 用户信息
        """
        url = f"{self.base_url}/users/{user_id}"
        headers = {}
        
        if jwt_token:
            headers["Authorization"] = f"Bearer {jwt_token}"
        
        try:
            async with aiohttp.ClientSession(timeout=self.timeout) as session:
                async with session.get(url, headers=headers) as response:
                    if response.status == 200:
                        data = await response.json()
                        if data.get('code') == 0:
                            return data.get('data', {})
                        else:
                            logger.warning(f"User service returned error: {data.get('msg')}")
                            return {}
                    else:
                        logger.warning(f"User service HTTP error: {response.status}")
                        return {}
        except asyncio.TimeoutError:
            logger.warning(f"User service timeout for user {user_id}")
            return {}
        except Exception as e:
            logger.error(f"User service error for user {user_id}: {e}")
            return {}
    
    async def get_user_nickname(self, user_id: str, jwt_token: Optional[str] = None) -> str:
        """
        获取用户昵称
        
        Args:
            user_id: 用户ID
            jwt_token: JWT token（可选）
            
        Returns:
            str: 用户昵称，如果获取失败则返回默认值
        """
        user_info = await self.get_user_info(user_id, jwt_token)
        
        # 尝试多个可能的昵称字段
        nickname = (
            user_info.get('nickname') or 
            user_info.get('display_name') or 
            user_info.get('username') or 
            f"用户{user_id[-4:]}"  # 默认昵称：用户+ID后4位
        )
        
        return nickname
    
    async def batch_get_user_nicknames(self, user_ids: list[str], jwt_token: Optional[str] = None) -> Dict[str, str]:
        """
        批量获取用户昵称
        
        Args:
            user_ids: 用户ID列表
            jwt_token: JWT token（可选）
            
        Returns:
            Dict[str, str]: 用户ID到昵称的映射
        """
        if not user_ids:
            return {}
        
        # 并发获取用户信息
        tasks = [self.get_user_nickname(user_id, jwt_token) for user_id in user_ids]
        nicknames = await asyncio.gather(*tasks, return_exceptions=True)
        
        result = {}
        for user_id, nickname in zip(user_ids, nicknames):
            if isinstance(nickname, Exception):
                logger.error(f"Failed to get nickname for user {user_id}: {nickname}")
                result[user_id] = f"用户{user_id[-4:]}"
            else:
                result[user_id] = nickname
        
        return result

# 全局用户服务客户端实例
user_service = UserServiceClient()

# 便捷函数
async def get_user_nickname(user_id: str, jwt_token: Optional[str] = None) -> str:
    """获取用户昵称的便捷函数"""
    return await user_service.get_user_nickname(user_id, jwt_token)

async def get_user_info(user_id: str, jwt_token: Optional[str] = None) -> Dict[str, Any]:
    """获取用户信息的便捷函数"""
    return await user_service.get_user_info(user_id, jwt_token)