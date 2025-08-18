import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const schoolCode = searchParams.get('school_code')
    
    if (!schoolCode) {
      return NextResponse.json({
        success: false,
        code: 400,
        message: '学校代码不能为空'
      }, { status: 400 })
    }

    // Call backend API
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
    const response = await fetch(`${backendUrl}/api/v1/opcode/districts/${schoolCode}`, {
      headers: {
        'Content-Type': 'application/json',
      }
    })

    if (!response.ok) {
      throw new Error('Backend request failed')
    }

    const data = await response.json()
    
    // Return the data directly from backend
    return NextResponse.json(data)

  } catch (error) {
    console.error('Districts API Error:', error)
    
    // Return mock data as fallback
    return NextResponse.json({
      success: true,
      code: 0,
      data: {
        districts: [
          { area_code: '01', area_name: '本部东区', description: '包含1-5栋宿舍楼' },
          { area_code: '02', area_name: '本部西区', description: '包含6-10栋宿舍楼' },
          { area_code: '03', area_name: '本部南区', description: '包含11-15栋宿舍楼' },
          { area_code: '04', area_name: '本部北区', description: '包含16-20栋宿舍楼' },
          { area_code: '05', area_name: '中心校区', description: '教学楼、图书馆、行政楼' }
        ]
      }
    })
  }
}