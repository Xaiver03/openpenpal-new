-- 创建扩展用户档案表
CREATE TABLE IF NOT EXISTS user_profiles_extended (
    user_id VARCHAR(36) PRIMARY KEY,
    bio TEXT,
    school VARCHAR(100),
    op_code VARCHAR(6),
    writing_level INT DEFAULT 1 CHECK (writing_level >= 0 AND writing_level <= 5),
    courier_level INT DEFAULT 0 CHECK (courier_level >= 0 AND courier_level <= 4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_op_code (op_code),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建用户统计表
CREATE TABLE IF NOT EXISTS user_stats (
    user_id VARCHAR(36) PRIMARY KEY,
    letters_sent INT DEFAULT 0,
    letters_received INT DEFAULT 0,
    museum_contributions INT DEFAULT 0,
    total_points INT DEFAULT 0,
    writing_points INT DEFAULT 0,
    courier_points INT DEFAULT 0,
    current_streak INT DEFAULT 0,
    max_streak INT DEFAULT 0,
    last_active_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建用户隐私设置表
CREATE TABLE IF NOT EXISTS user_privacy_settings (
    user_id VARCHAR(36) PRIMARY KEY,
    show_email BOOLEAN DEFAULT FALSE,
    show_op_code BOOLEAN DEFAULT TRUE,
    show_stats BOOLEAN DEFAULT TRUE,
    op_code_privacy VARCHAR(20) DEFAULT 'partial',
    profile_visible BOOLEAN DEFAULT TRUE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建用户成就表
CREATE TABLE IF NOT EXISTS user_achievements (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    category VARCHAR(50),
    unlocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    UNIQUE KEY idx_user_achievement (user_id, code),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 为现有用户创建默认档案数据
INSERT INTO user_profiles_extended (user_id, writing_level, courier_level)
SELECT id, 1, 
    CASE 
        WHEN role = 'courier_level1' THEN 1
        WHEN role = 'courier_level2' THEN 2
        WHEN role = 'courier_level3' THEN 3
        WHEN role = 'courier_level4' THEN 4
        ELSE 0
    END
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM user_profiles_extended WHERE user_id = users.id
);

-- 为现有用户创建默认统计数据
INSERT INTO user_stats (user_id)
SELECT id FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM user_stats WHERE user_id = users.id
);

-- 为现有用户创建默认隐私设置
INSERT INTO user_privacy_settings (user_id)
SELECT id FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM user_privacy_settings WHERE user_id = users.id
);

-- 添加一些测试数据（如果需要）
-- 更新alice用户的档案
UPDATE user_profiles_extended 
SET bio = '爱好写信的学生，希望通过文字传递温暖',
    school = '北京大学',
    op_code = 'PK5F3D',
    writing_level = 3
WHERE user_id = (SELECT id FROM users WHERE username = 'alice');

UPDATE user_stats
SET letters_sent = 15,
    letters_received = 12,
    museum_contributions = 3,
    total_points = 450,
    writing_points = 320,
    current_streak = 7
WHERE user_id = (SELECT id FROM users WHERE username = 'alice');

-- 给alice用户添加成就
INSERT INTO user_achievements (user_id, code, name, description, icon, category)
SELECT id, 'first_letter', '初次来信', '发送第一封信', '✉️', 'writing'
FROM users WHERE username = 'alice';

INSERT INTO user_achievements (user_id, code, name, description, icon, category)
SELECT id, 'active_writer', '活跃写手', '发送10封信', '✍️', 'writing'
FROM users WHERE username = 'alice';

INSERT INTO user_achievements (user_id, code, name, description, icon, category)
SELECT id, 'museum_contributor', '博物馆贡献者', '贡献第一封信到博物馆', '🏛️', 'museum'
FROM users WHERE username = 'alice';