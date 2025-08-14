-- Migration 004: Create Privacy System Tables
-- 隐私设置系统的数据库表结构

-- Privacy settings table
CREATE TABLE IF NOT EXISTS privacy_settings (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Profile visibility settings (embedded as columns with prefix)
    profile_bio VARCHAR(20) DEFAULT 'school',
    profile_school_info VARCHAR(20) DEFAULT 'public',
    profile_contact_info VARCHAR(20) DEFAULT 'friends',
    profile_activity_feed VARCHAR(20) DEFAULT 'school',
    profile_follow_lists VARCHAR(20) DEFAULT 'friends',
    profile_statistics VARCHAR(20) DEFAULT 'public',
    profile_last_active VARCHAR(20) DEFAULT 'school',
    
    -- Social privacy settings (embedded as columns with prefix)
    social_allow_follow_requests BOOLEAN DEFAULT TRUE,
    social_allow_comments BOOLEAN DEFAULT TRUE,
    social_allow_direct_messages BOOLEAN DEFAULT TRUE,
    social_show_in_discovery BOOLEAN DEFAULT TRUE,
    social_show_in_suggestions BOOLEAN DEFAULT TRUE,
    social_allow_school_search BOOLEAN DEFAULT TRUE,
    
    -- Notification privacy settings (embedded as columns with prefix)
    notification_new_followers BOOLEAN DEFAULT TRUE,
    notification_follow_requests BOOLEAN DEFAULT TRUE,
    notification_comments BOOLEAN DEFAULT TRUE,
    notification_mentions BOOLEAN DEFAULT TRUE,
    notification_direct_messages BOOLEAN DEFAULT TRUE,
    notification_system_updates BOOLEAN DEFAULT TRUE,
    notification_email_notifications BOOLEAN DEFAULT FALSE,
    
    -- Blocking settings (embedded as columns with prefix)
    blocking_blocked_users JSONB DEFAULT '[]'::jsonb,
    blocking_muted_users JSONB DEFAULT '[]'::jsonb,
    blocking_blocked_keywords JSONB DEFAULT '[]'::jsonb,
    blocking_auto_block_new_accounts BOOLEAN DEFAULT FALSE,
    blocking_block_non_school_users BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for privacy settings
CREATE INDEX IF NOT EXISTS idx_privacy_settings_user_id ON privacy_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_privacy_settings_updated_at ON privacy_settings(updated_at);

-- Privacy audit log table for tracking privacy-related actions
CREATE TABLE IF NOT EXISTS privacy_audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    target_user_id VARCHAR(36),
    target_data TEXT,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for privacy audit logs
CREATE INDEX IF NOT EXISTS idx_privacy_audit_logs_user_id ON privacy_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_privacy_audit_logs_action ON privacy_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_privacy_audit_logs_target_user_id ON privacy_audit_logs(target_user_id);
CREATE INDEX IF NOT EXISTS idx_privacy_audit_logs_created_at ON privacy_audit_logs(created_at);

-- User privacy cache table for performance optimization
CREATE TABLE IF NOT EXISTS privacy_cache (
    id VARCHAR(36) PRIMARY KEY,
    viewer_user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permissions JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(viewer_user_id, target_user_id)
);

-- Indexes for privacy cache
CREATE INDEX IF NOT EXISTS idx_privacy_cache_viewer_target ON privacy_cache(viewer_user_id, target_user_id);
CREATE INDEX IF NOT EXISTS idx_privacy_cache_expires_at ON privacy_cache(expires_at);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_privacy_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on privacy_settings
DROP TRIGGER IF EXISTS privacy_settings_update_timestamp ON privacy_settings;
CREATE TRIGGER privacy_settings_update_timestamp
    BEFORE UPDATE ON privacy_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_privacy_settings_updated_at();

-- Function to clean up expired privacy cache entries
CREATE OR REPLACE FUNCTION cleanup_expired_privacy_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM privacy_cache WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE privacy_settings IS 'User privacy settings and preferences';
COMMENT ON TABLE privacy_audit_logs IS 'Audit log for privacy-related actions and changes';
COMMENT ON TABLE privacy_cache IS 'Performance cache for privacy permission checks';

COMMENT ON COLUMN privacy_settings.profile_bio IS 'Visibility level for user bio: public, school, friends, private';
COMMENT ON COLUMN privacy_settings.profile_school_info IS 'Visibility level for school information';
COMMENT ON COLUMN privacy_settings.profile_contact_info IS 'Visibility level for contact information';
COMMENT ON COLUMN privacy_settings.profile_activity_feed IS 'Visibility level for activity feed';
COMMENT ON COLUMN privacy_settings.profile_follow_lists IS 'Visibility level for follower/following lists';
COMMENT ON COLUMN privacy_settings.profile_statistics IS 'Visibility level for user statistics';
COMMENT ON COLUMN privacy_settings.profile_last_active IS 'Visibility level for last active timestamp';

COMMENT ON COLUMN privacy_settings.social_allow_follow_requests IS 'Whether to allow follow requests from other users';
COMMENT ON COLUMN privacy_settings.social_allow_comments IS 'Whether to allow comments on profile';
COMMENT ON COLUMN privacy_settings.social_allow_direct_messages IS 'Whether to allow direct messages';
COMMENT ON COLUMN privacy_settings.social_show_in_discovery IS 'Whether to show profile in discovery/search';
COMMENT ON COLUMN privacy_settings.social_show_in_suggestions IS 'Whether to show in user suggestions';
COMMENT ON COLUMN privacy_settings.social_allow_school_search IS 'Whether to allow search by school members';

COMMENT ON COLUMN privacy_settings.blocking_blocked_users IS 'JSON array of blocked user IDs';
COMMENT ON COLUMN privacy_settings.blocking_muted_users IS 'JSON array of muted user IDs';
COMMENT ON COLUMN privacy_settings.blocking_blocked_keywords IS 'JSON array of blocked keywords for content filtering';

-- Sample privacy settings for development/testing
INSERT INTO privacy_settings (
    id, user_id,
    profile_bio, profile_school_info, profile_contact_info,
    social_allow_follow_requests, social_allow_comments,
    notification_new_followers, notification_email_notifications,
    created_at, updated_at
) VALUES (
    'privacy-demo-001', 
    (SELECT id FROM users WHERE username = 'alice' LIMIT 1),
    'school', 'public', 'friends',
    true, true,
    true, false,
    NOW(), NOW()
) ON CONFLICT (user_id) DO NOTHING;

-- Privacy system statistics view
CREATE OR REPLACE VIEW privacy_stats AS
SELECT
    COUNT(*) as total_settings,
    COUNT(CASE WHEN social_show_in_discovery = true THEN 1 END) as discoverable_users,
    COUNT(CASE WHEN social_allow_follow_requests = true THEN 1 END) as accept_follow_requests,
    COUNT(CASE WHEN notification_email_notifications = true THEN 1 END) as email_notifications_enabled,
    AVG(CASE 
        WHEN profile_bio = 'public' THEN 4
        WHEN profile_bio = 'school' THEN 3
        WHEN profile_bio = 'friends' THEN 2
        WHEN profile_bio = 'private' THEN 1
        ELSE 0
    END) as avg_bio_openness,
    NOW() as calculated_at
FROM privacy_settings;

COMMENT ON VIEW privacy_stats IS 'Privacy system usage statistics and insights';