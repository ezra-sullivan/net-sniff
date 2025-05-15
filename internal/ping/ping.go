package ping

import (
	"fmt"
	"github.com/ezra-sullivan/net-sniff/internal/global"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// Result 存储 ping 的结果
type Result struct {
	Success bool
	Host    string
	TTL     uint8
	Time    time.Duration
	Error   error
}

// SinglePing 对单个主机执行 ping 操作
func SinglePing(host string, timeout time.Duration) Result {
	result := Result{
		Host: host,
	}

	// 创建新的 pinger
	pinger, err := probing.NewPinger(host)
	if err != nil {
		result.Success = false
		result.Error = err
		return result
	}

	// 设置 ping 参数
	pinger.Count = 1
	pinger.Timeout = timeout   // 使用传入的超时值
	pinger.SetPrivileged(true) // Windows 系统需要设置为 true

	// 执行 ping
	startTime := time.Now()
	err = pinger.Run()
	result.Time = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = err
		return result
	}

	// 获取统计信息
	stats := pinger.Statistics()
	result.Success = stats.PacketsRecv > 0
	if result.Success {
		result.TTL = stats.TTLs[0]
	} else if result.Error == nil {
		// 当没有收到任何数据包但没有错误时，设置一个有意义的错误信息
		result.Error = fmt.Errorf("目标主机不可达或未响应")
	}

	// 直接输出结果
	result.output()

	return result
}

// BatchPing 对多个主机执行批量 ping 操作
func BatchPing(hosts []string, concurrency int, timeout time.Duration) []Result {
	results := make([]Result, 0, len(hosts))
	resultsChan := make(chan Result, len(hosts))

	// 使用信号量控制并发
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			resultsChan <- SinglePing(h, timeout) // 传递超时参数
		}(host)
	}

	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 收集结果
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// 定义 Result 结构体的 output 方法
func (result *Result) output() {
	// 获取全局的 consoleLogger
	consoleLogger := global.ConsoleLogger

	// 如果 Ping 成功
	if result.Success {
		// 打印Ping结果
		consoleLogger.Info(
			"Ping Result",
			"status", "success",
			"host", result.Host,
			"ttl", result.TTL,
			"time_ms", float64(result.Time.Microseconds())/1000.0)
	} else {
		// 打印 Ping 失败结果
		consoleLogger.Debug("Ping Result",
			"status", "failed",
			"host", result.Host,
			"err", result.Error,
		)
	}

	// 获取全局的 fileLogger
	fileLogger := global.FileLogger
	// 如果 fileLogger 不为空
	if fileLogger != nil {

		// 如果 Ping 成功
		if result.Success {
			// 打印Ping结果到文件
			fileLogger.Info(fmt.Sprintf("%s,%d,%s,%s,%.2f\n", result.Host, result.TTL, "success", "success", float64(result.Time.Microseconds())/1000.0))

		} else {
			// 打印 Ping 失败结果到文件
			fileLogger.Debug(fmt.Sprintf("%s,%d,%s,%s,%.2f\n", result.Host, result.TTL, "failed", result.Error, float64(result.Time.Microseconds())/1000.0))
		}
	}
}
