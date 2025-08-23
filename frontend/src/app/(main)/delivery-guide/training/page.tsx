'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackButton } from '@/components/ui/back-button'
import { 
  BookOpen,
  Play,
  CheckCircle,
  Clock,
  Star,
  Award,
  Users,
  Shield,
  MapPin,
  Smartphone,
  Target,
  TrendingUp,
  AlertTriangle,
  Lightbulb,
  Video,
  FileText,
  Headphones,
  ChevronRight,
  Download
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { usePermission } from '@/hooks/use-permission'

interface TrainingModule {
  id: string
  title: string
  description: string
  duration: number // minutes
  difficulty: 'beginner' | 'intermediate' | 'advanced'
  type: 'video' | 'text' | 'interactive'
  completed: boolean
  progress: number // 0-100
  icon: any
}

interface TrainingCategory {
  id: string
  title: string
  description: string
  icon: any
  color: string
  modules: TrainingModule[]
}

export default function CourierTrainingPage() {
  const { user } = useAuth()
  const { hasPermission, getRoleDisplayName } = usePermission()
  const [selectedCategory, setSelectedCategory] = useState<string>('')
  const [selectedModule, setSelectedModule] = useState<string>('')

  // 培训分类和模块
  const trainingCategories: TrainingCategory[] = [
    {
      id: 'basics',
      title: '基础知识',
      description: '信使工作的基本概念和流程',
      icon: BookOpen,
      color: 'bg-blue-50 text-blue-700 border-blue-200',
      modules: [
        {
          id: 'intro',
          title: 'OpenPenPal信使系统介绍',
          description: '了解四级信使体系和工作原理',
          duration: 15,
          difficulty: 'beginner',
          type: 'video',
          completed: false,
          progress: 0,
          icon: Users
        },
        {
          id: 'opcode-system',
          title: 'OP Code地址系统',
          description: '掌握6位地址编码的使用方法',
          duration: 20,
          difficulty: 'beginner',
          type: 'interactive',
          completed: true,
          progress: 100,
          icon: MapPin
        },
        {
          id: 'hierarchy',
          title: '信使等级与权限',
          description: '了解L1-L4等级权限和晋升路径',
          duration: 12,
          difficulty: 'beginner',
          type: 'text',
          completed: false,
          progress: 45,
          icon: Star
        }
      ]
    },
    {
      id: 'operations',
      title: '投递操作',
      description: '实际投递工作的标准流程',
      icon: Target,
      color: 'bg-green-50 text-green-700 border-green-200',
      modules: [
        {
          id: 'scanning',
          title: '扫码操作指南',
          description: '掌握扫码收件、投递的标准流程',
          duration: 18,
          difficulty: 'intermediate',
          type: 'video',
          completed: false,
          progress: 0,
          icon: Smartphone
        },
        {
          id: 'task-management',
          title: '任务管理系统',
          description: '学习任务接收、更新状态的操作',
          duration: 25,
          difficulty: 'intermediate',
          type: 'interactive',
          completed: false,
          progress: 30,
          icon: CheckCircle
        },
        {
          id: 'delivery-best-practices',
          title: '投递最佳实践',
          description: '提高投递效率的技巧和方法',
          duration: 22,
          difficulty: 'intermediate',
          type: 'video',
          completed: false,
          progress: 0,
          icon: TrendingUp
        }
      ]
    },
    {
      id: 'safety',
      title: '安全规范',
      description: '投递安全和应急处理',
      icon: Shield,
      color: 'bg-red-50 text-red-700 border-red-200',
      modules: [
        {
          id: 'safety-protocols',
          title: '安全操作规范',
          description: '投递过程中的安全注意事项',
          duration: 16,
          difficulty: 'beginner',
          type: 'text',
          completed: false,
          progress: 0,
          icon: Shield
        },
        {
          id: 'emergency-handling',
          title: '应急情况处理',
          description: '遇到突发情况的应对方法',
          duration: 20,
          difficulty: 'intermediate',
          type: 'video',
          completed: false,
          progress: 0,
          icon: AlertTriangle
        }
      ]
    },
    {
      id: 'advanced',
      title: '高级技能',
      description: '面向高等级信使的专业技能',
      icon: Award,
      color: 'bg-purple-50 text-purple-700 border-purple-200',
      modules: [
        {
          id: 'team-management',
          title: '团队管理技巧',
          description: 'L2+信使的下属管理和培训方法',
          duration: 35,
          difficulty: 'advanced',
          type: 'video',
          completed: false,
          progress: 0,
          icon: Users
        },
        {
          id: 'performance-optimization',
          title: '绩效优化策略',
          description: '提升团队整体投递效率的策略',
          duration: 28,
          difficulty: 'advanced',
          type: 'interactive',
          completed: false,
          progress: 0,
          icon: TrendingUp
        }
      ]
    }
  ]

  // 计算总体进度
  const calculateOverallProgress = () => {
    const allModules = trainingCategories.flatMap(cat => cat.modules)
    if (allModules.length === 0) return 0
    
    const totalProgress = allModules.reduce((sum, module) => sum + module.progress, 0)
    return Math.round(totalProgress / allModules.length)
  }

  // 获取完成的模块数
  const getCompletedModules = () => {
    const allModules = trainingCategories.flatMap(cat => cat.modules)
    return allModules.filter(module => module.completed).length
  }

  // 获取总模块数
  const getTotalModules = () => {
    return trainingCategories.flatMap(cat => cat.modules).length
  }

  // 获取难度标签
  const getDifficultyBadge = (difficulty: string) => {
    const styles = {
      beginner: 'bg-green-100 text-green-800',
      intermediate: 'bg-yellow-100 text-yellow-800',
      advanced: 'bg-red-100 text-red-800'
    }
    const labels = {
      beginner: '入门',
      intermediate: '中级',
      advanced: '高级'
    }
    return { style: styles[difficulty as keyof typeof styles], label: labels[difficulty as keyof typeof labels] }
  }

  // 获取类型图标
  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'video': return Video
      case 'text': return FileText
      case 'interactive': return Target
      default: return BookOpen
    }
  }

  const overallProgress = calculateOverallProgress()
  const completedModules = getCompletedModules()
  const totalModules = getTotalModules()

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <BackButton href="/delivery-guide" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <BookOpen className="h-8 w-8" />
              信使培训中心
            </h1>
            <p className="text-gray-600 mt-2">
              系统化的培训体系，助你成为专业信使
              {user && ` • ${getRoleDisplayName()}`}
            </p>
          </div>
        </div>
      </div>

      {/* 学习统计 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">学习进度</p>
                <p className="text-2xl font-bold">{overallProgress}%</p>
              </div>
              <TrendingUp className="h-5 w-5 text-blue-500" />
            </div>
            <Progress value={overallProgress} className="mt-2" />
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">完成课程</p>
                <p className="text-2xl font-bold">{completedModules}/{totalModules}</p>
              </div>
              <CheckCircle className="h-5 w-5 text-green-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">学习时长</p>
                <p className="text-2xl font-bold">
                  {trainingCategories.flatMap(cat => cat.modules)
                    .filter(m => m.completed)
                    .reduce((sum, m) => sum + m.duration, 0)}min
                </p>
              </div>
              <Clock className="h-5 w-5 text-purple-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">获得徽章</p>
                <p className="text-2xl font-bold">3</p>
              </div>
              <Award className="h-5 w-5 text-yellow-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 主要内容 */}
      <Tabs defaultValue="courses" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="courses">课程学习</TabsTrigger>
          <TabsTrigger value="progress">学习进度</TabsTrigger>
          <TabsTrigger value="certificates">认证考试</TabsTrigger>
        </TabsList>

        {/* 课程学习标签页 */}
        <TabsContent value="courses" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {trainingCategories.map((category) => (
              <Card key={category.id} className={`border ${category.color.split(' ').pop()}`}>
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <div className={`p-2 rounded-lg ${category.color}`}>
                      <category.icon className="h-6 w-6" />
                    </div>
                    <div className="flex-1">
                      <CardTitle className="flex items-center justify-between">
                        {category.title}
                        <Badge variant="outline" className="text-xs">
                          {category.modules.length} 课程
                        </Badge>
                      </CardTitle>
                      <CardDescription className="mt-1">
                        {category.description}
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-3">
                  {category.modules.map((module) => {
                    const TypeIcon = getTypeIcon(module.type)
                    const difficultyBadge = getDifficultyBadge(module.difficulty)
                    
                    return (
                      <div key={module.id} className="flex items-center gap-3 p-3 border rounded-lg hover:bg-gray-50 transition-colors">
                        <div className="flex items-center gap-3 flex-1">
                          <div className="p-2 bg-gray-100 rounded">
                            <TypeIcon className="h-4 w-4" />
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-1">
                              <h4 className="font-medium text-sm">{module.title}</h4>
                              <Badge variant="outline" className={`text-xs ${difficultyBadge.style}`}>
                                {difficultyBadge.label}
                              </Badge>
                              {module.completed && (
                                <CheckCircle className="w-4 h-4 text-green-500" />
                              )}
                            </div>
                            <p className="text-xs text-gray-600">{module.description}</p>
                            <div className="flex items-center gap-4 mt-1">
                              <span className="text-xs text-gray-500">{module.duration}分钟</span>
                              {module.progress > 0 && (
                                <div className="flex items-center gap-1">
                                  <Progress value={module.progress} className="w-16 h-1" />
                                  <span className="text-xs text-gray-500">{module.progress}%</span>
                                </div>
                              )}
                            </div>
                          </div>
                        </div>
                        <Button variant="outline" size="sm">
                          {module.completed ? (
                            <>
                              <CheckCircle className="w-3 h-3 mr-1" />
                              复习
                            </>
                          ) : module.progress > 0 ? (
                            <>
                              <Play className="w-3 h-3 mr-1" />
                              继续
                            </>
                          ) : (
                            <>
                              <Play className="w-3 h-3 mr-1" />
                              开始
                            </>
                          )}
                        </Button>
                      </div>
                    )
                  })}
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        {/* 学习进度标签页 */}
        <TabsContent value="progress" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2">
              <Card>
                <CardHeader>
                  <CardTitle>学习路径</CardTitle>
                  <CardDescription>
                    按推荐顺序完成以下学习模块
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {trainingCategories.map((category, categoryIndex) => (
                    <div key={category.id}>
                      <div className="flex items-center gap-2 mb-3">
                        <category.icon className="h-5 w-5 text-gray-600" />
                        <h3 className="font-semibold">{category.title}</h3>
                      </div>
                      {category.modules.map((module, moduleIndex) => (
                        <div key={module.id} className="flex items-center gap-4 ml-7 mb-2">
                          <div className={`w-4 h-4 rounded-full flex items-center justify-center text-xs ${
                            module.completed 
                              ? 'bg-green-500 text-white' 
                              : module.progress > 0 
                                ? 'bg-yellow-500 text-white'
                                : 'bg-gray-300'
                          }`}>
                            {module.completed ? '✓' : categoryIndex + 1}.{moduleIndex + 1}
                          </div>
                          <div className="flex-1">
                            <p className="text-sm font-medium">{module.title}</p>
                            {module.progress > 0 && !module.completed && (
                              <Progress value={module.progress} className="w-32 h-1 mt-1" />
                            )}
                          </div>
                          <span className="text-xs text-gray-500">{module.duration}min</span>
                        </div>
                      ))}
                    </div>
                  ))}
                </CardContent>
              </Card>
            </div>

            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">学习成就</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                      <BookOpen className="w-5 h-5 text-blue-600" />
                    </div>
                    <div>
                      <p className="font-medium text-sm">学习新手</p>
                      <p className="text-xs text-gray-600">完成第一个课程</p>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-3 opacity-50">
                    <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                      <Target className="w-5 h-5 text-green-600" />
                    </div>
                    <div>
                      <p className="font-medium text-sm">操作达人</p>
                      <p className="text-xs text-gray-600">完成所有操作课程</p>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-3 opacity-50">
                    <div className="w-10 h-10 bg-purple-100 rounded-full flex items-center justify-center">
                      <Award className="w-5 h-5 text-purple-600" />
                    </div>
                    <div>
                      <p className="font-medium text-sm">培训专家</p>
                      <p className="text-xs text-gray-600">完成全部课程</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">学习建议</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm text-gray-600">
                  <div className="flex items-start gap-2">
                    <Lightbulb className="w-4 h-4 text-yellow-500 mt-0.5" />
                    <div>
                      <p className="font-medium">循序渐进</p>
                      <p className="text-xs">建议按照推荐顺序学习</p>
                    </div>
                  </div>
                  
                  <div className="flex items-start gap-2">
                    <Clock className="w-4 h-4 text-blue-500 mt-0.5" />
                    <div>
                      <p className="font-medium">合理安排</p>
                      <p className="text-xs">每天学习20-30分钟效果最佳</p>
                    </div>
                  </div>
                  
                  <div className="flex items-start gap-2">
                    <Target className="w-4 h-4 text-green-500 mt-0.5" />
                    <div>
                      <p className="font-medium">实践结合</p>
                      <p className="text-xs">学习后立即在工作中实践</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </TabsContent>

        {/* 认证考试标签页 */}
        <TabsContent value="certificates" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Award className="h-5 w-5" />
                  信使基础认证
                </CardTitle>
                <CardDescription>
                  验证基本信使技能和知识掌握情况
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">前置要求</span>
                  <Badge variant="outline" className="text-xs">
                    完成基础课程
                  </Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">考试时间</span>
                  <span className="text-sm">30分钟</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">通过分数</span>
                  <span className="text-sm">80分</span>
                </div>
                <Button className="w-full" disabled>
                  需要完成前置课程
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Award className="h-5 w-5" />
                  高级信使认证
                </CardTitle>
                <CardDescription>
                  面向L2+信使的高级技能认证
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">前置要求</span>
                  <Badge variant="outline" className="text-xs">
                    完成高级课程
                  </Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">考试时间</span>
                  <span className="text-sm">45分钟</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">通过分数</span>
                  <span className="text-sm">85分</span>
                </div>
                <Button variant="outline" className="w-full" disabled>
                  需要完成前置课程
                </Button>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>认证历史</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-center py-8 text-gray-500">
                <Award className="h-12 w-12 mx-auto mb-2 opacity-50" />
                <p>还没有获得任何认证</p>
                <p className="text-sm">完成相应课程后即可参加认证考试</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}