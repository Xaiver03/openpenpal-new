'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Label } from '@/components/ui/label'
import { 
  Mail, 
  Lock,
  AlertCircle,
  User,
  Eye,
  EyeOff
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
export default function LoginPage() {
  const [mounted, setMounted] = useState(false)
  
  useEffect(() => {
    setMounted(true)
  }, [])
  
  if (!mounted) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }
  
  return <LoginForm />
}

function LoginForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { login } = useAuth()
  
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  })
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showPassword, setShowPassword] = useState(false)
  
  // 自动聚焦到用户名输入框
  useEffect(() => {
    const usernameInput = document.getElementById('username')
    if (usernameInput) {
      usernameInput.focus()
    }
  }, [])
  
  // 切换密码显示状态
  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    try {
      console.log('🔄 开始登录:', formData.username)
      await login(formData)
      console.log('✅ 登录成功，准备跳转')
      
      // 检查是否有返回URL参数
      const returnUrl = searchParams.get('returnUrl')
      
      // 登录成功，跳转到首页或返回URL
      const targetUrl = returnUrl || '/'
      console.log('🚀 跳转到:', targetUrl)
      router.push(targetUrl)
    } catch (error) {
      setError(error instanceof Error ? error.message : '登录失败，请稍后重试')
    } finally {
      setIsLoading(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({
      ...prev,
      [e.target.name]: e.target.value
    }))
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-letter-paper via-white to-letter-paper">
      <div className="absolute inset-0 bg-[url('/paper-texture.svg')] opacity-10" />
      
      <Card className="w-full max-w-md mx-4 relative">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-serif text-letter-ink">
            登录 OpenPenPal
          </CardTitle>
          <CardDescription>
            欢迎回来，继续你的信件之旅
          </CardDescription>
        </CardHeader>
        
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div className="space-y-2">
              <Label htmlFor="username">用户名或邮箱</Label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="username"
                  name="username"
                  type="text"
                  placeholder="请输入用户名或邮箱"
                  value={formData.username}
                  onChange={handleChange}
                  className="pl-10 focus:caret-letter-amber"
                  required
                  autoComplete="username"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="password"
                  name="password"
                  type={showPassword ? "text" : "password"}
                  placeholder="请输入密码"
                  value={formData.password}
                  onChange={handleChange}
                  className="pl-10 pr-10 focus:caret-letter-amber"
                  required
                  autoComplete="current-password"
                />
                <button
                  type="button"
                  onClick={togglePasswordVisibility}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  aria-label={showPassword ? "隐藏密码" : "显示密码"}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

            <Button 
              type="submit" 
              className="w-full font-serif"
              disabled={isLoading}
            >
              {isLoading ? (
                <>
                  <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                  登录中...
                </>
              ) : (
                <>
                  <Mail className="mr-2 h-4 w-4" />
                  登录
                </>
              )}
            </Button>
          </form>

          <div className="mt-6 text-center text-sm">
            <span className="text-muted-foreground">还没有账户？</span>
            <Link 
              href="/register" 
              className="font-medium text-letter-accent hover:text-letter-accent/80 ml-1"
            >
              立即注册
            </Link>
          </div>

          <div className="mt-4 text-center">
            <Link 
              href="/" 
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              ← 返回首页
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}