'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Loader2, CreditCard, Smartphone, Banknote, AlertCircle, CheckCircle } from 'lucide-react'
import { apiClient } from '@/lib/api-client'
import { toast } from '@/components/ui/use-toast'

export type PaymentMethod = 'alipay' | 'wechat' | 'card' | 'credit'

interface PaymentGatewayProps {
  orderId: string
  amount: number
  onSuccess: (paymentResult: any) => void
  onError: (error: string) => void
  onCancel?: () => void
}

export function PaymentGateway({ orderId, amount, onSuccess, onError, onCancel }: PaymentGatewayProps) {
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethod>('alipay')
  const [processing, setProcessing] = useState(false)
  const [paymentUrl, setPaymentUrl] = useState<string | null>(null)

  const paymentMethods = [
    {
      id: 'alipay' as PaymentMethod,
      name: '支付宝',
      description: '使用支付宝扫码支付',
      icon: Smartphone,
      color: 'text-blue-600',
      available: true
    },
    {
      id: 'wechat' as PaymentMethod,
      name: '微信支付',
      description: '使用微信扫码支付',
      icon: Smartphone,
      color: 'text-green-600',
      available: true
    },
    {
      id: 'card' as PaymentMethod,
      name: '银行卡',
      description: '使用借记卡或信用卡支付',
      icon: CreditCard,
      color: 'text-purple-600',
      available: true
    },
    {
      id: 'credit' as PaymentMethod,
      name: '积分支付',
      description: '使用账户积分支付',
      icon: Banknote,
      color: 'text-orange-600',
      available: true
    }
  ]

  // 发起支付
  const initiatePayment = async () => {
    try {
      setProcessing(true)
      
      const response = await apiClient.post(`/api/v1/shop/orders/${orderId}/pay`, {
        payment_method: selectedMethod,
        return_url: `${window.location.origin}/orders/payment/success?order_id=${orderId}`,
        cancel_url: `${window.location.origin}/orders/${orderId}`
      })

      const paymentData = ((response as any)?.data?.data || (response as any)?.data)?.data

      if (selectedMethod === 'alipay' || selectedMethod === 'wechat') {
        // 扫码支付 - 显示二维码或跳转到支付页面
        if (paymentData.qr_code) {
          setPaymentUrl(paymentData.qr_code)
          // 开始轮询支付状态
          pollPaymentStatus(paymentData.payment_id)
        } else if (paymentData.redirect_url) {
          // 跳转到支付页面
          window.location.href = paymentData.redirect_url
        }
      } else if (selectedMethod === 'card') {
        // 银行卡支付 - 跳转到支付页面
        if (paymentData.redirect_url) {
          window.location.href = paymentData.redirect_url
        }
      } else if (selectedMethod === 'credit') {
        // 积分支付 - 直接扣除积分
        onSuccess({
          payment_id: paymentData.payment_id,
          status: 'success',
          method: selectedMethod
        })
      }
      
    } catch (err: any) {
      setProcessing(false)
      const errorMessage = err.message || '支付发起失败'
      onError(errorMessage)
      toast({
        title: '支付失败',
        description: errorMessage,
        variant: 'destructive'
      })
    }
  }

  // 轮询支付状态（用于扫码支付）
  const pollPaymentStatus = async (paymentId: string) => {
    let attempts = 0
    const maxAttempts = 60 // 5分钟超时

    const poll = async () => {
      try {
        attempts++
        
        const response = await apiClient.get(`/api/v1/shop/payments/${paymentId}/status`)
        const status = ((response as any)?.data?.data || (response as any)?.data)?.data?.status

        if (status === 'paid') {
          setProcessing(false)
          onSuccess({
            payment_id: paymentId,
            status: 'success',
            method: selectedMethod
          })
        } else if (status === 'failed' || status === 'cancelled') {
          setProcessing(false)
          onError('支付失败或已取消')
        } else if (attempts < maxAttempts) {
          // 继续轮询
          setTimeout(poll, 5000) // 5秒间隔
        } else {
          // 超时
          setProcessing(false)
          onError('支付超时，请重试')
        }
      } catch (err) {
        if (attempts < maxAttempts) {
          setTimeout(poll, 5000)
        } else {
          setProcessing(false)
          onError('支付状态查询失败')
        }
      }
    }

    poll()
  }

  // 取消支付
  const cancelPayment = () => {
    setProcessing(false)
    setPaymentUrl(null)
    if (onCancel) {
      onCancel()
    }
  }

  return (
    <div className="space-y-6">
      {/* 支付金额显示 */}
      <Card>
        <CardContent className="p-6">
          <div className="text-center">
            <p className="text-sm text-gray-600 mb-2">支付金额</p>
            <p className="text-3xl font-bold text-red-600">¥{amount.toFixed(2)}</p>
          </div>
        </CardContent>
      </Card>

      {/* 支付方式选择 */}
      {!paymentUrl && (
        <Card>
          <CardHeader>
            <CardTitle>选择支付方式</CardTitle>
            <CardDescription>请选择您偏好的支付方式</CardDescription>
          </CardHeader>
          <CardContent>
            <RadioGroup 
              value={selectedMethod} 
              onValueChange={(value) => setSelectedMethod(value as PaymentMethod)}
              className="space-y-3"
            >
              {paymentMethods.map((method) => (
                <div key={method.id} className="flex items-center space-x-3">
                  <RadioGroupItem 
                    value={method.id} 
                    id={method.id}
                    disabled={!method.available}
                  />
                  <Label 
                    htmlFor={method.id} 
                    className={`flex-1 cursor-pointer ${!method.available ? 'opacity-50' : ''}`}
                  >
                    <div className="flex items-center gap-3 p-4 border rounded-lg hover:bg-gray-50 transition-colors">
                      <method.icon className={`h-6 w-6 ${method.color}`} />
                      <div>
                        <p className="font-medium">{method.name}</p>
                        <p className="text-sm text-gray-600">{method.description}</p>
                      </div>
                      {!method.available && (
                        <span className="text-xs text-red-500">暂不可用</span>
                      )}
                    </div>
                  </Label>
                </div>
              ))}
            </RadioGroup>
          </CardContent>
        </Card>
      )}

      {/* 扫码支付区域 */}
      {paymentUrl && (
        <Card>
          <CardHeader>
            <CardTitle className="text-center">扫码支付</CardTitle>
            <CardDescription className="text-center">
              请使用{selectedMethod === 'alipay' ? '支付宝' : '微信'}扫描下方二维码完成支付
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center space-y-4">
            <div className="inline-block p-4 bg-white border rounded-lg">
              <img 
                src={paymentUrl} 
                alt="支付二维码" 
                className="w-48 h-48 mx-auto"
              />
            </div>
            <p className="text-sm text-gray-600">
              支付后页面将自动跳转，请不要关闭此页面
            </p>
            <div className="flex items-center justify-center gap-2 text-blue-600">
              <Loader2 className="w-4 h-4 animate-spin" />
              <span className="text-sm">等待支付完成...</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 支付提示 */}
      {processing && !paymentUrl && (
        <Alert>
          <Loader2 className="h-4 w-4 animate-spin" />
          <AlertDescription>
            正在处理支付请求，请稍候...
          </AlertDescription>
        </Alert>
      )}

      {/* 操作按钮 */}
      <div className="flex gap-4">
        {!paymentUrl ? (
          <>
            <Button 
              onClick={initiatePayment}
              disabled={processing}
              className="flex-1"
            >
              {processing ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  处理中...
                </>
              ) : (
                <>
                  <CreditCard className="w-4 h-4 mr-2" />
                  确认支付
                </>
              )}
            </Button>
            {onCancel && (
              <Button variant="outline" onClick={onCancel} disabled={processing}>
                取消
              </Button>
            )}
          </>
        ) : (
          <>
            <Button variant="outline" onClick={cancelPayment} className="flex-1">
              取消支付
            </Button>
            <Button 
              onClick={() => window.location.reload()} 
              variant="outline"
            >
              刷新状态
            </Button>
          </>
        )}
      </div>

      {/* 安全提示 */}
      <Alert>
        <CheckCircle className="h-4 w-4" />
        <AlertDescription>
          <div className="space-y-1 text-sm">
            <p>• 支付过程使用SSL加密，保障您的资金安全</p>
            <p>• 支付遇到问题请联系客服：service@openpenpal.com</p>
            <p>• 请在30分钟内完成支付，否则订单将自动取消</p>
          </div>
        </AlertDescription>
      </Alert>
    </div>
  )
}

export default PaymentGateway