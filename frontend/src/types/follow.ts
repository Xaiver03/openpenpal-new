/**
 * Follow System Types - SOTA Implementation
 * 关注/粉丝系统类型定义 - 支持用户关系管理、社交互动、推荐系统
 */

import type { User } from './user'

// ================================
// Core Follow Relationship Types
// ================================

export interface FollowRelationship {
  id: string
  follower_id: string
  following_id: string
  created_at: string
  updated_at: string
  status: FollowStatus
  notification_enabled: boolean
}

export type FollowStatus = 'active' | 'blocked' | 'muted'

export interface FollowUser extends User {
  // Extended user info for follow contexts
  followers_count: number
  following_count: number
  letters_count: number
  is_following?: boolean
  is_follower?: boolean
  follow_status?: FollowStatus
  followed_at?: string
  mutual_followers_count?: number
}

// ================================
// API Request/Response Types
// ================================

export interface FollowActionRequest {
  user_id: string
  notification_enabled?: boolean
}

export interface FollowActionResponse {
  success: boolean
  is_following: boolean
  follower_count: number
  following_count: number
  followed_at?: string
  message?: string
}

export interface FollowListQuery {
  page?: number
  limit?: number
  sort_by?: 'created_at' | 'nickname' | 'letters_count'
  order?: 'asc' | 'desc'
  search?: string
  school_filter?: string
  status_filter?: FollowStatus[]
}

export interface FollowListResponse {
  users: FollowUser[]
  pagination: {
    page: number
    limit: number
    total: number
    pages: number
  }
}

export interface FollowStatsResponse {
  followers_count: number
  following_count: number
  mutual_followers_count: number
  recent_followers: FollowUser[]
  popular_following: FollowUser[]
}

export interface UserSuggestionsQuery {
  limit?: number
  based_on?: 'school' | 'mutual_followers' | 'activity' | 'interests'
  exclude_followed?: boolean
  min_activity_score?: number
}

export interface UserSuggestionsResponse {
  suggestions: FollowSuggestion[]
  algorithm_used: string
  refresh_available_at: string
}

export interface UserSearchQuery {
  query: string
  page?: number
  limit?: number
  sort_by?: 'activity' | 'followers' | 'joined' | 'relevance'
  order?: 'asc' | 'desc'
  search?: string
  school_filter?: string
  status_filter?: FollowStatus[]
}

export interface FollowSuggestion {
  user: FollowUser
  reason: SuggestionReason
  confidence_score: number
  mutual_followers?: FollowUser[]
  common_interests?: string[]
}

export type SuggestionReason = 
  | 'same_school'
  | 'mutual_followers'
  | 'similar_interests'
  | 'active_user'
  | 'new_user'
  | 'trending'

// ================================
// Follow Activity & Notifications
// ================================

export interface FollowActivity {
  id: string
  type: FollowActivityType
  actor_id: string
  target_id: string
  actor: FollowUser
  target?: FollowUser
  created_at: string
  is_read: boolean
  metadata?: Record<string, any>
}

export type FollowActivityType = 
  | 'new_follower'
  | 'followed_you_back'
  | 'user_joined'
  | 'milestone_reached'

export interface FollowNotificationSettings {
  new_followers: boolean
  follow_backs: boolean
  suggestions: boolean
  activity_digest: boolean
  email_notifications: boolean
}

// ================================
// UI Component Props
// ================================

export interface FollowButtonProps {
  user_id: string
  initial_is_following?: boolean
  initial_follower_count?: number
  size?: 'sm' | 'md' | 'lg'
  variant?: 'default' | 'outline' | 'ghost'
  show_count?: boolean
  disabled?: boolean
  className?: string
  onFollowChange?: (isFollowing: boolean, followerCount: number) => void
}

export interface FollowListProps {
  user_id?: string // If not provided, use current user
  type: 'followers' | 'following'
  initial_data?: FollowUser[]
  show_stats?: boolean
  enable_search?: boolean
  enable_filters?: boolean
  max_height?: string
  className?: string
}

export interface FollowStatsProps {
  user_id?: string
  show_detailed?: boolean
  show_recent?: boolean
  compact?: boolean
  className?: string
}

export interface UserSuggestionsProps {
  limit?: number
  show_reason?: boolean
  show_mutual?: boolean
  show_refresh?: boolean
  algorithm?: UserSuggestionsQuery['based_on']
  className?: string
  onUserFollow?: (user: FollowUser) => void
}

export interface FollowActivityFeedProps {
  user_id?: string
  limit?: number
  show_own_actions?: boolean
  show_avatars?: boolean
  real_time?: boolean
  className?: string
}

// ================================
// Store State Types
// ================================

export interface FollowState {
  // Current user's follow data
  followers: FollowUser[]
  following: FollowUser[]
  follower_count: number
  following_count: number
  
  // Cached data
  suggestions: FollowSuggestion[]
  recent_activities: FollowActivity[]
  
  // UI state
  loading: {
    followers: boolean
    following: boolean
    suggestions: boolean
    follow_action: boolean
  }
  
  // Last updated timestamps
  last_updated: {
    followers: number | null
    following: number | null
    suggestions: number | null
  }
  
  // Error states
  errors: {
    followers: string | null
    following: string | null
    suggestions: string | null
    follow_action: string | null
  }
  
  // Settings
  notification_settings: FollowNotificationSettings
}

export interface FollowActions {
  // Core actions
  followUser: (userId: string, options?: { notificationEnabled?: boolean }) => Promise<boolean>
  unfollowUser: (userId: string) => Promise<boolean>
  
  // Data fetching
  loadFollowers: (userId?: string, query?: FollowListQuery) => Promise<void>
  loadFollowing: (userId?: string, query?: FollowListQuery) => Promise<void>
  loadSuggestions: (query?: UserSuggestionsQuery) => Promise<void>
  refreshSuggestions: () => Promise<void>
  
  // Bulk actions
  followMultipleUsers: (userIds: string[]) => Promise<{ success: string[]; failed: string[] }>
  removeFollower: (userId: string) => Promise<boolean>
  
  // User data
  getUserFollowStatus: (userId: string) => Promise<{ isFollowing: boolean; isFollower: boolean }>
  searchUsers: (query: string, filters?: Partial<FollowListQuery>) => Promise<FollowUser[]>
  
  // Settings
  updateNotificationSettings: (settings: Partial<FollowNotificationSettings>) => Promise<void>
  
  // Cache management
  clearCache: () => void
  invalidateUser: (userId: string) => void
}

// ================================
// Utility Types
// ================================

export interface FollowRelationshipSummary {
  mutual_followers: FollowUser[]
  followers_only: FollowUser[]
  following_only: FollowUser[]
  suggestions_based_on_mutual: FollowSuggestion[]
}

export interface FollowMetrics {
  growth_rate: {
    followers_weekly: number
    following_weekly: number
  }
  engagement_rate: number
  mutual_follow_rate: number
  suggestion_acceptance_rate: number
}

export interface FollowError {
  code?: string
  message: string
  status?: number
  details?: Record<string, any>
}

// ================================
// Search and Discovery
// ================================

export interface UserSearchQuery {
  query?: string
  school_code?: string
  role?: string
  min_followers?: number
  max_followers?: number
  active_since?: string
  sort_by?: 'followers' | 'activity' | 'joined' | 'relevance'
  order?: 'asc' | 'desc'
  limit?: number
  offset?: number
}

export interface UserSearchResponse {
  users: FollowUser[]
  total: number
  query: string
  suggestions?: string[]
  filters_applied: Partial<UserSearchQuery>
}

// ================================
// Real-time Updates
// ================================

export interface FollowWebSocketEvent {
  type: 'follow_action' | 'follow_stats_update' | 'new_suggestion'
  data: {
    user_id: string
    target_id?: string
    action?: 'follow' | 'unfollow'
    new_stats?: Partial<FollowStatsResponse>
    suggestion?: FollowSuggestion
  }
  timestamp: string
}

// ================================
// Export convenience types
// ================================

export type {
  User as BaseUser,
}

// Default exports for common interfaces
export type FollowSystemUser = FollowUser
export type FollowSystemState = FollowState & FollowActions