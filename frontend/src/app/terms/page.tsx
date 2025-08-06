import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { FileText, ArrowLeft, AlertTriangle, CheckCircle, XCircle, Scale } from 'lucide-react'

export default function TermsPage() {
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
            <FileText className="w-10 h-10 text-amber-700" />
          </div>
          <h1 className="font-serif text-4xl font-bold text-amber-900 mb-4">
            服务条款
          </h1>
          <p className="text-lg text-amber-700">
            使用 OpenPenPal 服务前，请仔细阅读以下条款
          </p>
          <p className="text-sm text-amber-600 mt-2">
            最后更新时间：2024年1月20日
          </p>
        </div>

        <div className="space-y-8">
          {/* 接受条款 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <CheckCircle className="w-5 h-5" />
                条款接受
              </CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>
                欢迎使用 OpenPenPal 信使计划平台（以下简称"本平台"或"我们"）。在使用我们的服务之前，请仔细阅读本服务条款。
              </p>
              <p>
                通过注册账户、访问或使用本平台，您表示同意遵守本服务条款的所有规定。如果您不同意本条款的任何部分，请不要使用我们的服务。
              </p>
            </CardContent>
          </Card>

          {/* 服务描述 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">服务描述</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>OpenPenPal 是一个结合传统手写信和数字技术的校园社交平台，主要功能包括：</p>
              <ul className="list-disc list-inside space-y-2 text-sm">
                <li>数字化信件编辑和生成</li>
                <li>信件编号和二维码生成</li>
                <li>投递状态跟踪</li>
                <li>信使服务系统</li>
                <li>用户社区功能</li>
              </ul>
              <p className="text-sm bg-amber-100 p-3 rounded-md">
                我们保留随时修改、暂停或终止服务的权利，恕不另行通知。
              </p>
            </CardContent>
          </Card>

          {/* 用户责任 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">用户责任与义务</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <div>
                <h3 className="font-semibold mb-2 flex items-center gap-2">
                  <CheckCircle className="w-4 h-4 text-green-600" />
                  您同意：
                </h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>提供真实、准确、完整的注册信息</li>
                  <li>保护账户安全，不与他人共享登录信息</li>
                  <li>遵守相关法律法规和社会道德规范</li>
                  <li>尊重他人权利，不侵犯他人隐私</li>
                  <li>正确使用平台功能，不进行恶意操作</li>
                  <li>及时更新个人信息，确保信息准确性</li>
                </ul>
              </div>

              <div>
                <h3 className="font-semibold mb-2 flex items-center gap-2">
                  <XCircle className="w-4 h-4 text-red-600" />
                  您不得：
                </h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>发布违法、有害、威胁、辱骂、诽谤等内容</li>
                  <li>传播垃圾信息、广告、病毒或恶意代码</li>
                  <li>冒充他人身份或虚假陈述与他人的关系</li>
                  <li>干扰或破坏服务器和网络连接</li>
                  <li>使用自动化程序或机器人访问服务</li>
                  <li>未经授权访问他人账户或系统</li>
                  <li>从事任何违法犯罪活动</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* 内容规范 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">内容与行为规范</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <div>
                <h3 className="font-semibold mb-2">信件内容规范</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>不得包含违法、暴力、色情、恐怖等不当内容</li>
                  <li>不得包含人身攻击、歧视、仇恨言论</li>
                  <li>不得包含商业广告或推广信息</li>
                  <li>不得泄露他人隐私或个人信息</li>
                  <li>提倡积极正面、健康向上的内容</li>
                </ul>
              </div>

              <div>
                <h3 className="font-semibold mb-2">信使行为规范</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>及时、准确地投递信件</li>
                  <li>保护信件的完整性和隐私性</li>
                  <li>诚实上报投递状态</li>
                  <li>遵守投递时间和区域规定</li>
                  <li>维护良好的信使形象和声誉</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* 知识产权 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">知识产权</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <div>
                <h3 className="font-semibold mb-2">平台知识产权</h3>
                <p className="text-sm">
                  本平台的所有内容，包括但不限于软件、技术、程序、网页、文字、图片、音频、视频、图表、版面设计、商标、服务标记等，均受知识产权法保护。
                </p>
              </div>

              <div>
                <h3 className="font-semibold mb-2">用户内容</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>您保留对自己创作内容的知识产权</li>
                  <li>您授权我们使用、存储、传输您的内容以提供服务</li>
                  <li>您保证拥有发布内容的合法权利</li>
                  <li>您承担因内容侵权产生的一切责任</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* 隐私保护 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">隐私保护</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>
                我们非常重视用户隐私保护。详细的隐私处理方式请参阅我们的《隐私政策》。
              </p>
              <ul className="list-disc list-inside space-y-1 text-sm">
                <li>我们会采取合理措施保护您的个人信息</li>
                <li>未经您同意，我们不会向第三方披露您的个人信息</li>
                <li>您有权访问、更正或删除您的个人信息</li>
                <li>我们会及时通知您任何重大隐私政策变更</li>
              </ul>
              <div className="mt-4">
                <Button asChild variant="outline" size="sm" className="border-amber-300 text-amber-700">
                  <Link href="/privacy">
                    查看隐私政策
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* 服务中断 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <AlertTriangle className="w-5 h-5" />
                服务中断与免责
              </CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <div>
                <h3 className="font-semibold mb-2">服务中断</h3>
                <p className="text-sm">
                  在以下情况下，我们可能会中断或终止服务，且不承担任何责任：
                </p>
                <ul className="list-disc list-inside space-y-1 text-sm mt-2">
                  <li>系统维护、升级或故障</li>
                  <li>不可抗力因素（自然灾害、政府行为等）</li>
                  <li>用户违反本条款或相关法律法规</li>
                  <li>第三方服务中断或故障</li>
                  <li>其他超出我们控制范围的情况</li>
                </ul>
              </div>

              <div>
                <h3 className="font-semibold mb-2">免责声明</h3>
                <ul className="list-disc list-inside space-y-1 text-sm">
                  <li>我们不保证服务的绝对连续性和稳定性</li>
                  <li>用户自行承担使用服务的风险</li>
                  <li>我们不对用户之间的纠纷承担责任</li>
                  <li>第三方链接的内容与我们无关</li>
                </ul>
              </div>
            </CardContent>
          </Card>

          {/* 违约责任 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">违约处理</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>如果您违反本条款，我们有权采取以下措施：</p>
              <ul className="list-disc list-inside space-y-2 text-sm">
                <li><strong>警告</strong>：发出违规警告通知</li>
                <li><strong>限制功能</strong>：限制部分或全部功能的使用</li>
                <li><strong>暂停账户</strong>：暂时冻结您的账户</li>
                <li><strong>终止服务</strong>：永久关闭您的账户</li>
                <li><strong>法律追责</strong>：保留通过法律途径维权的权利</li>
              </ul>
              <p className="text-sm bg-red-50 border border-red-200 p-3 rounded-md mt-4">
                严重违规行为可能导致立即终止服务，并承担相应法律责任。
              </p>
            </CardContent>
          </Card>

          {/* 条款变更 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">条款变更</CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>
                我们保留随时修改本服务条款的权利。重大变更时，我们会通过以下方式通知您：
              </p>
              <ul className="list-disc list-inside space-y-1 text-sm">
                <li>在平台显著位置发布公告</li>
                <li>向您的注册邮箱发送通知</li>
                <li>应用内消息推送</li>
              </ul>
              <p className="text-sm mt-4">
                条款修改后，如果您继续使用服务，即视为接受修改后的条款。如不接受，请停止使用服务。
              </p>
            </CardContent>
          </Card>

          {/* 争议解决 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <Scale className="w-5 h-5" />
                争议解决
              </CardTitle>
            </CardHeader>
            <CardContent className="text-amber-800 space-y-4">
              <p>
                本条款的签订、履行、解释及争议解决均适用中华人民共和国法律。
              </p>
              <p>
                因本条款产生的争议，双方应首先友好协商解决。协商不成的，可向我们所在地的人民法院提起诉讼。
              </p>
            </CardContent>
          </Card>

          {/* 联系方式 */}
          <Card className="border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50">
            <CardContent className="p-6 text-center">
              <h3 className="text-xl font-bold text-amber-900 mb-4">联系我们</h3>
              <p className="text-amber-700 mb-4">
                如果您对本服务条款有任何疑问，请联系我们：
              </p>
              <div className="space-y-2 text-sm text-amber-600">
                <p>邮箱：legal@openpenpal.cn</p>
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