/**
 * SOTA Comment Types - Enhanced Multi-Target Comment System
 * SOTA评论系统类型定义 - 增强的多目标评论系统
 */

// ================================
// SOTA Core Types
// ================================

export type CommentTargetType = 'letter' | 'profile' | 'museum';

export type CommentStatus = 'active' | 'pending' | 'hidden' | 'deleted' | 'rejected';

export interface CommentUser {
  id: string;
  username: string;
  nickname: string;
  avatar?: string;
  role?: string;
  level?: number;
  is_verified?: boolean;
}

// ================================
// SOTA Comment Models
// ================================

export interface CommentSOTA {
  // Core fields
  id: string;
  target_id: string;
  target_type: CommentTargetType;
  user_id: string;
  parent_id?: string;
  root_id?: string;
  content: string;
  status: CommentStatus;

  // SOTA hierarchy fields
  level: number;
  path?: string;

  // Statistics
  like_count: number;
  reply_count: number;
  report_count: number;
  net_likes: number;

  // SOTA enhancement fields
  is_top: boolean;
  is_anonymous: boolean;
  is_edited: boolean;

  // Timestamps
  created_at: string;
  updated_at: string;

  // Relationships
  user?: CommentUser;
  replies?: CommentSOTA[];

  // User permissions and interactions
  is_liked: boolean;
  can_edit: boolean;
  can_delete: boolean;
  can_report: boolean;

  // Backward compatibility
  letter_id?: string;

  // UI state
  is_loading_replies?: boolean;
  is_expanded?: boolean;
}

// ================================
// SOTA Request/Response Types
// ================================

export interface CommentCreateRequestSOTA {
  // Multi-target support
  target_id: string;
  target_type: CommentTargetType;
  
  // Backward compatibility
  letter_id?: string;
  
  // Core fields
  parent_id?: string;
  content: string;
  is_anonymous?: boolean;
}

export interface CommentUpdateRequestSOTA {
  content: string;
  status?: CommentStatus;
}

export interface CommentListQuerySOTA {
  // Target specification
  target_id: string;
  target_type: CommentTargetType;
  
  // Backward compatibility
  letter_id?: string;
  
  // Filtering
  status?: CommentStatus;
  author_id?: string;
  parent_id?: string;
  root_id?: string;
  max_level?: number;
  only_top_level?: boolean;
  include_replies?: boolean;
  
  // Sorting and pagination
  sort_by?: 'created_at' | 'like_count' | 'reply_count';
  order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface CommentListResponseSOTA {
  comments: CommentSOTA[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
  stats: CommentStatsSOTA;
}

export interface CommentStatsSOTA {
  total_comments: number;
  total_replies: number;
  total_likes: number;
  active_comments: number;
  pending_comments: number;
  reported_comments: number;
}

// ================================
// SOTA Reporting Types
// ================================

export interface CommentReportRequestSOTA {
  reason: 'spam' | 'inappropriate' | 'offensive' | 'false_info' | 'other';
  description?: string;
}

export interface CommentModerationRequestSOTA {
  status: CommentStatus;
  reason?: string;
}

// ================================
// SOTA Component Props
// ================================

export interface CommentSystemPropsSOTA {
  target_id: string;
  target_type: CommentTargetType;
  target_title?: string;
  max_depth?: number;
  enable_nested?: boolean;
  show_stats?: boolean;
  allow_comments?: boolean;
  allow_anonymous?: boolean;
  initial_sort?: CommentListQuerySOTA['sort_by'];
  max_display?: number;
  className?: string;
}

export interface CommentItemPropsSOTA {
  comment: CommentSOTA;
  depth?: number;
  max_depth?: number;
  enable_reply?: boolean;
  enable_like?: boolean;
  enable_edit?: boolean;
  enable_delete?: boolean;
  enable_report?: boolean;
  show_replies?: boolean;
  show_path?: boolean;
  on_action?: (action: CommentActionSOTA) => void;
  className?: string;
}

export interface CommentFormPropsSOTA {
  target_id: string;
  target_type: CommentTargetType;
  parent_id?: string;
  parent_comment?: CommentSOTA;
  placeholder?: string;
  max_length?: number;
  auto_focus?: boolean;
  allow_anonymous?: boolean;
  submit_text?: string;
  cancel_text?: string;
  on_submit?: (data: CommentFormDataSOTA) => Promise<void>;
  on_cancel?: () => void;
  className?: string;
}

export interface CommentStatsPropsSOTA {
  target_id: string;
  target_type: CommentTargetType;
  show_icon?: boolean;
  format?: 'compact' | 'full';
  real_time?: boolean;
  className?: string;
}

// ================================
// SOTA Action Types
// ================================

export interface CommentActionSOTA {
  type: 'create' | 'reply' | 'edit' | 'delete' | 'like' | 'report' | 'moderate';
  comment_id?: string;
  data?: any;
}

export interface CommentFormDataSOTA {
  content: string;
  parent_id?: string;
  is_anonymous?: boolean;
}

// ================================
// SOTA Hook Types
// ================================

export interface UseCommentsSOTAOptions {
  target_id: string;
  target_type: CommentTargetType;
  initial_query?: Partial<CommentListQuerySOTA>;
  auto_load?: boolean;
  enable_real_time?: boolean;
  enable_polling?: boolean;
  polling_interval?: number;
}

export interface UseCommentsSOTAReturn {
  // Data
  comments: CommentSOTA[];
  stats: CommentStatsSOTA | null;
  loading: boolean;
  error: string | null;
  has_more: boolean;
  
  // Actions
  load_comments: (query?: Partial<CommentListQuerySOTA>) => Promise<void>;
  create_comment: (data: CommentCreateRequestSOTA) => Promise<CommentSOTA>;
  update_comment: (id: string, data: CommentUpdateRequestSOTA) => Promise<CommentSOTA>;
  delete_comment: (id: string) => Promise<void>;
  like_comment: (id: string) => Promise<{ is_liked: boolean; like_count: number }>;
  report_comment: (id: string, data: CommentReportRequestSOTA) => Promise<void>;
  moderate_comment: (id: string, data: CommentModerationRequestSOTA) => Promise<void>;
  load_replies: (parent_id: string, query?: Partial<CommentListQuerySOTA>) => Promise<CommentSOTA[]>;
  
  // Utilities
  refresh: () => Promise<void>;
  clear_error: () => void;
  get_comment_by_id: (id: string) => CommentSOTA | undefined;
  get_replies: (parent_id: string) => CommentSOTA[];
  build_comment_tree: (comments: CommentSOTA[]) => CommentSOTA[];
  toggle_replies_expanded: (comment_id: string) => void;
}

// ================================
// SOTA Tree and Hierarchy Types
// ================================

export interface CommentTreeNodeSOTA extends CommentSOTA {
  children: CommentTreeNodeSOTA[];
  depth: number;
  is_expanded: boolean;
  is_last_in_branch: boolean;
  ancestors: string[];
}

export interface CommentHierarchySOTA {
  roots: CommentSOTA[];
  total_depth: number;
  total_nodes: number;
  by_level: CommentSOTA[][];
  by_parent: Record<string, CommentSOTA[]>;
}

// ================================
// SOTA Error Types
// ================================

export interface CommentErrorSOTA extends Error {
  code?: string;
  status?: number;
  target_type?: CommentTargetType;
  target_id?: string;
  comment_id?: string;
  details?: any;
}

// ================================
// SOTA Real-time Event Types
// ================================

export interface CommentEventSOTA {
  type: 'comment_created' | 'comment_updated' | 'comment_deleted' | 'comment_liked' | 'comment_reported';
  target_id: string;
  target_type: CommentTargetType;
  comment_id: string;
  user_id: string;
  data?: any;
  timestamp: string;
}

// ================================
// SOTA API Response Types
// ================================

export interface CommentAPIResponseSOTA<T> {
  success: boolean;
  message: string;
  data: T;
  errors?: string[];
}

export interface CommentTargetInfoSOTA {
  id: string;
  type: CommentTargetType;
  title?: string;
  description?: string;
  owner_id?: string;
  allow_comments: boolean;
  allow_anonymous: boolean;
  moderation_enabled: boolean;
}

// ================================
// SOTA Analytics Types
// ================================

export interface CommentAnalyticsSOTA {
  engagement_rate: number;
  avg_comments_per_target: number;
  top_commenters: CommentUser[];
  activity_timeline: Array<{
    date: string;
    comments: number;
    replies: number;
    likes: number;
  }>;
  popular_targets: Array<{
    target_id: string;
    target_type: CommentTargetType;
    comment_count: number;
    engagement_score: number;
  }>;
}

// ================================
// Utility Types
// ================================

export type CommentPermissionSOTA = {
  can_view: boolean;
  can_comment: boolean;
  can_reply: boolean;
  can_like: boolean;
  can_edit: boolean;
  can_delete: boolean;
  can_report: boolean;
  can_moderate: boolean;
}

export type CommentDisplayModeSOTA = 'threaded' | 'flat' | 'tree' | 'timeline';

export type CommentSortModeSOTA = 'newest' | 'oldest' | 'popular' | 'controversial' | 'most_replied';

// ================================
// Export Compatibility Types
// ================================

// For backward compatibility with existing components
export type Comment = CommentSOTA;
export type CommentListQuery = CommentListQuerySOTA;
export type CommentListResponse = CommentListResponseSOTA;
export type CommentCreateRequest = CommentCreateRequestSOTA;
export type CommentUpdateRequest = CommentUpdateRequestSOTA;
export type CommentAction = CommentActionSOTA;
export type CommentFormData = CommentFormDataSOTA;