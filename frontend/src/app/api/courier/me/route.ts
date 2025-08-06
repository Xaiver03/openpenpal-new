import { NextRequest, NextResponse } from 'next/server'
import { COURIER_TEST_ACCOUNTS } from '@/config/courier-test-accounts'
import { requireCourier, PermissionMiddleware, AuthenticatedRequest } from '@/lib/middleware/permissions'

// 使用统一权限中间件保护API
export const GET = requireCourier(async function(request: NextRequest) {
  try {
    // 从中间件获取已验证的用户信息
    const user = (request as AuthenticatedRequest).user!
    
    // 查找对应的信使信息
    const courierAccount = COURIER_TEST_ACCOUNTS.find(
      account => account.username === user.username
    )
    
    if (!courierAccount) {
      return PermissionMiddleware.createResponse(
        404,
        '信使信息不存在'
      )
    }

    // 生成模拟的信使详细信息
    const courierInfo = {
      id: `courier_${courierAccount.username}`,
      userId: user.userId,
      username: courierAccount.username,
      level: courierAccount.level,
      parent_id: getParentId(courierAccount),
      region: courierAccount.zoneCode,
      zone_type: courierAccount.zoneType,
      status: 'active',
      total_points: Math.floor(Math.random() * 1000) + 500,
      completed_tasks: Math.floor(Math.random() * 50) + 20,
      success_rate: (95 + Math.random() * 4).toFixed(1), // 95-99%
      avg_delivery_time: '1.2h',
      join_date: '2024-01-15T00:00:00Z',
      last_active: new Date().toISOString(),
      // 管理的下级信使数量
      subordinate_count: getSubordinateCount(courierAccount.level),
      // 本月完成任务数
      monthly_tasks: Math.floor(Math.random() * 20) + 10,
      // 获得的奖励
      rewards: generateRewards(courierAccount.level)
    }

    return PermissionMiddleware.createResponse(
      200,
      '获取信使信息成功',
      courierInfo
    )

  } catch (error) {
    console.error('获取信使信息失败:', error)
    return PermissionMiddleware.createResponse(
      500,
      '获取信使信息失败'
    )
  }
})

// 获取上级信使ID
function getParentId(courierAccount: any): string | null {
  const parentAccount = COURIER_TEST_ACCOUNTS.find(
    account => account.username === courierAccount.parentUsername
  )
  return parentAccount ? `courier_${parentAccount.username}` : null
}

// 根据级别获取下级信使数量
function getSubordinateCount(level: number): number {
  switch (level) {
    case 4: return Math.floor(Math.random() * 5) + 3 // 3-7个三级信使
    case 3: return Math.floor(Math.random() * 8) + 5 // 5-12个二级信使
    case 2: return Math.floor(Math.random() * 10) + 8 // 8-17个一级信使
    case 1: return 0 // 一级信使无下级
    default: return 0
  }
}

// 生成奖励信息
function generateRewards(level: number) {
  const baseRewards = [
    { name: '投递达人', icon: '🚚', description: '完成100次投递' },
    { name: '准时之星', icon: '⭐', description: '准时率超过95%' }
  ]

  const levelRewards = {
    4: [
      { name: '城市之王', icon: '👑', description: '管理整个城市信使网络' },
      { name: '卓越领导', icon: '🏆', description: '城市级运营优秀' }
    ],
    3: [
      { name: '校园守护', icon: '🏫', description: '守护校园信件传递' },
      { name: '团队建设者', icon: '👥', description: '建设优秀信使团队' }
    ],
    2: [
      { name: '片区之光', icon: '✨', description: '片区管理出色' },
      { name: '协调专家', icon: '🤝', description: '协调能力优秀' }
    ],
    1: [
      { name: '勤劳蜜蜂', icon: '🐝', description: '勤勉投递每一封信' }
    ]
  }

  return [...baseRewards, ...(levelRewards[level as keyof typeof levelRewards] || [])]
}