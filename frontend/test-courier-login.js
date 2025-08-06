#!/usr/bin/env node

/**
 * 四级信使系统登录测试脚本
 * 测试所有级别的信使是否能正确登录并获取信使信息
 */

const accounts = [
  { username: 'courier_level4_city', password: 'city123', level: 4, name: '四级信使（城市总代）' },
  { username: 'courier_level3_school', password: 'school123', level: 3, name: '三级信使（校级）' },
  { username: 'courier_level2_zone', password: 'zone123', level: 2, name: '二级信使（片区/年级）' },
  { username: 'courier_level1_basic', password: 'basic123', level: 1, name: '一级信使（楼栋/班级）' }
];

const API_BASE = 'http://localhost:3000';

async function testLogin(account) {
  try {
    console.log(`\n🧪 测试 ${account.name} (${account.username}) 登录...`);
    
    // 1. 测试登录
    const loginResponse = await fetch(`${API_BASE}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: account.username,
        password: account.password
      })
    });
    
    const loginResult = await loginResponse.json();
    
    if (loginResult.code !== 0) {
      console.log(`❌ 登录失败: ${loginResult.message}`);
      return false;
    }
    
    console.log(`✅ 登录成功`);
    console.log(`   - 用户ID: ${loginResult.data.user.id}`);
    console.log(`   - 角色: ${loginResult.data.user.role}`);
    console.log(`   - 权限数量: ${loginResult.data.user.permissions.length}`);
    
    // 检查courierInfo
    if (loginResult.data.user.courierInfo) {
      const courierInfo = loginResult.data.user.courierInfo;
      console.log(`   - 信使级别: ${courierInfo.level}`);
      console.log(`   - 覆盖区域: ${courierInfo.zoneCode}`);
      console.log(`   - 区域类型: ${courierInfo.zoneType}`);
      console.log(`   - 积分: ${courierInfo.points}`);
      console.log(`   - 完成任务: ${courierInfo.taskCount}`);
    } else {
      console.log(`⚠️  courierInfo 缺失`);
    }
    
    // 2. 测试获取用户信息
    const token = loginResult.data.accessToken;
    const meResponse = await fetch(`${API_BASE}/api/auth/me`, {
      method: 'GET',
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    const meResult = await meResponse.json();
    
    if (meResult.code !== 0) {
      console.log(`❌ 获取用户信息失败: ${meResult.message}`);
      return false;
    }
    
    console.log(`✅ 用户信息获取成功`);
    if (meResult.data.courierInfo) {
      console.log(`   - /api/auth/me courierInfo: ✅ 存在`);
    } else {
      console.log(`   - /api/auth/me courierInfo: ❌ 缺失`);
    }
    
    // 3. 测试信使相关API（level 2+）
    if (account.level >= 2) {
      const subordinatesResponse = await fetch(`${API_BASE}/api/courier/subordinates`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      const subordinatesResult = await subordinatesResponse.json();
      
      if (subordinatesResult.success) {
        console.log(`✅ 下级信使查询成功，数量: ${subordinatesResult.data.couriers.length}`);
      } else {
        console.log(`❌ 下级信使查询失败: ${subordinatesResult.error}`);
      }
    } else {
      // Level 1 应该被拒绝
      const subordinatesResponse = await fetch(`${API_BASE}/api/courier/subordinates`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (subordinatesResponse.status === 403) {
        console.log(`✅ Level 1 正确被拒绝访问管理功能`);
      } else {
        console.log(`❌ Level 1 应该被拒绝访问，但返回状态: ${subordinatesResponse.status}`);
      }
    }
    
    // 4. 测试信使信息API
    const courierMeResponse = await fetch(`${API_BASE}/api/courier/me`, {
      method: 'GET',
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    const courierMeResult = await courierMeResponse.json();
    
    if (courierMeResult.success) {
      console.log(`✅ 信使信息API成功`);
      console.log(`   - API返回级别: ${courierMeResult.data.level}`);
      console.log(`   - API返回积分: ${courierMeResult.data.total_points}`);
    } else {
      console.log(`❌ 信使信息API失败: ${courierMeResult.error}`);
    }
    
    return true;
  } catch (error) {
    console.log(`❌ 测试出错: ${error.message}`);
    return false;
  }
}

async function main() {
  console.log('🚀 开始四级信使系统全面测试\n');
  console.log('='.repeat(60));
  
  let successCount = 0;
  const totalCount = accounts.length;
  
  for (const account of accounts) {
    const success = await testLogin(account);
    if (success) successCount++;
    
    console.log('-'.repeat(60));
  }
  
  console.log(`\n📊 测试结果汇总:`);
  console.log(`   - 成功: ${successCount}/${totalCount}`);
  console.log(`   - 失败: ${totalCount - successCount}/${totalCount}`);
  
  if (successCount === totalCount) {
    console.log(`\n🎉 所有测试通过！四级信使系统运行正常。`);
  } else {
    console.log(`\n⚠️  存在失败的测试，请检查系统配置。`);
  }
}

// 运行测试
main().catch(console.error);