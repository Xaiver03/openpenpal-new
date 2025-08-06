'use client'

export function CommunityStats() {
  return (
    <section className="py-16 bg-gradient-to-br from-amber-50 to-orange-50">
      <div className="container px-4">
        <div className="text-center mb-12">
          <h2 className="font-serif text-3xl font-bold text-amber-900 mb-4">
            社区数据
          </h2>
          <p className="text-amber-700 max-w-2xl mx-auto">
            每一个数字都代表着真实的连接与温暖
          </p>
        </div>

        <div className="grid grid-cols-2 lg:grid-cols-4 gap-8">
          <div className="text-center">
            <div className="text-3xl font-bold text-amber-600 mb-2">1,247</div>
            <div className="text-amber-700 text-sm">发布作品</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-amber-600 mb-2">8,439</div>
            <div className="text-amber-700 text-sm">社区成员</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-amber-600 mb-2">15,672</div>
            <div className="text-amber-700 text-sm">互动评论</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-amber-600 mb-2">23,891</div>
            <div className="text-amber-700 text-sm">点赞数量</div>
          </div>
        </div>
      </div>
    </section>
  )
}