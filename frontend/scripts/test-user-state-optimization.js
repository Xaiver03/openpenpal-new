#!/usr/bin/env node

/**
 * User State Management Optimization Test Script
 * 用户状态管理优化测试脚本
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

class UserStateOptimizationTester {
  constructor() {
    this.srcDir = path.join(process.cwd(), 'src')
    this.results = {
      files: {
        created: [],
        updated: [],
        errors: []
      },
      optimizations: {
        storeIntegration: false,
        duplicateDataElimination: false,
        optimisticUpdates: false,
        unifiedLoading: false,
        performanceOptimizations: false
      },
      compatibility: {
        legacyHooks: false,
        authContext: false,
        existingComponents: false
      }
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

  async analyzeFile(filePath) {
    try {
      const content = await fs.promises.readFile(filePath, 'utf8')
      return {
        path: filePath,
        size: content.length,
        lines: content.split('\n').length,
        content,
        exists: true
      }
    } catch (error) {
      return {
        path: filePath,
        exists: false,
        error: error.message
      }
    }
  }

  async testStoreCreation() {
    this.log('\n🔍 Testing Store Creation...', 'cyan')
    
    const storeFile = path.join(this.srcDir, 'stores', 'user-store.ts')
    const analysis = await this.analyzeFile(storeFile)
    
    if (analysis.exists) {
      this.log('  ✅ User store created successfully', 'green')
      this.results.files.created.push('stores/user-store.ts')
      
      // Check store features
      const { content } = analysis
      const features = {
        zustand: content.includes('create from zustand'),
        devtools: content.includes('devtools'),
        persist: content.includes('persist'),
        optimisticUpdates: content.includes('optimisticUpdate'),
        permissionsCache: content.includes('permissionsCache'),
        loadingStates: content.includes('LoadingState')
      }
      
      Object.entries(features).forEach(([feature, hasFeature]) => {
        if (hasFeature) {
          this.log(`  ✅ ${feature} integration: ✓`, 'green')
        } else {
          this.log(`  ❌ ${feature} integration: ✗`, 'red')
        }
      })
      
      this.results.optimizations.storeIntegration = true
      return true
    } else {
      this.log('  ❌ User store not found', 'red')
      this.results.files.errors.push('stores/user-store.ts not found')
      return false
    }
  }

  async testHookOptimization() {
    this.log('\n🔍 Testing Hook Optimization...', 'cyan')
    
    const hookFiles = [
      'hooks/use-courier-permission.ts',
      'hooks/use-unified-loading.ts',
      'hooks/use-optimized-subscriptions.ts'
    ]
    
    let optimizedHooks = 0
    
    for (const hookFile of hookFiles) {
      const fullPath = path.join(this.srcDir, hookFile)
      const analysis = await this.analyzeFile(fullPath)
      
      if (analysis.exists) {
        this.log(`  ✅ ${hookFile} exists`, 'green')
        this.results.files.created.push(hookFile)
        
        const { content } = analysis
        
        // Check for store integration
        if (content.includes('useUserStore') || content.includes('useCourier') || content.includes('usePermissions')) {
          this.log(`    ✅ Store integration: ✓`, 'green')
          optimizedHooks++
        } else {
          this.log(`    ❌ Store integration: ✗`, 'red')
        }
        
        // Check for performance optimizations
        if (content.includes('shallow') || content.includes('useCallback') || content.includes('useMemo')) {
          this.log(`    ✅ Performance optimizations: ✓`, 'green')
        } else {
          this.log(`    ⚠️  Performance optimizations: missing`, 'yellow')
        }
      } else {
        this.log(`  ❌ ${hookFile} not found`, 'red')
        this.results.files.errors.push(`${hookFile} not found`)
      }
    }
    
    this.results.optimizations.duplicateDataElimination = optimizedHooks > 0
    this.results.optimizations.performanceOptimizations = optimizedHooks >= 2
    
    return optimizedHooks >= 2
  }

  async testOptimisticUpdates() {
    this.log('\n🔍 Testing Optimistic Updates...', 'cyan')
    
    const storeFile = path.join(this.srcDir, 'stores', 'user-store.ts')
    const analysis = await this.analyzeFile(storeFile)
    
    if (analysis.exists) {
      const { content } = analysis
      
      const optimisticFeatures = {
        optimisticUpdateMethod: content.includes('optimisticUpdate:'),
        rollbackMechanism: content.includes('rollbackFn'),
        errorHandling: content.includes('catch (error)'),
        stateRestoration: content.includes('rollback()')
      }
      
      let implementedFeatures = 0
      
      Object.entries(optimisticFeatures).forEach(([feature, hasFeature]) => {
        if (hasFeature) {
          this.log(`  ✅ ${feature}: ✓`, 'green')
          implementedFeatures++
        } else {
          this.log(`  ❌ ${feature}: ✗`, 'red')
        }
      })
      
      this.results.optimizations.optimisticUpdates = implementedFeatures >= 3
      return implementedFeatures >= 3
    }
    
    return false
  }

  async testUnifiedLoading() {
    this.log('\n🔍 Testing Unified Loading States...', 'cyan')
    
    const loadingHook = path.join(this.srcDir, 'hooks', 'use-unified-loading.ts')
    const analysis = await this.analyzeFile(loadingHook)
    
    if (analysis.exists) {
      const { content } = analysis
      
      const loadingFeatures = {
        globalLoading: content.includes('globalLoading'),
        localLoading: content.includes('localLoading'),
        operationLoading: content.includes('useOperationLoading'),
        batchLoading: content.includes('useBatchLoading'),
        progressTracking: content.includes('progress'),
        timeoutSupport: content.includes('timeout'),
        retryMechanism: content.includes('retries')
      }
      
      let implementedFeatures = 0
      
      Object.entries(loadingFeatures).forEach(([feature, hasFeature]) => {
        if (hasFeature) {
          this.log(`  ✅ ${feature}: ✓`, 'green')
          implementedFeatures++
        } else {
          this.log(`  ❌ ${feature}: ✗`, 'red')
        }
      })
      
      this.results.optimizations.unifiedLoading = implementedFeatures >= 5
      return implementedFeatures >= 5
    } else {
      this.log('  ❌ Unified loading hook not found', 'red')
      return false
    }
  }

  async testCompatibility() {
    this.log('\n🔍 Testing Backward Compatibility...', 'cyan')
    
    // Check for auth context wrapper
    const authContextNew = path.join(this.srcDir, 'contexts', 'auth-context-new.tsx')
    const authAnalysis = await this.analyzeFile(authContextNew)
    
    if (authAnalysis.exists) {
      this.log('  ✅ Auth context wrapper created', 'green')
      this.results.compatibility.authContext = true
      
      const { content } = authAnalysis
      
      // Check compatibility features
      const compatFeatures = {
        legacyInterfaces: content.includes('interface User'),
        legacyMethods: content.includes('checkPermission') && content.includes('hasRole'),
        storeIntegration: content.includes('useUserStore'),
        eventEmission: content.includes('CustomEvent')
      }
      
      Object.entries(compatFeatures).forEach(([feature, hasFeature]) => {
        if (hasFeature) {
          this.log(`    ✅ ${feature}: ✓`, 'green')
        } else {
          this.log(`    ❌ ${feature}: ✗`, 'red')
        }
      })
    } else {
      this.log('  ❌ Auth context wrapper not found', 'red')
    }
    
    // Check courier permission hook compatibility
    const courierHook = path.join(this.srcDir, 'hooks', 'use-courier-permission.ts')
    const courierAnalysis = await this.analyzeFile(courierHook)
    
    if (courierAnalysis.exists) {
      this.log('  ✅ Courier permission hook maintained', 'green')
      this.results.compatibility.legacyHooks = true
      
      const { content } = courierAnalysis
      if (content.includes('legacyCourierInfo')) {
        this.log('    ✅ Legacy format compatibility: ✓', 'green')
      } else {
        this.log('    ⚠️  Legacy format compatibility: missing', 'yellow')
      }
    }
    
    return this.results.compatibility.authContext && this.results.compatibility.legacyHooks
  }

  async testPackageDependencies() {
    this.log('\n🔍 Testing Package Dependencies...', 'cyan')
    
    const packageJsonPath = path.join(process.cwd(), 'package.json')
    
    try {
      const packageJson = JSON.parse(await fs.promises.readFile(packageJsonPath, 'utf8'))
      const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies }
      
      const requiredPackages = {
        'zustand': 'State management',
        'react': 'React framework',
        'typescript': 'TypeScript support'
      }
      
      let installedPackages = 0
      
      Object.entries(requiredPackages).forEach(([pkg, description]) => {
        if (dependencies[pkg]) {
          this.log(`  ✅ ${pkg} (${dependencies[pkg]}): ${description}`, 'green')
          installedPackages++
        } else {
          this.log(`  ❌ ${pkg}: Missing - ${description}`, 'red')
        }
      })
      
      return installedPackages === Object.keys(requiredPackages).length
    } catch (error) {
      this.log('  ❌ Failed to check package.json', 'red')
      return false
    }
  }

  generateReport() {
    this.log('\\n' + '='.repeat(80), 'cyan')
    this.log('User State Management Optimization Report', 'bold')
    this.log('='.repeat(80), 'cyan')
    
    // File Status
    this.log('\\n📁 File Status:', 'cyan')
    this.log(`   Created: ${this.results.files.created.length}`, 'green')
    this.results.files.created.forEach(file => {
      this.log(`     • ${file}`, 'white')
    })
    
    if (this.results.files.updated.length > 0) {
      this.log(`   Updated: ${this.results.files.updated.length}`, 'yellow')
      this.results.files.updated.forEach(file => {
        this.log(`     • ${file}`, 'white')
      })
    }
    
    if (this.results.files.errors.length > 0) {
      this.log(`   Errors: ${this.results.files.errors.length}`, 'red')
      this.results.files.errors.forEach(error => {
        this.log(`     • ${error}`, 'red')
      })
    }
    
    // Optimization Status
    this.log('\\n🚀 Optimization Status:', 'cyan')
    const optimizations = [
      ['Store Integration', this.results.optimizations.storeIntegration],
      ['Duplicate Data Elimination', this.results.optimizations.duplicateDataElimination],
      ['Optimistic Updates', this.results.optimizations.optimisticUpdates],
      ['Unified Loading', this.results.optimizations.unifiedLoading],
      ['Performance Optimizations', this.results.optimizations.performanceOptimizations]
    ]
    
    optimizations.forEach(([name, status]) => {
      const icon = status ? '✅' : '❌'
      const color = status ? 'green' : 'red'
      this.log(`   ${icon} ${name}`, color)
    })
    
    // Compatibility Status
    this.log('\\n🔄 Compatibility Status:', 'cyan')
    const compatibility = [
      ['Legacy Hooks', this.results.compatibility.legacyHooks],
      ['Auth Context', this.results.compatibility.authContext],
      ['Existing Components', this.results.compatibility.existingComponents]
    ]
    
    compatibility.forEach(([name, status]) => {
      const icon = status ? '✅' : '❌'
      const color = status ? 'green' : 'red'
      this.log(`   ${icon} ${name}`, color)
    })
    
    // Overall Status
    const totalOptimizations = Object.values(this.results.optimizations).filter(Boolean).length
    const totalCompatibility = Object.values(this.results.compatibility).filter(Boolean).length
    const overallScore = ((totalOptimizations + totalCompatibility) / 8) * 100
    
    this.log('\\n📊 Overall Status:', 'cyan')
    this.log(`   Optimization Score: ${totalOptimizations}/5 (${Math.round((totalOptimizations/5)*100)}%)`, 
              totalOptimizations >= 4 ? 'green' : totalOptimizations >= 2 ? 'yellow' : 'red')
    this.log(`   Compatibility Score: ${totalCompatibility}/3 (${Math.round((totalCompatibility/3)*100)}%)`, 
              totalCompatibility >= 2 ? 'green' : totalCompatibility >= 1 ? 'yellow' : 'red')
    this.log(`   Overall Score: ${Math.round(overallScore)}%`, 
              overallScore >= 80 ? 'green' : overallScore >= 60 ? 'yellow' : 'red')
    
    // Recommendations
    this.log('\\n💡 Recommendations:', 'cyan')
    
    if (!this.results.optimizations.storeIntegration) {
      this.log('   • Complete user store implementation with Zustand', 'yellow')
    }
    
    if (!this.results.optimizations.duplicateDataElimination) {
      this.log('   • Refactor hooks to use centralized store', 'yellow')
    }
    
    if (!this.results.optimizations.optimisticUpdates) {
      this.log('   • Implement optimistic update mechanism', 'yellow')
    }
    
    if (!this.results.optimizations.unifiedLoading) {
      this.log('   • Create unified loading state management', 'yellow')
    }
    
    if (!this.results.optimizations.performanceOptimizations) {
      this.log('   • Add performance optimizations (selectors, memoization)', 'yellow')
    }
    
    if (!this.results.compatibility.authContext) {
      this.log('   • Create auth context compatibility wrapper', 'yellow')
    }
    
    if (overallScore >= 80) {
      this.log('\\n🎉 Excellent! User state management optimization is complete!', 'green')
    } else if (overallScore >= 60) {
      this.log('\\n👍 Good progress! A few more optimizations needed.', 'yellow')
    } else {
      this.log('\\n⚠️  More work needed to complete the optimization.', 'red')
    }
    
    this.log('\\n' + '='.repeat(80), 'cyan')
    
    return overallScore >= 80
  }

  async run() {
    this.log('🚀 Starting User State Management Optimization Test...', 'cyan')
    
    // Run all tests
    await this.testPackageDependencies()
    await this.testStoreCreation()
    await this.testHookOptimization()
    await this.testOptimisticUpdates()
    await this.testUnifiedLoading()
    await this.testCompatibility()
    
    // Generate report
    const success = this.generateReport()
    
    return success
  }
}

// Run the test
if (require.main === module) {
  const tester = new UserStateOptimizationTester()
  tester.run().then(success => {
    process.exit(success ? 0 : 1)
  }).catch(error => {
    console.error('Test failed:', error)
    process.exit(1)
  })
}

module.exports = UserStateOptimizationTester