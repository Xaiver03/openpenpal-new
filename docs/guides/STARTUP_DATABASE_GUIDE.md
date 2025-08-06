# OpenPenPal 启动脚本数据库配置指南

## 更新说明

启动脚本已更新，现在支持灵活的数据库配置，可以在 SQLite 和 PostgreSQL 之间自由切换。

## 快速使用

### 1. 使用 SQLite（开发模式）

```bash
# 默认使用 SQLite
./startup/quick-start.sh development
```

### 2. 使用 PostgreSQL（生产模式）

```bash
# 生产模式自动使用 PostgreSQL，无需额外配置！
./startup/quick-start.sh production

# 或使用自动打开浏览器
./startup/quick-start.sh production --auto-open
```

### 3. 在开发模式使用 PostgreSQL

```bash
# 如果开发环境也想使用 PostgreSQL
export DATABASE_TYPE=postgres
./startup/quick-start.sh development
```

### 3. 演示模式

```bash
# SQLite 演示
./startup/quick-start.sh demo --auto-open

# PostgreSQL 演示
export DATABASE_TYPE=postgres
./startup/quick-start.sh demo --auto-open
```

## 环境变量配置

启动脚本会自动根据 `DATABASE_TYPE` 设置相应的数据库配置：

### SQLite 模式（默认）
- `DATABASE_TYPE=sqlite`
- `DATABASE_URL=./openpenpal.db`

### PostgreSQL 模式
- `DATABASE_TYPE=postgres`
- `DB_HOST=localhost`
- `DB_PORT=5432`
- `DB_USER=openpenpal`
- `DB_PASSWORD=openpenpal123`
- `DB_NAME=openpenpal`
- `DB_SSLMODE=disable`

## 自定义配置

您可以通过环境变量覆盖默认配置：

```bash
# 自定义 PostgreSQL 配置
export DATABASE_TYPE=postgres
export DB_HOST=your-postgres-host
export DB_PORT=5432
export DB_USER=your-username
export DB_PASSWORD=your-password
export DB_NAME=your-database

./startup/quick-start.sh production
```

## 配置文件

后端会根据环境自动选择配置文件：

- **开发环境**: `backend/.env.development` (SQLite)
- **生产环境**: `backend/.env.production` (PostgreSQL)

## 验证配置

运行测试脚本验证配置：

```bash
./startup/test-database-startup.sh
```

这会测试：
1. SQLite 连接
2. PostgreSQL 连接
3. 启动脚本配置

## 常见问题

### 1. PostgreSQL 连接失败

确保 PostgreSQL 正在运行：
```bash
# macOS (Homebrew)
brew services start postgresql@15

# Docker
docker run --name openpenpal-postgres \
  -e POSTGRES_USER=openpenpal \
  -e POSTGRES_PASSWORD=openpenpal123 \
  -e POSTGRES_DB=openpenpal \
  -p 5432:5432 \
  -d postgres:15-alpine
```

### 2. 端口被占用

检查并停止占用端口的进程：
```bash
./startup/check-status.sh
./startup/stop-all.sh
```

### 3. 数据库迁移

从 SQLite 迁移到 PostgreSQL：
```bash
cd backend
go run cmd/migrate-data/main.go ./openpenpal.db
```

## 启动模式说明

| 模式 | 默认数据库 | 适用场景 |
|------|------------|----------|
| development | SQLite | 本地开发 |
| production | **PostgreSQL**（自动）| 生产部署 |
| demo | SQLite | 功能演示 |
| simple | SQLite | 最小化启动 |
| complete | SQLite | 全部服务 |

**重要**：`production` 模式会自动使用 PostgreSQL，无需手动设置！

## 最佳实践

1. **开发环境**：使用 SQLite，快速简单
2. **测试环境**：使用 PostgreSQL，接近生产
3. **生产环境**：必须使用 PostgreSQL

## 示例工作流

### 开发工作流
```bash
# 1. 启动开发环境
./startup/quick-start.sh development --auto-open

# 2. 开发和测试...

# 3. 停止服务
./startup/stop-all.sh
```

### 生产部署工作流
```bash
# 1. 设置 PostgreSQL
export DATABASE_TYPE=postgres

# 2. 编译后端
cd backend && go build -o openpenpal-backend

# 3. 启动生产服务
cd .. && ./startup/quick-start.sh production

# 4. 验证服务状态
./startup/check-status.sh
```

## 总结

启动脚本现在完全支持新的数据库配置系统：
- ✅ 自动检测数据库类型
- ✅ 灵活的环境变量配置
- ✅ SQLite 和 PostgreSQL 无缝切换
- ✅ 所有启动模式正常工作

如有问题，请参考 `GORM_POSTGRESQL_MIGRATION.md` 获取更多详细信息。