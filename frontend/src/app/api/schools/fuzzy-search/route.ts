import { NextRequest, NextResponse } from 'next/server'
import { query } from '@/lib/database'

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const keyword = searchParams.get('keyword')
    const limit = parseInt(searchParams.get('limit') || '20')
    
    console.log('Fuzzy search API called with keyword:', keyword)
    
    if (!keyword || keyword.trim().length === 0) {
      return NextResponse.json({
        code: 400,
        msg: '请输入搜索关键词',
        data: null
      }, { status: 400 })
    }

    // Fuzzy search across multiple fields in op_code_schools table
    const schools = await query(`
      SELECT 
        s.id,
        s.school_code,
        s.school_name,
        s.full_name,
        s.city,
        s.province,
        s.is_active
      FROM op_code_schools s
      WHERE 
        s.is_active = true
        AND (
          s.school_name ILIKE $1
          OR s.full_name ILIKE $1
          OR s.city ILIKE $1
          OR s.province ILIKE $1
          OR s.school_code ILIKE $1
        )
      ORDER BY 
        CASE 
          WHEN s.school_name ILIKE $2 THEN 1
          WHEN s.city ILIKE $2 THEN 2
          WHEN s.province ILIKE $2 THEN 3
          ELSE 4
        END,
        s.school_name ASC
      LIMIT $3
    `, [`%${keyword}%`, `${keyword}%`, limit])

    console.log('Query results:', schools.length, 'schools found')

    // Group by relevance
    const exactMatches = schools.filter(s => 
      s.school_name.includes(keyword) || 
      s.city === keyword || 
      s.province === keyword
    )
    
    const partialMatches = schools.filter(s => 
      !exactMatches.includes(s)
    )

    return NextResponse.json({
      code: 0,
      msg: 'success',
      data: {
        schools,
        total: schools.length,
        keyword,
        exactMatches: exactMatches.length,
        partialMatches: partialMatches.length
      },
      timestamp: new Date().toISOString()
    })

  } catch (error) {
    console.error('Fuzzy search error:', error)
    
    // Fallback data for demonstration
    if (error instanceof Error && error.message.includes('connect')) {
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

// Fallback data when database is unavailable
function getFallbackData(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const keyword = searchParams.get('keyword')?.toLowerCase() || ''
  
  const mockSchools = [
    {
      school_code: 'CS01',
      school_name: '中南大学',
      full_name: '中南大学',
      province: '湖南',
      city: '长沙',
      is_985: true,
      is_211: true,
      is_double_first_class: true
    },
    {
      school_code: 'CS02',
      school_name: '湖南大学',
      full_name: '湖南大学',
      province: '湖南',
      city: '长沙',
      is_985: true,
      is_211: true,
      is_double_first_class: true
    },
    {
      school_code: 'CS03',
      school_name: '湖南师范大学',
      full_name: '湖南师范大学',
      province: '湖南',
      city: '长沙',
      is_985: false,
      is_211: true,
      is_double_first_class: true
    },
    {
      school_code: 'CS04',
      school_name: '长沙理工大学',
      full_name: '长沙理工大学',
      province: '湖南',
      city: '长沙',
      is_985: false,
      is_211: false,
      is_double_first_class: false
    },
    {
      school_code: 'BJ01',
      school_name: '北京大学',
      full_name: '北京大学',
      province: '北京',
      city: '北京',
      is_985: true,
      is_211: true,
      is_double_first_class: true
    },
    {
      school_code: 'SH01',
      school_name: '复旦大学',
      full_name: '复旦大学',
      province: '上海',
      city: '上海',
      is_985: true,
      is_211: true,
      is_double_first_class: true
    }
  ]
  
  const filtered = mockSchools.filter(school => 
    school.school_name.includes(keyword) ||
    school.city.includes(keyword) ||
    school.province.includes(keyword)
  )
  
  return NextResponse.json({
    code: 0,
    msg: 'success (fallback mode)',
    data: {
      schools: filtered,
      total: filtered.length,
      keyword,
      exactMatches: filtered.filter(s => s.city.includes(keyword)).length,
      partialMatches: filtered.filter(s => !s.city.includes(keyword)).length
    },
    timestamp: new Date().toISOString()
  })
}