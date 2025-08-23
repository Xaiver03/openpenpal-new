'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { RefreshCw, LogIn, AlertTriangle } from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { TokenManager } from '@/lib/auth/cookie-token-manager'
import { getAuthSyncService } from '@/lib/auth/auth-sync-service'
import { toast } from 'sonner'

interface AuthFixBannerProps {
  onFixed?: () => void
}

export function AuthFixBanner({ onFixed }: AuthFixBannerProps) {
  const router = useRouter()
  const [isFixing, setIsFixing] = useState(false)
  const { refreshUser, logout } = useAuth()

  const handleQuickFix = async () => {
    setIsFixing(true)
    
    try {
      // 使用强化的认证同步服务进行修复
      const success = await getAuthSyncService().fixAuth()
      
      if (success) {
        toast.success('认证状态已修复！')
        onFixed?.()
      } else {
        toast.error('快速修复失败，请重新登录')
        handleForceRelogin()
      }
    } catch (error) {
      console.error('Quick fix failed:', error)
      toast.error('快速修复失败，请重新登录')
      handleForceRelogin()
    } finally {
      setIsFixing(false)
    }
  }

  const handleForceRelogin = async () => {
    TokenManager.clear()
    await logout()
    router.push('/login')
  }

  return (
    <Alert variant="destructive" className="mb-6">
      <AlertTriangle className="h-4 w-4" />
      <AlertDescription className="flex items-center justify-between">
        <div>
          <strong>认证状态异常</strong>
          <p className="text-sm mt-1">
            检测到认证token可能已过期或损坏，这会导致AI功能无法正常使用。
          </p>
        </div>
        <div className="flex gap-2 ml-4">
          <Button
            size="sm"
            variant="outline"
            onClick={handleQuickFix}
            disabled={isFixing}
            className="gap-2"
          >
            {isFixing ? (
              <>
                <RefreshCw className="h-3 w-3 animate-spin" />
                修复中...
              </>
            ) : (
              <>
                <RefreshCw className="h-3 w-3" />
                快速修复
              </>
            )}
          </Button>
          <Button
            size="sm"
            onClick={handleForceRelogin}
            className="gap-2"
          >
            <LogIn className="h-3 w-3" />
            重新登录
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  )
}