/**
 * Comment Components - Unified export for all comment-related components
 * 评论组件 - 统一导出所有评论相关组件
 */

// Core comment components
export { default as CommentList } from './comment-list'
export { default as CommentItem } from './comment-item'
export { default as CommentForm } from './comment-form'
export { default as CommentStats, CommentCountBadge } from './comment-stats'

// Re-export types for convenience
export type {
  Comment,
  CommentUser,
  CommentStatus,
  CommentCreateRequest,
  CommentUpdateRequest,
  CommentListQuery,
  CommentListResponse,
  CommentStatsResponse,
  CommentFormData,
  CommentAction,
  CommentTreeNode,
  CommentListProps,
  CommentItemProps,
  CommentFormProps,
  CommentStatsProps,
  CommentError,
} from '@/types/comment'

// Re-export hooks for convenience
export { useComments, useCommentList, useCommentStats } from '@/hooks/use-comments'

// Re-export API for advanced usage
export { commentApi } from '@/lib/api/comment'