'use client'

import React, { useState, useEffect } from 'react'
import { Calendar, Clock, Zap, Coffee, Sunrise, Moon, CalendarDays, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Calendar as CalendarComponent } from '@/components/ui/calendar'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'

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

interface DelayTimePickerProps {
  value?: DelayConfig
  onChange: (config: DelayConfig) => void
  className?: string
}

// 预设时间选项
const presetOptions = [
  {
    id: '1hour',
    label: '1小时后',
    icon: Zap,
    description: '稍后即达，给彼此一点思考的时间',
    color: 'bg-green-50 border-green-200 text-green-700',
    hover: 'hover:bg-green-100 hover:border-green-300 hover:scale-105'
  },
  {
    id: '3hours',
    label: '3小时后',
    icon: Coffee,
    description: '下午茶时光，温暖的午后邂逅',
    color: 'bg-orange-50 border-orange-200 text-orange-700',
    hover: 'hover:bg-orange-100 hover:border-orange-300 hover:scale-105'
  },
  {
    id: 'tomorrow_morning',
    label: '明天早上',
    icon: Sunrise,
    description: '清晨第一缕阳光，美好的一天开始',
    color: 'bg-yellow-50 border-yellow-200 text-yellow-700',
    hover: 'hover:bg-yellow-100 hover:border-yellow-300 hover:scale-105'
  },
  {
    id: 'tomorrow',
    label: '明天此时',
    icon: CalendarDays,
    description: '同一时刻的约定，跨越今日与明天',
    color: 'bg-blue-50 border-blue-200 text-blue-700',
    hover: 'hover:bg-blue-100 hover:border-blue-300 hover:scale-105'
  },
  {
    id: 'weekend',
    label: '周末上午',
    icon: Moon,
    description: '悠闲的周末时光，适合慢慢品读',
    color: 'bg-purple-50 border-purple-200 text-purple-700',
    hover: 'hover:bg-purple-100 hover:border-purple-300 hover:scale-105'
  },
  {
    id: 'nextweek',
    label: '下周此时',
    icon: Calendar,
    description: '一周的时间沉淀，让思绪更加深刻',
    color: 'bg-indigo-50 border-indigo-200 text-indigo-700',
    hover: 'hover:bg-indigo-100 hover:border-indigo-300 hover:scale-105'
  },
]

export function DelayTimePicker({
  value,
  onChange,
  className = ''
}: DelayTimePickerProps) {
  const [mode, setMode] = useState<'preset' | 'relative' | 'absolute'>('preset')
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  
  // 相对时间状态
  const [relativeDays, setRelativeDays] = useState(0)
  const [relativeHours, setRelativeHours] = useState(0)
  const [relativeMinutes, setRelativeMinutes] = useState(0)
  const [customDescription, setCustomDescription] = useState('')
  
  // 绝对时间状态
  const [selectedDate, setSelectedDate] = useState<Date>()
  const [selectedTime, setSelectedTime] = useState('09:00')

  // 初始化值
  useEffect(() => {
    if (value) {
      setMode(value.type)
      if (value.type === 'relative') {
        setRelativeDays(value.relativeDays || 0)
        setRelativeHours(value.relativeHours || 0)
        setRelativeMinutes(value.relativeMinutes || 0)
        setCustomDescription(value.userDescription || '')
      } else if (value.type === 'absolute' && value.absoluteTime) {
        setSelectedDate(value.absoluteTime)
        setSelectedTime(format(value.absoluteTime, 'HH:mm'))
      }
    }
  }, [value])

  // 处理预设选项点击
  const handlePresetClick = (presetOption: string) => {
    const config: DelayConfig = {
      type: 'preset',
      presetOption
    }
    onChange(config)
  }

  // 处理相对时间更新
  const handleRelativeTimeUpdate = () => {
    const config: DelayConfig = {
      type: 'relative',
      relativeDays,
      relativeHours,
      relativeMinutes,
      userDescription: customDescription || undefined
    }
    onChange(config)
  }

  // 处理绝对时间更新
  const handleAbsoluteTimeUpdate = (date?: Date) => {
    if (!date) return
    
    const [hours, minutes] = selectedTime.split(':').map(Number)
    const absoluteTime = new Date(date)
    absoluteTime.setHours(hours, minutes, 0, 0)
    
    const config: DelayConfig = {
      type: 'absolute',
      absoluteTime
    }
    onChange(config)
    setSelectedDate(date)
  }

  // 计算预期时间
  const getExpectedTime = () => {
    const now = new Date()
    
    if (!value) return null
    
    switch (value.type) {
      case 'preset':
        // 这里应该与后端逻辑保持一致
        switch (value.presetOption) {
          case '1hour':
            return new Date(now.getTime() + 60 * 60 * 1000)
          case '3hours':
            return new Date(now.getTime() + 3 * 60 * 60 * 1000)
          case 'tomorrow_morning':
            const tomorrow = new Date(now)
            tomorrow.setDate(tomorrow.getDate() + 1)
            tomorrow.setHours(8, 0, 0, 0)
            return tomorrow
          case 'tomorrow':
            return new Date(now.getTime() + 24 * 60 * 60 * 1000)
          case 'weekend':
            const daysUntilSaturday = (6 - now.getDay() + 7) % 7 || 7
            const saturday = new Date(now)
            saturday.setDate(saturday.getDate() + daysUntilSaturday)
            saturday.setHours(10, 0, 0, 0)
            return saturday
          case 'nextweek':
            return new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000)
        }
        break
        
      case 'relative':
        const relativeTime = new Date(now)
        relativeTime.setDate(relativeTime.getDate() + (value.relativeDays || 0))
        relativeTime.setHours(relativeTime.getHours() + (value.relativeHours || 0))
        relativeTime.setMinutes(relativeTime.getMinutes() + (value.relativeMinutes || 0))
        return relativeTime
        
      case 'absolute':
        return value.absoluteTime || null
    }
    
    return null
  }

  const expectedTime = getExpectedTime()

  // 格式化日期时间为中文格式
  const formatDateTime = (date: Date) => {
    return format(date, 'MM月dd日 HH:mm', { locale: zhCN })
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* 预设选项 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Clock className="h-5 w-5" />
            选择回信时间
          </CardTitle>
          <CardDescription>
            选择你希望收到AI回信的时间，让等待变成期待
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {presetOptions.map((option) => {
              const IconComponent = option.icon
              const isSelected = value?.type === 'preset' && value.presetOption === option.id
              
              return (
                <Button
                  key={option.id}
                  variant={isSelected ? "default" : "outline"}
                  className={`h-auto p-3 transition-all duration-200 w-full ${
                    isSelected 
                      ? 'shadow-md ring-2 ring-blue-200' 
                      : `${option.color} ${option.hover}`
                  }`}
                  onClick={() => {
                    setIsSubmitting(true)
                    handlePresetClick(option.id)
                    setTimeout(() => setIsSubmitting(false), 300)
                  }}
                  disabled={isSubmitting}
                >
                  <div className="flex flex-col items-center gap-2 text-center w-full">
                    <IconComponent className={`h-5 w-5 ${isSubmitting ? 'animate-pulse' : ''}`} />
                    <div className="w-full">
                      <div className="font-medium">{option.label}</div>
                      <div className="text-xs opacity-75 leading-tight mt-1">
                        {option.description}
                      </div>
                    </div>
                  </div>
                </Button>
              )
            })}
          </div>
        </CardContent>
      </Card>

      {/* 自定义时间选项 */}
      <Collapsible open={isAdvancedOpen} onOpenChange={setIsAdvancedOpen}>
        <CollapsibleTrigger asChild>
          <Button 
            variant="outline" 
            className="w-full transition-all duration-200 hover:bg-blue-50 hover:shadow-sm border-blue-200 text-blue-700"
          >
            <Plus className={`h-4 w-4 mr-2 transition-transform duration-200 ${
              isAdvancedOpen ? 'rotate-45' : ''
            }`} />
            <span className="font-medium">
              {isAdvancedOpen ? '收起自定义时间' : '🎯 自定义送达时间'}
            </span>
          </Button>
        </CollapsibleTrigger>
        
        <CollapsibleContent className="space-y-4">
          {/* 相对时间设置 */}
          <Card className="border-dashed border-gray-200">
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <Clock className="h-4 w-4" />
                相对时间设置
              </CardTitle>
              <CardDescription>
                相对于现在设置一个时间间隔，比如「2天3小时后」
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <Label htmlFor="days" className="text-sm font-medium text-gray-700">
                    天数
                  </Label>
                  <Input
                    id="days"
                    type="number"
                    min="0"
                    max="365"
                    value={relativeDays}
                    onChange={(e) => setRelativeDays(Math.max(0, Number(e.target.value)))}
                    onBlur={handleRelativeTimeUpdate}
                    className="text-center"
                    placeholder="0"
                  />
                  <p className="text-xs text-gray-500 mt-1">最多365天</p>
                </div>
                <div>
                  <Label htmlFor="hours" className="text-sm font-medium text-gray-700">
                    小时
                  </Label>
                  <Input
                    id="hours"
                    type="number"
                    min="0"
                    max="23"
                    value={relativeHours}
                    onChange={(e) => setRelativeHours(Math.max(0, Math.min(23, Number(e.target.value))))}
                    onBlur={handleRelativeTimeUpdate}
                    className="text-center"
                    placeholder="0"
                  />
                  <p className="text-xs text-gray-500 mt-1">0-23小时</p>
                </div>
                <div>
                  <Label htmlFor="minutes" className="text-sm font-medium text-gray-700">
                    分钟
                  </Label>
                  <Input
                    id="minutes"
                    type="number"
                    min="0"
                    max="59"
                    value={relativeMinutes}
                    onChange={(e) => setRelativeMinutes(Math.max(0, Math.min(59, Number(e.target.value))))}
                    onBlur={handleRelativeTimeUpdate}
                    className="text-center"
                    placeholder="0"
                  />
                  <p className="text-xs text-gray-500 mt-1">0-59分钟</p>
                </div>
              </div>
              
              <div>
                <Label htmlFor="description" className="text-sm font-medium text-gray-700">
                  个性化描述 <span className="text-gray-400">(可选)</span>
                </Label>
                <Input
                  id="description"
                  placeholder="例如：等我心情平复后 / 给我们一些思考的时间"
                  value={customDescription}
                  onChange={(e) => setCustomDescription(e.target.value)}
                  onBlur={handleRelativeTimeUpdate}
                  className="mt-1"
                />
                <p className="text-xs text-gray-500 mt-1">
                  添加一段个性化的时间描述，让等待更有意义
                </p>
              </div>
              
              {(relativeDays > 0 || relativeHours > 0 || relativeMinutes > 0) && (
                <div className="pt-2 border-t border-gray-100">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm text-gray-600">
                      将在 {relativeDays > 0 ? `${relativeDays}天` : ''} 
                      {relativeHours > 0 ? `${relativeHours}小时` : ''} 
                      {relativeMinutes > 0 ? `${relativeMinutes}分钟` : ''} 后送达
                    </span>
                  </div>
                  <Button
                    onClick={() => {
                      setMode('relative')
                      handleRelativeTimeUpdate()
                    }}
                    variant={value?.type === 'relative' ? 'default' : 'outline'}
                    className="w-full transition-all duration-200 hover:scale-105"
                  >
                    {value?.type === 'relative' ? '✓ 已选择相对时间' : '使用这个时间设置'}
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          {/* 绝对时间设置 */}
          <Card className="border-dashed border-gray-200">
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <CalendarDays className="h-4 w-4" />
                指定具体时间
              </CardTitle>
              <CardDescription>
                选择一个确切的日期和时间，让这份等待有确定的期待
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-4">
                <div className="flex-1">
                  <Label className="text-sm font-medium text-gray-700">
                    选择日期
                  </Label>
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        className={`w-full justify-start text-left font-normal mt-1 ${
                          selectedDate ? 'text-gray-900' : 'text-gray-500'
                        }`}
                      >
                        <CalendarDays className="mr-2 h-4 w-4" />
                        {selectedDate 
                          ? format(selectedDate, 'yyyy年MM月dd日', { locale: zhCN })
                          : '点击选择日期'
                        }
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0" align="start">
                      <CalendarComponent
                        mode="single"
                        selected={selectedDate}
                        onSelect={handleAbsoluteTimeUpdate}
                        initialFocus
                        disabled={(date) => date < new Date()}
                        className="rounded-md border"
                      />
                    </PopoverContent>
                  </Popover>
                  <p className="text-xs text-gray-500 mt-1">不能选择过去的日期</p>
                </div>
                
                <div className="flex-1">
                  <Label htmlFor="time" className="text-sm font-medium text-gray-700">
                    选择时间
                  </Label>
                  <Input
                    id="time"
                    type="time"
                    value={selectedTime}
                    onChange={(e) => {
                      setSelectedTime(e.target.value)
                      if (selectedDate) {
                        handleAbsoluteTimeUpdate(selectedDate)
                      }
                    }}
                    className="mt-1"
                  />
                  <p className="text-xs text-gray-500 mt-1">24小时格式</p>
                </div>
              </div>
              
              {selectedDate && (
                <div className="pt-2 border-t border-gray-100">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm text-gray-600">
                      将在 {format(selectedDate, 'MM月dd日 ', { locale: zhCN })}
                      {selectedTime} 准时送达
                    </span>
                  </div>
                  <Button
                    onClick={() => {
                      setMode('absolute')
                      handleAbsoluteTimeUpdate(selectedDate)
                    }}
                    variant={value?.type === 'absolute' ? 'default' : 'outline'}
                    className="w-full transition-all duration-200 hover:scale-105"
                  >
                    {value?.type === 'absolute' ? '✓ 已选择指定时间' : '使用这个时间'}
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </CollapsibleContent>
      </Collapsible>

      {/* 预期时间显示 */}
      {expectedTime && (
        <Card className="bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200 shadow-sm">
          <CardContent className="pt-4">
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-100 rounded-full">
                  <Clock className="h-4 w-4 text-blue-600" />
                </div>
                <div>
                  <div className="font-medium text-blue-900">
                    预计送达时间
                  </div>
                  <div className="text-lg font-semibold text-blue-800 mt-1">
                    {formatDateTime(expectedTime)}
                  </div>
                </div>
              </div>
              <Badge variant="secondary" className="bg-blue-100 text-blue-700">
                {(() => {
                  const diffMinutes = Math.ceil((expectedTime.getTime() - Date.now()) / (1000 * 60))
                  const diffHours = Math.floor(diffMinutes / 60)
                  const diffDays = Math.floor(diffHours / 24)
                  
                  if (diffDays > 0) {
                    return `${diffDays}天后`
                  } else if (diffHours > 0) {
                    return `${diffHours}小时后`
                  } else if (diffMinutes > 0) {
                    return `${diffMinutes}分钟后`
                  } else {
                    return '即将送达'
                  }
                })()}
              </Badge>
            </div>
            
            <div className="mt-3 p-2 bg-white/60 rounded-lg border border-blue-100">
              <p className="text-sm text-blue-700 leading-relaxed break-words">
                {value?.userDescription || '你的信将在选定的时间准时送达，请耐心等待这份美好的邂逅 ✨'}
              </p>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}