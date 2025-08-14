package config

import (
	"gorm.io/gorm"
	"log"
)

// CreateOptimizedIndexes 创建优化的复合索引 - SOTA性能优化
func CreateOptimizedIndexes(db *gorm.DB) error {
	log.Println("Creating optimized composite indexes...")

	// 索引创建SQL语句集合
	indexes := []string{
		// 1. Letter表性能优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_user_status_created ON letters(user_id, status, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_recipient_opcode_status ON letters(recipient_op_code, status) WHERE recipient_op_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_sender_opcode_created ON letters(sender_op_code, created_at DESC) WHERE sender_op_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_type_visibility_created ON letters(type, visibility, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_reply_to_created ON letters(reply_to, created_at DESC) WHERE reply_to != ''",

		// 2. LetterCode表优化索引 - FSD条码系统
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_codes_status_updated ON letter_codes(status, updated_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_codes_recipient_code_status ON letter_codes(recipient_code, status) WHERE recipient_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_codes_envelope_status ON letter_codes(envelope_id, status) WHERE envelope_id != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_codes_scanned_by_time ON letter_codes(last_scanned_by, last_scanned_at DESC) WHERE last_scanned_by != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_codes_bound_delivered ON letter_codes(bound_at, delivered_at) WHERE bound_at IS NOT NULL",

		// 3. ScanEvent表优化索引 - PRD扫描历史系统
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_barcode_timestamp ON scan_events(barcode_id, timestamp DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_scanner_type_time ON scan_events(scanned_by, scan_type, timestamp DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_opcode_type_time ON scan_events(op_code, scan_type, timestamp DESC) WHERE op_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_status_transition ON scan_events(old_status, new_status, timestamp DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_location_time ON scan_events(location, timestamp DESC) WHERE location != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_lettercode_timestamp ON scan_events(letter_code_id, timestamp DESC)",

		// 4. Courier表优化索引 - 信使系统
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_couriers_level_zone_active ON couriers(level, zone_code, deleted_at) WHERE deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_couriers_managed_prefix_level ON couriers(managed_op_code_prefix, level) WHERE managed_op_code_prefix != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_couriers_parent_level ON couriers(parent_id, level) WHERE parent_id IS NOT NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_couriers_created_by_time ON couriers(created_by_id, created_at DESC) WHERE created_by_id IS NOT NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_couriers_status_level ON couriers(status, level, deleted_at) WHERE deleted_at IS NULL",

		// 5. CourierTask表优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_courier_status_priority ON courier_tasks(courier_id, status, priority, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_pickup_delivery_opcode ON courier_tasks(pickup_op_code, delivery_op_code) WHERE pickup_op_code != '' AND delivery_op_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_current_opcode_status ON courier_tasks(current_op_code, status) WHERE current_op_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_deadline_status ON courier_tasks(deadline, status) WHERE deadline IS NOT NULL",

		// 6. LevelUpgradeRequest表优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_level_upgrade_requests_courier_status ON level_upgrade_requests(courier_id, status, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_level_upgrade_requests_reviewer_reviewed ON level_upgrade_requests(reviewed_by, reviewed_at DESC) WHERE reviewed_by != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_level_upgrade_requests_level_transition ON level_upgrade_requests(current_level, request_level, status)",

		// 7. OP Code系统优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_codes_school_area_point ON op_codes(school_code, area_code, point_code)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_codes_type_active_public ON op_codes(point_type, is_active, is_public)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_codes_managed_by_active ON op_codes(managed_by, is_active) WHERE managed_by != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_codes_binding_status ON op_codes(binding_type, binding_status) WHERE binding_type != ''",

		// 8. OPCodeApplication表优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_code_applications_user_status ON op_code_applications(user_id, status, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_code_applications_school_area_status ON op_code_applications(school_code, area_code, status)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_code_applications_reviewer_reviewed ON op_code_applications(reviewer_id, reviewed_at DESC) WHERE reviewer_id IS NOT NULL",

		// 9. OPCodeSchool和OPCodeArea表优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_code_schools_code_active ON op_code_schools(school_code, is_active)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_code_schools_managed_by ON op_code_schools(managed_by) WHERE managed_by != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_code_areas_school_code_active ON op_code_areas(school_code, area_code, is_active)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_op_code_areas_managed_by ON op_code_areas(managed_by) WHERE managed_by != ''",

		// 10. User表额外优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_role_school_active ON users(role, school_code, is_active) WHERE is_active = true",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created_role ON users(created_at DESC, role) WHERE deleted_at IS NULL",

		// 11. StatusLog表优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_status_logs_letter_status_created ON status_logs(letter_id, status, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_status_logs_updated_by_created ON status_logs(updated_by, created_at DESC) WHERE updated_by != ''",

		// 12. 扩展用户档案表索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_profiles_extended_opcode ON user_profiles_extended(op_code) WHERE op_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_profiles_extended_levels ON user_profiles_extended(writing_level, courier_level)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_stats_points_streak ON user_stats(total_points DESC, current_streak DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_stats_active_date ON user_stats(last_active_date DESC) WHERE last_active_date IS NOT NULL",

		// 13. 用户成就表索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_achievements_user_category ON user_achievements(user_id, category, unlocked_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_achievements_code_unlocked ON user_achievements(code, unlocked_at DESC)",

		// 14. 信件互动表索引（如果存在）
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_likes_user_created ON letter_likes(user_id, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_likes_letter_created ON letter_likes(letter_id, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_shares_platform_created ON letter_shares(platform, created_at DESC)",

		// 15. 高级复合索引 - 复杂查询优化
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_complex_search ON letters(user_id, status, visibility, created_at DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_geo_routing ON courier_tasks(pickup_op_code, delivery_op_code, status, priority) WHERE pickup_op_code != '' AND delivery_op_code != ''",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_analytics ON scan_events(scan_type, new_status, DATE(timestamp)) WHERE timestamp >= CURRENT_DATE - INTERVAL '30 days'",

		// 16. 部分索引 - 针对活跃数据
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_active_drafts ON letters(user_id, updated_at DESC) WHERE status = 'draft' AND deleted_at IS NULL",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_in_transit ON letters(recipient_op_code, updated_at DESC) WHERE status IN ('collected', 'in_transit')",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_couriers_active_hierarchy ON couriers(parent_id, level, created_at DESC) WHERE status = 'approved' AND deleted_at IS NULL",

		// 17. 时间范围查询优化
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_recent_activity ON scan_events(timestamp DESC, scan_type) WHERE timestamp >= CURRENT_DATE - INTERVAL '7 days'",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letter_codes_recent_scans ON letter_codes(last_scanned_at DESC, status) WHERE last_scanned_at >= CURRENT_DATE - INTERVAL '7 days'",

		// 18. 统计查询优化索引
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_events_hourly_stats ON scan_events(DATE_TRUNC('hour', timestamp), scan_type) WHERE timestamp >= CURRENT_DATE - INTERVAL '7 days'",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_daily_stats ON letters(DATE(created_at), status) WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'",
	}

	// 逐个执行索引创建
	for i, indexSQL := range indexes {
		log.Printf("Creating index %d/%d...", i+1, len(indexes))
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index %d: %v", i+1, err)
			// 继续执行其他索引，不因为单个索引失败而中断
		}
	}

	log.Println("Optimized composite indexes creation completed")
	return nil
}

// CreatePerformanceViews 创建性能优化的物化视图
func CreatePerformanceViews(db *gorm.DB) error {
	log.Println("Creating performance views...")

	views := []string{
		// 信使性能统计视图
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_courier_performance AS
		SELECT 
			c.id as courier_id,
			c.level,
			c.zone_code,
			c.managed_op_code_prefix,
			COUNT(ct.id) as total_tasks,
			COUNT(ct.id) FILTER (WHERE ct.status = 'completed') as completed_tasks,
			COUNT(ct.id) FILTER (WHERE ct.status = 'pending') as pending_tasks,
			AVG(EXTRACT(EPOCH FROM (ct.completed_at - ct.created_at))/3600) FILTER (WHERE ct.status = 'completed') as avg_completion_hours,
			SUM(ct.reward) FILTER (WHERE ct.status = 'completed') as total_rewards,
			MAX(ct.updated_at) as last_activity
		FROM couriers c
		LEFT JOIN courier_tasks ct ON c.id = ct.courier_id
		WHERE c.deleted_at IS NULL
		GROUP BY c.id, c.level, c.zone_code, c.managed_op_code_prefix`,

		// OP Code活跃度统计视图
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_opcode_activity AS
		SELECT 
			oc.code,
			oc.school_code,
			oc.area_code,
			oc.point_code,
			oc.point_type,
			COUNT(DISTINCT se.id) as total_scans,
			COUNT(DISTINCT se.scanned_by) as unique_scanners,
			COUNT(DISTINCT DATE(se.timestamp)) as active_days,
			MAX(se.timestamp) as last_scan_time,
			COUNT(DISTINCT lc.id) as associated_letters
		FROM op_codes oc
		LEFT JOIN scan_events se ON oc.code = se.op_code
		LEFT JOIN letter_codes lc ON oc.code = lc.recipient_code
		WHERE oc.is_active = true
		GROUP BY oc.code, oc.school_code, oc.area_code, oc.point_code, oc.point_type`,

		// 条码状态流转统计视图
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_barcode_status_flow AS
		SELECT 
			se.old_status,
			se.new_status,
			se.scan_type,
			COUNT(*) as transition_count,
			AVG(EXTRACT(EPOCH FROM (se.timestamp - prev_se.timestamp))/3600) as avg_transition_hours
		FROM scan_events se
		LEFT JOIN scan_events prev_se ON se.barcode_id = prev_se.barcode_id 
			AND prev_se.timestamp < se.timestamp
			AND prev_se.new_status = se.old_status
		WHERE se.timestamp >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY se.old_status, se.new_status, se.scan_type`,
	}

	// 创建物化视图
	for i, viewSQL := range views {
		log.Printf("Creating view %d/%d...", i+1, len(views))
		if err := db.Exec(viewSQL).Error; err != nil {
			log.Printf("Warning: Failed to create view %d: %v", i+1, err)
		}
	}

	// 为物化视图创建索引
	viewIndexes := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_courier_performance_id ON mv_courier_performance(courier_id)",
		"CREATE INDEX IF NOT EXISTS idx_mv_courier_performance_level ON mv_courier_performance(level, completed_tasks DESC)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_opcode_activity_code ON mv_opcode_activity(code)",
		"CREATE INDEX IF NOT EXISTS idx_mv_opcode_activity_school ON mv_opcode_activity(school_code, total_scans DESC)",
		"CREATE INDEX IF NOT EXISTS idx_mv_barcode_status_flow_transition ON mv_barcode_status_flow(old_status, new_status)",
	}

	for i, indexSQL := range viewIndexes {
		log.Printf("Creating view index %d/%d...", i+1, len(viewIndexes))
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create view index %d: %v", i+1, err)
		}
	}

	log.Println("Performance views creation completed")
	return nil
}

// RefreshMaterializedViews 刷新物化视图
func RefreshMaterializedViews(db *gorm.DB) error {
	log.Println("Refreshing materialized views...")

	views := []string{
		"REFRESH MATERIALIZED VIEW CONCURRENTLY mv_courier_performance",
		"REFRESH MATERIALIZED VIEW CONCURRENTLY mv_opcode_activity",
		"REFRESH MATERIALIZED VIEW CONCURRENTLY mv_barcode_status_flow",
		"REFRESH MATERIALIZED VIEW CONCURRENTLY mv_user_stats",
		"REFRESH MATERIALIZED VIEW CONCURRENTLY mv_courier_stats",
	}

	for i, refreshSQL := range views {
		log.Printf("Refreshing view %d/%d...", i+1, len(views))
		if err := db.Exec(refreshSQL).Error; err != nil {
			log.Printf("Warning: Failed to refresh view %d: %v", i+1, err)
		}
	}

	log.Println("Materialized views refresh completed")
	return nil
}
