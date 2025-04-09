package pscan

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// UDPScanResult 存储 UDP 端口扫描结果
type UDPScanResult struct {
	Host  string
	Port  int
	Open  bool
	Error error
	Time  time.Duration
}

// ScanUDPPort 扫描单个 UDP 端口
func ScanUDPPort(host string, port int, timeout time.Duration) UDPScanResult {
	result := UDPScanResult{
		Host: host,
		Port: port,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	startTime := time.Now()

	conn, err := net.DialTimeout("udp", address, timeout)
	if err != nil {
		result.Open = false
		result.Error = err
		result.Time = time.Since(startTime)
		return result
	}

	defer conn.Close()

	// 发送一个空的 UDP 包
	_, err = conn.Write([]byte{})
	if err != nil {
		result.Open = false
		result.Error = err
		result.Time = time.Since(startTime)
		return result
	}

	// 设置读取超时
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		result.Open = false
		result.Error = err
		result.Time = time.Since(startTime)
		return result
	}

	// 尝试读取响应
	buff := make([]byte, 1024)
	_, err = conn.Read(buff)
	result.Time = time.Since(startTime)

	// 如果收到 ICMP 不可达错误，端口可能是关闭的
	// 如果超时，端口可能是开放的但没有响应
	// 如果收到响应，端口肯定是开放的
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// 超时可能意味着端口是开放的但没有响应
			result.Open = true
			result.Error = nil
		} else {
			// 其他错误可能意味着端口是关闭的
			result.Open = false
			result.Error = err
		}
	} else {
		// 收到响应，端口是开放的
		result.Open = true
	}

	return result
}

// BatchScanUDPPorts 批量扫描多个主机的多个 UDP 端口
func BatchScanUDPPorts(hosts []string, ports []int, concurrency int, timeout time.Duration) []UDPScanResult {
	totalScans := len(hosts) * len(ports)
	results := make([]UDPScanResult, 0, totalScans)
	resultsChan := make(chan UDPScanResult, totalScans)

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

				resultsChan <- ScanUDPPort(h, p, timeout)
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
