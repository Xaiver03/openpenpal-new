/**
 * OpenPenPal API Gateway v2 - 标准化配置
 * 符合微服务网关最佳实践的重构版本
 */

const express = require('express')
const { createProxyMiddleware } = require('http-proxy-middleware')
const cors = require('cors')
const helmet = require('helmet')
const rateLimit = require('express-rate-limit')
const jwt = require('jsonwebtoken')
const WebSocket = require('ws')
const http = require('http')

const app = express()
const server = http.createServer(app)

// 环境配置
const PORT = process.env.GATEWAY_PORT || 8000
const JWT_SECRET = process.env.JWT_SECRET || 'your-super-secret-jwt-key-change-in-production'

// 服务配置 - 标准化服务发现配置
const SERVICES = {
  WRITE: {
    url: process.env.WRITE_SERVICE_URL || 'http://localhost:8001',
    healthPath: '/health',
    timeout: 5000
  },
  COURIER: {
    url: process.env.COURIER_SERVICE_URL || 'http://localhost:8002', 
    healthPath: '/health',
    timeout: 5000
  },
  ADMIN: {
    url: process.env.ADMIN_SERVICE_URL || 'http://localhost:8003',
    healthPath: '/health',
    timeout: 5000
  },
  OCR: {
    url: process.env.OCR_SERVICE_URL || 'http://localhost:8004',
    healthPath: '/health',
    timeout: 5000
  }
}

// 路由配置 - 标准化路由映射
const ROUTES = [
  {
    path: '/api/v1/auth',
    service: 'WRITE',
    target: '/auth',
    description: '认证服务'
  },
  {
    path: '/api/v1/users',
    service: 'WRITE', 
    target: '/users',
    description: '用户管理'
  },
  {
    path: '/api/v1/schools',
    service: 'WRITE',
    target: '/schools', 
    description: '学校管理'
  },
  {
    path: '/api/v1/letters',
    service: 'WRITE',
    target: '/letters',
    description: '信件服务'
  },
  {
    path: '/api/v1/museum',
    service: 'WRITE',
    target: '/museum',
    description: '博物馆服务'
  },
  {
    path: '/api/v1/courier',
    service: 'COURIER',
    target: '/courier',
    description: '信使服务'
  },
  {
    path: '/api/v1/admin',
    service: 'ADMIN', 
    target: '/api/admin',
    description: '管理服务'
  },
  {
    path: '/api/v1/ocr',
    service: 'OCR',
    target: '/ocr',
    description: 'OCR服务'
  }
]

// 公共端点配置 - 无需认证的端点（注意：路径已剥离/api前缀）
const PUBLIC_ENDPOINTS = [
  '/v1/health',
  '/v1/auth/login',
  '/v1/auth/register', 
  '/v1/auth/forgot-password',
  '/v1/auth/reset-password',
  '/v1/schools/search',
  '/v1/schools/provinces',
  '/v1/letters/public',
  '/v1/museum/posts',
  '/v1/museum/featured'
]

// 安全中间件
app.use(helmet({
  crossOriginEmbedderPolicy: false,
  contentSecurityPolicy: false
}))

app.use(cors({
  origin: process.env.FRONTEND_URL || 'http://localhost:3000',
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Client-Version']
}))

// 健康检查端点 (在所有中间件之前)
app.get('/api/v1/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: '2.0.0',
    services: Object.keys(SERVICES)
  })
})

// 速率限制
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15分钟
  max: 1000, // 每个IP最多1000个请求
  message: { error: 'Too many requests, please try again later.' },
  standardHeaders: true,
  legacyHeaders: false
})

// JWT认证中间件 - 标准化实现
const authenticateToken = (req, res, next) => {
  console.log(`[DEBUG] Auth check for: ${req.path}`)
  
  // 检查是否为公共端点
  const isPublicEndpoint = PUBLIC_ENDPOINTS.some(endpoint => 
    req.path === endpoint || req.path.startsWith(endpoint + '/')
  )
  
  console.log(`[DEBUG] Is public endpoint: ${isPublicEndpoint}`)
  
  if (isPublicEndpoint) {
    console.log(`[INFO] Public endpoint allowed: ${req.path}`)
    return next()
  }

  const authHeader = req.headers['authorization']
  const token = authHeader && authHeader.split(' ')[1]
  
  if (!token) {
    return res.status(401).json({ 
      error: 'Access token required',
      code: 'UNAUTHORIZED'
    })
  }

  jwt.verify(token, JWT_SECRET, (err, user) => {
    if (err) {
      return res.status(403).json({ 
        error: 'Invalid or expired token',
        code: 'FORBIDDEN'
      })
    }
    req.user = user
    next()
  })
}

// 应用中间件
app.use('/api', limiter)
app.use('/api', authenticateToken)

// 标准化代理中间件创建函数
const createStandardProxy = (serviceConfig, targetPath) => {
  return createProxyMiddleware({
    target: serviceConfig.url,
    changeOrigin: true,
    pathRewrite: (path, req) => {
      // 标准化路径重写逻辑
      const originalPath = path
      const rewrittenPath = path.replace(/^\/api\/v1\/[^\/]+/, targetPath)
      console.log(`[PROXY] ${originalPath} -> ${serviceConfig.url}${rewrittenPath}`)
      return rewrittenPath
    },
    timeout: serviceConfig.timeout,
    onError: (err, req, res) => {
      console.error(`[ERROR] Proxy error for ${req.url}:`, err.message)
      if (!res.headersSent) {
        res.status(502).json({
          error: 'Service unavailable',
          service: serviceConfig.url,
          code: 'SERVICE_UNAVAILABLE'
        })
      }
    },
    onProxyReq: (proxyReq, req, res) => {
      // 添加标准请求头
      if (req.user) {
        proxyReq.setHeader('X-User-ID', req.user.id)
        proxyReq.setHeader('X-User-Role', req.user.role)
      }
      
      const requestId = require('crypto').randomUUID()
      proxyReq.setHeader('X-Request-ID', requestId)
      proxyReq.setHeader('X-Gateway-Version', '2.0.0')
      req.requestId = requestId
    },
    onProxyRes: (proxyRes, req, res) => {
      // 标准响应头处理
      proxyRes.headers['Access-Control-Allow-Origin'] = process.env.FRONTEND_URL || 'http://localhost:3000'
      proxyRes.headers['Access-Control-Allow-Credentials'] = 'true'
      
      if (req.requestId) {
        proxyRes.headers['X-Request-ID'] = req.requestId
      }
      proxyRes.headers['X-Gateway-Version'] = '2.0.0'
    }
  })
}

// 动态注册路由 - 标准化路由注册
ROUTES.forEach(route => {
  const serviceConfig = SERVICES[route.service]
  if (!serviceConfig) {
    console.error(`[ERROR] Service ${route.service} not found in configuration`)
    return
  }
  
  app.use(route.path, createStandardProxy(serviceConfig, route.target))
  console.log(`[ROUTE] ${route.path} -> ${serviceConfig.url}${route.target} (${route.description})`)
})

// WebSocket服务器配置
const wss = new WebSocket.Server({ 
  server,
  path: '/ws',
  verifyClient: (info) => {
    const token = new URL(info.req.url, 'http://localhost').searchParams.get('token')
    
    if (!token) return false

    try {
      const user = jwt.verify(token, JWT_SECRET)
      info.req.user = user
      return true
    } catch (err) {
      return false
    }
  }
})

// WebSocket连接管理
const connections = new Map()

wss.on('connection', (ws, req) => {
  const user = req.user
  const userId = user.id
  
  console.log(`[WS] Connected: ${user.username} (${userId})`)
  connections.set(userId, ws)

  const heartbeat = setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.ping()
    }
  }, 30000)

  ws.on('close', () => {
    console.log(`[WS] Disconnected: ${user.username} (${userId})`)
    connections.delete(userId)
    clearInterval(heartbeat)
  })

  ws.on('error', (error) => {
    console.error('[WS] Error:', error)
    connections.delete(userId)
    clearInterval(heartbeat)
  })

  ws.send(JSON.stringify({
    type: 'connection:established',
    data: { user_id: userId, timestamp: new Date().toISOString() }
  }))
})

// 错误处理中间件
app.use((err, req, res, next) => {
  console.error('[ERROR] Gateway error:', err)
  res.status(500).json({
    error: 'Internal server error',
    code: 'INTERNAL_ERROR',
    message: process.env.NODE_ENV === 'development' ? err.message : 'Something went wrong'
  })
})

// 404处理
app.use((req, res) => {
  res.status(404).json({
    error: 'Not Found',
    code: 'NOT_FOUND',
    message: 'The requested resource was not found',
    path: req.path
  })
})

// 服务器启动
server.listen(PORT, () => {
  console.log(`🚀 OpenPenPal API Gateway v2.0 running on port ${PORT}`)
  console.log(`📡 WebSocket server running on ws://localhost:${PORT}/ws`)
  console.log('🔗 Registered routes:')
  
  ROUTES.forEach(route => {
    const serviceConfig = SERVICES[route.service]
    console.log(`   ${route.path} -> ${serviceConfig.url}${route.target}`)
  })
  
  console.log('🔓 Public endpoints:')
  PUBLIC_ENDPOINTS.forEach(endpoint => {
    console.log(`   ${endpoint}`)
  })
})

// 优雅关闭
process.on('SIGTERM', () => {
  console.log('[INFO] SIGTERM received, shutting down gracefully')
  server.close(() => {
    console.log('[INFO] Process terminated')
    process.exit(0)
  })
})

module.exports = app