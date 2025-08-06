'use client'

import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Card, CardContent } from '@/components/ui/card'
import { CheckCircle } from 'lucide-react'
import Link from 'next/link'

export default function SimpleRegisterPage() {
  const router = useRouter()
  const [mounted, setMounted] = useState(false)
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    schoolCode: '',
    realName: '',
    agreeToTerms: false,
    agreeToPrivacy: false,
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    setError('') // 清除错误信息
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // 基本验证
    if (!formData.username || !formData.email || !formData.password || 
        !formData.confirmPassword || !formData.schoolCode || !formData.realName) {
      setError('请填写所有必填字段')
      return
    }

    if (formData.password !== formData.confirmPassword) {
      setError('两次输入的密码不一致')
      return
    }

    if (!formData.agreeToTerms || !formData.agreeToPrivacy) {
      setError('请同意用户协议和隐私政策')
      return
    }

    setLoading(true)
    
    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: formData.username,
          email: formData.email,
          password: formData.password,
          confirmPassword: formData.confirmPassword,
          schoolCode: formData.schoolCode,
          realName: formData.realName,
          verificationCode: 'SKIP_VERIFICATION',
          agreeToTerms: formData.agreeToTerms,
          agreeToPrivacy: formData.agreeToPrivacy
        }),
      })

      const result = await response.json()

      if (result.code === 0) {
        setSuccess(true)
        console.log('用户注册成功:', result.data)
      } else {
        setError(result.message || '注册失败，请稍后重试')
      }
    } catch (err) {
      console.error('注册错误:', err)
      setError('网络错误，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  if (!mounted) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  if (success) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-letter-paper via-white to-letter-paper flex items-center justify-center py-12 px-4">
        <div className="max-w-md w-full">
          <Card>
            <CardContent className="p-8 text-center space-y-6">
              <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center">
                <CheckCircle className="w-8 h-8 text-green-500" />
              </div>
              
              <div>
                <h2 className="text-2xl font-bold text-gray-900">注册成功！</h2>
                <p className="text-gray-600 mt-2">
                  欢迎加入OpenPenPal！您的账户已创建成功。
                </p>
              </div>

              <div className="bg-gray-50 rounded-lg p-4 text-left">
                <h3 className="font-medium mb-2">接下来您可以：</h3>
                <ul className="text-sm text-gray-600 space-y-1">
                  <li>• 立即登录开始使用OpenPenPal</li>
                  <li>• 创建您的第一封电子信件</li>
                  <li>• 参与博物馆投稿活动</li>
                  <li>• 申请成为信使，参与信件投递</li>
                  <li>• 浏览精美的信件展览</li>
                </ul>
              </div>

              <div className="text-sm text-gray-500">
                <p>我们已向 <strong>{formData.email}</strong> 发送了欢迎邮件（开发环境下为模拟发送）</p>
              </div>

              <Button onClick={() => router.push('/login')} className="w-full">
                立即登录
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-letter-paper via-white to-letter-paper flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-serif font-bold text-letter-ink">
            快速注册 OpenPenPal
          </h1>
          <p className="text-gray-600 mt-2">
            创建账户，开始您的数字信件之旅
          </p>
        </div>

        <Card>
          <CardContent className="p-6">
            {error && (
              <Alert variant="destructive" className="mb-6">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <Label htmlFor="username">用户名 *</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => handleInputChange('username', e.target.value)}
                  placeholder="请输入用户名"
                  required
                />
              </div>

              <div>
                <Label htmlFor="email">邮箱 *</Label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) => handleInputChange('email', e.target.value)}
                  placeholder="请输入邮箱"
                  required
                />
              </div>

              <div>
                <Label htmlFor="password">密码 *</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => handleInputChange('password', e.target.value)}
                  placeholder="请输入密码"
                  required
                />
              </div>

              <div>
                <Label htmlFor="confirmPassword">确认密码 *</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  value={formData.confirmPassword}
                  onChange={(e) => handleInputChange('confirmPassword', e.target.value)}
                  placeholder="请再次输入密码"
                  required
                />
              </div>

              <div>
                <Label htmlFor="realName">真实姓名 *</Label>
                <Input
                  id="realName"
                  value={formData.realName}
                  onChange={(e) => handleInputChange('realName', e.target.value)}
                  placeholder="请输入真实姓名"
                  required
                />
              </div>

              <div>
                <Label htmlFor="schoolCode">学校编码 *</Label>
                <Input
                  id="schoolCode"
                  value={formData.schoolCode}
                  onChange={(e) => handleInputChange('schoolCode', e.target.value.toUpperCase())}
                  placeholder="请输入学校编码，如：BJUT2024"
                  required
                />
                <p className="text-xs text-gray-500 mt-1">
                  可用的测试编码：BJUT2024, THU2024, PKU2024 等
                </p>
              </div>

              <div className="space-y-3">
                <div className="flex items-start space-x-2">
                  <Checkbox
                    id="agreeToTerms"
                    checked={formData.agreeToTerms}
                    onCheckedChange={(checked) => handleInputChange('agreeToTerms', checked)}
                  />
                  <Label htmlFor="agreeToTerms" className="text-sm">
                    我已阅读并同意 <a href="/terms" target="_blank" className="text-blue-600 hover:underline">《用户协议》</a>
                  </Label>
                </div>
                
                <div className="flex items-start space-x-2">
                  <Checkbox
                    id="agreeToPrivacy"
                    checked={formData.agreeToPrivacy}
                    onCheckedChange={(checked) => handleInputChange('agreeToPrivacy', checked)}
                  />
                  <Label htmlFor="agreeToPrivacy" className="text-sm">
                    我已阅读并同意 <a href="/privacy" target="_blank" className="text-blue-600 hover:underline">《隐私政策》</a>
                  </Label>
                </div>
              </div>

              <Button type="submit" disabled={loading} className="w-full">
                {loading ? '注册中...' : '立即注册'}
              </Button>
            </form>
          </CardContent>
        </Card>

        <div className="text-center mt-6">
          <p className="text-sm text-gray-600">
            已有账户？{' '}
            <Link
              href="/login"
              className="text-letter-accent hover:text-letter-accent/80 font-medium"
            >
              立即登录
            </Link>
          </p>
        </div>
        
        <div className="text-center mt-4">
          <Link 
            href="/" 
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            ← 返回首页
          </Link>
        </div>
      </div>
    </div>
  )
}