/**
 * Mock 服务配置
 * 定义各个微服务的端口、路径、响应延迟等配置
 */

export const SERVICES = {
  'gateway': {
    name: 'API Gateway',
    port: 8000,
    basePath: '/api',
    delay: { min: 50, max: 200 },
    enabled: true
  },
  'main-backend': {
    name: 'Main Backend Service',
    port: 8080,
    basePath: '/api',
    delay: { min: 100, max: 300 },
    enabled: true
  },
  'write-service': {
    name: 'Write Service',
    port: 8001,
    basePath: '/api',
    delay: { min: 150, max: 400 },
    enabled: true
  },
  'courier-service': {
    name: 'Courier Service',
    port: 8002,
    basePath: '/api',
    delay: { min: 100, max: 250 },
    enabled: true
  },
  'admin-service': {
    name: 'Admin Service',
    port: 8003,
    basePath: '/api/admin',
    delay: { min: 200, max: 500 },
    enabled: true
  },
  'ocr-service': {
    name: 'OCR Service',
    port: 8004,
    basePath: '/api',
    delay: { min: 500, max: 2000 }, // OCR 服务通常较慢
    enabled: true
  }
};

// 默认配置
export const DEFAULT_CONFIG = {
  // 全局延迟设置
  globalDelay: {
    enabled: false,
    min: 100,
    max: 500
  },
  
  // 错误模拟概率 (0-1)
  errorSimulation: {
    enabled: false,
    probability: 0.1, // 10% 概率返回错误
    types: ['network', 'server', 'timeout']
  },
  
  // 日志配置
  logging: {
    level: 'info', // debug, info, warn, error
    format: 'combined', // combined, common, dev
    requests: true,
    responses: true,
    errors: true
  },
  
  // CORS 配置
  cors: {
    origin: ['http://localhost:3000', 'http://localhost:3001'],
    credentials: true,
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With']
  },
  
  // JWT 配置
  jwt: {
    secret: 'mock-service-secret-key-for-testing-only',
    expiresIn: '24h',
    issuer: 'openpenpal-mock-service',
    audience: 'openpenpal-frontend'
  }
};

// 获取服务配置
export function getServiceConfig(serviceName) {
  return SERVICES[serviceName] || null;
}

// 获取所有启用的服务
export function getEnabledServices() {
  return Object.entries(SERVICES)
    .filter(([_, config]) => config.enabled)
    .reduce((acc, [name, config]) => {
      acc[name] = config;
      return acc;
    }, {});
}

// 检查端口是否被占用
export function getAvailablePort(preferredPort) {
  // 在实际实现中，这里应该检查端口可用性
  // 这里简化处理，直接返回首选端口
  return preferredPort;
}