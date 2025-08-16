-- Check users
SELECT COUNT(*) as user_count FROM users;
SELECT username, role, is_active FROM users ORDER BY created_at DESC LIMIT 5;

-- Check letters
SELECT COUNT(*) as letter_count FROM letters;
SELECT id, title, status, created_at FROM letters ORDER BY created_at DESC LIMIT 5;

-- Check letter codes
SELECT COUNT(*) as code_count FROM letter_codes;
