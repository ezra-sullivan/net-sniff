package cmd

import (
	"github.com/ezra-sullivan/net-sniff/pkg/cmd/ping"
	"github.com/ezra-sullivan/net-sniff/pkg/cmd/tcp"
	"github.com/ezra-sullivan/net-sniff/pkg/cmd/udp"
	"github.com/ezra-sullivan/net-sniff/pkg/options"
	"github.com/spf13/cobra"
)

// Execute 执行根命令
//func Execute() {
//	if err := NewRootCmd().Execute(); err != nil {
//		os.Exit(1)
//	}
//}

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 初始化日志
			return opts.InitOpts()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// 如果是根命令直接执行，则显示帮助信息
			return cmd.Help()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// 添加命令执行后的清理工作
			err := opts.Close()
			if err != nil {
				opts.Logger.Error("关闭资源错误", "error", err)
				return err
			}
			return nil
		},
	}

	// 添加全局标志
	rootCmd.PersistentFlags().IntVarP(&opts.Timeout, "timeout", "t", 1000, "超时时间（毫秒）")
	rootCmd.PersistentFlags().IntVarP(&opts.Concurrency, "concurrency", "c", 100, "并发数")
	rootCmd.PersistentFlags().StringVarP(&opts.OutputFile, "output", "o", "", "输出文件路径")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "详细模式")
	rootCmd.PersistentFlags().StringVarP(&opts.LogLevel, "log-level", "l", "info", "日志级别: debug, info, warn, error")

	// 添加子命令
	rootCmd.AddCommand(ping.NewCmdPing(opts))
	rootCmd.AddCommand(tcp.NewCmdTCP(opts))
	rootCmd.AddCommand(udp.NewCmdUDP(opts))

	return rootCmd
}
