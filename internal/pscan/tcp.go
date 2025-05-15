package pscan

import (
	"fmt"
	"github.com/ezra-sullivan/net-sniff/internal/global"
	"net"
	"sync"
	"time"
)

// TCPScanResult 存储 TCP 端口扫描结果
type TCPScanResult struct {
	Success bool
	Host    string
	Port    int
	IsOpen  bool
	Error   error
	Time    time.Duration
}

// ScanTCPPort 扫描单个 TCP 端口
// ScanTCPPort 函数用于扫描指定的TCP端口是否开放
func ScanTCPPort(host string, port int, timeout time.Duration) TCPScanResult {

	// 将主机名和端口号拼接成地址
	address := fmt.Sprintf("%s:%d", host, port)
	// 记录开始时间
	startTime := time.Now()

	// 尝试连接指定的TCP端口
	conn, err := net.DialTimeout("tcp", address, timeout)

	// 在连接尝试后立即计算时间
	deration := time.Since(startTime)
	result := TCPScanResult{
		Host:   host,
		Port:   port,
		Time:   deration,
		Error:  err,
		IsOpen: err == nil,
	}

	// 如果连接失败，则设置结果为未开放，并返回结果
	if conn != nil {
		defer func() {
			err = conn.Close()
			if err != nil {
				result.Error = fmt.Errorf("注意内存溢出，关闭连接失败: %w", err)
			}
		}()
	}

	result.output()
	// 返回结果
	return result
}

// BatchScanTCPPorts 批量扫描多个主机的多个 TCP 端口
func BatchScanTCPPorts(hosts []string, ports []int, concurrency int, timeout time.Duration) []TCPScanResult {
	totalScans := len(hosts) * len(ports)
	results := make([]TCPScanResult, 0, totalScans)
	resultsChan := make(chan TCPScanResult, totalScans)

	// 使用信号量控制并发
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, host := range hosts {
		for _, port := range ports {
			wg.Add(1)
			go func(h string, p int) {
				defer wg.Done()
				sem <- struct{}{}        // 获取信号量
				defer func() { <-sem }() // 释放信号量

				resultsChan <- ScanTCPPort(h, p, timeout)
			}(host, port)
		}
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

// 定义一个名为output的函数
func (result *TCPScanResult) output() {
	// 获取全局的 consoleLogger
	consoleLogger := global.ConsoleLogger
	if result.IsOpen {
		// 打印TCP端口开放结果
		consoleLogger.Info("TCP Port Scan Result",
			"host", result.Host,
			"port", result.Port,
			"status", "open",
			"time_ms", float64(result.Time.Microseconds())/1000.0,
		)
	} else {
		consoleLogger.Debug("TCP Port Scan Result",
			"host", result.Host,
			"port", result.Port,
			"status", "closed",
			"time_ms", float64(result.Time.Microseconds())/1000.0,
			"err", result.Error,
		)
	}

	// 获取全局的 fileLogger
	fileLogger := global.FileLogger
	// 如果 fileLogger 不为空
	if fileLogger != nil {
		if result.IsOpen {
			// 打印 TCP 端口开放结果到文件
			fileLogger.Info(fmt.Sprintf("%s,%d,%s,%.2f\n", result.Host, result.Port, "open", float64(result.Time.Microseconds())/1000.0))
		} else {
			// 打印TCP端口关闭结果到文件
			fileLogger.Debug(fmt.Sprintf("%s,%d,%s,%.2f\n", result.Host, result.Port, "closed", float64(result.Time.Microseconds())/1000.0))
		}
	}

}
