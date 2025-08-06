/**
 * 用户权限展示组件
 * 展示当前用户的权限详情和状态
 */

'use client'

import React, { useState } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { Alert, AlertDescription } from '@/components/ui/alert'
// Using custom collapsible implementation instead of radix
import { 
  ChevronDown, 
  Shield, 
  User, 
  Crown, 
  AlertTriangle,
  CheckCircle2,
  Info,
  RefreshCw
} from 'lucide-react'
import { usePermissions } from '@/hooks/use-permissions'
import { PermissionCategory, RiskLevel } from '@/lib/permissions/permission-modules'

interface PermissionDisplayProps {
  className?: string
  showDetails?: boolean
}

export function PermissionDisplay({ className, showDetails = true }: PermissionDisplayProps) {
  const {
    permissionSummary,
    permissionStatuses,
    getPermissionsByCategory,
    hasHighRiskPermissions,
    hasMissingDependencies,
    refreshPermissions,
    loading,
    lastRefresh
  } = usePermissions()

  const [expandedCategories, setExpandedCategories] = useState<Set<PermissionCategory>>(new Set())

  const toggleCategory = (category: PermissionCategory) => {
    const newExpanded = new Set(expandedCategories)
    if (newExpanded.has(category)) {
      newExpanded.delete(category)
    } else {
      newExpanded.add(category)
    }
    setExpandedCategories(newExpanded)
  }

  const getRiskLevelColor = (risk: RiskLevel) => {
    switch (risk) {
      case 'low': return 'bg-green-100 text-green-800'
      case 'medium': return 'bg-yellow-100 text-yellow-800'
      case 'high': return 'bg-orange-100 text-orange-800'
      case 'critical': return 'bg-red-100 text-red-800'
    }
  }

  const getCategoryIcon = (category: PermissionCategory) => {
    switch (category) {
      case 'basic': return '📝'
      case 'courier': return '📮'
      case 'management': return '👥'
      case 'admin': return '🛡️'
      case 'system': return '⚙️'
    }
  }

  const getCategoryName = (category: PermissionCategory) => {
    switch (category) {
      case 'basic': return '基础功能'
      case 'courier': return '信使功能'
      case 'management': return '管理功能'
      case 'admin': return '管理员功能'
      case 'system': return '系统功能'
    }
  }

  if (loading && !permissionSummary) {
    return (
      <Card className={className}>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center">
            <RefreshCw className="h-4 w-4 animate-spin mr-2" />
            加载权限信息...
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className={className}>
      {/* 权限概览 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Shield className="h-5 w-5" />
              <CardTitle>我的权限</CardTitle>
            </div>
            <Button 
              variant="outline" 
              size="sm" 
              onClick={refreshPermissions} 
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 mr-1 ${loading ? 'animate-spin' : ''}`} />
              刷新
            </Button>
          </div>
          {lastRefresh && (
            <p className="text-sm text-gray-600">
              最后更新: {lastRefresh.toLocaleString()}
            </p>
          )}
        </CardHeader>

        <CardContent className="space-y-6">
          {/* 权限统计 */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">{permissionSummary.granted}</div>
              <div className="text-sm text-gray-600">已授权</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-600">{permissionSummary.total}</div>
              <div className="text-sm text-gray-600">权限总数</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {Math.round((permissionSummary.granted / permissionSummary.total) * 100)}%
              </div>
              <div className="text-sm text-gray-600">权限覆盖率</div>
            </div>
          </div>

          {/* 权限进度条 */}
          <div>
            <div className="flex justify-between text-sm mb-2">
              <span>权限覆盖率</span>
              <span>{permissionSummary.granted}/{permissionSummary.total}</span>
            </div>
            <Progress 
              value={(permissionSummary.granted / permissionSummary.total) * 100} 
              className="h-2"
            />
          </div>

          {/* 警告信息 */}
          {hasHighRiskPermissions() && (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                您拥有 {permissionSummary.highRiskPermissions.length} 个高风险权限，请谨慎使用。
              </AlertDescription>
            </Alert>
          )}

          {hasMissingDependencies() && (
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                检测到 {permissionSummary.missingDependencies.length} 个权限依赖缺失，可能影响功能正常使用。
              </AlertDescription>
            </Alert>
          )}

          {showDetails && (
            <>
              <Separator />

              {/* 按分类展示权限 */}
              <div className="space-y-4">
                <h3 className="font-semibold">权限详情</h3>
                
                {Object.entries(permissionSummary.byCategory).map(([category, stats]) => {
                  const categoryPermissions = getPermissionsByCategory(category as PermissionCategory)
                  const isExpanded = expandedCategories.has(category as PermissionCategory)
                  
                  return (
                    <div key={category}>
                      <div>
                        <Button
                          variant="ghost"
                          className="w-full justify-between p-3 h-auto"
                          onClick={() => toggleCategory(category as PermissionCategory)}
                        >
                          <div className="flex items-center space-x-3">
                            <span className="text-lg">{getCategoryIcon(category as PermissionCategory)}</span>
                            <div className="text-left">
                              <div className="font-medium">{getCategoryName(category as PermissionCategory)}</div>
                              <div className="text-sm text-gray-600">
                                {stats.granted}/{stats.total} 个权限
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center space-x-2">
                            <Progress 
                              value={stats.total > 0 ? (stats.granted / stats.total) * 100 : 0} 
                              className="w-16 h-2"
                            />
                            <ChevronDown className={`h-4 w-4 transition-transform ${isExpanded ? 'rotate-180' : ''}`} />
                          </div>
                        </Button>
                      </div>
                      
                      {isExpanded && (<div className="mt-2">)
                        <div className="pl-4 space-y-2">
                          {categoryPermissions.map(permission => (
                            <div 
                              key={permission.id} 
                              className="flex items-center justify-between p-2 rounded border-l-2 border-gray-200"
                            >
                              <div className="flex items-center space-x-3">
                                {permission.granted ? (
                                  <CheckCircle2 className="h-4 w-4 text-green-600" />
                                ) : (
                                  <div className="h-4 w-4 rounded-full border-2 border-gray-300" />
                                )}
                                <div>
                                  <div className="font-medium text-sm">{permission.name}</div>
                                  <div className="text-xs text-gray-600">{permission.description}</div>
                                </div>
                              </div>
                              <div className="flex items-center space-x-2">
                                <Badge variant="outline" className={getRiskLevelColor(permission.riskLevel)}>
                                  {permission.riskLevel}
                                </Badge>
                                <Badge variant="secondary" className="text-xs">
                                  {permission.source === 'role' ? '角色' :
                                   permission.source === 'courier_level' ? '信使' : '自定义'}
                                </Badge>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>)}
                    </div>
                  )
                })}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default PermissionDisplay