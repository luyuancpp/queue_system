package quest_system

import (
	"testing"
)

// Logger 抽象日志记录
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// DefaultLogger 使用测试框架中的日志记录
type DefaultLogger struct {
	t *testing.T
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	l.t.Logf(msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	l.t.Errorf(msg, args...)
}
