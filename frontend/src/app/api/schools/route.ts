import { NextRequest, NextResponse } from 'next/server'
import { searchSchools, getProvinces, getSchoolTypes, getSchoolStats } from '@/lib/database'
import type { SchoolSearchParams } from '@/lib/database'

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    
    // 构建查询参数
    const queryParams: SchoolSearchParams = {
      search: searchParams.get('search') || undefined,
      province: searchParams.get('province') || undefined,
      city: searchParams.get('city') || undefined,
      school_type: searchParams.get('type') || undefined,
      school_level: searchParams.get('level') || undefined,
      page: parseInt(searchParams.get('page') || '1'),
      limit: parseInt(searchParams.get('limit') || '50')
    }

    // 特殊属性过滤
    if (searchParams.get('is_985') === 'true') {
      queryParams.is_985 = true
    }
    if (searchParams.get('is_211') === 'true') {
      queryParams.is_211 = true
    }
    if (searchParams.get('is_double_first_class') === 'true') {
      queryParams.is_double_first_class = true
    }

    // 并行获取数据
    const [schoolsResult, provinces, types, stats] = await Promise.all([
      searchSchools(queryParams),
      getProvinces(),
      getSchoolTypes(),
      getSchoolStats()
    ])

    // 转换数据格式以匹配前端期望
    const transformedSchools = schoolsResult.data.map(school => ({
      code: school.code,
      name: school.name,
      fullName: school.full_name,
      englishName: school.english_name,
      province: school.province,
      city: school.city,
      district: school.district,
      address: school.address,
      type: school.school_type,
      level: school.school_level,
      is985: school.is_985,
      is211: school.is_211,
      isDoubleFirstClass: school.is_double_first_class,
      isActive: school.status === 'active',
      website: school.website,
      establishedYear: school.established_year,
      userCount: school.user_count,
      letterCount: school.letter_count,
      courierCount: school.courier_count
    }))

    return NextResponse.json({
      code: 0,
      msg: 'success',
      data: {
        schools: transformedSchools,
        total: schoolsResult.total,
        page: schoolsResult.page,
        limit: schoolsResult.limit,
        totalPages: schoolsResult.totalPages,
        provinces,
        types,
        stats
      },
      timestamp: new Date().toISOString()
    })

  } catch (error) {
    console.error('Schools API Error:', error)
    
    // 数据库连接失败时的降级处理 - 使用内存数据
    if (error instanceof Error && error.message.includes('connect')) {
      console.warn('Database connection failed, falling back to mock data')
      return getFallbackData(request)
    }
    
    return NextResponse.json({
      code: 500,
      msg: '服务器内部错误',
      data: null,
      timestamp: new Date().toISOString()
    }, { status: 500 })
  }
}

// 降级数据 - 当数据库不可用时使用
async function getFallbackData(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const search = searchParams.get('search')?.toLowerCase()
  const province = searchParams.get('province')
  const type = searchParams.get('type')
  const page = parseInt(searchParams.get('page') || '1')
  const limit = parseInt(searchParams.get('limit') || '50')

  // 内存中的备用数据
  const FALLBACK_SCHOOLS = [
    {
      code: 'BJDX01',
      name: '北京大学',
      fullName: '北京大学',
      province: '北京',
      city: '北京',
      type: '综合类',
      isActive: true,
      website: 'https://www.pku.edu.cn',
      establishedYear: 1898
    },
    {
      code: 'QHDX01',
      name: '清华大学',
      fullName: '清华大学',
      province: '北京',
      city: '北京',
      type: '理工类',
      isActive: true,
      website: 'https://www.tsinghua.edu.cn',
      establishedYear: 1911
    },
    {
      code: 'FDDX01',
      name: '复旦大学',
      fullName: '复旦大学',
      province: '上海',
      city: '上海',
      type: '综合类',
      isActive: true,
      website: 'https://www.fudan.edu.cn',
      establishedYear: 1905
    }
  ]

  let filteredSchools = FALLBACK_SCHOOLS.filter(school => school.isActive)

  // 简单过滤逻辑
  if (search) {
    filteredSchools = filteredSchools.filter(school => 
      school.name.toLowerCase().includes(search) ||
      school.fullName.toLowerCase().includes(search) ||
      school.city.toLowerCase().includes(search) ||
      school.code.toLowerCase().includes(search)
    )
  }

  if (province) {
    filteredSchools = filteredSchools.filter(school => 
      school.province === province
    )
  }

  if (type) {
    filteredSchools = filteredSchools.filter(school => 
      school.type === type
    )
  }

  const total = filteredSchools.length
  const startIndex = (page - 1) * limit
  const endIndex = startIndex + limit
  const paginatedSchools = filteredSchools.slice(startIndex, endIndex)

  const provinces = [...new Set(FALLBACK_SCHOOLS.map(school => school.province))].sort()
  const types = [...new Set(FALLBACK_SCHOOLS.map(school => school.type))].sort()

  return NextResponse.json({
    code: 0,
    msg: 'success (fallback mode)',
    data: {
      schools: paginatedSchools,
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
      provinces,
      types,
      stats: {
        totalSchools: FALLBACK_SCHOOLS.length,
        totalProvinces: provinces.length,
        totalCities: [...new Set(FALLBACK_SCHOOLS.map(s => s.city))].length
      }
    },
    timestamp: new Date().toISOString()
  })
}