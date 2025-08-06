#!/usr/bin/env node

/**
 * Smart Development Server Launcher
 * 自动检测可用端口并启动 Next.js 开发服务器
 */

const { spawn } = require('child_process')
const net = require('net')

// 默认端口范围
const DEFAULT_PORTS = [3000, 3001, 3002, 3003, 3004, 3005, 3006, 3007]
const PREFERRED_PORT = 3000

/**
 * 检查端口是否可用
 */
function checkPort(port) {
  return new Promise((resolve) => {
    const server = net.createServer()
    
    server.listen(port, () => {
      server.once('close', () => {
        resolve(true) // 端口可用
      })
      server.close()
    })
    
    server.on('error', () => {
      resolve(false) // 端口被占用
    })
  })
}

/**
 * 找到第一个可用端口
 */
async function findAvailablePort(ports = DEFAULT_PORTS) {
  for (const port of ports) {
    const isAvailable = await checkPort(port)
    if (isAvailable) {
      return port
    }
  }
  
  // 如果默认端口都被占用，尝试随机端口
  const randomPort = Math.floor(Math.random() * (9999 - 8000 + 1)) + 8000
  const isRandomAvailable = await checkPort(randomPort)
  if (isRandomAvailable) {
    return randomPort
  }
  
  throw new Error('无法找到可用端口')
}

/**
 * 启动开发服务器
 */
async function startDevServer() {
  try {
    console.log('🔍 正在检测可用端口...')
    
    const availablePort = await findAvailablePort()
    
    if (availablePort === PREFERRED_PORT) {
      console.log(`✅ 使用首选端口: ${availablePort}`)
    } else {
      console.log(`⚠️  端口 ${PREFERRED_PORT} 被占用，使用端口: ${availablePort}`)
    }
    
    console.log(`🚀 启动开发服务器: http://localhost:${availablePort}`)
    console.log('---')
    
    // 启动 Next.js 开发服务器
    const nextProcess = spawn('npx', ['next', 'dev', '--port', availablePort.toString()], {
      stdio: 'inherit',
      env: {
        ...process.env,
        PORT: availablePort.toString()
      }
    })
    
    // 处理进程退出
    nextProcess.on('close', (code) => {
      if (code !== 0) {
        console.error(`开发服务器退出，代码: ${code}`)
        process.exit(code)
      }
    })
    
    // 处理中断信号
    process.on('SIGINT', () => {
      console.log('\n👋 正在关闭开发服务器...')
      nextProcess.kill('SIGINT')
    })
    
    process.on('SIGTERM', () => {
      nextProcess.kill('SIGTERM')
    })
    
  } catch (error) {
    console.error('❌ 启动开发服务器失败:', error.message)
    process.exit(1)
  }
}

/**
 * 显示当前端口使用情况
 */
async function showPortStatus() {
  console.log('📊 端口使用情况:')
  
  for (const port of DEFAULT_PORTS) {
    const isAvailable = await checkPort(port)
    const status = isAvailable ? '✅ 可用' : '❌ 被占用'
    console.log(`  端口 ${port}: ${status}`)
  }
  
  console.log()
}

// 命令行参数处理
const args = process.argv.slice(2)

if (args.includes('--status') || args.includes('-s')) {
  showPortStatus()
} else if (args.includes('--help') || args.includes('-h')) {
  console.log(`
OpenPenPal 智能开发服务器启动器

用法:
  npm run smart-dev              启动开发服务器（自动检测端口）
  npm run smart-dev -- --status  显示端口使用情况
  npm run smart-dev -- --help    显示帮助

特性:
  - 自动检测可用端口 (优先使用 3000)
  - 端口被占用时自动切换到其他端口
  - 友好的启动日志和错误处理
  - 支持优雅关闭

默认端口范围: ${DEFAULT_PORTS.join(', ')}
`)
} else {
  startDevServer()
}