'use client'

import React, { useEffect, useState } from 'react'
import { Bot, Sparkles, CheckCircle } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea } from '@/components/ui/scroll-area'
import { aiService, AIPersona } from '@/lib/services/ai-service'

interface AIPersonaSelectorProps {
  value?: string
  onChange?: (personaId: string) => void
  className?: string
}

// 人设图标映射
const personaIcons: Record<string, string> = {
  poet: '🎭',
  philosopher: '🤔',
  artist: '🎨',
  scientist: '🔬',
  traveler: '✈️',
  historian: '📚',
  mentor: '💝',
  friend: '👫',
}

export function AIPersonaSelector({
  value,
  onChange,
  className = '',
}: AIPersonaSelectorProps) {
  const [personas, setPersonas] = useState<AIPersona[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchPersonas()
  }, [])

  const fetchPersonas = async () => {
    setLoading(true)
    try {
      const result = await aiService.getAIPersonas()
      setPersonas(result.personas || [])
    } catch (error) {
      console.error('Failed to fetch personas:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className={`space-y-2 ${className}`}>
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} className="p-4">
            <div className="flex items-center gap-3">
              <Skeleton className="h-12 w-12 rounded-full" />
              <div className="flex-1">
                <Skeleton className="h-4 w-24 mb-2" />
                <Skeleton className="h-3 w-full" />
              </div>
            </div>
          </Card>
        ))}
      </div>
    )
  }

  return (
    <ScrollArea className={`h-[400px] pr-4 ${className}`}>
      <RadioGroup value={value} onValueChange={onChange}>
        <div className="space-y-3">
          {personas.map((persona) => (
            <Card
              key={persona.id}
              className={`cursor-pointer transition-all ${
                value === persona.id
                  ? 'border-primary ring-2 ring-primary ring-offset-2'
                  : 'hover:border-gray-300'
              }`}
            >
              <Label htmlFor={persona.id} className="cursor-pointer">
                <CardContent className="p-4">
                  <div className="flex items-start gap-3">
                    <RadioGroupItem
                      value={persona.id}
                      id={persona.id}
                      className="mt-1"
                    />
                    <Avatar className="h-12 w-12">
                      {persona.avatar ? (
                        <AvatarImage src={persona.avatar} alt={persona.name} />
                      ) : (
                        <AvatarFallback className="text-lg">
                          {personaIcons[persona.id] || '🤖'}
                        </AvatarFallback>
                      )}
                    </Avatar>
                    <div className="flex-1 space-y-1">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium text-sm">{persona.name}</h4>
                        {value === persona.id && (
                          <CheckCircle className="h-4 w-4 text-primary" />
                        )}
                      </div>
                      <p className="text-xs text-muted-foreground leading-relaxed">
                        {persona.description}
                      </p>
                      {persona.id === 'poet' && (
                        <Badge variant="secondary" className="text-xs">
                          热门选择
                        </Badge>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Label>
            </Card>
          ))}
        </div>
      </RadioGroup>
    </ScrollArea>
  )
}

// 人设预览组件
export function AIPersonaPreview({ personaId }: { personaId: string }) {
  const [persona, setPersona] = useState<AIPersona | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchPersona = async () => {
      setLoading(true)
      try {
        const result = await aiService.getAIPersonas()
        const found = result.personas?.find((p: AIPersona) => p.id === personaId)
        setPersona(found || null)
      } catch (error) {
        console.error('Failed to fetch persona:', error)
      } finally {
        setLoading(false)
      }
    }

    if (personaId) {
      fetchPersona()
    }
  }, [personaId])

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-32 mb-2" />
          <Skeleton className="h-4 w-full" />
        </CardHeader>
      </Card>
    )
  }

  if (!persona) {
    return null
  }

  return (
    <Card className="bg-gradient-to-br from-purple-50 to-indigo-50 border-purple-200">
      <CardHeader>
        <div className="flex items-center gap-3">
          <Avatar className="h-10 w-10">
            <AvatarFallback className="text-lg bg-white">
              {personaIcons[persona.id] || '🤖'}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1">
            <CardTitle className="text-base flex items-center gap-2">
              {persona.name}
              <Bot className="h-4 w-4 text-purple-600" />
            </CardTitle>
            <CardDescription className="text-xs">
              AI笔友人设
            </CardDescription>
          </div>
          <Sparkles className="h-5 w-5 text-purple-600" />
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-gray-700 leading-relaxed">
          {persona.description}
        </p>
      </CardContent>
    </Card>
  )
}