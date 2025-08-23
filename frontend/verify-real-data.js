#!/usr/bin/env node

/**
 * 验证设置页面使用真实数据库数据
 */

const chalk = require('chalk') || { red: (s) => s, green: (s) => s, yellow: (s) => s, blue: (s) => s }

console.log(chalk.blue('\n🔍 验证设置页面数据源...\n'))

// 检查文件中是否还有 mock 相关的代码
const fs = require('fs')
const path = require('path')

const filesToCheck = [
  'src/stores/notification-store.ts',
  'src/stores/user-store.ts',
  'src/components/settings/notification-settings.tsx',
  'src/components/settings/security-settings.tsx',
  'src/components/profile/privacy-settings.tsx',
  'src/app/(main)/settings/profile/page.tsx'
]

const mockPatterns = [
  /generateMock/gi,
  /mockNotifications/gi,
  /TODO.*Replace.*with.*real.*API/gi,
  /Mock.*implementation/gi,
  /test_courier_mode/gi
]

let hasIssues = false

filesToCheck.forEach(file => {
  const filePath = path.join(__dirname, file)
  if (!fs.existsSync(filePath)) {
    console.log(chalk.yellow(`⚠️  文件不存在: ${file}`))
    return
  }
  
  const content = fs.readFileSync(filePath, 'utf8')
  const lines = content.split('\n')
  
  mockPatterns.forEach(pattern => {
    lines.forEach((line, index) => {
      if (pattern.test(line)) {
        hasIssues = true
        console.log(chalk.red(`❌ 发现可能的 mock 数据使用:`))
        console.log(`   文件: ${file}`)
        console.log(`   行号: ${index + 1}`)
        console.log(`   内容: ${line.trim()}`)
        console.log('')
      }
    })
  })
})

// 检查 localStorage 中的测试模式标记
console.log(chalk.blue('\n📦 检查需要清理的 localStorage 键:\n'))
const testKeys = [
  'test_courier_mode',
  'test_courier_level',
  'mock_data_enabled',
  'use_test_data',
  'openpenpal_privacy_settings' // fallback 存储
]

console.log('请在浏览器控制台运行以下命令清理测试数据:')
console.log(chalk.green(`
// 清理所有测试相关的 localStorage
${testKeys.map(key => `localStorage.removeItem('${key}')`).join('\n')}

// 刷新页面
location.reload()
`))

if (!hasIssues) {
  console.log(chalk.green('\n✅ 太好了！没有发现明显的 mock 数据使用。'))
  console.log(chalk.green('   所有设置页面组件都应该使用真实的数据库数据。\n'))
} else {
  console.log(chalk.red('\n⚠️  警告：发现一些可能使用 mock 数据的代码。'))
  console.log(chalk.red('   请检查并修复上述问题。\n'))
}

// 提供验证步骤
console.log(chalk.blue('\n🧪 验证步骤:\n'))
console.log('1. 打开浏览器开发者工具的 Network 标签')
console.log('2. 访问 http://localhost:3000/settings')
console.log('3. 检查以下 API 调用:')
console.log('   - GET /api/v1/notifications/preferences')
console.log('   - GET /api/v1/privacy/settings')
console.log('   - GET /api/v1/users/me')
console.log('4. 确保所有请求都返回真实数据，而不是 404 或错误')
console.log('')