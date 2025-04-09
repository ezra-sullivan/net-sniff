package cmd

import (
	"fmt"
	"os"
	"strings"

	"log/slog"

	"github.com/ezra-sullivan/net-sniff/pkg/cmd/ping"
	"github.com/ezra-sullivan/net-sniff/pkg/cmd/tcp"
	"github.com/ezra-sullivan/net-sniff/pkg/cmd/udp"
	"github.com/ezra-sullivan/net-sniff/pkg/options"
	"github.com/spf13/cobra"
)

// Execute 执行根命令
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// NewNetSniffCommand 创建根命令 (与 main.go 中的引用保持一致)
func NewNetSniffCommand() *cobra.Command {
	return NewRootCmd()
}

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	opts := options.NewOptions()

	rootCmd := &cobra.Command{
		Use:           "net-sniff",
		Short:         "网络探测工具",
		Long:          `网络探测工具，支持批量 Ping、TCP/UDP 端口扫描等功能。`,
		SilenceUsage:  false,
		SilenceErrors: false,
	}

	// 添加全局标志
	rootCmd.PersistentFlags().StringVarP(&opts.LogLevel, "log-level", "l", "info", "日志级别: debug, info, warn, error")
	rootCmd.PersistentFlags().StringVarP(&opts.Ports, "ports", "p", "", "端口列表，逗号分隔或范围（例如 80,443,8000-8100）")
	rootCmd.PersistentFlags().IntVarP(&opts.Timeout, "timeout", "t", 1000, "超时时间（毫秒）")

	// 初始化日志
	initLogger(opts)

	// 添加子命令
	rootCmd.AddCommand(ping.NewCmdPing(opts))
	rootCmd.AddCommand(tcp.NewCmdTCP(opts))
	rootCmd.AddCommand(udp.NewCmdUDP(opts))

	return rootCmd
}

// initLogger 初始化日志
func initLogger(opts *options.Options) {
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
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	// 设置全局日志记录器
	logger := slog.New(handler)
	slog.SetDefault(logger)
	opts.Logger = logger
}

// openOutputFile 打开输出文件
func openOutputFile(opts *options.Options) error {
	if opts.OutputFile != "" {
		file, err := os.Create(opts.OutputFile)
		if err != nil {
			return fmt.Errorf("无法创建输出文件: %w", err)
		}
		opts.OutputWriter = file
	}
	return nil
}
