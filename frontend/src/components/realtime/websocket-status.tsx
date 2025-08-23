'use client'

import React, { useState } from 'react'
import { 
  Wifi, 
  WifiOff, 
  Loader2, 
  AlertTriangle, 
  RefreshCw,
  Signal,
  Users,
  MessageSquare,
  Clock
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { 
  Tooltip, 
  TooltipContent, 
  TooltipProvider, 
  TooltipTrigger 
} from '@/components/ui/tooltip'
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuHeader, 
  DropdownMenuTrigger 
} from '@/components/ui/dropdown-menu'
import { useWebSocket, ConnectionStatus } from '@/contexts/websocket-context'
import { SafeTimestamp } from '@/components/ui/safe-timestamp'

interface WebSocketStatusProps {
  showDetails?: boolean
  className?: string
}

export function WebSocketStatus({ showDetails = false, className = '' }: WebSocketStatusProps) {
  const { 
    connectionStatus, 
    isConnected, 
    connectionInfo, 
    stats, 
    connect, 
    disconnect 
  } = useWebSocket()
  
  const [showDropdown, setShowDropdown] = useState(false)

  const getStatusInfo = (status: ConnectionStatus) => {
    switch (status) {
      case 'connected':
        return {
          icon: <Wifi className="w-4 h-4" />,
          text: '已连接',
          color: 'bg-green-500',
          textColor: 'text-green-700',
          bgColor: 'bg-green-50',
          borderColor: 'border-green-200'
        }
      case 'connecting':
        return {
          icon: <Loader2 className="w-4 h-4 animate-spin" />,
          text: '连接中',
          color: 'bg-yellow-500',
          textColor: 'text-yellow-700',
          bgColor: 'bg-yellow-50',
          borderColor: 'border-yellow-200'
        }
      case 'reconnecting':
        return {
          icon: <RefreshCw className="w-4 h-4 animate-spin" />,
          text: '重连中',
          color: 'bg-blue-500',
          textColor: 'text-blue-700',
          bgColor: 'bg-blue-50',
          borderColor: 'border-blue-200'
        }
      case 'error':
        return {
          icon: <AlertTriangle className="w-4 h-4" />,
          text: '连接错误',
          color: 'bg-red-500',
          textColor: 'text-red-700',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200'
        }
      default:
        return {
          icon: <WifiOff className="w-4 h-4" />,
          text: '未连接',
          color: 'bg-gray-500',
          textColor: 'text-gray-700',
          bgColor: 'bg-gray-50',
          borderColor: 'border-gray-200'
        }
    }
  }

  const statusInfo = getStatusInfo(connectionStatus)

  if (!showDetails) {
    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className={`flex items-center gap-2 ${className}`}>
              <div className="relative">
                <div className={`w-2 h-2 rounded-full ${statusInfo.color}`}></div>
                {connectionStatus === 'connected' && (
                  <div className={`absolute -inset-1 w-4 h-4 rounded-full ${statusInfo.color} opacity-75 animate-ping`}></div>
                )}
              </div>
              {showDetails && (
                <span className={`text-sm ${statusInfo.textColor}`}>
                  {statusInfo.text}
                </span>
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <p>实时连接状态: {statusInfo.text}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  return (
    <DropdownMenu open={showDropdown} onOpenChange={setShowDropdown}>
      <DropdownMenuTrigger asChild>
        <Button 
          variant="ghost" 
          size="sm" 
          className={`flex items-center gap-2 ${className}`}
        >
          {statusInfo.icon}
          <span className="text-sm">{statusInfo.text}</span>
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-80">
        <DropdownMenuHeader className="px-4 py-2 border-b">
          <h3 className="font-semibold">WebSocket 连接状态</h3>
        </DropdownMenuHeader>

        <div className="p-4 space-y-4">
          {/* 连接状态 */}
          <Card className={`${statusInfo.bgColor} ${statusInfo.borderColor}`}>
            <CardContent className="p-3">
              <div className="flex items-center gap-3">
                {statusInfo.icon}
                <div>
                  <h4 className={`font-medium ${statusInfo.textColor}`}>
                    {statusInfo.text}
                  </h4>
                  <p className="text-xs text-gray-600">
                    {connectionStatus === 'connected' && connectionInfo ? (
                      <>
                        连接于 <SafeTimestamp 
                          date={connectionInfo.connected_at} 
                          format="relative" 
                          fallback="刚刚"
                          className="inline"
                        />
                      </>
                    ) : (
                      '实时通信功能'
                    )}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* 连接信息 */}
          {connectionInfo && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium text-gray-900">连接信息</h4>
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <span className="text-gray-500">连接ID:</span>
                  <p className="font-mono">{connectionInfo.id?.slice(-8)}</p>
                </div>
                <div>
                  <span className="text-gray-500">用户角色:</span>
                  <p className="capitalize">{connectionInfo.role}</p>
                </div>
                <div>
                  <span className="text-gray-500">学校代码:</span>
                  <p>{connectionInfo.schoolCode}</p>
                </div>
                <div>
                  <span className="text-gray-500">房间数:</span>
                  <p>{connectionInfo.rooms?.length || 0}</p>
                </div>
              </div>
            </div>
          )}

          {/* 统计信息 */}
          {stats && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium text-gray-900">实时统计</h4>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex items-center gap-2">
                  <Users className="w-4 h-4 text-blue-500" />
                  <div>
                    <p className="text-sm font-medium">{stats.active_connections}</p>
                    <p className="text-xs text-gray-500">在线连接</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <MessageSquare className="w-4 h-4 text-green-500" />
                  <div>
                    <p className="text-sm font-medium">{stats.total_messages}</p>
                    <p className="text-xs text-gray-500">总消息数</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* 操作按钮 */}
          <div className="flex gap-2">
            {connectionStatus === 'connected' ? (
              <Button
                variant="outline"
                size="sm"
                onClick={disconnect}
                className="flex-1"
              >
                断开连接
              </Button>
            ) : (
              <Button
                variant="default"
                size="sm"
                onClick={connect}
                className="flex-1"
                disabled={connectionStatus === 'connecting'}
              >
                {connectionStatus === 'connecting' ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    连接中...
                  </>
                ) : (
                  '重新连接'
                )}
              </Button>
            )}
          </div>
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

// 简单的状态指示器
export function SimpleWebSocketIndicator({ className = '' }: { className?: string }) {
  const { connectionStatus } = useWebSocket()
  const statusInfo = getStatusInfo(connectionStatus)

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className={`flex items-center gap-2 ${className}`}>
            <div className="relative">
              <div className={`w-2 h-2 rounded-full ${statusInfo.color}`}></div>
              {connectionStatus === 'connected' && (
                <div className={`absolute -inset-1 w-4 h-4 rounded-full ${statusInfo.color} opacity-75 animate-ping`}></div>
              )}
            </div>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>实时连接: {statusInfo.text}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

// 连接质量指示器
export function ConnectionQualityIndicator({ className = '' }: { className?: string }) {
  const { connectionStatus, stats } = useWebSocket()
  
  const getQualityLevel = () => {
    // 这里可以根据实际的延迟、重连次数等指标来计算连接质量
    switch (connectionStatus) {
      case 'connected': return 3
      case 'reconnecting': return 1
      case 'connecting': return 2
      default: return 0
    }
  }

  const qualityLevel = getQualityLevel()

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className={`flex items-center gap-1 ${className}`}>
            <Signal className="w-4 h-4 text-gray-400" />
            <div className="flex gap-0.5">
              {[1, 2, 3].map((level) => (
                <div
                  key={level}
                  className={`w-1 h-3 rounded-sm ${
                    level <= qualityLevel
                      ? qualityLevel === 3
                        ? 'bg-green-500'
                        : qualityLevel === 2
                        ? 'bg-yellow-500'
                        : 'bg-red-500'
                      : 'bg-gray-300'
                  }`}
                />
              ))}
            </div>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>
            连接质量: {
              qualityLevel === 3 ? '优秀' : 
              qualityLevel === 2 ? '良好' : 
              qualityLevel === 1 ? '一般' : '断开'
            }
          </p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

function getStatusInfo(status: ConnectionStatus) {
  switch (status) {
    case 'connected':
      return {
        icon: <Wifi className="w-4 h-4" />,
        text: '已连接',
        color: 'bg-green-500',
        textColor: 'text-green-700',
        bgColor: 'bg-green-50',
        borderColor: 'border-green-200'
      }
    case 'connecting':
      return {
        icon: <Loader2 className="w-4 h-4 animate-spin" />,
        text: '连接中',
        color: 'bg-yellow-500',
        textColor: 'text-yellow-700',
        bgColor: 'bg-yellow-50',
        borderColor: 'border-yellow-200'
      }
    case 'reconnecting':
      return {
        icon: <RefreshCw className="w-4 h-4 animate-spin" />,
        text: '重连中',
        color: 'bg-blue-500',
        textColor: 'text-blue-700',
        bgColor: 'bg-blue-50',
        borderColor: 'border-blue-200'
      }
    case 'error':
      return {
        icon: <AlertTriangle className="w-4 h-4" />,
        text: '连接错误',
        color: 'bg-red-500',
        textColor: 'text-red-700',
        bgColor: 'bg-red-50',
        borderColor: 'border-red-200'
      }
    default:
      return {
        icon: <WifiOff className="w-4 h-4" />,
        text: '未连接',
        color: 'bg-gray-500',
        textColor: 'text-gray-700',
        bgColor: 'bg-gray-50',
        borderColor: 'border-gray-200'
      }
  }
}