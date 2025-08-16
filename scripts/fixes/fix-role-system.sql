-- OpenPenPal 角色系统简化脚本
-- 根据PRD要求，只保留：user, courier_level1-4, super_admin

-- 1. 首先查看当前系统中的所有角色
SELECT DISTINCT role, COUNT(*) as count FROM users GROUP BY role;

-- 2. 更新不符合PRD的角色
-- 将所有其他管理员角色统一为super_admin
UPDATE users SET role = 'super_admin' 
WHERE role IN ('platform_admin', 'school_admin');

-- 将所有其他信使角色映射到相应的level
UPDATE users SET role = 'courier_level1' 
WHERE role IN ('courier');

UPDATE users SET role = 'courier_level2' 
WHERE role IN ('senior_courier');

UPDATE users SET role = 'courier_level3' 
WHERE role IN ('courier_coordinator');

-- 3. 再次检查角色分布
SELECT DISTINCT role, COUNT(*) as count FROM users GROUP BY role ORDER BY role;

-- 4. 验证结果 - 应该只有以下角色：
-- user, courier_level1, courier_level2, courier_level3, courier_level4, super_admin