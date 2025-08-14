'use client'

import React, { useState } from 'react'
import { Users, Brain, Heart, Star, ChevronRight, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Progress } from '@/components/ui/progress'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { DelayTimePicker, type DelayConfig } from '@/components/ai/delay-time-picker'
import { aiService } from '@/lib/services/ai-service'
import { toast } from 'sonner'

interface PenpalMatch {
  userId: string
  username: string
  score: number
  reason: string
  common_tags: string[]
}

interface AIPenpalMatchProps {
  letterId: string
  onSelectMatch?: (match: PenpalMatch) => void
  className?: string
}

export function AIPenpalMatch({
  letterId,
  onSelectMatch,
  className = '',
}: AIPenpalMatchProps) {
  const [matches, setMatches] = useState<PenpalMatch[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [delayConfig, setDelayConfig] = useState<DelayConfig>({
    type: 'preset',
    presetOption: '1hour'
  })
  const [showDelayPicker, setShowDelayPicker] = useState(true)

  const fetchMatches = async () => {
    if (!letterId) {
      toast.error('请先保存信件草稿')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const result = await aiService.matchPenpal({
        letterId: letterId,
        max_matches: 3,
        delay_config: delayConfig,
      })

      if (result.matches && result.matches.length > 0) {
        setMatches(result.matches)
        toast.success(`找到 ${result.matches.length} 位合适的笔友`)
      } else {
        setError('暂未找到合适的笔友，请稍后再试')
      }
    } catch (err) {
      console.error('Failed to fetch matches:', err)
      setError('匹配失败，请稍后重试')
      toast.error('匹配失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSelectMatch = (match: PenpalMatch) => {
    if (onSelectMatch) {
      onSelectMatch(match)
      toast.success(`已选择 ${match.username} 作为收信人`)
    }
  }

  const getScoreColor = (score: number) => {
    if (score >= 0.8) return 'text-green-600'
    if (score >= 0.6) return 'text-yellow-600'
    return 'text-gray-600'
  }

  const getScoreLabel = (score: number) => {
    if (score >= 0.8) return '非常匹配'
    if (score >= 0.6) return '较为匹配'
    return '可以尝试'
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* 延迟时间选择器 */}
      {showDelayPicker && (
        <DelayTimePicker
          value={delayConfig}
          onChange={setDelayConfig}
          className="mb-6"
        />
      )}

      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Brain className="h-5 w-5 text-purple-600" />
          <h3 className="text-lg font-semibold">AI笔友匹配</h3>
        </div>
        <Button
          onClick={fetchMatches}
          disabled={loading || !letterId}
          className="gap-2"
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              匹配中...
            </>
          ) : (
            <>
              <Users className="h-4 w-4" />
              {matches.length > 0 ? '重新匹配' : '智能匹配'}
            </>
          )}
        </Button>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {matches.length === 0 && !error && !loading && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-8 text-center">
            <Users className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground mb-4">
              点击"智能匹配"按钮，AI将根据您的信件内容
              <br />
              为您推荐最合适的笔友
            </p>
          </CardContent>
        </Card>
      )}

      {matches.map((match, index) => (
        <Card
          key={match.userId}
          className="cursor-pointer transition-all hover:shadow-md hover:border-purple-200"
          onClick={() => handleSelectMatch(match)}
        >
          <CardHeader className="pb-3">
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                <Avatar>
                  <AvatarFallback>
                    {match.username.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <CardTitle className="text-base">{match.username}</CardTitle>
                  <div className="flex items-center gap-2 mt-1">
                    <Progress 
                      value={match.score * 100} 
                      className="w-20 h-2"
                    />
                    <span className={`text-sm font-medium ${getScoreColor(match.score)}`}>
                      {Math.round(match.score * 100)}%
                    </span>
                    <Badge variant="secondary" className="text-xs">
                      {getScoreLabel(match.score)}
                    </Badge>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-1">
                {index === 0 && (
                  <Badge variant="default" className="gap-1">
                    <Star className="h-3 w-3" />
                    推荐
                  </Badge>
                )}
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <CardDescription className="text-sm mb-3">
              <Heart className="h-3 w-3 inline mr-1 text-red-500" />
              {match.reason}
            </CardDescription>
            {match.common_tags.length > 0 && (
              <div className="flex items-center gap-1 flex-wrap">
                <span className="text-xs text-muted-foreground mr-1">
                  共同兴趣：
                </span>
                {match.common_tags.map((tag) => (
                  <Badge key={tag} variant="outline" className="text-xs">
                    {tag}
                  </Badge>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}