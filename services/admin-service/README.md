# OpenPenPal Admin Service

> 管理后台服务 - Agent #4 开发模块

## 🎯 功能概述

OpenPenPal管理后台服务，提供完整的用户管理、信件监控、信使管理、数据统计和系统配置功能。

### 核心功能
- 👥 **用户管理** - 用户CRUD、角色分配、权限管理
- 📧 **信件管理** - 信件监控、状态更新、批量操作
- 🚚 **信使管理** - 信使审核、任务分配、绩效统计
- 📊 **数据统计** - 多维度统计分析和报表
- ⚙️ **系统配置** - 动态配置管理和权限控制

## 🏗️ 技术架构

### 后端技术栈
- **框架**: Spring Boot 3.2.1 + Java 17
- **安全**: Spring Security + JWT认证
- **数据库**: PostgreSQL + Spring Data JPA
- **缓存**: Redis + Spring Data Redis  
- **文档**: SpringDoc OpenAPI 3.0
- **容器**: Docker + Docker Compose

### 权限系统
- **RBAC模型** - 基于角色的权限控制
- **细粒度权限** - 资源-操作级别的权限验证
- **范围限制** - 学校级别的数据访问控制
- **操作审计** - 完整的管理员操作日志

## 🚀 快速启动

### 开发环境

1. **启动开发环境**
```bash
# 启动数据库和Redis
./start-dev.sh

# 启动Spring Boot应用
cd backend
./mvnw spring-boot:run -Dspring-boot.run.profiles=dev
```

2. **访问服务**
- **API服务**: http://localhost:8003/api/admin
- **API文档**: http://localhost:8003/api/admin/swagger-ui.html
- **健康检查**: http://localhost:8003/api/admin/actuator/health

3. **管理工具**
- **PgAdmin**: http://localhost:5050 (admin@openpenpal.com/admin123)
- **Redis Commander**: http://localhost:8081

### 生产环境

```bash
# 构建和启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f admin-service
```

## 📡 API接口

### 认证方式
```http
Authorization: Bearer <jwt_token>
X-Admin-Permission: <required_permission>
```

### 主要接口

#### 用户管理
```http
GET    /api/admin/users              # 获取用户列表
GET    /api/admin/users/{id}         # 获取用户详情
PUT    /api/admin/users/{id}         # 更新用户信息
DELETE /api/admin/users/{id}         # 删除用户
POST   /api/admin/users/{id}/unlock  # 解锁用户
```

#### 信件管理
```http
GET    /api/admin/letters                    # 获取信件列表
GET    /api/admin/letters/{id}               # 获取信件详情
PUT    /api/admin/letters/{id}/status        # 更新信件状态
PUT    /api/admin/letters/{id}/urgent        # 标记紧急状态
PUT    /api/admin/letters/batch/status       # 批量更新状态
```

#### 统计分析
```http
GET    /api/admin/users/stats/role           # 用户角色统计
GET    /api/admin/users/stats/school         # 用户学校统计
GET    /api/admin/letters/stats/overview     # 信件概览统计
GET    /api/admin/letters/stats/by-status    # 信件状态统计
```

## 🔐 权限系统

### 权限级别
- **super_admin** - 超级管理员 (所有权限)
- **platform_admin** - 平台管理员 (跨校管理)
- **school_admin** - 学校管理员 (本校管理)
- **courier_manager** - 信使协调员 (信使管理)

### 权限范围
- **user.\*** - 用户管理权限
- **letter.\*** - 信件管理权限
- **courier.\*** - 信使管理权限
- **stats.\*** - 统计查看权限
- **config.\*** - 配置管理权限

### 使用示例
```java
@RequiresPermission("user.read")
public ApiResponse<PageResponse<UserDto>> getUsers(...) {
    // 需要用户读取权限
}

@RequiresPermission(value = "user.write", requireScope = true)
public ApiResponse<UserDto> updateUser(@PathVariable UUID userId, ...) {
    // 需要用户写入权限，并检查范围限制
}
```

## 📊 数据库设计

### 核心表结构
- **users** - 用户基础信息和安全字段
- **permissions** - 权限定义表
- **role_permissions** - 角色权限关联
- **admin_logs** - 管理员操作日志
- **system_config** - 系统配置表

### 权限数据初始化
数据库自动初始化包含：
- 基础权限定义 (24个权限)
- 角色权限关联 (4个角色)
- 系统配置项 (7个配置)

## 🔧 开发指南

### 环境变量
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=openpenpal
DB_USERNAME=postgres
DB_PASSWORD=postgres

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# 服务配置
WRITE_SERVICE_URL=http://localhost:8001
COURIER_SERVICE_URL=http://localhost:8002

# 安全配置
JWT_SECRET=your-secret-key
ADMIN_DEFAULT_PASSWORD=admin123
```

### 添加新权限
1. 在数据库中插入权限记录
2. 分配给相应角色
3. 在控制器上添加 `@RequiresPermission` 注解

### 添加新API
1. 创建DTO类 (请求/响应)
2. 实现Service接口和实现类
3. 创建Controller并添加权限注解
4. 更新OpenAPI文档

## 🐳 Docker配置

### 开发环境
```bash
# 启动开发数据库
docker-compose -f docker-compose.dev.yml up -d

# 停止
docker-compose -f docker-compose.dev.yml down
```

### 生产环境
```bash
# 完整服务栈
docker-compose up -d

# 扩展服务
docker-compose up -d --scale admin-service=3
```

## 📝 日志和监控

### 日志级别
- **开发环境**: DEBUG (详细SQL日志)
- **生产环境**: INFO (关键操作日志)

### 健康检查
- **Spring Actuator**: `/actuator/health`
- **数据库连接**: 自动检测
- **Redis连接**: 自动检测
- **外部服务**: 写信服务、信使服务

### 操作审计
所有管理员操作自动记录：
- 操作用户和目标
- 请求详情和结果
- IP地址和用户代理
- 时间戳和错误信息

## 🔄 与其他服务集成

### 服务调用
```java
// 调用写信服务
@Autowired
private WebClient.Builder webClientBuilder;

Map<String, Object> response = webClientBuilder.build()
    .get()
    .uri(writeServiceUrl + "/api/letters/" + letterId)
    .retrieve()
    .bodyToMono(Map.class)
    .block();
```

### WebSocket事件
监听和推送系统事件：
- 用户状态变更
- 信件状态更新
- 信使任务分配
- 系统通知

## 📚 相关文档

- [多Agent协同框架](../../MULTI_AGENT_COORDINATION.md)
- [统一API规范](../../docs/api/UNIFIED_API_SPECIFICATION.md)
- [共享配置管理](../../AGENT_CONTEXT_MANAGEMENT.md)
- [Agent任务卡片](../../agent-tasks/AGENT-4-ADMIN-SERVICE.md)

## 🛠️ 故障排除

### 常见问题

1. **数据库连接失败**
```bash
# 检查数据库状态
docker-compose -f docker-compose.dev.yml ps postgres-dev

# 查看数据库日志
docker-compose -f docker-compose.dev.yml logs postgres-dev
```

2. **权限验证失败**
```bash
# 检查JWT配置
grep JWT_SECRET backend/src/main/resources/application-dev.yml

# 查看权限日志
docker-compose logs admin-service | grep Permission
```

3. **服务调用超时**
```bash
# 检查服务连通性
curl http://localhost:8001/health  # 写信服务
curl http://localhost:8002/health  # 信使服务
```

---

**Agent #4 开发**: 企业级Spring Boot架构，完善的RBAC权限系统，生产就绪的管理后台服务。