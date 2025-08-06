import { NextRequest, NextResponse } from 'next/server'
import { PostcodeService } from '@/lib/services/postcode-service'
import type { AddressSearchResult } from '@/lib/types/postcode'

// 模拟数据 - 在实际环境中应该从数据库获取
const MOCK_ADDRESS_DATA: Record<string, AddressSearchResult> = {
  'PK5F3D': {
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
    matchScore: 1.0
  },
  'PK5F2A': {
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
    matchScore: 1.0
  }
}

export async function GET(
  request: NextRequest,
  { params }: { params: { code: string } }
) {
  try {
    const { code } = params
    
    // 验证编码格式
    const validation = PostcodeService.validatePostcode(code)
    if (!validation.isValid) {
      return NextResponse.json({
        code: 400,
        message: validation.errors.join(', '),
        data: null,
        timestamp: new Date().toISOString()
      }, { status: 400 })
    }

    // 首先尝试网关服务
    const gatewayUrl = process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080'
    
    try {
      const response = await fetch(`${gatewayUrl}/api/v1/postcode/${code}`, {
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
          message: '地址查询成功',
          data: result.data,
          timestamp: new Date().toISOString()
        })
      } else {
        throw new Error('网关查询失败，使用本地数据')
      }
    } catch (error) {
      console.log('网关服务不可用，使用本地模拟数据')
      
      // 本地模拟数据查询
      const addressData = MOCK_ADDRESS_DATA[code.toUpperCase()]
      
      if (!addressData) {
        return NextResponse.json({
          code: 404,
          message: '未找到对应的地址信息',
          data: null,
          timestamp: new Date().toISOString()
        }, { status: 404 })
      }

      return NextResponse.json({
        code: 0,
        message: '地址查询成功',
        data: addressData,
        timestamp: new Date().toISOString()
      })
    }
    
  } catch (error) {
    console.error('Postcode查询错误:', error)
    return NextResponse.json({
      code: 500,
      message: '地址查询失败，请稍后重试',
      data: null,
      timestamp: new Date().toISOString()
    }, { status: 500 })
  }
}