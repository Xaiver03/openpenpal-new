package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
)

// CreditLimitAdminHandler Phase 1.4: 积分限制管理界面处理器
type CreditLimitAdminHandler struct {
	limiterService *services.CreditLimiterService
}

// NewCreditLimitAdminHandler 创建管理界面处理器
func NewCreditLimitAdminHandler(limiterService *services.CreditLimiterService) *CreditLimitAdminHandler {
	return &CreditLimitAdminHandler{
		limiterService: limiterService,
	}
}

// ==================== 规则管理 ====================

// BatchCreateRules 批量创建规则
func (h *CreditLimitAdminHandler) BatchCreateRules(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req struct {
		Rules []struct {
			ActionType  string                    `json:"action_type" binding:"required"`
			LimitType   models.CreditLimitType    `json:"limit_type" binding:"required"`
			LimitPeriod models.CreditLimitPeriod  `json:"limit_period" binding:"required"`
			MaxCount    int                       `json:"max_count"`
			MaxPoints   int                       `json:"max_points"`
			Enabled     bool                      `json:"enabled"`
			Priority    int                       `json:"priority"`
			Description string                    `json:"description"`
		} `json:"rules" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdRules := make([]*models.CreditLimitRule, 0, len(req.Rules))
	errors := make([]string, 0)

	// 事务处理批量创建
	tx := h.limiterService.DB().Begin()
	for i, ruleReq := range req.Rules {
		rule := &models.CreditLimitRule{
			ID:          uuid.New().String(),
			ActionType:  ruleReq.ActionType,
			LimitType:   ruleReq.LimitType,
			LimitPeriod: ruleReq.LimitPeriod,
			MaxCount:    ruleReq.MaxCount,
			MaxPoints:   ruleReq.MaxPoints,
			Enabled:     ruleReq.Enabled,
			Priority:    ruleReq.Priority,
			Description: ruleReq.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := tx.Create(rule).Error; err != nil {
			errors = append(errors, fmt.Sprintf("规则%d创建失败: %v", i+1, err))
		} else {
			createdRules = append(createdRules, rule)
		}
	}

	if len(errors) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "批量创建失败",
			"details": errors,
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("成功创建%d条规则", len(createdRules)),
		"rules":   createdRules,
	})
}

// BatchUpdateRules 批量更新规则
func (h *CreditLimitAdminHandler) BatchUpdateRules(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req struct {
		Updates []struct {
			ID      string `json:"id" binding:"required"`
			Enabled *bool  `json:"enabled"`
			Priority int   `json:"priority"`
		} `json:"updates" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedCount := 0
	errors := make([]string, 0)

	tx := h.limiterService.DB().Begin()
	for _, update := range req.Updates {
		rule := &models.CreditLimitRule{}
		if err := tx.Where("id = ?", update.ID).First(rule).Error; err != nil {
			errors = append(errors, fmt.Sprintf("规则%s不存在", update.ID))
			continue
		}

		if update.Enabled != nil {
			rule.Enabled = *update.Enabled
		}
		if update.Priority > 0 {
			rule.Priority = update.Priority
		}
		rule.UpdatedAt = time.Now()

		if err := tx.Save(rule).Error; err != nil {
			errors = append(errors, fmt.Sprintf("规则%s更新失败: %v", update.ID, err))
		} else {
			updatedCount++
		}
	}

	if len(errors) > 0 && updatedCount == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "批量更新失败",
			"details": errors,
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("成功更新%d条规则", updatedCount),
		"updated_count": updatedCount,
		"errors":        errors,
	})
}

// ExportRules 导出规则配置
func (h *CreditLimitAdminHandler) ExportRules(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	format := c.DefaultQuery("format", "json") // json or csv

	var rules []models.CreditLimitRule
	if err := h.limiterService.DB().Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败"})
		return
	}

	switch format {
	case "csv":
		h.exportRulesAsCSV(c, rules)
	default:
		h.exportRulesAsJSON(c, rules)
	}
}

// ImportRules 导入规则配置
func (h *CreditLimitAdminHandler) ImportRules(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req struct {
		Rules     []models.CreditLimitRule `json:"rules" binding:"required"`
		Overwrite bool                     `json:"overwrite"` // 是否覆盖现有规则
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	importedCount := 0
	skippedCount := 0
	errors := make([]string, 0)

	tx := h.limiterService.DB().Begin()

	for _, rule := range req.Rules {
		// 检查是否已存在
		var existingRule models.CreditLimitRule
		exists := tx.Where("action_type = ? AND limit_type = ? AND limit_period = ?", 
			rule.ActionType, rule.LimitType, rule.LimitPeriod).First(&existingRule).Error == nil

		if exists && !req.Overwrite {
			skippedCount++
			continue
		}

		newRule := &models.CreditLimitRule{
			ID:          uuid.New().String(),
			ActionType:  rule.ActionType,
			LimitType:   rule.LimitType,
			LimitPeriod: rule.LimitPeriod,
			MaxCount:    rule.MaxCount,
			MaxPoints:   rule.MaxPoints,
			Enabled:     rule.Enabled,
			Priority:    rule.Priority,
			Description: rule.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if exists && req.Overwrite {
			// 更新现有规则
			existingRule.MaxCount = rule.MaxCount
			existingRule.MaxPoints = rule.MaxPoints
			existingRule.Enabled = rule.Enabled
			existingRule.Priority = rule.Priority
			existingRule.Description = rule.Description
			existingRule.UpdatedAt = time.Now()

			if err := tx.Save(&existingRule).Error; err != nil {
				errors = append(errors, fmt.Sprintf("更新规则失败: %v", err))
			} else {
				importedCount++
			}
		} else {
			// 创建新规则
			if err := tx.Create(newRule).Error; err != nil {
				errors = append(errors, fmt.Sprintf("创建规则失败: %v", err))
			} else {
				importedCount++
			}
		}
	}

	if len(errors) > 0 && importedCount == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "导入失败",
			"details": errors,
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":        "导入完成",
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"errors":         errors,
	})
}

// ==================== 统计报表 ====================

// GetDashboardStats 获取仪表板统计
func (h *CreditLimitAdminHandler) GetDashboardStats(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	timeRange := c.DefaultQuery("range", "24h") // 24h, 7d, 30d

	var since time.Time
	switch timeRange {
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
	default:
		since = time.Now().Add(-24 * time.Hour)
	}

	stats := make(map[string]interface{})

	// 基础统计
	var totalRules int64
	var activeRules int64
	var totalRiskUsers int64
	var blockedUsers int64

	h.limiterService.DB().Model(&models.CreditLimitRule{}).Count(&totalRules)
	h.limiterService.DB().Model(&models.CreditLimitRule{}).Where("enabled = ?", true).Count(&activeRules)
	h.limiterService.DB().Model(&models.CreditRiskUser{}).Count(&totalRiskUsers)
	h.limiterService.DB().Model(&models.CreditRiskUser{}).Where("risk_level = ?", models.RiskLevelBlocked).Count(&blockedUsers)

	stats["basic"] = map[string]interface{}{
		"total_rules":     totalRules,
		"active_rules":    activeRules,
		"total_risk_users": totalRiskUsers,
		"blocked_users":   blockedUsers,
	}

	// 时间范围内的活动统计
	var actions []struct {
		ActionType string
		Count      int64
		TotalPoints int64
	}

	h.limiterService.DB().Model(&models.UserCreditAction{}).
		Select("action_type, COUNT(*) as count, COALESCE(SUM(points), 0) as total_points").
		Where("created_at >= ?", since).
		Group("action_type").
		Order("count DESC").
		Scan(&actions)

	stats["actions"] = actions

	// 风险用户分布
	var riskDistribution []struct {
		RiskLevel string
		Count     int64
	}

	h.limiterService.DB().Model(&models.CreditRiskUser{}).
		Select("risk_level, COUNT(*) as count").
		Group("risk_level").
		Scan(&riskDistribution)

	stats["risk_distribution"] = riskDistribution

	// 检测日志统计
	var detectionStats struct {
		TotalDetections    int64
		AnomalousCount    int64
		HighRiskCount     int64
	}

	h.limiterService.DB().Model(&models.FraudDetectionLog{}).
		Where("created_at >= ?", since).
		Count(&detectionStats.TotalDetections)

	h.limiterService.DB().Model(&models.FraudDetectionLog{}).
		Where("created_at >= ? AND is_anomalous = ?", since, true).
		Count(&detectionStats.AnomalousCount)

	h.limiterService.DB().Model(&models.FraudDetectionLog{}).
		Where("created_at >= ? AND risk_score >= ?", since, 0.8).
		Count(&detectionStats.HighRiskCount)

	stats["detection"] = detectionStats

	// 每日趋势（最近7天）
	var dailyTrends []struct {
		Date            string
		ActionCount     int64
		DetectionCount  int64
		AnomalousCount  int64
	}

	for i := 6; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i)
		dayStart := day.Truncate(24 * time.Hour)
		dayEnd := dayStart.Add(24 * time.Hour)

		var actionCount, detectionCount, anomalousCount int64

		h.limiterService.DB().Model(&models.UserCreditAction{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&actionCount)

		h.limiterService.DB().Model(&models.FraudDetectionLog{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&detectionCount)

		h.limiterService.DB().Model(&models.FraudDetectionLog{}).
			Where("created_at >= ? AND created_at < ? AND is_anomalous = ?", dayStart, dayEnd, true).
			Count(&anomalousCount)

		dailyTrends = append(dailyTrends, struct {
			Date            string
			ActionCount     int64
			DetectionCount  int64
			AnomalousCount  int64
		}{
			Date:            day.Format("2006-01-02"),
			ActionCount:     actionCount,
			DetectionCount:  detectionCount,
			AnomalousCount:  anomalousCount,
		})
	}

	stats["daily_trends"] = dailyTrends

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetLimitUsageReport 获取限制使用报告
func (h *CreditLimitAdminHandler) GetLimitUsageReport(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	actionType := c.Query("action_type")
	period := c.DefaultQuery("period", "daily") // daily, weekly, monthly

	var periodStart time.Time
	switch period {
	case "weekly":
		periodStart = time.Now().AddDate(0, 0, -7)
	case "monthly":
		periodStart = time.Now().AddDate(0, -1, 0)
	default:
		periodStart = time.Now().Truncate(24 * time.Hour)
	}

	query := h.limiterService.DB().Model(&models.UserCreditAction{}).
		Where("created_at >= ?", periodStart)

	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}

	// 按用户统计使用情况
	var userUsage []struct {
		UserID      string
		ActionCount int64
		TotalPoints int64
		UniqueIPs   int64
		FirstAction time.Time
		LastAction  time.Time
	}

	query.Select(`
		user_id,
		COUNT(*) as action_count,
		COALESCE(SUM(points), 0) as total_points,
		COUNT(DISTINCT ip_address) as unique_ips,
		MIN(created_at) as first_action,
		MAX(created_at) as last_action
	`).
		Group("user_id").
		Order("action_count DESC").
		Limit(100).
		Scan(&userUsage)

	// 获取对应的限制规则
	var rules []models.CreditLimitRule
	rulesQuery := h.limiterService.DB().Where("enabled = ?", true)
	if actionType != "" {
		rulesQuery = rulesQuery.Where("action_type = ?", actionType)
	}
	rulesQuery.Find(&rules)

	c.JSON(http.StatusOK, gin.H{
		"period":     period,
		"start_time": periodStart,
		"usage":      userUsage,
		"rules":      rules,
	})
}

// GetFraudDetectionReport 获取防作弊检测报告
func (h *CreditLimitAdminHandler) GetFraudDetectionReport(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	timeRange := c.DefaultQuery("range", "7d")
	userID := c.Query("user_id")

	var since time.Time
	switch timeRange {
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
	default:
		since = time.Now().AddDate(0, 0, -7)
	}

	query := h.limiterService.DB().Model(&models.FraudDetectionLog{}).
		Where("created_at >= ?", since)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 检测模式统计
	var patternStats []struct {
		Pattern string
		Count   int64
	}

	// 由于detected_patterns是JSON字符串，需要特殊处理
	var logs []models.FraudDetectionLog
	query.Find(&logs)

	patternMap := make(map[string]int64)
	totalLogs := int64(len(logs))
	anomalousLogs := int64(0)
	highRiskLogs := int64(0)

	for _, log := range logs {
		if log.IsAnomalous {
			anomalousLogs++
		}
		if log.RiskScore >= 0.8 {
			highRiskLogs++
		}

		// 解析检测到的模式
		var patterns []string
		if log.DetectedPatterns != "" {
			json.Unmarshal([]byte(log.DetectedPatterns), &patterns)
			for _, pattern := range patterns {
				patternMap[pattern]++
			}
		}
	}

	for pattern, count := range patternMap {
		patternStats = append(patternStats, struct {
			Pattern string
			Count   int64
		}{Pattern: pattern, Count: count})
	}

	// 最近的高风险检测
	var recentHighRisk []models.FraudDetectionLog
	h.limiterService.DB().Where("created_at >= ? AND risk_score >= ?", since, 0.8).
		Order("created_at DESC").
		Limit(20).
		Find(&recentHighRisk)

	c.JSON(http.StatusOK, gin.H{
		"time_range":       timeRange,
		"total_logs":       totalLogs,
		"anomalous_logs":   anomalousLogs,
		"high_risk_logs":   highRiskLogs,
		"pattern_stats":    patternStats,
		"recent_high_risk": recentHighRisk,
	})
}

// ==================== 实时监控 ====================

// GetRealTimeAlerts 获取实时告警
func (h *CreditLimitAdminHandler) GetRealTimeAlerts(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	severity := c.Query("severity") // high, medium, low
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// 获取最近的检测日志作为告警
	query := h.limiterService.DB().Model(&models.FraudDetectionLog{}).
		Where("created_at >= ?", time.Now().Add(-time.Hour*24))

	if severity != "" {
		switch severity {
		case "high":
			query = query.Where("risk_score >= ?", 0.8)
		case "medium":
			query = query.Where("risk_score >= ? AND risk_score < ?", 0.4, 0.8)
		case "low":
			query = query.Where("risk_score < ?", 0.4)
		}
	}

	var alerts []struct {
		models.FraudDetectionLog
		UserInfo struct {
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user_info"`
	}

	query.Order("created_at DESC").Limit(limit).Find(&alerts)

	// 增加用户信息（简化版，实际应该join用户表）
	for i := range alerts {
		// 这里应该查询用户表获取用户名和邮箱
		alerts[i].UserInfo.Username = fmt.Sprintf("user_%s", alerts[i].UserID[:8])
		alerts[i].UserInfo.Email = fmt.Sprintf("user_%s@example.com", alerts[i].UserID[:8])
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetSystemHealth 获取系统健康状态
func (h *CreditLimitAdminHandler) GetSystemHealth(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	health := make(map[string]interface{})

	// 数据库连接状态
	if sqlDB, err := h.limiterService.DB().DB(); err == nil {
		if err := sqlDB.Ping(); err == nil {
			health["database"] = "healthy"
		} else {
			health["database"] = "unhealthy"
		}
	} else {
		health["database"] = "error"
	}

	// 检查最近的检测活动
	var recentDetections int64
	h.limiterService.DB().Model(&models.FraudDetectionLog{}).
		Where("created_at >= ?", time.Now().Add(-time.Hour)).
		Count(&recentDetections)

	health["detection_activity"] = map[string]interface{}{
		"recent_count": recentDetections,
		"status":       "active",
	}

	// 检查是否有大量高风险用户
	var highRiskUsers int64
	h.limiterService.DB().Model(&models.CreditRiskUser{}).
		Where("risk_level IN (?)", []string{string(models.RiskLevelHigh), string(models.RiskLevelBlocked)}).
		Count(&highRiskUsers)

	var totalUsers int64
	h.limiterService.DB().Model(&models.CreditRiskUser{}).Count(&totalUsers)

	riskRatio := float64(0)
	if totalUsers > 0 {
		riskRatio = float64(highRiskUsers) / float64(totalUsers)
	}

	health["risk_status"] = map[string]interface{}{
		"high_risk_users": highRiskUsers,
		"total_users":     totalUsers,
		"risk_ratio":      riskRatio,
		"status":          h.getRiskStatus(riskRatio),
	}

	// 检查规则配置状态
	var enabledRules int64
	h.limiterService.DB().Model(&models.CreditLimitRule{}).
		Where("enabled = ?", true).
		Count(&enabledRules)

	health["rules_status"] = map[string]interface{}{
		"enabled_rules": enabledRules,
		"status":        h.getRulesStatus(enabledRules),
	}

	health["overall_status"] = h.getOverallStatus(health)
	health["timestamp"] = time.Now()

	c.JSON(http.StatusOK, gin.H{"health": health})
}

// ==================== 高级搜索和筛选 ====================

// AdvancedSearch 高级搜索
func (h *CreditLimitAdminHandler) AdvancedSearch(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req struct {
		SearchType string `json:"search_type"` // users, actions, detections, rules
		Filters    struct {
			UserID       string    `json:"user_id"`
			ActionType   string    `json:"action_type"`
			RiskLevel    string    `json:"risk_level"`
			IPAddress    string    `json:"ip_address"`
			DeviceID     string    `json:"device_id"`
			DateFrom     time.Time `json:"date_from"`
			DateTo       time.Time `json:"date_to"`
			MinRiskScore float64   `json:"min_risk_score"`
			MaxRiskScore float64   `json:"max_risk_score"`
			IsAnomalous  *bool     `json:"is_anomalous"`
		} `json:"filters"`
		Pagination struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
		} `json:"pagination"`
		Sort struct {
			Field string `json:"field"`
			Order string `json:"order"` // asc, desc
		} `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Pagination.Page == 0 {
		req.Pagination.Page = 1
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = 20
	}

	switch req.SearchType {
	case "users":
		h.searchUsers(c, req)
	case "actions":
		h.searchActions(c, req)
	case "detections":
		h.searchDetections(c, req)
	case "rules":
		h.searchRules(c, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search_type"})
	}
}

// ==================== 私有方法 ====================

func (h *CreditLimitAdminHandler) isAdmin(c *gin.Context) bool {
	userRole := c.GetString("user_role")
	return userRole == "super_admin" || userRole == "admin" || userRole == "platform_admin"
}

func (h *CreditLimitAdminHandler) exportRulesAsJSON(c *gin.Context, rules []models.CreditLimitRule) {
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"credit_rules_%s.json\"", time.Now().Format("20060102_150405")))
	c.JSON(http.StatusOK, gin.H{"rules": rules, "exported_at": time.Now()})
}

func (h *CreditLimitAdminHandler) exportRulesAsCSV(c *gin.Context, rules []models.CreditLimitRule) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"credit_rules_%s.csv\"", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入标题行
	header := []string{"ID", "ActionType", "LimitType", "LimitPeriod", "MaxCount", "MaxPoints", "Enabled", "Priority", "Description", "CreatedAt"}
	writer.Write(header)

	// 写入数据行
	for _, rule := range rules {
		record := []string{
			rule.ID,
			rule.ActionType,
			string(rule.LimitType),
			string(rule.LimitPeriod),
			strconv.Itoa(rule.MaxCount),
			strconv.Itoa(rule.MaxPoints),
			strconv.FormatBool(rule.Enabled),
			strconv.Itoa(rule.Priority),
			rule.Description,
			rule.CreatedAt.Format(time.RFC3339),
		}
		writer.Write(record)
	}
}

func (h *CreditLimitAdminHandler) getRiskStatus(ratio float64) string {
	if ratio > 0.1 {
		return "critical"
	} else if ratio > 0.05 {
		return "warning"
	}
	return "normal"
}

func (h *CreditLimitAdminHandler) getRulesStatus(count int64) string {
	if count == 0 {
		return "error"
	} else if count < 5 {
		return "warning"
	}
	return "normal"
}

func (h *CreditLimitAdminHandler) getOverallStatus(health map[string]interface{}) string {
	// 简化的整体状态计算
	dbStatus := health["database"].(string)
	riskStatus := health["risk_status"].(map[string]interface{})["status"].(string)
	rulesStatus := health["rules_status"].(map[string]interface{})["status"].(string)

	if dbStatus == "unhealthy" || riskStatus == "critical" || rulesStatus == "error" {
		return "critical"
	} else if riskStatus == "warning" || rulesStatus == "warning" {
		return "warning"
	}
	return "healthy"
}

// searchUsers 搜索用户
func (h *CreditLimitAdminHandler) searchUsers(c *gin.Context, req interface{}) {
	// 实现用户搜索逻辑
	c.JSON(http.StatusOK, gin.H{"message": "User search not implemented yet"})
}

// searchActions 搜索行为记录
func (h *CreditLimitAdminHandler) searchActions(c *gin.Context, req interface{}) {
	// 实现行为记录搜索逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Action search not implemented yet"})
}

// searchDetections 搜索检测记录
func (h *CreditLimitAdminHandler) searchDetections(c *gin.Context, req interface{}) {
	// 实现检测记录搜索逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Detection search not implemented yet"})
}

// searchRules 搜索规则
func (h *CreditLimitAdminHandler) searchRules(c *gin.Context, req interface{}) {
	// 实现规则搜索逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Rule search not implemented yet"})
}