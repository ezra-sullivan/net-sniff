package ping

import (
	"fmt"
	"github.com/ezra-sullivan/net-sniff/internal/global"
	"github.com/ezra-sullivan/net-sniff/internal/initialize/logger"
	"github.com/ezra-sullivan/net-sniff/internal/initialize/options"
	"time"

	"github.com/ezra-sullivan/net-sniff/internal/ping"
	"github.com/ezra-sullivan/net-sniff/pkg/utils"
	"github.com/spf13/cobra"
)

// NewCmdPing 创建 ping 命令
func NewCmdPing(opts *options.Options) *cobra.Command {

	consoleLogger := global.ConsoleLogger

	cmd := &cobra.Command{
		Use:           "ping",
		Short:         "批量 Ping 主机",
		Long:          `对多个主机执行批量 Ping 操作，支持从文件读取主机列表。`,
		SilenceUsage:  false,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 添加 panic 恢复机制
			defer func() {
				if r := recover(); r != nil {
					consoleLogger.Error("命令执行过程中发生严重错误", "error", r)
				}
			}()

			// 检查主机列表是否为空
			if opts.Hosts == "" {
				return fmt.Errorf("必须指定主机列表")
			}
			return runPing(opts)
		},
	}

	// 添加命令特定的标志
	addFlags(cmd, opts)

	return cmd
}

// addFlags 添加命令特定的标志
func addFlags(cmd *cobra.Command, opts *options.Options) {
	cmd.Flags().StringVarP(&opts.Hosts, "hosts", "H", "", "主机列表，逗号分隔或文件路径")
}

// runPing 执行 ping 命令
func runPing(opts *options.Options) error {
	// 解析主机列表

	consoleLogger := global.ConsoleLogger

	hostList, err := utils.ParseHostList(opts.Hosts)
	if err != nil {
		consoleLogger.Error("解析主机列表错误", "error", err)
		return err
	}

	logger.OutputStart("Ping", len(hostList), 0)

	// 执行批量 Ping，传入超时参数
	results := ping.BatchPing(hostList, opts.Concurrency, time.Duration(opts.Timeout)*time.Millisecond)

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}

	}

	// 输出总结信息
	logger.OutputSummary("Ping", successCount, len(results))

	return nil
}
