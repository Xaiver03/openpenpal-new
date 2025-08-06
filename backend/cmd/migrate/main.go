package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"openpenpal-backend/internal/config"
)

func main() {
	var (
		optimize = flag.Bool("optimize", false, "Run SOTA database optimizations")
		refresh  = flag.Bool("refresh", false, "Refresh materialized views only")
		analyze  = flag.Bool("analyze", false, "Run performance analysis only")
		full     = flag.Bool("full", false, "Run full optimization suite")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := config.SetupDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create migration service
	migrationService := config.NewMigrationService(db, cfg)

	// Execute based on flags
	switch {
	case *full:
		fmt.Println("🎯 Running full SOTA optimization suite...")
		if err := migrationService.RunFullOptimization(); err != nil {
			log.Fatalf("Full optimization failed: %v", err)
		}

	case *optimize:
		fmt.Println("🚀 Running database optimizations...")
		if err := migrationService.RunOptimizations(); err != nil {
			log.Fatalf("Optimization failed: %v", err)
		}

	case *refresh:
		fmt.Println("🔄 Refreshing materialized views...")
		if err := migrationService.RefreshMaterializedViews(); err != nil {
			log.Fatalf("Refresh failed: %v", err)
		}

	case *analyze:
		fmt.Println("📊 Running performance analysis...")
		if err := migrationService.AnalyzePerformance(); err != nil {
			log.Fatalf("Analysis failed: %v", err)
		}

	default:
		fmt.Println("OpenPenPal Database Migration Tool")
		fmt.Println("Usage:")
		fmt.Println("  -full      Run complete SOTA optimization suite")
		fmt.Println("  -optimize  Run database optimizations only")
		fmt.Println("  -refresh   Refresh materialized views only")
		fmt.Println("  -analyze   Run performance analysis only")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/migrate/main.go -full")
		fmt.Println("  go run cmd/migrate/main.go -optimize")
		fmt.Println("  go run cmd/migrate/main.go -refresh")
		os.Exit(1)
	}

	fmt.Println("✅ Migration completed successfully!")
}
