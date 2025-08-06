import { NextRequest, NextResponse } from 'next/server'
import { JWTUtils } from '@/lib/auth/jwt-utils'
import { COURIER_TEST_ACCOUNTS, getSubordinateAccounts } from '@/config/courier-test-accounts'

export async function GET(request: NextRequest) {
  try {
    // 获取认证信息
    const authHeader = request.headers.get('authorization')
    if (!authHeader?.startsWith('Bearer ')) {
      return NextResponse.json({
        success: false,
        error: '未提供认证令牌'
      }, { status: 401 })
    }

    const token = authHeader.substring(7)
    
    try {
      const payload = JWTUtils.verifyAccessToken(token)
      
      // 查找当前信使信息
      const currentCourier = COURIER_TEST_ACCOUNTS.find(
        account => account.username === payload.username
      )
      
      if (!currentCourier) {
        return NextResponse.json({
          success: false,
          error: '信使信息不存在'
        }, { status: 404 })
      }

      // 检查是否有管理权限（2级及以上）
      if (currentCourier.level < 2) {
        return NextResponse.json({
          success: false,
          error: '权限不足，无法查看下级信使'
        }, { status: 403 })
      }

      // 获取下级信使列表
      const subordinates = getSubordinateAccounts(currentCourier.username)
      
      // 生成下级信使的详细信息
      const subordinateDetails = subordinates.map(subordinate => ({
        id: `courier_${subordinate.username}`,
        name: subordinate.levelName,
        username: subordinate.username,
        email: subordinate.email,
        level: subordinate.level,
        region: subordinate.zoneCode,
        zone_type: subordinate.zoneType,
        zone_name: getZoneDisplayName(subordinate.zoneCode),
        status: 'active',
        total_points: Math.floor(Math.random() * 800) + 200,
        completed_tasks: Math.floor(Math.random() * 40) + 15,
        success_rate: (92 + Math.random() * 6).toFixed(1), // 92-98%
        avg_delivery_time: `${(1.0 + Math.random()).toFixed(1)}h`,
        join_date: getRandomJoinDate(),
        last_active: getRandomLastActive(),
        // 本月任务数
        monthly_tasks: Math.floor(Math.random() * 15) + 8,
        // 管理的下级数量（如果有的话）
        subordinate_count: subordinate.level > 1 ? Math.floor(Math.random() * 5) + 2 : 0
      }))

      return NextResponse.json({
        success: true,
        data: {
          couriers: subordinateDetails,
          total: subordinateDetails.length,
          manager_level: currentCourier.level,
          manager_zone: currentCourier.zoneCode
        }
      })

    } catch (jwtError) {
      return NextResponse.json({
        success: false,
        error: '无效的认证令牌'
      }, { status: 401 })
    }

  } catch (error) {
    console.error('获取下级信使信息失败:', error)
    return NextResponse.json({
      success: false,
      error: '获取下级信使信息失败'
    }, { status: 500 })
  }
}

// 获取区域显示名称
function getZoneDisplayName(zoneCode: string): string {
  const zoneMap: Record<string, string> = {
    'BJDX_BUILDING_32': '32号楼',
    'BJDX_BUILDING_33': '33号楼',
    'BJDX_ZONE_A': '北大A区',
    'BJDX_ZONE_B': '北大B区',
    'BJDX': '北京大学',
    'BEIJING': '北京市'
  }
  return zoneMap[zoneCode] || zoneCode
}

// 生成随机加入日期
function getRandomJoinDate(): string {
  const start = new Date('2024-01-01')
  const end = new Date('2024-01-20')
  const randomDate = new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()))
  return randomDate.toISOString()
}

// 生成随机最后活跃时间
function getRandomLastActive(): string {
  const now = new Date()
  const hoursAgo = Math.floor(Math.random() * 24) // 0-24小时前
  const lastActive = new Date(now.getTime() - hoursAgo * 60 * 60 * 1000)
  return lastActive.toISOString()
}