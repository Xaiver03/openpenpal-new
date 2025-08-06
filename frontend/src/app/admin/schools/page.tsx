'use client'

import React, { useState, useEffect } from 'react'
import { Search, Plus, Edit, Trash2, School, MapPin, Users, FileText, Settings } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface School {
  id: string
  code: string
  name: string
  fullName: string
  province: string
  city: string
  type: string
  level: string
  is985: boolean
  is211: boolean
  isDoubleFirstClass: boolean
  userCount: number
  letterCount: number
  courierCount: number
  status: string
  website?: string
  establishedYear?: number
}

interface SchoolStats {
  totalSchools: number
  totalProvinces: number
  totalCities: number
  totalUsers: number
  totalLetters: number
  total985Schools: number
  total211Schools: number
}

export default function SchoolsAdminPage() {
  const [schools, setSchools] = useState<School[]>([])
  const [stats, setStats] = useState<SchoolStats | null>(null)
  const [provinces, setProvinces] = useState<string[]>([])
  const [types, setTypes] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  // 筛选和搜索状态
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedProvince, setSelectedProvince] = useState('')
  const [selectedType, setSelectedType] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  // 加载学校数据
  const loadSchools = async () => {
    try {
      setLoading(true)
      const params = new URLSearchParams()
      if (searchTerm) params.append('search', searchTerm)
      if (selectedProvince) params.append('province', selectedProvince)
      if (selectedType) params.append('type', selectedType)
      params.append('page', currentPage.toString())
      params.append('limit', '20')

      const response = await fetch(`/api/schools?${params.toString()}`)
      const result = await response.json()
      
      if (result.code === 0) {
        setSchools(result.data.schools)
        setStats(result.data.stats)
        setProvinces(result.data.provinces)
        setTypes(result.data.types)
        setTotalPages(result.data.totalPages)
        setError(null)
      } else {
        setError(result.msg)
      }
    } catch (err) {
      setError('加载学校数据失败')
      console.error('Load schools error:', err)
    } finally {
      setLoading(false)
    }
  }

  // 初始加载和依赖更新
  useEffect(() => {
    loadSchools()
  }, [searchTerm, selectedProvince, selectedType, currentPage])

  const handleSearch = (value: string) => {
    setSearchTerm(value)
    setCurrentPage(1)
  }

  const handleProvinceFilter = (value: string) => {
    setSelectedProvince(value)
    setCurrentPage(1)
  }

  const handleTypeFilter = (value: string) => {
    setSelectedType(value)
    setCurrentPage(1)
  }

  const getSchoolBadges = (school: School) => {
    const badges = []
    if (school.is985) badges.push({ label: '985', variant: 'destructive' })
    if (school.is211) badges.push({ label: '211', variant: 'secondary' })
    if (school.isDoubleFirstClass) badges.push({ label: '双一流', variant: 'default' })
    return badges
  }

  if (loading && schools.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <School className="w-8 h-8" />
            学校管理
          </h1>
          <p className="text-muted-foreground mt-1">
            管理OpenPenPal平台的学校主数据和配置
          </p>
        </div>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          添加学校
        </Button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总学校数</CardTitle>
              <School className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalSchools}</div>
              <p className="text-xs text-muted-foreground">
                覆盖 {stats.totalProvinces} 个省份
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">用户总数</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalUsers}</div>
              <p className="text-xs text-muted-foreground">
                平台注册用户
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">信件总数</CardTitle>
              <FileText className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalLetters}</div>
              <p className="text-xs text-muted-foreground">
                累计信件数量
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">重点院校</CardTitle>
              <Badge className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total985Schools}</div>
              <p className="text-xs text-muted-foreground">
                985/211院校数量
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 搜索和筛选 */}
      <Card>
        <CardHeader>
          <CardTitle>学校列表</CardTitle>
          <CardDescription>
            搜索和管理平台中的所有学校信息
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col sm:flex-row gap-4 mb-6">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="搜索学校名称、城市或编码..."
                value={searchTerm}
                onChange={(e) => handleSearch(e.target.value)}
                className="pl-10"
              />
            </div>

            <Select value={selectedProvince} onValueChange={handleProvinceFilter}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder="选择省份" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">全部省份</SelectItem>
                {provinces.map(province => (
                  <SelectItem key={province} value={province}>
                    {province}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={selectedType} onValueChange={handleTypeFilter}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder="学校类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">全部类型</SelectItem>
                {types.map(type => (
                  <SelectItem key={type} value={type}>
                    {type}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {error && (
            <Alert className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* 学校表格 */}
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>学校信息</TableHead>
                  <TableHead>地理位置</TableHead>
                  <TableHead>类型/属性</TableHead>
                  <TableHead>统计数据</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {schools.map((school) => (
                  <TableRow key={school.id}>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="font-medium">{school.name}</div>
                        <div className="text-sm text-muted-foreground">
                          {school.code} | {school.fullName}
                        </div>
                        {school.establishedYear && (
                          <div className="text-xs text-muted-foreground">
                            建校: {school.establishedYear}年
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1 text-sm">
                        <MapPin className="w-3 h-3" />
                        {school.province} · {school.city}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="space-y-2">
                        <div className="text-sm">{school.type}</div>
                        <div className="flex flex-wrap gap-1">
                          {getSchoolBadges(school).map((badge, index) => (
                            <Badge key={index} variant={badge.variant as any} className="text-xs">
                              {badge.label}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="space-y-1 text-xs">
                        <div>用户: {school.userCount}</div>
                        <div>信件: {school.letterCount}</div>
                        <div>信使: {school.courierCount}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button variant="ghost" size="sm">
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm">
                          <Settings className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm" className="text-destructive">
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-4">
              <div className="text-sm text-muted-foreground">
                第 {currentPage} 页，共 {totalPages} 页
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={currentPage <= 1}
                  onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                >
                  上一页
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={currentPage >= totalPages}
                  onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                >
                  下一页
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}