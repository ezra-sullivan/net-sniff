package udp

import (
	"fmt"
	"github.com/ezra-sullivan/net-sniff/internal/global"
	"github.com/ezra-sullivan/net-sniff/internal/initialize/logger"
	"github.com/ezra-sullivan/net-sniff/internal/initialize/options"
	"time"

	"github.com/ezra-sullivan/net-sniff/internal/pscan"
	"github.com/ezra-sullivan/net-sniff/pkg/utils"
	"github.com/spf13/cobra"
)

// NewCmdUDP 创建 UDP 命令
func NewCmdUDP(opts *options.Options) *cobra.Command {
	consoleLogger := global.ConsoleLogger
	cmd := &cobra.Command{
		Use:           "udp",
		Short:         "UDP 端口扫描",
		Long:          `对多个主机执行 UDP 端口扫描，支持指定端口范围。`,
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

			// 检查端口列表是否为空
			if opts.Ports == "" {
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
	consoleLogger := global.ConsoleLogger
	// 解析主机列表
	hostList, err := utils.ParseHostList(opts.Hosts)
	if err != nil {
		consoleLogger.Error("解析主机列表错误", "error", err)
		return err
	}

	// 解析端口范围
	portList, err := utils.ParsePortRange(opts.Ports)
	if err != nil {
		consoleLogger.Error("解析端口范围错误", "error", err)
		return err
	}

	logger.OutputStart("UDP 扫描", len(hostList), len(portList))

	// 执行 UDP 端口扫描
	results := pscan.BatchScanUDPPorts(hostList, portList, opts.Concurrency, time.Duration(opts.Timeout)*time.Millisecond)

	// 统计结果
	openCount := 0
	for _, result := range results {
		if result.IsOpen == pscan.UDP_PORT_OPEN {
			openCount++
		}
	}

	// 输出总结信息
	logger.OutputSummary("UDP 扫描", openCount, len(results))

	return nil
}
