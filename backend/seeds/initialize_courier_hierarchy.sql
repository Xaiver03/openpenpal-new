-- Initialize Courier Hierarchy with Shared Data
-- This script ensures all courier levels have proper data and relationships

-- First, ensure all courier users have courier records
INSERT INTO couriers (id, user_id, name, contact, school, zone, status, level, created_at, updated_at)
SELECT 
    gen_random_uuid()::text,
    u.id,
    u.nickname,
    u.email,
    u.school_code,
    CASE 
        WHEN u.role = 'courier_level4' THEN 'BEIJING'
        WHEN u.role = 'courier_level3' THEN 'PKU'
        WHEN u.role = 'courier_level2' THEN 'PKU-NORTH'
        WHEN u.role = 'courier_level1' THEN 'PKU-A-101'
    END as zone,
    'approved',
    CASE 
        WHEN u.role = 'courier_level4' THEN 4
        WHEN u.role = 'courier_level3' THEN 3
        WHEN u.role = 'courier_level2' THEN 2
        WHEN u.role = 'courier_level1' THEN 1
        ELSE 1
    END as level,
    NOW(),
    NOW()
FROM users u
WHERE u.role LIKE 'courier%'
AND NOT EXISTS (
    SELECT 1 FROM couriers c WHERE c.user_id = u.id
)
ON CONFLICT (user_id) DO NOTHING;

-- Update managed OP Code prefixes
UPDATE couriers c
SET managed_op_code_prefix = CASE 
    WHEN c.level = 4 THEN 'BJ'
    WHEN c.level = 3 THEN 'PK'
    WHEN c.level = 2 THEN 'PK5F'
    WHEN c.level = 1 THEN 'PK5F01'
END
FROM users u
WHERE c.user_id = u.id
AND u.username IN ('courier_level1', 'courier_level2', 'courier_level3', 'courier_level4');

-- Create sample letters
INSERT INTO letters (id, user_id, title, content, style, status, is_anonymous, read_count, created_at, updated_at)
SELECT 
    gen_random_uuid()::text,
    (SELECT id FROM users WHERE username = 'alice' LIMIT 1),
    title,
    content,
    'casual',
    'generated',
    false,
    0,
    NOW(),
    NOW()
FROM (VALUES 
    ('给远方朋友的新年祝福', '新的一年，希望你一切安好，愿我们的友谊长存...'),
    ('感谢信', '感谢你在我困难时期的帮助和支持...'),
    ('校园回忆', '还记得我们一起在梧桐树下读书的日子吗...'),
    ('生日祝福', '祝你生日快乐，愿你的每一天都充满阳光...')
) AS letters(title, content)
WHERE NOT EXISTS (
    SELECT 1 FROM letters WHERE title = letters.title
);

-- Create letter codes for the letters
INSERT INTO letter_codes (id, letter_id, code, created_at, updated_at)
SELECT 
    gen_random_uuid()::text,
    l.id,
    'LC' || LPAD(FLOOR(RANDOM() * 999999 + 1)::text, 6, '0'),
    NOW(),
    NOW()
FROM letters l
WHERE NOT EXISTS (
    SELECT 1 FROM letter_codes lc WHERE lc.letter_id = l.id
)
AND l.created_at >= NOW() - INTERVAL '5 minutes';

-- Create shared courier tasks (unassigned initially)
INSERT INTO courier_tasks (
    id, courier_id, letter_code, title, sender_name, 
    target_location, pickup_op_code, delivery_op_code, 
    status, priority, created_at, updated_at
)
SELECT 
    gen_random_uuid()::text,
    c.id, -- Temporarily assign to show in database, but we'll make some available
    lc.code,
    l.title,
    u.nickname,
    CASE (RANDOM() * 3)::int
        WHEN 0 THEN '北京大学3食堂'
        WHEN 1 THEN '北京大学图书馆'
        WHEN 2 THEN '北京大学7号楼'
        ELSE '北京大学5号楼'
    END,
    'PK5F01',  -- Pickup location
    CASE (RANDOM() * 3)::int
        WHEN 0 THEN 'PK3D12'
        WHEN 1 THEN 'PKTSG1'
        WHEN 2 THEN 'PK7B05'
        ELSE 'PK5F03'
    END,
    'pending',  -- Available for pickup
    CASE WHEN RANDOM() < 0.3 THEN 'urgent' ELSE 'normal' END,
    NOW(),
    NOW()
FROM letters l
JOIN letter_codes lc ON l.id = lc.letter_id
JOIN users u ON l.user_id = u.id
CROSS JOIN (
    SELECT id FROM couriers WHERE level = 1 LIMIT 1
) c
WHERE l.created_at >= NOW() - INTERVAL '5 minutes'
AND NOT EXISTS (
    SELECT 1 FROM courier_tasks ct WHERE ct.letter_code = lc.code
);

-- Make some tasks available by setting courier_id to a placeholder
-- This simulates unassigned tasks that any courier can claim
UPDATE courier_tasks 
SET courier_id = (SELECT id FROM couriers WHERE level = 1 LIMIT 1)
WHERE status = 'pending'
AND created_at >= NOW() - INTERVAL '5 minutes'
AND random() < 0.5;

-- Create some inter-school tasks that require higher level couriers
INSERT INTO courier_tasks (
    id, courier_id, letter_code, title, sender_name,
    target_location, pickup_op_code, delivery_op_code,
    status, priority, created_at, updated_at
)
SELECT 
    gen_random_uuid()::text,
    (SELECT id FROM couriers WHERE level = 3 LIMIT 1),
    'LC' || LPAD(FLOOR(RANDOM() * 999999 + 1000000)::text, 6, '0'),
    '跨校快递 - ' || l.title,
    u.nickname,
    '清华大学3号楼',
    'PK5F01',  -- Beijing University
    'QH3B02',  -- Tsinghua University
    'pending',
    'urgent',
    NOW(),
    NOW()
FROM letters l
JOIN users u ON l.user_id = u.id
WHERE l.created_at >= NOW() - INTERVAL '5 minutes'
LIMIT 2;

-- Show results
SELECT 'Courier hierarchy:' as info;
SELECT 
    c.id,
    u.username,
    c.level,
    c.zone,
    c.managed_op_code_prefix,
    c.status
FROM couriers c
JOIN users u ON c.user_id = u.id
ORDER BY c.level DESC, u.username;

SELECT 'Task distribution:' as info;
SELECT 
    ct.status,
    ct.priority,
    COUNT(*) as count,
    COUNT(DISTINCT ct.courier_id) as assigned_to_couriers
FROM courier_tasks ct
WHERE ct.created_at >= NOW() - INTERVAL '10 minutes'
GROUP BY ct.status, ct.priority
ORDER BY ct.status, ct.priority;