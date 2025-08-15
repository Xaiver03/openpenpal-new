'use client'

import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useRouter } from 'next/navigation'
import { cn } from '@/lib/utils'

interface BackButtonProps {
  href?: string
  label?: string
  className?: string
  variant?: 'default' | 'outline' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  onClick?: () => void
}

export function BackButton({ 
  href, 
  label = '返回上一级', 
  className,
  variant = 'ghost',
  size = 'sm',
  onClick 
}: BackButtonProps) {
  const router = useRouter()

  const handleClick = () => {
    if (onClick) {
      onClick()
    } else if (href) {
      router.push(href)
    } else {
      router.back()
    }
  }

  return (
    <Button
      variant={variant}
      size={size}
      onClick={handleClick}
      className={cn(
        'flex items-center gap-2 mb-4 hover:bg-gray-100 transition-colors',
        className
      )}
    >
      <ArrowLeft className="h-4 w-4" />
      <span className="hidden sm:inline">{label}</span>
    </Button>
  )
}