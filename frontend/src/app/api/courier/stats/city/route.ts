import { NextRequest, NextResponse } from 'next/server'
import { JWTUtils } from '@/lib/auth/jwt-utils'
import { COURIER_TEST_ACCOUNTS } from '@/config/courier-test-accounts'

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

      // 检查是否是四级信使
      if (currentCourier.level !== 4) {
        return NextResponse.json({
          success: false,
          error: '权限不足，只有四级信使可以查看城市统计'
        }, { status: 403 })
      }

      // 生成城市级统计数据
      const cityStats = {
        // 基础统计
        total_schools: 24,
        active_couriers: 156,
        total_tasks: 8947,
        completed_tasks: 8654,
        pending_tasks: 293,
        
        // 绩效指标
        success_rate: 96.7,
        avg_delivery_time: '2.3h',
        customer_satisfaction: 4.8,
        
        // 增长数据
        monthly_growth: 12.5,
        weekly_tasks: 1247,
        daily_average: 178,
        
        // 区域分布
        regions: [
          {
            name: '海淀区',
            schools: 8,
            couriers: 62,
            tasks: 3420,
            success_rate: 97.2
          },
          {
            name: '朝阳区',
            schools: 6,
            couriers: 45,
            tasks: 2689,
            success_rate: 96.8
          },
          {
            name: '东城区',
            schools: 5,
            couriers: 28,
            tasks: 1590,
            success_rate: 95.9
          },
          {
            name: '西城区',
            schools: 5,
            couriers: 21,
            tasks: 1248,
            success_rate: 97.5
          }
        ],
        
        // 信使层级分布
        courier_levels: {
          level_1: 89,  // 一级信使
          level_2: 42,  // 二级信使
          level_3: 18,  // 三级信使
          level_4: 7    // 四级信使
        },
        
        // 热门学校
        top_schools: [
          {
            name: '北京大学',
            code: 'PKU001',
            couriers: 23,
            tasks: 1456,
            success_rate: 98.2
          },
          {
            name: '清华大学',
            code: 'THU001',
            couriers: 21,
            tasks: 1389,
            success_rate: 97.8
          },
          {
            name: '中国人民大学',
            code: 'RUC001',
            couriers: 18,
            tasks: 1147,
            success_rate: 96.9
          },
          {
            name: '北京师范大学',
            code: 'BNU001',
            couriers: 15,
            tasks: 967,
            success_rate: 97.3
          }
        ],
        
        // 时间趋势（最近7天）
        daily_trends: generateDailyTrends(),
        
        // 异常情况
        exceptions: {
          total: 23,
          resolved: 21,
          pending: 2,
          types: {
            delivery_delay: 12,
            address_error: 6,
            recipient_unavailable: 3,
            system_error: 2
          }
        },
        
        // 奖励和激励
        rewards_distributed: {
          this_month: 156,
          points_awarded: 23400,
          top_performers: [
            { name: '三级信使A', points: 890, school: '北京大学' },
            { name: '三级信使B', points: 856, school: '清华大学' },
            { name: '三级信使C', points: 823, school: '人民大学' }
          ]
        }
      }

      return NextResponse.json({
        success: true,
        data: cityStats
      })

    } catch (jwtError) {
      return NextResponse.json({
        success: false,
        error: '无效的认证令牌'
      }, { status: 401 })
    }

  } catch (error) {
    console.error('获取城市统计失败:', error)
    return NextResponse.json({
      success: false,
      error: '获取城市统计失败'
    }, { status: 500 })
  }
}

// 生成最近7天的趋势数据
function generateDailyTrends() {
  const trends = []
  const today = new Date()
  
  for (let i = 6; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    
    trends.push({
      date: date.toISOString().split('T')[0],
      tasks: Math.floor(Math.random() * 50) + 150,
      success_rate: (94 + Math.random() * 4).toFixed(1),
      avg_delivery_time: (2.0 + Math.random() * 0.8).toFixed(1) + 'h',
      active_couriers: Math.floor(Math.random() * 20) + 140
    })
  }
  
  return trends
}