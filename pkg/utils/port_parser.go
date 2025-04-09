package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePortRange 解析端口范围
func ParsePortRange(ports string) ([]int, error) {
	var result []int

	// 解析逗号分隔的端口列表
	for _, portStr := range strings.Split(ports, ",") {
		portStr = strings.TrimSpace(portStr)
		if portStr == "" {
			continue
		}

		// 检查是否为端口范围
		if strings.Contains(portStr, "-") {
			// 解析端口范围
			rangeParts := strings.Split(portStr, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("无效的端口范围格式: %s", portStr)
			}

			startPort, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("无效的起始端口: %s", rangeParts[0])
			}

			endPort, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("无效的结束端口: %s", rangeParts[1])
			}

			// 验证端口范围
			if startPort < 1 || startPort > 65535 || endPort < 1 || endPort > 65535 {
				return nil, fmt.Errorf("端口必须在 1-65535 范围内")
			}

			if endPort < startPort {
				return nil, fmt.Errorf("结束端口必须大于或等于起始端口")
			}

			// 添加端口范围
			for port := startPort; port <= endPort; port++ {
				result = append(result, port)
			}
		} else {
			// 解析单个端口
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return nil, fmt.Errorf("无效的端口: %s", portStr)
			}

			// 验证端口
			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("端口必须在 1-65535 范围内")
			}

			result = append(result, port)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("未指定有效的端口")
	}

	return result, nil
}
