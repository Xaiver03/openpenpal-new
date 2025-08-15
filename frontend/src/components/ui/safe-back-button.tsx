'use client'

import { useState } from 'react'
import { BackButton } from './back-button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

interface SafeBackButtonProps {
  href?: string
  label?: string
  className?: string
  variant?: 'default' | 'outline' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  hasUnsavedChanges?: boolean
  onConfirmLeave?: () => void
  confirmTitle?: string
  confirmDescription?: string
}

export function SafeBackButton({
  href,
  label = '返回上一级',
  className,
  variant = 'ghost',
  size = 'sm',
  hasUnsavedChanges = false,
  onConfirmLeave,
  confirmTitle = '确认离开',
  confirmDescription = '您有未保存的更改，确定要离开吗？所有未保存的更改将会丢失。'
}: SafeBackButtonProps) {
  const [showConfirmDialog, setShowConfirmDialog] = useState(false)

  const handleBackClick = () => {
    if (hasUnsavedChanges) {
      setShowConfirmDialog(true)
    } else {
      onConfirmLeave?.()
    }
  }

  const handleConfirmLeave = () => {
    setShowConfirmDialog(false)
    onConfirmLeave?.()
  }

  return (
    <>
      <BackButton
        href={hasUnsavedChanges ? undefined : href}
        label={label}
        className={className}
        variant={variant}
        size={size}
        onClick={hasUnsavedChanges ? handleBackClick : undefined}
      />

      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{confirmTitle}</AlertDialogTitle>
            <AlertDialogDescription>
              {confirmDescription}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setShowConfirmDialog(false)}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction 
              onClick={handleConfirmLeave}
              className="bg-red-600 hover:bg-red-700"
            >
              确认离开
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}