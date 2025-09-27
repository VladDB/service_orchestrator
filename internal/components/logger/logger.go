package logger

import (
	"io"
	"log/slog"
	"os"
	"service_orchestrator/internal/components/globals"

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
func (l *Logger) Init() {
	// set log level
	var lvl slog.Level
	switch globals.LogLevel {
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
}

// logger deinit with closing file
func (l *Logger) Deinit() error {
	slog.Info("======== Stop application ========")
	return l.file.Close()
}

// ChangeLevel allows changing the logger level at runtime
func (l *Logger) ChangeLevel(logLevel int) {
	var lvl slog.Level
	switch logLevel {
	case LogDebug:
		lvl = slog.LevelDebug
	case LogInfo:
	default:
		lvl = slog.LevelInfo
	}

	// Re-create handler with new level
	handler := slog.NewTextHandler(l.writer, &slog.HandlerOptions{
		Level: lvl,
	})
	l.slog = slog.New(handler)
	slog.SetDefault(l.slog)
	slog.Info("Logger level set", "level", lvl.String())
}

// get logger instance
func (l *Logger) Logger() *slog.Logger {
	return l.slog
}
