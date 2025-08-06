/**
 * 管理员审计日志API
 * Admin Audit Logs API for OpenPenPal
 */

import { NextRequest, NextResponse } from 'next/server'
import { withSuperAdminAuth } from '@/lib/middleware/admin'
import { AuditLogAPI } from '@/lib/middleware/admin'
import { AuthenticatedRequest } from '@/lib/middleware/auth'

// GET - 获取审计日志（需要超级管理员权限）
async function getAuditLogs(request: AuthenticatedRequest) {
  try {
    const user = request.user!
    console.log(`超级管理员 ${user.username} 正在查看审计日志`)
    
    return await AuditLogAPI.getLogs(request)
  } catch (error) {
    console.error('获取审计日志失败:', error)
    return NextResponse.json({
      code: 500,
      message: '获取审计日志失败',
      data: null
    }, { status: 500 })
  }
}

// POST - 清理审计日志（需要超级管理员权限）
async function cleanupAuditLogs(request: AuthenticatedRequest) {
  try {
    const user = request.user!
    console.log(`超级管理员 ${user.username} 正在清理审计日志`)
    
    return await AuditLogAPI.cleanupLogs(request)
  } catch (error) {
    console.error('清理审计日志失败:', error)
    return NextResponse.json({
      code: 500,
      message: '清理审计日志失败',
      data: null
    }, { status: 500 })
  }
}

export const GET = withSuperAdminAuth({
  action: 'VIEW_AUDIT_LOGS',
  resource: '/api/admin/audit-logs'
})(getAuditLogs)

export const POST = withSuperAdminAuth({
  action: 'CLEANUP_AUDIT_LOGS',
  resource: '/api/admin/audit-logs'
})(cleanupAuditLogs)