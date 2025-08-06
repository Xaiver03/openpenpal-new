SELECT username, role, is_active, created_at 
FROM users 
WHERE username IN ('admin', 'courier_level4_city', 'courier_level3_school', 
                   'courier_level2_zone', 'courier_level1_building', 'test_user')
ORDER BY username;
