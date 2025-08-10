/**
 * CommentDemo - Demo component for testing comment system
 * 评论系统演示组件 - 用于测试评论功能
 */

'use client'

import React from 'react'
import { CommentList, CommentStats } from '@/components/comments'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface CommentDemoProps {
  letter_id?: string
}

export default function CommentDemo({ letter_id = 'demo-letter-123' }: CommentDemoProps) {
  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>评论系统演示</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Comment Stats Example */}
          <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
            <h3 className="text-lg font-medium">信件标题</h3>
            <CommentStats letter_id={letter_id} format="full" />
          </div>

          {/* Main Comment List */}
          <CommentList
            letter_id={letter_id}
            max_depth={3}
            enable_nested={true}
            show_stats={true}
            allow_comments={true}
            initial_sort="created_at"
          />
        </CardContent>
      </Card>
    </div>
  )
}