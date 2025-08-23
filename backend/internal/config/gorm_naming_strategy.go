package config

import (
	"fmt"
	"strings"

	"gorm.io/gorm/schema"
)

// CustomNamingStrategy 自定义命名策略，保持与现有数据库约束名称一致
type CustomNamingStrategy struct {
	schema.NamingStrategy
}

// IndexName 生成索引名称
func (ns CustomNamingStrategy) IndexName(table, column string) string {
	// 对于users表的特殊处理
	if table == "users" {
		switch column {
		case "username":
			return "unique_username"
		case "email":
			return "unique_email"
		}
	}
	// 默认使用GORM的命名策略
	return ns.NamingStrategy.IndexName(table, column)
}

// UniqueIndexName 生成唯一索引名称
func (ns CustomNamingStrategy) UniqueIndexName(table, column string) string {
	// 对于users表的特殊处理
	if table == "users" {
		switch column {
		case "username":
			return "unique_username"
		case "email":
			return "unique_email"
		}
	}
	// 默认使用GORM的命名策略
	return fmt.Sprintf("uni_%s_%s", table, column)
}

// ConstraintName 生成约束名称
func (ns CustomNamingStrategy) ConstraintName(table, name string) string {
	// 对于users表的特殊约束处理
	if table == "users" {
		if strings.Contains(name, "username") {
			return "unique_username"
		}
		if strings.Contains(name, "email") {
			return "unique_email"
		}
	}
	// 默认约束名称格式
	return fmt.Sprintf("fk_%s_%s", table, name)
}

// CheckerName 生成检查约束名称
func (ns CustomNamingStrategy) CheckerName(table, column string) string {
	return fmt.Sprintf("chk_%s_%s", table, column)
}

// RelationshipFKName 生成外键约束名称
func (ns CustomNamingStrategy) RelationshipFKName(relationship schema.Relationship) string {
	return ns.NamingStrategy.RelationshipFKName(relationship)
}

// JoinTableName 生成连接表名称
func (ns CustomNamingStrategy) JoinTableName(str string) string {
	return ns.NamingStrategy.JoinTableName(str)
}