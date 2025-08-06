'use client'

import React, { useState, useEffect, useMemo } from 'react'
import { Search, MapPin, School, ChevronDown, Check, AlertCircle, Loader2 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import SchoolService, { type School as SchoolType } from '@/lib/services/school-service'

interface School {
  code: string
  name: string
  fullName: string
  province: string
  city: string
  type: string
  isActive: boolean
}

interface SchoolSelectorProps {
  value?: string
  onChange: (schoolCode: string, schoolName: string) => void
  placeholder?: string
  className?: string
  required?: boolean
  disabled?: boolean
  error?: string
}

// 转换API数据格式到组件格式
function transformSchoolData(apiSchool: SchoolType): School {
  return {
    code: apiSchool.code,
    name: apiSchool.name,
    fullName: apiSchool.name,
    province: apiSchool.province,
    city: apiSchool.city,
    type: apiSchool.type,
    isActive: apiSchool.status === 'active'
  }
}

// 使用新的SchoolService获取学校数据
async function fetchSchools(searchParams: {
  keyword?: string
  province?: string
  page?: number
  limit?: number
} = {}): Promise<{
  schools: School[]
  provinces: string[]
  total: number
}> {
  try {
    // 并行获取学校数据和省份列表
    const [schoolsResponse, provincesResponse] = await Promise.all([
      SchoolService.searchSchools({
        keyword: searchParams.keyword,
        province: searchParams.province,
        page: searchParams.page || 1,
        limit: searchParams.limit || 50,
        sort_by: 'name',
        sort_order: 'asc'
      }),
      SchoolService.getProvinces()
    ])
    
    if (schoolsResponse.success && provincesResponse.success && schoolsResponse.data && provincesResponse.data) {
      return {
        schools: schoolsResponse.data.items.map(transformSchoolData),
        provinces: provincesResponse.data,
        total: schoolsResponse.data.total
      }
    } else {
      throw new Error('Failed to fetch schools data')
    }
  } catch (error) {
    console.error('Failed to fetch schools:', error)
    
    // 使用fallback数据
    return {
      schools: [
        {
          code: 'BJDX01',
          name: '北京大学',
          fullName: '北京大学',
          province: '北京',
          city: '北京',
          type: 'university',
          isActive: true
        },
        {
          code: 'THU001',
          name: '清华大学',
          fullName: '清华大学',
          province: '北京',
          city: '北京',
          type: 'university',
          isActive: true
        }
      ],
      provinces: ['北京', '上海', '广东', '江苏', '浙江'],
      total: 2
    }
  }
}

export function SchoolSelector({ 
  value, 
  onChange, 
  placeholder = "请选择您的学校", 
  className = "",
  required = false,
  disabled = false,
  error
}: SchoolSelectorProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedProvince, setSelectedProvince] = useState('')
  const [selectedSchool, setSelectedSchool] = useState<School | null>(null)
  const [schools, setSchools] = useState<School[]>([])
  const [provinces, setProvinces] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)

  // 使用新的API加载学校数据
  const loadSchools = async (searchParams: {
    keyword?: string
    province?: string
  } = {}) => {
    setIsLoading(true)
    setLoadError(null)
    try {
      const data = await fetchSchools(searchParams)
      setSchools(data.schools)
      setProvinces(data.provinces)
    } catch (error) {
      console.error('Failed to load schools:', error)
      setLoadError('加载学校数据失败，请稍后重试')
    } finally {
      setIsLoading(false)
    }
  }

  // 初始加载数据
  useEffect(() => {
    loadSchools()
  }, [])

  // 根据value找到对应的学校
  useEffect(() => {
    if (value && schools.length > 0) {
      const school = schools.find(s => s.code === value)
      setSelectedSchool(school || null)
    } else {
      setSelectedSchool(null)
    }
  }, [value, schools])

  // 搜索和筛选时重新加载数据
  useEffect(() => {
    loadSchools({
      keyword: searchTerm,
      province: selectedProvince
    })
  }, [searchTerm, selectedProvince])

  // 按省份分组的学校
  const schoolsByProvince = useMemo(() => {
    const groups = schools.reduce((acc, school) => {
      if (!acc[school.province]) {
        acc[school.province] = []
      }
      acc[school.province].push(school)
      return acc
    }, {} as Record<string, School[]>)

    // 对每个省份内的学校按名称排序
    Object.keys(groups).forEach(province => {
      groups[province].sort((a, b) => a.name.localeCompare(b.name))
    })

    return groups
  }, [schools])

  const handleSchoolSelect = (school: School) => {
    setSelectedSchool(school)
    onChange(school.code, school.name)
    setIsOpen(false)
    setSearchTerm('')
    setSelectedProvince('')
  }

  const handleClear = () => {
    setSelectedSchool(null)
    onChange('', '')
  }

  return (
    <div className={className}>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>
          <Button
            variant="outline"
            className={`w-full justify-between h-10 px-3 py-2 ${
              !selectedSchool ? 'text-muted-foreground' : ''
            }`}
          >
            <div className="flex items-center gap-2">
              <School className="w-4 h-4" />
              <span className="truncate">
                {selectedSchool ? selectedSchool.name : placeholder}
              </span>
              {selectedSchool && (
                <Badge variant="secondary" className="text-xs">
                  {selectedSchool.code}
                </Badge>
              )}
            </div>
            <ChevronDown className="w-4 h-4 opacity-50" />
          </Button>
        </DialogTrigger>

        <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-hidden">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <School className="w-5 h-5" />
              选择学校
            </DialogTitle>
            <DialogDescription>
              搜索并选择您所在的学校，系统会自动分配相应的学校编码
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {/* 搜索框 */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="搜索学校名称、城市或编码..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* 省份筛选 */}
            <div className="flex flex-wrap gap-2">
              <Button
                variant={selectedProvince === '' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setSelectedProvince('')}
              >
                全部省份
              </Button>
              {provinces.map(province => (
                <Button
                  key={province}
                  variant={selectedProvince === province ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setSelectedProvince(province)}
                >
                  {province}
                </Button>
              ))}
            </div>

            {/* 学校列表 */}
            <div className="max-h-96 overflow-y-auto space-y-4">
              {Object.keys(schoolsByProvince).length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <School className="w-12 h-12 mx-auto mb-4 opacity-50" />
                  <p>未找到匹配的学校</p>
                  <p className="text-sm">请尝试调整搜索条件</p>
                </div>
              ) : (
                Object.entries(schoolsByProvince).map(([province, schools]) => (
                  <div key={province} className="space-y-2">
                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                      <MapPin className="w-4 h-4" />
                      {province}
                    </div>
                    <div className="space-y-1">
                      {schools.map(school => (
                        <Card
                          key={school.code}
                          className={`cursor-pointer transition-all duration-200 hover:shadow-md ${
                            selectedSchool?.code === school.code
                              ? 'border-primary bg-primary/5'
                              : 'hover:border-primary/50'
                          }`}
                          onClick={() => handleSchoolSelect(school)}
                        >
                          <CardContent className="p-3">
                            <div className="flex items-center justify-between">
                              <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2 mb-1">
                                  <h4 className="font-medium text-sm truncate">
                                    {school.name}
                                  </h4>
                                  <Badge variant="outline" className="text-xs">
                                    {school.code}
                                  </Badge>
                                  <Badge variant="secondary" className="text-xs">
                                    {school.type}
                                  </Badge>
                                </div>
                                <p className="text-xs text-muted-foreground">
                                  {school.province} · {school.city}
                                </p>
                              </div>
                              {selectedSchool?.code === school.code && (
                                <Check className="w-4 h-4 text-primary" />
                              )}
                            </div>
                          </CardContent>
                        </Card>
                      ))}
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* 底部操作 */}
            {selectedSchool && (
              <div className="flex items-center justify-between pt-4 border-t">
                <div className="text-sm text-muted-foreground">
                  已选择: {selectedSchool.name} ({selectedSchool.code})
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={handleClear}>
                    清除选择
                  </Button>
                  <Button size="sm" onClick={() => setIsOpen(false)}>
                    确认选择
                  </Button>
                </div>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {required && !selectedSchool && (
        <p className="text-sm text-destructive mt-1">请选择您的学校</p>
      )}
    </div>
  )
}