const { Client } = require('pg');

async function checkCourierDatabaseConsistency() {
  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://rocalight:postgres@localhost:5432/openpenpal'
  });

  try {
    await client.connect();
    console.log('Connected to database');

    // 1. Check users with courier roles
    console.log('\n=== COURIER USERS ===');
    const courierUsers = await client.query(`
      SELECT id, username, role, nickname, school_code 
      FROM users 
      WHERE role LIKE 'courier%' 
      ORDER BY role
    `);
    console.log(`Found ${courierUsers.rows.length} courier users:`);
    courierUsers.rows.forEach(user => {
      console.log(`- ${user.username} (${user.role}) - ${user.nickname}`);
    });

    // 2. Check courier records
    console.log('\n=== COURIER RECORDS ===');
    const couriers = await client.query(`
      SELECT c.*, u.username 
      FROM couriers c
      JOIN users u ON c.user_id = u.id
      ORDER BY c.level DESC
    `);
    console.log(`Found ${couriers.rows.length} courier records:`);
    couriers.rows.forEach(courier => {
      console.log(`- ${courier.username} (L${courier.level}) | Zone: ${courier.zone_code} | Parent: ${courier.parent_id || 'None'} | Prefix: ${courier.managed_op_code_prefix || 'None'}`);
    });

    // 3. Check hierarchy
    console.log('\n=== COURIER HIERARCHY ===');
    const hierarchy = await client.query(`
      SELECT 
        c.id,
        u.username as courier_name,
        c.level,
        p.username as parent_name,
        c.zone_code
      FROM couriers c
      JOIN users u ON c.user_id = u.id
      LEFT JOIN couriers pc ON c.parent_id = pc.id
      LEFT JOIN users p ON pc.user_id = p.id
      ORDER BY c.level DESC
    `);
    hierarchy.rows.forEach(row => {
      console.log(`- ${row.courier_name} (L${row.level}) → ${row.parent_name || 'No Parent'}`);
    });

    // 4. Check shared tasks
    console.log('\n=== SHARED TASKS ===');
    const tasks = await client.query(`
      SELECT 
        ct.*,
        l.title as letter_title,
        c.username as assigned_to
      FROM courier_tasks ct
      LEFT JOIN letters l ON ct.letter_id = l.id
      LEFT JOIN couriers co ON ct.courier_id = co.id
      LEFT JOIN users c ON co.user_id = c.id
      ORDER BY ct.created_at DESC
    `);
    console.log(`Found ${tasks.rows.length} tasks:`);
    tasks.rows.forEach(task => {
      console.log(`- Task ${task.id.slice(0, 8)}... | Status: ${task.status} | Type: ${task.task_type} | Assigned to: ${task.assigned_to || 'UNASSIGNED'} | ${task.pickup_op_code} → ${task.delivery_op_code}`);
    });

    // 5. Check task distribution
    console.log('\n=== TASK DISTRIBUTION ===');
    const taskStats = await client.query(`
      SELECT 
        status,
        COUNT(*) as count,
        COUNT(DISTINCT courier_id) as courier_count
      FROM courier_tasks
      GROUP BY status
    `);
    taskStats.rows.forEach(stat => {
      console.log(`- ${stat.status}: ${stat.count} tasks (${stat.courier_count} couriers)`);
    });

    // 6. Check if couriers can see shared tasks
    console.log('\n=== TASK VISIBILITY CHECK ===');
    for (const courier of couriers.rows) {
      const visibleTasks = await client.query(`
        SELECT COUNT(*) as total,
               COUNT(CASE WHEN courier_id IS NULL THEN 1 END) as available,
               COUNT(CASE WHEN courier_id = $1 THEN 1 END) as assigned_to_me
        FROM courier_tasks
        WHERE status IN ('available', 'accepted', 'in_transit')
      `, [courier.id]);
      
      const stats = visibleTasks.rows[0];
      console.log(`- ${courier.username} can see: ${stats.total} tasks (${stats.available} available, ${stats.assigned_to_me} assigned to them)`);
    }

    // 7. Check OP Code assignments
    console.log('\n=== OP CODE ASSIGNMENTS ===');
    const opCodes = await client.query(`
      SELECT DISTINCT pickup_op_code, delivery_op_code, COUNT(*) as task_count
      FROM courier_tasks
      GROUP BY pickup_op_code, delivery_op_code
    `);
    opCodes.rows.forEach(code => {
      console.log(`- Route: ${code.pickup_op_code} → ${code.delivery_op_code} (${code.task_count} tasks)`);
    });

  } catch (error) {
    console.error('Error:', error.message);
  } finally {
    await client.end();
  }
}

// Run the check
checkCourierDatabaseConsistency();