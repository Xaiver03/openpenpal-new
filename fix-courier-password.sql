-- Fix courier_level1 password to 'password'
UPDATE users 
SET password_hash = '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'
WHERE username = 'courier_level1';

-- Verify the update
SELECT username, email, password_hash, role, is_active 
FROM users 
WHERE username IN ('courier_level1', 'courier1', 'admin', 'user1')
ORDER BY username;