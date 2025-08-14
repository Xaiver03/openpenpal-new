package logging

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String 返回日志级别字符串
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 结构化日志器接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
	WithContext(ctx context.Context) Logger
	WithRequestID(requestID string) Logger
	WithUserID(userID string) Logger
}

// StructuredLogger 结构化日志器实现
type StructuredLogger struct {
	level     LogLevel
	output    io.Writer
	baseLog   *log.Logger
	fields    map[string]interface{}
	service   string
	version   string
	requestID string
	userID    string
}

// LoggerConfig 日志器配置
type LoggerConfig struct {
	Level        LogLevel
	Output       io.Writer
	Service      string
	Version      string
	EnableCaller bool
	TimeFormat   string
}

// NewLogger 创建新的结构化日志器
func NewLogger(config LoggerConfig) *StructuredLogger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.TimeFormat == "" {
		config.TimeFormat = time.RFC3339
	}

	return &StructuredLogger{
		level:   config.Level,
		output:  config.Output,
		baseLog: log.New(config.Output, "", 0),
		fields:  make(map[string]interface{}),
		service: config.Service,
		version: config.Version,
	}
}

// NewDefaultLogger 创建默认日志器
func NewDefaultLogger(service string) *StructuredLogger {
	return NewLogger(LoggerConfig{
		Level:   LevelInfo,
		Output:  os.Stdout,
		Service: service,
		Version: "1.0.0",
	})
}

// Debug 记录调试日志
func (l *StructuredLogger) Debug(msg string, fields ...interface{}) {
	if l.level <= LevelDebug {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info 记录信息日志
func (l *StructuredLogger) Info(msg string, fields ...interface{}) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn 记录警告日志
func (l *StructuredLogger) Warn(msg string, fields ...interface{}) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error 记录错误日志
func (l *StructuredLogger) Error(msg string, fields ...interface{}) {
	if l.level <= LevelError {
		l.log(LevelError, msg, fields...)
	}
}

// Fatal 记录致命错误并退出
func (l *StructuredLogger) Fatal(msg string, fields ...interface{}) {
	l.log(LevelFatal, msg, fields...)
	os.Exit(1)
}

// With 添加字段
func (l *StructuredLogger) With(fields ...interface{}) Logger {
	newLogger := &StructuredLogger{
		level:     l.level,
		output:    l.output,
		baseLog:   l.baseLog,
		fields:    make(map[string]interface{}),
		service:   l.service,
		version:   l.version,
		requestID: l.requestID,
		userID:    l.userID,
	}

	// 复制现有字段
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// 添加新字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			newLogger.fields[key] = fields[i+1]
		}
	}

	return newLogger
}

// WithContext 从上下文中提取信息
func (l *StructuredLogger) WithContext(ctx context.Context) Logger {
	newLogger := l.clone()

	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		newLogger.requestID = requestID
	}

	if userID := GetUserIDFromContext(ctx); userID != "" {
		newLogger.userID = userID
	}

	return newLogger
}

// WithRequestID 添加请求ID
func (l *StructuredLogger) WithRequestID(requestID string) Logger {
	newLogger := l.clone()
	newLogger.requestID = requestID
	return newLogger
}

// WithUserID 添加用户ID
func (l *StructuredLogger) WithUserID(userID string) Logger {
	newLogger := l.clone()
	newLogger.userID = userID
	return newLogger
}

// clone 克隆日志器
func (l *StructuredLogger) clone() *StructuredLogger {
	newLogger := &StructuredLogger{
		level:     l.level,
		output:    l.output,
		baseLog:   l.baseLog,
		fields:    make(map[string]interface{}),
		service:   l.service,
		version:   l.version,
		requestID: l.requestID,
		userID:    l.userID,
	}

	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// log 执行实际的日志记录
func (l *StructuredLogger) log(level LogLevel, msg string, fields ...interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Service:   l.service,
		Version:   l.version,
		RequestID: l.requestID,
		UserID:    l.userID,
		Fields:    make(map[string]interface{}),
	}

	// 添加调用者信息
	if pc, file, line, ok := runtime.Caller(3); ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			entry.Caller = fmt.Sprintf("%s:%d %s", filepath.Base(file), line, fn.Name())
		}
	}

	// 复制基础字段
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// 添加新字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			entry.Fields[key] = fields[i+1]
		}
	}

	// 格式化并输出
	formatted := l.formatEntry(entry)
	l.baseLog.Println(formatted)
}

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// formatEntry 格式化日志条目
func (l *StructuredLogger) formatEntry(entry LogEntry) string {
	var parts []string

	// 时间戳
	parts = append(parts, entry.Timestamp.Format(time.RFC3339))

	// 级别
	parts = append(parts, fmt.Sprintf("[%s]", entry.Level))

	// 服务信息
	if entry.Service != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Service))
	}

	// 请求ID
	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("[req:%s]", entry.RequestID))
	}

	// 用户ID
	if entry.UserID != "" {
		parts = append(parts, fmt.Sprintf("[user:%s]", entry.UserID))
	}

	// 消息
	parts = append(parts, entry.Message)

	// 字段
	if len(entry.Fields) > 0 {
		var fieldParts []string
		for k, v := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("{%s}", strings.Join(fieldParts, ", ")))
	}

	// 调用者信息
	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("@%s", entry.Caller))
	}

	return strings.Join(parts, " ")
}

// 上下文相关函数

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

// GetRequestIDFromContext 从上下文获取请求ID
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithRequestID 在上下文中添加请求ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID 在上下文中添加用户ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GenerateRequestID 生成请求ID
func GenerateRequestID() string {
	return uuid.New().String()[:8]
}

// 全局日志器
var defaultLogger Logger

// InitDefaultLogger 初始化默认日志器
func InitDefaultLogger(service string, level LogLevel) {
	defaultLogger = NewLogger(LoggerConfig{
		Level:   level,
		Output:  os.Stdout,
		Service: service,
		Version: "1.0.0",
	})
}

// GetDefaultLogger 获取默认日志器
func GetDefaultLogger() Logger {
	if defaultLogger == nil {
		defaultLogger = NewDefaultLogger("courier-service")
	}
	return defaultLogger
}

// 全局日志函数
func Debug(msg string, fields ...interface{}) {
	GetDefaultLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
	GetDefaultLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	GetDefaultLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	GetDefaultLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
	GetDefaultLogger().Fatal(msg, fields...)
}
