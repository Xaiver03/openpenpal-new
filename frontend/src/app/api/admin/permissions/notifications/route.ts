/**
 * 权限变更实时通知API - 基于Server-Sent Events
 */

import { NextRequest, NextResponse } from 'next/server'

// 存储活跃的SSE连接
const activeConnections = new Set<ReadableStreamDefaultController>()

// 权限变更事件类型
interface PermissionChangeEvent {
  type: 'permission_updated' | 'permission_reset' | 'config_imported' | 'user_affected'
  data: {
    target: string
    targetType: 'role' | 'courier-level' | 'system'
    affectedUsers?: number
    modifiedBy: string
    timestamp: string
    changes?: {
      added: string[]
      removed: string[]
    }
  }
}

export async function GET(request: NextRequest) {
  // 设置Server-Sent Events
  const encoder = new TextEncoder()
  
  const customReadable = new ReadableStream({
    start(controller) {
      // 添加到活跃连接列表
      activeConnections.add(controller)
      
      // 发送初始连接消息
      controller.enqueue(encoder.encode(`data: ${JSON.stringify({
        type: 'connected',
        timestamp: new Date().toISOString()
      })}\n\n`))
      
      // 定期发送心跳
      const heartbeat = setInterval(() => {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify({
            type: 'heartbeat',
            timestamp: new Date().toISOString()
          })}\n\n`))
        } catch (error) {
          // 连接已关闭
          clearInterval(heartbeat)
          activeConnections.delete(controller)
        }
      }, 30000) // 30秒心跳

      // 清理函数
      request.signal.addEventListener('abort', () => {
        clearInterval(heartbeat)
        activeConnections.delete(controller)
        try {
          controller.close()
        } catch (error) {
          // 连接可能已经关闭
        }
      })
    },
    
    cancel() {
      // 连接取消时的清理
      activeConnections.delete(this as any)
    }
  })

  return new NextResponse(customReadable, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection': 'keep-alive',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Headers': 'Cache-Control'
    }
  })
}

export async function POST(request: NextRequest) {
  try {
    const event: PermissionChangeEvent = await request.json()
    
    // 广播权限变更事件给所有连接的客户端
    broadcastPermissionChange(event)
    
    return NextResponse.json({
      success: true,
      message: '权限变更通知已发送',
      activeConnections: activeConnections.size
    })
  } catch (error) {
    console.error('权限通知发送失败:', error)
    return NextResponse.json({
      success: false,
      error: '发送权限变更通知失败'
    }, { status: 500 })
  }
}

// 广播权限变更事件
function broadcastPermissionChange(event: PermissionChangeEvent) {
  const encoder = new TextEncoder()
  const message = `data: ${JSON.stringify(event)}\n\n`
  const encodedMessage = encoder.encode(message)
  
  // 向所有活跃连接发送消息
  activeConnections.forEach(controller => {
    try {
      controller.enqueue(encodedMessage)
    } catch (error) {
      // 连接已断开，从列表中移除
      activeConnections.delete(controller)
    }
  })
  
  console.log(`权限变更通知已发送给 ${activeConnections.size} 个客户端:`, event)
}

// Note: Helper functions should not be exported from route handlers
// Move this to a separate utility file if needed by other modules