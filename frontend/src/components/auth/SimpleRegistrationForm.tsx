import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Eye, EyeOff } from 'lucide-react';

interface SimpleRegistrationFormProps {
  onSuccess: (userId: string) => void;
  onBack: () => void;
}

export const SimpleRegistrationForm: React.FC<SimpleRegistrationFormProps> = ({
  onSuccess,
  onBack,
}) => {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    schoolCode: '',
    realName: '',
    agreeToTerms: false,
    agreeToPrivacy: false,
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // 基本验证
    if (!formData.username || !formData.email || !formData.password || 
        !formData.confirmPassword || !formData.schoolCode || !formData.realName) {
      setError('请填写所有必填字段');
      return;
    }

    if (formData.password !== formData.confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }

    if (!formData.agreeToTerms || !formData.agreeToPrivacy) {
      setError('请同意用户协议和隐私政策');
      return;
    }

    setLoading(true);
    
    try {
      // 使用前端 API 路由
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: formData.username,
          email: formData.email,
          password: formData.password,
          confirmPassword: formData.confirmPassword,
          schoolCode: formData.schoolCode,
          realName: formData.realName,
          verificationCode: '000000', // 使用临时验证码，实际应该传入真实验证码
          agreeToTerms: formData.agreeToTerms,
          agreeToPrivacy: formData.agreeToPrivacy
        }),
      });

      const result = await response.json();

      if (result.code === 0) {
        onSuccess(result.data.userId);
      } else {
        setError(result.message || '注册失败，请稍后重试');
      }
    } catch (err) {
      setError('网络错误，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 max-w-md mx-auto p-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold">用户注册</h2>
        <p className="text-gray-600 mt-2">请填写注册信息</p>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <Label htmlFor="username">用户名 *</Label>
          <Input
            id="username"
            value={formData.username}
            onChange={(e) => handleInputChange('username', e.target.value)}
            placeholder="请输入用户名"
            required
          />
        </div>

        <div>
          <Label htmlFor="email">邮箱 *</Label>
          <Input
            id="email"
            type="email"
            value={formData.email}
            onChange={(e) => handleInputChange('email', e.target.value)}
            placeholder="请输入邮箱"
            required
          />
        </div>

        <div>
          <Label htmlFor="password">密码 *</Label>
          <div className="relative">
            <Input
              id="password"
              type={showPassword ? "text" : "password"}
              value={formData.password}
              onChange={(e) => handleInputChange('password', e.target.value)}
              placeholder="请输入密码"
              className="pr-10"
              required
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
              aria-label={showPassword ? "隐藏密码" : "显示密码"}
            >
              {showPassword ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>

        <div>
          <Label htmlFor="confirmPassword">确认密码 *</Label>
          <div className="relative">
            <Input
              id="confirmPassword"
              type={showConfirmPassword ? "text" : "password"}
              value={formData.confirmPassword}
              onChange={(e) => handleInputChange('confirmPassword', e.target.value)}
              placeholder="请再次输入密码"
              className="pr-10"
              required
            />
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
              aria-label={showConfirmPassword ? "隐藏密码" : "显示密码"}
            >
              {showConfirmPassword ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>

        <div>
          <Label htmlFor="schoolCode">学校编码 *</Label>
          <Input
            id="schoolCode"
            value={formData.schoolCode}
            onChange={(e) => handleInputChange('schoolCode', e.target.value.toUpperCase())}
            placeholder="请输入学校编码，如：BJDX01"
            required
          />
        </div>

        <div>
          <Label htmlFor="realName">真实姓名 *</Label>
          <Input
            id="realName"
            value={formData.realName}
            onChange={(e) => handleInputChange('realName', e.target.value)}
            placeholder="请输入真实姓名"
            required
          />
        </div>

        <div className="space-y-3">
          <div className="flex items-start space-x-2">
            <Checkbox
              id="agreeToTerms"
              checked={formData.agreeToTerms}
              onCheckedChange={(checked) => handleInputChange('agreeToTerms', checked)}
            />
            <Label htmlFor="agreeToTerms" className="text-sm">
              我已阅读并同意 <a href="/terms" target="_blank" className="text-blue-600 hover:underline">《用户协议》</a>
            </Label>
          </div>
          
          <div className="flex items-start space-x-2">
            <Checkbox
              id="agreeToPrivacy"
              checked={formData.agreeToPrivacy}
              onCheckedChange={(checked) => handleInputChange('agreeToPrivacy', checked)}
            />
            <Label htmlFor="agreeToPrivacy" className="text-sm">
              我已阅读并同意 <a href="/privacy" target="_blank" className="text-blue-600 hover:underline">《隐私政策》</a>
            </Label>
          </div>
        </div>

        <div className="flex justify-between pt-4">
          <Button type="button" variant="outline" onClick={onBack}>
            返回
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? '注册中...' : '注册'}
          </Button>
        </div>
      </form>
    </div>
  );
};