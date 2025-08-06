#!/usr/bin/env node

// Script to reset AI usage for testing purposes
const { Client } = require('pg');

async function resetAIUsage() {
    const client = new Client({
        host: 'localhost',
        port: 5432,
        user: 'rocalight',
        password: 'password',
        database: 'openpenpal'
    });

    try {
        await client.connect();
        console.log('✅ Connected to PostgreSQL database');

        // First, let's check what tables exist
        const tablesResult = await client.query(`
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name LIKE '%usage%'
            ORDER BY table_name
        `);
        
        console.log('\n📋 Tables with "usage" in name:');
        tablesResult.rows.forEach(row => {
            console.log('  -', row.table_name);
        });

        // Check for ai_usage_logs table
        const aiLogsResult = await client.query(`
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name LIKE '%ai%'
            ORDER BY table_name
        `);
        
        console.log('\n📋 Tables with "ai" in name:');
        aiLogsResult.rows.forEach(row => {
            console.log('  -', row.table_name);
        });

        // First, let's see the current usage
        const checkResult = await client.query(`
            SELECT u.username, us.* 
            FROM user_daily_usages us 
            JOIN users u ON us.user_id = u.id 
            WHERE u.username = 'admin'
            AND DATE(us.date) = CURRENT_DATE
        `);
        
        if (checkResult.rows.length > 0) {
            console.log('\n📊 Current usage for admin:');
            console.log('  - Inspirations Used:', checkResult.rows[0].inspirations_used);
            console.log('  - AI Replies Generated:', checkResult.rows[0].ai_replies_generated);
            console.log('  - Penpal Matches:', checkResult.rows[0].penpal_matches);
            console.log('  - Letters Curated:', checkResult.rows[0].letters_curated);
        }

        // Reset the usage for admin user for today
        const resetResult = await client.query(`
            UPDATE user_daily_usages 
            SET inspirations_used = 0,
                ai_replies_generated = 0,
                penpal_matches = 0,
                letters_curated = 0,
                updated_at = NOW()
            WHERE user_id = (SELECT id FROM users WHERE username = 'admin')
            AND DATE(date) = CURRENT_DATE
            RETURNING *
        `);

        if (resetResult.rowCount > 0) {
            console.log('\n✅ Successfully reset AI usage for admin user!');
            console.log('  New values:');
            console.log('  - Inspirations Used:', resetResult.rows[0].inspirations_used);
            console.log('  - AI Replies Generated:', resetResult.rows[0].ai_replies_generated);
            console.log('  - Penpal Matches:', resetResult.rows[0].penpal_matches);
            console.log('  - Letters Curated:', resetResult.rows[0].letters_curated);
        } else {
            console.log('\n⚠️  No usage record found for admin user today');
        }

        // Also check if there are any active users
        const usersResult = await client.query(`
            SELECT username, email, role 
            FROM users 
            WHERE is_active = true 
            ORDER BY created_at 
            LIMIT 10
        `);

        console.log('\n👥 Active users in database:');
        usersResult.rows.forEach(user => {
            console.log(`  - ${user.username} (${user.email}) - Role: ${user.role}`);
        });

    } catch (error) {
        console.error('❌ Error:', error.message);
    } finally {
        await client.end();
        console.log('\n👋 Disconnected from database');
    }
}

// Run the reset
resetAIUsage();