// Package ui handles user interaction, logging, and progress display.
package ui

import (
	"log/slog"
	"os"
	"strings"
)

// Log is the global logger instance
var Log *slog.Logger

func init() {
	// Default to info level
	Setup("info")
}

// Setup initializes the global logger with the specified level
func Setup(levelStr string) {
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		// Fallback to info
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Use TextHandler for CLI, output to Stderr
	handler := slog.NewTextHandler(os.Stderr, opts)
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// Debug logs at Debug level
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

// Info logs at Info level
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Warn logs at Warn level
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Error logs at Error level
func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}
