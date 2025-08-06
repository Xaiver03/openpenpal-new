/**
 * API Gateway Configuration for OpenPenPal Frontend-Backend Integration
 * OpenPenPal前后端集成API网关配置
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

// 服务配置
const SERVICES = {
  WRITE: process.env.WRITE_SERVICE_URL || 'http://localhost:8001',
  COURIER: process.env.COURIER_SERVICE_URL || 'http://localhost:8002',
  ADMIN: process.env.ADMIN_SERVICE_URL || 'http://localhost:8003',
  OCR: process.env.OCR_SERVICE_URL || 'http://localhost:8004'
}

// 中间件配置
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
    services: Object.keys(SERVICES)
  })
})

// JWT中间件
const authenticateToken = (req, res, next) => {
  const authHeader = req.headers['authorization']
  const token = authHeader && authHeader.split(' ')[1]
  
  if (!token) {
    // 对于某些公共端点允许匿名访问
    const publicEndpoints = [
      '/v1/auth/login',
      '/v1/auth/register',
      '/v1/auth/forgot-password',
      '/v1/auth/reset-password',
      '/v1/schools/search',
      '/v1/schools/provinces',
      '/v1/letters/public',
      '/v1/museum/posts',
      '/v1/museum/featured',
      '/v1/health'
    ]
    
    if (publicEndpoints.some(endpoint => req.path.startsWith(endpoint))) {
      console.log(`[DEBUG] Public endpoint allowed: ${req.path}`)
      return next()
    }
    
    return res.status(401).json({ error: 'Access token required' })
  }

  jwt.verify(token, JWT_SECRET, (err, user) => {
    if (err) {
      return res.status(403).json({ error: 'Invalid or expired token' })
    }
    req.user = user
    next()
  })
}

// 速率限制
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15分钟
  max: 1000, // 每个IP最多1000个请求
  message: { error: 'Too many requests, please try again later.' },
  standardHeaders: true,
  legacyHeaders: false
})

app.use('/api', limiter)
app.use('/api', authenticateToken)

// Debug middleware to see all requests
app.use('/api', (req, res, next) => {
  console.log(`[DEBUG] Request: ${req.method} ${req.originalUrl} -> ${req.path}`)
  next()
})

// 服务代理配置
const createServiceProxy = (target, pathRewrite = {}) => {
  return createProxyMiddleware({
    target,
    changeOrigin: true,
    pathRewrite,
    logLevel: 'debug',
    selfHandleResponse: false,
    onError: (err, req, res) => {
      console.error(`Proxy error for ${req.url}:`, err.message)
      if (!res.headersSent) {
        res.status(502).json({
          error: 'Service temporarily unavailable',
          service: target,
          code: 'PROXY_ERROR'
        })
      }
    },
    onProxyReq: (proxyReq, req, res) => {
      console.log(`[DEBUG] Proxying ${req.method} ${req.url} to ${proxyReq.path}`)
      console.log(`[DEBUG] Original URL: ${req.originalUrl}`)
      console.log(`[DEBUG] Proxy target path: ${proxyReq.path}`)
      
      // 添加用户信息到请求头
      if (req.user) {
        proxyReq.setHeader('X-User-ID', req.user.id)
        proxyReq.setHeader('X-User-Role', req.user.role)
      }
      
      // 添加请求ID用于追踪
      const requestId = require('crypto').randomUUID()
      proxyReq.setHeader('X-Request-ID', requestId)
      req.requestId = requestId
    },
    onProxyRes: (proxyRes, req, res) => {
      // 添加CORS头
      proxyRes.headers['Access-Control-Allow-Origin'] = process.env.FRONTEND_URL || 'http://localhost:3000'
      proxyRes.headers['Access-Control-Allow-Credentials'] = 'true'
      
      // 添加请求ID到响应
      if (req.requestId) {
        proxyRes.headers['X-Request-ID'] = req.requestId
      }
    }
  })
}

// 路由配置
// 认证服务 - 写信服务处理
app.use('/api/v1/auth', createServiceProxy(SERVICES.WRITE, {
  '^/v1/auth': '/auth'
}))

// 学校服务 - 写信服务处理
app.use('/api/v1/schools', createServiceProxy(SERVICES.WRITE, {
  '^/v1/schools': '/schools'
}))

// 用户服务 - 写信服务处理
app.use('/api/v1/users', createServiceProxy(SERVICES.WRITE, {
  '^/v1/users': '/users'
}))

// 信件服务 - 写信服务处理
app.use('/api/v1/letters', createServiceProxy(SERVICES.WRITE, {
  '^/': '/letters/'
}))

// 博物馆服务 - 写信服务处理
app.use('/api/v1/museum', createServiceProxy(SERVICES.WRITE, {
  '^/v1/museum': '/museum'
}))

// 信使服务 - 信使服务处理
app.use('/api/v1/courier', createServiceProxy(SERVICES.COURIER, {
  '^/v1/courier': '/courier'
}))

// 管理服务 - 管理服务处理
app.use('/api/v1/admin', createServiceProxy(SERVICES.ADMIN, {
  '^/v1/admin': '/api/admin'
}))

// OCR服务 - OCR服务处理
app.use('/api/v1/ocr', createServiceProxy(SERVICES.OCR, {
  '^/v1/ocr': '/ocr'
}))

// WebSocket服务器配置
const wss = new WebSocket.Server({ 
  server,
  path: '/ws',
  verifyClient: (info) => {
    const token = new URL(info.req.url, 'http://localhost').searchParams.get('token')
    
    if (!token) {
      return false
    }

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
  
  console.log(`WebSocket connected: ${user.username} (${userId})`)
  connections.set(userId, ws)

  // 心跳检测
  const heartbeat = setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.ping()
    }
  }, 30000)

  ws.on('message', (data) => {
    try {
      const message = JSON.parse(data)
      
      // 处理不同类型的消息
      switch (message.type) {
        case 'user:online':
          // 广播用户上线状态给相关用户
          broadcastToSchool(user.school_code, {
            type: 'user:status',
            data: { user_id: userId, status: 'online', timestamp: new Date().toISOString() }
          })
          break
          
        case 'user:offline':
          broadcastToSchool(user.school_code, {
            type: 'user:status',
            data: { user_id: userId, status: 'offline', timestamp: new Date().toISOString() }
          })
          break
          
        case 'courier:location_update':
          if (user.role.includes('courier')) {
            // 转发信使位置更新
            broadcastToRole('courier_coordinator', {
              type: 'courier:location_update',
              data: { courier_id: userId, ...message.data }
            })
          }
          break
          
        default:
          console.log('Unknown message type:', message.type)
      }
    } catch (error) {
      console.error('WebSocket message parse error:', error)
    }
  })

  ws.on('close', () => {
    console.log(`WebSocket disconnected: ${user.username} (${userId})`)
    connections.delete(userId)
    clearInterval(heartbeat)
    
    // 广播用户离线状态
    broadcastToSchool(user.school_code, {
      type: 'user:status',
      data: { user_id: userId, status: 'offline', timestamp: new Date().toISOString() }
    })
  })

  ws.on('error', (error) => {
    console.error('WebSocket error:', error)
    connections.delete(userId)
    clearInterval(heartbeat)
  })

  // 发送连接确认
  ws.send(JSON.stringify({
    type: 'connection:established',
    data: { user_id: userId, timestamp: new Date().toISOString() }
  }))
})

// 广播函数
function broadcastToUser(userId, message) {
  const ws = connections.get(userId)
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message))
  }
}

function broadcastToSchool(schoolCode, message) {
  // 这里需要维护用户学校映射，简化实现
  connections.forEach((ws, userId) => {
    if (ws.user && ws.user.school_code === schoolCode && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message))
    }
  })
}

function broadcastToRole(role, message) {
  // 这里需要维护用户角色映射，简化实现
  connections.forEach((ws, userId) => {
    if (ws.user && ws.user.role === role && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message))
    }
  })
}

function broadcastToAll(message) {
  connections.forEach((ws) => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message))
    }
  })
}

// 导出广播函数供其他模块使用
global.wssBroadcast = {
  toUser: broadcastToUser,
  toSchool: broadcastToSchool,
  toRole: broadcastToRole,
  toAll: broadcastToAll
}

// 错误处理
app.use((err, req, res, next) => {
  console.error('Gateway error:', err)
  res.status(500).json({
    error: 'Internal server error',
    message: process.env.NODE_ENV === 'development' ? err.message : 'Something went wrong'
  })
})

// 404处理
app.use((req, res) => {
  res.status(404).json({
    error: 'Not Found',
    message: 'The requested resource was not found',
    path: req.path
  })
})

// 服务器启动
server.listen(PORT, () => {
  console.log(`🚀 API Gateway running on port ${PORT}`)
  console.log(`📡 WebSocket server running on ws://localhost:${PORT}/ws`)
  console.log('🔗 Service routes:')
  console.log(`   /api/v1/auth/* → ${SERVICES.WRITE}`)
  console.log(`   /api/v1/letters/* → ${SERVICES.WRITE}`)
  console.log(`   /api/v1/schools/* → ${SERVICES.WRITE}`)
  console.log(`   /api/v1/courier/* → ${SERVICES.COURIER}`)
  console.log(`   /api/v1/admin/* → ${SERVICES.ADMIN}`)
  console.log(`   /api/v1/ocr/* → ${SERVICES.OCR}`)
})

// 优雅关闭
process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully')
  
  // 通知所有连接的客户端服务器即将关闭
  broadcastToAll({
    type: 'system:maintenance',
    data: { message: 'Server is shutting down', timestamp: new Date().toISOString() }
  })
  
  server.close(() => {
    console.log('Process terminated')
    process.exit(0)
  })
})

module.exports = app