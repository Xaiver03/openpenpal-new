'use client'

import { useEffect } from 'react'
import { disableTestCourierMode } from '@/lib/auth/test-courier-mock'

export function DisableTestMode() {
  useEffect(() => {
    // 禁用测试模式
    disableTestCourierMode()
    
    // 清除所有测试相关的 localStorage
    if (typeof window !== 'undefined') {
      const keysToRemove = [
        'test_courier_mode',
        'test_courier_level',
        'mock_data_enabled',
        'use_test_data'
      ]
      
      keysToRemove.forEach(key => {
        localStorage.removeItem(key)
      })
      
      console.log('✅ 已禁用所有测试模式，现在使用真实数据库数据')
    }
  }, [])
  
  return null
}