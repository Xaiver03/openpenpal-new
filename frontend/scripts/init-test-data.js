#!/usr/bin/env node

/**
 * 测试数据初始化脚本
 * Test Data Initialization Script for OpenPenPal
 */

const crypto = require('crypto')
const bcrypt = require('bcryptjs')

/**
 * 生成安全的测试密码
 */
function generateSecurePassword(length = 16) {
  const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*'
  let password = ''
  for (let i = 0; i < length; i++) {
    password += charset.charAt(Math.floor(Math.random() * charset.length))
  }
  return password
}

/**
 * 生成密码哈希
 */
async function hashPassword(password) {
  return bcrypt.hash(password, 12)
}

/**
 * 生成测试账户配置
 */
async function generateTestAccountConfig() {
  const accounts = [
    'admin',
    'courier_building', 
    'senior_courier',
    'coordinator',
    'courier_level4_city',
    'courier_level3_school',
    'courier_level2_zone',
    'courier_level1_building'
  ]
  
  const config = {}
  const passwordDoc = []
  
  console.log('🔐 生成测试账户配置...\n')
  
  for (const account of accounts) {
    const password = generateSecurePassword()
    const hash = await hashPassword(password)
    
    const envVar = `TEST_ACCOUNT_${account.toUpperCase()}_PASSWORD`
    config[envVar] = password
    
    passwordDoc.push({
      username: account,
      password: password,
      hash: hash,
      envVar: envVar
    })
    
    console.log(`✅ ${account}: ${password}`)
  }
  
  return { config, passwordDoc }
}

/**
 * 生成环境配置文件内容
 */
function generateEnvConfig(config) {
  let envContent = `
# Test Data Configuration (Generated on ${new Date().toISOString()})
# WARNING: Only use in development environment!
ENABLE_TEST_DATA=true
`
  
  Object.entries(config).forEach(([key, value]) => {
    envContent += `${key}=${value}\n`
  })
  
  return envContent
}

/**
 * 生成生产环境清理配置
 */
function generateProductionConfig() {
  return `
# Production Environment Configuration
# Test data is DISABLED for security
NODE_ENV=production
ENABLE_TEST_DATA=false

# Use secure passwords from your secret management system
# TEST_ACCOUNT_*_PASSWORD variables should be unset or use strong passwords
`
}

/**
 * 验证当前环境
 */
function validateEnvironment() {
  const nodeEnv = process.env.NODE_ENV
  
  if (nodeEnv === 'production') {
    console.error('❌ 错误: 不能在生产环境中运行测试数据初始化脚本!')
    console.error('请设置 NODE_ENV=development 或在开发环境中运行')
    process.exit(1)
  }
  
  console.log(`✅ 环境检查通过: ${nodeEnv || 'development'}`)
}

/**
 * 主函数
 */
async function main() {
  try {
    console.log('🚀 OpenPenPal 测试数据初始化脚本\n')
    
    // 验证环境
    validateEnvironment()
    
    // 生成测试账户配置
    const { config, passwordDoc } = await generateTestAccountConfig()
    
    // 生成配置文件内容
    const envConfig = generateEnvConfig(config)
    const prodConfig = generateProductionConfig()
    
    console.log('\n📝 环境配置文件内容 (.env.local):')
    console.log('─'.repeat(50))
    console.log(envConfig)
    
    console.log('\n🏭 生产环境配置 (.env.production):')
    console.log('─'.repeat(50))
    console.log(prodConfig)
    
    console.log('\n📊 密码统计:')
    console.log(`- 生成账户数: ${passwordDoc.length}`)
    console.log(`- 密码长度: 16字符`)
    console.log(`- 密码强度: 高 (包含大小写字母、数字、特殊字符)`)
    
    console.log('\n🔒 安全提醒:')
    console.log('1. 这些密码仅用于开发和测试环境')
    console.log('2. 生产环境必须使用 ENABLE_TEST_DATA=false')
    console.log('3. 生产环境应使用密钥管理系统存储密码')
    console.log('4. 定期轮换测试环境密码')
    
    console.log('\n✅ 测试数据配置生成完成!')
    
  } catch (error) {
    console.error('❌ 初始化失败:', error.message)
    process.exit(1)
  }
}

// 运行脚本
if (require.main === module) {
  main()
}

module.exports = {
  generateSecurePassword,
  hashPassword,
  generateTestAccountConfig,
  generateEnvConfig,
  generateProductionConfig
}