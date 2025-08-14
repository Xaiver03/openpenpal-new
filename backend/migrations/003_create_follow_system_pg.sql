-- PostgreSQL compatible Follow System Migration

-- User relationships table (follows/following)
CREATE TABLE IF NOT EXISTS user_relationships (
    id VARCHAR(36) PRIMARY KEY,
    follower_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active',
    notification_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_follow_relationship UNIQUE (follower_id, following_id),
    CONSTRAINT no_self_follow CHECK (follower_id != following_id)
);

-- User follow statistics table
CREATE TABLE IF NOT EXISTS user_follow_stats (
    user_id VARCHAR(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    followers_count INTEGER DEFAULT 0 CHECK (followers_count >= 0),
    following_count INTEGER DEFAULT 0 CHECK (following_count >= 0),
    mutual_follow_count INTEGER DEFAULT 0 CHECK (mutual_follow_count >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Follow activities (for activity feed)
CREATE TABLE IF NOT EXISTS follow_activities (
    id VARCHAR(36) PRIMARY KEY,
    actor_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- User suggestions table
CREATE TABLE IF NOT EXISTS user_suggestions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    suggested_user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    suggestion_type VARCHAR(50) DEFAULT 'similar_interests',
    score DECIMAL(3,2) DEFAULT 0.50,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_user_suggestion UNIQUE (user_id, suggested_user_id),
    CONSTRAINT no_self_suggest CHECK (user_id != suggested_user_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_relationships_follower ON user_relationships(follower_id);
CREATE INDEX IF NOT EXISTS idx_user_relationships_following ON user_relationships(following_id);
CREATE INDEX IF NOT EXISTS idx_user_relationships_status ON user_relationships(status);
CREATE INDEX IF NOT EXISTS idx_user_relationships_created ON user_relationships(created_at);

CREATE INDEX IF NOT EXISTS idx_follow_activities_actor ON follow_activities(actor_id);
CREATE INDEX IF NOT EXISTS idx_follow_activities_target ON follow_activities(target_id);
CREATE INDEX IF NOT EXISTS idx_follow_activities_type ON follow_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_follow_activities_created ON follow_activities(created_at);

CREATE INDEX IF NOT EXISTS idx_user_suggestions_user ON user_suggestions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_suggestions_suggested ON user_suggestions(suggested_user_id);
CREATE INDEX IF NOT EXISTS idx_user_suggestions_score ON user_suggestions(score DESC);

-- Function to update user follow stats
CREATE OR REPLACE FUNCTION update_follow_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.status = 'active' THEN
        -- Increment following count for follower
        INSERT INTO user_follow_stats (user_id, following_count)
        VALUES (NEW.follower_id, 1)
        ON CONFLICT (user_id) 
        DO UPDATE SET following_count = user_follow_stats.following_count + 1,
                      updated_at = NOW();
        
        -- Increment followers count for followed user
        INSERT INTO user_follow_stats (user_id, followers_count)
        VALUES (NEW.following_id, 1)
        ON CONFLICT (user_id)
        DO UPDATE SET followers_count = user_follow_stats.followers_count + 1,
                      updated_at = NOW();
                      
    ELSIF TG_OP = 'DELETE' AND OLD.status = 'active' THEN
        -- Decrement following count for follower
        UPDATE user_follow_stats 
        SET following_count = GREATEST(0, following_count - 1),
            updated_at = NOW()
        WHERE user_id = OLD.follower_id;
        
        -- Decrement followers count for followed user
        UPDATE user_follow_stats 
        SET followers_count = GREATEST(0, followers_count - 1),
            updated_at = NOW()
        WHERE user_id = OLD.following_id;
    END IF;
    
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Triggers for automatic stats updates
DROP TRIGGER IF EXISTS follow_stats_trigger ON user_relationships;
CREATE TRIGGER follow_stats_trigger
    AFTER INSERT OR DELETE ON user_relationships
    FOR EACH ROW EXECUTE FUNCTION update_follow_stats();

-- Function to update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at timestamps
CREATE TRIGGER update_user_relationships_updated_at
    BEFORE UPDATE ON user_relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_follow_stats_updated_at
    BEFORE UPDATE ON user_follow_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Initialize stats for existing users
INSERT INTO user_follow_stats (user_id, followers_count, following_count)
SELECT id, 0, 0 FROM users 
ON CONFLICT (user_id) DO NOTHING;

-- Sample follow relationships for development
INSERT INTO user_relationships (id, follower_id, following_id, status, created_at)
SELECT 
    'follow-' || u1.id || '-' || u2.id,
    u1.id,
    u2.id,
    'active',
    NOW() - (random() * interval '30 days')
FROM 
    (SELECT id FROM users WHERE username IN ('alice', 'bob') LIMIT 1) u1,
    (SELECT id FROM users WHERE username IN ('courier_level1', 'admin') LIMIT 2) u2
WHERE u1.id != u2.id
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- Sample suggestions
INSERT INTO user_suggestions (id, user_id, suggested_user_id, suggestion_type, score)
SELECT 
    'suggest-' || u1.id || '-' || u2.id,
    u1.id,
    u2.id,
    'mutual_friends',
    0.75
FROM 
    (SELECT id FROM users LIMIT 3) u1,
    (SELECT id FROM users LIMIT 3 OFFSET 1) u2
WHERE u1.id != u2.id
ON CONFLICT (user_id, suggested_user_id) DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE user_relationships IS 'User follow/following relationships';
COMMENT ON TABLE user_follow_stats IS 'Cached user follow statistics for performance';
COMMENT ON TABLE follow_activities IS 'Follow-related activities for activity feeds';
COMMENT ON TABLE user_suggestions IS 'User follow suggestions based on various algorithms';