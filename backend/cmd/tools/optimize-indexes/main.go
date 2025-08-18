package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/migrations"
)

func main() {
	// 命令行参数
	var (
		mode      = flag.String("mode", "analyze", "Mode: analyze, create, drop, report")
		dryRun    = flag.Bool("dry-run", false, "Dry run mode (no changes)")
		verbose   = flag.Bool("verbose", false, "Verbose output")
		tableName = flag.String("table", "", "Specific table to optimize")
		indexName = flag.String("index", "", "Specific index name")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := config.SetupDatabaseDirect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建索引优化器
	optimizer := config.NewIndexOptimizer(db, *dryRun, *verbose)

	switch *mode {
	case "analyze":
		// 分析当前索引状态
		analyzeIndexes(optimizer)

	case "create":
		// 创建优化索引
		if *tableName != "" {
			createTableIndexes(optimizer, *tableName)
		} else {
			if err := optimizer.OptimizeAll(); err != nil {
				log.Fatalf("Optimization failed: %v", err)
			}
		}

	case "drop":
		// 删除指定索引
		if *indexName == "" {
			log.Fatal("Index name required for drop operation")
		}
		if err := optimizer.DropIndex(*indexName, true); err != nil {
			log.Fatalf("Failed to drop index: %v", err)
		}

	case "report":
		// 生成详细报告
		generateDetailedReport(optimizer)

	case "migrate":
		// 运行迁移
		if err := migrations.RegisterIndexMigration(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

// analyzeIndexes 分析当前索引
func analyzeIndexes(optimizer *config.IndexOptimizer) {
	fmt.Println("=== Index Analysis Report ===\n")

	// 获取索引统计
	stats, err := optimizer.GetIndexStats()
	if err != nil {
		log.Printf("Failed to get index stats: %v", err)
		return
	}

	fmt.Println("Current Indexes:")
	fmt.Println("Table Name | Index Name | Size | Scans | Tuples Read")
	fmt.Println("-----------|------------|------|-------|------------")
	
	for _, stat := range stats {
		fmt.Printf("%-10s | %-30s | %-8s | %-6d | %d\n",
			truncate(stat.TableName, 10),
			truncate(stat.IndexName, 30),
			stat.IndexSize,
			stat.ScanCount,
			stat.TuplesRead,
		)
	}

	// 获取未使用的索引
	fmt.Println("\n\nUnused Indexes (candidates for removal):")
	unusedIndexes, err := optimizer.GetUnusedIndexes(30)
	if err != nil {
		log.Printf("Failed to get unused indexes: %v", err)
		return
	}

	if len(unusedIndexes) == 0 {
		fmt.Println("No unused indexes found")
	} else {
		for _, idx := range unusedIndexes {
			fmt.Printf("- %s\n", idx)
		}
	}
}

// createTableIndexes 为特定表创建索引
func createTableIndexes(optimizer *config.IndexOptimizer, tableName string) {
	indexes := optimizer.GetCriticalIndexes()
	created := 0
	
	for _, idx := range indexes {
		if idx.TableName == tableName {
			if err := optimizer.CreateIndex(idx); err != nil {
				log.Printf("Failed to create index %s: %v", idx.IndexName, err)
			} else {
				created++
			}
		}
	}
	
	fmt.Printf("Created %d indexes for table %s\n", created, tableName)
}

// generateDetailedReport 生成详细报告
func generateDetailedReport(optimizer *config.IndexOptimizer) {
	fmt.Println("# PostgreSQL Index Optimization Report\n")
	
	// 获取所有推荐的索引
	criticalIndexes := optimizer.GetCriticalIndexes()
	textIndexes := optimizer.GetTextSearchIndexes()
	
	fmt.Println("## Recommended Indexes\n")
	fmt.Println("### Critical Performance Indexes")
	for _, idx := range criticalIndexes {
		fmt.Printf("- **%s** on `%s`\n", idx.IndexName, idx.TableName)
		fmt.Printf("  - Columns: %v\n", idx.Columns)
		if idx.Partial != "" {
			fmt.Printf("  - Condition: `%s`\n", idx.Partial)
		}
		if len(idx.Include) > 0 {
			fmt.Printf("  - Covering: %v\n", idx.Include)
		}
		fmt.Printf("  - Purpose: %s\n", idx.Comment)
		fmt.Println()
	}
	
	fmt.Println("\n### Full-Text Search Indexes")
	for _, idx := range textIndexes {
		fmt.Printf("- **%s** on `%s`\n", idx.IndexName, idx.TableName)
		fmt.Printf("  - Type: %s\n", idx.Type)
		fmt.Printf("  - Purpose: %s\n", idx.Comment)
		fmt.Println()
	}
	
	// 获取当前统计
	stats, err := optimizer.GetIndexStats()
	if err == nil {
		fmt.Println("\n## Current Index Usage")
		fmt.Println("\n### Most Used Indexes")
		count := 0
		for _, stat := range stats {
			if stat.ScanCount > 0 && count < 10 {
				fmt.Printf("- %s.%s: %d scans, %s\n", 
					stat.TableName, stat.IndexName, stat.ScanCount, stat.IndexSize)
				count++
			}
		}
	}
	
	fmt.Println("\n## Implementation Commands\n")
	fmt.Println("```bash")
	fmt.Println("# Dry run to preview changes")
	fmt.Println("go run cmd/tools/optimize-indexes/main.go --mode=create --dry-run --verbose")
	fmt.Println()
	fmt.Println("# Apply all optimizations")
	fmt.Println("go run cmd/tools/optimize-indexes/main.go --mode=create --verbose")
	fmt.Println()
	fmt.Println("# Optimize specific table")
	fmt.Println("go run cmd/tools/optimize-indexes/main.go --mode=create --table=letters")
	fmt.Println("```")
}

// truncate 截断字符串
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}