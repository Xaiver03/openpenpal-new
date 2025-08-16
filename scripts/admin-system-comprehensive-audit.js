#!/usr/bin/env node

const axios = require('axios');
const fs = require('fs');

// Configuration
const BASE_URL = 'http://localhost:8080';
const ADMIN_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJ0ZXN0LWFkbWluIiwicm9sZSI6InN1cGVyX2FkbWluIiwiaXNzIjoib3BlbnBlbnBhbCIsImV4cCI6MTc1NDE0MDA2NCwiaWF0IjoxNzU0MDUzNjY0LCJqdGkiOiI3ODgyZGRmMWEyZTk5MDA2YmE4MDFkNWZkYTMyM2NmMyJ9.D9VLMt14F4JpFV6k-r2pe7Rr_kziBmlpqTKsVo4VhaA';

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m'
};

// Set environment variables to bypass proxy
process.env.NO_PROXY = 'localhost,127.0.0.1,*.local,localhost:*,127.0.0.1:*';
process.env.no_proxy = 'localhost,127.0.0.1,*.local,localhost:*,127.0.0.1:*';
delete process.env.HTTP_PROXY;
delete process.env.HTTPS_PROXY;
delete process.env.http_proxy;
delete process.env.https_proxy;

// Configure axios
const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Authorization': `Bearer ${ADMIN_TOKEN}`,
    'Content-Type': 'application/json'
  }
});

// Helper functions
const log = (message, color = 'white') => {
  const timestamp = new Date().toISOString();
  console.log(`${colors[color]}[${timestamp}] ${message}${colors.reset}`);
};

// Admin System Structure from backend/main.go analysis
const ADMIN_ENDPOINTS = {
  dashboard: [
    { method: 'GET', path: '/api/v1/admin/dashboard/stats', description: 'Dashboard Statistics' },
    { method: 'GET', path: '/api/v1/admin/dashboard/activities', description: 'Recent Activities' },
    { method: 'GET', path: '/api/v1/admin/dashboard/analytics', description: 'Analytics Data' },
    { method: 'POST', path: '/api/v1/admin/seed-data', description: 'Inject Seed Data' }
  ],
  settings: [
    { method: 'GET', path: '/api/v1/admin/settings', description: 'Get System Settings' },
    { method: 'PUT', path: '/api/v1/admin/settings', description: 'Update System Settings' },
    { method: 'POST', path: '/api/v1/admin/settings', description: 'Reset System Settings' },
    { method: 'POST', path: '/api/v1/admin/settings/test-email', description: 'Test Email Config' }
  ],
  users: [
    { method: 'GET', path: '/api/v1/admin/users', description: 'Get User Management' },
    { method: 'GET', path: '/api/v1/admin/users/:id', description: 'Get Specific User' },
    { method: 'DELETE', path: '/api/v1/admin/users/:id', description: 'Deactivate User' },
    { method: 'POST', path: '/api/v1/admin/users/:id/reactivate', description: 'Reactivate User' }
  ],
  courier: [
    { method: 'GET', path: '/api/v1/admin/courier/applications', description: 'Get Pending Applications' },
    { method: 'POST', path: '/api/v1/admin/courier/:id/approve', description: 'Approve Courier Application' },
    { method: 'POST', path: '/api/v1/admin/courier/:id/reject', description: 'Reject Courier Application' }
  ],
  museum: [
    { method: 'POST', path: '/api/v1/admin/museum/items/:id/approve', description: 'Approve Museum Item' },
    { method: 'POST', path: '/api/v1/admin/museum/entries/:id/moderate', description: 'Moderate Museum Entry' },
    { method: 'GET', path: '/api/v1/admin/museum/entries/pending', description: 'Get Pending Museum Entries' },
    { method: 'POST', path: '/api/v1/admin/museum/exhibitions', description: 'Create Museum Exhibition' },
    { method: 'PUT', path: '/api/v1/admin/museum/exhibitions/:id', description: 'Update Museum Exhibition' },
    { method: 'DELETE', path: '/api/v1/admin/museum/exhibitions/:id', description: 'Delete Museum Exhibition' },
    { method: 'POST', path: '/api/v1/admin/museum/refresh-stats', description: 'Refresh Museum Stats' },
    { method: 'GET', path: '/api/v1/admin/museum/analytics', description: 'Get Museum Analytics' }
  ],
  analytics: [
    { method: 'GET', path: '/api/v1/admin/analytics/system', description: 'Get System Analytics' },
    { method: 'GET', path: '/api/v1/admin/analytics/dashboard', description: 'Get Analytics Dashboard' },
    { method: 'GET', path: '/api/v1/admin/analytics/reports', description: 'Get Analytics Reports' }
  ],
  moderation: [
    { method: 'POST', path: '/api/v1/admin/moderation/review', description: 'Review Content' },
    { method: 'GET', path: '/api/v1/admin/moderation/queue', description: 'Get Moderation Queue' },
    { method: 'GET', path: '/api/v1/admin/moderation/stats', description: 'Get Moderation Stats' },
    { method: 'GET', path: '/api/v1/admin/moderation/sensitive-words', description: 'Get Sensitive Words' },
    { method: 'POST', path: '/api/v1/admin/moderation/sensitive-words', description: 'Add Sensitive Word' },
    { method: 'PUT', path: '/api/v1/admin/moderation/sensitive-words/:id', description: 'Update Sensitive Word' },
    { method: 'DELETE', path: '/api/v1/admin/moderation/sensitive-words/:id', description: 'Delete Sensitive Word' },
    { method: 'GET', path: '/api/v1/admin/moderation/rules', description: 'Get Moderation Rules' },
    { method: 'POST', path: '/api/v1/admin/moderation/rules', description: 'Add Moderation Rule' },
    { method: 'PUT', path: '/api/v1/admin/moderation/rules/:id', description: 'Update Moderation Rule' },
    { method: 'DELETE', path: '/api/v1/admin/moderation/rules/:id', description: 'Delete Moderation Rule' }
  ],
  credits: [
    { method: 'GET', path: '/api/v1/admin/credits/users/:user_id', description: 'Get User Credit' },
    { method: 'POST', path: '/api/v1/admin/credits/users/add-points', description: 'Add User Points' },
    { method: 'POST', path: '/api/v1/admin/credits/users/spend-points', description: 'Spend User Points' },
    { method: 'GET', path: '/api/v1/admin/credits/leaderboard', description: 'Get Credits Leaderboard' },
    { method: 'GET', path: '/api/v1/admin/credits/rules', description: 'Get Credit Rules' }
  ],
  ai: [
    { method: 'GET', path: '/api/v1/admin/ai/config', description: 'Get AI Config' },
    { method: 'PUT', path: '/api/v1/admin/ai/config', description: 'Update AI Config' },
    { method: 'GET', path: '/api/v1/admin/ai/monitoring', description: 'Get AI Monitoring' },
    { method: 'GET', path: '/api/v1/admin/ai/analytics', description: 'Get AI Analytics' },
    { method: 'GET', path: '/api/v1/admin/ai/logs', description: 'Get AI Logs' },
    { method: 'POST', path: '/api/v1/admin/ai/test-provider', description: 'Test AI Provider' }
  ],
  shop: [
    { method: 'POST', path: '/api/v1/admin/shop/products', description: 'Create Product' },
    { method: 'PUT', path: '/api/v1/admin/shop/products/:id', description: 'Update Product' },
    { method: 'DELETE', path: '/api/v1/admin/shop/products/:id', description: 'Delete Product' },
    { method: 'PUT', path: '/api/v1/admin/shop/orders/:id/status', description: 'Update Order Status' },
    { method: 'GET', path: '/api/v1/admin/shop/stats', description: 'Get Shop Statistics' }
  ]
};

let auditResults = {
  totalEndpoints: 0,
  implemented: 0,
  working: 0,
  failing: 0,
  notImplemented: 0,
  categories: {},
  issues: [],
  recommendations: []
};

async function testEndpoint(category, endpoint) {
  const { method, path, description } = endpoint;
  
  auditResults.totalEndpoints++;
  
  if (!auditResults.categories[category]) {
    auditResults.categories[category] = {
      total: 0,
      working: 0,
      failing: 0,
      notImplemented: 0
    };
  }
  auditResults.categories[category].total++;
  
  // Replace placeholder params with test values
  let testPath = path
    .replace(':id', 'test-id-123')
    .replace(':user_id', 'test-admin');
  
  try {
    let response;
    const config = { validateStatus: () => true }; // Don't throw on any status code
    
    switch (method) {
      case 'GET':
        response = await api.get(testPath, config);
        break;
      case 'POST':
        const postData = getTestData(path);
        response = await api.post(testPath, postData, config);
        break;
      case 'PUT':
        const putData = getTestData(path);
        response = await api.put(testPath, putData, config);
        break;
      case 'DELETE':
        response = await api.delete(testPath, config);
        break;
      default:
        throw new Error(`Unsupported method: ${method}`);
    }
    
    if (response.status === 404) {
      log(`❌ NOT IMPLEMENTED: ${method} ${path} - ${description}`, 'red');
      auditResults.notImplemented++;
      auditResults.categories[category].notImplemented++;
      auditResults.issues.push(`${method} ${path} - Endpoint not implemented`);
    } else if (response.status >= 200 && response.status < 300) {
      log(`✅ WORKING: ${method} ${path} - ${description} (${response.status})`, 'green');
      auditResults.working++;
      auditResults.categories[category].working++;
      auditResults.implemented++;
    } else if (response.status >= 400 && response.status < 500) {
      // Client errors might be expected (validation, etc.)
      if (response.status === 401 || response.status === 403) {
        log(`🔒 AUTH ISSUE: ${method} ${path} - ${description} (${response.status})`, 'yellow');
        auditResults.issues.push(`${method} ${path} - Authentication/Authorization issue (${response.status})`);
      } else {
        log(`⚠️  CLIENT ERROR: ${method} ${path} - ${description} (${response.status})`, 'yellow');
      }
      auditResults.implemented++;
      auditResults.categories[category].working++;
    } else {
      log(`❌ SERVER ERROR: ${method} ${path} - ${description} (${response.status})`, 'red');
      auditResults.failing++;
      auditResults.categories[category].failing++;
      auditResults.implemented++;
      auditResults.issues.push(`${method} ${path} - Server error (${response.status})`);
    }
    
  } catch (error) {
    log(`❌ REQUEST FAILED: ${method} ${path} - ${description} - ${error.message}`, 'red');
    auditResults.failing++;
    auditResults.categories[category].failing++;
    auditResults.issues.push(`${method} ${path} - Request failed: ${error.message}`);
  }
}

function getTestData(path) {
  // Provide appropriate test data based on the endpoint
  if (path.includes('settings')) {
    return { maintenance_mode: false, max_upload_size: 10485760 };
  } else if (path.includes('sensitive-words')) {
    return { word: 'testword', severity: 'medium', category: 'spam' };
  } else if (path.includes('moderation/rules')) {
    return { name: 'Test Rule', pattern: 'test.*pattern', action: 'flag' };
  } else if (path.includes('credits')) {
    return { user_id: 'test-admin', amount: 100, reason: 'Admin test' };
  } else if (path.includes('ai/config')) {
    return { provider: 'siliconflow', api_key: 'test-key', enabled: true };
  } else if (path.includes('shop/products')) {
    return { name: 'Test Product', price: 99.99, description: 'Test product description' };
  } else if (path.includes('test-email')) {
    return { recipient: 'test@example.com' };
  } else if (path.includes('exhibitions')) {
    return { title: 'Test Exhibition', description: 'Test exhibition description' };
  }
  return {}; // Default empty object
}

async function performComprehensiveAudit() {
  log('🔍 STARTING COMPREHENSIVE ADMIN SYSTEM AUDIT', 'cyan');
  log(`🌐 Target: ${BASE_URL}`, 'cyan');
  log(`🔑 Token: ${ADMIN_TOKEN.substring(0, 20)}...`, 'cyan');
  log('', 'white');
  
  // Test all admin endpoints by category
  for (const [category, endpoints] of Object.entries(ADMIN_ENDPOINTS)) {
    log(`\n📁 Testing ${category.toUpperCase()} Module (${endpoints.length} endpoints)`, 'cyan');
    
    for (const endpoint of endpoints) {
      await testEndpoint(category, endpoint);
      // Small delay to avoid overwhelming the server
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  }
  
  // Generate comprehensive report
  generateAuditReport();
}

function generateAuditReport() {
  log('\n' + '='.repeat(80), 'cyan');
  log('📊 COMPREHENSIVE ADMIN SYSTEM AUDIT REPORT', 'cyan');
  log('='.repeat(80), 'cyan');
  
  // Overall Statistics
  log('\n📈 OVERALL STATISTICS:', 'white');
  log(`Total Endpoints Tested: ${auditResults.totalEndpoints}`, 'white');
  log(`✅ Working: ${auditResults.working} (${((auditResults.working/auditResults.totalEndpoints)*100).toFixed(1)}%)`, 'green');
  log(`❌ Failing: ${auditResults.failing} (${((auditResults.failing/auditResults.totalEndpoints)*100).toFixed(1)}%)`, 'red');
  log(`🚫 Not Implemented: ${auditResults.notImplemented} (${((auditResults.notImplemented/auditResults.totalEndpoints)*100).toFixed(1)}%)`, 'yellow');
  log(`📡 Implementation Rate: ${((auditResults.implemented/auditResults.totalEndpoints)*100).toFixed(1)}%`, 
      auditResults.implemented/auditResults.totalEndpoints > 0.8 ? 'green' : 'yellow');
  
  // Category Breakdown
  log('\n📋 MODULE BREAKDOWN:', 'white');
  Object.entries(auditResults.categories).forEach(([category, stats]) => {
    const successRate = stats.total > 0 ? ((stats.working / stats.total) * 100).toFixed(1) : 0;
    const implRate = stats.total > 0 ? (((stats.working + stats.failing) / stats.total) * 100).toFixed(1) : 0;
    
    log(`${category.padEnd(15)} | Working: ${stats.working.toString().padStart(2)}/${stats.total} (${successRate}%) | Impl: ${implRate}%`, 
        successRate > 80 ? 'green' : successRate > 50 ? 'yellow' : 'red');
  });
  
  // Critical Issues
  log('\n🚨 CRITICAL ISSUES:', 'red');
  if (auditResults.issues.length === 0) {
    log('None detected! 🎉', 'green');
  } else {
    auditResults.issues.slice(0, 10).forEach(issue => {
      log(`• ${issue}`, 'red');
    });
    if (auditResults.issues.length > 10) {
      log(`... and ${auditResults.issues.length - 10} more issues`, 'red');
    }
  }
  
  // Generate Recommendations
  generateRecommendations();
  
  // Production Readiness Assessment
  assessProductionReadiness();
  
  // Save detailed report to file
  saveDetailedReport();
}

function generateRecommendations() {
  log('\n💡 RECOMMENDATIONS:', 'yellow');
  
  if (auditResults.notImplemented > 0) {
    auditResults.recommendations.push('Implement missing admin endpoints for complete functionality');
  }
  
  if (auditResults.failing > 0) {
    auditResults.recommendations.push('Fix server errors in failing endpoints');
  }
  
  // Category-specific recommendations
  Object.entries(auditResults.categories).forEach(([category, stats]) => {
    const implRate = (stats.working + stats.failing) / stats.total;
    if (implRate < 0.5) {
      auditResults.recommendations.push(`${category} module needs significant development (${(implRate*100).toFixed(1)}% implemented)`);
    }
  });
  
  if (auditResults.recommendations.length === 0) {
    log('System appears to be well-implemented! 🎉', 'green');
  } else {
    auditResults.recommendations.forEach(rec => {
      log(`• ${rec}`, 'yellow');
    });
  }
}

function assessProductionReadiness() {
  log('\n🏭 PRODUCTION READINESS ASSESSMENT:', 'cyan');
  
  const implementationRate = auditResults.implemented / auditResults.totalEndpoints;
  const workingRate = auditResults.working / auditResults.totalEndpoints;
  const errorRate = auditResults.failing / auditResults.totalEndpoints;
  
  let readinessScore = 0;
  let readinessLevel = 'Not Ready';
  let readinessColor = 'red';
  
  if (implementationRate >= 0.9 && workingRate >= 0.8 && errorRate < 0.1) {
    readinessScore = 95;
    readinessLevel = 'Production Ready';
    readinessColor = 'green';
  } else if (implementationRate >= 0.8 && workingRate >= 0.7 && errorRate < 0.2) {
    readinessScore = 80;
    readinessLevel = 'Near Production Ready';
    readinessColor = 'yellow';
  } else if (implementationRate >= 0.6 && workingRate >= 0.5) {
    readinessScore = 60;
    readinessLevel = 'Development Stage';
    readinessColor = 'yellow';
  } else {
    readinessScore = 30;
    readinessLevel = 'Early Development';
    readinessColor = 'red';
  }
  
  log(`🎯 Readiness Score: ${readinessScore}/100`, readinessColor);
  log(`📊 Status: ${readinessLevel}`, readinessColor);
  
  // Production checklist
  log('\n✅ PRODUCTION CHECKLIST:', 'white');
  const checks = [
    { name: 'Basic Admin Functions', passed: workingRate > 0.7 },
    { name: 'User Management', passed: auditResults.categories.users?.working > 0 },
    { name: 'Content Moderation', passed: auditResults.categories.moderation?.working > 0 },
    { name: 'System Configuration', passed: auditResults.categories.settings?.working > 0 },
    { name: 'Analytics & Reporting', passed: auditResults.categories.analytics?.working > 0 },
    { name: 'Low Error Rate', passed: errorRate < 0.1 }
  ];
  
  checks.forEach(check => {
    log(`${check.passed ? '✅' : '❌'} ${check.name}`, check.passed ? 'green' : 'red');
  });
}

function saveDetailedReport() {
  const report = {
    timestamp: new Date().toISOString(),
    summary: {
      totalEndpoints: auditResults.totalEndpoints,
      working: auditResults.working,
      failing: auditResults.failing,
      notImplemented: auditResults.notImplemented,
      implementationRate: (auditResults.implemented / auditResults.totalEndpoints * 100).toFixed(2),
      workingRate: (auditResults.working / auditResults.totalEndpoints * 100).toFixed(2)
    },
    categories: auditResults.categories,
    issues: auditResults.issues,
    recommendations: auditResults.recommendations,
    endpoints: ADMIN_ENDPOINTS
  };
  
  try {
    fs.writeFileSync('admin-audit-report.json', JSON.stringify(report, null, 2));
    log('\n💾 Detailed report saved to: admin-audit-report.json', 'green');
  } catch (error) {
    log(`\n❌ Failed to save report: ${error.message}`, 'red');
  }
}

// Main execution
performComprehensiveAudit().catch(error => {
  log(`💥 Audit failed with error: ${error.message}`, 'red');
  console.error(error);
  process.exit(1);
});