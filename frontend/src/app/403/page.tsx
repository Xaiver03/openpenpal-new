import Link from 'next/link'
import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { BackButton } from '@/components/ui/back-button'

export default function Forbidden() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-letter-paper">
      <div className="text-center">
        <AlertTriangle className="mx-auto h-16 w-16 text-amber-600" />
        <h1 className="mt-4 text-3xl font-bold text-gray-900">权限不足</h1>
        <p className="mt-2 text-gray-600">抱歉，您没有权限访问此页面</p>
        <div className="mt-6 flex gap-4 justify-center">
          <BackButton>返回上页</BackButton>
          <Link href="/">
            <Button>返回首页</Button>
          </Link>
          <Link href="/login">
            <Button variant="outline">重新登录</Button>
          </Link>
        </div>
      </div>
    </div>
  )
}