'use client'

import React, { useState } from 'react'
import { useUserStore } from '@/stores/user-store'
import { validateOPCodeAccess, getManagedPrefixes, getPermissionDescription } from '@/lib/utils/courier-permission-utils'

interface CourierPermissionDebugProps {
  testOPCodes?: string[]
}

export function CourierPermissionDebug({ 
  testOPCodes = ['PK5F3D', 'PK5F01', 'PK9A01', 'QH1A01', 'BD2B02'] 
}: CourierPermissionDebugProps) {
  const { user } = useUserStore()
  const [isVisible, setIsVisible] = useState(true)

  if (!user || !user.role?.includes('courier_level')) {
    return null
  }

  // 安全解析信使等级 - 后端角色格式是 courier_level4 而不是 courier_level_4
  let courierLevel = 0
  if (user.role?.includes('courier_level')) {
    const levelStr = user.role.replace('courier_level', '')
    const parsed = parseInt(levelStr)
    courierLevel = isNaN(parsed) ? 0 : parsed
  }
  
  const courierInfo = {
    id: user.id || '',
    level: courierLevel,
    managedOPCodePrefix: user.managed_op_code_prefix || '',
    zoneCode: user.zone_code || ''
  }

  const managedPrefixes = getManagedPrefixes(courierInfo)
  const permissionDesc = getPermissionDescription(courierInfo)

  if (process.env.NODE_ENV !== 'development') {
    return null
  }

  if (!isVisible) {
    return (
      <button
        onClick={() => setIsVisible(true)}
        className="fixed bottom-4 left-4 bg-blue-600 text-white px-3 py-2 rounded-full shadow-lg z-50 text-xs hover:bg-blue-700"
      >
        🐛 Debug
      </button>
    )
  }

  return (
    <div className="fixed bottom-4 left-4 bg-white border rounded-lg shadow-lg p-4 max-w-sm z-50 text-xs">
      <div className="flex items-center justify-between mb-2">
        <div className="font-bold text-blue-600">🐛 信使权限调试</div>
        <button
          onClick={() => setIsVisible(false)}
          className="text-gray-500 hover:text-gray-700 text-lg leading-none"
        >
          ×
        </button>
      </div>
      
      <div className="space-y-2">
        <div>
          <span className="font-semibold">用户:</span> {user.username}
        </div>
        <div>
          <span className="font-semibold">角色:</span> {user.role}
        </div>
        <div>
          <span className="font-semibold">等级:</span> L{courierLevel}
        </div>
        <div>
          <span className="font-semibold">管理前缀:</span> {user.managed_op_code_prefix || '未设置'}
        </div>
        <div>
          <span className="font-semibold">区域代码:</span> {user.zone_code || '未设置'}
        </div>
        <div>
          <span className="font-semibold">权限描述:</span> {permissionDesc}
        </div>
        <div>
          <span className="font-semibold">管理范围:</span> {managedPrefixes.join(', ')}
        </div>
      </div>

      <div className="mt-3 border-t pt-3">
        <div className="font-semibold mb-2">OP Code权限测试:</div>
        <div className="space-y-1">
          {testOPCodes.map(code => {
            const permissions = validateOPCodeAccess(courierInfo, code)
            return (
              <div key={code} className="flex items-center justify-between text-xs">
                <span className="font-mono">{code}</span>
                <div className="flex gap-1">
                  <span className={permissions.canView ? 'text-green-600' : 'text-red-600'}>
                    {permissions.canView ? '👁' : '❌'}
                  </span>
                  <span className={permissions.canEdit ? 'text-green-600' : 'text-red-600'}>
                    {permissions.canEdit ? '✏️' : '❌'}
                  </span>
                  <span className={permissions.canCreate ? 'text-green-600' : 'text-red-600'}>
                    {permissions.canCreate ? '➕' : '❌'}
                  </span>
                  <span className={permissions.canDelete ? 'text-green-600' : 'text-red-600'}>
                    {permissions.canDelete ? '🗑' : '❌'}
                  </span>
                  <span className={permissions.canBatch ? 'text-green-600' : 'text-red-600'}>
                    {permissions.canBatch ? '📦' : '❌'}
                  </span>
                </div>
              </div>
            )
          })}
        </div>
        <div className="mt-2 text-xs text-gray-500">
          👁=查看 ✏️=编辑 ➕=创建 🗑=删除 📦=批量
        </div>
      </div>
    </div>
  )
}