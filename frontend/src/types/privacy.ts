/**
 * Privacy Settings Types - Comprehensive profile privacy controls
 * Based on SOTA privacy patterns and OpenPenPal requirements
 */

// Privacy level enumeration
export type PrivacyLevel = 
  | 'public'      // Anyone can see
  | 'school'      // Same school users only
  | 'friends'     // Following/followers only
  | 'private'     // Only self

// Profile field visibility settings
export interface ProfileVisibility {
  bio: PrivacyLevel
  school_info: PrivacyLevel
  contact_info: PrivacyLevel
  op_code: PrivacyLevel           // OP Code visibility
  activity_feed: PrivacyLevel
  follow_lists: PrivacyLevel
  statistics: PrivacyLevel
  achievements: PrivacyLevel       // Achievement badges visibility
  public_letters: PrivacyLevel     // Public letters visibility
  last_active: PrivacyLevel
}

// Social interaction privacy settings
export interface SocialPrivacy {
  allow_follow_requests: boolean
  allow_comments: boolean
  allow_direct_messages: boolean
  show_in_discovery: boolean
  show_in_suggestions: boolean
  allow_school_search: boolean
}

// Notification privacy settings
export interface NotificationPrivacy {
  new_followers: boolean
  follow_requests: boolean
  comments: boolean
  mentions: boolean
  direct_messages: boolean
  system_updates: boolean
  email_notifications: boolean
}

// Blocking and muting settings
export interface BlockingSettings {
  blocked_users: string[]         // User IDs
  muted_users: string[]          // User IDs  
  blocked_keywords: string[]     // Content filtering
  auto_block_new_accounts: boolean
  block_non_school_users: boolean
}

// Complete privacy settings structure
export interface PrivacySettings {
  id: string
  user_id: string
  profile_visibility: ProfileVisibility
  social_privacy: SocialPrivacy
  notification_privacy: NotificationPrivacy
  blocking_settings: BlockingSettings
  updated_at: string
  created_at: string
}

// Update privacy settings request
export interface UpdatePrivacySettingsRequest {
  profile_visibility?: Partial<ProfileVisibility>
  social_privacy?: Partial<SocialPrivacy>
  notification_privacy?: Partial<NotificationPrivacy>
  blocking_settings?: Partial<BlockingSettings>
}

// Privacy check result
export interface PrivacyCheckResult {
  can_view_profile: boolean
  can_view_bio: boolean
  can_view_school: boolean
  can_view_contact: boolean
  can_view_activity: boolean
  can_view_followers: boolean
  can_view_following: boolean
  can_view_stats: boolean
  can_follow: boolean
  can_comment: boolean
  can_message: boolean
  reason?: string
}

// Default privacy settings for new users
export const DEFAULT_PRIVACY_SETTINGS: Omit<PrivacySettings, 'id' | 'user_id' | 'created_at' | 'updated_at'> = {
  profile_visibility: {
    bio: 'school',
    school_info: 'public',
    contact_info: 'friends',
    op_code: 'school',           // Default: visible to same school
    activity_feed: 'school',
    follow_lists: 'friends',
    statistics: 'public',
    achievements: 'public',       // Default: public visible
    public_letters: 'public',     // Default: public visible
    last_active: 'school'
  },
  social_privacy: {
    allow_follow_requests: true,
    allow_comments: true,
    allow_direct_messages: true,
    show_in_discovery: true,
    show_in_suggestions: true,
    allow_school_search: true
  },
  notification_privacy: {
    new_followers: true,
    follow_requests: true,
    comments: true,
    mentions: true,
    direct_messages: true,
    system_updates: true,
    email_notifications: false
  },
  blocking_settings: {
    blocked_users: [],
    muted_users: [],
    blocked_keywords: [],
    auto_block_new_accounts: false,
    block_non_school_users: false
  }
}

// Privacy level labels for UI
export const PRIVACY_LEVEL_LABELS: Record<PrivacyLevel, string> = {
  public: '公开',
  school: '同校可见',
  friends: '好友可见',
  private: '仅自己'
}

// Privacy level descriptions
export const PRIVACY_LEVEL_DESCRIPTIONS: Record<PrivacyLevel, string> = {
  public: '所有用户都能看到',
  school: '只有同校用户可以看到',
  friends: '只有关注你或你关注的用户可以看到',
  private: '只有你自己可以看到'
}

// Profile field labels
export const PROFILE_FIELD_LABELS: Record<keyof ProfileVisibility, string> = {
  bio: '个人简介',
  school_info: '学校信息',
  contact_info: '联系方式',
  op_code: 'OP Code',
  activity_feed: '活动动态',
  follow_lists: '关注列表',
  statistics: '统计信息',
  achievements: '成就徽章',
  public_letters: '公开信件',
  last_active: '最近活跃'
}