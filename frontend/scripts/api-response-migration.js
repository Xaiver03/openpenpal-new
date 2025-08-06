#!/usr/bin/env node

/**
 * API响应格式迁移检查脚本
 * API Response Format Migration Checker
 */

const fs = require('fs')
const path = require('path')

const colors = {
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
}

class APIResponseMigrationChecker {
  constructor() {
    this.apiDir = path.join(process.cwd(), 'src', 'app', 'api')
    this.results = {
      standard: [],
      needsMigration: [],
      errors: []
    }
  }

  log(message, color = 'white') {
    console.log(`${colors[color]}${message}${colors.reset}`)
  }

  async scanFile(filePath) {
    try {
      const content = fs.readFileSync(filePath, 'utf8')
      const relativePath = path.relative(process.cwd(), filePath)
      
      const analysis = {
        file: relativePath,
        hasStandardFormat: false,
        hasOldFormat: false,
        responsePatterns: [],
        issues: []
      }

      // 检查是否使用了新的响应构建器
      if (content.includes('ApiResponseBuilder')) {
        analysis.hasStandardFormat = true
        analysis.responsePatterns.push('ApiResponseBuilder')
      }

      if (content.includes('from \'@/lib/api/response\'')) {
        analysis.hasStandardFormat = true
        analysis.responsePatterns.push('标准响应导入')
      }

      // 检查旧格式模式
      const oldPatterns = [
        { pattern: /success:\s*true/g, name: 'success: true' },
        { pattern: /success:\s*false/g, name: 'success: false' },
        { pattern: /{\s*success:/g, name: '{ success:' },
        { pattern: /NextResponse\.json\(\s*{\s*error:/g, name: 'NextResponse.json({ error:' },
        { pattern: /NextResponse\.json\(\s*{\s*message:/g, name: 'NextResponse.json({ message:' }
      ]

      oldPatterns.forEach(({ pattern, name }) => {
        const matches = content.match(pattern)
        if (matches) {
          analysis.hasOldFormat = true
          analysis.responsePatterns.push(`${name} (${matches.length}次)`)
        }
      })

      // 检查是否有响应返回
      const responseReturns = content.match(/return\s+NextResponse\.json/g)
      if (responseReturns) {
        analysis.responsePatterns.push(`NextResponse.json返回 (${responseReturns.length}次)`)
      }

      // 检查潜在问题
      if (content.includes('NextResponse.json') && !analysis.hasStandardFormat) {
        analysis.issues.push('使用NextResponse.json但未使用标准响应格式')
      }

      if (analysis.hasOldFormat && analysis.hasStandardFormat) {
        analysis.issues.push('混合使用新旧响应格式')
      }

      return analysis
    } catch (error) {
      return {
        file: path.relative(process.cwd(), filePath),
        error: error.message
      }
    }
  }

  async scanDirectory(dir) {
    const entries = fs.readdirSync(dir, { withFileTypes: true })
    
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name)
      
      if (entry.isDirectory()) {
        await this.scanDirectory(fullPath)
      } else if (entry.name === 'route.ts') {
        const analysis = await this.scanFile(fullPath)
        
        if (analysis.error) {
          this.results.errors.push(analysis)
        } else if (analysis.hasStandardFormat && !analysis.hasOldFormat) {
          this.results.standard.push(analysis)
        } else {
          this.results.needsMigration.push(analysis)
        }
      }
    }
  }

  generateReport() {
    this.log('\n' + '='.repeat(80), 'cyan')
    this.log('API响应格式迁移检查报告', 'bold')
    this.log('='.repeat(80), 'cyan')

    const total = this.results.standard.length + this.results.needsMigration.length
    const migrated = this.results.standard.length
    const needsMigration = this.results.needsMigration.length

    this.log(`\n📊 总览:`, 'cyan')
    this.log(`   📁 总API文件: ${total}`, 'white')
    this.log(`   ✅ 已标准化: ${migrated}`, 'green')
    this.log(`   🔄 需要迁移: ${needsMigration}`, 'yellow')
    this.log(`   ❌ 扫描错误: ${this.results.errors.length}`, 'red')

    if (migrated > 0) {
      this.log(`\n✅ 已使用标准格式的API:`, 'green')
      this.results.standard.forEach(api => {
        this.log(`   • ${api.file}`, 'green')
        if (api.responsePatterns.length > 0) {
          api.responsePatterns.forEach(pattern => {
            this.log(`     - ${pattern}`, 'white')
          })
        }
      })
    }

    if (needsMigration > 0) {
      this.log(`\n🔄 需要迁移的API:`, 'yellow')
      this.results.needsMigration.forEach(api => {
        this.log(`   • ${api.file}`, 'yellow')
        
        if (api.responsePatterns.length > 0) {
          this.log(`     响应模式:`, 'white')
          api.responsePatterns.forEach(pattern => {
            this.log(`     - ${pattern}`, 'white')
          })
        }
        
        if (api.issues.length > 0) {
          this.log(`     问题:`, 'red')
          api.issues.forEach(issue => {
            this.log(`     ⚠️  ${issue}`, 'red')
          })
        }
      })
    }

    if (this.results.errors.length > 0) {
      this.log(`\n❌ 扫描错误:`, 'red')
      this.results.errors.forEach(error => {
        this.log(`   • ${error.file}: ${error.error}`, 'red')
      })
    }

    // 迁移建议
    this.log(`\n💡 迁移建议:`, 'cyan')
    
    if (needsMigration > 0) {
      this.log(`   1. 优先迁移高频使用的API端点`, 'white')
      this.log(`   2. 使用 ApiResponseBuilder 替换 NextResponse.json`, 'white')
      this.log(`   3. 导入: import { ApiResponseBuilder } from '@/lib/api/response'`, 'white')
      this.log(`   4. 成功响应: ApiResponseBuilder.success(data, message)`, 'white')
      this.log(`   5. 错误响应: ApiResponseBuilder.error(statusCode, message)`, 'white')
    } else {
      this.log(`   🎉 所有API已经使用标准响应格式！`, 'green')
    }

    this.log('\n' + '='.repeat(80), 'cyan')
    
    return needsMigration === 0
  }

  generateMigrationGuide() {
    if (this.results.needsMigration.length === 0) return

    this.log(`\n📋 详细迁移指南:`, 'cyan')
    
    this.results.needsMigration.forEach(api => {
      this.log(`\n📄 ${api.file}:`, 'yellow')
      
      // 提供具体的迁移建议
      if (api.responsePatterns.some(p => p.includes('success: true'))) {
        this.log(`   🔄 替换: { success: true, data: ... }`, 'white')
        this.log(`   ➡️  改为: ApiResponseBuilder.success(data, message)`, 'green')
      }
      
      if (api.responsePatterns.some(p => p.includes('error:'))) {
        this.log(`   🔄 替换: { error: "message" }`, 'white')
        this.log(`   ➡️  改为: ApiResponseBuilder.error(statusCode, message)`, 'green')
      }
      
      if (api.responsePatterns.some(p => p.includes('NextResponse.json'))) {
        this.log(`   🔄 添加导入: import { ApiResponseBuilder } from '@/lib/api/response'`, 'green')
      }
    })
  }

  async run() {
    this.log('🔍 开始扫描API响应格式...', 'cyan')
    
    if (!fs.existsSync(this.apiDir)) {
      this.log(`❌ API目录不存在: ${this.apiDir}`, 'red')
      return false
    }

    await this.scanDirectory(this.apiDir)
    
    const allMigrated = this.generateReport()
    this.generateMigrationGuide()
    
    return allMigrated
  }
}

// 运行检查
if (require.main === module) {
  const checker = new APIResponseMigrationChecker()
  checker.run().then(success => {
    process.exit(success ? 0 : 1)
  }).catch(error => {
    console.error('迁移检查失败:', error)
    process.exit(1)
  })
}

module.exports = APIResponseMigrationChecker