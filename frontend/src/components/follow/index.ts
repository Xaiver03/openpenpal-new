/**
 * Follow System Components - Export Index
 * 关注系统组件导出 - 统一导出所有关注相关组件
 */

// Core components
export { FollowButton, CompactFollowButton, FollowButtonWithCount, HeartFollowButton } from './follow-button'
export { FollowStats, InlineFollowStats, BadgeFollowStats } from './follow-stats'
export { FollowList, FollowListTabs } from './follow-list'
export { UserCard, UserCardSkeleton } from './user-card'
export { UserSuggestions, CompactUserSuggestions } from './user-suggestions'

// Re-export types for convenience
export type {
  FollowButtonProps,
  FollowStatsProps,
  FollowListProps,
} from '@/types/follow'

// Re-export stores and hooks for convenience
export {
  useFollow,
  useFollowActions,
  useFollowStatus,
  useFollowStore
} from '@/stores/follow-store'

// Re-export API
export { followApi } from '@/lib/api/follow'