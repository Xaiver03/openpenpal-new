'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Send, Globe, Star, Heart, Users } from 'lucide-react'
import { useAuth, useCourier } from '@/stores/user-store'

export function JoinUsSection() {
  const [mounted, setMounted] = useState(false)
  const { isAuthenticated } = useAuth()
  const { isCourier } = useCourier()
  
  useEffect(() => {
    setMounted(true)
  }, [])
  
  // Default content for server-side rendering
  const isAuthenticatedCourier = mounted ? (isAuthenticated && isCourier) : false
  const showCourierCTA = mounted ? (!isAuthenticated || !isCourier) : true
  
  return (
    <section className="py-20 bg-gradient-to-br from-amber-100 to-orange-100">
      <div className="container px-4">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="font-serif text-3xl md:text-4xl font-bold text-amber-900 mb-6">
            {isAuthenticatedCourier ? '信使成长之路' : '成为连接世界的信使'}
          </h2>
          <p className="text-xl text-amber-700 mb-12 max-w-2xl mx-auto">
            {isAuthenticatedCourier 
              ? '继续您的信使旅程，帮助更多人传递温暖'
              : '加入我们的信使网络，成为传递温暖的使者，在帮助他人的同时收获成长与友谊'
            }
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
            <div className="text-center">
              <div className="w-16 h-16 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-4">
                <Star className="w-8 h-8 text-amber-700" />
              </div>
              <h3 className="font-semibold text-amber-900 mb-2">成长体系</h3>
              <p className="text-amber-700 text-sm">从新手信使到资深导师，见证自己的成长</p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-orange-200 rounded-full flex items-center justify-center mx-auto mb-4">
                <Heart className="w-8 h-8 text-orange-700" />
              </div>
              <h3 className="font-semibold text-amber-900 mb-2">温暖奖励</h3>
              <p className="text-amber-700 text-sm">每一次投递都有意义，收获感谢与友谊</p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-yellow-200 rounded-full flex items-center justify-center mx-auto mb-4">
                <Users className="w-8 h-8 text-yellow-700" />
              </div>
              <h3 className="font-semibold text-amber-900 mb-2">社区归属</h3>
              <p className="text-amber-700 text-sm">加入温暖的信使大家庭，结识志同道合的朋友</p>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            {showCourierCTA ? (
              <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif px-8">
                <Link href="/courier">
                  <Send className="mr-2 h-5 w-5" />
                  申请成为信使
                </Link>
              </Button>
            ) : (
              <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif px-8">
                <Link href="/courier">
                  <Send className="mr-2 h-5 w-5" />
                  继续信使之路
                </Link>
              </Button>
            )}
            <Button asChild variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50 font-serif px-8">
              <Link href="/about">
                <Globe className="mr-2 h-5 w-5" />
                了解合作方式
              </Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  )
}