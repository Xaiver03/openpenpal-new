// Museum type definitions - aligned with backend models

export interface MuseumEntry {
  id: string;
  letter_id: string;
  submission_id?: string;
  display_title: string;
  author_display_type: 'anonymous' | 'penName' | 'realName';
  author_display_name?: string;
  curator_type: 'system' | 'user' | 'admin';
  curator_id: string;
  categories: string[];
  tags: string[];
  status: MuseumItemStatus;
  moderation_status: MuseumItemStatus;
  view_count: number;
  like_count: number;
  bookmark_count: number;
  share_count: number;
  ai_metadata?: Record<string, any>;
  imageUrl?: string;
  created_at: string;
  updated_at: string;
  published_at?: string;
  featured_at?: string;
  
  // Relations
  letter?: any; // Letter type from letter.ts
  submission?: MuseumSubmission;
  curator?: any; // User type
}

export type MuseumItemStatus = 'pending' | 'approved' | 'rejected' | 'archived' | 'featured';

export interface MuseumSubmission {
  id: string;
  letter_id: string;
  submitter_id: string;
  display_preference: 'anonymous' | 'penName' | 'realName';
  pen_name?: string;
  submission_reason: string;
  curator_notes?: string;
  status: 'pending' | 'approved' | 'rejected' | 'withdrawn';
  submitted_at: string;
  reviewed_at?: string;
  reviewed_by?: string;
}

export interface MuseumExhibition {
  id: string;
  title: string;
  description: string;
  theme: string;
  cover_image?: string;
  start_date: string;
  end_date?: string;
  curator_id: string;
  status: 'draft' | 'active' | 'ended' | 'archived';
  entry_ids: string[];
  view_count: number;
  created_at: string;
  updated_at: string;
  
  // Relations
  entries?: MuseumEntry[];
  curator?: any; // User type
}

export interface MuseumTag {
  id: string;
  name: string;
  category: string;
  usage_count: number;
  created_at: string;
}

export interface MuseumListRequest {
  page?: number;
  limit?: number;
  category?: string;
  tags?: string[];
  search?: string;
  sort_by?: 'latest' | 'popular' | 'featured';
  status?: MuseumItemStatus;
}

export interface MuseumListResponse {
  entries: MuseumEntry[];
  total: number;
  page: number;
  limit: number;
  has_more: boolean;
}

export interface SubmissionRequest {
  letter_id: string;
  display_preference: 'anonymous' | 'penName' | 'realName';
  pen_name?: string;
  submission_reason: string;
}

export interface InteractionRequest {
  type: 'view' | 'like' | 'bookmark' | 'share';
}

export interface ReactionRequest {
  reaction_type: 'like' | 'love' | 'inspiring' | 'touching';
  comment?: string;
}

export interface ApprovalRequest {
  status: 'approved' | 'rejected';
  reason?: string;
  featured?: boolean;
}

export interface ExhibitionCreateRequest {
  title: string;
  description: string;
  theme: string;
  cover_image?: string;
  start_date: string;
  end_date?: string;
  entry_ids: string[];
}

export interface PopularMuseumEntry extends MuseumEntry {
  trend_score: number;
  recent_views: number;
  recent_likes: number;
}

export interface MuseumAnalytics {
  total_entries: number;
  total_views: number;
  total_likes: number;
  total_shares: number;
  popular_categories: Array<{
    category: string;
    count: number;
  }>;
  popular_tags: Array<{
    tag: string;
    count: number;
  }>;
  daily_stats: Array<{
    date: string;
    views: number;
    likes: number;
    submissions: number;
  }>;
}

export interface ApiError {
  code: string;
  message: string;
  details?: any;
}