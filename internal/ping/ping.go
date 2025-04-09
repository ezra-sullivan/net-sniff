package ping

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// PingResult 存储 ping 的结果
type PingResult struct {
	Host    string
	Success bool
	Time    time.Duration
	Error   error
}

// PingHost 对单个主机执行 ping 操作
func PingHost(host string) PingResult {
	result := PingResult{
		Host: host,
	}

	// 使用系统 ping 命令
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", host)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", host)
	}

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	result.Time = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = err
		return result
	}

	// 检查输出是否包含成功的响应
	outputStr := string(output)
	if runtime.GOOS == "windows" {
		result.Success = strings.Contains(outputStr, "TTL=")
	} else {
		result.Success = strings.Contains(outputStr, " 0% packet loss")
	}

	return result
}

// BatchPing 对多个主机执行批量 ping 操作
func BatchPing(hosts []string, concurrency int) []PingResult {
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

			resultsChan <- PingHost(h)
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

// ResolveHostname 将主机名解析为 IP 地址
func ResolveHostname(hostname string) (string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", err
	}

	// 返回第一个 IPv4 地址
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv4 address found for %s", hostname)
}
