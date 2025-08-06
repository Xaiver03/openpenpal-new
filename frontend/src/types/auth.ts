export interface EmailVerificationRequest {
  email: string;
}

export interface UserRegistrationRequest {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  schoolCode: string;
  realName: string;
  studentId?: string;
  major?: string;
  grade?: number;
  className?: string;
  phone?: string;
  verificationCode: string;
  agreeToTerms: boolean;
  agreeToPrivacy: boolean;
}

export interface UserRegistrationResponse {
  userId: string;
  username: string;
  email: string;
  status: string;
  registeredAt: string;
  nextStep: string;
}

export interface EmailValidationResponse {
  email: string;
  available: boolean;
  message: string;
}

export interface UsernameValidationResponse {
  username: string;
  available: boolean;
  message: string;
}

export interface SchoolValidationResponse {
  schoolCode: string;
  valid: boolean;
  message: string;
}

export interface VerificationCodeResponse {
  email: string;
  message: string;
  expiryMinutes: number;
  cooldownSeconds: number;
}

export interface VerificationStatusResponse {
  email: string;
  canSend: boolean;
  cooldownSeconds: number;
  message: string;
}

export interface EmailCodeValidationResponse {
  email: string;
  isValid: boolean;
  message: string;
}