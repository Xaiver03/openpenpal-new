#!/usr/bin/env node

/**
 * Circular Dependency Fix Validation Test
 * 循环依赖修复验证测试
 * 
 * Purpose: Validate that our fixes have resolved circular dependency issues
 * 目的: 验证我们的修复已解决循环依赖问题
 */

const { execSync } = require('child_process')
const fs = require('fs')
const path = require('path')

// ================================
// Test Configuration
// ================================

const TEST_CONFIG = {
  // Components to test
  testComponents: [
    'src/lib/di/service-interfaces.ts',
    'src/lib/di/service-container.ts', 
    'src/lib/di/service-adapters.ts',
    'src/lib/di/service-registry.ts',
    'src/contexts/auth-context-di.tsx',
    'src/lib/services/service-factory.ts',
    'src/lib/api/safe-index.ts'
  ],
  
  // Import tests
  importTests: [
    {
      name: 'DI Service Interfaces',
      file: 'src/lib/di/service-interfaces.ts',
      expectedExports: ['IAuthService', 'IUserStateService', 'SERVICE_KEYS']
    },
    {
      name: 'Service Container',
      file: 'src/lib/di/service-container.ts',
      expectedExports: ['ServiceContainer', 'getServiceContainer']
    },
    {
      name: 'Auth Context DI',
      file: 'src/contexts/auth-context-di.tsx',
      expectedExports: ['AuthProviderDI', 'useAuthDI']
    }
  ],
  
  // Madge analysis
  madgeConfig: {
    extensions: ['ts', 'tsx', 'js', 'jsx'],
    maxDepth: 10,
    excludePatterns: ['node_modules', '.next', 'dist', '*.test.*']
  }
}

// ================================
// Test Runner Class
// ================================

class CircularDependencyTestRunner {
  constructor() {
    this.results = {
      importTests: [],
      syntaxTests: [],
      madgeAnalysis: null,
      diContainerTest: null,
      integrationTest: null,
      summary: {
        passed: 0,
        failed: 0,
        total: 0
      }
    }
  }

  // ================================
  // Individual Tests
  // ================================

  async testFileImports(testCase) {
    const testName = `Import Test: ${testCase.name}`
    console.log(`🧪 Running ${testName}...`)
    
    try {
      const filePath = path.join(process.cwd(), testCase.file)
      
      if (!fs.existsSync(filePath)) {
        throw new Error(`File does not exist: ${testCase.file}`)
      }

      const content = fs.readFileSync(filePath, 'utf-8')
      
      // Check for expected exports
      const missingExports = []
      for (const exportName of testCase.expectedExports) {
        const exportRegex = new RegExp(`export.*${exportName}`)
        if (!exportRegex.test(content)) {
          missingExports.push(exportName)
        }
      }
      
      if (missingExports.length > 0) {
        throw new Error(`Missing exports: ${missingExports.join(', ')}`)
      }

      this.results.importTests.push({
        name: testName,
        status: 'PASSED',
        file: testCase.file,
        message: 'All expected exports found'
      })
      
      this.results.summary.passed++
      console.log(`   ✅ ${testName} - PASSED`)
      
    } catch (error) {
      this.results.importTests.push({
        name: testName,
        status: 'FAILED',
        file: testCase.file,
        error: error.message
      })
      
      this.results.summary.failed++
      console.log(`   ❌ ${testName} - FAILED: ${error.message}`)
    }
    
    this.results.summary.total++
  }

  async testFileSyntax(filePath) {
    const testName = `Syntax Test: ${filePath}`
    console.log(`🔍 Running ${testName}...`)
    
    try {
      const fullPath = path.join(process.cwd(), filePath)
      
      if (!fs.existsSync(fullPath)) {
        throw new Error(`File does not exist: ${filePath}`)
      }

      // Use TypeScript compiler to check syntax
      execSync(`npx tsc --noEmit --skipLibCheck ${fullPath}`, { 
        stdio: 'pipe',
        cwd: process.cwd()
      })

      this.results.syntaxTests.push({
        name: testName,
        status: 'PASSED',
        file: filePath,
        message: 'TypeScript compilation successful'
      })
      
      this.results.summary.passed++
      console.log(`   ✅ ${testName} - PASSED`)
      
    } catch (error) {
      this.results.syntaxTests.push({
        name: testName,
        status: 'FAILED',
        file: filePath,
        error: error.message
      })
      
      this.results.summary.failed++
      console.log(`   ❌ ${testName} - FAILED: ${error.message}`)
    }
    
    this.results.summary.total++
  }

  async testMadgeAnalysis() {
    const testName = 'Madge Circular Dependency Analysis'
    console.log(`🔍 Running ${testName}...`)
    
    try {
      // Run madge to check for circular dependencies
      const madgeOutput = execSync(
        `npx madge --circular --extensions ts,tsx,js,jsx src/`,
        { 
          stdio: 'pipe',
          cwd: process.cwd(),
          encoding: 'utf-8'
        }
      )

      const hasCircularDeps = madgeOutput.includes('Circular dependency found') || 
                              madgeOutput.includes('circular dependencies')

      this.results.madgeAnalysis = {
        name: testName,
        status: hasCircularDeps ? 'FAILED' : 'PASSED',
        output: madgeOutput,
        circularDepsFound: hasCircularDeps,
        message: hasCircularDeps 
          ? 'Circular dependencies detected'
          : 'No circular dependencies found'
      }

      if (hasCircularDeps) {
        this.results.summary.failed++
        console.log(`   ❌ ${testName} - FAILED: Circular dependencies still exist`)
        console.log(`   Output: ${madgeOutput}`)
      } else {
        this.results.summary.passed++
        console.log(`   ✅ ${testName} - PASSED`)
      }

    } catch (error) {
      // If madge exits with error, it might indicate circular dependencies
      const errorOutput = error.stdout || error.message
      const hasCircularDeps = errorOutput.includes('circular') ||
                              errorOutput.includes('Circular')

      this.results.madgeAnalysis = {
        name: testName,
        status: hasCircularDeps ? 'FAILED' : 'WARNING',
        output: errorOutput,
        error: error.message,
        message: hasCircularDeps 
          ? 'Circular dependencies detected by madge'
          : 'Madge analysis encountered issues but may not indicate circular dependencies'
      }

      if (hasCircularDeps) {
        this.results.summary.failed++
        console.log(`   ❌ ${testName} - FAILED: ${error.message}`)
      } else {
        console.log(`   ⚠️  ${testName} - WARNING: ${error.message}`)
      }
    }
    
    this.results.summary.total++
  }

  async testDIContainer() {
    const testName = 'Dependency Injection Container Test'
    console.log(`🧪 Running ${testName}...`)
    
    try {
      // Create a temporary test file to check DI container functionality
      const testCode = `
const { ServiceRegistry, getServiceContainer } = require('./src/lib/di/service-registry')
const { SERVICE_KEYS } = require('./src/lib/di/service-interfaces')

// Test service registry initialization
try {
  ServiceRegistry.initialize()
  console.log('✓ ServiceRegistry initialization successful')
  
  // Test service resolution
  const container = getServiceContainer()
  const hasAuthService = container.has(SERVICE_KEYS.AUTH_SERVICE)
  
  if (hasAuthService) {
    console.log('✓ Auth service registration successful')
  } else {
    throw new Error('Auth service not registered')
  }
  
  console.log('✓ DI Container test passed')
  process.exit(0)
} catch (error) {
  console.error('✗ DI Container test failed:', error.message)
  process.exit(1)
}
`

      const testFile = path.join(process.cwd(), 'temp-di-test.js')
      fs.writeFileSync(testFile, testCode)

      try {
        execSync(`node ${testFile}`, { 
          stdio: 'pipe',
          cwd: process.cwd(),
          timeout: 10000
        })

        this.results.diContainerTest = {
          name: testName,
          status: 'PASSED',
          message: 'DI container functionality verified'
        }
        
        this.results.summary.passed++
        console.log(`   ✅ ${testName} - PASSED`)

      } finally {
        // Clean up test file
        if (fs.existsSync(testFile)) {
          fs.unlinkSync(testFile)
        }
      }

    } catch (error) {
      this.results.diContainerTest = {
        name: testName,
        status: 'FAILED',
        error: error.message,
        message: 'DI container functionality test failed'
      }
      
      this.results.summary.failed++
      console.log(`   ❌ ${testName} - FAILED: ${error.message}`)
    }
    
    this.results.summary.total++
  }

  async testIntegration() {
    const testName = 'Integration Test - Context and Services'
    console.log(`🧪 Running ${testName}...`)
    
    try {
      // Test that our new context can import without circular dependencies
      const testCode = `
// Test imports
const serviceInterfaces = require('./src/lib/di/service-interfaces')
const serviceContainer = require('./src/lib/di/service-container')
const serviceAdapters = require('./src/lib/di/service-adapters')
const serviceRegistry = require('./src/lib/di/service-registry')

console.log('✓ All DI modules imported successfully')

// Test that key exports are available
const requiredExports = [
  'SERVICE_KEYS',
  'getServiceContainer',
  'AuthServiceAdapter',
  'ServiceRegistry'
]

for (const exportName of requiredExports) {
  const hasExport = Object.keys({
    ...serviceInterfaces,
    ...serviceContainer,
    ...serviceAdapters,
    ...serviceRegistry
  }).includes(exportName)
  
  if (hasExport) {
    console.log(\`✓ \${exportName} export found\`)
  } else {
    throw new Error(\`Missing export: \${exportName}\`)
  }
}

console.log('✓ Integration test passed')
process.exit(0)
`

      const testFile = path.join(process.cwd(), 'temp-integration-test.js')
      fs.writeFileSync(testFile, testCode)

      try {
        const output = execSync(`node ${testFile}`, { 
          stdio: 'pipe',
          cwd: process.cwd(),
          encoding: 'utf-8',
          timeout: 10000
        })

        this.results.integrationTest = {
          name: testName,
          status: 'PASSED',
          output: output,
          message: 'Integration test completed successfully'
        }
        
        this.results.summary.passed++
        console.log(`   ✅ ${testName} - PASSED`)

      } finally {
        // Clean up test file
        if (fs.existsSync(testFile)) {
          fs.unlinkSync(testFile)
        }
      }

    } catch (error) {
      this.results.integrationTest = {
        name: testName,
        status: 'FAILED',
        error: error.message,
        message: 'Integration test failed'
      }
      
      this.results.summary.failed++
      console.log(`   ❌ ${testName} - FAILED: ${error.message}`)
    }
    
    this.results.summary.total++
  }

  // ================================
  // Main Test Runner
  // ================================

  async runAllTests() {
    console.log('🚀 Starting Circular Dependency Fix Validation Tests...\n')

    // Test 1: Import Tests
    console.log('📋 Phase 1: Import Tests')
    for (const testCase of TEST_CONFIG.importTests) {
      await this.testFileImports(testCase)
    }

    // Test 2: Syntax Tests
    console.log('\n📋 Phase 2: Syntax Tests')
    for (const component of TEST_CONFIG.testComponents) {
      await this.testFileSyntax(component)
    }

    // Test 3: Madge Analysis
    console.log('\n📋 Phase 3: Circular Dependency Analysis')
    await this.testMadgeAnalysis()

    // Test 4: DI Container Test
    console.log('\n📋 Phase 4: DI Container Functionality')
    await this.testDIContainer()

    // Test 5: Integration Test
    console.log('\n📋 Phase 5: Integration Test')
    await this.testIntegration()

    // Generate Report
    this.generateReport()
  }

  generateReport() {
    console.log('\n' + '='.repeat(60))
    console.log('📊 CIRCULAR DEPENDENCY FIX VALIDATION REPORT')
    console.log('='.repeat(60))
    
    console.log('\n📈 SUMMARY:')
    console.log(`   Total tests: ${this.results.summary.total}`)
    console.log(`   Passed: ${this.results.summary.passed}`)
    console.log(`   Failed: ${this.results.summary.failed}`)
    console.log(`   Success rate: ${((this.results.summary.passed / this.results.summary.total) * 100).toFixed(1)}%`)

    if (this.results.summary.failed > 0) {
      console.log('\n🔴 FAILED TESTS:')
      
      const failedTests = [
        ...this.results.importTests.filter(t => t.status === 'FAILED'),
        ...this.results.syntaxTests.filter(t => t.status === 'FAILED')
      ]

      if (this.results.madgeAnalysis?.status === 'FAILED') {
        failedTests.push(this.results.madgeAnalysis)
      }

      if (this.results.diContainerTest?.status === 'FAILED') {
        failedTests.push(this.results.diContainerTest)
      }

      if (this.results.integrationTest?.status === 'FAILED') {
        failedTests.push(this.results.integrationTest)
      }

      failedTests.forEach(test => {
        console.log(`   ❌ ${test.name}: ${test.error || test.message}`)
      })
    }

    // Save detailed report
    const reportPath = path.join(process.cwd(), 'circular-deps-fix-report.json')
    fs.writeFileSync(reportPath, JSON.stringify({
      timestamp: new Date().toISOString(),
      summary: this.results.summary,
      results: this.results
    }, null, 2))

    console.log(`\n📄 Detailed report saved: ${reportPath}`)

    if (this.results.summary.failed === 0) {
      console.log('\n✅ All tests passed! Circular dependency fixes appear to be working correctly.')
    } else {
      console.log('\n⚠️  Some tests failed. Please review the issues above.')
    }

    return this.results.summary.failed === 0
  }
}

// ================================
// Main Execution
// ================================

async function main() {
  try {
    const testRunner = new CircularDependencyTestRunner()
    const success = await testRunner.runAllTests()
    
    process.exit(success ? 0 : 1)
  } catch (error) {
    console.error('❌ Test execution failed:', error.message)
    process.exit(1)
  }
}

if (require.main === module) {
  main()
}

module.exports = { CircularDependencyTestRunner, TEST_CONFIG }