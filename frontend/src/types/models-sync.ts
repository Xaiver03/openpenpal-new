// Auto-generated TypeScript interfaces from Go models
// Generated on: 2025-08-05T11:49:47.369Z
// DO NOT EDIT MANUALLY - Use sync-models.js to regenerate
// Note: Field names use camelCase due to backend transformation middleware - Use sync-models.js to regenerate

// Import types from other modules
import { UserRole } from '@/constants/roles'

// Enum types
export type LetterStyle = 'classic' | 'modern' | 'elegant' | 'casual'
export type LetterStatus = 'draft' | 'pending' | 'published' | 'delivered' | 'read' | 'replied' | 'archived'
export type LetterVisibility = 'private' | 'school' | 'public'
export type AIProvider = 'openai' | 'moonshot' | 'baidu' | 'custom'
export type MuseumSourceType = 'letter' | 'submission' | 'imported'
export type MuseumItemStatus = 'pending' | 'approved' | 'rejected' | 'archived'

// Additional interfaces for referenced types
export interface LetterCode {
  id: string
  letter_id: string
  code: string
  qr_code_url?: string
  is_active: boolean
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface StatusLog {
  id: string
  letter_id: string
  status: LetterStatus
  message?: string
  created_at: string
}

export interface LetterPhoto {
  id: string
  letter_id: string
  url: string
  caption?: string
  order: number
  created_at: string
}

export interface Envelope {
  id: string
  letter_id: string
  design_id: string
  customizations?: string
  created_at: string
  updated_at: string
}

export interface LetterLike {
  id: string
  letter_id: string
  user_id: string
  created_at: string
}

export interface LetterShare {
  id: string
  letter_id: string
  user_id: string
  platform: string
  shared_at: string
}

// GORM DeletedAt type
export interface DeletedAt {
  time: string
  valid: boolean
}

export interface User {
  id: string;
  username: string;
  email: string;
  nickname: string;
  avatar: string;
  role: UserRole;
  school_code: string;
  op_code: string; // OP Code地址
  is_active: boolean;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
  sent_letters?: Letter[];
  authored_letters?: Letter[];
}

export interface Letter {
  id: string;
  user_id: string;
  author_id: string;
  title: string;
  content: string;
  style: LetterStyle;
  status: LetterStatus;
  visibility: LetterVisibility;
  like_count: number;
  recipient_op_code: string;
  sender_op_code: string;
  share_count: number;
  view_count: number;
  reply_to?: string;
  envelope_id?: string | null;
  created_at: string;
  updated_at: string;
  user?: User | null;
  author?: User | null;
  code?: LetterCode | null;
  status_logs?: StatusLog[];
  photos?: LetterPhoto[];
  envelope?: Envelope | null;
  likes?: LetterLike[];
  shares?: LetterShare[];
}

export interface Courier {
  id: string;
  user_id: string;
  user: User;
  name: string;
  contact: string;
  school: string;
  zone: string;
  managed_op_code_prefix: string;
  has_printer: boolean;
  self_intro: string;
  can_mentor: string;
  weekly_hours: number;
  max_daily_tasks: number;
  transport_method: string;
  time_slots: string;
  status: string;
  level: number;
  task_count: number;
  points: number;
  created_at: string;
  updated_at: string;
  deleted_at: DeletedAt;
}

export interface AIConfig {
  id: string;
  provider: AIProvider;
  api_endpoint: string;
  model: string;
  temperature: number;
  max_tokens: number;
  is_active: boolean;
  priority: number;
  daily_quota: number;
  used_quota: number;
  quota_reset_at: string;
  created_at: string;
  updated_at: string;
}

export interface MuseumItem {
  id: string;
  source_type: MuseumSourceType;
  source_id: string;
  title: string;
  description: string;
  tags: string;
  status: MuseumItemStatus;
  submitted_by: string;
  approved_by: string | null;
  approved_at: string | null;
  view_count: number;
  like_count: number;
  share_count: number;
  origin_op_code?: string;
  created_at: string;
  updated_at: string;
  letter?: Letter | null;
  submitted_by_user?: User | null;
  approved_by_user?: User | null;
}

export interface MuseumEntry {
  id: string;
  letter_id: string;
  submission_id: string | null;
  display_title: string;
  author_display_type: string;
  author_display_name: string | null;
  curator_type: string;
  curator_id: string;
  categories: string[];
  tags: string[];
  status: MuseumItemStatus;
  moderation_status: MuseumItemStatus;
  view_count: number;
  like_count: number;
  bookmark_count: number;
  share_count: number;
}

