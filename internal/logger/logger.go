package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LogDebug = iota
	LogInfo
)

type Logger struct {
	file   *lumberjack.Logger
	slog   *slog.Logger
	writer io.Writer
}

// Create new logger with rotation
func New(logPath string, maxSizeMB, maxBackups, maxAgeDays int, compress bool) *Logger {
	file := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
	}
	return &Logger{file: file}
}

// logger initialization
func (l *Logger) Init(logLevel int) {
	// set log level
	var lvl slog.Leveler
	switch logLevel {
	case LogDebug:
		lvl = slog.LevelDebug
	case LogInfo:
	default:
		lvl = slog.LevelInfo
	}
	// set writing to stdout and file
	l.writer = io.MultiWriter(os.Stdout, l.file)
	// set slog handler
	handler := slog.NewTextHandler(l.writer, &slog.HandlerOptions{
		Level: lvl,
	})
	// create new slog
	l.slog = slog.New(handler)
	// set default for slog
	slog.SetDefault(l.slog)

	slog.Info("======== Start application ========")
	slog.Info("Logger was initialized")
	slog.Info("Debug log", "set", logLevel == LogDebug)
}

// logger deinit with closing file
func (l *Logger) Deinit() error {
	slog.Info("======== Stop application ========")
	return l.file.Close()
}

// get logger instance
func (l *Logger) Logger() *slog.Logger {
	return l.slog
}
