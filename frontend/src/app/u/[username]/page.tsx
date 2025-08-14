'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Suspense } from 'react'
import { UserLevelDisplay } from '@/components/user/level-badge'
import { OPCodeDisplay } from '@/components/user/opcode-display'
import { FollowButton } from '@/components/follow/follow-button'
import { useUser } from '@/stores/user-store'
import { useFollowStatus } from '@/stores/follow-store'
import { ProfileComments } from '@/components/profile/profile-comments'
import { UserActivityFeed } from '@/components/profile/user-activity-feed'

interface UserProfile {
  id: number
  username: string
  nickname?: string
  email?: string
  role: string
  avatar_url?: string
  bio?: string
  school?: string
  created_at: string
  op_code?: string
  writing_level?: number
  courier_level?: number
  stats?: {
    letters_sent: number
    letters_received: number
    museum_contributions: number
    total_points: number
    writing_points?: number
    courier_points?: number
    current_streak?: number
    achievements?: string[]
  }
  privacy?: {
    show_email: boolean
    show_op_code: boolean
    show_stats: boolean
    op_code_privacy: 'full' | 'partial' | 'hidden'
  }
}

interface UserLetter {
  id: number
  title?: string
  content_preview: string
  created_at: string
  status: string
  visibility: 'private' | 'public' | 'friends'
  recipient?: string
  sender?: string
}

function UserPageContent() {
  const { username } = useParams()
  const { user: currentUser } = useUser()
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [userLetters, setUserLetters] = useState<UserLetter[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'profile' | 'letters' | 'works' | 'courier' | 'collection' | 'comments'>('profile')
  const [followStats, setFollowStats] = useState({ follower_count: 0, following_count: 0 })
  const [isFollowing, setIsFollowing] = useState(false)

  useEffect(() => {
    if (username) {
      fetchUserProfile(username as string)
    }
  }, [username])

  const fetchUserProfile = async (username: string) => {
    try {
      setLoading(true)
      setError(null)

      // 获取用户基本信息
      const profileResponse = await fetch(`/api/users/${username}/profile`, {
        credentials: 'include'
      })

      if (!profileResponse.ok) {
        if (profileResponse.status === 404) {
          setError('用户不存在')
        } else if (profileResponse.status === 403) {
          setError('该用户的资料未公开')
        } else {
          setError('获取用户信息失败')
        }
        return
      }

      const profileData = await profileResponse.json()
      setUserProfile(profileData.data || profileData)
      
      // 获取关注统计信息
      try {
        const followStatsResponse = await fetch(`/api/users/${username}/follow-stats`, {
          credentials: 'include'
        })
        if (followStatsResponse.ok) {
          const statsData = await followStatsResponse.json()
          setFollowStats({
            follower_count: statsData.data?.follower_count || 0,
            following_count: statsData.data?.following_count || 0
          })
        }
      } catch (err) {
        console.log('获取关注统计失败')
      }
      
      // 检查当前用户是否已关注该用户
      if (currentUser) {
        try {
          const followStatusResponse = await fetch(`/api/users/${profileData.data?.id || profileData.id}/follow-status`, {
            credentials: 'include'
          })
          if (followStatusResponse.ok) {
            const statusData = await followStatusResponse.json()
            setIsFollowing(statusData.data?.is_following || false)
          }
        } catch (err) {
          console.log('获取关注状态失败')
        }
      }

      // 获取用户公开的信件（如果有权限）
      try {
        const lettersResponse = await fetch(`/api/users/${username}/letters?public=true`, {
          credentials: 'include'
        })
        if (lettersResponse.ok) {
          const lettersData = await lettersResponse.json()
          setUserLetters(lettersData.data || [])
        }
      } catch (err) {
        console.log('获取用户信件失败，可能是权限不够')
      }

    } catch (err) {
      console.error('获取用户信息失败:', err)
      setError('网络错误，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }

  const getRoleDisplayName = (role: string) => {
    const roleMap: Record<string, string> = {
      'student': '学生',
      'courier': '信使',
      'senior_courier': '高级信使',
      'coordinator': '协调员',
      'admin': '管理员',
      'super_admin': '超级管理员'
    }
    return roleMap[role] || role
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-letter-paper flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-amber-600 mx-auto mb-4"></div>
          <p className="text-amber-700">正在加载用户信息...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-letter-paper flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4">😔</div>
          <h1 className="text-2xl font-bold text-gray-800 mb-2">{error}</h1>
          <p className="text-gray-600">请检查用户名是否正确</p>
        </div>
      </div>
    )
  }

  if (!userProfile) {
    return (
      <div className="min-h-screen bg-letter-paper flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4">🤔</div>
          <h1 className="text-2xl font-bold text-gray-800 mb-2">用户不存在</h1>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-letter-paper">
      <div className="container mx-auto px-4 py-8 max-w-4xl">
        {/* 用户头像和基本信息 */}
        <div className="bg-white rounded-lg shadow-sm border border-amber-200 p-6 mb-6">
          <div className="flex items-start space-x-6">
            <div className="flex-shrink-0">
              {userProfile.avatar_url ? (
                <img 
                  src={userProfile.avatar_url} 
                  alt={userProfile.username}
                  className="w-24 h-24 rounded-full object-cover border-2 border-amber-200"
                />
              ) : (
                <div className="w-24 h-24 rounded-full bg-amber-100 flex items-center justify-center">
                  <span className="text-2xl text-amber-600">
                    {userProfile.username.charAt(0).toUpperCase()}
                  </span>
                </div>
              )}
            </div>
            
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-3">
                <h1 className="text-3xl font-bold text-gray-800">
                  {userProfile.nickname || userProfile.username}
                </h1>
                {/* 等级徽章 */}
                <UserLevelDisplay 
                  writingLevel={userProfile.writing_level}
                  courierLevel={userProfile.courier_level}
                  compact
                />
                {/* 关注按钮 - 只在用户已登录且不是查看自己时显示 */}
                {currentUser && currentUser.id !== userProfile.id.toString() && (
                  <FollowButton
                    user_id={userProfile.id.toString()}
                    initial_is_following={isFollowing}
                    initial_follower_count={followStats.follower_count}
                    onFollowChange={(following, newCount) => {
                      setIsFollowing(following)
                      setFollowStats(prev => ({ ...prev, follower_count: newCount }))
                    }}
                  />
                )}
              </div>
              
              {userProfile.nickname && (
                <p className="text-gray-600 mb-2">@{userProfile.username}</p>
              )}
              
              <div className="flex items-center space-x-4 text-sm text-gray-500 mb-3">
                <span className="bg-amber-100 text-amber-700 px-2 py-1 rounded">
                  {getRoleDisplayName(userProfile.role)}
                </span>
                {userProfile.school && (
                  <span>{userProfile.school}</span>
                )}
                <span>加入于 {formatDate(userProfile.created_at)}</span>
              </div>

              {/* OP Code 显示 */}
              {userProfile.op_code && userProfile.privacy?.show_op_code && (
                <div className="mb-3">
                  <OPCodeDisplay 
                    opCode={userProfile.op_code}
                    showPrivacy={userProfile.privacy?.op_code_privacy === 'partial'}
                  />
                </div>
              )}
              
              {userProfile.bio && (
                <p className="text-gray-700 mb-4">{userProfile.bio}</p>
              )}

              {/* 成就展示 */}
              {userProfile.stats?.achievements && userProfile.stats.achievements.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {userProfile.stats.achievements.map((achievement, index) => (
                    <span 
                      key={index}
                      className="text-xs bg-gradient-to-r from-yellow-100 to-yellow-200 text-yellow-800 px-2 py-1 rounded-full border border-yellow-300"
                    >
                      🏆 {achievement.replace(/_/g, ' ')}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* 统计信息 */}
          {userProfile.stats && (
            <div className="mt-6 pt-6 border-t border-amber-100">
              <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-amber-600">
                    {userProfile.stats.letters_sent}
                  </div>
                  <div className="text-sm text-gray-600">已发送</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-amber-600">
                    {userProfile.stats.letters_received}
                  </div>
                  <div className="text-sm text-gray-600">已收到</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-rose-600">
                    {followStats.follower_count}
                  </div>
                  <div className="text-sm text-gray-600">粉丝</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-600">
                    {followStats.following_count}
                  </div>
                  <div className="text-sm text-gray-600">关注</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-purple-600">
                    {userProfile.stats.museum_contributions}
                  </div>
                  <div className="text-sm text-gray-600">博物馆贡献</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-600">
                    {userProfile.stats.writing_points || 0}
                  </div>
                  <div className="text-sm text-gray-600">写信积分</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">
                    {userProfile.stats.courier_points || 0}
                  </div>
                  <div className="text-sm text-gray-600">信使积分</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-orange-600">
                    {userProfile.stats.current_streak || 0}
                  </div>
                  <div className="text-sm text-gray-600">连续天数</div>
                </div>
              </div>
              
              {/* 总积分展示 */}
              <div className="mt-4 text-center">
                <div className="inline-flex items-center gap-2 bg-gradient-to-r from-amber-100 to-yellow-100 px-4 py-2 rounded-full border border-amber-200">
                  <span className="text-sm text-gray-600">总积分</span>
                  <span className="text-xl font-bold text-amber-600">
                    {userProfile.stats.total_points}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 标签页 */}
        <div className="bg-white rounded-lg shadow-sm border border-amber-200">
          <div className="border-b border-amber-100">
            <nav className="flex space-x-8 px-6 overflow-x-auto">
              <button
                onClick={() => setActiveTab('profile')}
                className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === 'profile'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                个人资料
              </button>
              <button
                onClick={() => setActiveTab('letters')}
                className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === 'letters'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                📮 我的信件 ({userLetters.length})
              </button>
              <button
                onClick={() => setActiveTab('works')}
                className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === 'works'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                🖋 我的作品
              </button>
              {/* 只有信使才显示信使任务 */}
              {userProfile && userProfile.courier_level && userProfile.courier_level > 0 && (
                <button
                  onClick={() => setActiveTab('courier')}
                  className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                    activeTab === 'courier'
                      ? 'border-amber-500 text-amber-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                >
                  🎒 信使任务
                </button>
              )}
              <button
                onClick={() => setActiveTab('collection')}
                className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === 'collection'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                🏛 收藏 & 展览
              </button>
              <button
                onClick={() => setActiveTab('comments')}
                className={`py-4 px-2 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === 'comments'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                💬 留言板
              </button>
            </nav>
          </div>

          <div className="p-6">
            {activeTab === 'profile' && (
              <div className="space-y-6">
                {/* 基本信息 */}
                <div>
                  <h3 className="text-lg font-medium text-gray-800 mb-3">基本信息</h3>
                  <dl className="grid grid-cols-1 gap-x-4 gap-y-2 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm font-medium text-gray-500">用户名</dt>
                      <dd className="mt-1 text-sm text-gray-900">{userProfile.username}</dd>
                    </div>
                    {userProfile.nickname && (
                      <div>
                        <dt className="text-sm font-medium text-gray-500">昵称</dt>
                        <dd className="mt-1 text-sm text-gray-900">{userProfile.nickname}</dd>
                      </div>
                    )}
                    <div>
                      <dt className="text-sm font-medium text-gray-500">角色</dt>
                      <dd className="mt-1 text-sm text-gray-900">
                        {getRoleDisplayName(userProfile.role)}
                      </dd>
                    </div>
                    {userProfile.school && (
                      <div>
                        <dt className="text-sm font-medium text-gray-500">学校</dt>
                        <dd className="mt-1 text-sm text-gray-900">{userProfile.school}</dd>
                      </div>
                    )}
                    <div>
                      <dt className="text-sm font-medium text-gray-500">加入时间</dt>
                      <dd className="mt-1 text-sm text-gray-900">
                        {formatDate(userProfile.created_at)}
                      </dd>
                    </div>
                  </dl>
                </div>
                
                {/* 最近动态 */}
                <div>
                  <h3 className="text-lg font-medium text-gray-800 mb-3">最近动态</h3>
                  <UserActivityFeed 
                    user_id={userProfile.id.toString()}
                    max_items={5}
                    show_load_more={true}
                  />
                </div>
              </div>
            )}

            {activeTab === 'letters' && (
              <div>
                {userLetters.length > 0 ? (
                  <div className="space-y-4">
                    {userLetters.map((letter) => (
                      <div key={letter.id} className="border border-amber-100 rounded-lg p-4 hover:shadow-sm transition-shadow">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            {letter.title && (
                              <h4 className="text-lg font-medium text-gray-800 mb-2">
                                {letter.title}
                              </h4>
                            )}
                            <p className="text-gray-600 mb-3 line-clamp-2">
                              {letter.content_preview}
                            </p>
                            <div className="flex items-center space-x-4 text-sm text-gray-500">
                              <span>{formatDate(letter.created_at)}</span>
                              <span className="bg-green-100 text-green-700 px-2 py-1 rounded text-xs">
                                {letter.status}
                              </span>
                              {letter.recipient && (
                                <span className="text-blue-600">→ @{letter.recipient}</span>
                              )}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <div className="text-4xl mb-4">📭</div>
                    <p className="text-gray-500">暂无公开信件</p>
                  </div>
                )}
              </div>
            )}

            {activeTab === 'works' && (
              <div>
                {/* 作品分类和展示 */}
                <div className="mb-6">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-medium text-gray-800">我的作品</h3>
                    <div className="flex gap-2 text-sm">
                      <span className="bg-blue-100 text-blue-700 px-2 py-1 rounded">全部 {userLetters.length}</span>
                      <span className="bg-green-100 text-green-700 px-2 py-1 rounded">公开 {userLetters.filter(l => l.visibility === 'public').length}</span>
                    </div>
                  </div>
                  
                  {userLetters.length > 0 ? (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {userLetters.map((letter) => (
                        <div key={letter.id} className="border border-amber-100 rounded-lg p-4 hover:shadow-sm transition-shadow">
                          <div className="flex items-start justify-between mb-2">
                            <h4 className="text-lg font-medium text-gray-800 flex-1">
                              {letter.title || '无标题'}
                            </h4>
                            <div className="flex gap-1">
                              {letter.visibility === 'public' && (
                                <span className="text-xs bg-green-100 text-green-600 px-1.5 py-0.5 rounded">公开</span>
                              )}
                            </div>
                          </div>
                          <p className="text-gray-600 mb-3 line-clamp-3 text-sm">
                            {letter.content_preview}
                          </p>
                          <div className="flex items-center justify-between text-xs text-gray-500">
                            <span>{formatDate(letter.created_at)}</span>
                            <div className="flex items-center gap-2">
                              <span>📝 {letter.content_preview.length} 字</span>
                              <span>👀 暂无阅读数</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <div className="text-4xl mb-4">🖋️</div>
                      <p className="text-gray-500">还没有发布作品</p>
                    </div>
                  )}
                </div>
              </div>
            )}

            {activeTab === 'courier' && userProfile && userProfile.courier_level && userProfile.courier_level > 0 && (
              <div>
                <div className="mb-6">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-medium text-gray-800">信使任务中心</h3>
                    <div className="flex items-center gap-2">
                      <UserLevelDisplay courierLevel={userProfile.courier_level} />
                    </div>
                  </div>
                  
                  {/* 信使统计 */}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                    <div className="bg-blue-50 p-4 rounded-lg text-center">
                      <div className="text-2xl font-bold text-blue-600">12</div>
                      <div className="text-sm text-gray-600">待处理任务</div>
                    </div>
                    <div className="bg-green-50 p-4 rounded-lg text-center">
                      <div className="text-2xl font-bold text-green-600">45</div>
                      <div className="text-sm text-gray-600">已完成任务</div>
                    </div>
                    <div className="bg-orange-50 p-4 rounded-lg text-center">
                      <div className="text-2xl font-bold text-orange-600">3</div>
                      <div className="text-sm text-gray-600">今日任务</div>
                    </div>
                    <div className="bg-purple-50 p-4 rounded-lg text-center">
                      <div className="text-2xl font-bold text-purple-600">95%</div>
                      <div className="text-sm text-gray-600">完成率</div>
                    </div>
                  </div>

                  {/* 任务管理区域提示 */}
                  <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 text-center">
                    <div className="text-4xl mb-2">🎒</div>
                    <p className="text-gray-700">信使任务管理</p>
                    <p className="text-sm text-gray-500 mt-1">
                      完整的任务管理功能请访问 
                      <a href="/courier" className="text-amber-600 hover:text-amber-700 underline ml-1">信使中心</a>
                    </p>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'collection' && (
              <div>
                <div className="mb-6">
                  <h3 className="text-lg font-medium text-gray-800 mb-4">收藏 & 展览</h3>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {/* 我的收藏 */}
                    <div className="border border-amber-100 rounded-lg p-4">
                      <div className="flex items-center gap-2 mb-3">
                        <span className="text-lg">💝</span>
                        <h4 className="font-medium text-gray-800">我的收藏</h4>
                        <span className="text-sm text-gray-500">(5)</span>
                      </div>
                      <p className="text-sm text-gray-600 mb-3">收藏的精彩信件和作品</p>
                      <div className="text-center py-4">
                        <p className="text-sm text-gray-500">功能开发中...</p>
                      </div>
                    </div>

                    {/* 展览参与 */}
                    <div className="border border-amber-100 rounded-lg p-4">
                      <div className="flex items-center gap-2 mb-3">
                        <span className="text-lg">🏛️</span>
                        <h4 className="font-medium text-gray-800">展览参与</h4>
                        <span className="text-sm text-gray-500">(2)</span>
                      </div>
                      <p className="text-sm text-gray-600 mb-3">参与的博物馆展览活动</p>
                      <div className="text-center py-4">
                        <p className="text-sm text-gray-500">功能开发中...</p>
                      </div>
                    </div>
                  </div>
                  
                  {/* 成就展示区 */}
                  {userProfile?.stats?.achievements && userProfile.stats.achievements.length > 0 && (
                    <div className="mt-6 p-4 bg-gradient-to-r from-yellow-50 to-orange-50 rounded-lg border border-yellow-200">
                      <h4 className="font-medium text-gray-800 mb-3 flex items-center gap-2">
                        <span>🏆</span>
                        获得的成就
                      </h4>
                      <div className="flex flex-wrap gap-2">
                        {userProfile.stats.achievements.map((achievement, index) => (
                          <div 
                            key={index}
                            className="flex items-center gap-2 bg-white px-3 py-2 rounded-lg border border-yellow-200 shadow-sm"
                          >
                            <span className="text-lg">🏆</span>
                            <span className="text-sm font-medium text-gray-700">
                              {achievement.replace(/_/g, ' ').toUpperCase()}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {activeTab === 'comments' && (
              <ProfileComments
                profile_id={userProfile.id.toString()}
                profile_username={userProfile.username}
                allow_comments={true}
                max_display={20}
              />
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default function UserPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-letter-paper flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-amber-600"></div>
      </div>
    }>
      <UserPageContent />
    </Suspense>
  )
}