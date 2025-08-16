// Package performance provides AI-powered index recommendations
package performance

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// IndexAdvisor provides AI-powered index recommendations
type IndexAdvisor struct {
	queryPatterns     map[string]*QueryPattern
	tableAnalysis     map[string]*TableAnalysis
	indexRecommendations map[string][]*core.IndexRecommendation
	mu                sync.RWMutex
}

// TableAnalysis contains analysis data for a table
type TableAnalysis struct {
	TableName       string
	ColumnUsage     map[string]*ColumnUsage
	QueryPatterns   []string
	JoinPatterns    []string
	FilterPatterns  []string
	LastAnalyzed    string
}

// ColumnUsage tracks how columns are used in queries
type ColumnUsage struct {
	ColumnName      string
	SelectCount     int
	WhereCount      int
	JoinCount       int
	OrderByCount    int
	GroupByCount    int
	Selectivity     float64
}

// NewIndexAdvisor creates a new index advisor
func NewIndexAdvisor() *IndexAdvisor {
	return &IndexAdvisor{
		queryPatterns:        make(map[string]*QueryPattern),
		tableAnalysis:        make(map[string]*TableAnalysis),
		indexRecommendations: make(map[string][]*core.IndexRecommendation),
	}
}

// AnalyzeQuery analyzes a query and returns index recommendations
func (ia *IndexAdvisor) AnalyzeQuery(query string) []*core.IndexRecommendation {
	recommendations := make([]*core.IndexRecommendation, 0)
	
	// Parse the query to extract table and column information
	queryInfo := ia.parseQuery(query)
	
	// Update table analysis
	for _, table := range queryInfo.Tables {
		ia.updateTableAnalysis(table, queryInfo)
	}
	
	// Generate recommendations based on query patterns
	for _, table := range queryInfo.Tables {
		tableRecommendations := ia.generateRecommendationsForTable(table, queryInfo)
		recommendations = append(recommendations, tableRecommendations...)
	}
	
	return recommendations
}

// GetRecommendationsForTable returns index recommendations for a specific table
func (ia *IndexAdvisor) GetRecommendationsForTable(tableName string) []*core.IndexRecommendation {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	
	recommendations, exists := ia.indexRecommendations[tableName]
	if !exists {
		return make([]*core.IndexRecommendation, 0)
	}
	
	return recommendations
}

// AnalyzeTableUsage analyzes how tables are used across all queries
func (ia *IndexAdvisor) AnalyzeTableUsage() map[string]*TableAnalysis {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*TableAnalysis)
	for k, v := range ia.tableAnalysis {
		result[k] = v
	}
	
	return result
}

// Private methods

func (ia *IndexAdvisor) parseQuery(query string) *QueryInfo {
	queryInfo := &QueryInfo{
		Query:        query,
		Tables:       make([]string, 0),
		Columns:      make(map[string][]string),
		WhereColumns: make(map[string][]string),
		JoinColumns:  make(map[string][]string),
		OrderColumns: make(map[string][]string),
		GroupColumns: make(map[string][]string),
	}
	
	queryLower := strings.ToLower(query)
	
	// Extract table names
	queryInfo.Tables = ia.extractTableNames(queryLower)
	
	// Extract column usage patterns
	for _, table := range queryInfo.Tables {
		queryInfo.Columns[table] = ia.extractColumnsForTable(queryLower, table)
		queryInfo.WhereColumns[table] = ia.extractWhereColumns(queryLower, table)
		queryInfo.JoinColumns[table] = ia.extractJoinColumns(queryLower, table)
		queryInfo.OrderColumns[table] = ia.extractOrderColumns(queryLower, table)
		queryInfo.GroupColumns[table] = ia.extractGroupColumns(queryLower, table)
	}
	
	return queryInfo
}

func (ia *IndexAdvisor) extractTableNames(query string) []string {
	tables := make([]string, 0)
	
	// Simple regex to extract table names from FROM and JOIN clauses
	fromRegex := regexp.MustCompile(`from\s+(\w+)`)
	joinRegex := regexp.MustCompile(`join\s+(\w+)`)
	
	fromMatches := fromRegex.FindAllStringSubmatch(query, -1)
	for _, match := range fromMatches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}
	
	joinMatches := joinRegex.FindAllStringSubmatch(query, -1)
	for _, match := range joinMatches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}
	
	return ia.removeDuplicates(tables)
}

func (ia *IndexAdvisor) extractColumnsForTable(query, table string) []string {
	columns := make([]string, 0)
	
	// Simple extraction - in practice, you'd use a proper SQL parser
	selectRegex := regexp.MustCompile(`select\s+(.+?)\s+from`)
	matches := selectRegex.FindStringSubmatch(query)
	
	if len(matches) > 1 {
		columnList := matches[1]
		if columnList != "*" {
			parts := strings.Split(columnList, ",")
			for _, part := range parts {
				column := strings.TrimSpace(part)
				// Remove table prefix if present
				if strings.Contains(column, ".") {
					parts := strings.Split(column, ".")
					if len(parts) > 1 && parts[0] == table {
						columns = append(columns, parts[1])
					}
				} else {
					columns = append(columns, column)
				}
			}
		}
	}
	
	return columns
}

func (ia *IndexAdvisor) extractWhereColumns(query, table string) []string {
	columns := make([]string, 0)
	
	// Extract columns from WHERE clause
	whereRegex := regexp.MustCompile(`where\s+(.+?)(?:\s+order\s+by|\s+group\s+by|\s+limit|$)`)
	matches := whereRegex.FindStringSubmatch(query)
	
	if len(matches) > 1 {
		whereClause := matches[1]
		columnRegex := regexp.MustCompile(`(\w+)\s*[=<>!]`)
		columnMatches := columnRegex.FindAllStringSubmatch(whereClause, -1)
		
		for _, match := range columnMatches {
			if len(match) > 1 {
				columns = append(columns, match[1])
			}
		}
	}
	
	return ia.removeDuplicates(columns)
}

func (ia *IndexAdvisor) extractJoinColumns(query, table string) []string {
	columns := make([]string, 0)
	
	// Extract columns from JOIN conditions
	joinRegex := regexp.MustCompile(`join\s+\w+\s+on\s+(.+?)(?:\s+where|\s+order|\s+group|$)`)
	matches := joinRegex.FindAllStringSubmatch(query, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			joinCondition := match[1]
			columnRegex := regexp.MustCompile(`(\w+)\s*=\s*(\w+)`)
			columnMatches := columnRegex.FindAllStringSubmatch(joinCondition, -1)
			
			for _, columnMatch := range columnMatches {
				if len(columnMatch) > 2 {
					columns = append(columns, columnMatch[1], columnMatch[2])
				}
			}
		}
	}
	
	return ia.removeDuplicates(columns)
}

func (ia *IndexAdvisor) extractOrderColumns(query, table string) []string {
	columns := make([]string, 0)
	
	// Extract columns from ORDER BY clause
	orderRegex := regexp.MustCompile(`order\s+by\s+(.+?)(?:\s+limit|$)`)
	matches := orderRegex.FindStringSubmatch(query)
	
	if len(matches) > 1 {
		orderClause := matches[1]
		parts := strings.Split(orderClause, ",")
		
		for _, part := range parts {
			column := strings.TrimSpace(part)
			// Remove ASC/DESC
			column = regexp.MustCompile(`\s+(asc|desc)$`).ReplaceAllString(column, "")
			columns = append(columns, column)
		}
	}
	
	return ia.removeDuplicates(columns)
}

func (ia *IndexAdvisor) extractGroupColumns(query, table string) []string {
	columns := make([]string, 0)
	
	// Extract columns from GROUP BY clause
	groupRegex := regexp.MustCompile(`group\s+by\s+(.+?)(?:\s+order|\s+limit|$)`)
	matches := groupRegex.FindStringSubmatch(query)
	
	if len(matches) > 1 {
		groupClause := matches[1]
		parts := strings.Split(groupClause, ",")
		
		for _, part := range parts {
			column := strings.TrimSpace(part)
			columns = append(columns, column)
		}
	}
	
	return ia.removeDuplicates(columns)
}

func (ia *IndexAdvisor) updateTableAnalysis(table string, queryInfo *QueryInfo) {
	ia.mu.Lock()
	defer ia.mu.Unlock()
	
	analysis, exists := ia.tableAnalysis[table]
	if !exists {
		analysis = &TableAnalysis{
			TableName:   table,
			ColumnUsage: make(map[string]*ColumnUsage),
		}
		ia.tableAnalysis[table] = analysis
	}
	
	// Update column usage statistics
	for _, column := range queryInfo.Columns[table] {
		usage := ia.getOrCreateColumnUsage(analysis, column)
		usage.SelectCount++
	}
	
	for _, column := range queryInfo.WhereColumns[table] {
		usage := ia.getOrCreateColumnUsage(analysis, column)
		usage.WhereCount++
	}
	
	for _, column := range queryInfo.JoinColumns[table] {
		usage := ia.getOrCreateColumnUsage(analysis, column)
		usage.JoinCount++
	}
	
	for _, column := range queryInfo.OrderColumns[table] {
		usage := ia.getOrCreateColumnUsage(analysis, column)
		usage.OrderByCount++
	}
	
	for _, column := range queryInfo.GroupColumns[table] {
		usage := ia.getOrCreateColumnUsage(analysis, column)
		usage.GroupByCount++
	}
}

func (ia *IndexAdvisor) getOrCreateColumnUsage(analysis *TableAnalysis, column string) *ColumnUsage {
	usage, exists := analysis.ColumnUsage[column]
	if !exists {
		usage = &ColumnUsage{
			ColumnName:   column,
			Selectivity:  0.5, // Default selectivity
		}
		analysis.ColumnUsage[column] = usage
	}
	return usage
}

func (ia *IndexAdvisor) generateRecommendationsForTable(table string, queryInfo *QueryInfo) []*core.IndexRecommendation {
	recommendations := make([]*core.IndexRecommendation, 0)
	
	// Recommend indexes for frequently used WHERE columns
	for _, column := range queryInfo.WhereColumns[table] {
		rec := &core.IndexRecommendation{
			Table:            table,
			Columns:          []string{column},
			Type:             "btree",
			EstimatedBenefit: 25.0,
			Reason:           fmt.Sprintf("Column '%s' is frequently used in WHERE clauses", column),
			CreateStatement:  fmt.Sprintf("CREATE INDEX idx_%s_%s ON %s (%s);", table, column, table, column),
		}
		recommendations = append(recommendations, rec)
	}
	
	// Recommend indexes for JOIN columns
	for _, column := range queryInfo.JoinColumns[table] {
		rec := &core.IndexRecommendation{
			Table:            table,
			Columns:          []string{column},
			Type:             "btree",
			EstimatedBenefit: 30.0,
			Reason:           fmt.Sprintf("Column '%s' is used in JOIN operations", column),
			CreateStatement:  fmt.Sprintf("CREATE INDEX idx_%s_%s_join ON %s (%s);", table, column, table, column),
		}
		recommendations = append(recommendations, rec)
	}
	
	// Recommend composite indexes for ORDER BY columns
	if len(queryInfo.OrderColumns[table]) > 1 {
		columns := queryInfo.OrderColumns[table]
		rec := &core.IndexRecommendation{
			Table:            table,
			Columns:          columns,
			Type:             "btree",
			EstimatedBenefit: 20.0,
			Reason:           "Composite index for ORDER BY optimization",
			CreateStatement:  fmt.Sprintf("CREATE INDEX idx_%s_order ON %s (%s);", 
				table, table, strings.Join(columns, ", ")),
		}
		recommendations = append(recommendations, rec)
	}
	
	// Store recommendations
	ia.mu.Lock()
	ia.indexRecommendations[table] = recommendations
	ia.mu.Unlock()
	
	return recommendations
}

func (ia *IndexAdvisor) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := make([]string, 0)
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// Data structures

// QueryInfo contains parsed information about a query
type QueryInfo struct {
	Query        string
	Tables       []string
	Columns      map[string][]string
	WhereColumns map[string][]string
	JoinColumns  map[string][]string
	OrderColumns map[string][]string
	GroupColumns map[string][]string
}