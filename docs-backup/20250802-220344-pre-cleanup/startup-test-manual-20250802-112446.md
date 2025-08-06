# OpenPenPal Startup Modes Test Report (Manual)
Date: 2025年 8月 2日 星期六 11时24分46秒 CST

## System Information
- Platform: Darwin
- Node Version: v24.2.0
- Go Version: go version go1.24.5 darwin/arm64
- Python Version: Python 3.9.6
- Java Version: The operation couldn’t be completed. Unable to locate a Java Runtime.

## Test Results


### Mode: simple
Start Time: 11:24:46

**Status**: ✅ Started successfully
**Duration**: 11 seconds

**Service Health Check:**

- ✅ Go Backend (port 8080): Healthy
- ✅ Frontend (port 3000): Healthy

**Actually Running Processes:**
```
Port 3000: node 51428 *:hbci
Port 8080: openpenpa 51380 *:http-alt
```

End Time: 11:25:09

---

### Mode: demo
Start Time: 11:25:14

**Status**: ✅ Started successfully
**Duration**: 11 seconds

**Service Health Check:**

- ✅ Go Backend (port 8080): Healthy
- ✅ Frontend (port 3000): Healthy

**Actually Running Processes:**
```
Port 3000: node 52245 *:hbci
Port 8080: openpenpa 52165 *:http-alt
```

End Time: 11:25:37

---

### Mode: development
Start Time: 11:25:42

**Status**: ✅ Started successfully
**Duration**: 11 seconds

**Service Health Check:**

- ✅ Go Backend (port 8080): Healthy
- ✅ Frontend (port 3000): Healthy

**Actually Running Processes:**
```
Port 3000: node 52993 *:hbci
Port 8080: openpenpa 52943 *:http-alt
```

End Time: 11:26:05

---

### Mode: mock
Start Time: 11:26:10

**Status**: ✅ Started successfully
**Duration**: 10 seconds

**Service Health Check:**

- ✅ Simple Mock (port 8000): Healthy
- ✅ Frontend (port 3000): Healthy

**Actually Running Processes:**
```
Port 3000: node 53719 *:hbci
Port 8000: node 53604 *:irdmi
Port 8001: node 53604 *:vcom-tunnel
Port 8002: node 53604 *:teradataordbms
Port 8003: node 53604 *:8003
Port 8004: node 53604 *:8004
```

End Time: 11:26:33

---

### Mode: production
Start Time: 11:26:38

**Status**: ✅ Started successfully
**Duration**: 60 seconds

**Service Health Check:**

- ✅ Go Backend (port 8080): Healthy
- ✅ Frontend (port 3000): Healthy
- ⚠️ Gateway (port 8000): Not started
- ⚠️ Write Service (port 8001): Not started
- ⚠️ Courier Service (port 8002): Not started
- ⚠️ Admin Service (port 8003): Not started (Java required)
- ⚠️ OCR Service (port 8004): Not started
- ⚠️ Admin Frontend (port 3001): Not started

**Actually Running Processes:**
```
Port 3000: node 55068 *:hbci
Port 8080: openpenpa 54997 *:http-alt
```

End Time: 11:27:51

---

### Mode: complete
Start Time: 11:27:56

**Status**: ❌ Failed to start or timed out
**Duration**: 151 seconds

**Error Log (last 30 lines):**
```
[0;34m[INFO][0m [11:28:00] 模式: complete
[0;34m[INFO][0m [11:28:00] 检查系统要求...
[0;32m[SUCCESS][0m [11:28:00] ✓ Node.js v24.2.0
[0;32m[SUCCESS][0m [11:28:00] ✓ npm 11.5.1
[0;34m[INFO][0m [11:28:00] 检查项目依赖...
[0;32m[SUCCESS][0m [11:28:00] 依赖检查完成
[0;34m[INFO][0m [11:28:00] 准备启动环境...
[0;34m[INFO][0m [11:28:00] 清理可能运行的服务...
[0;32m[SUCCESS][0m [11:28:00] 通过端口停止了 8 个服务
[0;32m[SUCCESS][0m [11:28:01] 所有服务已成功停止
[0;32m[SUCCESS][0m [11:28:04] 环境准备完成
[0;34m[INFO][0m [11:28:04] 启动所有服务...
[0;34m[INFO][0m [11:28:04] 启动服务: go-backend
[0;32m[SUCCESS][0m [11:28:05] Go后端启动成功 (PID: 56082, 端口: 8080)
[0;34m[INFO][0m [11:28:05]   • 数据库: PostgreSQL (rocalight@localhost:5432/openpenpal)
[0;34m[INFO][0m [11:28:05]   • WebSocket: ws://localhost:8080/api/v1/ws/connect
[0;34m[INFO][0m [11:28:05]   • 健康检查: http://localhost:8080/health
[0;34m[INFO][0m [11:28:07] 启动服务: real-gateway
[0;31m[ERROR][0m [11:28:39] 网关服务启动失败
[0;31m[ERROR][0m [11:28:39] real-gateway 启动失败
[0;34m[INFO][0m [11:28:41] 启动服务: real-write-service
[0;31m[ERROR][0m [11:29:13] 写信服务启动失败
[0;31m[ERROR][0m [11:29:13] real-write-service 启动失败
[0;34m[INFO][0m [11:29:15] 启动服务: real-courier-service
[0;31m[ERROR][0m [11:29:47] 信使服务启动失败
[0;31m[ERROR][0m [11:29:47] real-courier-service 启动失败
[0;34m[INFO][0m [11:29:49] 启动服务: real-admin-service
[0;31m[ERROR][0m [11:30:20] 管理服务启动失败
[0;31m[ERROR][0m [11:30:20] real-admin-service 启动失败
[0;34m[INFO][0m [11:30:22] 启动服务: real-ocr-service
```

End Time: 11:30:33

---

## Summary

Test completed at: 2025年 8月 2日 星期六 11时30分38秒 CST

### Key Findings
- Simple modes (simple, demo, development, mock) should start quickly
- Complex modes (production, complete) may take longer and some services may fail
- Admin Service (port 8003) is expected to fail if Java is not installed
- Python-based services (Write, OCR) may fail if Python virtual environments are not set up
