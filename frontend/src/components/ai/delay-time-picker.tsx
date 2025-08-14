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
    description: '快速匹配',
    color: 'bg-green-50 border-green-200 text-green-700'
  },
  {
    id: '3hours',
    label: '3小时后',
    icon: Coffee,
    description: '下午茶时间',
    color: 'bg-orange-50 border-orange-200 text-orange-700'
  },
  {
    id: 'tomorrow_morning',
    label: '明天早上',
    icon: Sunrise,
    description: '早上8点',
    color: 'bg-yellow-50 border-yellow-200 text-yellow-700'
  },
  {
    id: 'tomorrow',
    label: '明天此时',
    icon: CalendarDays,
    description: '24小时后',
    color: 'bg-blue-50 border-blue-200 text-blue-700'
  },
  {
    id: 'weekend',
    label: '周末上午',
    icon: Moon,
    description: '周六10点',
    color: 'bg-purple-50 border-purple-200 text-purple-700'
  },
  {
    id: 'nextweek',
    label: '下周此时',
    icon: Calendar,
    description: '7天后',
    color: 'bg-indigo-50 border-indigo-200 text-indigo-700'
  },
]

export function DelayTimePicker({
  value,
  onChange,
  className = ''
}: DelayTimePickerProps) {
  const [mode, setMode] = useState<'preset' | 'relative' | 'absolute'>('preset')
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false)
  
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
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            {presetOptions.map((option) => {
              const IconComponent = option.icon
              const isSelected = value?.type === 'preset' && value.presetOption === option.id
              
              return (
                <Button
                  key={option.id}
                  variant={isSelected ? "default" : "outline"}
                  className={`h-auto p-4 ${isSelected ? '' : option.color}`}
                  onClick={() => handlePresetClick(option.id)}
                >
                  <div className="flex flex-col items-center gap-2 text-center">
                    <IconComponent className="h-5 w-5" />
                    <div>
                      <div className="font-medium">{option.label}</div>
                      <div className="text-xs opacity-70">{option.description}</div>
                    </div>
                  </div>
                </Button>
              )
            })}
          </div>
        </CardContent>
      </Card>

      {/* 高级选项 */}
      <Collapsible open={isAdvancedOpen} onOpenChange={setIsAdvancedOpen}>
        <CollapsibleTrigger asChild>
          <Button variant="ghost" className="w-full">
            <Plus className="h-4 w-4 mr-2" />
            高级时间设置
          </Button>
        </CollapsibleTrigger>
        
        <CollapsibleContent className="space-y-4">
          {/* 相对时间设置 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">相对时间</CardTitle>
              <CardDescription>设置相对于现在的时间</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <Label htmlFor="days">天</Label>
                  <Input
                    id="days"
                    type="number"
                    min="0"
                    value={relativeDays}
                    onChange={(e) => setRelativeDays(Number(e.target.value))}
                    onBlur={handleRelativeTimeUpdate}
                  />
                </div>
                <div>
                  <Label htmlFor="hours">小时</Label>
                  <Input
                    id="hours"
                    type="number"
                    min="0"
                    max="23"
                    value={relativeHours}
                    onChange={(e) => setRelativeHours(Number(e.target.value))}
                    onBlur={handleRelativeTimeUpdate}
                  />
                </div>
                <div>
                  <Label htmlFor="minutes">分钟</Label>
                  <Input
                    id="minutes"
                    type="number"
                    min="0"
                    max="59"
                    value={relativeMinutes}
                    onChange={(e) => setRelativeMinutes(Number(e.target.value))}
                    onBlur={handleRelativeTimeUpdate}
                  />
                </div>
              </div>
              
              <div>
                <Label htmlFor="description">自定义描述（可选）</Label>
                <Input
                  id="description"
                  placeholder="例如：等我心情平复后"
                  value={customDescription}
                  onChange={(e) => setCustomDescription(e.target.value)}
                  onBlur={handleRelativeTimeUpdate}
                />
              </div>
              
              {(relativeDays > 0 || relativeHours > 0 || relativeMinutes > 0) && (
                <Button
                  onClick={() => setMode('relative')}
                  variant={value?.type === 'relative' ? 'default' : 'outline'}
                  className="w-full"
                >
                  使用相对时间
                </Button>
              )}
            </CardContent>
          </Card>

          {/* 绝对时间设置 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">指定日期和时间</CardTitle>
              <CardDescription>选择具体的日期和时间</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-4">
                <div className="flex-1">
                  <Label>选择日期</Label>
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        className="w-full justify-start text-left font-normal"
                      >
                        <CalendarDays className="mr-2 h-4 w-4" />
                        {selectedDate 
                          ? format(selectedDate, 'yyyy年MM月dd日', { locale: zhCN })
                          : '选择日期'
                        }
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0">
                      <CalendarComponent
                        mode="single"
                        selected={selectedDate}
                        onSelect={handleAbsoluteTimeUpdate}
                        initialFocus
                        disabled={(date) => date < new Date()}
                      />
                    </PopoverContent>
                  </Popover>
                </div>
                
                <div className="flex-1">
                  <Label htmlFor="time">选择时间</Label>
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
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </CollapsibleContent>
      </Collapsible>

      {/* 预期时间显示 */}
      {expectedTime && (
        <Card className="bg-blue-50 border-blue-200">
          <CardContent className="pt-4">
            <div className="flex items-center gap-2 text-blue-800">
              <Clock className="h-4 w-4" />
              <span className="font-medium">
                预计回信时间：
                {formatDateTime(expectedTime)}
              </span>
            </div>
            <div className="text-sm text-blue-600 mt-1">
              还有 {Math.ceil((expectedTime.getTime() - Date.now()) / (1000 * 60))} 分钟
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}