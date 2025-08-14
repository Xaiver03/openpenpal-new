-- Migration 003: Create Follow System Tables
-- Date: 2025-08-14
-- Description: Create tables for user follow/follower relationships, suggestions, and activities

-- 用户关系表 (用户关注关系)
CREATE TABLE IF NOT EXISTS user_relationships (
    id VARCHAR(36) PRIMARY KEY,
    follower_id VARCHAR(36) NOT NULL,
    following_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    notification_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 外键约束
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 索引
    INDEX idx_user_relationships_follower (follower_id),
    INDEX idx_user_relationships_following (following_id),
    INDEX idx_user_relationships_status (status),
    
    -- 唯一约束：防止重复关注
    UNIQUE KEY unique_follow_relationship (follower_id, following_id),
    
    -- 检查约束：用户不能关注自己
    CONSTRAINT chk_no_self_follow CHECK (follower_id != following_id)
);

-- 用户关注统计表
CREATE TABLE IF NOT EXISTS user_follow_stats (
    user_id VARCHAR(36) PRIMARY KEY,
    followers_count INT NOT NULL DEFAULT 0,
    following_count INT NOT NULL DEFAULT 0,
    mutual_follows_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 索引
    INDEX idx_follow_stats_followers (followers_count),
    INDEX idx_follow_stats_following (following_count)
);

-- 关注活动记录表
CREATE TABLE IF NOT EXISTS follow_activities (
    id VARCHAR(36) PRIMARY KEY,
    actor_id VARCHAR(36) NOT NULL,
    target_id VARCHAR(36) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 索引
    INDEX idx_follow_activities_actor (actor_id),
    INDEX idx_follow_activities_target (target_id),
    INDEX idx_follow_activities_type (type),
    INDEX idx_follow_activities_created (created_at)
);

-- 用户推荐表
CREATE TABLE IF NOT EXISTS user_suggestions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    suggested_user_id VARCHAR(36) NOT NULL,
    reason VARCHAR(100),
    score DECIMAL(5,4) DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (suggested_user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 索引
    INDEX idx_user_suggestions_user (user_id),
    INDEX idx_user_suggestions_suggested (suggested_user_id),
    INDEX idx_user_suggestions_score (score),
    INDEX idx_user_suggestions_created (created_at),
    
    -- 唯一约束：防止重复推荐
    UNIQUE KEY unique_user_suggestion (user_id, suggested_user_id)
);

-- 初始化所有现有用户的关注统计
INSERT IGNORE INTO user_follow_stats (user_id, followers_count, following_count, mutual_follows_count, updated_at)
SELECT id, 0, 0, 0, NOW()
FROM users
WHERE is_active = true;

-- 创建触发器：自动维护关注统计
DELIMITER //

-- 关注时增加统计
CREATE TRIGGER IF NOT EXISTS tr_follow_stats_insert
AFTER INSERT ON user_relationships
FOR EACH ROW
BEGIN
    IF NEW.status = 'active' THEN
        -- 增加被关注者的粉丝数
        INSERT INTO user_follow_stats (user_id, followers_count, following_count, updated_at)
        VALUES (NEW.following_id, 1, 0, NOW())
        ON DUPLICATE KEY UPDATE 
            followers_count = followers_count + 1,
            updated_at = NOW();
            
        -- 增加关注者的关注数
        INSERT INTO user_follow_stats (user_id, followers_count, following_count, updated_at)
        VALUES (NEW.follower_id, 0, 1, NOW())
        ON DUPLICATE KEY UPDATE 
            following_count = following_count + 1,
            updated_at = NOW();
    END IF;
END//

-- 取消关注时减少统计
CREATE TRIGGER IF NOT EXISTS tr_follow_stats_delete
AFTER DELETE ON user_relationships
FOR EACH ROW
BEGIN
    IF OLD.status = 'active' THEN
        -- 减少被关注者的粉丝数
        UPDATE user_follow_stats 
        SET followers_count = GREATEST(followers_count - 1, 0),
            updated_at = NOW()
        WHERE user_id = OLD.following_id;
        
        -- 减少关注者的关注数
        UPDATE user_follow_stats 
        SET following_count = GREATEST(following_count - 1, 0),
            updated_at = NOW()
        WHERE user_id = OLD.follower_id;
    END IF;
END//

-- 更新关注状态时调整统计
CREATE TRIGGER IF NOT EXISTS tr_follow_stats_update
AFTER UPDATE ON user_relationships
FOR EACH ROW
BEGIN
    -- 从非活跃变为活跃
    IF OLD.status != 'active' AND NEW.status = 'active' THEN
        UPDATE user_follow_stats 
        SET followers_count = followers_count + 1,
            updated_at = NOW()
        WHERE user_id = NEW.following_id;
        
        UPDATE user_follow_stats 
        SET following_count = following_count + 1,
            updated_at = NOW()
        WHERE user_id = NEW.follower_id;
    -- 从活跃变为非活跃
    ELSEIF OLD.status = 'active' AND NEW.status != 'active' THEN
        UPDATE user_follow_stats 
        SET followers_count = GREATEST(followers_count - 1, 0),
            updated_at = NOW()
        WHERE user_id = NEW.following_id;
        
        UPDATE user_follow_stats 
        SET following_count = GREATEST(following_count - 1, 0),
            updated_at = NOW()
        WHERE user_id = NEW.follower_id;
    END IF;
END//

DELIMITER ;

-- 为现有用户创建一些示例关注关系（开发环境）
-- 注意：这些数据仅用于开发测试，生产环境应该删除
INSERT IGNORE INTO user_relationships (id, follower_id, following_id, status, notification_enabled, created_at, updated_at)
SELECT 
    CONCAT(u1.username, '-follows-', u2.username) as id,
    u1.id as follower_id,
    u2.id as following_id,
    'active' as status,
    true as notification_enabled,
    NOW() as created_at,
    NOW() as updated_at
FROM users u1 
CROSS JOIN users u2 
WHERE u1.id != u2.id 
  AND u1.username IN ('alice', 'admin') 
  AND u2.username IN ('alice', 'admin', 'courier_level1', 'courier_level2')
  AND u1.username != u2.username
LIMIT 5;

-- 创建一些用户推荐数据（开发环境）
INSERT IGNORE INTO user_suggestions (id, user_id, suggested_user_id, reason, score, created_at)
SELECT 
    CONCAT('suggestion-', u1.username, '-', u2.username) as id,
    u1.id as user_id,
    u2.id as suggested_user_id,
    CASE 
        WHEN u1.school_code = u2.school_code THEN '同校推荐'
        ELSE '活跃用户'
    END as reason,
    ROUND(RAND() * 0.5 + 0.5, 4) as score,
    NOW() as created_at
FROM users u1 
CROSS JOIN users u2 
WHERE u1.id != u2.id 
  AND u1.is_active = true 
  AND u2.is_active = true
  AND NOT EXISTS (
    SELECT 1 FROM user_relationships ur 
    WHERE ur.follower_id = u1.id AND ur.following_id = u2.id
  )
LIMIT 20;

-- 添加一些关注活动记录（开发环境）
INSERT IGNORE INTO follow_activities (id, actor_id, target_id, type, created_at)
SELECT 
    CONCAT('activity-', ur.id) as id,
    ur.follower_id as actor_id,
    ur.following_id as target_id,
    'new_follower' as type,
    ur.created_at
FROM user_relationships ur
WHERE ur.status = 'active'
LIMIT 10;