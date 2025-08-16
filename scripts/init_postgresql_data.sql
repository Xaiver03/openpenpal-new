-- Initialize PostgreSQL with basic OpenPenPal data
-- Run this after clearing the database

-- Insert test users
INSERT INTO users (id, username, email, password_hash, nickname, role, school_code, is_active, created_at, updated_at) VALUES
('user-alice', 'alice', 'alice@example.com', '$2a$12$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW', 'Alice', 'user', 'BJDX01', true, NOW(), NOW()),
('user-bob', 'bob', 'bob@example.com', '$2a$12$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW', 'Bob', 'user', 'BJDX01', true, NOW(), NOW()),
('admin-1', 'admin', 'admin@example.com', '$2a$12$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW', 'Admin', 'super_admin', 'SYSTEM', true, NOW(), NOW());

-- Insert test letters
INSERT INTO letters (id, user_id, title, content, style, status, visibility, created_at, updated_at) VALUES
('letter-1', 'user-alice', '测试信件', '这是一封测试信件的内容', 'classic', 'delivered', 'private', NOW(), NOW()),
('letter-2', 'user-bob', '感谢信', '感谢你的帮助', 'modern', 'delivered', 'public', NOW(), NOW());

-- Insert letter codes
INSERT INTO letter_codes (id, letter_id, code, created_at, updated_at) VALUES
('code-1', 'letter-1', 'LC000001', NOW(), NOW()),
('code-2', 'letter-2', 'LC000002', NOW(), NOW());

-- Insert couriers
INSERT INTO couriers (id, user_id, name, contact, school, zone, level, status, managed_op_code_prefix, created_at, updated_at) VALUES
('courier-1', 'user-alice', 'Alice Courier', 'alice@example.com', '北京大学', 'BJDX-A-101', 1, 'approved', 'BJDX5F01', NOW(), NOW());

-- Insert basic schools data
INSERT INTO schools (id, code, name, province, city, created_at, updated_at) VALUES
('school-1', 'PK', '北京大学', '北京', '北京市', NOW(), NOW()),
('school-2', 'QH', '清华大学', '北京', '北京市', NOW(), NOW());

-- Success message
SELECT 'PostgreSQL data initialization completed successfully!' as result;