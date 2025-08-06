# PostgreSQL 快速启动指南

## 安装后的快速设置

既然您已经安装了 PostgreSQL，以下是快速启动步骤：

### 1. 数据库已配置

✅ 数据库和用户已创建完成：
- 数据库名：`openpenpal`
- 用户名：`openpenpal`
- 密码：`openpenpal123`

### 2. 测试连接

```bash
# 测试数据库连接
PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal -d openpenpal -c "SELECT version();"
```

### 3. 启动应用

#### 开发模式（使用 SQLite）
```bash
./startup/quick-start.sh development --auto-open
```

#### 生产模式（使用 PostgreSQL）
```bash
./startup/quick-start.sh production --auto-open
```

### 4. 验证运行状态

```bash
# 检查服务状态
./startup/check-status.sh

# 测试 API
curl http://localhost:8080/health
```

### 5. 测试账号

- 管理员：`admin` / `admin123`
- 普通用户：`alice` / `secret`
- 信使：`courier1` / `courier123`

## 环境变量配置

如果需要自定义 PostgreSQL 连接：

```bash
export DATABASE_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=openpenpal
export DB_PASSWORD=openpenpal123
export DB_NAME=openpenpal
```

## 常见问题

### PostgreSQL 服务管理

```bash
# macOS (Homebrew)
brew services start postgresql@15
brew services stop postgresql@15
brew services restart postgresql@15

# 查看状态
brew services list | grep postgresql
```

### 数据库管理

```bash
# 连接到数据库
PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal -d openpenpal

# 查看所有表
\dt

# 查看用户数
SELECT COUNT(*) FROM users;

# 退出
\q
```

### 数据迁移

如果需要从 SQLite 迁移数据到 PostgreSQL：

```bash
cd backend
go run cmd/migrate-data/main.go ./openpenpal.db
```

## 总结

您的系统现在已完全支持双数据库模式：
- ✅ 开发使用 SQLite（快速简单）
- ✅ 生产使用 PostgreSQL（稳定可靠）
- ✅ 自动识别环境并选择数据库
- ✅ 所有测试通过