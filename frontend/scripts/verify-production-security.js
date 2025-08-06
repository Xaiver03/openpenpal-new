#!/usr/bin/env node

/**
 * 生产环境安全验证脚本
 * Production Environment Security Verification Script
 */

const fs = require('fs')
const path = require('path')

/**
 * 检查硬编码密码
 */
function checkHardcodedPasswords() {
  const suspiciousPatterns = [
    /password.*=.*['"][a-zA-Z0-9]{1,20}['"](?!.*dev_)/gi,
    /admin.*123/gi,
    /courier.*123/gi,
    /test.*123/gi,
    /\/\/.*admin123|\/\/.*courier123|\/\/.*test123/gi,
    /\/\*.*admin123.*\*\/|\/\*.*courier123.*\*\/|\/\*.*test123.*\*\//gi
  ]
  
  const files = [
    'src/app/api/auth/login/route.ts',
    'src/config/courier-test-accounts.ts',
    'src/lib/auth/test-data-manager.ts'
  ]
  
  const issues = []
  
  files.forEach(filePath => {
    if (fs.existsSync(filePath)) {
      const content = fs.readFileSync(filePath, 'utf8')
      
      suspiciousPatterns.forEach((pattern, index) => {
        const matches = content.match(pattern)
        if (matches) {
          // 过滤掉合法的加密哈希生成
          const filteredMatches = matches.filter(match => 
            !match.includes('crypto.createHash') && 
            !match.includes('generateDefaultPassword') &&
            !match.includes('dev_${basePassword}')
          )
          
          if (filteredMatches.length > 0) {
            issues.push({
              file: filePath,
              pattern: pattern.toString(),
              matches: filteredMatches
            })
          }
        }
      })
    }
  })
  
  return issues
}

/**
 * 检查环境变量配置
 */
function checkEnvironmentConfig() {
  const envFiles = ['.env.local', '.env.production', '.env.example']
  const issues = []
  
  envFiles.forEach(envFile => {
    if (fs.existsSync(envFile)) {
      const content = fs.readFileSync(envFile, 'utf8')
      
      // 检查是否启用了测试数据
      if (content.includes('ENABLE_TEST_DATA=true') && envFile === '.env.production') {
        issues.push({
          file: envFile,
          issue: 'Production environment has test data enabled',
          line: content.split('\n').findIndex(line => line.includes('ENABLE_TEST_DATA=true')) + 1
        })
      }
      
      // 检查是否有默认测试密码
      const defaultPasswords = ['admin123', 'courier123', 'test123']
      defaultPasswords.forEach(password => {
        if (content.includes(password) && envFile === '.env.production') {
          issues.push({
            file: envFile,
            issue: `Production environment contains default password: ${password}`,
            line: content.split('\n').findIndex(line => line.includes(password)) + 1
          })
        }
      })
    }
  })
  
  return issues
}

/**
 * 检查测试数据管理器
 */
function checkTestDataManager() {
  const testDataManagerPath = 'src/lib/auth/test-data-manager.ts'
  const issues = []
  
  if (fs.existsSync(testDataManagerPath)) {
    const content = fs.readFileSync(testDataManagerPath, 'utf8')
    
    // 检查是否有生产环境保护
    if (!content.includes('isProduction()')) {
      issues.push({
        file: testDataManagerPath,
        issue: 'Test data manager missing production environment protection'
      })
    }
    
    // 检查是否有测试数据清理功能
    if (!content.includes('cleanupTestData')) {
      issues.push({
        file: testDataManagerPath,
        issue: 'Test data manager missing cleanup functionality'
      })
    }
  }
  
  return issues
}

/**
 * 验证JWT配置
 */
function checkJWTConfig() {
  const issues = []
  
  // 检查是否使用默认JWT密钥
  const envFiles = ['.env.local', '.env.production']
  
  envFiles.forEach(envFile => {
    if (fs.existsSync(envFile)) {
      const content = fs.readFileSync(envFile, 'utf8')
      
      if (content.includes('your-super-secret-jwt-key')) {
        issues.push({
          file: envFile,
          issue: 'Using default JWT secret key'
        })
      }
      
      if (content.includes('your-super-secret-refresh-key')) {
        issues.push({
          file: envFile,
          issue: 'Using default JWT refresh secret key'
        })
      }
    }
  })
  
  return issues
}

/**
 * 生成安全报告
 */
function generateSecurityReport() {
  const report = {
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || 'development',
    checks: {
      hardcodedPasswords: checkHardcodedPasswords(),
      environmentConfig: checkEnvironmentConfig(),
      testDataManager: checkTestDataManager(),
      jwtConfig: checkJWTConfig()
    }
  }
  
  const totalIssues = Object.values(report.checks).reduce((sum, issues) => sum + issues.length, 0)
  report.summary = {
    totalIssues,
    severity: totalIssues === 0 ? 'SECURE' : totalIssues < 3 ? 'LOW' : totalIssues < 6 ? 'MEDIUM' : 'HIGH'
  }
  
  return report
}

/**
 * 打印安全报告
 */
function printSecurityReport(report) {
  console.log('🔍 OpenPenPal 生产环境安全验证报告')
  console.log('═'.repeat(50))
  console.log(`📅 时间: ${report.timestamp}`)
  console.log(`🌍 环境: ${report.environment}`)
  console.log(`📊 安全等级: ${report.summary.severity}`)
  console.log(`⚠️  发现问题: ${report.summary.totalIssues}`)
  console.log('')
  
  // 打印各项检查结果
  Object.entries(report.checks).forEach(([checkName, issues]) => {
    const status = issues.length === 0 ? '✅' : '❌'
    console.log(`${status} ${checkName}: ${issues.length} 问题`)
    
    if (issues.length > 0) {
      issues.forEach(issue => {
        console.log(`   - ${issue.file || 'Unknown'}: ${issue.issue || issue.matches?.[0] || 'Issue detected'}`)
        if (issue.line) {
          console.log(`     行号: ${issue.line}`)
        }
      })
    }
  })
  
  console.log('')
  
  // 安全建议
  if (report.summary.totalIssues > 0) {
    console.log('🔧 修复建议:')
    console.log('1. 移除所有硬编码密码和密码提示')
    console.log('2. 确保生产环境设置 ENABLE_TEST_DATA=false')
    console.log('3. 使用强随机JWT密钥')
    console.log('4. 使用环境变量管理敏感配置')
    console.log('5. 定期运行安全验证脚本')
  } else {
    console.log('🎉 恭喜！您的配置通过了所有安全检查！')
  }
  
  console.log('')
  console.log('📖 更多安全指南: docs/DEPLOYMENT_JWT_CONFIG.md')
}

/**
 * 主函数
 */
async function main() {
  try {
    const report = generateSecurityReport()
    printSecurityReport(report)
    
    // 如果在CI环境中运行，根据安全等级决定退出码
    if (process.env.CI) {
      if (report.summary.severity === 'HIGH') {
        process.exit(1)
      }
    }
    
  } catch (error) {
    console.error('❌ 安全验证失败:', error.message)
    process.exit(1)
  }
}

// 运行脚本
if (require.main === module) {
  main()
}

module.exports = {
  checkHardcodedPasswords,
  checkEnvironmentConfig,
  checkTestDataManager,
  checkJWTConfig,
  generateSecurityReport
}