package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ParseHostList 解析主机列表，支持逗号分隔、文件路径和 CIDR 格式
func ParseHostList(hosts string) ([]string, error) {
	// 如果是文件路径，从文件读取
	if _, err := os.Stat(hosts); err == nil {
		return readHostsFromFile(hosts)
	}

	// 否则按逗号分隔处理
	hostList := strings.Split(hosts, ",")
	result := make([]string, 0, len(hostList))

	for _, host := range hostList {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}

		// 检查是否是 CIDR 格式
		if isCIDR(host) {
			ipList, err := expandCIDR(host)
			if err != nil {
				return nil, err
			}
			result = append(result, ipList...)
			continue
		}

		// 检查是否是 IP 范围格式 (如 192.168.1.1-10)
		if strings.Contains(host, "-") {
			ipList, err := expandIPRange(host)
			if err != nil {
				return nil, err
			}
			result = append(result, ipList...)
			continue
		}

		// 单个主机
		if net.ParseIP(host) != nil || isValidHostname(host) {
			// 直接添加有效的 IP 地址或主机名
			result = append(result, host)
		} else if isInvalidIPFormat(host) {
			// 检查是否看起来像 IP 地址但格式不正确
			return nil, fmt.Errorf("无效的 IP 地址格式: %s", host)
		} else {
			// 既不是有效的 IP 也不是有效的主机名
			return nil, fmt.Errorf("无效的 IP 地址或主机名: %s", host)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("未指定有效的主机")
	}

	return result, nil
}

// isInvalidIPFormat 检查字符串是否看起来像 IP 地址但格式不正确
func isInvalidIPFormat(s string) bool {
	// 检查是否包含点，看起来像 IP 地址
	if strings.Count(s, ".") > 0 {
		// 使用正则表达式检查是否为有效的 IPv4 地址格式
		validIP := regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$`)
		return !validIP.MatchString(s)
	}
	return false
}

// readHostsFromFile 从文件读取主机列表
func readHostsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			os.Exit(1)
		}
	}(file)

	var hosts []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if host != "" && !strings.HasPrefix(host, "#") {
			// 验证主机格式
			if err := validateHost(host); err != nil {
				return nil, fmt.Errorf("文件中包含无效的主机: %s, %w", host, err)
			}
			hosts = append(hosts, host)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

// validateHost 验证主机是否为有效的 IP 地址或主机名
func validateHost(host string) error {
	// 检查是否为 CIDR 格式
	if isCIDR(host) {
		return nil
	}

	// 检查是否为 IP 范围
	if strings.Contains(host, "-") {
		return nil
	}

	// 检查是否为有效的 IPv4 地址
	ip := net.ParseIP(host)
	if ip == nil {
		// 检查是否看起来像 IP 地址但格式不正确
		if isInvalidIPFormat(host) {
			return fmt.Errorf("无效的 IP 地址格式: %s", host)
		}

		// 不是有效的 IP 地址，检查是否为有效的主机名
		if !isValidHostname(host) {
			return fmt.Errorf("不是有效的 IP 地址或主机名: %s", host)
		}
	}
	return nil
}

// isValidHostname 检查是否为有效的主机名
func isValidHostname(hostname string) bool {
	// 主机名规则: 只能包含字母、数字、连字符和点，且不能以连字符或点开头或结尾
	if len(hostname) > 255 {
		return false
	}

	// 检查主机名格式
	validHostname := regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)
	return validHostname.MatchString(hostname)
}

// isCIDR 检查字符串是否是 CIDR 格式
func isCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// expandCIDR 展开 CIDR 格式为 IP 列表
func expandCIDR(cidr string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	// 确保 IP 是 IPv4 格式
	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("只支持 IPv4 地址")
	}

	// 复制 IP 以避免修改原始值
	startIP := make(net.IP, len(ip))
	copy(startIP, ip)

	// 计算网段中的 IP 数量
	mask := ipNet.Mask
	ones, bits := mask.Size()
	size := 1 << uint(bits-ones)

	// 生成所有 IP
	for i := 0; i < size; i++ {
		ips = append(ips, startIP.String())
		incrementIP(startIP)
	}

	// 移除网络地址和广播地址（如果有超过2个地址）
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

// incrementIP 增加 IP 地址
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// expandIPRange 展开 IP 范围格式为 IP 列表
func expandIPRange(ipRange string) ([]string, error) {
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的 IP 范围格式: %s", ipRange)
	}

	baseIP := parts[0]
	endRange := parts[1]

	// 检查是否是完整的 IP 地址范围 (如 192.168.1.1-192.168.1.10)
	if strings.Count(endRange, ".") == 3 {
		startIP := net.ParseIP(baseIP)
		endIP := net.ParseIP(endRange)
		if startIP == nil || endIP == nil {
			return nil, fmt.Errorf("无效的 IP 地址范围: %s", ipRange)
		}
		return expandFullIPRange(startIP, endIP)
	}

	// 处理简化的范围 (如 192.168.1.1-10)
	ipParts := strings.Split(baseIP, ".")
	if len(ipParts) != 4 {
		return nil, fmt.Errorf("无效的 IP 地址格式: %s", baseIP)
	}

	// 获取基础 IP 的最后一部分
	lastPart, err := strconv.Atoi(ipParts[3])
	if err != nil {
		return nil, fmt.Errorf("无效的 IP 地址格式: %s", baseIP)
	}

	// 获取范围的结束值
	endPart, err := strconv.Atoi(endRange)
	if err != nil {
		return nil, fmt.Errorf("无效的 IP 范围结束值: %s", endRange)
	}

	if lastPart > endPart {
		return nil, fmt.Errorf("IP 范围结束值必须大于或等于起始值")
	}

	var ips []string
	basePrefix := strings.Join(ipParts[:3], ".") + "."
	for i := lastPart; i <= endPart; i++ {
		ips = append(ips, basePrefix+strconv.Itoa(i))
	}

	return ips, nil
}

// expandFullIPRange 展开完整的 IP 地址范围
func expandFullIPRange(startIP, endIP net.IP) ([]string, error) {
	var ips []string
	ip := make(net.IP, len(startIP))
	copy(ip, startIP)

	// 确保 IP 是 IPv4 格式
	ip = ip.To4()
	endIP = endIP.To4()
	if ip == nil || endIP == nil {
		return nil, fmt.Errorf("只支持 IPv4 地址范围")
	}

	// 比较 IP 地址
	if bytes.Compare(ip, endIP) > 0 {
		return nil, fmt.Errorf("结束 IP 地址必须大于或等于起始 IP 地址")
	}

	for bytes.Compare(ip, endIP) <= 0 {
		ips = append(ips, ip.String())
		incrementIP(ip)
	}

	return ips, nil
}
