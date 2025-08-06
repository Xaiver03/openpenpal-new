'use client'

import { useEffect, useRef, useCallback } from 'react'
import { useRouter } from 'next/navigation'

interface UseUnsavedChangesOptions {
  hasUnsavedChanges: boolean
  message?: string
  onSave?: () => Promise<void> | void
  onDiscard?: () => void
}

/**
 * Hook to detect and warn about unsaved changes before page navigation
 * 检测未保存更改并在页面导航前警告的 Hook
 */
export function useUnsavedChanges({
  hasUnsavedChanges,
  message = '您有未保存的更改。是否要在离开前保存？',
  onSave,
  onDiscard
}: UseUnsavedChangesOptions) {
  const router = useRouter()
  const isLeavingRef = useRef(false)
  
  // 浏览器刷新/关闭警告
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedChanges && !isLeavingRef.current) {
        e.preventDefault()
        e.returnValue = message
        return message
      }
    }
    
    if (hasUnsavedChanges) {
      window.addEventListener('beforeunload', handleBeforeUnload)
      return () => window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [hasUnsavedChanges, message])
  
  // 确认离开的方法
  const confirmLeave = useCallback(async (): Promise<boolean> => {
    if (!hasUnsavedChanges) return true
    
    const userChoice = await showUnsavedDialog()
    
    if (userChoice === 'save' && onSave) {
      try {
        await onSave()
        return true
      } catch (error) {
        console.error('保存失败:', error)
        return false
      }
    } else if (userChoice === 'discard') {
      if (onDiscard) onDiscard()
      return true
    }
    
    return false // 用户选择了取消
  }, [hasUnsavedChanges, onSave, onDiscard])
  
  // 显示未保存更改对话框
  const showUnsavedDialog = (): Promise<'save' | 'discard' | 'cancel'> => {
    return new Promise((resolve) => {
      const dialog = document.createElement('div')
      dialog.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50'
      dialog.innerHTML = `
        <div class="bg-white rounded-lg p-6 max-w-md mx-4 shadow-xl">
          <div class="flex items-center gap-3 mb-4">
            <svg class="w-6 h-6 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
            <h3 class="text-lg font-semibold text-gray-900">未保存的更改</h3>
          </div>
          <p class="text-gray-600 mb-6">${message}</p>
          <div class="flex justify-end gap-3">
            <button id="cancel-btn" class="px-4 py-2 text-gray-500 hover:text-gray-700 transition-colors">
              取消
            </button>
            <button id="discard-btn" class="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 transition-colors">
              不保存
            </button>
            <button id="save-btn" class="px-4 py-2 bg-amber-600 text-white rounded hover:bg-amber-700 transition-colors">
              保存并离开
            </button>
          </div>
        </div>
      `
      
      document.body.appendChild(dialog)
      
      const cleanup = () => document.body.removeChild(dialog)
      
      dialog.querySelector('#save-btn')?.addEventListener('click', () => {
        cleanup()
        resolve('save')
      })
      
      dialog.querySelector('#discard-btn')?.addEventListener('click', () => {
        cleanup()
        resolve('discard')
      })
      
      dialog.querySelector('#cancel-btn')?.addEventListener('click', () => {
        cleanup()
        resolve('cancel')
      })
      
      // 点击背景关闭
      dialog.addEventListener('click', (e) => {
        if (e.target === dialog) {
          cleanup()
          resolve('cancel')
        }
      })
      
      // ESC 键关闭
      const handleEsc = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
          document.removeEventListener('keydown', handleEsc)
          cleanup()
          resolve('cancel')
        }
      }
      document.addEventListener('keydown', handleEsc)
    })
  }
  
  // 安全导航方法
  const safeNavigate = useCallback(async (url: string) => {
    const canLeave = await confirmLeave()
    if (canLeave) {
      isLeavingRef.current = true
      router.push(url)
    }
  }, [confirmLeave, router])
  
  // 创建安全的后退按钮
  const createSafeBackButton = useCallback((href?: string) => {
    return {
      onClick: async (e: React.MouseEvent) => {
        e.preventDefault()
        const canLeave = await confirmLeave()
        if (canLeave) {
          isLeavingRef.current = true
          if (href) {
            router.push(href)
          } else {
            router.back()
          }
        }
      }
    }
  }, [confirmLeave, router])
  
  return {
    hasUnsavedChanges,
    confirmLeave,
    safeNavigate,
    createSafeBackButton
  }
}