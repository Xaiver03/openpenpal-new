#!/usr/bin/env node

const bcrypt = require('bcrypt');

// Test password hashes from database
const tests = [
    {
        username: 'courier_level1', 
        password: 'secret',
        hash: '$2a$10$KuNOKKOmFExYEe/BYHOQWOtuwywR3mHeOeBm7On0ZAozMWVqcmoU.'
    },
    {
        username: 'admin',
        password: 'admin123', 
        hash: '$2a$10$cH8Xq3cHw.nxkHBtepdYBekdP/85F1cn1LMBqii7tjB.VSmjInf/i'
    }
];

async function testPasswordVerification() {
    console.log('🔐 Testing Password Verification...\n');
    
    for (const test of tests) {
        console.log(`Testing ${test.username}:`);
        console.log(`  Password: ${test.password}`);
        console.log(`  Hash: ${test.hash}`);
        
        try {
            const isValid = await bcrypt.compare(test.password, test.hash);
            console.log(`  ✅ Result: ${isValid ? 'VALID' : 'INVALID'}`);
            
            if (!isValid) {
                console.log(`  🔧 Generating correct hash for '${test.password}':`);
                const correctHash = await bcrypt.hash(test.password, 10);
                console.log(`  📝 Correct hash: ${correctHash}`);
            }
        } catch (error) {
            console.log(`  ❌ Error: ${error.message}`);
        }
        console.log('');
    }
    
    // Test a few common passwords to see what works
    console.log('🧪 Testing common password variations...');
    const variations = ['secret', 'admin123', 'courier_level1', 'admin'];
    
    for (const variation of variations) {
        console.log(`\nTesting password: "${variation}"`);
        for (const test of tests) {
            try {
                const isValid = await bcrypt.compare(variation, test.hash);
                if (isValid) {
                    console.log(`  ✅ MATCH: "${variation}" works for ${test.username}`);
                }
            } catch (error) {
                console.log(`  ❌ Error testing ${test.username}: ${error.message}`);
            }
        }
    }
}

testPasswordVerification().catch(console.error);