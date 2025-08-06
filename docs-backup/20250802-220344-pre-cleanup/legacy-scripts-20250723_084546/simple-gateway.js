const express = require('express')
const { createProxyMiddleware } = require('http-proxy-middleware')
const cors = require('cors')

const app = express()
const PORT = 8005

app.use(cors({
  origin: 'http://localhost:3000',
  credentials: true
}))

app.use(express.json())

// Simple proxy without authentication for testing
app.use('/api/v1/auth', createProxyMiddleware({
  target: 'http://localhost:8001',
  changeOrigin: true,
  pathRewrite: {
    '^/api/v1/auth': '/auth'
  },
  onError: (err, req, res) => {
    console.error('Proxy error:', err.message)
    res.status(502).json({ error: 'Gateway error' })
  },
  onProxyReq: (proxyReq, req) => {
    console.log(`Proxying ${req.method} ${req.url} to ${proxyReq.path}`)
  }
}))

app.listen(PORT, () => {
  console.log(`Simple Gateway running on port ${PORT}`)
  console.log('Route: /api/v1/auth/* → http://localhost:8001/auth/*')
})