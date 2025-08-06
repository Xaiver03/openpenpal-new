'use client'

import { Button } from './button'
import { ArrowLeft } from 'lucide-react'
import { useRouter } from 'next/navigation'

interface SafeBackButtonProps {
  href?: string
  children?: React.ReactNode
  className?: string
  variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
  size?: "default" | "sm" | "lg" | "icon"
  onBeforeBack?: () => Promise<boolean> | boolean
  hasUnsavedChanges?: boolean
  unsavedMessage?: string
}

export function SafeBackButton({ 
  href, 
  children = '返回上一级', 
  className = '',
  variant = 'outline',
  size = 'sm',
  onBeforeBack,
  hasUnsavedChanges = false,
  unsavedMessage = '您有未保存的更改。确定要离开吗？'
}: SafeBackButtonProps) {
  const router = useRouter()
  
  const handleBack = async () => {
    // 如果有未保存更改，显示确认对话框
    if (hasUnsavedChanges) {
      const shouldLeave = window.confirm(unsavedMessage)
      if (!shouldLeave) return
    }
    
    // 如果有自定义的前置检查
    if (onBeforeBack) {
      const canProceed = await onBeforeBack()
      if (!canProceed) return
    }
    
    // 执行导航
    if (href) {
      router.push(href)
    } else {
      router.back()
    }
  }
  
  return (
    <Button
      variant={variant}
      size={size}
      onClick={handleBack}
      className={`flex items-center gap-2 ${className}`}
    >
      <ArrowLeft className="w-4 h-4" />
      {children}
    </Button>
  )
}