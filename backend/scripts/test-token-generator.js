#!/usr/bin/env node

/**
 * 🔐 OpenPenPal 安全测试令牌生成器
 * 
 * 功能：为测试环境生成安全的JWT令牌，替代硬编码令牌
 * 安全原则：
 * 1. 使用环境变量存储密钥
 * 2. 生成短期有效令牌
 * 3. 包含明确的测试标识
 * 4. 支持不同角色和权限
 */

const jwt = require('jsonwebtoken');
const crypto = require('crypto');

// 测试环境专用密钥 (绝不在生产环境使用)
const TEST_JWT_SECRET = process.env.TEST_JWT_SECRET || 'test_secret_for_local_development_only_never_use_in_production_' + crypto.randomBytes(16).toString('hex');

// 预定义测试用户角色
const TEST_ROLES = {
  ADMIN: {
    role: 'super_admin',
    permissions: [
      'MANAGE_USERS', 'VIEW_ANALYTICS', 'MODERATE_CONTENT', 
      'MANAGE_SCHOOLS', 'MANAGE_EXHIBITIONS', 'SYSTEM_CONFIG',
      'AUDIT_SUBMISSIONS', 'HANDLE_REPORTS'
    ],
    userId: 'test-admin-' + crypto.randomBytes(4).toString('hex')
  },
  USER: {
    role: 'user',
    permissions: ['CREATE_LETTER', 'VIEW_LETTERS', 'PARTICIPATE_ACTIVITIES'],
    userId: 'test-user-' + crypto.randomBytes(4).toString('hex')
  },
  COURIER_L1: {
    role: 'courier_level1',
    permissions: ['courier_scan_code', 'courier_deliver_letter', 'courier_view_own_tasks'],
    userId: 'test-courier-l1-' + crypto.randomBytes(4).toString('hex')
  },
  COURIER_L2: {
    role: 'courier_level2', 
    permissions: ['courier_scan_code', 'courier_deliver_letter', 'courier_manage_subordinates'],
    userId: 'test-courier-l2-' + crypto.randomBytes(4).toString('hex')
  },
  COURIER_L3: {
    role: 'courier_level3',
    permissions: ['courier_manage_school_zone', 'courier_view_school_analytics'],
    userId: 'test-courier-l3-' + crypto.randomBytes(4).toString('hex')
  },
  COURIER_L4: {
    role: 'courier_level4',
    permissions: ['courier_manage_city_operations', 'courier_view_city_analytics'],
    userId: 'test-courier-l4-' + crypto.randomBytes(4).toString('hex')
  }
};

/**
 * 生成测试JWT令牌
 * @param {string} roleType - 角色类型 (ADMIN, USER, COURIER_L1-L4)
 * @param {Object} customPayload - 自定义载荷
 * @param {string} expiresIn - 过期时间 (默认: 2h)
 * @returns {string} JWT令牌
 */
function generateTestToken(roleType = 'USER', customPayload = {}, expiresIn = '2h') {
  const roleConfig = TEST_ROLES[roleType];
  if (!roleConfig) {
    throw new Error(`Unknown role type: ${roleType}. Available: ${Object.keys(TEST_ROLES).join(', ')}`);
  }

  const now = Math.floor(Date.now() / 1000);
  
  const payload = {
    // 标准字段
    userId: roleConfig.userId,
    username: `test_${roleConfig.role}`,
    role: roleConfig.role,
    permissions: roleConfig.permissions,
    
    // JWT标准声明
    iss: 'openpenpal-test', // 明确标识为测试环境
    aud: 'openpenpal-client',
    iat: now,
    jti: crypto.randomBytes(16).toString('hex'), // 唯一标识符
    
    // 测试环境标识
    env: 'test',
    schoolCode: roleType.includes('COURIER') ? 'TEST01' : 'PKU001',
    
    // 合并自定义载荷
    ...customPayload
  };

  return jwt.sign(payload, TEST_JWT_SECRET, { 
    expiresIn,
    algorithm: 'HS256'
  });
}

/**
 * 验证测试令牌
 * @param {string} token - JWT令牌
 * @returns {Object} 解码后的载荷
 */
function verifyTestToken(token) {
  try {
    return jwt.verify(token, TEST_JWT_SECRET);
  } catch (error) {
    throw new Error(`Token verification failed: ${error.message}`);
  }
}

/**
 * 解码令牌信息（不验证签名）
 * @param {string} token - JWT令牌
 * @returns {Object} 解码后的载荷
 */
function decodeTestToken(token) {
  return jwt.decode(token);
}

/**
 * 生成长期有效的令牌（用于长时间运行的测试）
 * @param {string} roleType - 角色类型
 * @returns {string} 长期有效令牌
 */
function generateLongLivedToken(roleType = 'ADMIN') {
  return generateTestToken(roleType, {}, '30d');
}

/**
 * 批量生成测试令牌
 * @returns {Object} 包含所有角色的令牌对象
 */
function generateAllTestTokens() {
  const tokens = {};
  Object.keys(TEST_ROLES).forEach(roleType => {
    tokens[roleType.toLowerCase()] = generateTestToken(roleType);
  });
  return tokens;
}

// 命令行使用
if (require.main === module) {
  const args = process.argv.slice(2);
  const command = args[0] || 'admin';
  
  console.log('🔐 OpenPenPal 安全测试令牌生成器\n');
  
  try {
    switch (command.toLowerCase()) {
      case 'admin':
        console.log('管理员令牌:');
        console.log(generateTestToken('ADMIN'));
        break;
        
      case 'user':
        console.log('普通用户令牌:');
        console.log(generateTestToken('USER'));
        break;
        
      case 'courier':
        const level = args[1] || '1';
        const courierType = `COURIER_L${level}`;
        console.log(`${level}级信使令牌:`);
        console.log(generateTestToken(courierType));
        break;
        
      case 'all':
        console.log('所有测试令牌:');
        const allTokens = generateAllTestTokens();
        Object.entries(allTokens).forEach(([role, token]) => {
          console.log(`${role.toUpperCase()}:`);
          console.log(`  ${token}\n`);
        });
        break;
        
      case 'long':
        console.log('长期有效管理员令牌 (30天):');
        console.log(generateLongLivedToken('ADMIN'));
        break;
        
      case 'verify':
        if (!args[1]) {
          console.error('请提供要验证的令牌');
          process.exit(1);
        }
        const decoded = verifyTestToken(args[1]);
        console.log('令牌验证成功:');
        console.log(JSON.stringify(decoded, null, 2));
        break;
        
      default:
        console.log('使用方法:');
        console.log('  node test-token-generator.js admin     # 生成管理员令牌');
        console.log('  node test-token-generator.js user      # 生成用户令牌');
        console.log('  node test-token-generator.js courier 1 # 生成1级信使令牌');
        console.log('  node test-token-generator.js all       # 生成所有角色令牌');
        console.log('  node test-token-generator.js long      # 生成长期令牌');
        console.log('  node test-token-generator.js verify <token> # 验证令牌');
    }
  } catch (error) {
    console.error('❌ 错误:', error.message);
    process.exit(1);
  }
}

module.exports = {
  generateTestToken,
  verifyTestToken,
  decodeTestToken,
  generateLongLivedToken,
  generateAllTestTokens,
  TEST_ROLES
};