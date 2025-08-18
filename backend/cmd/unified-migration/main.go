package main

import (
	"flag"
	"fmt"
	"log"

	"openpenpal-backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 统一迁移工具 - 命令行界面
func main() {
	fmt.Println("🎯 OpenPenPal Unified Migration Tool")
	fmt.Println("=====================================")

	// 解析命令行参数
	var (
		dryRun       = flag.Bool("dry-run", false, "Run migration in dry-run mode (no actual changes)")
		strategy     = flag.String("strategy", "unified", "Migration strategy: unified, safe, extended")
		rollback     = flag.Bool("rollback", false, "Rollback mode")
		// configPath   = flag.String("config", "", "Path to configuration file") // TODO: implement config file support
		dbHost       = flag.String("host", "localhost", "Database host")
		dbPort       = flag.String("port", "5432", "Database port")
		dbUser       = flag.String("user", "rocalight", "Database user")
		dbPassword   = flag.String("password", "password", "Database password")
		dbName       = flag.String("database", "openpenpal", "Database name")
		sslMode      = flag.String("ssl", "disable", "SSL mode")
		sslCert      = flag.String("ssl-cert", "", "SSL certificate file")
		sslKey       = flag.String("ssl-key", "", "SSL private key file")
		sslRootCert  = flag.String("ssl-root-cert", "", "SSL root certificate file")
		verbose      = flag.Bool("verbose", false, "Verbose output")
		skipOptim    = flag.Bool("skip-optimization", false, "Skip performance optimizations")
		coordinated  = flag.Bool("coordinated", true, "Use coordinated migration across services")
	)
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// 创建数据库配置
	dbConfig := &config.Config{
		DatabaseType:   "postgres",
		DBHost:         *dbHost,
		DBPort:         *dbPort,
		DBUser:         *dbUser,
		DBPassword:     *dbPassword,
		DatabaseName:   *dbName,
		DBSSLMode:      *sslMode,
		DBSSLCert:      *sslCert,
		DBSSLKey:       *sslKey,
		DBSSLRootCert:  *sslRootCert,
	}

	// 连接数据库
	db, err := connectDatabase(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Printf("✅ Connected to PostgreSQL database: %s@%s:%s/%s", 
		*dbUser, *dbHost, *dbPort, *dbName)

	// 执行迁移策略
	switch *strategy {
	case "unified":
		err = executeUnifiedMigration(db, dbConfig, *dryRun, *rollback, *skipOptim)
	case "coordinated":
		err = executeCoordinatedMigration(db, dbConfig, *dryRun)
	case "safe":
		err = executeSafeMigration(db, dbConfig)
	case "extended":
		err = executeExtendedMigration(db, dbConfig)
	default:
		log.Fatalf("❌ Unknown migration strategy: %s", *strategy)
	}

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// 如果使用协调迁移，执行协调器
	if *coordinated && *strategy == "unified" {
		log.Println("🔄 Starting coordinated migration...")
		if err := executeCoordinatedMigration(db, dbConfig, *dryRun); err != nil {
			log.Printf("⚠️  Warning: Coordinated migration failed: %v", err)
		}
	}

	fmt.Println("🎉 Migration completed successfully!")
	fmt.Println("📋 Next steps:")
	fmt.Println("  1. Verify database connection in your application")
	fmt.Println("  2. Run application tests")
	fmt.Println("  3. Monitor performance metrics")
	fmt.Println("  4. Schedule regular maintenance tasks")
}

// connectDatabase 连接数据库
func connectDatabase(config *config.Config) (*gorm.DB, error) {
	// 构建基础DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		config.DBHost, config.DBUser, config.DBPassword, config.DatabaseName, config.DBPort, config.DBSSLMode)
	
	// 添加SSL证书参数
	if config.DBSSLMode != "disable" && config.DBSSLMode != "allow" {
		if config.DBSSLRootCert != "" {
			dsn += fmt.Sprintf(" sslrootcert=%s", config.DBSSLRootCert)
		}
		if config.DBSSLCert != "" {
			dsn += fmt.Sprintf(" sslcert=%s", config.DBSSLCert)
		}
		if config.DBSSLKey != "" {
			dsn += fmt.Sprintf(" sslkey=%s", config.DBSSLKey)
		}
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// executeUnifiedMigration 执行统一迁移策略
func executeUnifiedMigration(db *gorm.DB, dbConfig *config.Config, dryRun, rollback, skipOptim bool) error {
	log.Println("🚀 Executing unified migration strategy...")

	opts := &config.MigrationOptions{
		DryRun:              dryRun,
		RollbackMode:        rollback,
		SkipOptimizations:   skipOptim,
		ConcurrentIndexes:   true,
		Timeout:             30 * 60, // 30 minutes
		FailureStrategy:     "stop",
		BackupBeforeMigrate: true,
	}

	strategy := config.NewMigrationStrategy(db, dbConfig, opts)
	return strategy.ExecuteUnifiedMigration()
}

// executeCoordinatedMigration 执行协调迁移
func executeCoordinatedMigration(db *gorm.DB, dbConfig *config.Config, dryRun bool) error {
	log.Println("🎯 Executing coordinated migration across services...")

	coordinator := config.NewMigrationCoordinator(db, dbConfig)
	return coordinator.ExecuteCoordinatedMigration()
}

// executeSafeMigration 执行安全迁移
func executeSafeMigration(db *gorm.DB, dbConfig *config.Config) error {
	log.Println("🛡️  Executing safe migration...")

	// 使用现有的安全迁移逻辑
	allModels := config.GetAllModels()
	return config.SafeAutoMigrate(db, allModels...)
}

// executeExtendedMigration 执行扩展迁移
func executeExtendedMigration(db *gorm.DB, dbConfig *config.Config) error {
	log.Println("🔧 Executing extended models migration...")

	// 先执行核心迁移
	if err := executeSafeMigration(db, dbConfig); err != nil {
		return fmt.Errorf("core migration failed: %w", err)
	}

	// 然后执行扩展迁移
	return config.MigrateExtendedModels(db)
}

// showUsage 显示使用说明
func showUsage() {
	fmt.Println(`
OpenPenPal Unified Migration Tool
================================

USAGE:
  go run main.go [OPTIONS]

OPTIONS:
  --strategy STRING     Migration strategy (unified, coordinated, safe, extended) [default: unified]
  --dry-run            Run in dry-run mode (no actual changes)
  --rollback           Enable rollback mode
  --host STRING        Database host [default: localhost]
  --port STRING        Database port [default: 5432]
  --user STRING        Database user [default: rocalight]
  --password STRING    Database password [default: password]
  --database STRING    Database name [default: openpenpal]
  --ssl STRING         SSL mode [default: disable]
  --verbose            Enable verbose output
  --skip-optimization  Skip performance optimizations
  --coordinated        Use coordinated migration [default: true]

EXAMPLES:
  # Run unified migration with all features
  go run main.go --strategy=unified

  # Dry run to see what would be executed
  go run main.go --dry-run

  # Safe migration only (core models)
  go run main.go --strategy=safe

  # Extended migration (core + extended models)
  go run main.go --strategy=extended

  # Coordinated migration across all services
  go run main.go --strategy=coordinated

  # Connect to specific database
  go run main.go --host=prod-db.example.com --user=admin --database=openpenpal_prod

MIGRATION STRATEGIES:
  unified     - Complete migration with all optimizations (recommended)
  coordinated - Cross-service coordination with dependency management
  safe        - Core models only using SafeAutoMigrate
  extended    - Core + extended models migration

DATABASE REQUIREMENTS:
  - PostgreSQL 12+ (recommended: PostgreSQL 14+)
  - User with CREATE, ALTER, DROP privileges
  - Sufficient disk space for backup and indexes
  - Extensions: pg_stat_statements (optional)

NOTES:
  - Always backup your database before running migrations
  - Use --dry-run first to preview changes
  - Monitor logs for warnings and errors
  - Run ANALYZE after migration for optimal performance
`)
}