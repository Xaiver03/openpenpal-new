# PostgreSQL 本地安装配置指南

由于您还未安装 Docker，这里提供本地安装 PostgreSQL 的方法。

## macOS 安装 PostgreSQL

### 方法 1: 使用 Homebrew（推荐）

```bash
# 安装 Homebrew（如果还没有）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 PostgreSQL
brew install postgresql@15

# 启动 PostgreSQL 服务
brew services start postgresql@15

# 检查服务状态
brew services list
```

### 方法 2: 使用 Postgres.app

1. 访问 https://postgresapp.com/
2. 下载并安装 Postgres.app
3. 启动应用程序
4. 点击 "Initialize" 创建默认数据库

## 创建 OpenPenPal 数据库

安装完成后，创建项目所需的数据库和用户：

```bash
# 进入 PostgreSQL 命令行
psql postgres

# 在 psql 中执行以下命令：
CREATE USER openpenpal WITH PASSWORD 'openpenpal123';
CREATE DATABASE openpenpal OWNER openpenpal;
GRANT ALL PRIVILEGES ON DATABASE openpenpal TO openpenpal;
\q
```

## 测试连接

创建测试脚本验证连接：

```bash
# 创建测试文件
cat > test-postgres.sh << 'EOF'
#!/bin/bash
echo "测试 PostgreSQL 连接..."
psql -h localhost -U openpenpal -d openpenpal -c "SELECT version();"
EOF

chmod +x test-postgres.sh
./test-postgres.sh
```

## 配置后端使用 PostgreSQL

1. 复制生产环境配置：
```bash
cd backend
cp .env.production .env
```

2. 编辑 .env 文件，确保以下配置正确：
```
DATABASE_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=openpenpal
DB_PASSWORD=openpenpal123
DB_NAME=openpenpal
DB_SSLMODE=disable
```

3. 测试后端连接：
```bash
go run main.go
```

## 如果遇到问题

### 1. 密码认证失败
编辑 PostgreSQL 配置文件：
```bash
# macOS Homebrew 安装位置
nano /usr/local/var/postgresql@15/pg_hba.conf

# 将 METHOD 改为 trust 或 md5
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
```

重启 PostgreSQL：
```bash
brew services restart postgresql@15
```

### 2. 端口被占用
检查端口：
```bash
lsof -i :5432
```

### 3. 连接被拒绝
确保 PostgreSQL 正在运行：
```bash
pg_ctl -D /usr/local/var/postgresql@15 status
```

## 下一步

PostgreSQL 安装配置完成后：

1. 运行后端应用
2. 数据库表会自动创建
3. 可以开始迁移数据

需要帮助请告诉我！