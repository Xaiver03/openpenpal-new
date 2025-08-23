'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent } from '@/components/ui/card'
import { Loader2 } from 'lucide-react'

// 重定向组件 - 将用户重定向到新的设置页面
export default function ProfileRedirectPage() {
  const router = useRouter()

  useEffect(() => {
    // 立即重定向到设置页面的个人资料标签页
    router.replace('/settings?tab=profile')
  }, [router])

  return (
    <div className="container max-w-2xl mx-auto px-4 py-8">
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="text-center">
            <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
            <p className="text-muted-foreground">正在跳转到设置页面...</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}