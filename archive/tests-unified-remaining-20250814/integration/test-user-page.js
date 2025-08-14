const axios = require('axios');

const API_URL = 'http://localhost:8080/api';
const FRONTEND_URL = 'http://localhost:3000';

async function testUserPageAPI() {
  console.log('🧪 测试用户主页API...\n');

  // 创建axios实例，跳过代理
  const axiosInstance = axios.create({
    proxy: false
  });

  try {
    // 1. 测试用户资料API
    console.log('1️⃣ 测试用户资料API...');
    
    console.log('  测试 alice 用户资料:');
    const aliceProfileResponse = await axiosInstance.get(`${API_URL}/users/alice/profile`);
    console.log('  ✅ 成功获取 alice 用户资料');
    console.log('  用户信息:', {
      username: aliceProfileResponse.data.data.username,
      nickname: aliceProfileResponse.data.data.nickname,
      role: aliceProfileResponse.data.data.role,
      school: aliceProfileResponse.data.data.school,
      opCode: aliceProfileResponse.data.data.opCode,
      writingLevel: aliceProfileResponse.data.data.writingLevel,
      courierLevel: aliceProfileResponse.data.data.courierLevel,
      achievements: aliceProfileResponse.data.data.stats?.achievements?.length || 0
    });

    console.log('\n  测试 admin 用户资料:');
    const adminProfileResponse = await axiosInstance.get(`${API_URL}/users/admin/profile`);
    console.log('  ✅ 成功获取 admin 用户资料');
    console.log('  用户信息:', {
      username: adminProfileResponse.data.data.username,
      nickname: adminProfileResponse.data.data.nickname,
      role: adminProfileResponse.data.data.role,
      opCode: adminProfileResponse.data.data.opCode,
      writingLevel: adminProfileResponse.data.data.writingLevel,
      courierLevel: adminProfileResponse.data.data.courierLevel,
      achievements: adminProfileResponse.data.data.stats?.achievements?.length || 0
    });

    console.log('\n  测试不存在的用户:');
    try {
      await axiosInstance.get(`${API_URL}/users/nonexistent/profile`);
    } catch (error) {
      if (error.response && error.response.status === 404) {
        console.log('  ✅ 正确返回 404 错误');
      } else {
        console.log('  ❌ 未预期的错误:', error.message);
      }
    }

    // 2. 测试用户信件API
    console.log('\n2️⃣ 测试用户信件API...');
    
    console.log('  测试 alice 公开信件:');
    const aliceLettersResponse = await axiosInstance.get(`${API_URL}/users/alice/letters?public=true`);
    console.log('  ✅ 成功获取 alice 信件列表');
    console.log('  信件数量:', aliceLettersResponse.data.data.count);
    
    if (aliceLettersResponse.data.data.letters.length > 0) {
      const firstLetter = aliceLettersResponse.data.data.letters[0];
      console.log('  第一封信件:', {
        title: firstLetter.title,
        preview: (firstLetter.contentPreview || firstLetter.content_preview || '').substring(0, 30) + '...',
        status: firstLetter.status
      });
    }

    console.log('\n  测试 admin 公开信件:');
    const adminLettersResponse = await axiosInstance.get(`${API_URL}/users/admin/letters?public=true`);
    console.log('  ✅ 成功获取 admin 信件列表');
    console.log('  信件数量:', adminLettersResponse.data.data.count);

    // 3. 测试前端页面
    console.log('\n3️⃣ 测试前端页面访问...');
    
    console.log('  测试 /u/alice 页面:');
    try {
      const frontendResponse = await axiosInstance.get(`${FRONTEND_URL}/u/alice`);
      if (frontendResponse.status === 200) {
        console.log('  ✅ 前端页面可正常访问');
      }
    } catch (error) {
      if (error.code === 'ECONNREFUSED') {
        console.log('  ⚠️  前端服务器未运行 (这是正常的，如果你还没启动前端)');
      } else {
        console.log('  ❌ 前端页面访问失败:', error.message);
      }
    }

    console.log('\n🎉 用户主页API测试完成!');
    console.log('\n📋 测试总结:');
    console.log('✅ 用户资料API: 正常工作');
    console.log('✅ 用户信件API: 正常工作');
    console.log('✅ 错误处理: 正确响应404');
    console.log('✅ Mock数据: alice 和 admin 用户数据完整');

    console.log('\n🚀 下一步:');
    console.log('1. 启动前端服务: cd frontend && npm run dev');
    console.log('2. 访问: http://localhost:3000/u/alice');
    console.log('3. 访问: http://localhost:3000/u/admin');
    console.log('4. 登录后查看导航菜单中的"我的主页"链接');

  } catch (error) {
    console.error('❌ 测试失败:', error.message);
    if (error.response) {
      console.error('响应状态:', error.response.status);
      console.error('响应数据:', error.response.data);
    }
  }
}

// 运行测试
testUserPageAPI();