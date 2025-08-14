-- OpenPenPal 晋升系统数据库测试脚本
-- 用于验证晋升系统的数据库功能

-- 1. 查看所有信使用户及其等级
SELECT 
    u.id,
    u.username,
    u.nickname,
    u.role,
    c.level as courier_level,
    c.zone,
    c.status as courier_status,
    c.points,
    c.task_count,
    c.created_at as courier_since
FROM users u
LEFT JOIN couriers c ON u.id = c.user_id
WHERE u.role LIKE '%courier%' OR u.role = 'super_admin'
ORDER BY 
    CASE 
        WHEN c.level IS NULL THEN 999 
        ELSE c.level 
    END DESC,
    u.created_at;

-- 2. 查看信使层级关系树
WITH RECURSIVE courier_hierarchy AS (
    -- 顶级信使（没有上级）
    SELECT 
        c.id,
        c.user_id,
        c.level,
        u.nickname,
        c.parent_id,
        CAST(u.nickname AS TEXT) as hierarchy_path,
        0 as depth
    FROM couriers c
    JOIN users u ON c.user_id = u.id
    WHERE c.parent_id IS NULL
    
    UNION ALL
    
    -- 递归查找下级
    SELECT 
        c.id,
        c.user_id,
        c.level,
        u.nickname,
        c.parent_id,
        ch.hierarchy_path || ' -> ' || u.nickname,
        ch.depth + 1
    FROM couriers c
    JOIN users u ON c.user_id = u.id
    JOIN courier_hierarchy ch ON c.parent_id = ch.id
)
SELECT 
    REPEAT('  ', depth) || nickname as courier_tree,
    level,
    hierarchy_path
FROM courier_hierarchy
ORDER BY hierarchy_path;

-- 3. 查看晋升申请记录（如果表存在）
-- 注意：这个表可能还未创建，如果报错请忽略
SELECT 
    cur.id as request_id,
    u.username,
    u.nickname,
    cur.current_level,
    cur.request_level,
    cur.reason,
    cur.status,
    cur.created_at as requested_at,
    cur.reviewed_at,
    reviewer.nickname as reviewed_by
FROM courier_upgrade_requests cur
JOIN users u ON cur.courier_id = u.id
LEFT JOIN users reviewer ON cur.reviewer_id = reviewer.id
ORDER BY cur.created_at DESC;

-- 4. 查看每个等级的信使数量统计
SELECT 
    c.level,
    COUNT(*) as courier_count,
    AVG(c.points) as avg_points,
    SUM(c.task_count) as total_tasks
FROM couriers c
GROUP BY c.level
ORDER BY c.level;

-- 5. 查看可以管理下级的信使（3级及以上）
SELECT 
    u.username,
    u.nickname,
    c.level,
    c.zone,
    CASE 
        WHEN c.level = 4 THEN '可管理: 1-3级信使'
        WHEN c.level = 3 THEN '可管理: 1-2级信使'
        WHEN c.level = 2 THEN '可管理: 1级信使'
        ELSE '无管理权限'
    END as management_permission
FROM users u
JOIN couriers c ON u.id = c.user_id
WHERE c.level >= 2
ORDER BY c.level DESC;

-- 6. 创建测试晋升申请记录（如果需要）
-- 取消注释以下内容来创建测试数据
/*
INSERT INTO courier_upgrade_requests (
    id,
    courier_id,
    current_level,
    request_level,
    reason,
    status,
    created_at
) VALUES 
    (
        'test-upgrade-001',
        (SELECT user_id FROM couriers WHERE level = 1 LIMIT 1),
        1,
        2,
        '完成了10个投递任务，申请晋升到二级信使',
        'pending',
        CURRENT_TIMESTAMP
    ),
    (
        'test-upgrade-002',
        (SELECT user_id FROM couriers WHERE level = 2 LIMIT 1),
        2,
        3,
        '管理能力突出，申请晋升到三级信使',
        'pending',
        CURRENT_TIMESTAMP
    );
*/

-- 7. 检查晋升权限逻辑
-- 显示谁可以审批谁的晋升申请
SELECT 
    'Level ' || reviewer.level || ' (' || reviewer_user.nickname || ')' as reviewer,
    'Can approve Level ' || (reviewer.level - 1) || ' and below' as approval_permission
FROM couriers reviewer
JOIN users reviewer_user ON reviewer.user_id = reviewer_user.id
WHERE reviewer.level >= 2
ORDER BY reviewer.level DESC;