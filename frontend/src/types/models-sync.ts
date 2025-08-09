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
  letterId: string
  code: string
  qrCodeUrl?: string
  isActive: boolean
  expiresAt?: string
  createdAt: string
  updatedAt: string
}

export interface StatusLog {
  id: string
  letterId: string
  status: LetterStatus
  message?: string
  createdAt: string
}

export interface LetterPhoto {
  id: string
  letterId: string
  url: string
  caption?: string
  order: number
  createdAt: string
}

export interface Envelope {
  id: string
  letterId: string
  designId: string
  customizations?: string
  createdAt: string
  updatedAt: string
}

export interface LetterLike {
  id: string
  letterId: string
  userId: string
  createdAt: string
}

export interface LetterShare {
  id: string
  letterId: string
  userId: string
  platform: string
  sharedAt: string
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
  schoolCode: string;
  isActive: boolean;
  lastLoginAt: string | null;
  createdAt: string;
  updatedAt: string;
  sentLetters?: Letter[];
  authoredLetters?: Letter[];
}

export interface Letter {
  id: string;
  userId: string;
  authorId: string;
  title: string;
  content: string;
  style: LetterStyle;
  status: LetterStatus;
  visibility: LetterVisibility;
  likeCount: number;
  recipientOpCode: string;
  senderOpCode: string;
  shareCount: number;
  viewCount: number;
  replyTo?: string;
  envelopeId?: string | null;
  createdAt: string;
  updatedAt: string;
  user?: User | null;
  author?: User | null;
  code?: LetterCode | null;
  statusLogs?: StatusLog[];
  photos?: LetterPhoto[];
  envelope?: Envelope | null;
  likes?: LetterLike[];
  shares?: LetterShare[];
}

export interface Courier {
  id: string;
  userId: string;
  user: User;
  name: string;
  contact: string;
  school: string;
  zone: string;
  managedOpCodePrefix: string;
  hasPrinter: boolean;
  selfIntro: string;
  canMentor: string;
  weeklyHours: number;
  maxDailyTasks: number;
  transportMethod: string;
  timeSlots: string;
  status: string;
  level: number;
  taskCount: number;
  points: number;
  createdAt: string;
  updatedAt: string;
  deletedAt: DeletedAt;
}

export interface AIConfig {
  id: string;
  provider: AIProvider;
  apiEndpoint: string;
  model: string;
  temperature: number;
  maxTokens: number;
  isActive: boolean;
  priority: number;
  dailyQuota: number;
  usedQuota: number;
  quotaResetAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface MuseumItem {
  id: string;
  sourceType: MuseumSourceType;
  sourceId: string;
  title: string;
  description: string;
  tags: string;
  status: MuseumItemStatus;
  submittedBy: string;
  approvedBy: string | null;
  approvedAt: string | null;
  viewCount: number;
  likeCount: number;
  shareCount: number;
  originOpCode?: string;
  createdAt: string;
  updatedAt: string;
  letter?: Letter | null;
  submittedByUser?: User | null;
  approvedByUser?: User | null;
}

export interface MuseumEntry {
  id: string;
  letterId: string;
  submissionId: string | null;
  displayTitle: string;
  authorDisplayType: string;
  authorDisplayName: string | null;
  curatorType: string;
  curatorId: string;
  categories: string[];
  tags: string[];
  status: MuseumItemStatus;
  moderationStatus: MuseumItemStatus;
  viewCount: number;
  likeCount: number;
  bookmarkCount: number;
  shareCount: number;
}

