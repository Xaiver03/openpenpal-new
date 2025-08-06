'use client'

import React, { useState } from 'react'
import { Sparkles, RefreshCw, BookOpen, Hash } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { aiService } from '@/lib/services/ai-service'
import { toast } from 'sonner'

interface Inspiration {
  id: string
  theme: string
  prompt: string
  style: string
  tags: string[]
}

interface AIWritingInspirationProps {
  theme?: string
  onSelectInspiration?: (inspiration: Inspiration) => void
  className?: string
}

export function AIWritingInspiration({
  theme,
  onSelectInspiration,
  className = '',
}: AIWritingInspirationProps) {
  const [inspirations, setInspirations] = useState<Inspiration[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchInspirations = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const result = await aiService.generateWritingPrompt({
        theme: theme || '日常生活',
        count: 3,
      })
      
      if (result.inspirations) {
        setInspirations(result.inspirations)
      } else {
        throw new Error('未获取到灵感数据')
      }
    } catch (err) {
      console.error('Failed to fetch inspirations:', err)
      setError('获取灵感失败，请稍后重试')
      toast.error('获取灵感失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSelectInspiration = (inspiration: Inspiration) => {
    if (onSelectInspiration) {
      onSelectInspiration(inspiration)
      toast.success('已选择写作灵感')
    }
  }

  if (loading) {
    return (
      <div className={`space-y-4 ${className}`}>
        {[1, 2, 3].map((i) => (
          <Card key={i} className="p-4">
            <Skeleton className="h-4 w-3/4 mb-2" />
            <Skeleton className="h-3 w-full mb-2" />
            <Skeleton className="h-3 w-2/3" />
          </Card>
        ))}
      </div>
    )
  }

  return (
    <div className={`space-y-4 ${className}`}>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-amber-600" />
          <h3 className="text-lg font-semibold">AI写作灵感</h3>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={fetchInspirations}
          disabled={loading}
          className="gap-2"
        >
          <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          {inspirations.length > 0 ? '换一批' : '获取灵感'}
        </Button>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {inspirations.length === 0 && !error && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-8 text-center">
            <BookOpen className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground mb-4">
              点击"获取灵感"按钮，让AI为您提供写作建议
            </p>
            <Button onClick={fetchInspirations} className="gap-2">
              <Sparkles className="h-4 w-4" />
              获取写作灵感
            </Button>
          </CardContent>
        </Card>
      )}

      {inspirations.map((inspiration) => (
        <Card
          key={inspiration.id}
          className="cursor-pointer transition-all hover:shadow-md hover:border-amber-200"
          onClick={() => handleSelectInspiration(inspiration)}
        >
          <CardHeader className="pb-3">
            <div className="flex items-start justify-between">
              <CardTitle className="text-base">{inspiration.theme}</CardTitle>
              <Badge variant="secondary" className="ml-2">
                {inspiration.style}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <CardDescription className="text-sm leading-relaxed mb-3">
              {inspiration.prompt}
            </CardDescription>
            <div className="flex items-center gap-1 flex-wrap">
              <Hash className="h-3 w-3 text-muted-foreground" />
              {inspiration.tags.map((tag) => (
                <Badge key={tag} variant="outline" className="text-xs">
                  {tag}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}