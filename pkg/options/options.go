package options

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options 定义全局配置选项
type Options struct {
	Hosts        string
	Ports        string
	Concurrency  int
	Timeout      int
	Verbose      bool
	Mode         string
	OutputFile   string
	LogLevel     string
	Logger       *slog.Logger
	OutputWriter io.WriteCloser // 添加输出文件写入器
}

// NewOptions 创建默认选项
func NewOptions() *Options {
	return &Options{
		Concurrency: 100,
		Timeout:     1000,
		LogLevel:    "info",
	}
}

func (opts *Options) InitOpts() error {
	var err error
	// 初始化日志
	err = opts.InitLogger()
	if err != nil {
		return err
	}
	// 初始化输出文件
	err = opts.InitOutputFile()
	if err != nil {
		return err
	}

	return nil
}

// InitLogger 初始化日志
func (opts *Options) InitLogger() error {
	// 设置日志级别
	var level slog.Level
	switch strings.ToLower(opts.LogLevel) {
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
	// 创建日志处理器
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	// 设置全局日志记录器
	logger := slog.New(handler)
	slog.SetDefault(logger)
	opts.Logger = logger

	return nil
}

// InitOutputFile 初始化输出文件
func (opts *Options) InitOutputFile() error {
	var err error
	if opts.OutputFile != "" {
		opts.OutputWriter, err = os.Create(opts.OutputFile)
		if err != nil {
			if opts.Logger != nil {
				opts.Logger.Error("创建输出文件错误", "error", err)
			}
			return err
		}
	}
	return nil
}

// Close 实现 io.Closer 接口，关闭资源
func (opts *Options) Close() error {
	if opts.OutputWriter != nil {
		return opts.OutputWriter.Close()
	}
	return nil
}
