import { NextRequest, NextResponse } from 'next/server'
import type { AddressSearchResult } from '@/lib/types/postcode'

// 模拟地址数据库 - 在实际环境中应该从数据库搜索
const MOCK_ADDRESSES: AddressSearchResult[] = [
  {
    postcode: 'PK5F3D',
    fullAddress: '北京大学 第五片区 F栋宿舍 3D宿舍',
    hierarchy: {
      school: {
        id: 'school_pk',
        code: 'PK',
        name: '北京大学',
        fullName: '北京大学',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level4_city'
      },
      area: {
        id: 'area_pk5',
        schoolCode: 'PK',
        code: '5',
        name: '第五片区',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level3_school'
      },
      building: {
        id: 'building_pk5f',
        schoolCode: 'PK',
        areaCode: '5',
        code: 'F',
        name: 'F栋宿舍',
        type: 'dormitory',
        floors: 6,
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level2_zone'
      },
      room: {
        id: 'room_pk5f3d',
        schoolCode: 'PK',
        areaCode: '5',
        buildingCode: 'F',
        code: '3D',
        name: '3D宿舍',
        type: 'dormitory',
        capacity: 4,
        floor: 3,
        fullPostcode: 'PK5F3D',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level1_basic'
      }
    },
    matchScore: 0.95
  },
  {
    postcode: 'PK5F2A',
    fullAddress: '北京大学 第五片区 F栋宿舍 2A宿舍',
    hierarchy: {
      school: {
        id: 'school_pk',
        code: 'PK',
        name: '北京大学',
        fullName: '北京大学',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level4_city'
      },
      area: {
        id: 'area_pk5',
        schoolCode: 'PK',
        code: '5',
        name: '第五片区',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level3_school'
      },
      building: {
        id: 'building_pk5f',
        schoolCode: 'PK',
        areaCode: '5',
        code: 'F',
        name: 'F栋宿舍',
        type: 'dormitory',
        floors: 6,
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level2_zone'
      },
      room: {
        id: 'room_pk5f2a',
        schoolCode: 'PK',
        areaCode: '5',
        buildingCode: 'F',
        code: '2A',
        name: '2A宿舍',
        type: 'dormitory',
        capacity: 4,
        floor: 2,
        fullPostcode: 'PK5F2A',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level1_basic'
      }
    },
    matchScore: 0.93
  },
  {
    postcode: 'PK3A1B',
    fullAddress: '北京大学 第三片区 A栋教学楼 1B教室',
    hierarchy: {
      school: {
        id: 'school_pk',
        code: 'PK',
        name: '北京大学',
        fullName: '北京大学',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level4_city'
      },
      area: {
        id: 'area_pk3',
        schoolCode: 'PK',
        code: '3',
        name: '第三片区',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level3_school'
      },
      building: {
        id: 'building_pk3a',
        schoolCode: 'PK',
        areaCode: '3',
        code: 'A',
        name: 'A栋教学楼',
        type: 'teaching',
        floors: 5,
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level2_zone'
      },
      room: {
        id: 'room_pk3a1b',
        schoolCode: 'PK',
        areaCode: '3',
        buildingCode: 'A',
        code: '1B',
        name: '1B教室',
        type: 'classroom',
        capacity: 50,
        floor: 1,
        fullPostcode: 'PK3A1B',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level1_basic'
      }
    },
    matchScore: 0.88
  },
  {
    postcode: 'QH1C2E',
    fullAddress: '清华大学 第一片区 C栋宿舍 2E宿舍',
    hierarchy: {
      school: {
        id: 'school_qh',
        code: 'QH',
        name: '清华大学',
        fullName: '清华大学',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level4_city'
      },
      area: {
        id: 'area_qh1',
        schoolCode: 'QH',
        code: '1',
        name: '第一片区',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level3_school'
      },
      building: {
        id: 'building_qh1c',
        schoolCode: 'QH',
        areaCode: '1',
        code: 'C',
        name: 'C栋宿舍',
        type: 'dormitory',
        floors: 8,
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level2_zone'
      },
      room: {
        id: 'room_qh1c2e',
        schoolCode: 'QH',
        areaCode: '1',
        buildingCode: 'C',
        code: '2E',
        name: '2E宿舍',
        type: 'dormitory',
        capacity: 4,
        floor: 2,
        fullPostcode: 'QH1C2E',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        managedBy: 'courier_level1_basic'
      }
    },
    matchScore: 0.85
  }
]

// 简单的模糊搜索实现
function fuzzySearch(query: string, addresses: AddressSearchResult[]): AddressSearchResult[] {
  const lowerQuery = query.toLowerCase()
  
  return addresses
    .map(address => {
      let score = 0
      const fullAddress = address.fullAddress.toLowerCase()
      const postcode = address.postcode.toLowerCase()
      
      // 完全匹配获得最高分
      if (fullAddress.includes(lowerQuery)) {
        score += 1.0
      }
      
      // Postcode 匹配
      if (postcode.includes(lowerQuery)) {
        score += 0.8
      }
      
      // 学校名匹配
      if (address.hierarchy.school.name.toLowerCase().includes(lowerQuery)) {
        score += 0.6
      }
      
      // 楼栋名匹配
      if (address.hierarchy.building?.name.toLowerCase().includes(lowerQuery)) {
        score += 0.4
      }
      
      // 房间名匹配
      if (address.hierarchy.room?.name.toLowerCase().includes(lowerQuery)) {
        score += 0.3
      }
      
      return {
        ...address,
        matchScore: score
      }
    })
    .filter(address => address.matchScore > 0)
    .sort((a, b) => b.matchScore - a.matchScore)
}

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const query = searchParams.get('q')
    const limit = parseInt(searchParams.get('limit') || '10')
    
    if (!query || query.trim().length === 0) {
      return NextResponse.json({
        code: 400,
        message: '搜索关键词不能为空',
        data: [],
        timestamp: new Date().toISOString()
      }, { status: 400 })
    }

    if (query.trim().length < 2) {
      return NextResponse.json({
        code: 400,
        message: '搜索关键词至少需要2个字符',
        data: [],
        timestamp: new Date().toISOString()
      }, { status: 400 })
    }

    // 首先尝试网关服务
    const gatewayUrl = process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080'
    
    try {
      const response = await fetch(`${gatewayUrl}/api/v1/address/search?q=${encodeURIComponent(query)}&limit=${limit}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        signal: AbortSignal.timeout(5000)
      })
      
      const result = await response.json()
      
      if (result.success) {
        return NextResponse.json({
          code: 0,
          message: '地址搜索成功',
          data: result.data,
          timestamp: new Date().toISOString()
        })
      } else {
        throw new Error('网关搜索失败，使用本地搜索')
      }
    } catch (error) {
      console.log('网关服务不可用，使用本地模拟搜索')
      
      // 本地模糊搜索
      const results = fuzzySearch(query, MOCK_ADDRESSES)
      const limitedResults = results.slice(0, limit)

      return NextResponse.json({
        code: 0,
        message: '地址搜索成功',
        data: limitedResults,
        timestamp: new Date().toISOString()
      })
    }
    
  } catch (error) {
    console.error('地址搜索错误:', error)
    return NextResponse.json({
      code: 500,
      message: '地址搜索失败，请稍后重试',
      data: [],
      timestamp: new Date().toISOString()
    }, { status: 500 })
  }
}