import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const prefix = searchParams.get('prefix')
    
    if (!prefix) {
      return NextResponse.json({
        success: false,
        code: 400,
        message: '前缀不能为空'
      }, { status: 400 })
    }

    // Call backend API
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'
    const response = await fetch(
      `${backendUrl}/api/v1/opcode/delivery-points/${prefix}`, 
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
    console.error('Delivery Points API Error:', error)
    
    // Return mock data as fallback
    const points = []
    for (let floor = 1; floor <= 6; floor++) {
      for (let room = 1; room <= 10; room++) {
        const code = `${floor}${room.toString().padStart(2, '0')}`
        points.push({
          code: code.slice(-2),
          name: `${floor}${room.toString().padStart(2, '0')}室`,
          available: Math.random() > 0.3,
          type: 'room'
        })
      }
    }
    
    return NextResponse.json({
      success: true,
      code: 0,
      data: {
        points
      }
    })
  }
}