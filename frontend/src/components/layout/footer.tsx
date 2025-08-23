import Link from 'next/link'
import { Mail, Heart, Github, Twitter } from 'lucide-react'
import { cn } from '@/lib/utils'

interface FooterProps {
  className?: string
}

export function Footer({ className }: FooterProps) {
  const currentYear = new Date().getFullYear()

  const footerLinks = {
    product: [
      { label: '写信', href: '/letters/write' },
      { label: '投递', href: '/letters/send' },
      { label: '信件', href: '/letters' },
      { label: '信使', href: '/courier/dashboard' },
    ],
    support: [
      { label: '帮助中心', href: '/help' },
      { label: '使用指南', href: '/guide' },
      { label: '常见问题', href: '/faq' },
      { label: '联系我们', href: '/contact' },
    ],
    company: [
      { label: '关于我们', href: '/about' },
      { label: '隐私政策', href: '/privacy' },
      { label: '服务条款', href: '/terms' },
      { label: '加入我们', href: '/careers' },
    ],
  }

  const socialLinks = [
    { label: 'GitHub', href: 'https://github.com/openpenpal', icon: Github },
    { label: 'Twitter', href: 'https://twitter.com/openpenpal', icon: Twitter },
  ]

  return (
    <footer className={cn(
      'border-t bg-background mt-auto',
      className
    )}>
      <div className="container px-4 py-12">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-5">
          {/* Brand */}
          <div className="lg:col-span-2">
            <Link href="/" className="flex items-center space-x-2 mb-4">
              <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
                <Mail className="h-5 w-5" />
              </div>
              <span className="font-serif text-xl font-bold text-letter-ink">
                OpenPenPal
              </span>
            </Link>
            <p className="text-muted-foreground text-sm leading-relaxed max-w-sm">
              实体手写信 + 数字跟踪平台，重建校园社群的温度感知与精神连接。
              让每一封信都成为连接心灵的桥梁。
            </p>
            <div className="flex items-center space-x-4 mt-6">
              {socialLinks.map((social) => {
                const Icon = social.icon
                return (
                  <Link
                    key={social.label}
                    href={social.href}
                    className="text-muted-foreground hover:text-foreground transition-colors"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Icon className="h-5 w-5" />
                    <span className="sr-only">{social.label}</span>
                  </Link>
                )
              })}
            </div>
          </div>

          {/* Product Links */}
          <div>
            <h3 className="font-semibold text-foreground mb-4">产品</h3>
            <ul className="space-y-3">
              {footerLinks.product.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Support Links */}
          <div>
            <h3 className="font-semibold text-foreground mb-4">支持</h3>
            <ul className="space-y-3">
              {footerLinks.support.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Company Links */}
          <div>
            <h3 className="font-semibold text-foreground mb-4">公司</h3>
            <ul className="space-y-3">
              {footerLinks.company.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Bottom */}
        <div className="border-t mt-12 pt-8 flex flex-col md:flex-row items-center justify-between">
          <p className="text-sm text-muted-foreground">
            © {currentYear} OpenPenPal. 保留所有权利。
          </p>
          <p className="text-sm text-muted-foreground mt-4 md:mt-0 flex items-center">
            用 <Heart className="h-4 w-4 mx-1 text-red-500" /> 为校园社交而构建
          </p>
        </div>
      </div>
    </footer>
  )
}