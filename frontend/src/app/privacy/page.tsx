import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Shield, ArrowLeft, Eye, Lock, UserCheck, Database } from 'lucide-react'

export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-4xl mx-auto px-4 py-8">
        {/* 返回按钮 */}
        <div className="mb-8">
          <Button asChild variant="outline" size="sm" className="border-amber-300 text-amber-700 hover:bg-amber-50">
            <Link href="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              返回首页
            </Link>
          </Button>
        </div>

        {/* 页面标题 */}
        <div className="text-center mb-12">
          <div className="w-20 h-20 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-6">
            <Shield className="w-10 h-10 text-amber-700" />
          </div>
          <h1 className="font-serif text-4xl font-bold text-amber-900 mb-4">
            隐私政策
          </h1>
          <p className="text-lg text-amber-700">
            我们致力于保护您的隐私和数据安全
          </p>
          <p className="text-sm text-amber-600 mt-2">
            最后更新时间：2024年1月20日
          </p>
        </div>

        <div className="space-y-8">
          {/* 概述 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <Eye className="w-5 h-5" />
                隐私保护概述
              </CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>
                OpenPenPal（以下简称"我们"）非常重视用户隐私保护。本隐私政策详细说明了我们如何收集、使用、存储和保护您的个人信息。
              </p>
              <p>
                使用我们的服务即表示您同意本隐私政策的条款。如果您不同意本政策，请停止使用我们的服务。
              </p>
            </CardContent>
          </Card>

          {/* 信息收集 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <Database className="w-5 h-5" />
                我们收集的信息
              </CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <div>
                <h3 className="font-semibold mb-2">个人信息</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>注册信息：用户名、邮箱、密码、昵称</li>
                  <li>身份信息：学校代码（用于身份验证）</li>
                  <li>联系信息：手机号码（可选）</li>
                  <li>个人资料：头像、个人简介（可选）</li>
                </ul>
              </div>
              
              <div>
                <h3 className="font-semibold mb-2">使用数据</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>信件内容和状态信息</li>
                  <li>投递记录和统计数据</li>
                  <li>登录日志和操作记录</li>
                  <li>设备信息和IP地址</li>
                </ul>
              </div>

              <div>
                <h3 className="font-semibold mb-2">自动收集的信息</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>浏览器类型和版本</li>
                  <li>操作系统信息</li>
                  <li>访问时间和页面</li>
                  <li>Cookie和本地存储数据</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* 信息使用 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <UserCheck className="w-5 h-5" />
                信息使用方式
              </CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>我们收集的信息仅用于以下目的：</p>
              <ul className="list-disc list-inside space-y-2 text-sm">
                <li><strong>提供服务</strong>：处理注册、登录、信件创建和投递等核心功能</li>
                <li><strong>身份验证</strong>：确保用户身份真实性，维护平台安全</li>
                <li><strong>个性化体验</strong>：根据用户偏好优化界面和功能</li>
                <li><strong>客户支持</strong>：响应用户咨询和技术问题</li>
                <li><strong>平台改进</strong>：分析使用数据以改善服务质量</li>
                <li><strong>安全防护</strong>：检测和预防欺诈、滥用等行为</li>
                <li><strong>法律合规</strong>：履行法律义务和保护合法权益</li>
              </ul>
            </CardContent>
          </Card>

          {/* 信息保护 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <Lock className="w-5 h-5" />
                数据安全保护
              </CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <div>
                <h3 className="font-semibold mb-2">技术保护措施</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>HTTPS加密传输，保护数据传输安全</li>
                  <li>密码加密存储，使用bcrypt算法</li>
                  <li>JWT令牌认证，确保访问权限</li>
                  <li>数据库访问控制和权限管理</li>
                  <li>定期安全审计和漏洞修复</li>
                </ul>
              </div>

              <div>
                <h3 className="font-semibold mb-2">管理保护措施</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>员工隐私培训和保密协议</li>
                  <li>最小权限原则，限制数据访问</li>
                  <li>数据备份和灾难恢复机制</li>
                  <li>安全事件响应和处理流程</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* 信息共享 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">信息共享和披露</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>我们承诺不会向第三方出售、交易或转让您的个人信息。但在以下情况下，我们可能会共享必要的信息：</p>
              <ul className="list-disc list-inside space-y-2 text-sm">
                <li><strong>获得明确同意</strong>：在获得您明确同意的情况下</li>
                <li><strong>法律要求</strong>：根据法律法规、法院命令或政府要求</li>
                <li><strong>安全保护</strong>：为保护用户、公众或我们的权利和安全</li>
                <li><strong>服务提供商</strong>：与我们的技术服务提供商（仅限于提供服务所需）</li>
                <li><strong>业务转移</strong>：在合并、收购或资产转移时（用户将被提前通知）</li>
              </ul>
            </CardContent>
          </Card>

          {/* 用户权利 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">您的权利</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>您对自己的个人信息享有以下权利：</p>
              <ul className="list-disc list-inside space-y-2 text-sm">
                <li><strong>访问权</strong>：查看我们持有的您的个人信息</li>
                <li><strong>更正权</strong>：更新或修正不准确的个人信息</li>
                <li><strong>删除权</strong>：要求删除您的个人信息</li>
                <li><strong>限制处理权</strong>：限制我们处理您的个人信息</li>
                <li><strong>数据可携权</strong>：获得您的个人信息副本</li>
                <li><strong>撤回同意权</strong>：撤回之前给予的同意</li>
              </ul>
              <p className="mt-4 text-sm bg-amber-100 p-3 rounded-md">
                如需行使上述权利，请通过邮箱 privacy@openpenpal.cn 联系我们。
              </p>
            </CardContent>
          </Card>

          {/* Cookie政策 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">Cookie 和本地存储</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>我们使用Cookie和本地存储技术来改善用户体验：</p>
              <ul className="list-disc list-inside space-y-2 text-sm">
                <li><strong>必要Cookie</strong>：维持登录状态、保存用户偏好</li>
                <li><strong>功能Cookie</strong>：记住您的设置和选择</li>
                <li><strong>分析Cookie</strong>：了解网站使用情况（匿名数据）</li>
              </ul>
              <p className="text-sm mt-4">
                您可以通过浏览器设置管理Cookie，但这可能影响某些功能的正常使用。
              </p>
            </CardContent>
          </Card>

          {/* 未成年人保护 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">未成年人保护</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>
                我们非常重视未成年人的隐私保护。如果您未满18周岁，请在监护人的陪同下阅读本政策，并在获得监护人同意后使用我们的服务。
              </p>
              <p>
                如果我们发现在未获得监护人同意的情况下收集了未成年人的个人信息，会立即删除相关信息。
              </p>
            </CardContent>
          </Card>

          {/* 政策更新 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">政策更新</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>
                我们可能会不定期更新本隐私政策。重大变更时，我们会通过网站公告、邮件或其他方式通知您。
              </p>
              <p>
                建议您定期查看本政策，以了解我们如何保护您的信息。继续使用服务即表示您接受更新后的政策。
              </p>
            </CardContent>
          </Card>

          {/* 联系我们 */}
          <Card className="border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50">
            <CardContent className="p-6 text-center">
              <h3 className="text-xl font-bold text-amber-900 mb-4">联系我们</h3>
              <p className="text-amber-700 mb-4">
                如果您对本隐私政策有任何疑问或建议，请联系我们：
              </p>
              <div className="space-y-2 text-sm text-amber-600">
                <p>邮箱：privacy@openpenpal.cn</p>
                <p>电话：400-123-4567</p>
                <p>地址：北京市海淀区中关村大街1号</p>
              </div>
              <div className="mt-6">
                <Button asChild className="bg-amber-600 hover:bg-amber-700 text-white">
                  <Link href="/contact">
                    联系我们
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}