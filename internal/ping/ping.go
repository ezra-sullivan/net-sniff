package ping

import (
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// PingResult 存储 ping 的结果
type PingResult struct {
	Host    string
	Success bool
	Time    time.Duration
	Error   error
}

// SinglePing 对单个主机执行 ping 操作
func SinglePing(host string, timeout time.Duration) PingResult {
	result := PingResult{
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

	return result
}

// BatchPing 对多个主机执行批量 ping 操作
func BatchPing(hosts []string, concurrency int, timeout time.Duration) []PingResult {
	results := make([]PingResult, 0, len(hosts))
	resultsChan := make(chan PingResult, len(hosts))

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
