package pscan

import (
	"errors"
	"fmt"
	"github.com/ezra-sullivan/net-sniff/internal/global"
	"net"
	"sync"
	"time"
)

const (
	UDP_PORT_CLOSED           uint8 = 0
	UDP_PORT_OPEN             uint8 = 1
	UDP_PORT_OPEN_OR_FILTERED uint8 = 2
)

// UDPScanResult 存储 UDP 端口扫描结果
type UDPScanResult struct {
	Host   string
	Port   int
	IsOpen uint8
	Error  error
	Time   time.Duration
}

// ScanUDPPort 扫描单个 UDP 端口
func ScanUDPPort(host string, port int, timeout time.Duration) UDPScanResult {

	address := fmt.Sprintf("%s:%d", host, port)
	startTime := time.Now()

	conn, err := net.DialTimeout("udp", address, timeout)

	// 在连接尝试后立即计算时间
	result := UDPScanResult{
		Host:   host,
		Port:   port,
		Error:  err,
		IsOpen: UDP_PORT_CLOSED,
	}

	if err != nil || conn == nil {
		result.Time = time.Since(startTime)
		return result
	}

	defer func() {
		if err = conn.Close(); err != nil {
			result.Error = fmt.Errorf("关闭连接失败: %w", err)
		}
	}()

	// 发送一个的 UDP 包（DNS）
	_, err = conn.Write([]byte{
		0x12, 0x34, // ID
		0x01, 0x00, // 标准查询
		0x00, 0x01, // 1 个问题
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 无答案等
		0x03, 'w', 'w', 'w',
		0x06, 'g', 'o', 'o', 'g', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // 终止
		0x00, 0x01, // 类型 A
		0x00, 0x01, // 类 IN
	})
	if err != nil {
		result.Error = err
		result.Time = time.Since(startTime)
		return result
	}

	// 设置读取超时
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		result.IsOpen = 0
		result.Error = err
		result.Time = time.Since(startTime)
		return result
	}

	// 尝试读取响应
	buff := make([]byte, 1024)
	_, err = conn.Read(buff)
	result.Time = time.Since(startTime)

	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			// 超时无响应：开放或过滤
			result.IsOpen = UDP_PORT_OPEN_OR_FILTERED
			result.Error = nil
		}
	} else {
		// 收到数据包：开放
		result.IsOpen = UDP_PORT_OPEN
	}

	result.output()

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

func (result *UDPScanResult) output() {
	// 获取全局的 consoleLogger
	consoleLogger := global.ConsoleLogger

	if result.IsOpen == UDP_PORT_OPEN {
		consoleLogger.Info("UDP Port Scan Result ",

			"host", result.Host,
			"port", result.Port,
			"status", "open",
			"time", result.Time)
	} else if result.IsOpen == UDP_PORT_CLOSED {
		consoleLogger.Error("UDP Port Scan Result ",
			"host", result.Host,
			"port", result.Port,
			"status", "closed",
			"time", result.Time,
			"err", result.Error)
	} else {
		consoleLogger.Error("UDP Port Scan Result ",
			"host", result.Host,
			"port", result.Port,
			"status", "open|filtered",
		)
	}

	// 获取全局的 fileLogger
	fileLogger := global.FileLogger
	// 如果 fileLogger 不为空
	if fileLogger != nil {
		if result.IsOpen == UDP_PORT_OPEN {
			// 打印 TCP 端口开放结果到文件
			fileLogger.Info(fmt.Sprintf("%s,%d,%s,%.2f\n", result.Host, result.Port, "open", float64(result.Time.Microseconds())/1000.0))
		} else if result.IsOpen == UDP_PORT_CLOSED {
			// 打印TCP端口关闭结果到文件
			fileLogger.Debug(fmt.Sprintf("%s,%d,%s,%.2f\n", result.Host, result.Port, "closed", float64(result.Time.Microseconds())/1000.0))
		} else {
			fileLogger.Debug(fmt.Sprintf("%s,%d,%s,%.2f\n", result.Host, result.Port, "open|filtered", float64(result.Time.Microseconds())/1000.0))
		}
	}
}
