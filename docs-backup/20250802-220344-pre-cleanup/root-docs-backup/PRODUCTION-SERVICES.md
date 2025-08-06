# OpenPenPal 生产模式服务清单

## 基础设施服务（必需）

### 1. PostgreSQL 数据库
- **端口**: 5432
- **启动方式**: `docker-compose up -d postgres` 或本地PostgreSQL
- **状态**: ✅ 已在运行（端口5432有进程）

### 2. Redis 缓存（可选但推荐）
- **端口**: 6379
- **启动方式**: `docker-compose up -d redis` 或本地Redis
- **用途**: 缓存、会话、队列

## 核心后端服务（必需）

### 1. Go主后端服务
- **端口**: 8080  
- **路径**: `/backend`
- **启动命令**: `cd backend && ./openpenpal-backend`
- **状态**: ✅ 已在运行
- **功能**: 核心API、用户认证、信件管理、WebSocket等

### 2. API网关服务
- **端口**: 8000
- **路径**: `/services/gateway`
- **启动命令**: `cd services/gateway && ./bin/gateway`
- **功能**: 统一入口、路由、限流、监控

## 微服务（按需启动）

### 1. 写信服务（Python）
- **端口**: 8001
- **路径**: `/services/write-service`
- **启动命令**: `cd services/write-service && python app/main.py`
- **功能**: 信件写作、草稿管理、广场功能

### 2. 信使服务（Go）
- **端口**: 8002
- **路径**: `/services/courier-service`
- **启动命令**: `cd services/courier-service && ./bin/courier-service`
- **功能**: 四级信使系统、任务分配、扫码功能

### 3. 管理服务（Java）
- **端口**: 8003
- **路径**: `/services/admin-service/backend`
- **启动命令**: `cd services/admin-service/backend && ./mvnw spring-boot:run`
- **功能**: 后台管理、内容审核、系统配置

### 4. OCR服务（Python）
- **端口**: 8004
- **路径**: `/services/ocr-service`
- **启动命令**: `cd services/ocr-service && python app/main.py`
- **功能**: 图像文字识别、批量处理

## 前端服务

### 1. 主前端
- **端口**: 3000
- **路径**: `/frontend`
- **启动命令**: `cd frontend && npm run dev`
- **状态**: ✅ 已在运行

### 2. 管理后台前端
- **端口**: 3001
- **路径**: `/services/admin-service/frontend`
- **启动命令**: `cd services/admin-service/frontend && npm run dev`

## 生产模式启动命令

### 完整启动（推荐）
```bash
# 启动所有服务（包括基础设施）
./startup/quick-start.sh production --auto-open

# 查看服务状态
./startup/check-status.sh

# 停止所有服务
./startup/stop-all.sh
```

### 手动启动基础设施
```bash
# 1. 启动PostgreSQL
docker-compose up -d postgres

# 2. 启动Redis（可选）
docker-compose up -d redis

# 3. 运行数据库迁移
cd backend && go run main.go migrate
```

### 逐个启动服务
```bash
# 1. 启动主后端
cd backend && ./openpenpal-backend &

# 2. 启动网关
cd services/gateway && go build -o bin/gateway cmd/main.go && ./bin/gateway &

# 3. 启动写信服务
cd services/write-service && python app/main.py &

# 4. 启动信使服务  
cd services/courier-service && go build -o bin/courier-service cmd/main.go && ./bin/courier-service &

# 5. 启动前端
cd frontend && npm run dev &
```

## 健康检查端点

- 主后端: http://localhost:8080/health
- 网关: http://localhost:8000/health
- 写信服务: http://localhost:8001/health
- 信使服务: http://localhost:8002/health
- 管理服务: http://localhost:8003/health
- OCR服务: http://localhost:8004/health
- 前端: http://localhost:3000/health
- 管理后台: http://localhost:3001/health

## 注意事项

1. **数据库连接**: 生产模式使用PostgreSQL，确保数据库服务正在运行
2. **端口冲突**: 确保所需端口未被占用
3. **依赖安装**: Python服务需要虚拟环境和依赖包
4. **编译Go服务**: Go服务需要先编译再运行
5. **Java环境**: 管理服务需要Java 11+和Maven