-- Fix Courier Hierarchy and Create Shared Data
-- This script ensures all courier levels have proper data and relationships

-- First, let's check existing users with courier roles
SELECT 'Existing courier users:' as info;
SELECT id, username, role FROM users WHERE role LIKE 'courier%' ORDER BY role;

-- Create courier records for all courier users if they don't exist
INSERT INTO couriers (id, user_id, level, zone_code, zone_type, status, created_at, updated_at)
SELECT 
    gen_random_uuid()::text,
    u.id,
    CASE 
        WHEN u.role = 'courier_level4' THEN 4
        WHEN u.role = 'courier_level3' THEN 3
        WHEN u.role = 'courier_level2' THEN 2
        WHEN u.role = 'courier_level1' THEN 1
        ELSE 1
    END as level,
    CASE 
        WHEN u.role = 'courier_level4' THEN 'BEIJING'
        WHEN u.role = 'courier_level3' THEN 'BJDX'
        WHEN u.role = 'courier_level2' THEN 'BJDX-NORTH'
        WHEN u.role = 'courier_level1' THEN 'BJDX-A-101'
    END as zone_code,
    CASE 
        WHEN u.role = 'courier_level4' THEN 'city'
        WHEN u.role = 'courier_level3' THEN 'school'
        WHEN u.role = 'courier_level2' THEN 'zone'
        WHEN u.role = 'courier_level1' THEN 'building'
    END as zone_type,
    'active',
    NOW(),
    NOW()
FROM users u
WHERE u.role LIKE 'courier%'
AND NOT EXISTS (
    SELECT 1 FROM couriers c WHERE c.user_id = u.id
);

-- Update hierarchy relationships
-- First, ensure we have the courier IDs
DO $$
DECLARE
    l4_courier_id TEXT;
    l3_courier_id TEXT;
    l2_courier_id TEXT;
BEGIN
    -- Get Level 4 courier ID
    SELECT c.id INTO l4_courier_id
    FROM couriers c
    JOIN users u ON c.user_id = u.id
    WHERE u.username = 'courier_level4'
    LIMIT 1;

    -- Get Level 3 courier ID
    SELECT c.id INTO l3_courier_id
    FROM couriers c
    JOIN users u ON c.user_id = u.id
    WHERE u.username = 'courier_level3'
    LIMIT 1;

    -- Get Level 2 courier ID
    SELECT c.id INTO l2_courier_id
    FROM couriers c
    JOIN users u ON c.user_id = u.id
    WHERE u.username = 'courier_level2'
    LIMIT 1;

    -- Update parent relationships
    -- Level 3 reports to Level 4
    UPDATE couriers SET parent_id = l4_courier_id
    WHERE id = l3_courier_id;

    -- Level 2 reports to Level 3
    UPDATE couriers SET parent_id = l3_courier_id
    WHERE id = l2_courier_id;

    -- Level 1 reports to Level 2
    UPDATE couriers SET parent_id = l2_courier_id
    WHERE id IN (
        SELECT c.id FROM couriers c
        JOIN users u ON c.user_id = u.id
        WHERE u.username = 'courier_level1'
    );

    -- Update managed OP Code prefixes
    UPDATE couriers SET managed_op_code_prefix = 'BJ' WHERE level = 4;
    UPDATE couriers SET managed_op_code_prefix = 'BJDX' WHERE level = 3 AND zone_code = 'BJDX';
    UPDATE couriers SET managed_op_code_prefix = 'BJDX5F' WHERE level = 2 AND zone_code = 'BJDX-NORTH';
    UPDATE couriers SET managed_op_code_prefix = 'BJDX5F01' WHERE level = 1 AND zone_code LIKE 'BJDX-A-%';
END $$;

-- Create shared letters that need delivery
INSERT INTO letters (id, user_id, title, content, recipient_type, is_anonymous, status, created_at, updated_at)
VALUES
    (gen_random_uuid()::text, (SELECT id FROM users WHERE username = 'alice' LIMIT 1), 
     '给远方朋友的信', '希望这封信能带去我的思念...', 'specific', false, 'pending', NOW(), NOW()),
    (gen_random_uuid()::text, (SELECT id FROM users WHERE username = 'alice' LIMIT 1), 
     '春节祝福', '新年快乐，万事如意！', 'random', false, 'pending', NOW(), NOW()),
    (gen_random_uuid()::text, (SELECT id FROM users WHERE username = 'bob' LIMIT 1), 
     '感谢信', '感谢你一直以来的陪伴...', 'specific', false, 'pending', NOW(), NOW()),
    (gen_random_uuid()::text, (SELECT id FROM users WHERE username = 'charlie' LIMIT 1), 
     '生日祝福', '祝你生日快乐！', 'specific', false, 'pending', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Create shared courier tasks for these letters
-- These tasks are NOT assigned to specific couriers initially
INSERT INTO courier_tasks (
    id, letter_id, pickup_location, delivery_location, 
    pickup_op_code, delivery_op_code, task_type,
    status, priority, created_at, updated_at
)
SELECT 
    gen_random_uuid()::text,
    l.id,
    'BJDX5F01',  -- 北大5号楼1层 (pickup)
    CASE (random() * 3)::int
        WHEN 0 THEN 'BJDX3D12'  -- 北大3食堂12号桌
        WHEN 1 THEN 'QHUA2G03'  -- 清华2号门3区
        WHEN 2 THEN 'BJDX7B05'  -- 北大7号楼B区5号
        ELSE 'BJDX5F03'  -- 北大5号楼3层
    END,
    'BJDX5F01',  -- pickup OP code
    CASE (random() * 3)::int
        WHEN 0 THEN 'BJDX3D12'
        WHEN 1 THEN 'QHUA2G03'
        WHEN 2 THEN 'BJDX7B05'
        ELSE 'BJDX5F03'
    END,
    'standard',
    'available',  -- Available for any courier to claim
    CASE WHEN random() < 0.3 THEN 'high' ELSE 'medium' END,
    NOW(),
    NOW()
FROM letters l
WHERE l.status = 'pending'
AND NOT EXISTS (
    SELECT 1 FROM courier_tasks ct WHERE ct.letter_id = l.id
);

-- Create some tasks already assigned to show hierarchy
-- Assign some tasks to Level 1 couriers
UPDATE courier_tasks 
SET 
    courier_id = (
        SELECT c.id FROM couriers c 
        JOIN users u ON c.user_id = u.id 
        WHERE u.username = 'courier_level1' 
        LIMIT 1
    ),
    status = 'accepted',
    accepted_at = NOW()
WHERE id IN (
    SELECT id FROM courier_tasks 
    WHERE status = 'available' 
    LIMIT 2
);

-- Create inter-school delivery tasks (requiring higher level couriers)
INSERT INTO courier_tasks (
    id, letter_id, pickup_location, delivery_location,
    pickup_op_code, delivery_op_code, task_type,
    status, priority, required_level, created_at, updated_at
)
SELECT 
    gen_random_uuid()::text,
    l.id,
    'BJDX5F01',  -- 北大
    'QHUA3B02',  -- 清华
    'BJDX5F01',
    'QHUA3B02',
    'inter_school',
    'available',
    'high',
    3,  -- Requires Level 3 or higher
    NOW(),
    NOW()
FROM letters l
WHERE l.id IN (
    SELECT id FROM letters 
    WHERE status = 'pending' 
    ORDER BY created_at DESC 
    LIMIT 1
);

-- Update courier statistics
UPDATE couriers c
SET 
    completed_tasks = (
        SELECT COUNT(*) 
        FROM courier_tasks ct 
        WHERE ct.courier_id = c.id 
        AND ct.status = 'delivered'
    ),
    performance_score = 95.0 + (random() * 5),  -- 95-100 score
    updated_at = NOW();

-- Create courier activities log
INSERT INTO courier_activities (
    id, courier_id, task_id, action, location, notes, created_at
)
SELECT 
    gen_random_uuid()::text,
    ct.courier_id,
    ct.id,
    'task_accepted',
    ct.pickup_location,
    '接受配送任务',
    ct.accepted_at
FROM courier_tasks ct
WHERE ct.courier_id IS NOT NULL
AND ct.accepted_at IS NOT NULL;

-- Show final hierarchy
SELECT 'Final courier hierarchy:' as info;
SELECT 
    c.id,
    u.username,
    c.level,
    c.zone_code,
    c.zone_type,
    c.parent_id,
    p.username as parent_username,
    c.managed_op_code_prefix
FROM couriers c
JOIN users u ON c.user_id = u.id
LEFT JOIN couriers pc ON c.parent_id = pc.id
LEFT JOIN users p ON pc.user_id = p.id
ORDER BY c.level DESC, u.username;

-- Show task distribution
SELECT 'Task distribution:' as info;
SELECT 
    ct.status,
    ct.task_type,
    ct.required_level,
    COUNT(*) as count
FROM courier_tasks ct
GROUP BY ct.status, ct.task_type, ct.required_level
ORDER BY ct.status, ct.task_type;