'use client'

import { useState, useEffect } from 'react'
import { getLeaderboard, getPointsHistory, getCourierStats } from '@/lib/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  Trophy, 
  Award, 
  TrendingUp, 
  Crown, 
  Medal,
  Star,
  Gift,
  Target,
  Users,
  Calendar,
  Zap
} from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'

interface CourierRanking {
  id: string
  rank: number
  username: string
  level: number
  points: number
  taskCount: number
  averageRating: number
  badge?: string
  isCurrentUser?: boolean
}

interface PointsHistory {
  id: string
  action: string
  points: number
  description: string
  timestamp: string
}

interface LevelInfo {
  currentLevel: number
  currentLevelName: string
  nextLevel: number
  nextLevelName: string
  currentPoints: number
  nextLevelPoints: number
  progress: number
}

export default function CourierPointsPage() {
  const { courierInfo, getCourierLevelName } = useCourierPermission()
  
  const [rankingScope, setRankingScope] = useState<'building' | 'zone' | 'school' | 'city' | 'national'>('school')
  const [schoolRanking, setSchoolRanking] = useState<CourierRanking[]>([])
  const [pointsHistory, setPointsHistory] = useState<PointsHistory[]>([])
  const [levelInfo, setLevelInfo] = useState<LevelInfo>({
    currentLevel: 2,
    currentLevelName: '二级信使',
    nextLevel: 3,
    nextLevelName: '三级信使',
    currentPoints: 2850,
    nextLevelPoints: 5000,
    progress: 57
  })

  // 从API获取数据
  useEffect(() => {
    const fetchData = async () => {
      try {
        // 获取当前选中范围的排行榜数据
        const leaderboardResponse = await getLeaderboard()
        const leaderboardData = (leaderboardResponse.data as any)?.leaderboard || []
        
        // 转换API数据格式
        const apiRanking: CourierRanking[] = leaderboardData.map((entry: any) => ({
          id: entry.id,
          rank: entry.rank,
          username: entry.name,
          level: entry.level,
          points: entry.total_points,
          taskCount: 0, // TODO: 从API获取任务数量
          averageRating: 4.8, // TODO: 从API获取评分
          badge: entry.rank === 1 ? '金牌信使' : entry.rank === 2 ? '银牌信使' : entry.rank === 3 ? '铜牌信使' : undefined,
          isCurrentUser: false // TODO: 检查是否为当前用户
        }))
        
        setSchoolRanking(apiRanking)
        
        // 获取积分历史
        const historyResponse = await getPointsHistory()
        const historyData = (historyResponse.data as any)?.history || []
        
        const apiHistory: PointsHistory[] = historyData.map((item: any) => ({
          id: item.id,
          action: item.action,
          points: item.points,
          description: item.action,
          timestamp: item.created_at
        }))
        
        setPointsHistory(apiHistory)
        
        // 获取用户统计信息以更新等级进度
        const statsResponse = await getCourierStats()
        const stats = (statsResponse.data as any) || {}
        
        if (stats?.level_progress) {
          setLevelInfo({
            currentLevel: stats?.level_progress?.current_level || 1,
            currentLevelName: `${stats?.level_progress?.current_level || 1}级信使`,
            nextLevel: stats?.level_progress?.next_level || 2,
            nextLevelName: `${stats?.level_progress?.next_level || 2}级信使`,
            currentPoints: stats?.total_points || 0,
            nextLevelPoints: (stats?.level_progress?.current_level || 1) * 1000, // 简单计算
            progress: stats?.level_progress?.progress_percentage || 0
          })
        }
        
      } catch (error) {
        console.error('Failed to load points data:', error)
        // 如果API失败，使用模拟数据
        const mockRanking: CourierRanking[] = [
        {
          id: 'r001',
          rank: 1,
          username: '北大最佳信使',
          level: 3,
          points: 4850,
          taskCount: 256,
          averageRating: 4.9,
          badge: '金牌信使'
        },
        {
          id: 'r002',
          rank: 2,
          username: '传递温暖小天使',
          level: 2,
          points: 3420,
          taskCount: 198,
          averageRating: 4.8,
          badge: '银牌信使'
        },
        {
          id: 'r003',
          rank: 3,
          username: '当前用户',
          level: 2,
          points: 2850,
          taskCount: 145,
          averageRating: 4.7,
          badge: '铜牌信使',
          isCurrentUser: true
        },
        {
          id: 'r004',
          rank: 4,
          username: '校园快递员',
          level: 2,
          points: 2650,
          taskCount: 132,
          averageRating: 4.6
        },
        {
          id: 'r005',
          rank: 5,
          username: '爱心投递员',
          level: 2,
          points: 2480,
          taskCount: 127,
          averageRating: 4.5
        }
        ]
        setSchoolRanking(mockRanking)

        // 使用模拟积分历史作为后备
        const mockHistory: PointsHistory[] = [
          {
            id: 'h001',
            action: '投递完成',
            points: 50,
            description: '成功投递信件至宿舍3号楼',
            timestamp: '2024-01-21T10:30:00Z'
          }
        ]
        setPointsHistory(mockHistory)
      }
    }

    fetchData()
  }, [rankingScope])

  const getRankIcon = (rank: number) => {
    switch (rank) {
      case 1: return <Crown className="w-5 h-5 text-yellow-500" />
      case 2: return <Medal className="w-5 h-5 text-gray-400" />
      case 3: return <Award className="w-5 h-5 text-amber-600" />
      default: return <span className="w-5 h-5 flex items-center justify-center text-sm font-bold text-amber-700">{rank}</span>
    }
  }

  const getLevelColor = (level: number) => {
    switch (level) {
      case 1: return 'bg-yellow-100 text-yellow-800'
      case 2: return 'bg-orange-100 text-orange-800'
      case 3: return 'bg-amber-100 text-amber-800'
      case 4: return 'bg-purple-100 text-purple-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getScopeText = (scope: string) => {
    switch (scope) {
      case 'building': return '楼栋排行'
      case 'zone': return '片区排行'
      case 'school': return '学校排行'
      case 'city': return '城市排行'
      case 'national': return '全国排行'
      default: return '排行榜'
    }
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-6xl mx-auto px-4 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Trophy className="w-8 h-8 text-amber-600" />
            <h1 className="text-3xl font-bold text-amber-900">信使积分中心</h1>
          </div>
          <p className="text-amber-700">追踪积分、等级进度和排行榜表现</p>
        </div>

        {/* 等级进度卡片 */}
        <Card className="border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50 mb-8">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-amber-900">
              <Star className="w-6 h-6" />
              我的等级进度
            </CardTitle>
            <CardDescription>完成更多任务提升信使等级</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              {/* 当前等级状态 */}
              <div className="flex items-center justify-between">
                <div>
                  <Badge className={`${getLevelColor(levelInfo.currentLevel)} mb-2`}>
                    {levelInfo.currentLevelName}
                  </Badge>
                  <div className="text-2xl font-bold text-amber-900">{levelInfo.currentPoints} 积分</div>
                </div>
                <div className="text-right">
                  <div className="text-sm text-amber-600 mb-1">下一等级</div>
                  <div className="font-semibold text-amber-900">{levelInfo.nextLevelName}</div>
                  <div className="text-sm text-amber-600">需要 {levelInfo.nextLevelPoints} 积分</div>
                </div>
              </div>

              {/* 进度条 */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm text-amber-700">
                  <span>升级进度</span>
                  <span>{levelInfo.progress}%</span>
                </div>
                <Progress value={levelInfo.progress} className="h-3" />
                <div className="flex justify-between text-xs text-amber-600">
                  <span>{levelInfo.currentPoints}</span>
                  <span>还需 {levelInfo.nextLevelPoints - levelInfo.currentPoints} 积分</span>
                  <span>{levelInfo.nextLevelPoints}</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Tabs defaultValue="ranking" className="space-y-6">
          <TabsList className="bg-amber-100">
            <TabsTrigger value="ranking" className="data-[state=active]:bg-amber-200">积分排行榜</TabsTrigger>
            <TabsTrigger value="history" className="data-[state=active]:bg-amber-200">积分历史</TabsTrigger>
            <TabsTrigger value="rewards" className="data-[state=active]:bg-amber-200">奖励兑换</TabsTrigger>
          </TabsList>

          <TabsContent value="ranking" className="space-y-6">
            <Card className="border-amber-200">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-amber-900">积分排行榜</CardTitle>
                    <CardDescription>与其他信使比较积分和表现</CardDescription>
                  </div>
                  <Select value={rankingScope} onValueChange={(value: any) => setRankingScope(value)}>
                    <SelectTrigger className="w-48 border-amber-200">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="building">楼栋排行</SelectItem>
                      <SelectItem value="zone">片区排行</SelectItem>
                      <SelectItem value="school">学校排行</SelectItem>
                      <SelectItem value="city">城市排行</SelectItem>
                      <SelectItem value="national">全国排行</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {schoolRanking.map((courier) => (
                    <Card 
                      key={courier.id} 
                      className={`border transition-all ${
                        courier.isCurrentUser 
                          ? 'border-amber-400 bg-amber-50' 
                          : 'border-amber-200 hover:border-amber-300'
                      }`}
                    >
                      <CardContent className="p-4">
                        <div className="flex items-center space-x-4">
                          {/* 排名图标 */}
                          <div className="flex-shrink-0">
                            {getRankIcon(courier.rank)}
                          </div>

                          {/* 用户头像 */}
                          <div className="w-12 h-12 bg-amber-600 text-white rounded-full flex items-center justify-center font-bold flex-shrink-0">
                            {courier.username.charAt(0)}
                          </div>

                          {/* 用户信息 */}
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                              <h3 className={`font-semibold ${courier.isCurrentUser ? 'text-amber-900' : 'text-gray-900'}`}>
                                {courier.username}
                                {courier.isCurrentUser && <span className="text-amber-600"> (我)</span>}
                              </h3>
                              <Badge className={getLevelColor(courier.level)}>
                                {courier.level}级信使
                              </Badge>
                              {courier.badge && (
                                <Badge variant="outline" className="border-amber-300 text-amber-700">
                                  {courier.badge}
                                </Badge>
                              )}
                            </div>
                            <div className="text-sm text-gray-600 space-y-1">
                              <div className="flex items-center gap-4">
                                <div className="flex items-center gap-1">
                                  <Trophy className="w-4 h-4" />
                                  <span>{courier.points} 积分</span>
                                </div>
                                <div className="flex items-center gap-1">
                                  <Target className="w-4 h-4" />
                                  <span>完成 {courier.taskCount} 任务</span>
                                </div>
                                <div className="flex items-center gap-1">
                                  <Star className="w-4 h-4" />
                                  <span>{courier.averageRating}/5.0</span>
                                </div>
                              </div>
                            </div>
                          </div>

                          {/* 排名 */}
                          <div className="text-right flex-shrink-0">
                            <div className={`text-2xl font-bold ${
                              courier.rank <= 3 ? 'text-amber-600' : 'text-gray-600'
                            }`}>
                              #{courier.rank}
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="history" className="space-y-6">
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="text-amber-900">积分历史</CardTitle>
                <CardDescription>查看最近的积分获得记录</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {pointsHistory.map((record) => (
                    <div key={record.id} className="flex items-center space-x-4 p-4 border border-amber-100 rounded-lg">
                      <div className="w-10 h-10 bg-green-100 text-green-600 rounded-full flex items-center justify-center flex-shrink-0">
                        <Zap className="w-5 h-5" />
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h4 className="font-semibold text-gray-900">{record.action}</h4>
                          <Badge variant="outline" className="border-green-300 text-green-700">
                            +{record.points} 积分
                          </Badge>
                        </div>
                        <p className="text-sm text-gray-600">{record.description}</p>
                        <p className="text-xs text-gray-500 mt-1">
                          {new Date(record.timestamp).toLocaleString()}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="rewards" className="space-y-6">
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="text-amber-900">奖励兑换</CardTitle>
                <CardDescription>使用积分兑换精美奖品和特权</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-center py-12">
                  <Gift className="w-12 h-12 text-amber-400 mx-auto mb-4" />
                  <h3 className="text-lg font-semibold text-amber-900 mb-2">奖励兑换功能开发中</h3>
                  <p className="text-amber-700">即将推出积分兑换商城，敬请期待！</p>
                  <Button className="mt-4 bg-amber-600 hover:bg-amber-700 text-white">
                    <Users className="w-4 h-4 mr-2" />
                    加入Beta测试
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}