package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	// TODO: Switch to PostgreSQL when needed
	// "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

// TODO: Re-enable when PostgreSQL import is needed
func main() {
	fmt.Println("=== OP Codes CSV Import Tool ===")
	fmt.Println("This tool is currently disabled as we are using PostgreSQL-only setup.")
	fmt.Println("To enable this tool, update database connection to PostgreSQL.")
	return

	/*
	// TODO: 连接PostgreSQL数据库
	// dbPath := "../../openpenpal.db"
	// if _, err := os.Stat(dbPath); os.IsNotExist(err) {
	//	log.Fatalf("Database file does not exist: %s", dbPath)
	// }

	// db, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// CSV文件路径
	csvPath := "../../../../back-up-database/gd_2024_admission_with_opcodes.csv"
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		log.Fatalf("CSV file does not exist: %s", csvPath)
	}

	fmt.Printf("开始导入CSV数据到OPCode数据库...\n")
	fmt.Printf("数据库路径: %s\n", dbPath)
	fmt.Printf("CSV文件路径: %s\n", csvPath)

	// 读取CSV文件
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) < 2 {
		log.Fatalf("CSV file must have at least 2 rows (header + data)")
	}

	// 检查CSV头部格式
	header := records[0]
	fmt.Printf("CSV头部: %v\n", header)

	// 期望的格式: 院校代码,院校名称,OP_Code_前缀
	if len(header) < 3 {
		log.Fatalf("CSV格式错误，期望至少3列")
	}

	successCount := 0
	skipCount := 0
	errorCount := 0

	// 开始事务
	tx := db.Begin()

	for i, record := range records[1:] { // 跳过头部
		if len(record) < 3 {
			log.Printf("跳过第%d行，列数不足: %v", i+2, record)
			skipCount++
			continue
		}

		// 解析CSV数据
		schoolCodeStr := strings.TrimSpace(record[0])   // 院校代码
		schoolName := strings.TrimSpace(record[1])      // 院校名称
		opCodePrefix := strings.TrimSpace(record[2])    // OP_Code_前缀

		if schoolCodeStr == "" || schoolName == "" || opCodePrefix == "" {
			log.Printf("跳过第%d行，数据为空: %v", i+2, record)
			skipCount++
			continue
		}

		// 生成ID
		schoolID := fmt.Sprintf("school_%s", strings.ToLower(opCodePrefix))

		// 推断城市和省份（基于学校名称）
		city, province := inferLocationFromSchoolName(schoolName)

		// 创建OPCodeSchool记录
		school := models.OPCodeSchool{
			ID:         schoolID,
			SchoolCode: strings.ToUpper(opCodePrefix),
			SchoolName: schoolName,
			FullName:   schoolName,
			City:       city,
			Province:   province,
			IsActive:   true,
			ManagedBy:  "system",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// 插入数据（使用FirstOrCreate避免重复）
		var existingSchool models.OPCodeSchool
		result := tx.Where("school_code = ?", school.SchoolCode).FirstOrCreate(&existingSchool, school)
		if result.Error != nil {
			log.Printf("插入第%d行失败 (%s - %s): %v", i+2, opCodePrefix, schoolName, result.Error)
			errorCount++
			continue
		}

		if result.RowsAffected > 0 {
			fmt.Printf("✅ 成功导入: %s - %s (%s, %s)\n", opCodePrefix, schoolName, city, province)
			successCount++
		} else {
			fmt.Printf("⚠️  已存在: %s - %s\n", opCodePrefix, schoolName)
			skipCount++
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("事务提交失败: %v", err)
	}

	fmt.Printf("\n=== 导入完成统计 ===\n")
	fmt.Printf("成功导入: %d 条\n", successCount)
	fmt.Printf("跳过记录: %d 条\n", skipCount)
	fmt.Printf("错误记录: %d 条\n", errorCount)
	fmt.Printf("总记录数: %d 条\n", len(records)-1)

	// 验证导入结果
	var totalCount int64
	db.Model(&models.OPCodeSchool{}).Count(&totalCount)
	fmt.Printf("数据库中总学校数: %d\n", totalCount)

	// 显示一些示例数据
	var sampleSchools []models.OPCodeSchool
	db.Limit(10).Find(&sampleSchools)
	
	fmt.Printf("\n=== 示例数据 ===\n")
	for _, school := range sampleSchools {
		fmt.Printf("%s - %s (%s, %s)\n", school.SchoolCode, school.SchoolName, school.City, school.Province)
	}

	fmt.Printf("\n✅ CSV数据导入完成！\n")
}

// inferLocationFromSchoolName 根据学校名称推断城市和省份
func inferLocationFromSchoolName(schoolName string) (city, province string) {
	// 城市映射表
	cityMappings := map[string][]string{
		"北京": {"北京", "北京市"},
		"上海": {"上海", "上海市"},
		"天津": {"天津", "天津市"},
		"重庆": {"重庆", "重庆市"},
		"广州": {"广州", "广东省"},
		"深圳": {"深圳", "广东省"},
		"杭州": {"杭州", "浙江省"},
		"南京": {"南京", "江苏省"},
		"武汉": {"武汉", "湖北省"},
		"成都": {"成都", "四川省"},
		"西安": {"西安", "陕西省"},
		"长沙": {"长沙", "湖南省"},
		"沈阳": {"沈阳", "辽宁省"},
		"大连": {"大连", "辽宁省"},
		"青岛": {"青岛", "山东省"},
		"济南": {"济南", "山东省"},
		"郑州": {"郑州", "河南省"},
		"合肥": {"合肥", "安徽省"},
		"南昌": {"南昌", "江西省"},
		"福州": {"福州", "福建省"},
		"厦门": {"厦门", "福建省"},
		"昆明": {"昆明", "云南省"},
		"贵阳": {"贵阳", "贵州省"},
		"兰州": {"兰州", "甘肃省"},
		"银川": {"银川", "宁夏回族自治区"},
		"西宁": {"西宁", "青海省"},
		"乌鲁木齐": {"乌鲁木齐", "新疆维吾尔自治区"},
		"拉萨": {"拉萨", "西藏自治区"},
		"呼和浩特": {"呼和浩特", "内蒙古自治区"},
		"南宁": {"南宁", "广西壮族自治区"},
		"海口": {"海口", "海南省"},
		"石家庄": {"石家庄", "河北省"},
		"太原": {"太原", "山西省"},
		"哈尔滨": {"哈尔滨", "黑龙江省"},
		"长春": {"长春", "吉林省"},
	}

	// 省份映射表
	provinceMappings := map[string]string{
		"河北": "河北省", "山西": "山西省", "辽宁": "辽宁省", "吉林": "吉林省",
		"黑龙江": "黑龙江省", "江苏": "江苏省", "浙江": "浙江省", "安徽": "安徽省",
		"福建": "福建省", "江西": "江西省", "山东": "山东省", "河南": "河南省",
		"湖北": "湖北省", "湖南": "湖南省", "广东": "广东省", "海南": "海南省",
		"四川": "四川省", "贵州": "贵州省", "云南": "云南省", "陕西": "陕西省",
		"甘肃": "甘肃省", "青海": "青海省", "台湾": "台湾省",
	}

	// 优先匹配城市
	for cityName, location := range cityMappings {
		if strings.Contains(schoolName, cityName) {
			return location[0], location[1]
		}
	}

	// 匹配省份
	for provName, fullProvName := range provinceMappings {
		if strings.Contains(schoolName, provName) {
			return provName, fullProvName
		}
	}

	// 特殊处理：自治区
	if strings.Contains(schoolName, "广西") {
		return "南宁", "广西壮族自治区"
	}
	if strings.Contains(schoolName, "内蒙古") {
		return "呼和浩特", "内蒙古自治区"
	}
	if strings.Contains(schoolName, "新疆") {
		return "乌鲁木齐", "新疆维吾尔自治区"
	}
	if strings.Contains(schoolName, "西藏") {
		return "拉萨", "西藏自治区"
	}
	if strings.Contains(schoolName, "宁夏") {
		return "银川", "宁夏回族自治区"
	}

	// 默认值
	return "未知", "未知省份"
	*/
}