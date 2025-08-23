'use client'

/**
 * Profile Privacy Integration - Seamless privacy controls within profile pages
 * 个人资料隐私集成 - 在个人资料页面中的无缝隐私控制
 */

import React, { useEffect, useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { 
  Shield, 
  Eye, 
  EyeOff, 
  Settings,
  UserMinus,
  Volume2,
  VolumeX,
  Info,
  ExternalLink
} from 'lucide-react'

import { privacyApi } from '../../lib/api/privacy'
import type { PrivacyCheckResult } from '../../types/privacy'
import { useAuth } from '@/contexts/auth-context-new'
import { toast } from 'sonner'

interface ProfilePrivacyIntegrationProps {
  targetUserId: string
  targetUsername: string
  className?: string
  showQuickActions?: boolean
}

export function ProfilePrivacyIntegration({ 
  targetUserId, 
  targetUsername,
  className,
  showQuickActions = true
}: ProfilePrivacyIntegrationProps) {
  const { user: currentUser } = useAuth()
  const [privacyStatus, setPrivacyStatus] = useState<PrivacyCheckResult | null>(null)
  const [isBlocked, setIsBlocked] = useState(false)
  const [isMuted, setIsMuted] = useState(false)
  const [loading, setLoading] = useState(false)

  // Don't show privacy controls for own profile
  if (!currentUser || currentUser.id === targetUserId) {
    return null
  }

  useEffect(() => {
    loadPrivacyStatus()
  }, [targetUserId, currentUser])

  const loadPrivacyStatus = async () => {
    try {
      setLoading(true)
      
      // Check privacy permissions
      const permissions = await privacyApi.checkPrivacy(targetUserId, 'view_profile')
      setPrivacyStatus(permissions)

      // Check if user is blocked/muted
      const blocked = await privacyApi.isUserBlocked(targetUserId)
      const muted = await privacyApi.isUserMuted(targetUserId)
      setIsBlocked(blocked)
      setIsMuted(muted)
    } catch (error) {
      console.error('Failed to load privacy status:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleBlockUser = async () => {
    try {
      await privacyApi.blockUser(targetUserId)
      setIsBlocked(true)
      toast.success(`已屏蔽 ${targetUsername}`)
    } catch (error) {
      console.error('Failed to block user:', error)
      toast.error('屏蔽用户失败')
    }
  }

  const handleUnblockUser = async () => {
    try {
      await privacyApi.unblockUser(targetUserId)
      setIsBlocked(false)
      toast.success(`已取消屏蔽 ${targetUsername}`)
    } catch (error) {
      console.error('Failed to unblock user:', error)
      toast.error('取消屏蔽失败')
    }
  }

  const handleMuteUser = async () => {
    try {
      await privacyApi.muteUser(targetUserId)
      setIsMuted(true)
      toast.success(`已静音 ${targetUsername}`)
    } catch (error) {
      console.error('Failed to mute user:', error)
      toast.error('静音用户失败')
    }
  }

  const handleUnmuteUser = async () => {
    try {
      await privacyApi.unmuteUser(targetUserId)
      setIsMuted(false)
      toast.success(`已取消静音 ${targetUsername}`)
    } catch (error) {
      console.error('Failed to unmute user:', error)
      toast.error('取消静音失败')
    }
  }

  if (loading) {
    return (
      <div className={`animate-pulse ${className}`}>
        <div className="h-8 w-24 bg-gray-200 rounded"></div>
      </div>
    )
  }

  // Show privacy restriction if user is blocked
  if (privacyStatus && !privacyStatus.can_view_profile) {
    return (
      <Alert className={className}>
        <Shield className="h-4 w-4" />
        <AlertDescription>
          由于隐私设置限制，您无法查看此用户的完整资料。
          {privacyStatus.reason && ` 原因：${privacyStatus.reason}`}
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <div className={`flex items-center gap-2 ${className}`}>
      {/* Privacy Status Indicators */}
      <div className="flex items-center gap-1">
        {isBlocked && (
          <Badge variant="destructive" className="text-xs">
            <UserMinus className="h-3 w-3 mr-1" />
            已屏蔽
          </Badge>
        )}
        {isMuted && (
          <Badge variant="outline" className="text-xs">
            <VolumeX className="h-3 w-3 mr-1" />
            已静音
          </Badge>
        )}
      </div>

      {/* Quick Actions */}
      {showQuickActions && (
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="outline" size="sm">
              <Shield className="h-4 w-4 mr-1" />
              隐私
            </Button>
          </SheetTrigger>
          <SheetContent>
            <SheetHeader>
              <SheetTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5" />
                隐私控制
              </SheetTitle>
              <SheetDescription>
                管理您与 {targetUsername} 的互动设置
              </SheetDescription>
            </SheetHeader>

            <div className="space-y-6 mt-6">
              {/* Privacy Information */}
              <div className="space-y-3">
                <h4 className="font-medium">当前权限状态</h4>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span>查看资料</span>
                    {privacyStatus?.can_view_profile ? (
                      <Eye className="h-4 w-4 text-green-500" />
                    ) : (
                      <EyeOff className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                  <div className="flex items-center justify-between">
                    <span>发送消息</span>
                    {privacyStatus?.can_message ? (
                      <Eye className="h-4 w-4 text-green-500" />
                    ) : (
                      <EyeOff className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                  <div className="flex items-center justify-between">
                    <span>关注用户</span>
                    {privacyStatus?.can_follow ? (
                      <Eye className="h-4 w-4 text-green-500" />
                    ) : (
                      <EyeOff className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                  <div className="flex items-center justify-between">
                    <span>发表评论</span>
                    {privacyStatus?.can_comment ? (
                      <Eye className="h-4 w-4 text-green-500" />
                    ) : (
                      <EyeOff className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                </div>
              </div>

              {/* Quick Actions */}
              <div className="space-y-3">
                <h4 className="font-medium">快速操作</h4>
                <div className="space-y-2">
                  {/* Block/Unblock */}
                  <Button
                    variant={isBlocked ? "outline" : "destructive"}
                    size="sm"
                    onClick={isBlocked ? handleUnblockUser : handleBlockUser}
                    className="w-full justify-start"
                  >
                    <UserMinus className="h-4 w-4 mr-2" />
                    {isBlocked ? '取消屏蔽' : '屏蔽用户'}
                  </Button>

                  {/* Mute/Unmute */}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={isMuted ? handleUnmuteUser : handleMuteUser}
                    className="w-full justify-start"
                  >
                    {isMuted ? (
                      <Volume2 className="h-4 w-4 mr-2" />
                    ) : (
                      <VolumeX className="h-4 w-4 mr-2" />
                    )}
                    {isMuted ? '取消静音' : '静音用户'}
                  </Button>
                </div>
              </div>

              {/* Advanced Settings Link */}
              <div className="pt-4 border-t">
                <Link href="/settings/privacy">
                  <Button variant="ghost" size="sm" className="w-full justify-start">
                    <Settings className="h-4 w-4 mr-2" />
                    高级隐私设置
                    <ExternalLink className="h-3 w-3 ml-auto" />
                  </Button>
                </Link>
              </div>

              {/* Privacy Info */}
              <Alert>
                <Info className="h-4 w-4" />
                <AlertDescription className="text-xs">
                  屏蔽用户后，您将无法看到对方的内容，对方也无法与您互动。
                  静音用户只会隐藏对方的通知，但不影响正常查看。
                </AlertDescription>
              </Alert>
            </div>
          </SheetContent>
        </Sheet>
      )}

      {/* Settings Link for Advanced Users */}
      <Link href="/settings/privacy">
        <Button variant="ghost" size="sm" className="text-xs">
          <Settings className="h-3 w-3" />
        </Button>
      </Link>
    </div>
  )
}

export default ProfilePrivacyIntegration