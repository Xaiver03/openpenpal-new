/**
 * UserCard Component - SOTA Implementation
 * 用户卡片组件 - 支持多种布局、关注操作、用户信息展示
 */

'use client'

import React from 'react'
import Link from 'next/link'
import { Card, CardContent, CardDescription, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { 
  Users, 
  MapPin, 
  Calendar,
  MessageCircle,
  Heart,
  Eye,
  Award,
  CheckCircle,
  Mail
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatRelativeTime } from '@/lib/utils'
import { FollowButton, CompactFollowButton } from './follow-button'
import { BadgeFollowStats } from './follow-stats'
import type { FollowUser } from '@/types/follow'

interface UserCardProps {
  user: FollowUser
  variant?: 'default' | 'compact' | 'detailed' | 'minimal'
  show_follow_button?: boolean
  show_stats?: boolean
  show_mutual?: boolean
  show_bio?: boolean
  show_school?: boolean
  clickable?: boolean
  className?: string
  onUserClick?: (user: FollowUser) => void
  onFollowChange?: (isFollowing: boolean, followerCount: number) => void
}

export function UserCard({
  user,
  variant = 'default',
  show_follow_button = true,
  show_stats = true,
  show_mutual = true,
  show_bio = true,
  show_school = true,
  clickable = true,
  className,
  onUserClick,
  onFollowChange,
}: UserCardProps) {
  
  const handleCardClick = () => {
    if (clickable && onUserClick) {
      onUserClick(user)
    }
  }
  
  const renderUserAvatar = (size: 'sm' | 'md' | 'lg' = 'md') => {
    const sizeClasses = {
      sm: 'h-8 w-8',
      md: 'h-12 w-12',
      lg: 'h-16 w-16'
    }
    
    return (
      <Avatar className={sizeClasses[size]}>
        <AvatarImage src={user.avatar} alt={user.nickname} />
        <AvatarFallback className="bg-gradient-to-r from-blue-400 to-purple-500 text-white">
          {user.nickname.charAt(0).toUpperCase()}
        </AvatarFallback>
      </Avatar>
    )
  }
  
  const renderUserInfo = (showDetails: boolean = true) => (
    <div className="flex-1 min-w-0">
      <div className="flex items-center gap-2 mb-1">
        <h3 className={cn(
          'font-medium truncate',
          variant === 'compact' ? 'text-sm' : 'text-base'
        )}>
          {user.nickname}
        </h3>
        {(user.role === 'platform_admin' || user.role === 'super_admin') && (
          <Badge variant="secondary" className="text-xs">
            <Award className="h-3 w-3 mr-1" />
            管理员
          </Badge>
        )}
        {user.courierInfo && (
          <Badge variant="outline" className="text-xs">
            信使L{user.courierInfo.level}
          </Badge>
        )}
      </div>
      
      <p className={cn(
        'text-muted-foreground truncate',
        variant === 'compact' ? 'text-xs' : 'text-sm'
      )}>
        @{user.username}
      </p>
      
      {showDetails && show_school && user.school_name && (
        <div className="flex items-center gap-1 mt-1">
          <MapPin className="h-3 w-3 text-muted-foreground" />
          <span className="text-xs text-muted-foreground truncate">
            {user.school_name}
          </span>
        </div>
      )}
      
      {showDetails && show_bio && user.bio && variant !== 'compact' && (
        <p className="text-sm text-muted-foreground mt-2 line-clamp-2">
          {user.bio}
        </p>
      )}
    </div>
  )
  
  const renderMutualInfo = () => {
    if (!show_mutual || !user.mutual_followers_count || user.mutual_followers_count === 0) {
      return null
    }
    
    return (
      <div className="flex items-center gap-1 text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded">
        <Users className="h-3 w-3" />
        <span>{user.mutual_followers_count} 位共同关注</span>
      </div>
    )
  }
  
  const renderStats = () => {
    if (!show_stats) return null
    
    if (variant === 'compact') {
      return (
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          <span>{user.followers_count || 0} 关注者</span>
          {user.letters_count && user.letters_count > 0 && (
            <span>{user.letters_count} 信件</span>
          )}
        </div>
      )
    }
    
    return (
      <div className="flex items-center gap-4 text-sm text-muted-foreground">
        <div className="flex items-center gap-1">
          <Users className="h-4 w-4" />
          <span>{user.followers_count || 0}</span>
        </div>
        <div className="flex items-center gap-1">
          <Mail className="h-4 w-4" />
          <span>{user.letters_count || 0}</span>
        </div>
        {user.last_login_at && (
          <div className="flex items-center gap-1">
            <Eye className="h-4 w-4" />
            <span>{formatRelativeTime(new Date(user.last_login_at))}</span>
          </div>
        )}
      </div>
    )
  }
  
  // Minimal variant - single line display
  if (variant === 'minimal') {
    return (
      <div className={cn(
        'flex items-center gap-3 p-2 rounded-lg hover:bg-muted/50 transition-colors',
        clickable && 'cursor-pointer',
        className
      )} onClick={handleCardClick}>
        {renderUserAvatar('sm')}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm truncate">{user.nickname}</span>
            {user.is_following && (
              <CheckCircle className="h-3 w-3 text-green-500 flex-shrink-0" />
            )}
          </div>
          <span className="text-xs text-muted-foreground">@{user.username}</span>
        </div>
        {show_follow_button && (
          <CompactFollowButton
            user_id={user.id}
            initial_is_following={user.is_following}
            onFollowChange={onFollowChange}
          />
        )}
      </div>
    )
  }
  
  // Compact variant - horizontal layout
  if (variant === 'compact') {
    return (
      <Card className={cn('p-3', className)}>
        <div className="flex items-center gap-3">
          {clickable ? (
            <Link href={`/u/${user.username}`} className="flex items-center gap-3 flex-1 min-w-0">
              {renderUserAvatar('sm')}
              {renderUserInfo(false)}
            </Link>
          ) : (
            <>
              {renderUserAvatar('sm')}
              {renderUserInfo(false)}
            </>
          )}
          
          <div className="flex flex-col items-end gap-2">
            {show_follow_button && (
              <CompactFollowButton
                user_id={user.id}
                initial_is_following={user.is_following}
                onFollowChange={onFollowChange}
              />
            )}
            {renderMutualInfo()}
          </div>
        </div>
        
        {show_stats && (
          <div className="mt-3 pt-3 border-t">
            {renderStats()}
          </div>
        )}
      </Card>
    )
  }
  
  // Detailed variant - full information display
  if (variant === 'detailed') {
    return (
      <Card className={cn('overflow-hidden', className)}>
        <CardHeader className="pb-3">
          <div className="flex items-start gap-4">
            {clickable ? (
              <Link href={`/u/${user.username}`}>
                {renderUserAvatar('lg')}
              </Link>
            ) : (
              renderUserAvatar('lg')
            )}
            
            <div className="flex-1 min-w-0">
              {clickable ? (
                <Link href={`/u/${user.username}`} className="block">
                  {renderUserInfo()}
                </Link>
              ) : (
                renderUserInfo()
              )}
              
              {renderMutualInfo()}
              
              <div className="flex items-center gap-4 mt-3">
                <BadgeFollowStats
                  followerCount={user.followers_count || 0}
                  followingCount={user.following_count || 0}
                />
                {user.created_at && (
                  <div className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Calendar className="h-3 w-3" />
                    <span>加入于 {new Date(user.created_at).toLocaleDateString()}</span>
                  </div>
                )}
              </div>
            </div>
            
            {show_follow_button && (
              <FollowButton
                user_id={user.id}
                initial_is_following={user.is_following}
                initial_follower_count={user.followers_count}
                onFollowChange={onFollowChange}
              />
            )}
          </div>
        </CardHeader>
        
        {user.bio && (
          <CardContent className="pt-0">
            <p className="text-sm text-muted-foreground">{user.bio}</p>
          </CardContent>
        )}
      </Card>
    )
  }
  
  // Default variant - balanced layout
  return (
    <Card className={cn('overflow-hidden hover:shadow-md transition-all', className)}>
      <CardContent className="p-4">
        <div className="flex items-start gap-4">
          {clickable ? (
            <Link href={`/u/${user.username}`}>
              {renderUserAvatar('md')}
            </Link>
          ) : (
            renderUserAvatar('md')
          )}
          
          <div className="flex-1 min-w-0">
            {clickable ? (
              <Link href={`/u/${user.username}`} className="block hover:opacity-80">
                {renderUserInfo()}
              </Link>
            ) : (
              renderUserInfo()
            )}
            
            <div className="flex items-center gap-2 mt-2">
              {renderMutualInfo()}
              {user.is_following && (
                <Badge variant="secondary" className="text-xs">
                  <Heart className="h-3 w-3 mr-1 fill-current" />
                  已关注
                </Badge>
              )}
            </div>
            
            {show_stats && (
              <div className="mt-3">
                {renderStats()}
              </div>
            )}
          </div>
          
          {show_follow_button && (
            <div className="flex flex-col gap-2">
              <FollowButton
                user_id={user.id}
                initial_is_following={user.is_following}
                initial_follower_count={user.followers_count}
                size="sm"
                onFollowChange={onFollowChange}
              />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

/**
 * UserCardSkeleton - Loading state component
 */
export function UserCardSkeleton({ variant = 'default' }: { variant?: UserCardProps['variant'] }) {
  if (variant === 'minimal') {
    return (
      <div className="flex items-center gap-3 p-2">
        <div className="h-8 w-8 bg-muted rounded-full animate-pulse" />
        <div className="flex-1 space-y-1">
          <div className="h-4 bg-muted rounded w-24 animate-pulse" />
          <div className="h-3 bg-muted rounded w-16 animate-pulse" />
        </div>
        <div className="h-6 w-16 bg-muted rounded animate-pulse" />
      </div>
    )
  }
  
  if (variant === 'compact') {
    return (
      <Card className="p-3">
        <div className="flex items-center gap-3">
          <div className="h-8 w-8 bg-muted rounded-full animate-pulse" />
          <div className="flex-1 space-y-1">
            <div className="h-4 bg-muted rounded w-32 animate-pulse" />
            <div className="h-3 bg-muted rounded w-24 animate-pulse" />
          </div>
          <div className="h-6 w-16 bg-muted rounded animate-pulse" />
        </div>
      </Card>
    )
  }
  
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-start gap-4">
          <div className="h-12 w-12 bg-muted rounded-full animate-pulse" />
          <div className="flex-1 space-y-2">
            <div className="h-5 bg-muted rounded w-40 animate-pulse" />
            <div className="h-4 bg-muted rounded w-32 animate-pulse" />
            <div className="h-3 bg-muted rounded w-48 animate-pulse" />
          </div>
          <div className="h-8 w-20 bg-muted rounded animate-pulse" />
        </div>
      </CardContent>
    </Card>
  )
}

export default UserCard