import { NextRequest, NextResponse } from 'next/server'
import { validateSchoolCode, getSchoolByCode } from '@/lib/database'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { schoolCode } = body

    if (!school_code) {
      return NextResponse.json({
        code: 400,
        msg: '学校代码不能为空',
        data: null,
        timestamp: new Date().toISOString()
      }, { status: 400 })
    }

    const codeUpper = school_code.toUpperCase()
    
    // 验证代码格式
    if (!/^[A-Z0-9]{4,10}$/.test(codeUpper)) {
      return NextResponse.json({
        code: 400,
        msg: '学校代码格式不正确',
        data: {
          valid: false,
          schoolCode: codeUpper,
          suggestion: '学校代码应为4-10位大写字母和数字组合'
        },
        timestamp: new Date().toISOString()
      }, { status: 400 })
    }

    try {
      // 从数据库验证学校代码
      const isValid = await validateSchoolCode(codeUpper)
      
      if (!isValid) {
        return NextResponse.json({
          code: 400,
          msg: '无效的学校代码',
          data: {
            valid: false,
            schoolCode: codeUpper,
            suggestion: '请从学校列表中选择正确的学校'
          },
          timestamp: new Date().toISOString()
        }, { status: 400 })
      }

      // 获取学校详细信息
      const school = await getSchoolByCode(codeUpper)

      return NextResponse.json({
        code: 0,
        msg: '学校代码验证成功',
        data: {
          valid: true,
          schoolCode: codeUpper,
          school_name: school?.name,
          school_full_name: school?.fullName,
          province: school?.province,
          city: school?.city
        },
        timestamp: new Date().toISOString()
      })

    } catch (dbError) {
      console.error('Database validation error:', dbError)
      
      // 数据库不可用时的降级处理
      const FALLBACK_CODES = [
        'BJDX01', 'QHDX01', 'FDDX01', 'JDDX01', 'ZJDX01', 'NJDX01',
        'HZDX01', 'XADX01', 'SCDX01', 'ZSDX01', 'HNDX01', 'DLDX01',
        'BJLG01', 'BJHG01', 'TJDX01'
      ]
      
      const isValidFallback = FALLBACK_CODES.includes(codeUpper)
      
      if (!isValidFallback) {
        return NextResponse.json({
          code: 400,
          msg: '无效的学校代码 (降级模式)',
          data: {
            valid: false,
            schoolCode: codeUpper,
            suggestion: '请从学校列表中选择正确的学校'
          },
          timestamp: new Date().toISOString()
        }, { status: 400 })
      }

      return NextResponse.json({
        code: 0,
        msg: '学校代码验证成功 (降级模式)',
        data: {
          valid: true,
          schoolCode: codeUpper,
          fallback_mode: true
        },
        timestamp: new Date().toISOString()
      })
    }

  } catch (error) {
    console.error('School validation error:', error)
    return NextResponse.json({
      code: 500,
      msg: '服务器内部错误',
      data: null,
      timestamp: new Date().toISOString()
    }, { status: 500 })
  }
}