#!/usr/bin/env node

/**
 * Auth Context Migration Script
 * 认证上下文迁移脚本
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

class AuthContextMigrator {
  constructor() {
    this.srcDir = path.join(process.cwd(), 'src')
    this.layoutFile = path.join(this.srcDir, 'app', 'layout.tsx')
    this.results = {
      migrated: [],
      errors: [],
      skipped: []
    }
  }

  log(message, color = 'white') {
    console.log(`${colors[color]}${message}${colors.reset}`)
  }

  async checkFileExists(filePath) {
    try {
      await fs.promises.access(filePath, fs.constants.F_OK)
      return true
    } catch {
      return false
    }
  }

  async readFile(filePath) {
    try {
      return await fs.promises.readFile(filePath, 'utf8')
    } catch (error) {
      throw new Error(`Failed to read ${filePath}: ${error.message}`)
    }
  }

  async writeFile(filePath, content) {
    try {
      await fs.promises.writeFile(filePath, content, 'utf8')
      return true
    } catch (error) {
      throw new Error(`Failed to write ${filePath}: ${error.message}`)
    }
  }

  async migrateRootLayout() {
    this.log('\n🔄 Migrating Root Layout...', 'cyan')
    
    if (!(await this.checkFileExists(this.layoutFile))) {
      this.log('  ❌ Root layout file not found', 'red')
      this.results.errors.push('Root layout file not found')
      return false
    }

    try {
      const content = await this.readFile(this.layoutFile)
      
      // Check if already using new auth context
      if (content.includes('@/contexts/auth-context-new')) {
        this.log('  ⏭️  Already using new auth context', 'yellow')
        this.results.skipped.push('Root layout already migrated')
        return true
      }

      // Replace old auth context import with new one
      let newContent = content.replace(
        /import { AuthProvider } from '@\/contexts\/auth-context'/g,
        "import { AuthProvider } from '@/contexts/auth-context-new'"
      )

      // If no replacement was made, check if AuthProvider is imported differently
      if (newContent === content && content.includes('AuthProvider')) {
        // Try to find and replace any auth context import
        newContent = content.replace(
          /from '@\/contexts\/auth-context'/g,
          "from '@/contexts/auth-context-new'"
        )
      }

      // If still no replacement and we found AuthProvider usage, add the import
      if (newContent === content && content.includes('<AuthProvider>')) {
        // Add import at the top of the imports
        const importInsertPoint = content.indexOf("import")
        if (importInsertPoint !== -1) {
          newContent = content.slice(0, importInsertPoint) +
            "import { AuthProvider } from '@/contexts/auth-context-new'\n" +
            content.slice(importInsertPoint)
        }
      }

      if (newContent !== content) {
        await this.writeFile(this.layoutFile, newContent)
        this.log('  ✅ Root layout migrated successfully', 'green')
        this.results.migrated.push('Root layout')
        return true
      } else {
        this.log('  ⚠️  No migration needed or AuthProvider not found', 'yellow')
        this.results.skipped.push('Root layout - no changes needed')
        return true
      }
    } catch (error) {
      this.log(`  ❌ Error migrating root layout: ${error.message}`, 'red')
      this.results.errors.push(`Root layout migration error: ${error.message}`)
      return false
    }
  }

  async createMigrationBackup() {
    this.log('\n💾 Creating Migration Backup...', 'cyan')
    
    const backupDir = path.join(process.cwd(), '.migration-backup')
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const backupPath = path.join(backupDir, `auth-context-backup-${timestamp}`)

    try {
      // Create backup directory
      await fs.promises.mkdir(backupPath, { recursive: true })

      // Backup root layout
      if (await this.checkFileExists(this.layoutFile)) {
        const layoutContent = await this.readFile(this.layoutFile)
        await this.writeFile(path.join(backupPath, 'layout.tsx'), layoutContent)
      }

      // Create backup info file
      const backupInfo = {
        timestamp: new Date().toISOString(),
        files: ['layout.tsx'],
        migration: 'auth-context-new',
        description: 'Backup before migrating to optimized auth context'
      }

      await this.writeFile(
        path.join(backupPath, 'backup-info.json'),
        JSON.stringify(backupInfo, null, 2)
      )

      this.log(`  ✅ Backup created at: ${backupPath}`, 'green')
      return true
    } catch (error) {
      this.log(`  ⚠️  Backup failed: ${error.message}`, 'yellow')
      return false
    }
  }

  async testMigration() {
    this.log('\n🧪 Testing Migration...', 'cyan')
    
    try {
      // Check if new auth context exists
      const newAuthContextPath = path.join(this.srcDir, 'contexts', 'auth-context-new.tsx')
      if (!(await this.checkFileExists(newAuthContextPath))) {
        this.log('  ❌ New auth context file not found', 'red')
        return false
      }

      // Check if user store exists
      const userStorePath = path.join(this.srcDir, 'stores', 'user-store.ts')
      if (!(await this.checkFileExists(userStorePath))) {
        this.log('  ❌ User store not found', 'red')
        return false
      }

      // Check if optimized hooks exist
      const optimizedHooksPath = path.join(this.srcDir, 'hooks', 'use-optimized-subscriptions.ts')
      if (!(await this.checkFileExists(optimizedHooksPath))) {
        this.log('  ❌ Optimized hooks not found', 'red')
        return false
      }

      this.log('  ✅ All required files present', 'green')
      return true
    } catch (error) {
      this.log(`  ❌ Test failed: ${error.message}`, 'red')
      return false
    }
  }

  generateReport() {
    this.log('\n' + '='.repeat(80), 'cyan')
    this.log('Auth Context Migration Report', 'bold')
    this.log('='.repeat(80), 'cyan')
    
    // Migration Status
    this.log('\n📁 Migration Status:', 'cyan')
    
    if (this.results.migrated.length > 0) {
      this.log(`   Migrated: ${this.results.migrated.length}`, 'green')
      this.results.migrated.forEach(item => {
        this.log(`     ✅ ${item}`, 'green')
      })
    }
    
    if (this.results.skipped.length > 0) {
      this.log(`   Skipped: ${this.results.skipped.length}`, 'yellow')
      this.results.skipped.forEach(item => {
        this.log(`     ⏭️  ${item}`, 'yellow')
      })
    }
    
    if (this.results.errors.length > 0) {
      this.log(`   Errors: ${this.results.errors.length}`, 'red')
      this.results.errors.forEach(error => {
        this.log(`     ❌ ${error}`, 'red')
      })
    }

    // Instructions
    this.log('\n📝 Next Steps:', 'cyan')
    this.log('   1. Test the application to ensure auth still works', 'white')
    this.log('   2. Update remaining components to use optimized hooks', 'white')
    this.log('   3. Run the user state optimization test script', 'white')
    this.log('   4. Monitor for any performance improvements', 'white')

    // Performance Notes
    this.log('\n🚀 Expected Improvements:', 'cyan')
    this.log('   • Reduced duplicate user data requests', 'green')
    this.log('   • Optimistic updates for better UX', 'green')
    this.log('   • Unified loading state management', 'green')
    this.log('   • Better performance with selective subscriptions', 'green')
    this.log('   • Maintained backward compatibility', 'green')

    const success = this.results.errors.length === 0
    
    if (success) {
      this.log('\n🎉 Migration completed successfully!', 'green')
    } else {
      this.log('\n⚠️  Migration completed with errors. Please review.', 'yellow')
    }
    
    this.log('\n' + '='.repeat(80), 'cyan')
    
    return success
  }

  async run() {
    this.log('🚀 Starting Auth Context Migration...', 'cyan')
    
    // Create backup first
    await this.createMigrationBackup()
    
    // Test prerequisites
    const testPassed = await this.testMigration()
    if (!testPassed) {
      this.log('\n❌ Prerequisites not met. Aborting migration.', 'red')
      return false
    }
    
    // Perform migration
    await this.migrateRootLayout()
    
    // Generate report
    return this.generateReport()
  }
}

// Run the migration
if (require.main === module) {
  const migrator = new AuthContextMigrator()
  migrator.run().then(success => {
    process.exit(success ? 0 : 1)
  }).catch(error => {
    console.error('Migration failed:', error)
    process.exit(1)
  })
}

module.exports = AuthContextMigrator