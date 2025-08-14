#!/usr/bin/env node

/**
 * Follow System API Integration Test
 * Tests the complete Follow system backend integration
 */

const https = require('https');
const http = require('http');

const BASE_URL = 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;

// Test accounts
const TEST_USERS = {
    alice: { username: 'alice', password: 'secret123' },
    admin: { username: 'admin', password: 'admin123' }
};

let authTokens = {};

// HTTP request helper
function makeRequest(method, path, data = null, token = null) {
    return new Promise((resolve, reject) => {
        const url = new URL(`${API_BASE}${path}`);
        const options = {
            method,
            hostname: url.hostname,
            port: url.port,
            path: url.pathname + url.search,
            headers: {
                'Content-Type': 'application/json',
                ...(token && { 'Authorization': `Bearer ${token}` })
            }
        };

        const req = http.request(options, (res) => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                try {
                    const result = body ? JSON.parse(body) : {};
                    resolve({ status: res.statusCode, data: result, headers: res.headers });
                } catch (e) {
                    resolve({ status: res.statusCode, data: body, headers: res.headers });
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

// Test runner
async function runFollowSystemTests() {
    console.log('🚀 Starting Follow System Integration Tests');
    console.log('==========================================\n');

    try {
        // Step 1: Login test users
        console.log('📋 Step 1: Authenticating test users...');
        for (const [name, creds] of Object.entries(TEST_USERS)) {
            try {
                const response = await makeRequest('POST', '/auth/login', creds);
                if (response.status === 200 && response.data.token) {
                    authTokens[name] = response.data.token;
                    console.log(`✅ ${name} authenticated successfully`);
                } else {
                    console.log(`❌ ${name} authentication failed:`, response.status, response.data);
                }
            } catch (error) {
                console.log(`❌ ${name} authentication error:`, error.message);
            }
        }

        if (Object.keys(authTokens).length === 0) {
            throw new Error('No users authenticated - cannot proceed with tests');
        }

        console.log('\n📋 Step 2: Testing Follow System APIs...\n');

        // Test 2: Get user suggestions
        console.log('🔍 Testing user suggestions...');
        try {
            const response = await makeRequest('GET', '/follow/suggestions?limit=5', null, authTokens.alice);
            console.log(`Status: ${response.status}`);
            if (response.status === 200) {
                console.log(`✅ Got ${response.data.data?.suggestions?.length || 0} user suggestions`);
            } else {
                console.log(`⚠️ User suggestions: ${response.status} - ${JSON.stringify(response.data)}`);
            }
        } catch (error) {
            console.log(`❌ User suggestions error: ${error.message}`);
        }

        // Test 3: Search users
        console.log('\n🔍 Testing user search...');
        try {
            const response = await makeRequest('GET', '/follow/users/search?query=admin&limit=5', null, authTokens.alice);
            console.log(`Status: ${response.status}`);
            if (response.status === 200) {
                console.log(`✅ Found ${response.data.data?.users?.length || 0} users matching 'admin'`);
            } else {
                console.log(`⚠️ User search: ${response.status} - ${JSON.stringify(response.data)}`);
            }
        } catch (error) {
            console.log(`❌ User search error: ${error.message}`);
        }

        // Test 4: Follow user
        if (authTokens.alice && authTokens.admin) {
            console.log('\n👥 Testing follow user...');
            try {
                // Get admin user ID first (if possible)
                const adminUser = await makeRequest('GET', '/auth/me', null, authTokens.admin);
                if (adminUser.status === 200 && adminUser.data.user?.id) {
                    const adminId = adminUser.data.user.id;
                    
                    const followResponse = await makeRequest('POST', '/follow/users', {
                        user_id: adminId,
                        notification_enabled: true
                    }, authTokens.alice);
                    
                    console.log(`Follow Status: ${followResponse.status}`);
                    if (followResponse.status === 200) {
                        console.log(`✅ Alice successfully followed admin`);
                        console.log(`Follow response:`, followResponse.data);
                    } else {
                        console.log(`⚠️ Follow failed: ${followResponse.status} - ${JSON.stringify(followResponse.data)}`);
                    }
                } else {
                    console.log(`❌ Could not get admin user ID for follow test`);
                }
            } catch (error) {
                console.log(`❌ Follow user error: ${error.message}`);
            }
        }

        // Test 5: Get followers/following lists
        console.log('\n📊 Testing followers/following lists...');
        try {
            const followersResponse = await makeRequest('GET', '/follow/followers?limit=10', null, authTokens.alice);
            console.log(`Followers Status: ${followersResponse.status}`);
            if (followersResponse.status === 200) {
                console.log(`✅ Alice has ${followersResponse.data.data?.users?.length || 0} followers`);
            }

            const followingResponse = await makeRequest('GET', '/follow/following?limit=10', null, authTokens.alice);
            console.log(`Following Status: ${followingResponse.status}`);
            if (followingResponse.status === 200) {
                console.log(`✅ Alice is following ${followingResponse.data.data?.users?.length || 0} users`);
            }
        } catch (error) {
            console.log(`❌ Followers/following lists error: ${error.message}`);
        }

        // Test 6: Get follow status
        if (authTokens.alice && authTokens.admin) {
            console.log('\n🔍 Testing follow status check...');
            try {
                const adminUser = await makeRequest('GET', '/auth/me', null, authTokens.admin);
                if (adminUser.status === 200 && adminUser.data.user?.id) {
                    const adminId = adminUser.data.user.id;
                    
                    const statusResponse = await makeRequest('GET', `/follow/users/${adminId}/status`, null, authTokens.alice);
                    console.log(`Follow Status Check: ${statusResponse.status}`);
                    if (statusResponse.status === 200) {
                        console.log(`✅ Follow status retrieved:`, statusResponse.data.data);
                    }
                }
            } catch (error) {
                console.log(`❌ Follow status check error: ${error.message}`);
            }
        }

        console.log('\n🎉 Follow System Integration Test Complete!');
        console.log('==========================================');

    } catch (error) {
        console.error('💥 Test runner error:', error.message);
        process.exit(1);
    }
}

// Health check first
async function checkBackendHealth() {
    try {
        const response = await makeRequest('GET', '/../health');
        if (response.status === 200) {
            console.log('✅ Backend is healthy');
            return true;
        } else {
            console.log('❌ Backend health check failed:', response.status);
            return false;
        }
    } catch (error) {
        console.log('❌ Backend is not running:', error.message);
        return false;
    }
}

// Main execution
async function main() {
    console.log('🔍 Checking backend health...');
    const isHealthy = await checkBackendHealth();
    
    if (!isHealthy) {
        console.log('\n⚠️ Backend is not running or not healthy.');
        console.log('Please start the backend with: cd backend && go run main.go');
        console.log('Then run this test again.');
        process.exit(1);
    }

    console.log('\n✅ Backend is running, proceeding with Follow System tests...\n');
    await runFollowSystemTests();
}

// Execute if run directly
if (require.main === module) {
    main().catch(console.error);
}

module.exports = { runFollowSystemTests, makeRequest };