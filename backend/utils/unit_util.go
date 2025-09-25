package utils

import (
	"log"
	"strconv"
	"strings"
)

// ParsePacketCount 解析包数，支持K/M/G单位
func ParsePacketCount(countStr string) uint64 {
	if countStr == "" || countStr == "--" {
		return 0
	}

	// 移除可能的逗号
	countStr = strings.ReplaceAll(countStr, ",", "")

	// 检查是否有单位后缀
	if len(countStr) == 0 {
		return 0
	}

	lastChar := countStr[len(countStr)-1]
	var multiplier uint64 = 1
	var numStr string

	switch lastChar {
	case 'K', 'k':
		multiplier = 1000
		numStr = countStr[:len(countStr)-1]
	case 'M', 'm':
		multiplier = 1000000
		numStr = countStr[:len(countStr)-1]
	case 'G', 'g':
		multiplier = 1000000000
		numStr = countStr[:len(countStr)-1]
	default:
		// 没有单位，直接解析
		numStr = countStr
	}

	// 解析数值部分
	if numStr == "" {
		return 0
	}

	// 支持小数点（如 2.3K）
	if strings.Contains(numStr, ".") {
		floatVal, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			log.Printf("[WARN] Failed to parse float packet count '%s': %v", numStr, err)
			return 0
		}
		return uint64(floatVal * float64(multiplier))
	}

	// 整数解析
	intVal, err := strconv.ParseUint(numStr, 10, 64)
	if err != nil {
		log.Printf("[WARN] Failed to parse packet count '%s': %v", numStr, err)
		return 0
	}

	return intVal * multiplier
}
