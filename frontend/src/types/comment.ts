// Comment type definitions - aligned with backend models
// 评论系统类型定义 - 与后端模型对齐

export interface CommentUser {
  id: string;
  username: string;
  nickname: string;
  avatar?: string;
}

export interface Comment {
  id: string;
  letter_id: string;
  user_id: string;
  parent_id?: string;
  content: string;
  status: CommentStatus;
  like_count: number;
  reply_count: number;
  is_top: boolean;
  created_at: string;
  updated_at: string;
  user?: CommentUser;
  is_liked?: boolean;
  // For nested display
  replies?: Comment[];
  is_loading_replies?: boolean;
}

export type CommentStatus = 'active' | 'deleted' | 'hidden';

// Request/Response DTOs matching backend

export interface CommentCreateRequest {
  letter_id: string;
  parent_id?: string;
  content: string;
}

export interface CommentUpdateRequest {
  content: string;
  status?: CommentStatus;
}

export interface CommentListQuery {
  page?: number;
  limit?: number;
  sort_by?: 'created_at' | 'like_count';
  order?: 'asc' | 'desc';
  only_top_level?: boolean;
  parent_id?: string;
  letter_id?: string;
}

export interface CommentListResponse {
  comments: Comment[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    pages: number;
  };
}

export interface CommentStatsResponse {
  comment_count: number;
}

// UI-specific types

export interface CommentFormData {
  content: string;
  parent_id?: string;
}

export interface CommentAction {
  type: 'reply' | 'edit' | 'delete' | 'like' | 'report';
  comment_id: string;
  data?: any;
}

export interface CommentTreeNode extends Comment {
  children: CommentTreeNode[];
  depth: number;
  is_expanded: boolean;
}

// Component Props

export interface CommentListProps {
  letter_id: string;
  max_depth?: number;
  enable_nested?: boolean;
  show_stats?: boolean;
  allow_comments?: boolean;
  initial_sort?: CommentListQuery['sort_by'];
  className?: string;
}

export interface CommentItemProps {
  comment: Comment;
  depth?: number;
  max_depth?: number;
  enable_reply?: boolean;
  enable_like?: boolean;
  enable_edit?: boolean;
  enable_delete?: boolean;
  show_replies?: boolean;
  on_action?: (action: CommentAction) => void;
  className?: string;
}

export interface CommentFormProps {
  letter_id: string;
  parent_id?: string;
  parent_comment?: Comment;
  placeholder?: string;
  max_length?: number;
  auto_focus?: boolean;
  submit_text?: string;
  cancel_text?: string;
  on_submit?: (data: CommentFormData) => Promise<void>;
  on_cancel?: () => void;
  className?: string;
}

export interface CommentStatsProps {
  letter_id: string;
  show_icon?: boolean;
  format?: 'compact' | 'full';
  className?: string;
}

// Hook types

export interface UseCommentsOptions {
  letter_id: string;
  initial_query?: Partial<CommentListQuery>;
  auto_load?: boolean;
  enable_real_time?: boolean;
}

export interface UseCommentsReturn {
  comments: Comment[];
  stats: CommentStatsResponse | null;
  loading: boolean;
  error: string | null;
  has_more: boolean;
  
  // Actions
  load_comments: (query?: Partial<CommentListQuery>) => Promise<void>;
  create_comment: (data: CommentCreateRequest) => Promise<Comment>;
  update_comment: (id: string, data: CommentUpdateRequest) => Promise<Comment>;
  delete_comment: (id: string) => Promise<void>;
  like_comment: (id: string) => Promise<void>;
  load_replies: (parent_id: string, query?: Partial<CommentListQuery>) => Promise<Comment[]>;
  
  // Utilities
  refresh: () => Promise<void>;
  clear_error: () => void;
  get_comment_by_id: (id: string) => Comment | undefined;
  get_replies: (parent_id: string) => Comment[];
}

// Error types

export interface CommentError extends Error {
  code?: string;
  status?: number;
  details?: any;
}

// Real-time event types

export interface CommentEvent {
  type: 'comment_created' | 'comment_updated' | 'comment_deleted' | 'comment_liked';
  letter_id: string;
  comment_id: string;
  user_id: string;
  data?: any;
  timestamp: string;
}