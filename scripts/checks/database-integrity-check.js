// 数据库完整性和API交互全面检查 - SOTA级别验证
const { Client } = require('pg');

async function comprehensiveDatabaseCheck() {
  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://rocalight:@localhost:5432/openpenpal'
  });
  
  try {
    await client.connect();
    console.log('🔗 数据库连接成功');
    
    // 1. 检查所有晋升系统相关表的结构
    console.log('\n📊 1. 表结构完整性检查');
    console.log('='.repeat(60));
    
    const promotionTables = [
      'courier_upgrade_requests',
      'courier_promotion_history', 
      'courier_level_requirements',
      'couriers',
      'users'
    ];
    
    for (const tableName of promotionTables) {
      console.log(`\n📋 检查表: ${tableName}`);
      
      // 检查表是否存在
      const tableExists = await client.query(`
        SELECT EXISTS (
          SELECT FROM information_schema.tables 
          WHERE table_schema = 'public' 
          AND table_name = $1
        );
      `, [tableName]);
      
      if (!tableExists.rows[0].exists) {
        console.log(`❌ 表 ${tableName} 不存在`);
        continue;
      }
      
      // 检查表结构
      const columns = await client.query(`
        SELECT 
          column_name, 
          data_type, 
          is_nullable,
          column_default,
          character_maximum_length
        FROM information_schema.columns 
        WHERE table_name = $1
        ORDER BY ordinal_position;
      `, [tableName]);
      
      console.log(`✅ 表存在，共 ${columns.rows.length} 个字段:`);
      columns.rows.forEach(col => {
        const nullable = col.is_nullable === 'YES' ? 'NULL' : 'NOT NULL';
        const length = col.character_maximum_length ? `(${col.character_maximum_length})` : '';
        const defaultVal = col.column_default ? ` DEFAULT ${col.column_default}` : '';
        console.log(`   - ${col.column_name}: ${col.data_type}${length} ${nullable}${defaultVal}`);
      });
    }
    
    // 2. 检查外键约束
    console.log('\n🔗 2. 外键约束检查');
    console.log('='.repeat(60));
    
    const foreignKeys = await client.query(`
      SELECT 
        tc.constraint_name,
        tc.table_name,
        kcu.column_name,
        ccu.table_name AS foreign_table_name,
        ccu.column_name AS foreign_column_name
      FROM information_schema.table_constraints AS tc 
      JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
        AND tc.table_schema = kcu.table_schema
      JOIN information_schema.constraint_column_usage AS ccu
        ON ccu.constraint_name = tc.constraint_name
        AND ccu.table_schema = tc.table_schema
      WHERE tc.constraint_type = 'FOREIGN KEY' 
        AND tc.table_name IN ('courier_upgrade_requests', 'courier_promotion_history')
      ORDER BY tc.table_name, tc.constraint_name;
    `);
    
    console.log(`✅ 找到 ${foreignKeys.rows.length} 个外键约束:`);
    foreignKeys.rows.forEach(fk => {
      console.log(`   - ${fk.table_name}.${fk.column_name} -> ${fk.foreign_table_name}.${fk.foreign_column_name}`);
    });
    
    // 3. 检查数据完整性
    console.log('\n📊 3. 数据完整性检查');
    console.log('='.repeat(60));
    
    // 检查用户-信使关系
    const userCourierIntegrity = await client.query(`
      SELECT 
        u.id as user_id,
        u.username,
        u.role,
        c.id as courier_id,
        c.level,
        c.zone
      FROM users u
      LEFT JOIN couriers c ON u.id = c.user_id
      WHERE u.role LIKE '%courier%'
      ORDER BY c.level, u.username;
    `);
    
    console.log(`✅ 用户-信使关系检查 (${userCourierIntegrity.rows.length} 条记录):`);
    userCourierIntegrity.rows.forEach(row => {
      const status = row.courier_id ? '✅' : '❌';
      console.log(`   ${status} ${row.username} (${row.role}) -> Level ${row.level || 'N/A'} - ${row.zone || 'N/A'}`);
    });
    
    // 4. 检查晋升申请数据
    console.log('\n📋 4. 晋升申请数据检查');
    console.log('='.repeat(60));
    
    const upgradeRequests = await client.query(`
      SELECT 
        ur.id,
        ur.courier_id,
        u.username,
        ur.current_level,
        ur.request_level,
        ur.status,
        ur.created_at,
        ur.expires_at
      FROM courier_upgrade_requests ur
      JOIN users u ON ur.courier_id = u.id
      ORDER BY ur.created_at DESC;
    `);
    
    console.log(`✅ 晋升申请记录 (${upgradeRequests.rows.length} 条):`);
    upgradeRequests.rows.forEach(req => {
      const isExpired = new Date(req.expires_at) < new Date();
      const expiredStatus = isExpired ? '⏰ 已过期' : '✅ 有效';
      console.log(`   - ${req.username}: ${req.current_level} -> ${req.request_level} (${req.status}) ${expiredStatus}`);
    });
    
    // 5. 检查晋升历史数据
    console.log('\n📈 5. 晋升历史数据检查');
    console.log('='.repeat(60));
    
    const promotionHistory = await client.query(`
      SELECT 
        ph.id,
        ph.courier_id,
        u.username,
        ph.from_level,
        ph.to_level,
        ph.promoted_by,
        ph.promoted_at
      FROM courier_promotion_history ph
      JOIN users u ON ph.courier_id = u.id
      ORDER BY ph.promoted_at DESC;
    `);
    
    console.log(`✅ 晋升历史记录 (${promotionHistory.rows.length} 条):`);
    promotionHistory.rows.forEach(hist => {
      console.log(`   - ${hist.username}: ${hist.from_level} -> ${hist.to_level} (${hist.promoted_at?.toISOString()?.split('T')[0] || 'N/A'})`);
    });
    
    // 6. 检查等级要求配置
    console.log('\n⚙️ 6. 等级要求配置检查');
    console.log('='.repeat(60));
    
    const levelRequirements = await client.query(`
      SELECT 
        from_level,
        to_level,
        requirement_type,
        requirement_value,
        is_mandatory,
        description
      FROM courier_level_requirements
      ORDER BY from_level, to_level, requirement_type;
    `);
    
    console.log(`✅ 等级要求配置 (${levelRequirements.rows.length} 条):`);
    const groupedReqs = {};
    levelRequirements.rows.forEach(req => {
      const key = `${req.from_level}->${req.to_level}`;
      if (!groupedReqs[key]) groupedReqs[key] = [];
      groupedReqs[key].push(req);
    });
    
    Object.entries(groupedReqs).forEach(([key, reqs]) => {
      console.log(`   📊 ${key}:`);
      reqs.forEach(req => {
        const mandatory = req.is_mandatory ? '必需' : '可选';
        console.log(`      - ${req.requirement_type} (${mandatory}): ${req.description}`);
      });
    });
    
    // 7. 数据一致性检查
    console.log('\n🔍 7. 数据一致性检查');
    console.log('='.repeat(60));
    
    // 检查孤儿申请（申请者不存在）
    const orphanRequests = await client.query(`
      SELECT ur.id, ur.courier_id 
      FROM courier_upgrade_requests ur
      LEFT JOIN users u ON ur.courier_id = u.id
      WHERE u.id IS NULL;
    `);
    
    if (orphanRequests.rows.length > 0) {
      console.log(`❌ 发现 ${orphanRequests.rows.length} 个孤儿申请记录`);
      orphanRequests.rows.forEach(req => {
        console.log(`   - 申请ID: ${req.id}, 无效用户ID: ${req.courier_id}`);
      });
    } else {
      console.log(`✅ 所有申请记录的用户引用都有效`);
    }
    
    // 检查申请等级的合理性
    const invalidLevelRequests = await client.query(`
      SELECT 
        ur.id,
        u.username,
        c.level as current_actual_level,
        ur.current_level as request_current_level,
        ur.request_level
      FROM courier_upgrade_requests ur
      JOIN users u ON ur.courier_id = u.id
      JOIN couriers c ON ur.courier_id = c.user_id
      WHERE c.level != ur.current_level 
         OR ur.request_level != c.level + 1;
    `);
    
    if (invalidLevelRequests.rows.length > 0) {
      console.log(`⚠️ 发现 ${invalidLevelRequests.rows.length} 个等级不一致的申请:`);
      invalidLevelRequests.rows.forEach(req => {
        console.log(`   - ${req.username}: 实际等级${req.current_actual_level}, 申请中当前等级${req.request_current_level}, 申请等级${req.request_level}`);
      });
    } else {
      console.log(`✅ 所有申请的等级信息一致`);
    }
    
    // 8. 统计信息汇总
    console.log('\n📊 8. 系统统计汇总');
    console.log('='.repeat(60));
    
    const stats = await client.query(`
      SELECT 
        'users' as table_name,
        COUNT(*) as count
      FROM users
      WHERE role LIKE '%courier%'
      
      UNION ALL
      
      SELECT 
        'couriers' as table_name,
        COUNT(*) as count
      FROM couriers
      
      UNION ALL
      
      SELECT 
        'upgrade_requests' as table_name,
        COUNT(*) as count
      FROM courier_upgrade_requests
      
      UNION ALL
      
      SELECT 
        'promotion_history' as table_name,
        COUNT(*) as count
      FROM courier_promotion_history
      
      UNION ALL
      
      SELECT 
        'level_requirements' as table_name,
        COUNT(*) as count
      FROM courier_level_requirements;
    `);
    
    console.log('📈 表记录统计:');
    stats.rows.forEach(stat => {
      console.log(`   - ${stat.table_name}: ${stat.count} 条记录`);
    });
    
    // 按状态分组的申请统计
    const requestStatusStats = await client.query(`
      SELECT status, COUNT(*) as count
      FROM courier_upgrade_requests
      GROUP BY status
      ORDER BY count DESC;
    `);
    
    console.log('\n📊 申请状态分布:');
    requestStatusStats.rows.forEach(stat => {
      console.log(`   - ${stat.status}: ${stat.count} 个申请`);
    });
    
    console.log('\n🎉 数据库完整性检查完成！');
    
  } catch (error) {
    console.error('❌ 数据库检查失败:', error.message);
  } finally {
    await client.end();
  }
}

comprehensiveDatabaseCheck();