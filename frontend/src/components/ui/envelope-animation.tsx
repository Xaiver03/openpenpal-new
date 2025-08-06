'use client'

import React, { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Mail, Heart, Eye, Calendar, MapPin } from 'lucide-react'

interface EnvelopeAnimationProps {
  letter: {
    id: number
    title: string
    preview: string
    author: string
    type: string
    date: string
    location: string
    likes: number
    views: number
    significance: string
    featured: boolean
  }
  onReadLetter?: () => void
}

export function EnvelopeAnimation({ letter, onReadLetter }: EnvelopeAnimationProps) {
  const [isOpened, setIsOpened] = useState(false)
  const [isAnimating, setIsAnimating] = useState(false)

  const handleOpenEnvelope = () => {
    if (isAnimating) return
    
    setIsAnimating(true)
    setTimeout(() => {
      setIsOpened(true)
      setIsAnimating(false)
    }, 800)
  }

  const handleCloseEnvelope = () => {
    if (isAnimating) return
    
    setIsAnimating(true)
    setTimeout(() => {
      setIsOpened(false)
      setIsAnimating(false)
    }, 500)
  }

  return (
    <div className="relative perspective-1000 group">
      <Card className={`
        relative transition-all duration-500 cursor-pointer overflow-hidden
        ${letter.featured ? 'border-amber-400 bg-gradient-to-br from-amber-50 to-orange-50' : 'border-amber-200'}
        ${isOpened ? 'shadow-2xl scale-105' : 'hover:shadow-lg hover:scale-[1.02]'}
      `}>
        {/* 信封效果 */}
        <div className={`
          absolute inset-0 bg-gradient-to-br from-amber-100 to-orange-100 
          transition-all duration-800 ease-in-out transform-gpu
          ${isOpened ? 'opacity-0 scale-110 rotate-6' : 'opacity-100'}
        `}>
          {/* 信封顶部折叠 */}
          <div className={`
            absolute top-0 left-0 right-0 h-24 bg-gradient-to-b from-amber-200 to-amber-100
            transition-transform duration-800 ease-in-out origin-top transform-gpu
            ${isAnimating ? (isOpened ? 'rotate-x-90' : 'rotate-x-0') : ''}
            ${isOpened ? 'rotate-x-90' : ''}
          `}>
            {/* 信封装饰线条 */}
            <div className="absolute top-4 left-4 right-4 h-px bg-amber-300 opacity-60" />
            <div className="absolute top-6 left-6 right-6 h-px bg-amber-300 opacity-40" />
            
            {/* 邮票区域 */}
            <div className="absolute top-2 right-2 w-12 h-8 bg-gradient-to-br from-red-400 to-red-500 rounded-sm flex items-center justify-center">
              <Mail className="w-4 h-4 text-white" />
            </div>
          </div>

          {/* 信封主体 */}
          <div className="absolute inset-0 top-24 bg-gradient-to-b from-amber-100 to-amber-50 flex items-center justify-center">
            <div className="text-center text-amber-800 opacity-60">
              <Mail className="w-12 h-12 mx-auto mb-2" />
              <div className="text-sm font-medium">点击拆开信件</div>
            </div>
          </div>

          {/* 信封底部装饰 */}
          <div className="absolute bottom-0 left-0 right-0 h-2 bg-amber-200" />
        </div>

        {/* 信件内容 */}
        <div className={`
          relative transition-all duration-800 ease-out transform-gpu
          ${isOpened ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'}
        `}>
          <CardContent className="p-6">
            {/* 信件标题区 */}
            <div className="mb-4">
              <div className="flex items-start justify-between mb-3">
                <div className="flex flex-wrap gap-2">
                  <span className="px-3 py-1 bg-amber-100 text-amber-800 text-xs rounded-full">
                    {letter.type}
                  </span>
                  {letter.featured && (
                    <span className="px-3 py-1 bg-gold-100 text-gold-800 text-xs rounded-full flex items-center gap-1">
                      ⭐ 精选
                    </span>
                  )}
                </div>
              </div>
              
              <h3 className="font-serif text-xl text-amber-900 line-clamp-2 mb-2">
                {letter.title}
              </h3>
              
              <div className="text-xs text-amber-600 italic">
                {letter.significance}
              </div>
            </div>

            {/* 信件内容预览 */}
            <div className="mb-4 p-4 bg-white/60 rounded-lg border-l-4 border-amber-300">
              <p className="text-amber-700 text-sm line-clamp-4 leading-relaxed font-serif">
                {letter.preview}
              </p>
            </div>
            
            {/* 元信息 */}
            <div className="space-y-2 mb-4 text-xs text-amber-600">
              <div className="flex items-center gap-4">
                <span className="flex items-center gap-1">
                  <Calendar className="w-3 h-3" />
                  {letter.date}
                </span>
                <span className="flex items-center gap-1">
                  <MapPin className="w-3 h-3" />
                  {letter.location}
                </span>
              </div>
              <div className="flex items-center gap-1">
                <span>作者：{letter.author}</span>
              </div>
            </div>

            {/* 统计和操作 */}
            <div className="flex items-center justify-between pt-3 border-t border-amber-200">
              <div className="flex items-center gap-3 text-xs text-amber-600">
                <span className="flex items-center gap-1">
                  <Heart className="w-3 h-3" />
                  {letter.likes}
                </span>
                <span className="flex items-center gap-1">
                  <Eye className="w-3 h-3" />
                  {letter.views}
                </span>
              </div>
              <div className="flex gap-2">
                <Button 
                  size="sm" 
                  variant="ghost" 
                  onClick={onReadLetter}
                  className="text-amber-700 hover:bg-amber-50 text-xs"
                >
                  阅读全文
                </Button>
                <Button 
                  size="sm" 
                  variant="outline" 
                  onClick={handleCloseEnvelope}
                  className="text-amber-700 border-amber-300 hover:bg-amber-50 text-xs"
                >
                  收起
                </Button>
              </div>
            </div>
          </CardContent>
        </div>

        {/* 点击遮罩（用于打开信封） */}
        {!isOpened && (
          <div 
            className="absolute inset-0 z-10 cursor-pointer"
            onClick={handleOpenEnvelope}
          />
        )}

        {/* 撕开效果装饰 */}
        {isOpened && (
          <>
            <div className="absolute top-0 left-0 w-8 h-8 bg-amber-100 transform rotate-45 -translate-x-2 -translate-y-2 opacity-60" />
            <div className="absolute top-0 right-0 w-6 h-6 bg-amber-100 transform -rotate-45 translate-x-1 -translate-y-1 opacity-60" />
            <div className="absolute top-4 left-8 w-4 h-4 bg-amber-200 transform rotate-12 opacity-40" />
          </>
        )}

        {/* 发光效果 */}
        {letter.featured && !isOpened && (
          <div className="absolute -inset-1 bg-gradient-to-r from-amber-300 via-orange-300 to-amber-300 rounded-lg opacity-20 blur-sm group-hover:opacity-40 transition-opacity duration-300" />
        )}
      </Card>
    </div>
  )
}