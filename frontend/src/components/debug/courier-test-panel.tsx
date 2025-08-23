'use client'

import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { 
  enableTestCourierMode, 
  disableTestCourierMode, 
  isTestMode,
  getTestCourierLevel 
} from '@/lib/auth/test-courier-mock'
import { useAuth } from '@/stores/user-store'
import { Crown, Truck, Users, Building, X, Move } from 'lucide-react'

export function CourierTestPanel() {
  const [testModeEnabled, setTestModeEnabled] = useState(isTestMode())
  const [selectedLevel, setSelectedLevel] = useState<1 | 2 | 3 | 4>(getTestCourierLevel())
  const [isOpen, setIsOpen] = useState(false)
  const [position, setPosition] = useState({ x: 20, y: 100 })
  const [isDragging, setIsDragging] = useState(false)
  const dragRef = useRef<HTMLDivElement>(null)
  const dragOffset = useRef({ x: 0, y: 0 })
  const { user, refreshUser } = useAuth()
  
  // 只在开发环境显示
  if (process.env.NODE_ENV !== 'development') {
    return null
  }

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging || typeof window === 'undefined') return
      setPosition({
        x: Math.max(0, Math.min(window.innerWidth - 320, e.clientX - dragOffset.current.x)),
        y: Math.max(0, Math.min(window.innerHeight - 400, e.clientY - dragOffset.current.y))
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
  }, [isDragging])
  
  const handleToggleTestMode = async (enabled: boolean) => {
    if (enabled) {
      enableTestCourierMode(selectedLevel)
    } else {
      disableTestCourierMode()
    }
    
    setTestModeEnabled(enabled)
    
    // 刷新用户状态以应用更改
    if (user) {
      await refreshUser()
      if (typeof window !== 'undefined') {
        window.location.reload() // 强制刷新页面以确保状态更新
      }
    }
  }
  
  const handleLevelChange = (level: string) => {
    const newLevel = parseInt(level) as 1 | 2 | 3 | 4
    setSelectedLevel(newLevel)
    
    if (testModeEnabled) {
      enableTestCourierMode(newLevel)
      // 刷新用户状态
      refreshUser().then(() => {
        if (typeof window !== 'undefined') {
          window.location.reload()
        }
      })
    }
  }
  
  const levelIcons = {
    1: <Building className="w-4 h-4 text-orange-600" />,
    2: <Users className="w-4 h-4 text-orange-600" />,
    3: <Truck className="w-4 h-4 text-orange-600" />,
    4: <Crown className="w-4 h-4 text-orange-600" />
  }
  
  const levelNames = {
    1: '一级信使 (楼栋)',
    2: '二级信使 (片区)',
    3: '三级信使 (学校)',
    4: '四级信使 (城市)'
  }

  const handleMouseDown = (e: React.MouseEvent) => {
    const rect = dragRef.current?.getBoundingClientRect()
    if (rect) {
      dragOffset.current = {
        x: e.clientX - rect.left,
        y: e.clientY - rect.top
      }
      setIsDragging(true)
    }
  }

  // 浮动按钮 - 未展开时显示
  if (!isOpen) {
    return (
      <Button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-4 right-4 bg-orange-500 hover:bg-orange-600 text-white shadow-lg z-50 rounded-full"
        size="icon"
        title="打开信使测试面板"
      >
        🧪
      </Button>
    )
  }
  
  return (
    <Card 
      ref={dragRef}
      className="fixed w-80 shadow-lg border-2 border-orange-200 bg-orange-50/95 backdrop-blur-sm z-50 select-none"
      style={{ left: `${position.x}px`, top: `${position.y}px` }}
    >
      <CardHeader 
        className="pb-3 cursor-move" 
        onMouseDown={handleMouseDown}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Move className="w-4 h-4 text-orange-600" />
            <CardTitle className="text-sm font-semibold text-orange-900">
              🧪 信使测试面板
            </CardTitle>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-orange-700 border-orange-300 text-xs">
              开发模式
            </Badge>
            <Button
              onClick={() => setIsOpen(false)}
              variant="ghost"
              size="icon"
              className="h-6 w-6 hover:bg-orange-200"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
        <CardDescription className="text-xs text-orange-700">
          用于测试各级信使功能，仅在开发环境可用
        </CardDescription>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* 测试模式开关 */}
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium text-orange-900">启用测试模式</label>
          <Switch 
            checked={testModeEnabled}
            onCheckedChange={handleToggleTestMode}
          />
        </div>
        
        {/* 信使等级选择 */}
        <div className="space-y-2">
          <label className="text-sm font-medium text-orange-900">信使等级</label>
          <Select value={selectedLevel.toString()} onValueChange={handleLevelChange}>
            <SelectTrigger className="w-full">
              <SelectValue>
                <div className="flex items-center gap-2">
                  {levelIcons[selectedLevel]}
                  {levelNames[selectedLevel]} 
                </div>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {[1, 2, 3, 4].map((level) => (
                <SelectItem key={level} value={level.toString()}>
                  <div className="flex items-center gap-2">
                    {levelIcons[level as 1 | 2 | 3 | 4]}
                    {levelNames[level as 1 | 2 | 3 | 4]}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        
        {/* 当前状态 */}
        <div className="bg-white/80 rounded-lg p-3 space-y-2">
          <div className="text-xs font-medium text-orange-900">当前状态:</div>
          <div className="text-xs text-orange-700 space-y-1">
            <div>用户: {user?.username || '未登录'}</div>
            <div>角色: {user?.role || 'N/A'}</div>
            <div>
              信使等级: {user?.courierInfo?.level ? 
                `${levelNames[user.courierInfo.level as 1 | 2 | 3 | 4]}` : 
                '无'
              }
            </div>
            <div className="flex items-center gap-2">
              测试模式: 
              {testModeEnabled ? 
                <Badge className="text-xs bg-green-100 text-green-700 border-green-300">✅ 启用</Badge> : 
                <Badge className="text-xs bg-gray-100 text-gray-700 border-gray-300">❌ 禁用</Badge>
              }
            </div>
          </div>
        </div>
        
        {/* 提示信息 */}
        <div className="text-xs text-orange-600 bg-orange-100/50 rounded p-2">
          💡 启用测试模式后，当前用户将临时获得选定等级的信使权限
        </div>
      </CardContent>
    </Card>
  )
}