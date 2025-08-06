-- 检查晋升系统基础数据

-- 1. 查看所有信使用户
SELECT 
    u.id,
    u.username,
    u.nickname,
    u.role,
    c.level as courier_level,
    c.zone,
    c.status,
    c.points
FROM users u
LEFT JOIN couriers c ON u.id = c.user_id
WHERE u.role LIKE '%courier%'
ORDER BY u.username;

-- 2. 查看信使等级分布
SELECT 
    level,
    COUNT(*) as count
FROM couriers
GROUP BY level
ORDER BY level;

-- 3. 查看有信使记录的用户
SELECT 
    u.username,
    c.level,
    c.zone,
    c.status,
    c.task_count,
    c.points
FROM couriers c
JOIN users u ON c.user_id = u.id
ORDER BY c.level DESC, c.points DESC;