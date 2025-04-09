package pscan

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPScanResult 存储 TCP 端口扫描结果
type TCPScanResult struct {
	Host  string
	Port  int
	Open  bool
	Error error
	Time  time.Duration
}

// ScanTCPPort 扫描单个 TCP 端口
func ScanTCPPort(host string, port int, timeout time.Duration) TCPScanResult {
	result := TCPScanResult{
		Host: host,
		Port: port,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	startTime := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	result.Time = time.Since(startTime)

	if err != nil {
		result.Open = false
		result.Error = err
		return result
	}

	defer conn.Close()
	result.Open = true
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
