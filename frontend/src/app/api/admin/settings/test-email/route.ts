import { NextRequest, NextResponse } from 'next/server'
import nodemailer from 'nodemailer'
import { withSuperAdminAuth } from '@/lib/middleware/admin'
import { AuthenticatedRequest } from '@/lib/middleware/auth'

// POST - 测试邮件配置（需要超级管理员权限）
async function testEmailConfig(request: AuthenticatedRequest) {
  try {
    const user = request.user!
    const body = await request.json()
    const { smtp_host, smtp_port, smtp_username, smtp_password, smtp_encryption, email_from_address, test_email } = body
    
    console.log(`管理员 ${user.username} 正在测试邮件配置到: ${test_email}`)
    
    if (!test_email) {
      return NextResponse.json({
        code: 400,
        message: '请提供测试邮箱地址',
        data: null
      }, { status: 400 })
    }
    
    // 创建邮件传输配置
    const transporter = nodemailer.createTransport({
      host: smtp_host,
      port: smtp_port,
      secure: smtp_encryption === 'ssl', // true for SSL, false for TLS
      auth: {
        user: smtp_username,
        pass: smtp_password,
      },
      tls: {
        rejectUnauthorized: false // 允许自签名证书（仅用于测试）
      }
    })
    
    // 验证SMTP连接
    await transporter.verify()
    
    // 发送测试邮件
    const mailOptions = {
      from: `"OpenPenPal System" <${email_from_address}>`,
      to: test_email,
      subject: '邮件配置测试 - OpenPenPal',
      html: `
        <div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
          <h2 style="color: #333;">邮件配置测试成功！</h2>
          <p>这是来自 OpenPenPal 系统的测试邮件。</p>
          <div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <h3>配置信息：</h3>
            <ul>
              <li>SMTP服务器: ${smtp_host}</li>
              <li>端口: ${smtp_port}</li>
              <li>加密方式: ${smtp_encryption}</li>
              <li>发送时间: ${new Date().toLocaleString('zh-CN')}</li>
            </ul>
          </div>
          <p style="color: #666;">如果您收到这封邮件，说明您的邮件配置正确！</p>
        </div>
      `
    }
    
    await transporter.sendMail(mailOptions)
    
    console.log(`管理员 ${user.username} 测试邮件发送成功到: ${test_email}`)
    
    return NextResponse.json({
      code: 0,
      message: '测试邮件发送成功',
      data: {
        test_email,
        sent_at: new Date().toISOString(),
        sent_by: user.username
      }
    })
    
  } catch (error) {
    console.error('测试邮件发送失败:', error)
    
    let errorMessage = '邮件配置测试失败'
    if (error instanceof Error) {
      if (error.message.includes('EAUTH')) {
        errorMessage = '邮箱认证失败，请检查用户名和密码'
      } else if (error.message.includes('ECONNECTION')) {
        errorMessage = '无法连接到SMTP服务器，请检查主机和端口'
      } else if (error.message.includes('ETIMEDOUT')) {
        errorMessage = '连接超时，请检查网络和服务器配置'
      } else {
        errorMessage = `邮件发送失败: ${error.message}`
      }
    }
    
    return NextResponse.json({
      code: 500,
      message: errorMessage,
      data: null
    }, { status: 500 })
  }
}

export const POST = withSuperAdminAuth({
  action: 'TEST_EMAIL_CONFIG',
  resource: '/api/admin/settings/test-email'
})(testEmailConfig)