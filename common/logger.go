package common

import (
	"log"
	"os"
	"strings"
)

// Logger 分级日志输出，敏感内容自动脱敏
type Logger struct {
	level int
}

const (
	levelDebug = iota
	levelInfo
	levelWarn
	levelError
)

var Log = &Logger{level: levelInfo}

func InitLogger(level string) {
	switch strings.ToLower(level) {
	case "debug":
		Log.level = levelDebug
	case "info":
		Log.level = levelInfo
	case "warn", "warning":
		Log.level = levelWarn
	case "error":
		Log.level = levelError
	default:
		Log.level = levelInfo
	}
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= levelDebug {
		log.Printf("[DEBUG] "+DesensitizeLog(format), v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= levelInfo {
		log.Printf("[INFO] "+DesensitizeLog(format), v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= levelWarn {
		log.Printf("[WARN] "+DesensitizeLog(format), v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= levelError {
		log.Printf("[ERROR] "+DesensitizeLog(format), v...)
	}
}
