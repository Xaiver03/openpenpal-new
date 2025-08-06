#!/usr/bin/env python3
"""
批量操作API测试脚本

测试所有批量操作功能，包括：
1. 批量删除信件
2. 批量更新信件状态
3. 批量归档/恢复
4. 批量导出
5. 批量创建
6. 作业状态查询
"""

import asyncio
import aiohttp
import json
from datetime import datetime
from typing import List, Dict, Any


class BatchOperationTester:
    """批量操作测试客户端"""
    
    def __init__(self, base_url: str = "http://localhost:8001", auth_token: str = None):
        self.base_url = base_url.rstrip('/')
        self.auth_token = auth_token
        self.session = None
    
    async def __aenter__(self):
        self.session = aiohttp.ClientSession(
            headers={
                "Authorization": f"Bearer {self.auth_token}" if self.auth_token else None,
                "Content-Type": "application/json"
            }
        )
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    async def test_health_check(self):
        """测试批量操作服务健康检查"""
        print("\n🩺 测试批量操作服务健康检查...")
        
        try:
            async with self.session.get(f"{self.base_url}/api/batch/health") as response:
                if response.status == 200:
                    data = await response.json()
                    print(f"✅ 服务健康检查通过: {data['data']['status']}")
                    print(f"   支持的操作: {', '.join(data['data']['supported_operations'])}")
                    print(f"   支持的目标: {', '.join(data['data']['supported_targets'])}")
                    print(f"   活跃作业数: {data['data']['active_jobs']}")
                    return True
                else:
                    print(f"❌ 健康检查失败: {response.status}")
                    return False
        except Exception as e:
            print(f"❌ 健康检查异常: {e}")
            return False
    
    async def test_batch_validation(self):
        """测试批量操作验证（试运行）"""
        print("\n🔍 测试批量操作验证（试运行）...")
        
        validation_request = {
            "operation": "delete",
            "target_type": "letters", 
            "target_ids": ["TEST001", "TEST002"],
            "operation_data": {
                "soft_delete": True,
                "delete_reason": "测试删除"
            },
            "dry_run": True  # 试运行模式
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/api/batch/validate",
                json=validation_request
            ) as response:
                data = await response.json()
                if response.status == 200:
                    print(f"✅ 批量操作验证成功")
                    print(f"   验证结果: {data['msg']}")
                    if 'data' in data:
                        result = data['data']
                        print(f"   操作ID: {result.get('operation_id')}")
                        print(f"   预计处理: {result.get('total_count')} 个项目")
                    return True
                else:
                    print(f"❌ 验证失败: {data.get('detail', 'Unknown error')}")
                    return False
        except Exception as e:
            print(f"❌ 验证异常: {e}")
            return False
    
    async def test_batch_delete_letters(self):
        """测试批量删除信件"""
        print("\n🗑️ 测试批量删除信件...")
        
        delete_request = {
            "target_ids": ["LETTER001", "LETTER002", "LETTER003"],
            "soft_delete": True,
            "delete_reason": "批量清理测试信件"
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/api/batch/letters/delete",
                json=delete_request
            ) as response:
                data = await response.json()
                if response.status == 200:
                    print(f"✅ 批量删除信件成功: {data['msg']}")
                    result = data['data']
                    print(f"   操作ID: {result['operation_id']}")
                    print(f"   成功/总数: {result['success_count']}/{result['total_count']}")
                    return result['operation_id']
                else:
                    print(f"❌ 批量删除失败: {data.get('detail', 'Unknown error')}")
                    return None
        except Exception as e:
            print(f"❌ 批量删除异常: {e}")
            return None
    
    async def test_batch_status_update(self):
        """测试批量状态更新"""
        print("\n🔄 测试批量状态更新...")
        
        status_update_request = {
            "target_ids": ["LETTER004", "LETTER005"], 
            "new_status": "generated",
            "reason": "批量生成信件编码",
            "force": False
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/api/batch/letters/status",
                json=status_update_request
            ) as response:
                data = await response.json()
                if response.status == 200:
                    print(f"✅ 批量状态更新成功: {data['msg']}")
                    result = data['data']
                    print(f"   操作ID: {result['operation_id']}")
                    print(f"   成功/总数: {result['success_count']}/{result['total_count']}")
                    return result['operation_id']
                else:
                    print(f"❌ 批量状态更新失败: {data.get('detail', 'Unknown error')}")
                    return None
        except Exception as e:
            print(f"❌ 批量状态更新异常: {e}")
            return None
    
    async def test_batch_export(self):
        """测试批量导出"""
        print("\n📤 测试批量导出...")
        
        export_request = {
            "target_ids": ["LETTER001", "LETTER002"],
            "export_format": "json",
            "include_fields": ["id", "title", "status", "created_at"],
            "exclude_fields": ["content"]  # 排除敏感内容
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/api/batch/export?target_type=letters",
                json=export_request
            ) as response:
                data = await response.json()
                if response.status == 200:
                    print(f"✅ 批量导出成功: {data['msg']}")
                    result = data['data']
                    print(f"   操作ID: {result['operation_id']}")
                    print(f"   成功/总数: {result['success_count']}/{result['total_count']}")
                    
                    # 检查导出结果
                    if result['results']:
                        export_id = result['results'][0].get('data', {}).get('export_id')
                        if export_id:
                            await self.test_download_export(export_id)
                    
                    return result['operation_id']
                else:
                    print(f"❌ 批量导出失败: {data.get('detail', 'Unknown error')}")
                    return None
        except Exception as e:
            print(f"❌ 批量导出异常: {e}")
            return None
    
    async def test_download_export(self, export_id: str):
        """测试下载导出文件"""
        print(f"\n📥 测试下载导出文件 {export_id}...")
        
        try:
            async with self.session.get(
                f"{self.base_url}/api/batch/export/{export_id}"
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    print(f"✅ 导出文件下载成功")
                    print(f"   格式: {data.get('format')}")
                    print(f"   记录数: {len(data.get('data', []))}")
                    print(f"   创建时间: {data.get('created_at')}")
                    return True
                else:
                    data = await response.json()
                    print(f"❌ 下载导出文件失败: {data.get('detail', 'Unknown error')}")
                    return False
        except Exception as e:
            print(f"❌ 下载导出文件异常: {e}")
            return False
    
    async def test_batch_archive_restore(self):
        """测试批量归档和恢复"""
        print("\n📦 测试批量归档...")
        
        archive_request = {
            "target_ids": ["LETTER006", "LETTER007"],
            "archive_reason": "定期归档旧信件",
            "archive_location": "archive/2024/letters"
        }
        
        try:
            # 先归档
            async with self.session.post(
                f"{self.base_url}/api/batch/archive?target_type=letters",
                json=archive_request
            ) as response:
                data = await response.json()
                if response.status == 200:
                    print(f"✅ 批量归档成功: {data['msg']}")
                    archive_operation_id = data['data']['operation_id']
                    
                    # 然后恢复
                    print("\n📤 测试批量恢复...")
                    restore_data = {
                        "target_ids": ["LETTER006", "LETTER007"]
                    }
                    
                    async with self.session.post(
                        f"{self.base_url}/api/batch/restore?target_type=letters",
                        json=restore_data
                    ) as restore_response:
                        restore_data = await restore_response.json()
                        if restore_response.status == 200:
                            print(f"✅ 批量恢复成功: {restore_data['msg']}")
                            return restore_data['data']['operation_id']
                        else:
                            print(f"❌ 批量恢复失败: {restore_data.get('detail', 'Unknown error')}")
                            
                    return archive_operation_id
                else:
                    print(f"❌ 批量归档失败: {data.get('detail', 'Unknown error')}")
                    return None
        except Exception as e:
            print(f"❌ 批量归档/恢复异常: {e}")
            return None
    
    async def test_batch_create(self):
        """测试批量创建"""
        print("\n📝 测试批量创建...")
        
        create_request = {
            "items": [
                {
                    "title": "批量创建测试信件1",
                    "content": "这是批量创建的测试信件内容",
                    "anonymous": False,
                    "priority": "normal"
                },
                {
                    "title": "批量创建测试信件2", 
                    "content": "这是另一封批量创建的测试信件",
                    "anonymous": True,
                    "priority": "urgent"
                }
            ],
            "skip_validation": False,
            "continue_on_error": True
        }
        
        try:
            async with self.session.post(
                f"{self.base_url}/api/batch/bulk-create?target_type=letters",
                json=create_request
            ) as response:
                data = await response.json()
                if response.status == 200:
                    print(f"✅ 批量创建成功: {data['msg']}")
                    result = data['data']
                    print(f"   操作ID: {result['operation_id']}")
                    print(f"   成功/总数: {result['success_count']}/{result['total_count']}")
                    
                    # 显示创建的ID
                    created_ids = [
                        r.get('data', {}).get('created_id') 
                        for r in result['results'] 
                        if r['success']
                    ]
                    if created_ids:
                        print(f"   创建的ID: {', '.join(filter(None, created_ids))}")
                    
                    return result['operation_id']
                else:
                    print(f"❌ 批量创建失败: {data.get('detail', 'Unknown error')}")
                    return None
        except Exception as e:
            print(f"❌ 批量创建异常: {e}")
            return None
    
    async def test_job_status(self, job_id: str):
        """测试作业状态查询"""
        print(f"\n📊 测试作业状态查询 {job_id}...")
        
        try:
            async with self.session.get(
                f"{self.base_url}/api/batch/jobs/{job_id}"
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    job_data = data['data']
                    print(f"✅ 作业状态查询成功")
                    print(f"   作业ID: {job_data['job_id']}")
                    print(f"   状态: {job_data['status']}")
                    print(f"   进度: {job_data['progress']:.1f}%")
                    print(f"   创建时间: {job_data['created_at']}")
                    print(f"   更新时间: {job_data['updated_at']}")
                    
                    if job_data.get('error_message'):
                        print(f"   错误信息: {job_data['error_message']}")
                    
                    return True
                else:
                    data = await response.json()
                    print(f"❌ 作业状态查询失败: {data.get('detail', 'Unknown error')}")
                    return False
        except Exception as e:
            print(f"❌ 作业状态查询异常: {e}")
            return False
    
    async def test_admin_functions(self):
        """测试管理员功能"""
        print("\n👑 测试管理员功能...")
        
        try:
            # 获取所有作业状态
            async with self.session.get(
                f"{self.base_url}/api/batch/admin/jobs"
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    jobs = data['data']['jobs']
                    print(f"✅ 获取所有作业状态成功，共 {len(jobs)} 个作业")
                    
                    # 清理已完成作业
                    async with self.session.post(
                        f"{self.base_url}/api/batch/admin/cleanup?older_than_hours=1"
                    ) as cleanup_response:
                        if cleanup_response.status == 200:
                            cleanup_data = await cleanup_response.json()
                            print(f"✅ 清理已完成作业成功: {cleanup_data['msg']}")
                            print(f"   清理数量: {cleanup_data['data']['cleaned_jobs']}")
                            return True
                        else:
                            cleanup_data = await cleanup_response.json()
                            print(f"❌ 清理作业失败: {cleanup_data.get('detail', 'Unknown error')}")
                            return False
                else:
                    data = await response.json()
                    print(f"❌ 获取作业状态失败: {data.get('detail', 'Unknown error')}")
                    return False
        except Exception as e:
            print(f"❌ 管理员功能异常: {e}")
            return False
    
    async def run_all_tests(self):
        """运行所有测试"""
        print("🚀 开始批量操作API测试...")
        print("=" * 60)
        
        test_results = {}
        
        # 1. 健康检查
        test_results['health'] = await self.test_health_check()
        
        # 2. 验证功能
        test_results['validation'] = await self.test_batch_validation()
        
        # 3. 批量删除
        delete_job_id = await self.test_batch_delete_letters()
        test_results['delete'] = delete_job_id is not None
        
        # 4. 批量状态更新
        status_job_id = await self.test_batch_status_update()
        test_results['status_update'] = status_job_id is not None
        
        # 5. 批量导出
        export_job_id = await self.test_batch_export()
        test_results['export'] = export_job_id is not None
        
        # 6. 批量归档/恢复
        archive_job_id = await self.test_batch_archive_restore()
        test_results['archive_restore'] = archive_job_id is not None
        
        # 7. 批量创建
        create_job_id = await self.test_batch_create()
        test_results['create'] = create_job_id is not None
        
        # 8. 作业状态查询
        if delete_job_id:
            test_results['job_status'] = await self.test_job_status(delete_job_id)
        else:
            test_results['job_status'] = False
        
        # 9. 管理员功能（可选，需要管理员权限）
        try:
            test_results['admin'] = await self.test_admin_functions()
        except:
            test_results['admin'] = False
            print("⚠️ 管理员功能测试跳过（可能需要管理员权限）")
        
        # 测试结果汇总
        print("\n" + "=" * 60)
        print("📋 测试结果汇总:")
        
        passed = sum(test_results.values())
        total = len(test_results)
        
        for test_name, result in test_results.items():
            status = "✅ PASS" if result else "❌ FAIL"
            print(f"   {test_name:15}: {status}")
        
        print(f"\n🎯 总计: {passed}/{total} 测试通过")
        print(f"📊 通过率: {passed/total*100:.1f}%")
        
        if passed == total:
            print("\n🎉 所有批量操作功能测试通过！")
        else:
            print(f"\n⚠️ 有 {total-passed} 项测试失败，请检查相关功能")
        
        return test_results


async def main():
    """主函数"""
    # 配置测试参数
    BASE_URL = "http://localhost:8001"
    AUTH_TOKEN = "your-test-jwt-token"  # 需要替换为真实的JWT token
    
    print("批量操作API测试工具")
    print("=" * 60)
    print(f"📍 服务地址: {BASE_URL}")
    print(f"🔑 认证方式: {'JWT Token' if AUTH_TOKEN != 'your-test-jwt-token' else '未配置（某些功能可能失败）'}")
    print(f"⏰ 开始时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    async with BatchOperationTester(BASE_URL, AUTH_TOKEN) as tester:
        await tester.run_all_tests()
    
    print(f"\n⏰ 结束时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")


if __name__ == "__main__":
    asyncio.run(main())