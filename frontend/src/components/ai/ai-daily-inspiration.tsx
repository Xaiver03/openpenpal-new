'use client'

import React, { useEffect, useState } from 'react'
import { Sparkles, Calendar, Quote } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { aiService } from '@/lib/services/ai-service'

interface DailyInspiration {
  date: string
  theme: string
  prompt: string
  quote: string
  tips: string[]
}

interface AIDailyInspirationProps {
  onSelectPrompt?: (prompt: string) => void
  className?: string
}

export function AIDailyInspiration({
  onSelectPrompt,
  className = '',
}: AIDailyInspirationProps) {
  const [inspiration, setInspiration] = useState<DailyInspiration | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchDailyInspiration()
  }, [])

  const fetchDailyInspiration = async () => {
    setLoading(true)
    try {
      const result = await aiService.getDailyInspiration()
      setInspiration(result)
    } catch (error) {
      console.error('Failed to fetch daily inspiration:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSelectPrompt = () => {
    if (onSelectPrompt && inspiration) {
      onSelectPrompt(inspiration.prompt)
    }
  }

  if (loading) {
    return (
      <Card className={className}>
        <CardHeader>
          <Skeleton className="h-6 w-32 mb-2" />
          <Skeleton className="h-4 w-24" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-20 w-full mb-4" />
          <Skeleton className="h-16 w-full mb-4" />
          <Skeleton className="h-4 w-full mb-2" />
          <Skeleton className="h-4 w-3/4" />
        </CardContent>
      </Card>
    )
  }

  if (!inspiration) {
    return null
  }

  return (
    <Card 
      className={`bg-gradient-to-br from-amber-50 to-orange-50 border-amber-200 ${className}`}
      onClick={handleSelectPrompt}
    >
      <CardHeader>
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-amber-600" />
            <CardTitle>今日灵感</CardTitle>
          </div>
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <Calendar className="h-3 w-3" />
            <span>{new Date().toLocaleDateString('zh-CN')}</span>
          </div>
        </div>
        <Badge variant="secondary" className="w-fit">
          {inspiration.theme}
        </Badge>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="bg-white/60 rounded-lg p-4 cursor-pointer hover:bg-white/80 transition-colors">
          <p className="text-base leading-relaxed font-medium text-gray-800">
            {inspiration.prompt}
          </p>
        </div>

        {inspiration.quote && (
          <div className="flex gap-2 items-start bg-amber-100/50 rounded-lg p-3">
            <Quote className="h-4 w-4 text-amber-700 mt-1 flex-shrink-0" />
            <p className="text-sm italic text-amber-900">
              {inspiration.quote}
            </p>
          </div>
        )}

        {inspiration.tips.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold mb-2 text-gray-700">
              写作小贴士：
            </h4>
            <ul className="space-y-1">
              {inspiration.tips.map((tip, index) => (
                <li key={index} className="flex items-start gap-2 text-sm text-gray-600">
                  <span className="text-amber-600 mt-0.5">•</span>
                  <span>{tip}</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        <CardDescription className="text-xs text-center mt-4">
          点击使用今日写作主题
        </CardDescription>
      </CardContent>
    </Card>
  )
}