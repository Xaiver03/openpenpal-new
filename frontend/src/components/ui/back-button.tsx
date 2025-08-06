'use client'

import { Button } from './button'
import { ArrowLeft } from 'lucide-react'
import { useRouter } from 'next/navigation'

interface BackButtonProps {
  href?: string
  children?: React.ReactNode
  className?: string
  variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
  size?: "default" | "sm" | "lg" | "icon"
}

export function BackButton({ 
  href, 
  children = '返回上一级', 
  className = '',
  variant = 'outline',
  size = 'sm'
}: BackButtonProps) {
  const router = useRouter()
  
  const handleBack = () => {
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