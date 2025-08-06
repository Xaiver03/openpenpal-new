#!/usr/bin/env node

/**
 * 初始化数据库测试用户脚本
 * Database Test Users Initialization Script
 * 
 * 这个脚本将测试用户数据持久化到PostgreSQL数据库中，提高测试稳定性
 */

const { Pool } = require('pg')
const bcrypt = require('bcryptjs')
const crypto = require('crypto')

// 数据库配置
const dbConfig = {
  user: process.env.DATABASE_USER || 'postgres',
  host: process.env.DATABASE_HOST || 'localhost',
  database: process.env.DATABASE_NAME || 'openpenpal',
  password: process.env.DATABASE_PASSWORD || 'OpenPenPal_Secure_DB_P@ssw0rd_2025',
  port: parseInt(process.env.DATABASE_PORT || '5432'),
}

const pool = new Pool(dbConfig)

// 生成标准用户ID
function generateStandardUserId(username, config) {
  const hash = crypto.createHash('sha256').update(`${username}_${config.level || 'user'}_2025`).digest('hex')
  return hash.substring(0, 16)
}

// 测试用户配置
const TEST_USERS = [
  // 基础管理员用户
  {
    id: 'admin_001',
    username: 'admin',
    email: 'admin@openpenpal.com',
    realName: '系统管理员',
    role: 'super_admin',
    permissions: [
      'MANAGE_USERS', 'VIEW_ANALYTICS', 'MODERATE_CONTENT',
      'MANAGE_SCHOOLS', 'MANAGE_EXHIBITIONS', 'SYSTEM_CONFIG',
      'AUDIT_SUBMISSIONS', 'HANDLE_REPORTS'
    ],
    school_code: 'SYSTEM',
    school_name: '系统管理',
    status: 'active',
    password: process.env.TEST_ACCOUNT_ADMIN_PASSWORD || 'admin123'
  },
  
  // 基础信使用户
  {
    id: 'courier_001',
    username: 'courier_building',
    email: 'courier@openpenpal.com',
    realName: '建筑楼信使',
    role: 'courier',
    permissions: [
      'DELIVER_LETTER', 'SCAN_CODE', 'VIEW_TASKS',
      'WRITE_LETTER', 'READ_LETTER', 'MANAGE_PROFILE'
    ],
    school_code: 'BJDX01',
    school_name: '北京大学',
    status: 'active',
    password: process.env.TEST_ACCOUNT_COURIER_BUILDING_PASSWORD || 'courier123'
  },
  
  {
    id: 'courier_002', 
    username: 'senior_courier',
    email: 'senior.courier@openpenpal.com',
    realName: '高级信使',
    role: 'senior_courier',
    permissions: [
      'DELIVER_LETTER', 'SCAN_CODE', 'VIEW_TASKS',
      'WRITE_LETTER', 'READ_LETTER', 'MANAGE_PROFILE',
      'VIEW_REPORTS'
    ],
    school_code: 'BJDX01',
    school_name: '北京大学',
    status: 'active',
    password: process.env.TEST_ACCOUNT_SENIOR_COURIER_PASSWORD || 'senior123'
  },
  
  {
    id: 'coord_001',
    username: 'coordinator',
    email: 'coordinator@openpenpal.com',
    realName: '信使协调员',
    role: 'courier_coordinator',
    permissions: [
      'DELIVER_LETTER', 'SCAN_CODE', 'VIEW_TASKS',
      'WRITE_LETTER', 'READ_LETTER', 'MANAGE_PROFILE',
      'MANAGE_COURIERS', 'ASSIGN_TASKS', 'VIEW_REPORTS'
    ],
    school_code: 'BJDX01',
    school_name: '北京大学',
    status: 'active',
    password: process.env.TEST_ACCOUNT_COORDINATOR_PASSWORD || 'coord123'
  }
]

// 层级信使配置
const COURIER_LEVELS = [
  {
    username: 'courier_level4_city',
    email: 'city.courier@openpenpal.com',
    level: 4,
    levelName: '四级信使（城市总代）',
    zoneCode: 'BEIJING',
    zoneType: 'city',
    description: '北京市信使总负责人，管理全市学校信使网络',
    permissions: [
      'courier_scan_code', 'courier_deliver_letter', 'courier_view_own_tasks',
      'courier_report_exception', 'courier_manage_subordinates', 'courier_assign_tasks',
      'courier_view_subordinate_reports', 'courier_create_subordinate',
      'courier_manage_school_zone', 'courier_view_school_analytics',
      'courier_coordinate_cross_zone', 'courier_manage_city_operations',
      'courier_create_school_courier', 'courier_view_city_analytics'
    ],
    password: process.env.TEST_ACCOUNT_COURIER_LEVEL4_CITY_PASSWORD || 'city123'
  },
  
  {
    username: 'courier_level3_school',
    email: 'school.courier@openpenpal.com',
    level: 3,
    levelName: '三级信使（校级）',
    zoneCode: 'BJDX',
    zoneType: 'school',
    parentUsername: 'courier_level4_city',
    description: '北京大学信使负责人，管理全校信使团队',
    permissions: [
      'courier_scan_code', 'courier_deliver_letter', 'courier_view_own_tasks',
      'courier_report_exception', 'courier_manage_subordinates', 'courier_assign_tasks',
      'courier_view_subordinate_reports', 'courier_create_subordinate',
      'courier_manage_school_zone', 'courier_view_school_analytics',
      'courier_coordinate_cross_zone'
    ],
    password: process.env.TEST_ACCOUNT_COURIER_LEVEL3_SCHOOL_PASSWORD || 'school123'
  },
  
  {
    username: 'courier_level2_zone',
    email: 'zone.courier@openpenpal.com',
    level: 2,
    levelName: '二级信使（片区/年级）',
    zoneCode: 'BJDX_EAST',
    zoneType: 'zone',
    parentUsername: 'courier_level3_school',
    description: '负责东区宿舍楼群信件收发管理',
    permissions: [
      'courier_scan_code', 'courier_deliver_letter', 'courier_view_own_tasks',
      'courier_report_exception', 'courier_manage_subordinates', 'courier_assign_tasks',
      'courier_view_subordinate_reports', 'courier_create_subordinate'
    ],
    password: process.env.TEST_ACCOUNT_COURIER_LEVEL2_ZONE_PASSWORD || 'zone123'
  },
  
  {
    username: 'courier_level1_building',
    email: 'building.courier@openpenpal.com',
    level: 1,
    levelName: '一级信使（楼栋/班级）',
    zoneCode: 'BJDX_EAST_A1',
    zoneType: 'building',
    parentUsername: 'courier_level2_zone',
    description: '负责A1楼信件投递和收取',
    permissions: [
      'courier_scan_code', 'courier_deliver_letter', 'courier_view_own_tasks',
      'courier_report_exception'
    ],
    password: process.env.TEST_ACCOUNT_COURIER_LEVEL1_BUILDING_PASSWORD || 'building123'
  }
]

// 创建用户表结构
async function createUsersTable() {
  const createTableSQL = `
    CREATE TABLE IF NOT EXISTS test_users (
      id VARCHAR(50) PRIMARY KEY,
      username VARCHAR(100) UNIQUE NOT NULL,
      email VARCHAR(255) UNIQUE NOT NULL,
      real_name VARCHAR(100) NOT NULL,
      password_hash TEXT NOT NULL,
      role VARCHAR(50) NOT NULL,
      permissions JSONB DEFAULT '[]'::jsonb,
      school_code VARCHAR(20),
      school_name VARCHAR(100),
      status VARCHAR(20) DEFAULT 'active',
      courier_level INTEGER,
      courier_info JSONB,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    -- 创建索引
    CREATE INDEX IF NOT EXISTS idx_test_users_username ON test_users(username);
    CREATE INDEX IF NOT EXISTS idx_test_users_email ON test_users(email);
    CREATE INDEX IF NOT EXISTS idx_test_users_role ON test_users(role);
    CREATE INDEX IF NOT EXISTS idx_test_users_school_code ON test_users(school_code);
    CREATE INDEX IF NOT EXISTS idx_test_users_status ON test_users(status);
  `
  
  await pool.query(createTableSQL)
  console.log('✅ 用户表创建成功')
}

// 插入基础用户
async function insertBaseUsers() {
  console.log('📝 开始插入基础用户...')
  
  for (const user of TEST_USERS) {
    try {
      // 检查用户是否已存在
      const existingUser = await pool.query(
        'SELECT id FROM test_users WHERE username = $1',
        [user.username]
      )
      
      if (existingUser.rows.length > 0) {
        console.log(`⚠️  用户 ${user.username} 已存在，跳过`)
        continue
      }
      
      // 哈希密码
      const passwordHash = await bcrypt.hash(user.password, 12)
      
      // 插入用户
      await pool.query(`
        INSERT INTO test_users (
          id, username, email, real_name, password_hash, role, 
          permissions, school_code, school_name, status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP)
      `, [
        user.id,
        user.username,
        user.email,
        user.realName,
        passwordHash,
        user.role,
        JSON.stringify(user.permissions),
        user.school_code,
        user.school_name,
        user.status
      ])
      
      console.log(`✅ 插入基础用户: ${user.username} (${user.realName})`)
    } catch (error) {
      console.error(`❌ 插入用户 ${user.username} 失败:`, error.message)
    }
  }
}

// 插入层级信使用户
async function insertCourierUsers() {
  console.log('📝 开始插入层级信使用户...')
  
  for (const courier of COURIER_LEVELS) {
    try {
      // 检查用户是否已存在
      const existingUser = await pool.query(
        'SELECT id FROM test_users WHERE username = $1',
        [courier.username]
      )
      
      if (existingUser.rows.length > 0) {
        console.log(`⚠️  信使 ${courier.username} 已存在，跳过`)
        continue
      }
      
      // 生成用户ID
      const userId = generateStandardUserId(courier.username, courier)
      
      // 哈希密码
      const passwordHash = await bcrypt.hash(courier.password, 12)
      
      // 构建信使信息
      const courierInfo = {
        level: courier.level,
        zoneCode: courier.zoneCode,
        zoneType: courier.zoneType,
        status: 'active',
        points: Math.floor(Math.random() * 1000) + 500,
        taskCount: Math.floor(Math.random() * 50) + 20
      }
      
      // 插入信使用户
      await pool.query(`
        INSERT INTO test_users (
          id, username, email, real_name, password_hash, role,
          permissions, school_code, school_name, status, 
          courier_level, courier_info, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP)
      `, [
        userId,
        courier.username,
        courier.email,
        courier.levelName,
        passwordHash,
        'courier',
        JSON.stringify(courier.permissions),
        courier.zoneCode.includes('BJDX') ? 'BJDX01' : 'SYSTEM',
        courier.zoneCode.includes('BJDX') ? '北京大学' : '系统测试',
        'active',
        courier.level,
        JSON.stringify(courierInfo)
      ])
      
      console.log(`✅ 插入层级信使: ${courier.username} (${courier.levelName})`)
    } catch (error) {
      console.error(`❌ 插入信使 ${courier.username} 失败:`, error.message)
    }
  }
}

// 验证数据插入
async function verifyData() {
  console.log('🔍 验证数据插入...')
  
  const result = await pool.query('SELECT COUNT(*) as count FROM test_users')
  const count = parseInt(result.rows[0].count)
  
  console.log(`📊 数据库中共有 ${count} 个测试用户`)
  
  // 显示用户列表
  const users = await pool.query(`
    SELECT username, real_name, role, courier_level, status 
    FROM test_users 
    ORDER BY courier_level DESC NULLS LAST, username
  `)
  
  console.log('\n📋 测试用户列表:')
  users.rows.forEach(user => {
    const levelInfo = user.courier_level ? `Level ${user.courier_level}` : 'N/A'
    console.log(`  - ${user.username.padEnd(25)} | ${user.real_name.padEnd(20)} | ${user.role.padEnd(15)} | ${levelInfo}`)
  })
}

// 主函数
async function main() {
  try {
    console.log('🚀 开始初始化数据库测试用户...\n')
    
    // 测试数据库连接
    await pool.query('SELECT 1')
    console.log('✅ 数据库连接成功')
    
    // 创建表结构
    await createUsersTable()
    
    // 插入测试数据
    await insertBaseUsers()
    await insertCourierUsers()
    
    // 验证数据
    await verifyData()
    
    console.log('\n🎉 测试用户初始化完成！')
    console.log('\n💡 现在可以使用以下密码登录测试账户:')
    console.log('  - admin: admin123')
    console.log('  - courier_level2_zone: zone123')
    console.log('  - courier_level3_school: school123')
    console.log('  - courier_level4_city: city123')
    console.log('  - 其他密码请查看 .env.local 文件')
    
  } catch (error) {
    console.error('❌ 初始化失败:', error)
    process.exit(1)
  } finally {
    await pool.end()
  }
}

// 执行脚本
if (require.main === module) {
  main()
}

module.exports = { main }