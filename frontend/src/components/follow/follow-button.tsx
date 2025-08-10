/**
 * FollowButton Component - SOTA Implementation
 * 关注/取消关注按钮 - 支持乐观更新、加载状态、错误处理
 */

'use client'

import React from 'react'
import { Button } from '@/components/ui/button'
import { 
  UserPlus, 
  UserMinus, 
  Loader2,
  Heart,
  UserCheck,
  Users
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useFollow, useFollowActions, useFollowStatus } from '@/stores/follow-store'
import { useUser } from '@/stores/user-store'
import { toast } from 'sonner'
import type { FollowButtonProps } from '@/types/follow'

interface FollowButtonState {
  isLoading: boolean
  isFollowing: boolean
  followerCount: number
}

export function FollowButton({
  user_id,
  initial_is_following = false,
  initial_follower_count = 0,
  size = 'md',
  variant = 'default',
  show_count = false,
  disabled = false,
  className,
  onFollowChange,
}: FollowButtonProps) {
  const { user: currentUser } = useUser()
  const { followUser, unfollowUser } = useFollowActions()
  const { isLoading } = useFollow()
  const { isFollowing: storeIsFollowing } = useFollowStatus(user_id)
  
  // Local state for optimistic updates
  const [localState, setLocalState] = React.useState<FollowButtonState>({
    isLoading: false,
    isFollowing: initial_is_following,
    followerCount: initial_follower_count,
  })
  
  // Sync with store state when available
  React.useEffect(() => {
    if (storeIsFollowing !== undefined) {
      setLocalState(prev => ({
        ...prev,
        isFollowing: storeIsFollowing,
      }))
    }
  }, [storeIsFollowing])
  
  // Don't show follow button for current user
  if (currentUser?.id === user_id) {
    return null
  }
  
  const handleFollowToggle = async () => {
    if (!currentUser) {
      toast.error('请先登录')
      return
    }
    
    if (disabled || localState.isLoading) return
    
    const wasFollowing = localState.isFollowing
    const previousCount = localState.followerCount
    
    // Optimistic update
    setLocalState(prev => ({
      ...prev,
      isLoading: true,
      isFollowing: !wasFollowing,
      followerCount: wasFollowing ? prev.followerCount - 1 : prev.followerCount + 1,
    }))
    
    // Notify parent component
    onFollowChange?.(!wasFollowing, wasFollowing ? previousCount - 1 : previousCount + 1)
    
    try {
      if (wasFollowing) {
        await unfollowUser(user_id)
        toast.success('已取消关注')
      } else {
        await followUser(user_id)
        toast.success('关注成功')
      }
    } catch (error) {
      // Rollback optimistic update
      setLocalState(prev => ({
        ...prev,
        isFollowing: wasFollowing,
        followerCount: previousCount,
      }))
      
      // Rollback parent notification
      onFollowChange?.(wasFollowing, previousCount)
      
      // Show error
      const errorMessage = error instanceof Error ? error.message : '操作失败'
      toast.error(errorMessage)
    } finally {
      setLocalState(prev => ({ ...prev, isLoading: false }))
    }
  }
  
  const getButtonContent = () => {
    if (localState.isLoading || isLoading.follow_action) {
      return (
        <>
          <Loader2 className="h-4 w-4 animate-spin" />
          {size !== 'sm' && <span className="ml-2">处理中...</span>}
        </>
      )
    }
    
    if (localState.isFollowing) {
      switch (variant) {
        case 'outline':
          return (
            <>
              <UserCheck className="h-4 w-4" />
              {size !== 'sm' && <span className="ml-2">已关注</span>}
              {show_count && localState.followerCount > 0 && (
                <span className="ml-1">({localState.followerCount})</span>
              )}
            </>
          )
        case 'ghost':
          return (
            <>
              <Heart className="h-4 w-4 fill-current" />
              {size !== 'sm' && <span className="ml-2">关注中</span>}
              {show_count && localState.followerCount > 0 && (
                <span className="ml-1">{localState.followerCount}</span>
              )}
            </>
          )
        default:
          return (
            <>
              <UserMinus className="h-4 w-4" />
              {size !== 'sm' && <span className="ml-2">取消关注</span>}
              {show_count && localState.followerCount > 0 && (
                <span className="ml-1">({localState.followerCount})</span>
              )}
            </>
          )
      }
    } else {
      return (
        <>
          <UserPlus className="h-4 w-4" />
          {size !== 'sm' && <span className="ml-2">关注</span>}
          {show_count && localState.followerCount > 0 && (
            <span className="ml-1">({localState.followerCount})</span>
          )}
        </>
      )
    }
  }
  
  const getButtonVariant = () => {
    if (localState.isFollowing && variant === 'default') {
      return 'outline'
    }
    return variant
  }
  
  const getButtonSize = () => {
    switch (size) {
      case 'sm': return 'sm'
      case 'lg': return 'lg'
      default: return 'default'
    }
  }
  
  const getHoverClasses = () => {
    if (localState.isFollowing) {
      switch (variant) {
        case 'ghost':
          return 'hover:bg-red-50 hover:text-red-600 hover:border-red-200'
        case 'outline':
          return 'hover:bg-red-50 hover:text-red-600 hover:border-red-300'
        default:
          return 'hover:bg-red-600 hover:border-red-600'
      }
    }
    return ''
  }
  
  return (
    <Button
      variant={getButtonVariant()}
      size={getButtonSize()}
      onClick={handleFollowToggle}
      disabled={disabled || localState.isLoading || isLoading.follow_action}
      className={cn(
        'transition-all duration-200',
        localState.isFollowing && 'group',
        getHoverClasses(),
        className
      )}
    >
      {getButtonContent()}
    </Button>
  )
}

/**
 * Compact FollowButton for use in cards or limited space
 */
export function CompactFollowButton({
  user_id,
  initial_is_following = false,
  className,
  onFollowChange,
}: Pick<FollowButtonProps, 'user_id' | 'initial_is_following' | 'className' | 'onFollowChange'>) {
  return (
    <FollowButton
      user_id={user_id}
      initial_is_following={initial_is_following}
      size="sm"
      variant="ghost"
      className={cn('h-8 px-2', className)}
      onFollowChange={onFollowChange}
    />
  )
}

/**
 * FollowButton with count display
 */
export function FollowButtonWithCount({
  user_id,
  initial_is_following = false,
  initial_follower_count = 0,
  className,
  onFollowChange,
}: Pick<FollowButtonProps, 'user_id' | 'initial_is_following' | 'initial_follower_count' | 'className' | 'onFollowChange'>) {
  return (
    <FollowButton
      user_id={user_id}
      initial_is_following={initial_is_following}
      initial_follower_count={initial_follower_count}
      show_count={true}
      className={className}
      onFollowChange={onFollowChange}
    />
  )
}

/**
 * Heart-style follow button (like social media)
 */
export function HeartFollowButton({
  user_id,
  initial_is_following = false,
  className,
  onFollowChange,
}: Pick<FollowButtonProps, 'user_id' | 'initial_is_following' | 'className' | 'onFollowChange'>) {
  return (
    <FollowButton
      user_id={user_id}
      initial_is_following={initial_is_following}
      size="sm"
      variant="ghost"
      className={cn('h-8 w-8 p-0 hover:bg-red-50', className)}
      onFollowChange={onFollowChange}
    />
  )
}

export default FollowButton