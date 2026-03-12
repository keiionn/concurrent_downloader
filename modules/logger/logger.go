package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger 日志管理器
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// NewLogger 创建一个新的日志管理器
func NewLogger() *Logger {
	// 创建log文件夹
	logDir := "log"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
	}

	// 创建日志文件
	now := time.Now()
	logFile := filepath.Join(logDir, fmt.Sprintf("downloader_%s.log", now.Format("2006-01-02_15-04-05")))

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("打开日志文件失败: %v\n", err)
		// 如果打开文件失败，使用标准输出
		return &Logger{
			infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
			errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
			debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime),
		}
	}

	// 创建日志记录器
	return &Logger{
		infoLogger:  log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info 记录信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
	// 同时输出到控制台
	fmt.Printf("[INFO] "+format+"\n", v...)
}

// Error 记录错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
	// 同时输出到控制台
	fmt.Printf("[ERROR] "+format+"\n", v...)
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
	// 同时输出到控制台
	fmt.Printf("[DEBUG] "+format+"\n", v...)
}

// 全局日志实例
var logger *Logger

// GetLogger 获取日志实例
func GetLogger() *Logger {
	if logger == nil {
		logger = NewLogger()
	}
	return logger
}

// 辅助函数，方便直接使用
func LogInfo(format string, v ...interface{}) {
	GetLogger().Info(format, v...)
}

func LogError(format string, v ...interface{}) {
	GetLogger().Error(format, v...)
}

func LogDebug(format string, v ...interface{}) {
	GetLogger().Debug(format, v...)
}
