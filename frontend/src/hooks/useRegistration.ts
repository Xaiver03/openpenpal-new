import { useState, useCallback } from 'react';
import { 
  EmailVerificationRequest, 
  UserRegistrationRequest, 
  UserRegistrationResponse,
  EmailValidationResponse,
  UsernameValidationResponse,
  SchoolValidationResponse,
  VerificationCodeResponse,
  VerificationStatusResponse,
  EmailCodeValidationResponse 
} from '@/types/auth';

// 使用前端 API 路由作为临时解决方案
const API_BASE = '';

interface UseRegistrationReturn {
  // 状态
  loading: boolean;
  error: string | null;
  
  // 验证码相关
  sendVerificationCode: (request: EmailVerificationRequest) => Promise<VerificationCodeResponse>;
  resendVerificationCode: (email: string) => Promise<VerificationCodeResponse>;
  verifyEmailCode: (email: string, code: string) => Promise<EmailCodeValidationResponse>;
  getVerificationStatus: (email: string) => Promise<VerificationStatusResponse>;
  
  // 验证相关
  checkEmailAvailability: (email: string) => Promise<EmailValidationResponse>;
  checkUsernameAvailability: (username: string) => Promise<UsernameValidationResponse>;
  validateSchoolCode: (schoolCode: string) => Promise<SchoolValidationResponse>;
  
  // 注册
  registerUser: (request: UserRegistrationRequest) => Promise<UserRegistrationResponse>;
}

export const useRegistration = (): UseRegistrationReturn => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const apiCall = useCallback(async <T>(
    endpoint: string, 
    options: RequestInit = {}
  ): Promise<T> => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`${API_BASE}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        ...options,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || data.msg || '请求失败');
      }

      if (data.code !== 0) {
        throw new Error(data.message || data.msg || '操作失败');
      }

      return data.data;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '未知错误';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const sendVerificationCode = useCallback(async (
    request: EmailVerificationRequest
  ): Promise<VerificationCodeResponse> => {
    return apiCall('/api/auth/send-verification-code', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }, [apiCall]);

  const resendVerificationCode = useCallback(async (
    email: string
  ): Promise<VerificationCodeResponse> => {
    return apiCall(`/api/auth/resend-verification-code?email=${encodeURIComponent(email)}`, {
      method: 'POST',
    });
  }, [apiCall]);

  const verifyEmailCode = useCallback(async (
    email: string,
    code: string
  ): Promise<EmailCodeValidationResponse> => {
    return apiCall(`/api/auth/verify-email?email=${encodeURIComponent(email)}&code=${code}`);
  }, [apiCall]);

  const getVerificationStatus = useCallback(async (
    email: string
  ): Promise<VerificationStatusResponse> => {
    return apiCall(`/api/auth/verification-status?email=${encodeURIComponent(email)}`);
  }, [apiCall]);

  const checkEmailAvailability = useCallback(async (
    email: string
  ): Promise<EmailValidationResponse> => {
    return apiCall(`/api/auth/check-email?email=${encodeURIComponent(email)}`);
  }, [apiCall]);

  const checkUsernameAvailability = useCallback(async (
    username: string
  ): Promise<UsernameValidationResponse> => {
    return apiCall(`/api/auth/check-username?username=${encodeURIComponent(username)}`);
  }, [apiCall]);

  const validateSchoolCode = useCallback(async (
    schoolCode: string
  ): Promise<SchoolValidationResponse> => {
    return apiCall(`/api/auth/validate-school?schoolCode=${encodeURIComponent(schoolCode)}`);
  }, [apiCall]);

  const registerUser = useCallback(async (
    request: UserRegistrationRequest
  ): Promise<UserRegistrationResponse> => {
    return apiCall('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }, [apiCall]);

  return {
    loading,
    error,
    sendVerificationCode,
    resendVerificationCode,
    verifyEmailCode,
    getVerificationStatus,
    checkEmailAvailability,
    checkUsernameAvailability,
    validateSchoolCode,
    registerUser,
  };
};