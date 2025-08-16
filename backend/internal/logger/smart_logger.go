package logger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO", 
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// SmartLogger 智能日志器，支持级别控制和限流
type SmartLogger struct {
	level        LogLevel
	rateLimiters map[string]*RateLimiter
	mu           sync.RWMutex
}

// RateLimiter 日志限流器
type RateLimiter struct {
	lastLog   time.Time
	count     int
	threshold int
	window    time.Duration
	mu        sync.Mutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter(threshold int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		threshold: threshold,
		window:    window,
		lastLog:   time.Now(),
	}
}

// Allow 检查是否允许记录日志
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	if now.Sub(rl.lastLog) > rl.window {
		rl.count = 0
		rl.lastLog = now
	}
	
	if rl.count >= rl.threshold {
		return false
	}
	
	rl.count++
	return true
}

var (
	defaultLogger *SmartLogger
	once          sync.Once
)

// GetLogger 获取默认日志器实例
func GetLogger() *SmartLogger {
	once.Do(func() {
		level := INFO
		if env := os.Getenv("LOG_LEVEL"); env != "" {
			switch env {
			case "DEBUG":
				level = DEBUG
			case "INFO":
				level = INFO
			case "WARN":
				level = WARN
			case "ERROR":
				level = ERROR
			case "FATAL":
				level = FATAL
			}
		}
		
		// 生产环境默认使用WARN级别
		if os.Getenv("GIN_MODE") == "release" {
			level = WARN
		}
		
		defaultLogger = &SmartLogger{
			level:        level,
			rateLimiters: make(map[string]*RateLimiter),
		}
	})
	return defaultLogger
}

// getRateLimiter 获取或创建限流器
func (sl *SmartLogger) getRateLimiter(key string) *RateLimiter {
	sl.mu.RLock()
	limiter, exists := sl.rateLimiters[key]
	sl.mu.RUnlock()
	
	if !exists {
		sl.mu.Lock()
		limiter = NewRateLimiter(10, time.Minute) // 默认每分钟最多10条相同日志
		sl.rateLimiters[key] = limiter
		sl.mu.Unlock()
	}
	
	return limiter
}

// shouldLog 检查是否应该记录日志
func (sl *SmartLogger) shouldLog(level LogLevel) bool {
	return level >= sl.level
}

// logWithLevel 通用日志记录方法
func (sl *SmartLogger) logWithLevel(level LogLevel, key string, format string, args ...interface{}) {
	if !sl.shouldLog(level) {
		return
	}
	
	// 对于DEBUG和INFO级别的日志进行限流
	if level <= INFO && key != "" {
		limiter := sl.getRateLimiter(key)
		if !limiter.Allow() {
			return
		}
	}
	
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)
	
	log.Printf("[%s] [%s] %s", timestamp, levelName, message)
}

// Debug 调试日志
func (sl *SmartLogger) Debug(format string, args ...interface{}) {
	sl.logWithLevel(DEBUG, "", format, args...)
}

// DebugWithKey 带限流键的调试日志
func (sl *SmartLogger) DebugWithKey(key string, format string, args ...interface{}) {
	sl.logWithLevel(DEBUG, key, format, args...)
}

// Info 信息日志
func (sl *SmartLogger) Info(format string, args ...interface{}) {
	sl.logWithLevel(INFO, "", format, args...)
}

// InfoWithKey 带限流键的信息日志
func (sl *SmartLogger) InfoWithKey(key string, format string, args ...interface{}) {
	sl.logWithLevel(INFO, key, format, args...)
}

// Warn 警告日志
func (sl *SmartLogger) Warn(format string, args ...interface{}) {
	sl.logWithLevel(WARN, "", format, args...)
}

// Error 错误日志
func (sl *SmartLogger) Error(format string, args ...interface{}) {
	sl.logWithLevel(ERROR, "", format, args...)
}

// Fatal 致命错误日志
func (sl *SmartLogger) Fatal(format string, args ...interface{}) {
	sl.logWithLevel(FATAL, "", format, args...)
	os.Exit(1)
}

// 全局便捷方法
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func DebugWithKey(key string, format string, args ...interface{}) {
	GetLogger().DebugWithKey(key, format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func InfoWithKey(key string, format string, args ...interface{}) {
	GetLogger().InfoWithKey(key, format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}

// GORM Logger Implementation

// CustomGormLogger 自定义GORM日志器，集成智能日志系统
type CustomGormLogger struct {
	smartLogger *SmartLogger
	config      gormlogger.Config
}

// NewCustomGormLogger 创建自定义GORM日志器
func NewCustomGormLogger() gormlogger.Interface {
	return &CustomGormLogger{
		smartLogger: GetLogger(),
		config: gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
			LogLevel:                  gormlogger.Warn,        // 只记录警告和错误
			IgnoreRecordNotFoundError: true,                   // 忽略记录未找到错误
			Colorful:                  false,                  // 生产环境不使用颜色
		},
	}
}

// LogMode 设置日志模式
func (l *CustomGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.config.LogLevel = level
	return &newlogger
}

// Info 信息日志
func (l *CustomGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Info {
		l.smartLogger.InfoWithKey("gorm_info", msg, data...)
	}
}

// Warn 警告日志
func (l *CustomGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Warn {
		l.smartLogger.Warn(msg, data...)
	}
}

// Error 错误日志
func (l *CustomGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Error {
		l.smartLogger.Error(msg, data...)
	}
}

// Trace 跟踪SQL执行
func (l *CustomGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.config.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.config.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.config.IgnoreRecordNotFoundError):
		l.smartLogger.Error("SQL Error: %v [%.3fms] [rows:%v] %s", err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case elapsed > l.config.SlowThreshold && l.config.SlowThreshold != 0 && l.config.LogLevel >= gormlogger.Warn:
		l.smartLogger.Warn("Slow SQL [%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case l.config.LogLevel == gormlogger.Info:
		// 使用限流键避免SQL日志过多
		l.smartLogger.DebugWithKey("sql_trace", "[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}