package udp

import (
	"fmt"
	"time"

	"github.com/ezra-sullivan/net-sniff/internal/pscan"
	"github.com/ezra-sullivan/net-sniff/pkg/options"
	"github.com/ezra-sullivan/net-sniff/pkg/utils"
	"github.com/spf13/cobra"
)

// NewCmdUDP 创建 UDP 命令
func NewCmdUDP(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "udp",
		Short:         "UDP 端口扫描",
		Long:          `对多个主机执行 UDP 端口扫描，支持指定端口范围。`,
		SilenceUsage:  false,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 添加panic恢复机制
			defer func() {
				if r := recover(); r != nil {
					opts.Logger.Error("命令执行过程中发生严重错误", "error", r)
				}
			}()

			// 检查主机列表是否为空
			if opts.Hosts == "" {
				opts.Logger.Error("必须指定主机列表")
				return fmt.Errorf("必须指定主机列表")
			}

			// 检查端口列表是否为空
			if opts.Ports == "" {
				opts.Logger.Error("必须指定端口列表")
				return fmt.Errorf("必须指定端口列表")
			}

			return runUDP(opts)
		},
	}

	// 添加命令特定的标志
	addFlags(cmd, opts)

	return cmd
}

// addFlags 添加命令特定的标志
func addFlags(cmd *cobra.Command, opts *options.Options) {
	cmd.Flags().StringVarP(&opts.Hosts, "hosts", "H", "", "主机列表，逗号分隔或范围")
	cmd.Flags().StringVarP(&opts.Ports, "ports", "p", "", "端口列表，逗号分隔或范围")
}

// runUDP 执行 UDP 扫描命令
func runUDP(opts *options.Options) error {
	// 解析主机列表
	hostList, err := utils.ParseHostList(opts.Hosts)
	if err != nil {
		opts.Logger.Error("解析主机列表错误", "error", err)
		return err
	}

	// 解析端口范围
	portList, err := utils.ParsePortRange(opts.Ports)
	if err != nil {
		opts.Logger.Error("解析端口范围错误", "error", err)
		return err
	}

	opts.Logger.Info("开始 UDP 扫描",
		"主机数量", len(hostList),
		"端口数量", len(portList),
		"并发数", opts.Concurrency,
		"超时", opts.Timeout)

	// 执行 UDP 端口扫描
	results := pscan.BatchScanUDPPorts(hostList, portList, opts.Concurrency, time.Duration(opts.Timeout)*time.Millisecond)

	// 统计结果
	openCount := 0
	for _, result := range results {
		if result.Open {
			openCount++
		}

		// 输出到控制台
		if opts.Verbose || result.Open {
			status := "可能开放"
			if !result.Open {
				status = "关闭"
			}

			if result.Open {
				opts.Logger.Info("UDP 扫描结果",
					"host", result.Host,
					"port", result.Port,
					"status", status,
					"time_ms", float64(result.Time.Microseconds())/1000.0)
			} else {
				opts.Logger.Debug("UDP 扫描结果",
					"host", result.Host,
					"port", result.Port,
					"status", status,
					"time_ms", float64(result.Time.Microseconds())/1000.0)
			}
		}

		// 输出到文件
		if opts.OutputWriter != nil && result.Open {
			_, err := fmt.Fprintf(opts.OutputWriter, "%s,%d,UDP,可能开放,%.2f\n",
				result.Host, result.Port, float64(result.Time.Microseconds())/1000.0)
			if err != nil {
				return err
			}
		}
	}

	// 输出总结信息
	opts.Logger.Info("UDP 扫描完成", "可能开放端口数", openCount, "总扫描数", len(results))

	return nil
}
