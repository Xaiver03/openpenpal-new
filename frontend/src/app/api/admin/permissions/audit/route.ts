/**
 * 权限变更审计日志API
 */

import { NextRequest, NextResponse } from 'next/server'

// 模拟的审计日志存储（实际应用中应该使用数据库）
let auditLogs: PermissionAuditLog[] = []

interface PermissionAuditLog {
  id: string
  type: 'permission_updated' | 'permission_reset' | 'config_imported' | 'config_exported' | 'system_reset'
  target: string
  targetType: 'role' | 'courier-level' | 'system'
  modifiedBy: string
  timestamp: string
  changes?: {
    added: string[]
    removed: string[]
  }
  metadata?: {
    userAgent?: string
    ip?: string
    sessionId?: string
  }
}

// ================================
// 获取审计日志
// ================================

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const pageSize = parseInt(searchParams.get('pageSize') || '20')
    const target = searchParams.get('target')
    const type = searchParams.get('type')
    const modifiedBy = searchParams.get('modifiedBy')
    const startDate = searchParams.get('startDate')
    const endDate = searchParams.get('endDate')

    // 过滤日志
    let filteredLogs = [...auditLogs]

    if (target) {
      filteredLogs = filteredLogs.filter(log => 
        log.target.toLowerCase().includes(target.toLowerCase())
      )
    }

    if (type) {
      filteredLogs = filteredLogs.filter(log => log.type === type)
    }

    if (modifiedBy) {
      filteredLogs = filteredLogs.filter(log => 
        log.modifiedBy.toLowerCase().includes(modifiedBy.toLowerCase())
      )
    }

    if (startDate) {
      filteredLogs = filteredLogs.filter(log => 
        new Date(log.timestamp) >= new Date(startDate)
      )
    }

    if (endDate) {
      filteredLogs = filteredLogs.filter(log => 
        new Date(log.timestamp) <= new Date(endDate)
      )
    }

    // 按时间倒序排序
    filteredLogs.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())

    // 分页
    const startIndex = (page - 1) * pageSize
    const endIndex = startIndex + pageSize
    const paginatedLogs = filteredLogs.slice(startIndex, endIndex)

    // 统计信息
    const stats = {
      total: filteredLogs.length,
      byType: filteredLogs.reduce((acc, log) => {
        acc[log.type] = (acc[log.type] || 0) + 1
        return acc
      }, {} as Record<string, number>),
      byTarget: filteredLogs.reduce((acc, log) => {
        acc[log.targetType] = (acc[log.targetType] || 0) + 1
        return acc
      }, {} as Record<string, number>),
      recentActivity: filteredLogs.slice(0, 10).map(log => ({
        timestamp: log.timestamp,
        action: `${log.type} - ${log.target}`,
        modifiedBy: log.modifiedBy
      }))
    }

    return NextResponse.json({
      success: true,
      data: {
        logs: paginatedLogs,
        pagination: {
          page,
          pageSize,
          total: filteredLogs.length,
          totalPages: Math.ceil(filteredLogs.length / pageSize)
        },
        stats
      }
    })
  } catch (error) {
    console.error('Audit API GET error:', error)
    return NextResponse.json({
      success: false,
      error: '获取审计日志失败'
    }, { status: 500 })
  }
}

// ================================
// 创建审计日志
// ================================

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { type, target, targetType, modifiedBy, changes, metadata } = body

    if (!type || !target || !targetType || !modifiedBy) {
      return NextResponse.json({
        success: false,
        error: '缺少必要的审计信息'
      }, { status: 400 })
    }

    const auditLog: PermissionAuditLog = {
      id: `audit_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type,
      target,
      targetType,
      modifiedBy,
      timestamp: new Date().toISOString(),
      changes,
      metadata: {
        userAgent: request.headers.get('user-agent') || undefined,
        ip: request.headers.get('x-forwarded-for') || request.headers.get('x-real-ip') || undefined,
        ...metadata
      }
    }

    // 添加到日志列表（最多保留1000条）
    auditLogs.unshift(auditLog)
    if (auditLogs.length > 1000) {
      auditLogs = auditLogs.slice(0, 1000)
    }

    console.log('权限审计日志已记录:', auditLog)

    return NextResponse.json({
      success: true,
      data: auditLog
    })
  } catch (error) {
    console.error('Audit API POST error:', error)
    return NextResponse.json({
      success: false,
      error: '记录审计日志失败'
    }, { status: 500 })
  }
}

// ================================
// 导出审计日志
// ================================

export async function PUT(request: NextRequest) {
  try {
    const body = await request.json()
    const { action, filters, format = 'json' } = body

    if (action !== 'export') {
      return NextResponse.json({
        success: false,
        error: '无效的操作'
      }, { status: 400 })
    }

    // 应用过滤器
    let logsToExport = [...auditLogs]
    
    if (filters) {
      if (filters.startDate) {
        logsToExport = logsToExport.filter(log => 
          new Date(log.timestamp) >= new Date(filters.startDate)
        )
      }
      
      if (filters.endDate) {
        logsToExport = logsToExport.filter(log => 
          new Date(log.timestamp) <= new Date(filters.endDate)
        )
      }
      
      if (filters.type) {
        logsToExport = logsToExport.filter(log => log.type === filters.type)
      }
    }

    // 排序
    logsToExport.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())

    let exportData: string
    let contentType: string
    let filename: string

    switch (format) {
      case 'csv':
        const csvHeaders = 'ID,Type,Target,Target Type,Modified By,Timestamp,Changes Added,Changes Removed\n'
        const csvRows = logsToExport.map(log => {
          const addedPermissions = log.changes?.added?.join(';') || ''
          const removedPermissions = log.changes?.removed?.join(';') || ''
          return `"${log.id}","${log.type}","${log.target}","${log.targetType}","${log.modifiedBy}","${log.timestamp}","${addedPermissions}","${removedPermissions}"`
        }).join('\n')
        exportData = csvHeaders + csvRows
        contentType = 'text/csv'
        filename = `permission-audit-${new Date().toISOString().slice(0, 10)}.csv`
        break

      case 'json':
      default:
        exportData = JSON.stringify({
          exportedAt: new Date().toISOString(),
          totalRecords: logsToExport.length,
          filters,
          logs: logsToExport
        }, null, 2)
        contentType = 'application/json'
        filename = `permission-audit-${new Date().toISOString().slice(0, 10)}.json`
        break
    }

    return NextResponse.json({
      success: true,
      data: {
        content: exportData,
        contentType,
        filename,
        recordCount: logsToExport.length
      }
    })
  } catch (error) {
    console.error('Audit export error:', error)
    return NextResponse.json({
      success: false,
      error: '导出审计日志失败'
    }, { status: 500 })
  }
}

// ================================
// 清理审计日志
// ================================

export async function DELETE(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const action = searchParams.get('action')
    const beforeDate = searchParams.get('beforeDate')
    const modifiedBy = searchParams.get('modifiedBy')

    if (!modifiedBy) {
      return NextResponse.json({
        success: false,
        error: '缺少操作者信息'
      }, { status: 400 })
    }

    let deletedCount = 0

    switch (action) {
      case 'cleanup':
        if (!beforeDate) {
          return NextResponse.json({
            success: false,
            error: '清理操作需要指定截止日期'
          }, { status: 400 })
        }
        
        const originalLength = auditLogs.length
        auditLogs = auditLogs.filter(log => 
          new Date(log.timestamp) >= new Date(beforeDate)
        )
        deletedCount = originalLength - auditLogs.length
        break

      case 'clear':
        deletedCount = auditLogs.length
        auditLogs = []
        break

      default:
        return NextResponse.json({
          success: false,
          error: '无效的清理操作'
        }, { status: 400 })
    }

    // 记录清理操作的审计日志
    const cleanupLog: PermissionAuditLog = {
      id: `audit_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type: 'system_reset',
      target: 'audit_logs',
      targetType: 'system',
      modifiedBy,
      timestamp: new Date().toISOString(),
      metadata: {
        deletedCount,
        beforeDate
      } as any
    }
    auditLogs.unshift(cleanupLog)

    return NextResponse.json({
      success: true,
      message: `已清理 ${deletedCount} 条审计日志`,
      data: {
        deletedCount,
        remainingCount: auditLogs.length
      }
    })
  } catch (error) {
    console.error('Audit cleanup error:', error)
    return NextResponse.json({
      success: false,
      error: '清理审计日志失败'
    }, { status: 500 })
  }
}

// Note: Helper functions should not be exported from route handlers
// This function should be moved to a separate utility file if needed elsewhere