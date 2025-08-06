import { NextRequest, NextResponse } from 'next/server'
import { withSuperAdminAuth } from '@/lib/middleware/admin'
import { AuthenticatedRequest } from '@/lib/middleware/auth'

// 模拟系统配置存储
declare global {
  var systemConfig: any | undefined
}

const DEFAULT_CONFIG = {
  site_name: 'OpenPenPal',
  site_description: '温暖的校园信件投递平台',
  site_logo: '',
  maintenance_mode: false,
  
  smtp_host: 'smtp.gmail.com',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  smtp_encryption: 'tls',
  email_from_name: 'OpenPenPal',
  email_from_address: 'noreply@openpenpal.com',
  
  max_letter_length: 5000,
  allowed_file_types: ['jpg', 'png', 'pdf'],
  max_file_size: 10,
  letter_review_required: false,
  auto_delivery_enabled: true,
  
  user_registration_enabled: true,
  email_verification_required: true,
  max_users_per_school: 10000,
  user_inactive_days: 90,
  
  courier_application_enabled: true,
  courier_auto_approval: false,
  max_delivery_distance: 10,
  courier_rating_required: true,
  
  password_min_length: 6,
  password_require_symbols: false,
  password_require_numbers: true,
  session_timeout: 3600,
  max_login_attempts: 5,
  
  email_notifications: true,
  sms_notifications: false,
  push_notifications: true,
  admin_notifications: true
}

if (!global.systemConfig) {
  global.systemConfig = { ...DEFAULT_CONFIG }
}

// GET - 获取系统配置（需要超级管理员权限）
async function getSystemConfig(request: AuthenticatedRequest) {
  try {
    const user = request.user!
    console.log(`管理员 ${user.username} 正在查看系统配置`)
    
    return NextResponse.json({
      code: 0,
      message: 'success',
      data: global.systemConfig
    })
  } catch (error) {
    console.error('获取系统配置失败:', error)
    return NextResponse.json({
      code: 500,
      message: '获取系统配置失败',
      data: null
    }, { status: 500 })
  }
}

export const GET = withSuperAdminAuth({
  action: 'VIEW_SYSTEM_CONFIG',
  resource: '/api/admin/settings'
})(getSystemConfig)

// PUT - 更新系统配置（需要超级管理员权限）
async function updateSystemConfig(request: AuthenticatedRequest) {
  try {
    const user = request.user!
    const body = await request.json()
    
    console.log(`管理员 ${user.username} 正在更新系统配置:`, body)
    
    // 验证必填字段
    const requiredFields = ['site_name', 'email_from_address']
    for (const field of requiredFields) {
      if (!body[field]) {
        return NextResponse.json({
          code: 400,
          message: `字段 ${field} 不能为空`,
          data: null
        }, { status: 400 })
      }
    }
    
    // 记录变更前的配置（用于审计）
    const previousConfig = { ...global.systemConfig }
    
    // 更新配置
    global.systemConfig = {
      ...global.systemConfig,
      ...body,
      updatedAt: new Date().toISOString(),
      updated_by: user.username
    }
    
    console.log(`系统配置已由 ${user.username} 更新:`, global.systemConfig)
    
    return NextResponse.json({
      code: 0,
      message: '系统配置更新成功',
      data: global.systemConfig
    })
  } catch (error) {
    console.error('更新系统配置失败:', error)
    return NextResponse.json({
      code: 500,
      message: '更新系统配置失败',
      data: null
    }, { status: 500 })
  }
}

export const PUT = withSuperAdminAuth({
  action: 'UPDATE_SYSTEM_CONFIG',
  resource: '/api/admin/settings'
})(updateSystemConfig)

// POST - 重置为默认配置（需要超级管理员权限）
async function resetSystemConfig(request: AuthenticatedRequest) {
  try {
    const user = request.user!
    
    console.log(`管理员 ${user.username} 正在重置系统配置`)
    
    global.systemConfig = {
      ...DEFAULT_CONFIG,
      reset_at: new Date().toISOString(),
      reset_by: user.username
    }
    
    console.log(`系统配置已由 ${user.username} 重置为默认值`)
    
    return NextResponse.json({
      code: 0,
      message: '系统配置已重置为默认值',
      data: global.systemConfig
    })
  } catch (error) {
    console.error('重置系统配置失败:', error)
    return NextResponse.json({
      code: 500,
      message: '重置系统配置失败',
      data: null
    }, { status: 500 })
  }
}

export const POST = withSuperAdminAuth({
  action: 'RESET_SYSTEM_CONFIG',
  resource: '/api/admin/settings'
})(resetSystemConfig)