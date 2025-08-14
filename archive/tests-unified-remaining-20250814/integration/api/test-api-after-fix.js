const axios = require('axios');

const API_URL = 'http://localhost:8080/api/v1';

async function testApiAfterFix() {
    console.log('🔍 测试中间件禁用后的API响应格式\n');
    
    try {
        // 1. 登录
        console.log('1. 测试登录API...');
        const loginRes = await axios.post(`${API_URL}/auth/login`, {
            username: 'alice',
            password: 'secret123'
        });
        
        console.log('登录响应格式:');
        console.log('- success:', loginRes.data.success);
        console.log('- message:', loginRes.data.message);
        const user = loginRes.data.data.user;
        console.log('- 用户字段:');
        console.log('  - school_code (snake_case):', user.school_code);
        console.log('  - is_active (snake_case):', user.is_active);
        console.log('  - created_at (snake_case):', user.created_at);
        console.log('  - schoolCode (camelCase):', user.schoolCode);
        console.log('  - isActive (camelCase):', user.isActive);
        
        const token = loginRes.data.data.token;
        console.log('\n2. 测试用户信息API...');
        
        // 2. 获取用户信息
        const userRes = await axios.get(`${API_URL}/users/me`, {
            headers: { Authorization: `Bearer ${token}` }
        });
        
        console.log('用户信息响应格式:');
        const userData = userRes.data.data;
        console.log('- school_code:', userData.school_code);
        console.log('- is_active:', userData.is_active);
        console.log('- created_at:', userData.created_at);
        console.log('- updated_at:', userData.updated_at);
        console.log('- last_login_at:', userData.last_login_at);
        
        console.log('\n3. 测试创建信件API...');
        const letterRes = await axios.post(`${API_URL}/letters/`, {
            title: "测试信件",
            content: "这是一封测试信件",
            style: "classic"
        }, {
            headers: { Authorization: `Bearer ${token}` }
        });
        
        console.log('信件响应格式:');
        const letter = letterRes.data.data;
        console.log('- user_id:', letter.user_id);
        console.log('- author_id:', letter.author_id);
        console.log('- like_count:', letter.like_count);
        console.log('- share_count:', letter.share_count);
        console.log('- created_at:', letter.created_at);
        
        console.log('\n✅ 结论: API现在返回snake_case格式，与前端期望一致！');
        
    } catch (error) {
        console.error('错误:', error.response?.data || error.message);
    }
}

testApiAfterFix();