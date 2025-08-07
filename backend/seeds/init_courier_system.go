package main

import (
	"fmt"
	"log"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InitializeCourierSystem creates a complete 4-level courier hierarchy with shared data
func InitializeCourierSystem(db *gorm.DB) error {
	log.Println("Initializing complete courier system with shared data...")

	// Step 1: Ensure all courier users have courier records
	var courierUsers []models.User
	if err := db.Where("role LIKE ?", "courier%").Find(&courierUsers).Error; err != nil {
		return fmt.Errorf("failed to find courier users: %w", err)
	}

	courierMap := make(map[string]*models.Courier)
	for _, user := range courierUsers {
		var courier models.Courier
		if err := db.Where("user_id = ?", user.ID).First(&courier).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create courier record
				level := 1
				zoneCode := ""
				zoneType := ""
				managedPrefix := ""

				switch user.Username {
				case "courier_level4":
					level = 4
					zoneCode = "BEIJING"
					zoneType = "city"
					managedPrefix = "BJ"
				case "courier_level3":
					level = 3
					zoneCode = "BJDX"
					zoneType = "school"
					managedPrefix = "BJDX"
				case "courier_level2":
					level = 2
					zoneCode = "BJDX-NORTH"
					zoneType = "zone"
					managedPrefix = "BJDX5F"
				case "courier_level1":
					level = 1
					zoneCode = "BJDX-A-101"
					zoneType = "building"
					managedPrefix = "BJDX5F01"
				}

				courier = models.Courier{
					ID:                   uuid.New().String(),
					UserID:               user.ID,
					Level:                level,
					ZoneCode:             zoneCode,
					ZoneType:             zoneType,
					Status:               models.CourierStatusActive,
					ManagedOPCodePrefix:  managedPrefix,
					PerformanceScore:     95.0 + float64(level),
					CreatedAt:            time.Now(),
					UpdatedAt:            time.Now(),
				}

				if err := db.Create(&courier).Error; err != nil {
					return fmt.Errorf("failed to create courier for %s: %w", user.Username, err)
				}
				log.Printf("Created courier record for %s (Level %d)", user.Username, level)
			} else {
				return fmt.Errorf("failed to query courier: %w", err)
			}
		}
		courierMap[user.Username] = &courier
	}

	// Step 2: Establish hierarchy relationships
	if l4 := courierMap["courier_level4"]; l4 != nil {
		if l3 := courierMap["courier_level3"]; l3 != nil {
			l3.ParentID = &l4.ID
			db.Save(l3)
			
			if l2 := courierMap["courier_level2"]; l2 != nil {
				l2.ParentID = &l3.ID
				db.Save(l2)
				
				if l1 := courierMap["courier_level1"]; l1 != nil {
					l1.ParentID = &l2.ID
					db.Save(l1)
				}
			}
		}
	}
	log.Println("Established courier hierarchy relationships")

	// Step 3: Create sample letters that need delivery
	var alice models.User
	db.Where("username = ?", "alice").First(&alice)
	
	sampleLetters := []models.Letter{
		{
			ID:            uuid.New().String(),
			UserID:        alice.ID,
			Title:         "给远方朋友的新年祝福",
			Content:       "新的一年，希望你一切安好，愿我们的友谊长存...",
			RecipientType: models.RecipientSpecific,
			Status:        models.StatusPending,
			IsAnonymous:   false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New().String(),
			UserID:        alice.ID,
			Title:         "感谢信",
			Content:       "感谢你在我困难时期的帮助和支持...",
			RecipientType: models.RecipientRandom,
			Status:        models.StatusPending,
			IsAnonymous:   false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New().String(),
			UserID:        alice.ID,
			Title:         "校园回忆",
			Content:       "还记得我们一起在梧桐树下读书的日子吗...",
			RecipientType: models.RecipientSpecific,
			Status:        models.StatusPending,
			IsAnonymous:   false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, letter := range sampleLetters {
		db.Create(&letter)
	}
	log.Printf("Created %d sample letters", len(sampleLetters))

	// Step 4: Create shared courier tasks (not assigned to specific couriers)
	tasks := []models.CourierTask{
		// Standard delivery tasks (can be handled by any level)
		{
			ID:               uuid.New().String(),
			LetterID:         sampleLetters[0].ID,
			PickupLocation:   "北京大学5号楼1层",
			DeliveryLocation: "北京大学3食堂",
			PickupOPCode:     "BJDX5F01",
			DeliveryOPCode:   "BJDX3D12",
			TaskType:         models.TaskTypeStandard,
			Status:           models.TaskStatusAvailable,
			Priority:         models.PriorityMedium,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               uuid.New().String(),
			LetterID:         sampleLetters[1].ID,
			PickupLocation:   "北京大学7号楼B区",
			DeliveryLocation: "北京大学图书馆",
			PickupOPCode:     "BJDX7B05",
			DeliveryOPCode:   "BJDXTSG01",
			TaskType:         models.TaskTypeStandard,
			Status:           models.TaskStatusAvailable,
			Priority:         models.PriorityHigh,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		// Inter-school delivery (requires Level 3+)
		{
			ID:               uuid.New().String(),
			LetterID:         sampleLetters[2].ID,
			PickupLocation:   "北京大学5号楼",
			DeliveryLocation: "清华大学3号楼",
			PickupOPCode:     "BJDX5F01",
			DeliveryOPCode:   "QHUA3B02",
			TaskType:         models.TaskTypeExpress,
			Status:           models.TaskStatusAvailable,
			Priority:         models.PriorityHigh,
			RequiredLevel:    3, // Only Level 3+ can handle inter-school
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	for _, task := range tasks {
		db.Create(&task)
	}
	log.Printf("Created %d shared courier tasks", len(tasks))

	// Step 5: Assign one task to show task flow
	if l1 := courierMap["courier_level1"]; l1 != nil {
		now := time.Now()
		tasks[0].CourierID = &l1.ID
		tasks[0].Status = models.TaskStatusAccepted
		tasks[0].AcceptedAt = &now
		db.Save(&tasks[0])
		
		// Update courier stats
		l1.TotalTasks++
		db.Save(l1)
		
		log.Printf("Assigned one task to Level 1 courier for demonstration")
	}

	// Step 6: Create some scan records to show activity
	if l1 := courierMap["courier_level1"]; l1 != nil && len(tasks) > 0 {
		scanRecord := models.ScanRecord{
			ID:          uuid.New().String(),
			CourierID:   l1.ID,
			TaskID:      tasks[0].ID,
			ScanType:    models.ScanTypePickup,
			Location:    tasks[0].PickupLocation,
			OPCode:      tasks[0].PickupOPCode,
			DeviceInfo:  "iOS App v1.0",
			IsValid:     true,
			CreatedAt:   time.Now(),
		}
		db.Create(&scanRecord)
		log.Println("Created sample scan record")
	}

	// Step 7: Display final state
	log.Println("\n=== Courier System Initialized ===")
	
	// Show hierarchy
	var allCouriers []models.Courier
	db.Preload("User").Order("level DESC").Find(&allCouriers)
	
	log.Println("\nCourier Hierarchy:")
	for _, c := range allCouriers {
		parentStr := "None"
		if c.ParentID != nil {
			var parent models.Courier
			db.Preload("User").First(&parent, "id = ?", *c.ParentID)
			parentStr = parent.User.Username
		}
		log.Printf("- %s (L%d) | Zone: %s | Parent: %s | Prefix: %s",
			c.User.Username, c.Level, c.ZoneCode, parentStr, c.ManagedOPCodePrefix)
	}

	// Show task distribution
	var taskStats []struct {
		Status string
		Count  int64
	}
	db.Model(&models.CourierTask{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&taskStats)
	
	log.Println("\nTask Distribution:")
	for _, stat := range taskStats {
		log.Printf("- %s: %d tasks", stat.Status, stat.Count)
	}

	return nil
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := config.SetupDatabaseDirect(cfg)
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// Initialize courier system
	if err := InitializeCourierSystem(db); err != nil {
		log.Fatal("Failed to initialize courier system:", err)
	}

	log.Println("\nCourier system initialization complete!")
}