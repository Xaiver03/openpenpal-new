// AI服务相关的API调用
import { apiClient } from '../api-client'

// 写作灵感相关类型
export interface WritingInspirationRequest {
  theme?: string
  style?: string
  tags?: string[]
  count?: number // default: 1, max: 5
}

export interface WritingInspiration {
  id: string
  theme: string
  prompt: string
  style: string
  tags: string[]
}

export interface WritingInspirationResponse {
  inspirations: WritingInspiration[]
}

// 每日灵感类型
export interface DailyInspiration {
  date: string
  theme: string
  prompt: string
  quote: string
  tips: string[]
}

// AI人设类型
export interface AIPersona {
  id: string
  name: string
  description: string
  avatar?: string
}

export interface AIPersonasResponse {
  personas: AIPersona[]
  total: number
}

// 延迟配置类型（与后端保持一致）
export interface DelayConfig {
  type: 'preset' | 'relative' | 'absolute'
  presetOption?: string
  relativeDays?: number
  relativeHours?: number
  relativeMinutes?: number
  absoluteTime?: Date
  timezone?: string
  userDescription?: string
}

// 笔友匹配类型
export interface PenpalMatchRequest {
  letterId: string
  max_matches?: number // default: 3
  delay_config?: DelayConfig // 新增：延迟配置
}

export interface PenpalMatch {
  userId: string
  username: string
  score: number
  reason: string
  common_tags: string[]
}

export interface PenpalMatchResponse {
  matches: PenpalMatch[]
}

// AI回信类型
export interface AIReplyRequest {
  letterId: string
  persona: string // 后端期望的字段名是 persona 而不是 persona_id
  delay_hours?: number // default: 24
}

export interface AIReplyResponse {
  reply_content: string
  persona_name: string
  estimated_delay: number
}

// AI回信角度建议类型
export interface AIReplyAdviceRequest {
  letterId: string
  persona_type: 'custom' | 'predefined' | 'deceased' | 'distant_friend' | 'unspoken_love'
  persona_name: string
  persona_desc?: string
  relationship?: string
  delivery_days?: number // 0-7 days
}

export interface AIReplyAdvice {
  id: string
  letterId: string
  userId: string
  persona_type: string
  persona_name: string
  persona_desc: string
  perspectives: string[] // 角度建议数组
  emotional_tone: string // 情感基调
  suggested_topics: string // 建议话题
  writing_style: string // 写作风格
  key_points: string // 关键要点
  delivery_delay: number
  scheduled_for?: string
  provider: string
  createdAt: string
  used_at?: string
}

// AI使用统计类型
export interface AIUsageStats {
  userId: number
  usage: {
    matches_created: number
    replies_generated: number
    inspirations_used: number
    letters_curated: number
  }
  limits: {
    daily_matches: number
    daily_replies: number
    daily_inspirations: number
    daily_curations: number
  }
  remaining: {
    matches: number
    replies: number
    inspirations: number
    curations: number
  }
}

class AIService {
  private baseUrl = '/api/v1/ai'  // Direct path to AI service

  // Helper function to convert ApiResponse to BaseApiResponse
  private convertResponse<T>(response: any): T {
    if (response.code === 0 && response.data) {
      return response.data
    }
    if (response.data) {
      return response.data
    }
    throw new Error(response.message || 'API request failed')
  }

  // 生成写作灵感
  async generateWritingPrompt(request: WritingInspirationRequest): Promise<WritingInspirationResponse> {
    const response = await apiClient.post(`${this.baseUrl}/inspiration`, request)
    return this.convertResponse<WritingInspirationResponse>(response)
  }

  // 获取每日灵感
  async getDailyInspiration(): Promise<DailyInspiration> {
    const response = await apiClient.get(`${this.baseUrl}/daily-inspiration`)
    return this.convertResponse<DailyInspiration>(response)
  }

  // 获取AI人设列表
  async getAIPersonas(): Promise<AIPersonasResponse> {
    const response = await apiClient.get(`${this.baseUrl}/personas`)
    return this.convertResponse<AIPersonasResponse>(response)
  }

  // AI笔友匹配
  async matchPenpal(request: PenpalMatchRequest): Promise<PenpalMatchResponse> {
    const response = await apiClient.post(`${this.baseUrl}/match`, request)
    return this.convertResponse<PenpalMatchResponse>(response)
  }

  // 生成AI回信
  async generateReply(request: AIReplyRequest): Promise<AIReplyResponse> {
    const response = await apiClient.post(`${this.baseUrl}/reply`, request)
    return this.convertResponse<AIReplyResponse>(response)
  }

  // 生成延迟AI回信（新方法）
  async scheduleDelayedReply(request: AIReplyRequest): Promise<{
    conversation_id: string;
    scheduled_at: string;
    delay_hours: number;
  }> {
    const response = await apiClient.post(`${this.baseUrl}/reply`, request)
    return this.convertResponse<{
      conversation_id: string;
      scheduled_at: string;
      delay_hours: number;
    }>(response)
  }

  // 生成AI回信角度建议
  async generateReplyAdvice(request: AIReplyAdviceRequest): Promise<AIReplyAdvice> {
    const response = await apiClient.post(`${this.baseUrl}/reply-advice`, request)
    return this.convertResponse<AIReplyAdvice>(response)
  }

  // 获取AI使用统计
  async getAIStats(): Promise<AIUsageStats> {
    const response = await apiClient.get(`${this.baseUrl}/stats`)
    return this.convertResponse<AIUsageStats>(response)
  }
}

export const aiService = new AIService()