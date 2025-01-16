package quest_system

import (
	"log"
	"sync"
)

// Logger 接口定义了基本的日志功能
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// DefaultLogger 是基于标准库 log 的日志实现
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("INFO: "+msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("ERROR: "+msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	log.Printf("WARN: "+msg, args...)
}

// 全局变量，用于保存当前的日志实例
var currentLogger Logger = &DefaultLogger{}
var loggerMutex sync.Mutex

// SetLogger 设置全局的日志实例
func SetLogger(logger Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	currentLogger = logger
}

// GetLogger 获取当前的日志实例
func GetLogger() Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	return currentLogger
}
