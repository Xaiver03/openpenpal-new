package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("🚀 Starting OP Code School Data Import Tool...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// 建立数据库连接
	db, err := connectDatabase(cfg)
	if err != nil {
		log.Fatal("❌ Failed to connect database:", err)
	}

	// 确保数据库表结构正确
	if err := ensureDatabaseSchema(db); err != nil {
		log.Fatal("❌ Failed to ensure database schema:", err)
	}

	// 确定CSV文件路径
	csvFile := "/Users/rocalight/同步空间/opplc/back-up-database/gd_2024_admission_with_opcodes.csv"
	if len(os.Args) > 1 {
		csvFile = os.Args[1]
	}

	// 验证CSV文件存在
	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		log.Fatalf("❌ CSV file not found: %s", csvFile)
	}

	log.Printf("📄 Using CSV file: %s", csvFile)

	// 执行完整的数据导入流程
	if err := performCompleteImport(db, csvFile); err != nil {
		log.Fatal("❌ Data import failed:", err)
	}

	// 验证导入结果
	if err := validateImportResults(db); err != nil {
		log.Fatal("❌ Import validation failed:", err)
	}

	log.Println("✅ OP Code school data import completed successfully!")
}

// connectDatabase 建立数据库连接（使用完整的连接逻辑）
func connectDatabase(cfg *config.Config) (*gorm.DB, error) {
	log.Println("🔌 Connecting to database...")

	// 构建PostgreSQL DSN
	var dsn string
	if cfg.DatabaseURL != "" && cfg.DatabaseURL != "./openpenpal.db" && strings.HasPrefix(cfg.DatabaseURL, "postgres") {
		dsn = cfg.DatabaseURL
	} else {
		// 构建标准DSN
		host := cfg.DBHost
		if host == "" {
			host = "localhost"
		}
		
		port := cfg.DBPort
		if port == "" {
			port = "5432"
		}
		
		sslMode := cfg.DBSSLMode
		if sslMode == "" {
			sslMode = "disable"
		}

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, cfg.DBUser, cfg.DBPassword, cfg.DatabaseName, sslMode)
	}

	log.Printf("🔗 PostgreSQL DSN: %s", maskPassword(dsn))

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// 建立连接
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("✅ Database connection established")
	return db, nil
}

// ensureDatabaseSchema 确保数据库表结构正确
func ensureDatabaseSchema(db *gorm.DB) error {
	log.Println("🔧 Ensuring database schema...")

	// 检查表是否存在
	if !db.Migrator().HasTable(&models.OPCodeSchool{}) {
		log.Println("📋 Creating OPCodeSchool table...")
		if err := db.Migrator().CreateTable(&models.OPCodeSchool{}); err != nil {
			return fmt.Errorf("failed to create OPCodeSchool table: %w", err)
		}
		log.Println("✅ OPCodeSchool table created")
	} else {
		log.Println("✅ OPCodeSchool table already exists")
	}

	// 确保索引存在
	if err := ensureIndexes(db); err != nil {
		return fmt.Errorf("failed to ensure indexes: %w", err)
	}

	return nil
}

// ensureIndexes 确保必要的索引存在
func ensureIndexes(db *gorm.DB) error {
	log.Println("📊 Ensuring database indexes...")

	indexes := []struct {
		name    string
		table   string
		columns []string
	}{
		{"idx_op_code_schools_school_code", "op_code_schools", []string{"school_code"}},
		{"idx_op_code_schools_city", "op_code_schools", []string{"city"}},
		{"idx_op_code_schools_province", "op_code_schools", []string{"province"}},
		{"idx_op_code_schools_name", "op_code_schools", []string{"school_name"}},
		{"idx_op_code_schools_active", "op_code_schools", []string{"is_active"}},
	}

	for _, idx := range indexes {
		// GORM的HasIndex方法可能不够可靠，我们直接尝试创建索引
		sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
			idx.name, idx.table, strings.Join(idx.columns, ", "))
		
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️ Warning: Failed to create index %s: %v", idx.name, err)
			// 不要因为索引失败而停止整个流程
		} else {
			log.Printf("✅ Index %s ensured", idx.name)
		}
	}

	return nil
}

// performCompleteImport 执行完整的数据导入流程
func performCompleteImport(db *gorm.DB, csvFile string) error {
	log.Println("📥 Starting complete data import...")

	// 读取并解析CSV文件
	records, err := readCSVFile(csvFile)
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	log.Printf("📊 Processing %d records from CSV file", len(records))

	// 开始事务
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ Transaction rolled back due to panic: %v", r)
		}
	}()

	// 批量处理数据
	batchSize := 100
	successCount := 0
	failCount := 0

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		schools := make([]models.OPCodeSchool, 0, len(batch))

		// 处理当前批次
		for j, record := range batch {
			school, err := processRecord(record, i+j+2) // +2 因为CSV有标题行，行号从1开始
			if err != nil {
				log.Printf("⚠️ Skipping record %d: %v", i+j+2, err)
				failCount++
				continue
			}
			schools = append(schools, *school)
		}

		// 批量插入
		if len(schools) > 0 {
			result := tx.CreateInBatches(schools, 50)
			if result.Error != nil {
				// 如果批量插入失败，尝试单个插入
				log.Printf("⚠️ Batch insert failed, trying individual inserts: %v", result.Error)
				for _, school := range schools {
					if err := tx.Create(&school).Error; err != nil {
						log.Printf("⚠️ Failed to insert school %s: %v", school.SchoolName, err)
						failCount++
					} else {
						successCount++
					}
				}
			} else {
				successCount += len(schools)
			}
		}

		// 进度报告
		if (i+batchSize)%500 == 0 || end == len(records) {
			log.Printf("📈 Progress: %d/%d records processed (%d success, %d failed)",
				end, len(records), successCount, failCount)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Import completed: %d successful, %d failed", successCount, failCount)
	return nil
}

// readCSVFile 读取CSV文件
func readCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// 跳过标题行
	return records[1:], nil
}

// processRecord 处理单条CSV记录
func processRecord(record []string, rowNumber int) (*models.OPCodeSchool, error) {
	if len(record) < 3 {
		return nil, fmt.Errorf("insufficient columns (expected 3, got %d)", len(record))
	}

	institutionCode := strings.TrimSpace(record[0])
	schoolName := strings.TrimSpace(record[1])
	opcodePrefix := strings.TrimSpace(record[2])

	// 验证必填字段
	if institutionCode == "" || schoolName == "" || opcodePrefix == "" {
		return nil, fmt.Errorf("empty required fields")
	}

	// 处理OP Code前缀
	opcodePrefix = normalizeOPCodePrefix(opcodePrefix)
	if len(opcodePrefix) != 2 {
		return nil, fmt.Errorf("invalid OP Code prefix length: %s", opcodePrefix)
	}

	// 推断地理信息
	cityInfo := inferGeographicInfo(schoolName)
	
	// 生成唯一ID
	schoolID := generateUUID()

	// 创建学校记录
	school := &models.OPCodeSchool{
		ID:         schoolID,
		SchoolCode: opcodePrefix,
		SchoolName: schoolName,
		FullName:   schoolName, // 可以后续优化为完整名称
		City:       cityInfo.City,
		Province:   cityInfo.Province,
		IsActive:   true,
		ManagedBy:  "system_import",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	return school, nil
}

// GeographicInfo 地理信息结构
type GeographicInfo struct {
	City     string
	Province string
}

// normalizeOPCodePrefix 标准化OP Code前缀
func normalizeOPCodePrefix(prefix string) string {
	prefix = strings.TrimSpace(strings.ToUpper(prefix))
	
	// 如果是单字符，前面补0
	if len(prefix) == 1 {
		prefix = "0" + prefix
	}
	
	// 如果超过2位，截取前2位
	if len(prefix) > 2 {
		prefix = prefix[:2]
	}
	
	return prefix
}

// inferGeographicInfo 推断地理信息
func inferGeographicInfo(schoolName string) GeographicInfo {
	// 扩展的城市映射表
	cityMappings := map[string]GeographicInfo{
		// 直辖市
		"北京": {"北京", "北京"},
		"上海": {"上海", "上海"},
		"天津": {"天津", "天津"},
		"重庆": {"重庆", "重庆"},
		
		// 省会城市和重要城市
		"广州": {"广州", "广东"},
		"深圳": {"深圳", "广东"},
		"东莞": {"东莞", "广东"},
		"佛山": {"佛山", "广东"},
		"珠海": {"珠海", "广东"},
		"汕头": {"汕头", "广东"},
		"湛江": {"湛江", "广东"},
		"肇庆": {"肇庆", "广东"},
		"惠州": {"惠州", "广东"},
		"韶关": {"韶关", "广东"},
		
		"杭州": {"杭州", "浙江"},
		"宁波": {"宁波", "浙江"},
		"温州": {"温州", "浙江"},
		"嘉兴": {"嘉兴", "浙江"},
		"湖州": {"湖州", "浙江"},
		"绍兴": {"绍兴", "浙江"},
		"金华": {"金华", "浙江"},
		"台州": {"台州", "浙江"},
		
		"南京": {"南京", "江苏"},
		"苏州": {"苏州", "江苏"},
		"无锡": {"无锡", "江苏"},
		"常州": {"常州", "江苏"},
		"镇江": {"镇江", "江苏"},
		"南通": {"南通", "江苏"},
		"泰州": {"泰州", "江苏"},
		"扬州": {"扬州", "江苏"},
		"盐城": {"盐城", "江苏"},
		"连云港": {"连云港", "江苏"},
		"淮安": {"淮安", "江苏"},
		"宿迁": {"宿迁", "江苏"},
		"徐州": {"徐州", "江苏"},
		
		"武汉": {"武汉", "湖北"},
		"宜昌": {"宜昌", "湖北"},
		"襄阳": {"襄阳", "湖北"},
		"荆州": {"荆州", "湖北"},
		"黄冈": {"黄冈", "湖北"},
		"十堰": {"十堰", "湖北"},
		"孝感": {"孝感", "湖北"},
		"荆门": {"荆门", "湖北"},
		"鄂州": {"鄂州", "湖北"},
		"黄石": {"黄石", "湖北"},
		"咸宁": {"咸宁", "湖北"},
		
		"成都": {"成都", "四川"},
		"绵阳": {"绵阳", "四川"},
		"德阳": {"德阳", "四川"},
		"南充": {"南充", "四川"},
		"宜宾": {"宜宾", "四川"},
		"自贡": {"自贡", "四川"},
		"乐山": {"乐山", "四川"},
		"泸州": {"泸州", "四川"},
		"达州": {"达州", "四川"},
		"内江": {"内江", "四川"},
		"遂宁": {"遂宁", "四川"},
		"广安": {"广安", "四川"},
		"眉山": {"眉山", "四川"},
		"资阳": {"资阳", "四川"},
		
		"西安": {"西安", "陕西"},
		"宝鸡": {"宝鸡", "陕西"},
		"咸阳": {"咸阳", "陕西"},
		"渭南": {"渭南", "陕西"},
		"延安": {"延安", "陕西"},
		"汉中": {"汉中", "陕西"},
		"榆林": {"榆林", "陕西"},
		"安康": {"安康", "陕西"},
		"商洛": {"商洛", "陕西"},
		"铜川": {"铜川", "陕西"},
		
		"长沙": {"长沙", "湖南"},
		"株洲": {"株洲", "湖南"},
		"湘潭": {"湘潭", "湖南"},
		"衡阳": {"衡阳", "湖南"},
		"邵阳": {"邵阳", "湖南"},
		"岳阳": {"岳阳", "湖南"},
		"常德": {"常德", "湖南"},
		"张家界": {"张家界", "湖南"},
		"益阳": {"益阳", "湖南"},
		"郴州": {"郴州", "湖南"},
		"永州": {"永州", "湖南"},
		"怀化": {"怀化", "湖南"},
		"娄底": {"娄底", "湖南"},
		
		"济南": {"济南", "山东"},
		"青岛": {"青岛", "山东"},
		"淄博": {"淄博", "山东"},
		"枣庄": {"枣庄", "山东"},
		"东营": {"东营", "山东"},
		"烟台": {"烟台", "山东"},
		"潍坊": {"潍坊", "山东"},
		"济宁": {"济宁", "山东"},
		"泰安": {"泰安", "山东"},
		"威海": {"威海", "山东"},
		"日照": {"日照", "山东"},
		"莱芜": {"莱芜", "山东"},
		"临沂": {"临沂", "山东"},
		"德州": {"德州", "山东"},
		"聊城": {"聊城", "山东"},
		"滨州": {"滨州", "山东"},
		"菏泽": {"菏泽", "山东"},
		
		// 其他省份省会
		"哈尔滨": {"哈尔滨", "黑龙江"},
		"长春": {"长春", "吉林"},
		"沈阳": {"沈阳", "辽宁"},
		"大连": {"大连", "辽宁"},
		"石家庄": {"石家庄", "河北"},
		"太原": {"太原", "山西"},
		"呼和浩特": {"呼和浩特", "内蒙古"},
		"银川": {"银川", "宁夏"},
		"西宁": {"西宁", "青海"},
		"乌鲁木齐": {"乌鲁木齐", "新疆"},
		"拉萨": {"拉萨", "西藏"},
		"昆明": {"昆明", "云南"},
		"贵阳": {"贵阳", "贵州"},
		"南宁": {"南宁", "广西"},
		"海口": {"海口", "海南"},
		"福州": {"福州", "福建"},
		"厦门": {"厦门", "福建"},
		"南昌": {"南昌", "江西"},
		"合肥": {"合肥", "安徽"},
		"郑州": {"郑州", "河南"},
		"兰州": {"兰州", "甘肃"},
	}

	// 按城市名长度排序，优先匹配较长的城市名
	for cityName, info := range cityMappings {
		if strings.Contains(schoolName, cityName) {
			return info
		}
	}

	// 如果没有匹配到，尝试从其他关键词推断
	if strings.Contains(schoolName, "理工") {
		return GeographicInfo{"未知", "未知"}
	}

	// 默认值
	return GeographicInfo{"未知", "未知"}
}

// validateImportResults 验证导入结果
func validateImportResults(db *gorm.DB) error {
	log.Println("🔍 Validating import results...")

	var totalCount int64
	if err := db.Model(&models.OPCodeSchool{}).Count(&totalCount).Error; err != nil {
		return fmt.Errorf("failed to count total records: %w", err)
	}

	var activeCount int64
	if err := db.Model(&models.OPCodeSchool{}).Where("is_active = ?", true).Count(&activeCount).Error; err != nil {
		return fmt.Errorf("failed to count active records: %w", err)
	}

	// 检查是否有重复的SchoolCode
	var duplicateCount int64
	if err := db.Model(&models.OPCodeSchool{}).
		Select("school_code").
		Group("school_code").
		Having("COUNT(*) > 1").
		Count(&duplicateCount).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to check duplicates: %v", err)
	}

	log.Printf("📊 Import validation results:")
	log.Printf("   Total records: %d", totalCount)
	log.Printf("   Active records: %d", activeCount)
	log.Printf("   Duplicate school codes: %d", duplicateCount)

	if totalCount == 0 {
		return fmt.Errorf("no records were imported")
	}

	if duplicateCount > 0 {
		log.Printf("⚠️ Warning: Found %d duplicate school codes", duplicateCount)
	}

	// 样本数据检查
	var sampleSchools []models.OPCodeSchool
	if err := db.Limit(3).Find(&sampleSchools).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to fetch sample data: %v", err)
	} else {
		log.Println("📋 Sample records:")
		for i, school := range sampleSchools {
			log.Printf("   %d. %s (%s) - %s, %s", 
				i+1, school.SchoolName, school.SchoolCode, school.City, school.Province)
		}
	}

	log.Println("✅ Import validation completed")
	return nil
}

// maskPassword 掩码密码信息
func maskPassword(dsn string) string {
	// 简单的密码掩码，避免在日志中暴露敏感信息
	if strings.Contains(dsn, "password=") {
		parts := strings.Split(dsn, " ")
		for i, part := range parts {
			if strings.HasPrefix(part, "password=") {
				parts[i] = "password=***"
				break
			}
		}
		return strings.Join(parts, " ")
	}
	return dsn
}

// generateUUID 生成UUID
func generateUUID() string {
	return uuid.New().String()
}