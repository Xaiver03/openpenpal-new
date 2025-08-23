'use client'

import { useState, useRef, useEffect, useCallback, memo } from 'react'
import { useAuth, usePermissions } from '@/stores/user-store'
import useUserStore from '@/stores/user-store'
import { useUserBasicInfo, useUserRoleInfo } from '@/hooks/use-optimized-subscriptions'
import { ChevronUp, ChevronDown, X, Bug, Minimize2, Maximize2 } from 'lucide-react'
import { log } from '@/utils/logger'

const UserDebugPanel = memo(function UserDebugPanel() {
  // Get complete user info from store
  const { user, isAuthenticated } = useUserStore()
  const { hasRole } = usePermissions()
  
  // Compatibility aliases
  const { username, nickname, role } = user || { username: undefined, nickname: undefined, role: undefined }
  const canAccessAdmin = user ? hasRole('platform_admin') || hasRole('super_admin') : false
  const isCourier = user ? (
    hasRole('courier_level1') || 
    hasRole('courier_level2') || 
    hasRole('courier_level3') || 
    hasRole('courier_level4')
  ) : false
  
  const [isExpanded, setIsExpanded] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)
  const [isDragging, setIsDragging] = useState(false)
  // 响应式默认位置：桌面端右下角，移动端右上角
  const [position, setPosition] = useState(() => {
    if (typeof window !== 'undefined') {
      const isMobile = window.innerWidth < 768
      return isMobile ? { x: 16, y: 80 } : { x: 16, y: 16 }
    }
    return { x: 16, y: 16 }
  })
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 })
  const panelRef = useRef<HTMLDivElement>(null)

  // 只在开发环境且用户已登录时显示
  if (process.env.NODE_ENV !== 'development' || !isAuthenticated) {
    return null
  }

  // 调试信息 - 只在开发环境输出
  log.dev('Debug Panel Data', {
    user,
    isAuthenticated,
    userRole: role,
    isCourier,
    canAccessAdmin,
    hasRoleCourier: isCourier,
    hasRolePlatformAdmin: hasRole('platform_admin')
  }, 'UserDebugPanel')

  // 处理拖拽开始 - 使用 useCallback 优化性能
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (!panelRef.current) return
    
    setIsDragging(true)
    const rect = panelRef.current.getBoundingClientRect()
    setDragOffset({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top
    })
    e.preventDefault()
  }, [])

  // 处理拖拽移动
  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging) return
      
      const newX = e.clientX - dragOffset.x
      const newY = e.clientY - dragOffset.y
      
      // 限制在视窗范围内
      if (typeof window === 'undefined') return
      const maxX = window.innerWidth - (panelRef.current?.offsetWidth || 0)
      const maxY = window.innerHeight - (panelRef.current?.offsetHeight || 0)
      
      setPosition({
        x: Math.max(0, Math.min(newX, maxX)),
        y: Math.max(0, Math.min(newY, maxY))
      })
    }

    const handleMouseUp = () => {
      setIsDragging(false)
    }

    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, dragOffset])

  // 响应式位置调整
  useEffect(() => {
    const handleResize = () => {
      if (!panelRef.current || typeof window === 'undefined') return
      
      const isMobile = window.innerWidth < 768
      const maxX = window.innerWidth - panelRef.current.offsetWidth
      const maxY = window.innerHeight - panelRef.current.offsetHeight
      
      setPosition(prev => {
        // 移动端避免与头部导航栏重叠
        const newY = isMobile ? Math.max(80, Math.min(prev.y, maxY)) : Math.max(0, Math.min(prev.y, maxY))
        return {
          x: Math.max(0, Math.min(prev.x, maxX)),
          y: newY
        }
      })
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  // 最小化状态
  if (isMinimized) {
    return (
      <div
        ref={panelRef}
        className="fixed bg-black/80 backdrop-blur-sm text-white rounded-full p-2 shadow-lg cursor-move touch-manipulation select-none z-[9999] hover:bg-black/90 transition-colors
                   sm:p-2 p-1.5 sm:w-auto w-12 sm:h-auto h-12"
        style={{ 
          right: `${position.x}px`, 
          bottom: `${position.y}px`,
          transform: isDragging ? 'scale(1.05)' : 'scale(1)'
        }}
        onMouseDown={handleMouseDown}
        onClick={() => setIsMinimized(false)}
      >
        <Bug className="w-4 h-4 sm:w-5 sm:h-5" />
      </div>
    )
  }

  return (
    <div
      ref={panelRef}
      className={`fixed bg-black/90 backdrop-blur-sm text-white rounded-lg shadow-2xl border border-gray-700 z-[9999] transition-all duration-200 select-none ${
        isDragging ? 'scale-105 shadow-3xl' : 'scale-100'
      } ${
        isExpanded 
          ? 'max-w-xs sm:max-w-sm md:max-w-md' 
          : 'max-w-[280px] sm:max-w-xs'
      }
      max-h-[calc(100vh-160px)] sm:max-h-[80vh] overflow-y-auto`}
      style={{ 
        right: `${position.x}px`, 
        bottom: `${position.y}px`,
        maxHeight: isExpanded ? '70vh' : 'auto'
      }}
    >
      {/* 标题栏 - 可拖拽 */}
      <div 
        className="flex items-center justify-between p-2 sm:p-3 bg-gray-800/50 rounded-t-lg cursor-move touch-manipulation"
        onMouseDown={handleMouseDown}
      >
        <div className="flex items-center space-x-1 sm:space-x-2">
          <Bug className="w-3 h-3 sm:w-4 sm:h-4 text-green-400" />
          <span className="font-bold text-xs sm:text-sm">Debug Panel</span>
        </div>
        
        <div className="flex items-center space-x-1">
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-1 hover:bg-gray-700 rounded transition-colors touch-manipulation"
            title={isExpanded ? "收起详情" : "展开详情"}
          >
            {isExpanded ? (
              <ChevronDown className="w-3 h-3 sm:w-4 sm:h-4" />
            ) : (
              <ChevronUp className="w-3 h-3 sm:w-4 sm:h-4" />
            )}
          </button>
          
          <button
            onClick={() => setIsMinimized(true)}
            className="p-1 hover:bg-gray-700 rounded transition-colors touch-manipulation"
            title="最小化"
          >
            <Minimize2 className="w-3 h-3 sm:w-4 sm:h-4" />
          </button>
        </div>
      </div>

      {/* 内容区域 */}
      <div className={`p-2 sm:p-3 space-y-1 sm:space-y-2 text-xs sm:text-sm ${
        isExpanded ? 'max-h-[calc(100vh-240px)] sm:max-h-[60vh] overflow-y-auto' : ''
      }`}>
        {/* 基础信息 - 始终显示 */}
        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <span>认证状态:</span>
            <span className={isAuthenticated ? 'text-green-400' : 'text-red-400'}>
              {isAuthenticated ? '✅ 已登录' : '❌ 未登录'}
            </span>
          </div>
          
          <div className="flex items-center justify-between">
            <span>用户名:</span>
            <span className="text-blue-300 truncate max-w-[120px]">
              {user?.username || 'null'}
            </span>
          </div>
          
          <div className="flex items-center justify-between">
            <span>角色:</span>
            <span className="text-yellow-300 truncate max-w-[120px]">
              {user?.role || 'null'}
            </span>
          </div>
        </div>

        {/* 扩展信息 - 展开时显示 */}
        {isExpanded && (
          <>
            <div className="border-t border-gray-700 pt-1 sm:pt-2 space-y-1">
              <div className="text-gray-400 font-medium text-xs">详细信息:</div>
              
              <div className="flex items-center justify-between">
                <span>用户ID:</span>
                <span className="text-gray-300 truncate max-w-[150px] font-mono text-xs">
                  {user?.id || 'null'}
                </span>
              </div>
              
              <div className="flex items-center justify-between">
                <span>学校代码:</span>
                <span className="text-gray-300 truncate max-w-[120px]">
                  {user?.school_code || 'null'}
                </span>
              </div>
              
              <div className="flex items-center justify-between">
                <span>学校名称:</span>
                <span className="text-gray-300 truncate max-w-[120px]">
                  {user?.school_name || 'null'}
                </span>
              </div>
            </div>

            <div className="border-t border-gray-700 pt-1 sm:pt-2 space-y-1">
              <div className="text-gray-400 font-medium text-xs">权限检查:</div>
              
              <div className="grid grid-cols-2 gap-1 text-xs">
                <div className="flex items-center justify-between">
                  <span>信使:</span>
                  <span className={isCourier ? 'text-green-400' : 'text-red-400'}>
                    {isCourier ? '✅' : '❌'}
                  </span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span>管理员:</span>
                  <span className={canAccessAdmin ? 'text-green-400' : 'text-red-400'}>
                    {canAccessAdmin ? '✅' : '❌'}
                  </span>
                </div>
              </div>
              
              <div className="space-y-0.5 text-xs">
                <div className="flex items-center justify-between">
                  <span>信使等级1:</span>
                  <span className={hasRole('courier_level1') ? 'text-green-400' : 'text-red-400'}>
                    {hasRole('courier_level1') ? '✅' : '❌'}
                  </span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span>信使等级2:</span>
                  <span className={hasRole('courier_level2') ? 'text-green-400' : 'text-red-400'}>
                    {hasRole('courier_level2') ? '✅' : '❌'}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span>信使等级3:</span>
                  <span className={hasRole('courier_level3') ? 'text-green-400' : 'text-red-400'}>
                    {hasRole('courier_level3') ? '✅' : '❌'}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span>信使等级4:</span>
                  <span className={hasRole('courier_level4') ? 'text-green-400' : 'text-red-400'}>
                    {hasRole('courier_level4') ? '✅' : '❌'}
                  </span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span>platform_admin:</span>
                  <span className={hasRole('platform_admin') ? 'text-green-400' : 'text-red-400'}>
                    {hasRole('platform_admin') ? '✅' : '❌'}
                  </span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span>super_admin:</span>
                  <span className={hasRole('super_admin') ? 'text-green-400' : 'text-red-400'}>
                    {hasRole('super_admin') ? '✅' : '❌'}
                  </span>
                </div>
              </div>
            </div>

            {/* 权限列表 */}
            {user?.permissions && user.permissions.length > 0 && (
              <div className="border-t border-gray-700 pt-1 sm:pt-2">
                <div className="text-gray-400 font-medium text-xs mb-1">
                  权限列表 ({user.permissions.length}):
                </div>
                <div className="space-y-0.5 max-h-20 overflow-y-auto">
                  {user.permissions.slice(0, isExpanded ? undefined : 3).map((permission: string) => (
                    <div key={permission} className="text-xs text-purple-300 truncate">
                      • {permission}
                    </div>
                  ))}
                  {!isExpanded && user.permissions.length > 3 && (
                    <div className="text-xs text-gray-500">
                      ... +{user.permissions.length - 3} more
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* 信使信息 */}
            {user?.courierInfo && (
              <div className="border-t border-gray-700 pt-1 sm:pt-2">
                <div className="text-gray-400 font-medium text-xs mb-1">信使信息:</div>
                <div className="space-y-0.5 text-xs">
                  <div className="flex items-center justify-between">
                    <span>级别:</span>
                    <span className="text-orange-300">
                      {user.courierInfo.level}级
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span>区域:</span>
                    <span className="text-orange-300 truncate max-w-[100px]">
                      {user.courierInfo.zoneCode}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span>积分:</span>
                    <span className="text-orange-300">
                      {user.courierInfo.points}
                    </span>
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* 底部状态条 */}
      <div className="px-2 sm:px-3 pb-2 sm:pb-3">
        <div className="text-xs text-gray-500 text-center">
          {isDragging ? '🔄 拖拽中...' : '🖱️ 可拖拽移动'}
        </div>
      </div>
    </div>
  )
})

export { UserDebugPanel }