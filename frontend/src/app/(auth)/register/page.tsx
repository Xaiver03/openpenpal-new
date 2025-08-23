'use client'

import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { EmailVerificationStep } from '@/components/auth/EmailVerificationStep'
import { SimpleRegistrationForm } from '@/components/auth/SimpleRegistrationForm'
import { Card, CardContent } from '@/components/ui/card'
import { CheckCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import Link from 'next/link'

type RegistrationStep = 'email-verification' | 'user-info' | 'success'

export default function RegisterPage() {
  const router = useRouter()
  const [mounted, setMounted] = useState(false)
  const [currentStep, setCurrentStep] = useState<RegistrationStep>('email-verification')
  const [email, setEmail] = useState('')
  const [verificationCode, setVerificationCode] = useState('')
  const [, setRegisteredUserId] = useState<string>('')
  
  useEffect(() => {
    setMounted(true)
  }, [])

  const handleEmailVerificationNext = () => {
    setCurrentStep('user-info')
  }

  const handleEmailVerificationBack = () => {
    router.push('/login')
  }

  const handleUserInfoBack = () => {
    setCurrentStep('email-verification')
  }

  const handleRegistrationSuccess = (userId: string) => {
    setRegisteredUserId(userId)
    setCurrentStep('success')
  }

  const handleLoginRedirect = () => {
    router.push('/login')
  }
  
  if (!mounted) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  const renderProgressSteps = () => {
    const steps = [
      { key: 'email-verification', label: '邮箱验证', completed: currentStep !== 'email-verification' },
      { key: 'user-info', label: '完善信息', completed: currentStep === 'success' },
      { key: 'success', label: '注册成功', completed: currentStep === 'success' },
    ]

    return (
      <div className="flex justify-center mb-8">
        <div className="flex items-center space-x-4">
          {steps.map((step, index) => (
            <React.Fragment key={step.key}>
              <div className="flex items-center">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                    step.completed
                      ? 'bg-green-500 text-white'
                      : currentStep === step.key
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-600'
                  }`}
                >
                  {step.completed ? <CheckCircle className="w-4 h-4" /> : index + 1}
                </div>
                <span
                  className={`ml-2 text-sm ${
                    step.completed || currentStep === step.key
                      ? 'text-gray-900'
                      : 'text-gray-500'
                  }`}
                >
                  {step.label}
                </span>
              </div>
              {index < steps.length - 1 && (
                <div
                  className={`w-12 h-0.5 ${
                    steps[index + 1].completed ? 'bg-green-500' : 'bg-gray-200'
                  }`}
                />
              )}
            </React.Fragment>
          ))}
        </div>
      </div>
    )
  }

  const renderCurrentStep = () => {
    switch (currentStep) {
      case 'email-verification':
        return (
          <EmailVerificationStep
            email={email}
            onEmailChange={setEmail}
            verificationCode={verificationCode}
            onVerificationCodeChange={setVerificationCode}
            onNext={handleEmailVerificationNext}
            onBack={handleEmailVerificationBack}
          />
        )

      case 'user-info':
        return (
          <SimpleRegistrationForm
            onSuccess={handleRegistrationSuccess}
            onBack={handleUserInfoBack}
          />
        )

      case 'success':
        return (
          <div className="text-center space-y-6">
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
              <p>我们已向 <strong>{email}</strong> 发送了欢迎邮件，</p>
              <p>请注意查收。</p>
            </div>

            <Button onClick={handleLoginRedirect} className="w-full">
              立即登录
            </Button>
          </div>
        )

      default:
        return null
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-letter-paper via-white to-letter-paper flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="absolute inset-0 bg-[url('/paper-texture.svg')] opacity-10" />
      
      <div className="max-w-md w-full relative">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-serif font-bold text-letter-ink">
            加入OpenPenPal
          </h1>
          <p className="text-gray-600 mt-2">
            创建账户，开始您的数字信件之旅
          </p>
        </div>

        {renderProgressSteps()}

        <Card>
          <CardContent className="p-6">
            {renderCurrentStep()}
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

