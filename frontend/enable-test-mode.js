/**
 * 快速启用测试信使模式脚本
 * 在浏览器控制台运行此脚本
 */

// 测试模式配置
const TEST_CONFIGS = {
  1: { name: '一级信使（楼栋）', icon: '🏢' },
  2: { name: '二级信使（片区）', icon: '👥' },
  3: { name: '三级信使（学校）', icon: '🚚' },
  4: { name: '四级信使（城市）', icon: '👑' }
}

// 快速启用函数
function enableCourierTestMode(level = 2) {
  if (level < 1 || level > 4) {
    console.error('❌ 等级必须在1-4之间')
    return
  }
  
  const config = TEST_CONFIGS[level]
  
  console.log(`
🧪 启用测试信使模式
==================
等级: ${config.icon} ${config.name}
状态: ✅ 已启用

注意: 页面将在3秒后刷新...
  `)
  
  localStorage.setItem('test_courier_mode', 'true')
  localStorage.setItem('test_courier_level', level.toString())
  
  setTimeout(() => {
    location.reload()
  }, 3000)
}

// 禁用函数
function disableCourierTestMode() {
  console.log('🧪 禁用测试信使模式...')
  localStorage.removeItem('test_courier_mode')
  localStorage.removeItem('test_courier_level')
  location.reload()
}

// 检查当前状态
function checkTestModeStatus() {
  const enabled = localStorage.getItem('test_courier_mode') === 'true'
  const level = localStorage.getItem('test_courier_level') || 'N/A'
  
  if (enabled) {
    const config = TEST_CONFIGS[level] || { name: '未知', icon: '❓' }
    console.log(`
📊 测试模式状态
==============
状态: ✅ 已启用
等级: ${config.icon} ${config.name}
    `)
  } else {
    console.log(`
📊 测试模式状态
==============
状态: ❌ 未启用
    `)
  }
}

// 使用说明
console.log(`
🧪 OpenPenPal 测试信使模式控制台
================================

可用命令:
---------
enableCourierTestMode(1)  - 启用一级信使模式
enableCourierTestMode(2)  - 启用二级信使模式（默认）
enableCourierTestMode(3)  - 启用三级信使模式
enableCourierTestMode(4)  - 启用四级信使模式

disableCourierTestMode()  - 禁用测试模式
checkTestModeStatus()     - 检查当前状态

示例:
-----
enableCourierTestMode(4)  // 成为四级信使
`)

// 自动检查当前状态
checkTestModeStatus()