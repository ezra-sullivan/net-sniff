package initialize

import (
	"os"

	"github.com/ezra-sullivan/net-sniff/pkg/options"
	"github.com/spf13/cobra"
)

// Command 初始化命令相关资源
func Command(cmd *cobra.Command, opts *options.Options) error {
	// 初始化日志记录器
	InitLogger(opts)

	// 检查必要参数
	if opts.Hosts == "" && cmd.Name() != "help" && cmd.Parent() != nil && cmd.Parent().Name() != "help" {
		opts.Logger.Error("必须指定主机列表")
		cmd.Help()
		return cmd.Usage()
	}

	// 准备输出文件
	if opts.OutputFile != "" {
		return InitOutputFile(opts)
	}

	return nil
}

// InitLogger 初始化日志记录器
func InitLogger(opts *options.Options) {
	opts.InitLogger()
}

// InitOutputFile 初始化输出文件
func InitOutputFile(opts *options.Options) error {
	var err error
	opts.OutputWriter, err = os.Create(opts.OutputFile)
	if err != nil {
		opts.Logger.Error("创建输出文件错误", "error", err)
		return err
	}
	return nil
}

// Finalize 完成命令执行后的清理工作
func Finalize(opts *options.Options) error {
	if opts.OutputWriter != nil {
		return opts.OutputWriter.Close()
	}
	return nil
}