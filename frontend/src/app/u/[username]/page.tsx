'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Suspense } from 'react'

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
  stats?: {
    letters_sent: number
    letters_received: number
    museum_contributions: number
    total_points: number
  }
}

interface UserLetter {
  id: number
  title?: string
  content_preview: string
  created_at: string
  status: string
  recipient?: string
  sender?: string
}

function UserPageContent() {
  const { username } = useParams()
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [userLetters, setUserLetters] = useState<UserLetter[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'profile' | 'letters' | 'museum'>('profile')

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
              <h1 className="text-3xl font-bold text-gray-800 mb-2">
                {userProfile.nickname || userProfile.username}
              </h1>
              {userProfile.nickname && (
                <p className="text-gray-600 mb-2">@{userProfile.username}</p>
              )}
              <div className="flex items-center space-x-4 text-sm text-gray-500 mb-4">
                <span className="bg-amber-100 text-amber-700 px-2 py-1 rounded">
                  {getRoleDisplayName(userProfile.role)}
                </span>
                {userProfile.school && (
                  <span>{userProfile.school}</span>
                )}
                <span>加入于 {formatDate(userProfile.created_at)}</span>
              </div>
              {userProfile.bio && (
                <p className="text-gray-700 mb-4">{userProfile.bio}</p>
              )}
            </div>
          </div>

          {/* 统计信息 */}
          {userProfile.stats && (
            <div className="mt-6 pt-6 border-t border-amber-100">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
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
                  <div className="text-2xl font-bold text-amber-600">
                    {userProfile.stats.museum_contributions}
                  </div>
                  <div className="text-sm text-gray-600">博物馆贡献</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-amber-600">
                    {userProfile.stats.total_points}
                  </div>
                  <div className="text-sm text-gray-600">总积分</div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 标签页 */}
        <div className="bg-white rounded-lg shadow-sm border border-amber-200">
          <div className="border-b border-amber-100">
            <nav className="flex space-x-8 px-6">
              <button
                onClick={() => setActiveTab('profile')}
                className={`py-4 px-2 border-b-2 font-medium text-sm ${
                  activeTab === 'profile'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                个人资料
              </button>
              <button
                onClick={() => setActiveTab('letters')}
                className={`py-4 px-2 border-b-2 font-medium text-sm ${
                  activeTab === 'letters'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                公开信件 ({userLetters.length})
              </button>
              <button
                onClick={() => setActiveTab('museum')}
                className={`py-4 px-2 border-b-2 font-medium text-sm ${
                  activeTab === 'museum'
                    ? 'border-amber-500 text-amber-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                博物馆贡献
              </button>
            </nav>
          </div>

          <div className="p-6">
            {activeTab === 'profile' && (
              <div className="space-y-6">
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
              </div>
            )}

            {activeTab === 'letters' && (
              <div>
                {userLetters.length > 0 ? (
                  <div className="space-y-4">
                    {userLetters.map((letter) => (
                      <div key={letter.id} className="border border-amber-100 rounded-lg p-4">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            {letter.title && (
                              <h4 className="text-lg font-medium text-gray-800 mb-2">
                                {letter.title}
                              </h4>
                            )}
                            <p className="text-gray-600 mb-2 line-clamp-2">
                              {letter.content_preview}
                            </p>
                            <div className="flex items-center space-x-4 text-sm text-gray-500">
                              <span>{formatDate(letter.created_at)}</span>
                              <span className="bg-green-100 text-green-700 px-2 py-1 rounded">
                                {letter.status}
                              </span>
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

            {activeTab === 'museum' && (
              <div className="text-center py-8">
                <div className="text-4xl mb-4">🏛️</div>
                <p className="text-gray-500">博物馆贡献功能开发中...</p>
              </div>
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