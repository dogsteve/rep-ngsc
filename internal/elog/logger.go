package elog

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

var service string

type Level int32

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var level atomic.Int32

// Fields is a helper for structured log fields.
type Fields map[string]interface{}

// Init initializes the logger with a level string and service name.
func Init(levelStr string, svc string) error {
	service = svc
	switch levelStr {
	case "debug", "DEBUG":
		level.Store(int32(LevelDebug))
	case "info", "INFO":
		level.Store(int32(LevelInfo))
	case "warn", "WARN":
		level.Store(int32(LevelWarn))
	case "error", "ERROR":
		level.Store(int32(LevelError))
	case "fatal", "FATAL":
		level.Store(int32(LevelFatal))
	default:
		// default to info
		level.Store(int32(LevelInfo))
	}
	return nil
}

func shouldLog(l Level) bool {
	return Level(level.Load()) <= l
}

func write(entry map[string]interface{}, toStderr bool) {
	b, err := json.Marshal(entry)
	if err != nil {
		// fallback to fmt if marshal fails
		if toStderr {
			fmt.Fprintln(os.Stderr, "log marshal error:", err)
		} else {
			fmt.Println("log marshal error:", err)
		}
		return
	}
	if toStderr {
		fmt.Fprintln(os.Stderr, string(b))
	} else {
		fmt.Fprintln(os.Stdout, string(b))
	}
}

func baseEntry(levelStr, msg string, fields Fields) map[string]interface{} {
	entry := map[string]interface{}{
		"ts":      time.Now().Format(time.RFC3339),
		"level":   levelStr,
		"service": service,
		"msg":     msg,
	}
	for k, v := range fields {
		// avoid overwriting base keys
		if k == "ts" || k == "level" || k == "service" || k == "msg" {
			continue
		}
		entry[k] = v
	}
	return entry
}

func Debug(msg string, f Fields) {
	if !shouldLog(LevelDebug) {
		return
	}
	if f == nil {
		f = Fields{}
	}
	write(baseEntry("debug", msg, f), false)
}

func Info(msg string, f Fields) {
	if !shouldLog(LevelInfo) {
		return
	}
	if f == nil {
		f = Fields{}
	}
	write(baseEntry("info", msg, f), false)
}

func Warn(msg string, f Fields) {
	if !shouldLog(LevelWarn) {
		return
	}
	if f == nil {
		f = Fields{}
	}
	write(baseEntry("warn", msg, f), false)
}

func Error(msg string, f Fields) {
	if !shouldLog(LevelError) {
		return
	}
	if f == nil {
		f = Fields{}
	}
	write(baseEntry("error", msg, f), true)
}

func Fatal(msg string, f Fields) {
	if f == nil {
		f = Fields{}
	}
	write(baseEntry("fatal", msg, f), true)
	os.Exit(1)
}

func F(k string, v interface{}) Fields {
	return Fields{k: v}
}
