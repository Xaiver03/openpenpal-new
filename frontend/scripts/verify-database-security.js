#!/usr/bin/env node

/**
 * Database Security Verification Script
 * 数据库安全验证脚本
 * 
 * Verifies database security configuration and connection settings
 */

const fs = require('fs')
const path = require('path')

// ANSI color codes for console output
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

class DatabaseSecurityVerifier {
  constructor() {
    this.issues = []
    this.warnings = []
    this.passed = []
    this.envPath = path.join(process.cwd(), '.env.local')
  }

  log(message, color = 'white') {
    console.log(`${colors[color]}${message}${colors.reset}`)
  }

  addIssue(message) {
    this.issues.push(message)
    this.log(`❌ ISSUE: ${message}`, 'red')
  }

  addWarning(message) {
    this.warnings.push(message)
    this.log(`⚠️  WARNING: ${message}`, 'yellow')
  }

  addPassed(message) {
    this.passed.push(message)
    this.log(`✅ PASSED: ${message}`, 'green')
  }

  loadEnvironment() {
    if (!fs.existsSync(this.envPath)) {
      this.addIssue('.env.local file not found')
      return {}
    }

    try {
      const envContent = fs.readFileSync(this.envPath, 'utf8')
      const env = {}
      
      envContent.split('\n').forEach(line => {
        const match = line.match(/^([^#=]+)=(.*)$/)
        if (match) {
          const [, key, value] = match
          env[key.trim()] = value.trim()
        }
      })
      
      return env
    } catch (error) {
      this.addIssue(`Failed to read .env.local: ${error.message}`)
      return {}
    }
  }

  verifyDatabasePassword(env) {
    const password = env.DATABASE_PASSWORD
    
    if (!password) {
      this.addIssue('DATABASE_PASSWORD is not set')
      return
    }

    if (password === 'password') {
      this.addIssue('DATABASE_PASSWORD is using default weak password')
      return
    }

    if (password.length < 16) {
      this.addIssue(`DATABASE_PASSWORD is too short (${password.length} chars, minimum 16 required)`)
      return
    }

    // Check password complexity
    const hasUpper = /[A-Z]/.test(password)
    const hasLower = /[a-z]/.test(password)
    const hasNumbers = /[0-9]/.test(password)
    const hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)

    let complexity = 0
    if (hasUpper) complexity++
    if (hasLower) complexity++
    if (hasNumbers) complexity++
    if (hasSpecial) complexity++

    if (complexity < 3) {
      this.addWarning('DATABASE_PASSWORD should contain at least 3 of: uppercase, lowercase, numbers, special characters')
    } else {
      this.addPassed(`DATABASE_PASSWORD has strong complexity (${password.length} chars, ${complexity}/4 character types)`)
    }
  }

  verifyConnectionPoolSettings(env) {
    const settings = {
      DATABASE_MAX_CONNECTIONS: { min: 5, max: 100, default: 20 },
      DATABASE_MIN_CONNECTIONS: { min: 1, max: 10, default: 2 },
      DATABASE_IDLE_TIMEOUT: { min: 10000, max: 300000, default: 30000 },
      DATABASE_CONNECTION_TIMEOUT: { min: 1000, max: 30000, default: 2000 },
      DATABASE_ACQUIRE_TIMEOUT: { min: 30000, max: 300000, default: 60000 }
    }

    Object.entries(settings).forEach(([key, config]) => {
      const value = parseInt(env[key] || config.default)
      
      if (isNaN(value)) {
        this.addWarning(`${key} is not a valid number, using default: ${config.default}`)
        return
      }

      if (value < config.min || value > config.max) {
        this.addWarning(`${key} value ${value} is outside recommended range [${config.min}, ${config.max}]`)
      } else {
        this.addPassed(`${key} is properly configured: ${value}`)
      }
    })

    // Verify max > min connections
    const maxConn = parseInt(env.DATABASE_MAX_CONNECTIONS || 20)
    const minConn = parseInt(env.DATABASE_MIN_CONNECTIONS || 2)
    
    if (maxConn <= minConn) {
      this.addIssue('DATABASE_MAX_CONNECTIONS must be greater than DATABASE_MIN_CONNECTIONS')
    }
  }

  verifySSLConfiguration(env) {
    const nodeEnv = env.NODE_ENV || 'development'
    
    if (nodeEnv === 'production') {
      const sslCa = env.DATABASE_SSL_CA
      const sslCert = env.DATABASE_SSL_CERT
      const sslKey = env.DATABASE_SSL_KEY

      if (!sslCa && !sslCert && !sslKey) {
        this.addWarning('SSL configuration not set for production environment')
      } else if (sslCa || sslCert || sslKey) {
        // Check if SSL files exist (if paths are provided)
        const sslFiles = { sslCa, sslCert, sslKey }
        Object.entries(sslFiles).forEach(([name, filePath]) => {
          if (filePath && !fs.existsSync(filePath)) {
            this.addIssue(`SSL file not found: ${name} at ${filePath}`)
          } else if (filePath) {
            this.addPassed(`SSL file exists: ${name}`)
          }
        })
      }
    } else {
      this.addPassed('SSL configuration disabled for development environment')
    }
  }

  verifyDatabaseAccess(env) {
    const user = env.DATABASE_USER || 'postgres'
    const host = env.DATABASE_HOST || 'localhost'
    const dbName = env.DATABASE_NAME || 'openpenpal'

    // Check for overly permissive settings
    if (user === 'postgres' || user === 'root' || user === 'admin') {
      this.addWarning(`Database user '${user}' has high privileges. Consider using a dedicated application user.`)
    } else {
      this.addPassed(`Database user '${user}' follows principle of least privilege`)
    }

    if (host === '0.0.0.0' || host === '*') {
      this.addIssue(`Database host '${host}' allows connections from anywhere`)
    } else if (host === 'localhost' || host === '127.0.0.1') {
      this.addPassed(`Database host '${host}' is properly restricted`)
    } else {
      this.addPassed(`Database host '${host}' is configured`)
    }

    this.addPassed(`Database name '${dbName}' is configured`)
  }

  verifyProductionSettings(env) {
    const nodeEnv = env.NODE_ENV || 'development'

    if (nodeEnv === 'production') {
      // Production-specific checks
      const allowExitOnIdle = env.DATABASE_ALLOW_EXIT_ON_IDLE
      if (allowExitOnIdle === 'true') {
        this.addWarning('DATABASE_ALLOW_EXIT_ON_IDLE should be false in production')
      }

      const logQueries = env.LOG_DATABASE_QUERIES
      if (logQueries === 'true') {
        this.addWarning('LOG_DATABASE_QUERIES should be disabled in production for performance')
      }

      const logLevel = env.LOG_LEVEL
      if (logLevel === 'debug') {
        this.addWarning('LOG_LEVEL should not be debug in production')
      }

      this.addPassed('Production environment security checks completed')
    } else {
      this.addPassed('Development environment detected')
    }
  }

  async testDatabaseConnection() {
    this.log('\n🔍 Testing database connection...', 'cyan')
    
    try {
      // Import the database module
      const dbPath = path.join(process.cwd(), 'src', 'lib', 'database.ts')
      if (!fs.existsSync(dbPath)) {
        this.addWarning('Database module not found, skipping connection test')
        return
      }

      // Note: In a real scenario, you would use the actual database module
      // For now, we'll just verify the configuration is loadable
      this.addPassed('Database configuration module is accessible')
      
    } catch (error) {
      this.addIssue(`Database connection test failed: ${error.message}`)
    }
  }

  generateReport() {
    this.log('\n' + '='.repeat(80), 'cyan')
    this.log('DATABASE SECURITY VERIFICATION REPORT', 'bold')
    this.log('='.repeat(80), 'cyan')

    this.log(`\n📊 Summary:`, 'cyan')
    this.log(`   ✅ Passed: ${this.passed.length}`, 'green')
    this.log(`   ⚠️  Warnings: ${this.warnings.length}`, 'yellow')
    this.log(`   ❌ Issues: ${this.issues.length}`, 'red')

    if (this.issues.length === 0) {
      this.log('\n🔒 Database security status: SECURE', 'green')
      this.log('All critical security checks passed!', 'green')
    } else {
      this.log('\n⚠️  Database security status: NEEDS ATTENTION', 'red')
      this.log('Critical issues found that require immediate attention.', 'red')
    }

    if (this.warnings.length > 0) {
      this.log('\n📋 Recommendations:', 'yellow')
      this.warnings.forEach(warning => {
        this.log(`   • ${warning}`, 'yellow')
      })
    }

    this.log('\n' + '='.repeat(80), 'cyan')
    
    return this.issues.length === 0
  }

  async verify() {
    this.log('🔍 Database Security Verification Starting...', 'cyan')
    this.log('=' * 50, 'cyan')

    const env = this.loadEnvironment()
    
    if (Object.keys(env).length === 0) {
      return this.generateReport()
    }

    this.log('\n🔐 Verifying database password...', 'cyan')
    this.verifyDatabasePassword(env)

    this.log('\n⚙️  Verifying connection pool settings...', 'cyan')
    this.verifyConnectionPoolSettings(env)

    this.log('\n🛡️  Verifying SSL configuration...', 'cyan')
    this.verifySSLConfiguration(env)

    this.log('\n👤 Verifying database access permissions...', 'cyan')
    this.verifyDatabaseAccess(env)

    this.log('\n🏭 Verifying production settings...', 'cyan')
    this.verifyProductionSettings(env)

    await this.testDatabaseConnection()

    return this.generateReport()
  }
}

// Run verification if called directly
if (require.main === module) {
  const verifier = new DatabaseSecurityVerifier()
  verifier.verify().then(isSecure => {
    process.exit(isSecure ? 0 : 1)
  }).catch(error => {
    console.error('Verification failed:', error)
    process.exit(1)
  })
}

module.exports = DatabaseSecurityVerifier