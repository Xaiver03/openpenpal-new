#!/usr/bin/env python3
"""
广场API测试脚本
"""

import requests
import json
import sys
import os
from datetime import datetime

# 🔐 安全令牌生成 - 替代硬编码令牌
sys.path.append(os.path.join(os.path.dirname(__file__), '../../backend/scripts'))
from test_token_generator import get_admin_token

# 配置
BASE_URL = "http://localhost:8001"
PLAZA_API = f"{BASE_URL}/api/plaza"

# 安全生成的测试令牌
TEST_TOKEN = get_admin_token()

def get_headers():
    """获取请求头"""
    return {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {TEST_TOKEN}"
    }

def test_get_categories():
    """测试获取分类列表"""
    print("🧪 测试：获取分类列表")
    
    try:
        response = requests.get(f"{PLAZA_API}/categories")
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 成功获取分类列表")
            print(f"分类数量: {len(data.get('data', {}).get('categories', []))}")
            
            categories = data.get('data', {}).get('categories', [])
            for category in categories[:3]:  # 显示前3个分类
                print(f"  - {category['name']}: {category['description']}")
        else:
            print(f"❌ 获取分类失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
    
    print("-" * 50)

def test_create_post():
    """测试创建帖子"""
    print("🧪 测试：创建广场帖子")
    
    post_data = {
        "title": "测试帖子标题",
        "content": "这是一个测试帖子的内容。包含了一些示例文字，用于测试广场功能是否正常工作。",
        "category": "thoughts",
        "tags": ["测试", "开发", "API"],
        "allow_comments": True,
        "anonymous": False
    }
    
    try:
        response = requests.post(
            f"{PLAZA_API}/posts",
            headers=get_headers(),
            json=post_data
        )
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 成功创建帖子")
            print(f"帖子ID: {data.get('data', {}).get('post_id')}")
            return data.get('data', {}).get('post_id')
        else:
            print(f"❌ 创建帖子失败: {response.text}")
            return None
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
        return None
    
    print("-" * 50)

def test_get_posts():
    """测试获取帖子列表"""
    print("🧪 测试：获取帖子列表")
    
    try:
        # 测试基本列表
        response = requests.get(f"{PLAZA_API}/posts")
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 成功获取帖子列表")
            posts_data = data.get('data', {})
            print(f"帖子总数: {posts_data.get('total', 0)}")
            print(f"当前页: {posts_data.get('page', 1)}")
            
            posts = posts_data.get('posts', [])
            for post in posts[:2]:  # 显示前2个帖子
                print(f"  - {post['title']} ({post['author_nickname']})")
        else:
            print(f"❌ 获取帖子列表失败: {response.text}")
            
        # 测试分类过滤
        print("\n测试分类过滤...")
        response = requests.get(f"{PLAZA_API}/posts?category=thoughts")
        if response.status_code == 200:
            data = response.json()
            filtered_count = data.get('data', {}).get('total', 0)
            print(f"✅ 思想分类帖子数量: {filtered_count}")
        
    except Exception as e:
        print(f"❌ 请求异常: {e}")
    
    print("-" * 50)

def test_get_post_detail(post_id):
    """测试获取帖子详情"""
    if not post_id:
        print("⏩ 跳过帖子详情测试（没有有效的帖子ID）")
        return
        
    print(f"🧪 测试：获取帖子详情 (ID: {post_id})")
    
    try:
        response = requests.get(f"{PLAZA_API}/posts/{post_id}")
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 成功获取帖子详情")
            post_data = data.get('data', {})
            print(f"标题: {post_data.get('title')}")
            print(f"作者: {post_data.get('author_nickname')}")
            print(f"浏览次数: {post_data.get('view_count')}")
            print(f"点赞次数: {post_data.get('like_count')}")
        else:
            print(f"❌ 获取帖子详情失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
    
    print("-" * 50)

def test_like_post(post_id):
    """测试点赞帖子"""
    if not post_id:
        print("⏩ 跳过点赞测试（没有有效的帖子ID）")
        return
        
    print(f"🧪 测试：点赞帖子 (ID: {post_id})")
    
    try:
        response = requests.post(
            f"{PLAZA_API}/posts/{post_id}/like",
            headers=get_headers()
        )
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 点赞操作成功")
            like_data = data.get('data', {})
            print(f"点赞状态: {'已点赞' if like_data.get('liked') else '已取消点赞'}")
            print(f"总点赞数: {like_data.get('like_count')}")
        else:
            print(f"❌ 点赞操作失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
    
    print("-" * 50)

def test_add_comment(post_id):
    """测试添加评论"""
    if not post_id:
        print("⏩ 跳过评论测试（没有有效的帖子ID）")
        return
        
    print(f"🧪 测试：添加评论 (ID: {post_id})")
    
    comment_data = {
        "content": "这是一条测试评论，用于验证评论功能是否正常工作。"
    }
    
    try:
        response = requests.post(
            f"{PLAZA_API}/posts/{post_id}/comments",
            headers=get_headers(),
            json=comment_data
        )
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 评论添加成功")
            comment_data = data.get('data', {})
            print(f"评论ID: {comment_data.get('comment_id')}")
            print(f"帖子总评论数: {comment_data.get('comment_count')}")
            return comment_data.get('comment_id')
        else:
            print(f"❌ 添加评论失败: {response.text}")
            return None
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
        return None
    
    print("-" * 50)

def test_get_comments(post_id):
    """测试获取评论列表"""
    if not post_id:
        print("⏩ 跳过评论列表测试（没有有效的帖子ID）")
        return
        
    print(f"🧪 测试：获取评论列表 (ID: {post_id})")
    
    try:
        response = requests.get(f"{PLAZA_API}/posts/{post_id}/comments")
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 成功获取评论列表")
            comments_data = data.get('data', {})
            print(f"评论总数: {comments_data.get('total', 0)}")
            
            comments = comments_data.get('comments', [])
            for comment in comments[:2]:  # 显示前2条评论
                print(f"  - {comment['user_nickname']}: {comment['content'][:50]}...")
        else:
            print(f"❌ 获取评论列表失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
    
    print("-" * 50)

def test_get_popular_tags():
    """测试获取热门标签"""
    print("🧪 测试：获取热门标签")
    
    try:
        response = requests.get(f"{PLAZA_API}/tags/popular")
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 成功获取热门标签")
            tags_data = data.get('data', {})
            tags = tags_data.get('tags', [])
            print(f"热门标签: {', '.join(tags[:10])}")  # 显示前10个标签
        else:
            print(f"❌ 获取热门标签失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")
    
    print("-" * 50)

def main():
    """主测试函数"""
    print("🚀 开始广场API测试")
    print("=" * 50)
    
    # 检查服务是否运行
    try:
        response = requests.get(f"{BASE_URL}/health")
        if response.status_code != 200:
            print("❌ 写信服务未运行，请先启动服务")
            return
        print("✅ 写信服务运行正常")
        print("-" * 50)
    except Exception as e:
        print(f"❌ 无法连接到写信服务: {e}")
        return
    
    # 执行测试
    test_get_categories()
    post_id = test_create_post()
    test_get_posts()
    test_get_post_detail(post_id)
    test_like_post(post_id)
    test_add_comment(post_id)
    test_get_comments(post_id)
    test_get_popular_tags()
    
    print("🎉 广场API测试完成")
    print("=" * 50)
    
    if post_id:
        print(f"💡 创建的测试帖子ID: {post_id}")
        print("💡 你可以访问 http://localhost:8001/docs 查看完整的API文档")

if __name__ == "__main__":
    main()