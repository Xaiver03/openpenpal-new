import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import Link from 'next/link'
import { Home, ArrowLeft, Mail, Search } from 'lucide-react'

export default function NotFound() {
  const quickLinks = [
    { href: '/write', label: '写信', icon: Mail },
    { href: '/mailbox', label: '我的信箱', icon: Mail },
    { href: '/deliver', label: '投递信件', icon: Mail },
    { href: '/about', label: '关于我们', icon: Search },
  ]

  return (
    <div className="min-h-screen bg-amber-50 flex items-center justify-center">
      <div className="container max-w-2xl mx-auto px-4 text-center">
        <Card className="border-amber-200 bg-white shadow-lg">
          <CardHeader className="pb-8">
            <div className="mx-auto w-24 h-24 bg-amber-100 rounded-full flex items-center justify-center mb-6">
              <Search className="w-12 h-12 text-amber-600" />
            </div>
            <CardTitle className="font-serif text-3xl font-bold text-amber-900">
              页面未找到
            </CardTitle>
            <CardDescription className="text-lg text-amber-700">
              抱歉，您访问的页面不存在或正在开发中
            </CardDescription>
          </CardHeader>
          
          <CardContent className="space-y-6">
            <div className="p-4 bg-amber-50 rounded-lg border border-amber-200">
              <p className="text-amber-800">
                您可能是通过旧链接访问的，或者该功能还在开发中。
                <br />
                请尝试访问以下页面：
              </p>
            </div>

            {/* Quick Links */}
            <div className="grid grid-cols-2 gap-3">
              {quickLinks.map((link) => {
                const Icon = link.icon
                return (
                  <Button
                    key={link.href}
                    asChild
                    variant="outline"
                    className="border-amber-300 text-amber-700 hover:bg-amber-50 h-auto py-3"
                  >
                    <Link href={link.href} className="flex flex-col items-center gap-2">
                      <Icon className="w-5 h-5" />
                      <span className="text-sm">{link.label}</span>
                    </Link>
                  </Button>
                )
              })}
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 justify-center pt-4">
              <Button asChild className="bg-amber-600 hover:bg-amber-700 text-white">
                <Link href="/">
                  <Home className="mr-2 h-4 w-4" />
                  返回首页
                </Link>
              </Button>
              <Button asChild variant="outline" className="border-amber-300 text-amber-700 hover:bg-amber-50">
                <Link href="/write">
                  <Mail className="mr-2 h-4 w-4" />
                  开始写信
                </Link>
              </Button>
            </div>

            <div className="text-sm text-amber-600 pt-4 border-t border-amber-200">
              <p>如果您认为这是一个错误，请联系我们的支持团队</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}