import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useRegistration } from '@/hooks/useRegistration';

interface EmailVerificationStepProps {
  email: string;
  onEmailChange: (email: string) => void;
  verificationCode: string;
  onVerificationCodeChange: (code: string) => void;
  onNext: () => void;
  onBack: () => void;
}

export const EmailVerificationStep: React.FC<EmailVerificationStepProps> = ({
  email,
  onEmailChange,
  verificationCode,
  onVerificationCodeChange,
  onNext,
  onBack,
}) => {
  const {
    loading,
    error,
    sendVerificationCode,
    resendVerificationCode,
    verifyEmailCode,
    checkEmailAvailability,
    getVerificationStatus,
  } = useRegistration();

  const [emailError, setEmailError] = useState<string>('');
  const [codeError, setCodeError] = useState<string>('');
  const [codeSent, setCodeSent] = useState(false);
  const [cooldown, setCooldown] = useState(0);
  const [emailValidated, setEmailValidated] = useState(false);
  const [codeValidated, setCodeValidated] = useState(false);

  // 倒计时效果
  useEffect(() => {
    if (cooldown > 0) {
      const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [cooldown]);

  // 验证邮箱格式
  const isValidEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  // 检查邮箱可用性
  const handleEmailBlur = async () => {
    if (!email) {
      setEmailError('请输入邮箱地址');
      return;
    }

    if (!isValidEmail(email)) {
      setEmailError('邮箱格式不正确');
      return;
    }

    try {
      const response = await checkEmailAvailability(email);
      if (!response.available) {
        setEmailError('该邮箱已被注册');
        setEmailValidated(false);
      } else {
        setEmailError('');
        setEmailValidated(true);
      }
    } catch (err) {
      setEmailError('邮箱验证失败，请重试');
      setEmailValidated(false);
    }
  };

  // 发送验证码
  const handleSendCode = async () => {
    // 基本邮箱格式验证
    if (!email || !isValidEmail(email)) {
      setEmailError('请输入正确的邮箱地址');
      return;
    }

    console.log('🚀 准备发送验证码到:', email);

    try {
      // 先检查邮箱可用性
      if (!emailValidated) {
        const availabilityResponse = await checkEmailAvailability(email);
        if (!availabilityResponse.available) {
          setEmailError('该邮箱已被注册');
          return;
        }
        setEmailValidated(true);
        setEmailError('');
      }

      // 发送验证码
      console.log('📧 调用发送验证码API...');
      const response = await sendVerificationCode({ email });
      console.log('📧 验证码发送响应:', response);
      
      setCodeSent(true);
      setCooldown(response.cooldownSeconds);
      setCodeError('');
    } catch (err) {
      console.error('❌ 验证码发送错误:', err);
      setCodeError('验证码发送失败，请重试');
    }
  };

  // 重新发送验证码
  const handleResendCode = async () => {
    try {
      // 先检查冷却状态
      const status = await getVerificationStatus(email);
      if (!status.canSend) {
        setCooldown(status.cooldownSeconds);
        setCodeError(`请等待 ${status.cooldownSeconds} 秒后再重新发送`);
        return;
      }

      const response = await resendVerificationCode(email);
      setCooldown(response.cooldownSeconds);
      setCodeError('');
    } catch (err) {
      setCodeError('验证码发送失败，请重试');
    }
  };

  // 验证验证码
  const handleVerifyCode = async () => {
    if (!verificationCode || verificationCode.length !== 6) {
      setCodeError('请输入6位验证码');
      return;
    }

    try {
      const response = await verifyEmailCode(email, verificationCode);
      if (response.isValid) {
        setCodeError('');
        setCodeValidated(true);
      } else {
        setCodeError('验证码无效或已过期');
        setCodeValidated(false);
      }
    } catch (err) {
      setCodeError('验证码验证失败');
      setCodeValidated(false);
    }
  };

  // 继续下一步
  const handleNext = () => {
    if (!emailValidated) {
      setEmailError('请先验证邮箱');
      return;
    }
    if (!codeValidated) {
      setCodeError('请先验证邮箱验证码');
      return;
    }
    onNext();
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">邮箱验证</h2>
        <p className="text-gray-600 mt-2">我们将向您的邮箱发送验证码以确认身份</p>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-4">
        {/* 邮箱输入 */}
        <div>
          <Label htmlFor="email">邮箱地址</Label>
          <div className="flex gap-2 mt-1">
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => onEmailChange(e.target.value)}
              onBlur={handleEmailBlur}
              placeholder="请输入您的邮箱地址"
              className={emailError ? 'border-red-500' : emailValidated ? 'border-green-500' : ''}
            />
            <Button
              type="button"
              onClick={handleSendCode}
              disabled={loading || !emailValidated || cooldown > 0}
              className="whitespace-nowrap"
            >
              {cooldown > 0 ? `${cooldown}s` : codeSent ? '重新发送' : '发送验证码'}
            </Button>
          </div>
          {emailError && <p className="text-red-500 text-sm mt-1">{emailError}</p>}
          {emailValidated && !emailError && (
            <p className="text-green-500 text-sm mt-1">✓ 邮箱可用</p>
          )}
        </div>

        {/* 验证码输入 */}
        {codeSent && (
          <div>
            <Label htmlFor="verificationCode">验证码</Label>
            <div className="flex gap-2 mt-1">
              <Input
                id="verificationCode"
                type="text"
                value={verificationCode}
                onChange={(e) => {
                  const value = e.target.value.replace(/\D/g, '').slice(0, 6);
                  onVerificationCodeChange(value);
                }}
                placeholder="请输入6位验证码"
                maxLength={6}
                className={codeError ? 'border-red-500' : codeValidated ? 'border-green-500' : ''}
              />
              <Button
                type="button"
                onClick={handleVerifyCode}
                disabled={loading || verificationCode.length !== 6}
                className="whitespace-nowrap"
              >
                验证
              </Button>
            </div>
            {codeError && <p className="text-red-500 text-sm mt-1">{codeError}</p>}
            {codeValidated && !codeError && (
              <p className="text-green-500 text-sm mt-1">✓ 验证码正确</p>
            )}
            
            {cooldown === 0 && !codeValidated && (
              <Button
                type="button"
                variant="link"
                onClick={handleResendCode}
                disabled={loading}
                className="text-sm p-0 h-auto mt-1"
              >
                没收到验证码？重新发送
              </Button>
            )}
          </div>
        )}
      </div>

      {/* 操作按钮 */}
      <div className="flex justify-between">
        <Button type="button" variant="outline" onClick={onBack}>
          上一步
        </Button>
        <Button
          type="button"
          onClick={handleNext}
          disabled={loading || !emailValidated || !codeValidated}
        >
          下一步
        </Button>
      </div>
    </div>
  );
};