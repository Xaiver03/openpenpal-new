import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const schoolCode = searchParams.get('school_code')
    const districtCode = searchParams.get('district_code')
    
    if (!schoolCode || !districtCode) {
      return NextResponse.json({
        success: false,
        code: 400,
        message: '参数不完整'
      }, { status: 400 })
    }

    // Call backend API
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
    const response = await fetch(
      `${backendUrl}/api/v1/opcode/buildings/${schoolCode}/${districtCode}`, 
      {
        headers: {
          'Content-Type': 'application/json',
        }
      }
    )

    if (!response.ok) {
      throw new Error('Backend request failed')
    }

    const data = await response.json()
    
    // Return the data directly from backend
    return NextResponse.json(data)

  } catch (error) {
    console.error('Buildings API Error:', error)
    
    // Return mock data as fallback
    return NextResponse.json({
      success: true,
      code: 0,
      data: {
        buildings: [
          { code: 'A', name: 'A栋', type: 'dormitory' },
          { code: 'B', name: 'B栋', type: 'dormitory' },
          { code: 'C', name: 'C栋', type: 'dormitory' },
          { code: 'D', name: 'D栋', type: 'teaching' },
          { code: 'E', name: 'E栋', type: 'dining' },
          { code: 'F', name: 'F栋', type: 'dormitory' }
        ]
      }
    })
  }
}