/**
 * 权限系统测试
 * 验证用户权限、角色访问控制等功能
 */

import { createLogger } from '../src/utils/logger.js';
import { USERS, hasPermission, canAccessService } from '../src/config/users.js';
import { generateToken, verifyToken } from '../src/middleware/auth.js';

const logger = createLogger('permission-test');

// 测试结果统计
let totalTests = 0;
let passedTests = 0;
let failedTests = 0;

// 测试工具函数
function runTest(testName, testFn) {
  totalTests++;
  logger.info(`运行测试: ${testName}`);
  
  try {
    const result = testFn();
    if (result) {
      passedTests++;
      logger.success(`✓ ${testName}`);
    } else {
      failedTests++;
      logger.error(`✗ ${testName} - 断言失败`);
    }
  } catch (error) {
    failedTests++;
    logger.error(`✗ ${testName} - 异常:`, error.message);
  }
}

function assert(condition, message = '断言失败') {
  if (!condition) {
    throw new Error(message);
  }
  return true;
}

// 用户权限测试
function testUserPermissions() {
  runTest('超级管理员拥有所有权限', () => {
    const admin = USERS.admin;
    return assert(hasPermission(admin, 'ANY_PERMISSION'), '超级管理员应该有任意权限');
  });
  
  runTest('学生用户有基础权限', () => {
    const alice = USERS.alice;
    return assert(
      hasPermission(alice, 'WRITE_READ') && 
      hasPermission(alice, 'LETTER_SEND') &&
      !hasPermission(alice, 'ADMIN_WRITE'),
      '学生应该有写信权限但没有管理权限'
    );
  });
  
  runTest('信使用户有配送权限', () => {
    const courier = USERS.courier1;
    return assert(
      hasPermission(courier, 'COURIER_READ') &&
      hasPermission(courier, 'TASK_ACCEPT') &&
      !hasPermission(courier, 'USER_MANAGE'),
      '信使应该有配送权限但没有用户管理权限'
    );
  });
  
  runTest('审核员有内容审核权限', () => {
    const moderator = USERS.moderator;
    return assert(
      hasPermission(moderator, 'CONTENT_MODERATE') &&
      hasPermission(moderator, 'MUSEUM_MODERATE') &&
      !hasPermission(moderator, 'SYSTEM_CONFIG'),
      '审核员应该有内容审核权限但没有系统配置权限'
    );
  });
}

// 服务访问权限测试
function testServiceAccess() {
  runTest('学生可以访问写信服务', () => {
    const alice = USERS.alice;
    return assert(canAccessService(alice, 'write-service'), '学生应该能访问写信服务');
  });
  
  runTest('信使可以访问信使服务', () => {
    const courier = USERS.courier1;
    return assert(canAccessService(courier, 'courier-service'), '信使应该能访问信使服务');
  });
  
  runTest('普通用户不能访问管理服务', () => {
    const alice = USERS.alice;
    return assert(!canAccessService(alice, 'admin-service'), '普通用户不应该能访问管理服务');
  });
  
  runTest('管理员可以访问所有服务', () => {
    const admin = USERS.admin;
    return assert(
      canAccessService(admin, 'admin-service') &&
      canAccessService(admin, 'write-service') &&
      canAccessService(admin, 'courier-service'),
      '管理员应该能访问所有服务'
    );
  });
}

// JWT Token 测试
function testJWTFunctionality() {
  runTest('Token 生成和验证', () => {
    const alice = USERS.alice;
    const token = generateToken(alice);
    
    assert(typeof token === 'string' && token.length > 0, 'Token 应该是非空字符串');
    
    const decoded = verifyToken(token);
    assert(decoded && decoded.id === alice.id, 'Token 验证应该返回正确的用户信息');
    
    return true;
  });
  
  runTest('无效 Token 验证失败', () => {
    const invalidToken = 'invalid.token.here';
    const decoded = verifyToken(invalidToken);
    return assert(decoded === null, '无效 Token 应该验证失败');
  });
  
  runTest('Token 包含正确的用户信息', () => {
    const courier = USERS.courier1;
    const token = generateToken(courier);
    const decoded = verifyToken(token);
    
    return assert(
      decoded.id === courier.id &&
      decoded.username === courier.username &&
      decoded.role === courier.role &&
      Array.isArray(decoded.permissions),
      'Token 应该包含完整的用户信息'
    );
  });
}

// 边界情况测试
function testEdgeCases() {
  runTest('空用户权限检查', () => {
    return assert(
      !hasPermission(null, 'ANY_PERMISSION') &&
      !hasPermission(undefined, 'ANY_PERMISSION'),
      '空用户应该没有任何权限'
    );
  });
  
  runTest('无权限用户', () => {
    const userWithoutPermissions = { id: 'test', permissions: [] };
    return assert(
      !hasPermission(userWithoutPermissions, 'ANY_PERMISSION'),
      '没有权限的用户应该被拒绝访问'
    );
  });
  
  runTest('未知服务访问', () => {
    const alice = USERS.alice;
    return assert(
      !canAccessService(alice, 'unknown-service'),
      '未知服务应该拒绝访问'
    );
  });
}

// 角色权限一致性测试
function testRoleConsistency() {
  runTest('用户权限与角色一致性', () => {
    // 检查所有用户的权限是否与其角色匹配
    for (const [username, user] of Object.entries(USERS)) {
      if (user.role === 'student') {
        assert(
          hasPermission(user, 'WRITE_READ') &&
          hasPermission(user, 'PROFILE_READ'),
          `学生用户 ${username} 应该有基础权限`
        );
      } else if (user.role === 'courier') {
        assert(
          hasPermission(user, 'COURIER_READ') &&
          hasPermission(user, 'TASK_READ'),
          `信使用户 ${username} 应该有配送权限`
        );
      } else if (user.role === 'super_admin') {
        assert(
          hasPermission(user, 'ALL'),
          `超级管理员 ${username} 应该有所有权限`
        );
      }
    }
    return true;
  });
}

// HTTP 请求模拟测试
async function testHTTPScenarios() {
  runTest('模拟登录请求', () => {
    // 模拟登录逻辑
    const username = 'alice';
    const password = 'secret';
    
    const user = USERS[username];
    assert(user, '用户应该存在');
    assert(user.password === password, '密码应该匹配');
    
    const token = generateToken(user);
    assert(token, '应该生成有效 Token');
    
    return true;
  });
  
  runTest('模拟权限检查中间件', () => {
    const alice = USERS.alice;
    const token = generateToken(alice);
    const decoded = verifyToken(token);
    
    // 模拟访问写信服务
    assert(canAccessService(decoded, 'write-service'), '应该允许访问写信服务');
    assert(hasPermission(decoded, 'WRITE_CREATE'), '应该允许创建信件');
    
    // 模拟访问管理服务
    assert(!canAccessService(decoded, 'admin-service'), '应该拒绝访问管理服务');
    
    return true;
  });
}

// 主测试函数
async function runAllTests() {
  logger.info('🧪 开始权限系统测试');
  logger.info('='.repeat(50));
  
  // 运行所有测试组
  testUserPermissions();
  testServiceAccess();
  testJWTFunctionality();
  testEdgeCases();
  testRoleConsistency();
  await testHTTPScenarios();
  
  // 输出测试结果
  logger.info('='.repeat(50));
  logger.info('📊 测试结果统计');
  logger.info(`总测试数: ${totalTests}`);
  logger.success(`通过: ${passedTests}`);
  
  if (failedTests > 0) {
    logger.error(`失败: ${failedTests}`);
    logger.error('❌ 部分测试失败');
    process.exit(1);
  } else {
    logger.success('✅ 所有测试通过！');
    logger.success('🎉 权限系统验证完成');
  }
}

// 如果直接运行此文件
if (import.meta.url === `file://${process.argv[1]}`) {
  runAllTests().catch((error) => {
    logger.error('测试运行失败:', error);
    process.exit(1);
  });
}

export { runAllTests };