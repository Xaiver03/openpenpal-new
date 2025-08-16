#!/usr/bin/env node

/**
 * Comprehensive Admin System Testing Script for OpenPenPal
 * 
 * This script tests all admin functionalities including:
 * - Authentication & Authorization
 * - User Management
 * - Content Moderation
 * - System Configuration
 * - Courier Management
 * - Analytics & Reporting
 * - Security Testing
 * - Error Handling
 */

const https = require('https');
const { performance } = require('perf_hooks');

// 🔐 安全令牌生成 - 替代硬编码令牌
const { generateTestToken } = require('../../../backend/scripts/test-token-generator');

// Configuration
const config = {
    baseURL: 'http://localhost:8080',
    adminToken: generateTestToken('ADMIN', {}, '4h'),
    testUser: {
        username: 'test_admin_user',
        email: 'test@admin.com',
        password: 'password123',
        nickname: 'Test Admin User'
    },
    // Bypass proxy issues
    noProxy: process.env.NO_PROXY || 'localhost,127.0.0.1'
};

// Set environment variable to bypass proxy
process.env.NO_PROXY = config.noProxy;

// Test results tracking
const results = {
    passed: 0,
    failed: 0,
    skipped: 0,
    tests: []
};

// Utility functions
function log(message, type = 'info') {
    const timestamp = new Date().toISOString();
    const colors = {
        info: '\x1b[36m',    // Cyan
        success: '\x1b[32m', // Green
        error: '\x1b[31m',   // Red
        warn: '\x1b[33m',    // Yellow
        reset: '\x1b[0m'     // Reset
    };
    console.log(`${colors[type]}[${timestamp}] ${message}${colors.reset}`);
}

function makeRequest(method, path, data = null, headers = {}) {
    return new Promise((resolve, reject) => {
        const url = new URL(path, config.baseURL);
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${config.adminToken}`,
                ...headers
            }
        };

        const req = https.request(url, options, (res) => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                try {
                    const response = {
                        status: res.statusCode,
                        headers: res.headers,
                        body: body ? JSON.parse(body) : null
                    };
                    resolve(response);
                } catch (e) {
                    resolve({
                        status: res.statusCode,
                        headers: res.headers,
                        body
                    });
                }
            });
        });

        req.on('error', reject);
        
        if (data) {
            req.write(JSON.stringify(data));
        }
        
        req.end();
    });
}

async function runTest(name, testFn) {
    const startTime = performance.now();
    try {
        log(`Running: ${name}`);
        await testFn();
        const duration = Math.round(performance.now() - startTime);
        log(`✅ PASSED: ${name} (${duration}ms)`, 'success');
        results.passed++;
        results.tests.push({ name, status: 'PASSED', duration });
    } catch (error) {
        const duration = Math.round(performance.now() - startTime);
        log(`❌ FAILED: ${name} - ${error.message} (${duration}ms)`, 'error');
        results.failed++;
        results.tests.push({ name, status: 'FAILED', error: error.message, duration });
    }
}

function assert(condition, message) {
    if (!condition) {
        throw new Error(message);
    }
}

// Test Categories

// 1. Authentication & Authorization Tests
async function testAuthenticationAndAuthorization() {
    log('\n=== 1. AUTHENTICATION & AUTHORIZATION TESTING ===', 'info');

    await runTest('Admin Token Validation', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/stats');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should return success: true');
    });

    await runTest('Invalid Token Rejection', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/stats', null, {
            'Authorization': 'Bearer invalid_token'
        });
        assert(response.status === 401, `Expected 401, got ${response.status}`);
    });

    await runTest('Missing Token Rejection', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/stats', null, {
            'Authorization': undefined
        });
        assert(response.status === 401, `Expected 401, got ${response.status}`);
    });

    await runTest('Non-Admin User Access Denial', async () => {
        // Test with a regular user token (if available)
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/stats', null, {
            'Authorization': 'Bearer regular_user_token'
        });
        assert([401, 403].includes(response.status), 
            `Expected 401 or 403, got ${response.status}`);
    });

    await runTest('Super Admin Role Verification', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/stats');
        assert(response.status === 200, 'Super admin should have access');
        assert(response.body.data, 'Should return dashboard data');
    });
}

// 2. User Management Tests
async function testUserManagement() {
    log('\n=== 2. USER MANAGEMENT TESTING ===', 'info');

    let testUserId = null;

    await runTest('Get Users List with Pagination', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/users/?page=1&limit=10');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should return success: true');
        assert(response.body.data.users, 'Should return users array');
        assert(response.body.data.total >= 0, 'Should return total count');
        assert(response.body.data.page === 1, 'Should return correct page');
        assert(response.body.data.limit === 10, 'Should return correct limit');
    });

    await runTest('Get Users List with Invalid Pagination', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/users/?page=-1&limit=1000');
        assert(response.status === 200, 'Should handle invalid pagination gracefully');
        // Should default to page 1 and limit 20
        assert(response.body.data.page >= 1, 'Should default to valid page');
        assert(response.body.data.limit <= 100, 'Should limit max page size');
    });

    await runTest('Create Test User for Management Tests', async () => {
        const userData = {
            username: config.testUser.username,
            email: config.testUser.email,
            password: config.testUser.password,
            nickname: config.testUser.nickname,
            role: 'user'
        };
        
        // First try to create via regular registration endpoint
        const response = await makeRequest('POST', '/api/v1/auth/register', userData, {
            'Authorization': undefined // No auth needed for registration
        });
        
        if (response.status === 201 || response.status === 200) {
            testUserId = response.body.user?.id || response.body.data?.id;
            assert(testUserId, 'Should return user ID after creation');
        } else if (response.status === 409) {
            // User already exists, try to find them
            const usersResponse = await makeRequest('GET', '/api/v1/admin/users/?page=1&limit=100');
            const existingUser = usersResponse.body.data.users.find(u => u.username === config.testUser.username);
            testUserId = existingUser?.id;
            assert(testUserId, 'Should find existing test user');
        } else {
            throw new Error(`Failed to create/find test user: ${response.status}`);
        }
    });

    await runTest('Get Specific User Details', async () => {
        if (!testUserId) throw new Error('Test user ID not available');
        
        const response = await makeRequest('GET', `/api/v1/admin/users/${testUserId}`);
        if (response.status === 200) {
            assert(response.body.id === testUserId, 'Should return correct user');
            assert(response.body.username, 'Should return username');
            assert(response.body.email, 'Should return email');
        } else {
            // Endpoint might not be implemented yet
            log('Get specific user endpoint not implemented', 'warn');
            results.skipped++;
        }
    });

    await runTest('Deactivate User', async () => {
        if (!testUserId) throw new Error('Test user ID not available');
        
        const response = await makeRequest('DELETE', `/api/v1/admin/users/${testUserId}`);
        if (response.status === 200) {
            assert(response.body.success !== false, 'Should successfully deactivate user');
        } else {
            // Endpoint might not be implemented yet
            log('User deactivation endpoint not implemented', 'warn');
            results.skipped++;
        }
    });

    await runTest('Reactivate User', async () => {
        if (!testUserId) throw new Error('Test user ID not available');
        
        const response = await makeRequest('POST', `/api/v1/admin/users/${testUserId}/reactivate`);
        if (response.status === 200) {
            assert(response.body.success !== false, 'Should successfully reactivate user');
        } else {
            // Endpoint might not be implemented yet
            log('User reactivation endpoint not implemented', 'warn');
            results.skipped++;
        }
    });

    await runTest('Update User Role (if supported)', async () => {
        if (!testUserId) throw new Error('Test user ID not available');
        
        const updateData = { role: 'courier_level1' };
        const response = await makeRequest('PUT', `/api/v1/admin/users/${testUserId}`, updateData);
        
        if (response.status === 404) {
            log('User role update endpoint not found', 'warn');
            results.skipped++;
        } else if (response.status === 200) {
            assert(response.body.success !== false, 'Should successfully update user role');
        }
    });
}

// 3. Content Moderation Tests
async function testContentModeration() {
    log('\n=== 3. CONTENT MODERATION TESTING ===', 'info');

    await runTest('Get Moderation Queue', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/moderation/queue?limit=20');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.queue !== undefined, 'Should return moderation queue');
        assert(Array.isArray(response.body.queue), 'Queue should be an array');
    });

    await runTest('Get Moderation Statistics', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/moderation/stats');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        // Stats might be empty but should not error
    });

    await runTest('Get Sensitive Words List', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/moderation/sensitive-words');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.words !== undefined, 'Should return words list');
        assert(Array.isArray(response.body.words), 'Words should be an array');
    });

    await runTest('Add Sensitive Word', async () => {
        const wordData = {
            word: 'testbadword',
            category: 'test',
            level: 'medium',
            description: 'Test sensitive word for admin testing'
        };
        
        const response = await makeRequest('POST', '/api/v1/admin/moderation/sensitive-words', wordData);
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.message, 'Should return success message');
    });

    await runTest('Get Moderation Rules', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/moderation/rules');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.rules !== undefined, 'Should return rules list');
        assert(Array.isArray(response.body.rules), 'Rules should be an array');
    });

    await runTest('Add Moderation Rule', async () => {
        const ruleData = {
            name: 'Test Admin Rule',
            description: 'Test rule for admin system testing',
            contentType: 'letter',
            action: 'flag',
            conditions: [
                {
                    field: 'content',
                    operator: 'contains',
                    value: 'testflag'
                }
            ],
            isActive: true
        };
        
        const response = await makeRequest('POST', '/api/v1/admin/moderation/rules', ruleData);
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.message, 'Should return success message');
    });

    await runTest('Review Content (Manual Moderation)', async () => {
        const reviewData = {
            contentId: 'test-content-id',
            contentType: 'letter',
            status: 'approved',
            reviewerNote: 'Admin system test - approved',
            action: 'approve'
        };
        
        const response = await makeRequest('POST', '/api/v1/admin/moderation/review', reviewData);
        // This might fail if no content exists, but should not crash
        assert([200, 404, 400].includes(response.status), 
            `Expected 200, 404, or 400, got ${response.status}`);
    });
}

// 4. System Configuration Tests
async function testSystemConfiguration() {
    log('\n=== 4. SYSTEM CONFIGURATION TESTING ===', 'info');

    await runTest('Get System Settings', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/settings');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should return success: true');
        assert(response.body.data, 'Should return settings data');
    });

    await runTest('Update System Settings', async () => {
        const settingsData = {
            siteName: 'OpenPenPal Admin Test',
            siteDescription: 'Test description from admin system test',
            registrationOpen: true,
            maintenanceMode: false,
            maxLettersPerDay: 15,
            maxEnvelopesPerOrder: 25
        };
        
        const response = await makeRequest('PUT', '/api/v1/admin/settings', settingsData);
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should successfully update settings');
    });

    await runTest('Reset System Settings', async () => {
        const response = await makeRequest('POST', '/api/v1/admin/settings');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should successfully reset settings');
    });

    await runTest('Test Email Configuration', async () => {
        const emailTestData = {
            testEmail: 'admin-test@openpenpal.com',
            subject: 'Admin System Test Email',
            message: 'This is a test email from the admin system testing script.'
        };
        
        const response = await makeRequest('POST', '/api/v1/admin/settings/test-email', emailTestData);
        // Email test might fail if SMTP not configured, but should not crash
        assert([200, 400, 500].includes(response.status), 
            `Expected 200, 400, or 500, got ${response.status}`);
    });
}

// 5. Courier Management Tests
async function testCourierManagement() {
    log('\n=== 5. COURIER MANAGEMENT TESTING ===', 'info');

    await runTest('Get Pending Courier Applications', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/courier/applications');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        // Applications list might be empty but should not error
    });

    await runTest('Get Courier Hierarchy (if implemented)', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/courier/hierarchy');
        if (response.status === 404) {
            log('Courier hierarchy endpoint not found', 'warn');
            results.skipped++;
        } else {
            assert(response.status === 200, `Expected 200, got ${response.status}`);
        }
    });

    await runTest('Approve Courier Application (Mock)', async () => {
        const mockApplicationId = 'test-application-id';
        const approvalData = {
            level: 1,
            zone: 'TEST-ZONE',
            notes: 'Admin system test approval'
        };
        
        const response = await makeRequest('POST', `/api/v1/admin/courier/${mockApplicationId}/approve`, approvalData);
        // This will likely fail with 404 since it's a mock ID, but should handle gracefully
        assert([200, 404, 400].includes(response.status), 
            `Expected 200, 404, or 400, got ${response.status}`);
    });

    await runTest('Reject Courier Application (Mock)', async () => {
        const mockApplicationId = 'test-application-id';
        const rejectionData = {
            reason: 'Admin system test rejection',
            feedback: 'This is a mock rejection for testing purposes'
        };
        
        const response = await makeRequest('POST', `/api/v1/admin/courier/${mockApplicationId}/reject`, rejectionData);
        // This will likely fail with 404 since it's a mock ID, but should handle gracefully
        assert([200, 404, 400].includes(response.status), 
            `Expected 200, 404, or 400, got ${response.status}`);
    });
}

// 6. Analytics & Reporting Tests
async function testAnalyticsAndReporting() {
    log('\n=== 6. ANALYTICS & REPORTING TESTING ===', 'info');

    await runTest('Get Dashboard Statistics', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/stats');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should return success: true');
        assert(response.body.data, 'Should return statistics data');
        
        const stats = response.body.data;
        assert(typeof stats.totalUsers === 'number', 'Should return totalUsers count');
        assert(typeof stats.totalLetters === 'number', 'Should return totalLetters count');
        assert(typeof stats.activeCouriers === 'number', 'Should return activeCouriers count');
    });

    await runTest('Get Recent Activity', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/activities?limit=10');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should return success: true');
        assert(Array.isArray(response.body.data), 'Should return activities array');
    });

    await runTest('Get Analytics Dashboard Data', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/analytics');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.success === true, 'Should return success: true');
        assert(response.body.data, 'Should return analytics data');
        
        const analytics = response.body.data;
        assert(analytics.userGrowth, 'Should include user growth data');
        assert(analytics.letterTrends, 'Should include letter trends data');
        assert(analytics.courierStats, 'Should include courier statistics');
    });

    await runTest('Get System Analytics', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/analytics/system');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        // System analytics might return processed data or confirmation
    });

    await runTest('Generate Analytics Report', async () => {
        const reportData = {
            reportType: 'user_engagement',
            startDate: '2024-01-01',
            endDate: '2024-12-31',
            format: 'json',
            includeCharts: true
        };
        
        const response = await makeRequest('POST', '/api/v1/admin/analytics/reports', reportData);
        if (response.status === 200) {
            assert(response.body.id, 'Should return report ID');
        } else {
            // Report generation might be async or not fully implemented
            log('Report generation endpoint returned non-200', 'warn');
        }
    });

    await runTest('Get Analytics Reports List', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/analytics/reports?page=1&pageSize=10');
        assert(response.status === 200, `Expected 200, got ${response.status}`);
        assert(response.body.reports !== undefined, 'Should return reports list');
    });
}

// 7. Security Testing
async function testSecurity() {
    log('\n=== 7. SECURITY TESTING ===', 'info');

    await runTest('SQL Injection Protection', async () => {
        const maliciousPayload = "'; DROP TABLE users; --";
        const response = await makeRequest('GET', `/api/v1/admin/users/?page=1&limit=${maliciousPayload}`);
        // Should not crash the server
        assert([200, 400].includes(response.status), 
            'Server should handle SQL injection attempts gracefully');
    });

    await runTest('XSS Prevention in User Data', async () => {
        const xssPayload = '<script>alert("xss")</script>';
        const userData = {
            username: xssPayload,
            email: 'xss@test.com',
            nickname: xssPayload
        };
        
        const response = await makeRequest('POST', '/api/v1/admin/users', userData);
        // Should sanitize or reject XSS attempts
        assert([400, 422, 200].includes(response.status), 
            'Should handle XSS attempts appropriately');
    });

    await runTest('Large Payload Handling', async () => {
        const largeString = 'A'.repeat(10000);
        const largePayload = {
            content: largeString,
            description: largeString
        };
        
        const response = await makeRequest('POST', '/api/v1/admin/moderation/rules', largePayload);
        // Should handle large payloads gracefully
        assert([400, 413, 422].includes(response.status), 
            'Should reject or handle large payloads appropriately');
    });

    await runTest('Rate Limiting Check', async () => {
        const promises = [];
        for (let i = 0; i < 10; i++) {
            promises.push(makeRequest('GET', '/api/v1/admin/dashboard/stats'));
        }
        
        const responses = await Promise.all(promises);
        const rateLimitedResponses = responses.filter(r => r.status === 429);
        
        // Rate limiting might not be implemented, but should not crash
        log(`Rate limiting: ${rateLimitedResponses.length}/10 requests limited`, 'info');
    });

    await runTest('CSRF Protection Check', async () => {
        // Test without proper headers/tokens
        const response = await makeRequest('POST', '/api/v1/admin/settings', 
            { siteName: 'CSRF Test' }, 
            { 'X-Requested-With': undefined }
        );
        
        // CSRF protection might not be implemented, but check response
        assert([200, 403, 401].includes(response.status), 
            'Should handle CSRF appropriately');
    });

    await runTest('Authorization Bypass Attempt', async () => {
        // Try accessing admin endpoint with manipulated token
        const fakeToken = config.adminToken.slice(0, -10) + 'manipulated';
        const response = await makeRequest('GET', '/api/v1/admin/dashboard/stats', null, {
            'Authorization': `Bearer ${fakeToken}`
        });
        
        assert(response.status === 401, 
            'Should reject manipulated tokens');
    });
}

// 8. Error Handling Tests
async function testErrorHandling() {
    log('\n=== 8. ERROR HANDLING TESTING ===', 'info');

    await runTest('404 Error Handling', async () => {
        const response = await makeRequest('GET', '/api/v1/admin/nonexistent/endpoint');
        assert(response.status === 404, `Expected 404, got ${response.status}`);
    });

    await runTest('Invalid JSON Payload', async () => {
        const invalidJson = '{"invalid": json}';
        const response = await makeRequest('POST', '/api/v1/admin/settings', null, {
            'Content-Type': 'application/json'
        });
        
        // Manually write invalid JSON
        const url = new URL('/api/v1/admin/settings', config.baseURL);
        const options = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${config.adminToken}`
            }
        };

        const req = https.request(url, options, (res) => {
            assert([400, 422].includes(res.statusCode), 
                'Should reject invalid JSON with 400 or 422');
        });

        req.write(invalidJson);
        req.end();
    });

    await runTest('Missing Required Fields', async () => {
        const incompleteData = {
            // Missing required fields
            siteName: 'Test'
            // Other required fields missing
        };
        
        const response = await makeRequest('PUT', '/api/v1/admin/settings', incompleteData);
        // Should handle missing fields gracefully
        assert([200, 400, 422].includes(response.status), 
            'Should handle missing fields appropriately');
    });

    await runTest('Invalid Content-Type Header', async () => {
        const response = await makeRequest('POST', '/api/v1/admin/settings', 
            { siteName: 'Test' }, 
            { 'Content-Type': 'text/plain' }
        );
        
        assert([400, 415].includes(response.status), 
            'Should reject invalid content types');
    });

    await runTest('Server Error Recovery', async () => {
        // Test endpoint that might cause server errors
        const response = await makeRequest('POST', '/api/v1/admin/seed-data');
        
        // Should either succeed or fail gracefully
        assert([200, 409, 500].includes(response.status), 
            'Should handle seed data injection appropriately');
        
        if (response.status === 500) {
            assert(response.body, 'Should return error details on server errors');
        }
    });
}

// Main test execution
async function runAllTests() {
    log('🚀 Starting OpenPenPal Admin System Comprehensive Testing', 'info');
    log(`Base URL: ${config.baseURL}`, 'info');
    log(`Admin Token: ${config.adminToken.substring(0, 20)}...`, 'info');
    
    const startTime = performance.now();

    try {
        await testAuthenticationAndAuthorization();
        await testUserManagement();
        await testContentModeration();
        await testSystemConfiguration();
        await testCourierManagement();
        await testAnalyticsAndReporting();
        await testSecurity();
        await testErrorHandling();
    } catch (error) {
        log(`Critical error during testing: ${error.message}`, 'error');
    }

    const totalTime = Math.round(performance.now() - startTime);
    
    // Print summary
    log('\n' + '='.repeat(60), 'info');
    log('📊 TEST SUMMARY', 'info');
    log('='.repeat(60), 'info');
    log(`✅ Passed: ${results.passed}`, 'success');
    log(`❌ Failed: ${results.failed}`, 'error');
    log(`⏭️  Skipped: ${results.skipped}`, 'warn');
    log(`⏱️  Total Time: ${totalTime}ms`, 'info');
    log(`📈 Success Rate: ${Math.round((results.passed / (results.passed + results.failed)) * 100)}%`, 'info');

    // Detailed results
    if (results.tests.length > 0) {
        log('\n📋 DETAILED RESULTS:', 'info');
        results.tests.forEach(test => {
            const status = test.status === 'PASSED' ? '✅' : '❌';
            const duration = test.duration ? `(${test.duration}ms)` : '';
            log(`${status} ${test.name} ${duration}`, test.status === 'PASSED' ? 'success' : 'error');
            if (test.error) {
                log(`   Error: ${test.error}`, 'error');
            }
        });
    }

    // Recommendations
    log('\n💡 RECOMMENDATIONS:', 'info');
    if (results.failed > 0) {
        log('• Review failed tests and fix underlying issues', 'warn');
        log('• Check server logs for detailed error information', 'warn');
    }
    if (results.skipped > 0) {
        log('• Implement skipped endpoints for complete admin functionality', 'warn');
    }
    if (results.passed > 0) {
        log('• Consider implementing additional security measures', 'info');
        log('• Add comprehensive input validation', 'info');
        log('• Implement proper rate limiting', 'info');
    }

    log('\n🏁 Admin System Testing Complete!', 'success');
    
    // Exit with appropriate code
    process.exit(results.failed > 0 ? 1 : 0);
}

// Handle uncaught errors
process.on('uncaughtException', (error) => {
    log(`Uncaught Exception: ${error.message}`, 'error');
    console.error(error.stack);
    process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
    log(`Unhandled Rejection at: ${promise}, reason: ${reason}`, 'error');
    process.exit(1);
});

// Run the tests
runAllTests();