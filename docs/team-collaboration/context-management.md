# OpenPenPal 共享上下文配置
# 所有Agent开发时的必要信息

project:
  name: OpenPenPal
  version: 2.0.0
  description: 跨校笔友信件传递系统
  completion_rate: 97%
  status: production_ready
  last_updated: "2025-07-22T15:45:00Z"
  updated_by: "Agent-3"

services:
  frontend:
    port: 3000
    url: http://localhost:3000
    agent: Agent-1
    status: production_ready
    completion: 98%
    features: "4级信使管理后台、任命系统、积分排行榜、博物馆功能"
    
  backend:
    port: 8080
    url: http://localhost:8080
    agent: Agent-1
    status: production_ready
    completion: 95%
    features: "认证服务、WebSocket通信、权限控制"
    
  write_service:
    port: 8001
    url: http://localhost:8001
    agent: Agent-2
    status: production_ready
    completion: 100%
    api_prefix: /api/letters
    features: "信件管理、博物馆、广场、商城、批量操作"
    
  courier_service:
    port: 8002
    url: http://localhost:8002
    agent: Agent-3
    status: production_ready
    completion: 98%
    api_prefix: /api/courier
    features: "4级信使系统、智能任务分配、积分排行榜、信号编码"
    
  admin_service:
    port: 8003
    url: http://localhost:8003
    agent: Agent-4
    status: production_ready
    completion: 95%
    api_prefix: /api/admin
    features: "用户管理、权限控制、内容审核、系统统计"
    
  ocr_service:
    port: 8004
    url: http://localhost:8004
    agent: Agent-5
    status: production_ready
    completion: 100%
    api_prefix: /api/ocr
    features: "图像识别、批量处理、缓存优化"
    
  gateway:
    port: 8000
    url: http://localhost:8000
    agent: Agent-3
    status: production_ready
    completion: 100%
    api_prefix: /api
    features: "统一网关、服务发现、负载均衡、认证授权、限流防护"

database:
  host: localhost
  port: 5432
  name: openpenpal
  schema_version: 2.0
  
redis:
  host: localhost
  port: 6379
  purpose: 
    - session_store
    - task_queue
    - websocket_pubsub

authentication:
  type: JWT
  secret_key: ${JWT_SECRET}
  token_expiry: 7d
  refresh_token_expiry: 30d
  
websocket:
  url: ws://localhost:8080/ws
  events:
    - LETTER_STATUS_UPDATE
    - COURIER_LOCATION_UPDATE
    - NEW_TASK_ASSIGNMENT
    - NEW_MESSAGE
    - SYSTEM_NOTIFICATION

api_standards:
  response_format:
    success:
      code: 0
      msg: "success"
      data: {}
      timestamp: "ISO 8601"
    error:
      code: "error_code"
      msg: "error message"
      error: {}
      timestamp: "ISO 8601"
      
  status_codes:
    - code: 0
      meaning: 成功
      http: 200
    - code: 1
      meaning: 参数错误
      http: 400
    - code: 2
      meaning: 无权限
      http: 403
    - code: 3
      meaning: 资源不存在
      http: 404
    - code: 500
      meaning: 服务器错误
      http: 500

letter_status_flow:
  - draft          # 草稿
  - generated      # 已生成二维码
  - collected      # 已收取
  - in_transit     # 投递中
  - delivered      # 已投递
  - failed         # 投递失败

user_roles:
  - user           # 普通用户
  - courier        # 信使
  - admin          # 管理员
  - super_admin    # 超级管理员