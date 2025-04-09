package options

import (
	"log/slog"
	"os"
)

// Options 定义全局配置选项
type Options struct {
	Hosts       string
	Ports       string
	Concurrency int
	Timeout     int
	OutputFile  string
	Verbose     bool
	Mode        string
	LogLevel    string
	OutputWriter *os.File
	Logger      *slog.Logger
}

// NewOptions 创建默认选项
func NewOptions() *Options {
	return &Options{
		Concurrency: 100,
		Timeout:     1000,
		LogLevel:    "info",
	}
}

// InitLogger 初始化日志记录器
func (o *Options) InitLogger() {
	var level slog.Level
	switch o.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	o.Logger = slog.New(handler)
	slog.SetDefault(o.Logger)
}