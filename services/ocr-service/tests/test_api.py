import pytest
import json
import os
from unittest.mock import patch, MagicMock
from io import BytesIO

from app.main import create_app


@pytest.fixture
def app():
    """创建测试应用"""
    app = create_app('testing')
    app.config['TESTING'] = True
    return app


@pytest.fixture
def client(app):
    """创建测试客户端"""
    return app.test_client()


@pytest.fixture
def auth_headers():
    """模拟JWT认证头"""
    # 这里应该使用实际的JWT token，暂时使用模拟
    return {
        'Authorization': 'Bearer test-jwt-token'
    }


class TestHealthAPI:
    """健康检查API测试"""
    
    def test_health_check(self, client):
        """测试健康检查接口"""
        response = client.get('/health')
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert data['data']['service'] == 'ocr-service'
        assert data['data']['status'] == 'healthy'
    
    def test_ping(self, client):
        """测试ping接口"""
        response = client.get('/ping')
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert data['data'] == 'pong'


class TestOCRAPI:
    """OCR API测试"""
    
    def create_test_image(self):
        """创建测试图片"""
        # 创建一个简单的测试图片
        from PIL import Image
        import io
        
        img = Image.new('RGB', (100, 100), color='white')
        img_bytes = io.BytesIO()
        img.save(img_bytes, format='JPEG')
        img_bytes.seek(0)
        return img_bytes
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_recognize_without_auth(self, mock_decode, client):
        """测试无认证的OCR识别"""
        mock_decode.return_value = None
        
        response = client.post('/api/ocr/recognize')
        assert response.status_code == 403
        
        data = json.loads(response.data)
        assert data['code'] == 2  # 权限错误
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_recognize_no_file(self, mock_decode, client):
        """测试没有上传文件的OCR识别"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        response = client.post('/api/ocr/recognize', headers={'Authorization': 'Bearer test-token'})
        assert response.status_code == 400
        
        data = json.loads(response.data)
        assert data['code'] == 1  # 参数错误
        assert '缺少图片文件' in data['msg']
    
    @patch('app.utils.auth.decode_jwt_token')
    @patch('app.services.ocr_engine.MultiEngineOCR')
    def test_recognize_success(self, mock_ocr_class, mock_decode, client):
        """测试OCR识别成功"""
        # 模拟认证
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        # 模拟OCR结果
        mock_ocr = MagicMock()
        mock_result = {
            'text': '测试文本',
            'confidence': 0.95,
            'processing_time': 1.5,
            'blocks': [{'text': '测试文本', 'confidence': 0.95, 'bbox': [0, 0, 100, 20], 'line': 1}],
            'engine': 'tesseract'
        }
        mock_ocr.recognize_single_engine.return_value = mock_result
        mock_ocr_class.return_value = mock_ocr
        
        # 创建测试文件
        test_image = self.create_test_image()
        
        response = client.post(
            '/api/ocr/recognize',
            headers={'Authorization': 'Bearer test-token'},
            data={
                'image': (test_image, 'test.jpg'),
                'language': 'zh',
                'enhance': 'true'
            },
            content_type='multipart/form-data'
        )
        
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert data['msg'] == '识别成功'
        assert 'task_id' in data['data']
        assert data['data']['status'] == 'completed'
        assert data['data']['results']['text'] == '测试文本'
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_get_models(self, mock_decode, client):
        """测试获取模型列表"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        response = client.get('/api/ocr/models', headers={'Authorization': 'Bearer test-token'})
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert 'available_models' in data['data']
        assert isinstance(data['data']['available_models'], list)


class TestTasksAPI:
    """任务管理API测试"""
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_get_task_status(self, mock_decode, client):
        """测试获取任务状态"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        task_id = 'ocr_task_123456'
        response = client.get(f'/api/ocr/tasks/{task_id}', headers={'Authorization': 'Bearer test-token'})
        
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert data['data']['task_id'] == task_id
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_get_invalid_task(self, mock_decode, client):
        """测试获取无效任务"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        task_id = 'invalid_task_id'
        response = client.get(f'/api/ocr/tasks/{task_id}', headers={'Authorization': 'Bearer test-token'})
        
        assert response.status_code == 404
        
        data = json.loads(response.data)
        assert data['code'] == 3  # 资源不存在
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_get_user_tasks(self, mock_decode, client):
        """测试获取用户任务列表"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        response = client.get('/api/ocr/tasks/', headers={'Authorization': 'Bearer test-token'})
        
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert 'items' in data['data']
        assert 'pagination' in data['data']


class TestCacheAPI:
    """缓存API测试"""
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_get_cache_stats(self, mock_decode, client):
        """测试获取缓存统计"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        response = client.get('/api/ocr/cache/stats', headers={'Authorization': 'Bearer test-token'})
        
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert 'cache_type' in data['data']
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_clear_cache_no_permission(self, mock_decode, client):
        """测试非管理员清理缓存"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        response = client.post(
            '/api/ocr/cache/clear',
            headers={'Authorization': 'Bearer test-token'},
            json={'pattern': 'test_*'}
        )
        
        assert response.status_code == 403
        
        data = json.loads(response.data)
        assert data['code'] == 2  # 权限错误
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_clear_cache_admin(self, mock_decode, client):
        """测试管理员清理缓存"""
        mock_decode.return_value = {'user_id': 'admin_user', 'role': 'admin'}
        
        response = client.post(
            '/api/ocr/cache/clear',
            headers={'Authorization': 'Bearer admin-token'},
            json={'pattern': 'test_*'}
        )
        
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert data['msg'] == '缓存清理成功'


class TestImageProcessing:
    """图像处理测试"""
    
    @patch('app.utils.auth.decode_jwt_token')
    def test_enhance_image(self, mock_decode, client):
        """测试图像增强"""
        mock_decode.return_value = {'user_id': 'test_user', 'role': 'user'}
        
        test_image = BytesIO(b'fake image data')
        
        response = client.post(
            '/api/ocr/enhance',
            headers={'Authorization': 'Bearer test-token'},
            data={
                'image': (test_image, 'test.jpg'),
                'operations': '["denoise", "contrast"]',
                'return_enhanced': 'false'
            },
            content_type='multipart/form-data'
        )
        
        assert response.status_code == 200
        
        data = json.loads(response.data)
        assert data['code'] == 0
        assert 'operations_applied' in data['data']


if __name__ == '__main__':
    pytest.main([__file__])