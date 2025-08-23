import { apiClient } from '@/lib/api-client'

export interface OPCodeInfo {
  code: string
  school_code: string
  school_name: string
  area_code: string
  area_name: string
  building_code: string
  building_name: string
  point_code: string
  point_name: string
  full_address: string
  is_active: boolean
  privacy_level: 'public' | 'partial' | 'private'
  created_at: string
  updated_at: string
}

export interface OPCodeValidation {
  code: string
  is_valid: boolean
  format_valid: boolean
  exists: boolean
  error?: string
}

export interface OPCodeApplication {
  id: string
  user_id: string
  location_type: 'dormitory' | 'shop' | 'box' | 'club'
  location_name: string
  location_address: string
  school_code: string
  area_code: string
  building_code: string
  proof_images: string[]
  status: 'pending' | 'approved' | 'rejected'
  assigned_code?: string
  reject_reason?: string
  created_at: string
  updated_at: string
}

class OPCodeService {
  /**
   * 验证OP Code格式和有效性
   * 公开接口：任何人都可以调用
   * 如果用户已登录，会自动返回额外信息
   */
  async validateOPCode(code: string): Promise<OPCodeValidation & {
    additional_info?: any
    user_role?: string
  }> {
    try {
      const response = await apiClient.get(`/api/v1/opcode/validate?code=${encodeURIComponent(code)}`)
      return (response as any).data?.data || (response as any).data
    } catch (error) {
      console.error('Failed to validate OP Code:', error)
      throw error
    }
  }

  /**
   * 获取OP Code详细信息
   * 公开接口：返回公开信息
   * 如果用户已登录，管理员和信使可以看到私有信息
   * 信使还能看到是否有管理权限
   */
  async getOPCode(code: string): Promise<OPCodeInfo & {
    access_level?: 'basic' | 'full'
    can_manage?: boolean
  }> {
    try {
      const response = await apiClient.get(`/api/v1/opcode/${code}`)
      return (response as any).data?.data || (response as any).data
    } catch (error) {
      console.error('Failed to get OP Code:', error)
      throw error
    }
  }

  /**
   * 申请OP Code（需要认证）
   */
  async applyForOPCode(data: {
    location_type: string
    location_name: string
    location_address: string
    school_code: string
    area_code: string
    building_code: string
    proof_images: string[]
  }): Promise<OPCodeApplication> {
    try {
      const response = await apiClient.post('/api/v1/opcode/apply', data)
      return (response as any).data as OPCodeApplication
    } catch (error) {
      console.error('Failed to apply for OP Code:', error)
      throw error
    }
  }

  /**
   * 搜索OP Code（需要认证）
   */
  async searchOPCodes(params: {
    keyword?: string
    school_code?: string
    area_code?: string
    page?: number
    page_size?: number
  }) {
    try {
      const response = await apiClient.get(`/api/v1/opcode/search?${new URLSearchParams(params as any).toString()}`)
      return (response as any).data
    } catch (error) {
      console.error('Failed to search OP Codes:', error)
      throw error
    }
  }

  /**
   * 获取学校的OP Code统计信息（需要认证）
   */
  async getOPCodeStats(schoolCode: string) {
    try {
      const response = await apiClient.get(`/api/v1/opcode/stats/${schoolCode}`)
      return response.data
    } catch (error) {
      console.error('Failed to get OP Code stats:', error)
      throw error
    }
  }

  /**
   * 格式化OP Code显示
   * @param code OP Code
   * @param privacyLevel 隐私级别
   */
  formatOPCode(code: string, privacyLevel: 'full' | 'partial' | 'public' = 'full'): string {
    if (!code || code.length !== 6) return code || ''
    
    switch (privacyLevel) {
      case 'partial':
        // 隐藏最后两位
        return code.substring(0, 4) + '**'
      case 'public':
        // 只显示学校代码
        return code.substring(0, 2) + '****'
      default:
        return code
    }
  }

  /**
   * 解析OP Code
   */
  parseOPCode(code: string) {
    if (!code || code.length !== 6) return null
    
    return {
      schoolCode: code.substring(0, 2),
      areaCode: code.substring(2, 4),
      pointCode: code.substring(4, 6)
    }
  }
}

export const opcodeService = new OPCodeService()