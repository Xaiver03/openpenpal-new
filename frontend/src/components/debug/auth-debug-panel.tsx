/**
 * 认证调试面板 - 用于调试认证状态问题
 * Auth Debug Panel - For debugging authentication state issues
 */

'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { AuthStateFixer } from '@/lib/auth/auth-state-fixer'
import { getAuthSyncService } from '@/lib/auth/auth-sync-service'
import { TokenManager } from '@/lib/auth/cookie-token-manager'
import { useUserStore } from '@/stores/user-store'
import { JWTUtils } from '@/lib/auth/jwt-utils'

export function AuthDebugPanel() {
  const [isVisible, setIsVisible] = useState(false)
  const [diagReport, setDiagReport] = useState<string>('')
  const [fixResults, setFixResults] = useState<any>(null)
  
  const store = useUserStore()
  
  // 只在开发环境显示
  if (process.env.NODE_ENV !== 'development') {
    return null
  }
  
  const refreshReport = () => {
    const report = AuthStateFixer.generateDiagnosticReport()
    setDiagReport(report)
  }
  
  const handleAutoFix = async () => {
    const result = await AuthStateFixer.autoFix()
    setFixResults(result)
    refreshReport()
  }
  
  const handleForceReauth = async () => {
    const success = await AuthStateFixer.forceReauth()
    setFixResults({ 
      success, 
      message: success ? '强制重新认证成功' : '强制重新认证失败',
      actions: ['强制重新认证']
    })
    refreshReport()
  }
  
  const handleSyncFix = async () => {
    const success = await getAuthSyncService().fixAuth()
    setFixResults({ 
      success, 
      message: success ? '同步服务修复成功' : '同步服务修复失败',
      actions: ['同步服务修复']
    })
    refreshReport()
  }
  
  const handleClearAll = () => {
    TokenManager.clear()
    store.reset()
    localStorage.clear()
    setFixResults({ 
      success: true, 
      message: '已清理所有认证状态',
      actions: ['清理Token', '重置Store', '清理LocalStorage']
    })
    refreshReport()
  }
  
  const getTokenInfo = () => {
    const token = TokenManager.get()
    if (!token) return null
    
    try {
      const payload = JWTUtils.decodeToken(token)
      const isExpired = JWTUtils.isTokenExpired(token)
      const ttl = JWTUtils.getTokenTimeToLive(token)
      
      return {
        payload,
        isExpired,
        ttl,
        expiresAt: payload?.exp ? new Date(payload.exp * 1000).toLocaleString() : 'N/A'
      }
    } catch (error) {
      return { error: error instanceof Error ? error.message : String(error) }
    }
  }
  
  useEffect(() => {
    refreshReport()
  }, [])
  
  if (!isVisible) {
    return (
      <div className="fixed bottom-4 right-4 z-50">
        <Button 
          onClick={() => setIsVisible(true)}
          variant="outline" 
          size="sm"
          className="bg-orange-500 text-white hover:bg-orange-600"
        >
          🔧 Auth Debug
        </Button>
      </div>
    )
  }
  
  const tokenInfo = getTokenInfo()
  const state = AuthStateFixer.checkAuthState()
  
  return (
    <div className="fixed inset-0 z-50 bg-black bg-opacity-50 flex items-center justify-center p-4">
      <Card className="w-full max-w-4xl max-h-[90vh] overflow-auto">
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            🔧 认证状态调试面板
            <Button 
              onClick={() => setIsVisible(false)}
              variant="outline" 
              size="sm"
            >
              关闭
            </Button>
          </CardTitle>
        </CardHeader>
        
        <CardContent className="space-y-6">
          {/* 状态概览 */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-sm text-gray-500">Token状态</div>
              <Badge variant={state.hasToken && state.tokenValid ? 'default' : 'destructive'}>
                {state.hasToken ? (state.tokenValid ? '有效' : '无效') : '无Token'}
              </Badge>
            </div>
            <div className="text-center">
              <div className="text-sm text-gray-500">用户数据</div>
              <Badge variant={state.hasUser ? 'default' : 'secondary'}>
                {state.hasUser ? '存在' : '不存在'}
              </Badge>
            </div>
            <div className="text-center">
              <div className="text-sm text-gray-500">Store状态</div>
              <Badge variant={state.storeAuthenticated ? 'default' : 'secondary'}>
                {state.storeAuthenticated ? '已认证' : '未认证'}
              </Badge>
            </div>
            <div className="text-center">
              <div className="text-sm text-gray-500">一致性</div>
              <Badge variant={state.consistent ? 'default' : 'destructive'}>
                {state.consistent ? '一致' : '不一致'}
              </Badge>
            </div>
          </div>
          
          {/* Token信息 */}
          {tokenInfo && (
            <div>
              <h3 className="font-semibold mb-2">Token信息</h3>
              {tokenInfo.error ? (
                <div className="text-red-500">错误: {tokenInfo.error}</div>
              ) : (
                <div className="bg-gray-50 p-3 rounded text-sm space-y-1">
                  <div><strong>用户ID:</strong> {tokenInfo.payload?.userId}</div>
                  <div><strong>角色:</strong> {tokenInfo.payload?.role}</div>
                  <div><strong>过期时间:</strong> {tokenInfo.expiresAt}</div>
                  <div><strong>是否过期:</strong> {tokenInfo.isExpired ? '是' : '否'}</div>
                  <div><strong>剩余时间:</strong> {tokenInfo.ttl}秒</div>
                </div>
              )}
            </div>
          )}
          
          {/* 问题列表 */}
          {state.issues.length > 0 && (
            <div>
              <h3 className="font-semibold mb-2">发现的问题</h3>
              <ul className="list-disc list-inside space-y-1 text-sm">
                {state.issues.map((issue, index) => (
                  <li key={index} className="text-red-600">{issue}</li>
                ))}
              </ul>
            </div>
          )}
          
          {/* 修复结果 */}
          {fixResults && (
            <div>
              <h3 className="font-semibold mb-2">修复结果</h3>
              <div className={`p-3 rounded ${fixResults.success ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'}`}>
                <div><strong>状态:</strong> {fixResults.success ? '成功' : '失败'}</div>
                <div><strong>消息:</strong> {fixResults.message}</div>
                {fixResults.actions && fixResults.actions.length > 0 && (
                  <div><strong>操作:</strong> {fixResults.actions.join(', ')}</div>
                )}
              </div>
            </div>
          )}
          
          {/* 操作按钮 */}
          <div className="flex gap-2 flex-wrap">
            <Button onClick={refreshReport} variant="outline" size="sm">
              刷新状态
            </Button>
            <Button onClick={handleAutoFix} variant="default" size="sm">
              自动修复
            </Button>
            <Button onClick={handleForceReauth} variant="secondary" size="sm">
              强制重认证
            </Button>
            <Button onClick={handleSyncFix} variant="secondary" size="sm">
              同步修复
            </Button>
            <Button onClick={handleClearAll} variant="destructive" size="sm">
              清理所有
            </Button>
          </div>
          
          {/* 详细报告 */}
          <div>
            <h3 className="font-semibold mb-2">详细诊断报告</h3>
            <pre className="bg-gray-50 p-3 rounded text-xs overflow-auto max-h-60">
              {diagReport}
            </pre>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}