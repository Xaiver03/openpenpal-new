# GORM + PostgreSQL 迁移方案

## 1. 为什么选择 GORM + PostgreSQL

### 优势
- ✅ **最小改动**：只需更改数据库驱动和连接字符串
- ✅ **保留现有代码**：所有业务逻辑和模型定义无需改动
- ✅ **性能优秀**：Go 原生性能 + PostgreSQL 的强大功能
- ✅ **部署简单**：单一 Go 二进制文件，无需 Node.js 运行时

### 对比 Prisma
| 特性 | GORM | Prisma |
|------|------|---------|
| 改动成本 | 低（只改配置） | 高（重写数据层） |
| 性能 | 优秀 | 良好 |
| 类型安全 | 良好 | 优秀 |
| 迁移工具 | 基础 | 优秀 |
| 团队熟悉度 | 高 | 低 |

## 2. 迁移步骤

### Step 1: 安装 PostgreSQL

```bash
# Docker 方式（推荐）
docker run --name openpenpal-postgres \
  -e POSTGRES_USER=openpenpal \
  -e POSTGRES_PASSWORD=openpenpal123 \
  -e POSTGRES_DB=openpenpal \
  -p 5432:5432 \
  -d postgres:15-alpine

# 或使用之前创建的 docker-compose.yml
docker-compose up -d postgres
```

### Step 2: 更新数据库配置

```go
// backend/internal/config/database.go
package config

import (
    "fmt"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func SetupDatabase(cfg *Config) (*gorm.DB, error) {
    var db *gorm.DB
    var err error

    if cfg.DatabaseType == "postgres" {
        // PostgreSQL 配置
        dsn := cfg.DatabaseURL
        if dsn == "" {
            dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
                getEnv("DB_HOST", "localhost"),
                getEnv("DB_USER", "openpenpal"),
                getEnv("DB_PASSWORD", "openpenpal123"),
                getEnv("DB_NAME", "openpenpal"),
                getEnv("DB_PORT", "5432"),
            )
        }
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
            PrepareStmt: true,
        })
    } else {
        // SQLite (开发环境)
        db, err = gorm.Open(sqlite.Open(cfg.DatabaseURL), &gorm.Config{})
    }

    if err != nil {
        return nil, err
    }

    // 配置连接池
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)

    return db, nil
}
```

### Step 3: 更新模型以支持 PostgreSQL

```go
// backend/internal/models/user.go
type User struct {
    ID           string         `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    Username     string         `gorm:"uniqueIndex;not null"`
    Email        string         `gorm:"uniqueIndex;not null"`
    PasswordHash string         `gorm:"not null"`
    Role         string         `gorm:"type:varchar(20);default:'user'"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// 添加 PostgreSQL 特定功能
func (db *gorm.DB) EnablePostgreSQLExtensions() error {
    // 启用 UUID 扩展
    return db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
}
```

### Step 4: 创建迁移脚本

```go
// backend/cmd/migrate/main.go
package main

import (
    "log"
    "openpenpal-backend/internal/config"
    "openpenpal-backend/internal/models"
)

func main() {
    cfg := config.LoadConfig()
    db, err := config.SetupDatabase(cfg)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // 启用 PostgreSQL 扩展
    if cfg.DatabaseType == "postgres" {
        if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
            log.Fatal("Failed to create extension:", err)
        }
    }

    // 自动迁移
    err = db.AutoMigrate(
        &models.User{},
        &models.Letter{},
        &models.LetterCode{},
        &models.Courier{},
        &models.CourierTask{},
        &models.UserCredit{},
        &models.CreditHistory{},
        &models.MuseumEntry{},
        &models.MuseumQRCode{},
        &models.AdminAction{},
    )

    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    log.Println("Database migration completed successfully!")
}
```

### Step 5: 数据迁移工具

```go
// backend/cmd/migrate-data/main.go
package main

import (
    "log"
    "gorm.io/driver/sqlite"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func migrateData(sqliteDB, postgresDB *gorm.DB) error {
    // 迁移用户数据
    var users []models.User
    sqliteDB.Find(&users)
    for _, user := range users {
        postgresDB.Create(&user)
    }

    // 迁移其他数据...
    log.Printf("迁移了 %d 个用户", len(users))
    
    return nil
}
```

## 3. 环境变量配置

```bash
# .env.production
DATABASE_TYPE=postgres
DATABASE_URL=postgresql://openpenpal:openpenpal123@localhost:5432/openpenpal
# 或分开配置
DB_HOST=localhost
DB_USER=openpenpal
DB_PASSWORD=openpenpal123
DB_NAME=openpenpal
DB_PORT=5432

# .env.development
DATABASE_TYPE=sqlite
DATABASE_URL=./openpenpal.db
```

## 4. 性能优化建议

### 添加索引
```go
type Letter struct {
    ID        string `gorm:"type:uuid;primary_key"`
    UserID    string `gorm:"type:uuid;index"`
    Status    string `gorm:"index"`
    CreatedAt time.Time `gorm:"index"`
    // 复合索引
}

// 手动创建复合索引
db.Exec("CREATE INDEX idx_letters_user_status ON letters(user_id, status)")
```

### 使用连接池
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### 查询优化
```go
// 使用预加载减少 N+1 查询
db.Preload("User").Preload("Codes").Find(&letters)

// 使用 Select 只查询需要的字段
db.Select("id", "title", "created_at").Find(&letters)

// 使用原生 SQL 处理复杂查询
db.Raw(`
    SELECT l.*, u.username, COUNT(lc.id) as code_count
    FROM letters l
    JOIN users u ON l.user_id = u.id
    LEFT JOIN letter_codes lc ON l.id = lc.letter_id
    GROUP BY l.id, u.username
`).Scan(&results)
```

## 5. 测试迁移

```bash
# 1. 备份 SQLite 数据
cp backend/openpenpal.db backend/openpenpal.db.backup

# 2. 启动 PostgreSQL
docker-compose up -d postgres

# 3. 运行迁移
go run backend/cmd/migrate/main.go

# 4. 迁移数据
go run backend/cmd/migrate-data/main.go

# 5. 测试应用
DATABASE_TYPE=postgres go run backend/main.go
```

## 6. 生产部署清单

- [ ] 设置 PostgreSQL 主从复制
- [ ] 配置自动备份
- [ ] 设置连接池参数
- [ ] 添加监控（pg_stat_statements）
- [ ] 配置 SSL 连接
- [ ] 设置适当的内存和缓存参数

## 7. 回滚方案

如果需要回滚到 SQLite：
1. 修改 `DATABASE_TYPE=sqlite`
2. 重启应用
3. 数据不会自动同步，需要手动处理

## 总结

使用 GORM + PostgreSQL 的方案可以：
- 保留所有现有代码
- 获得 PostgreSQL 的强大功能
- 保持高性能
- 降低迁移风险

这是最适合 OpenPenPal 项目当前状态的方案。