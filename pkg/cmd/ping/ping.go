package ping

import (
	"fmt"

	"github.com/ezra-sullivan/net-sniff/internal/ping"
	"github.com/ezra-sullivan/net-sniff/pkg/options"
	"github.com/ezra-sullivan/net-sniff/pkg/utils"
	"github.com/spf13/cobra"
)

// NewCmdPing 创建 ping 命令
func NewCmdPing(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ping",
		Short:         "批量 Ping 主机",
		Long:          `对多个主机执行批量 Ping 操作，支持从文件读取主机列表。`,
		SilenceUsage:  false,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 检查主机列表是否为空
			if opts.Hosts == "" {
				opts.Logger.Error("必须指定主机列表")
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
	cmd.Flags().IntVarP(&opts.Concurrency, "concurrency", "c", 100, "并发数")
	cmd.Flags().StringVarP(&opts.OutputFile, "output", "o", "", "输出文件路径")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "显示详细信息")
}

// runPing 执行 ping 命令
func runPing(opts *options.Options) error {
	// 解析主机列表
	hostList, err := utils.ParseHostList(opts.Hosts)
	if err != nil {
		opts.Logger.Error("解析主机列表错误", "error", err)
		return err
	}

	opts.Logger.Info("开始 Ping", "主机数量", len(hostList), "并发数", opts.Concurrency)

	// 执行批量 Ping
	results := ping.BatchPing(hostList, opts.Concurrency)

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}

		// 输出到控制台
		if opts.Verbose || result.Success {
			status := "成功"
			if !result.Success {
				status = "失败"
			}

			if result.Success {
				opts.Logger.Info("Ping 结果",
					"host", result.Host,
					"status", status,
					"time_ms", float64(result.Time.Microseconds())/1000.0)
			} else {
				opts.Logger.Debug("Ping 结果",
					"host", result.Host,
					"status", status,
					"time_ms", float64(result.Time.Microseconds())/1000.0)
			}
		}

		// 输出到文件
		if opts.OutputWriter != nil {
			status := "成功"
			if !result.Success {
				status = "失败"
			}
			_, err := fmt.Fprintf(opts.OutputWriter, "%s,%s,%.2f\n", result.Host, status, float64(result.Time.Microseconds())/1000.0)
			if err != nil {
				return err
			}
		}
	}

	// 计算成功率并输出总结信息
	successRate := float64(successCount) / float64(len(results)) * 100
	opts.Logger.Info("Ping 完成",
		"成功数", successCount,
		"总数", len(results),
		"成功率", fmt.Sprintf("%.2f%%", successRate))

	return nil
}
