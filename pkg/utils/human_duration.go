package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseDuration(d string) (time.Duration, error) {
	d = strings.TrimSpace(d)

	// 先尝试标准解析
	if dr, err := time.ParseDuration(d); err == nil {
		return dr, nil
	}

	var total time.Duration
	var currentNum string

	// 遍历字符串，分别处理数字和单位
	for i := 0; i < len(d); i++ {
		c := d[i]
		if c >= '0' && c <= '9' || c == '.' {
			currentNum += string(c)
			continue
		}

		// 如果没有数字，返回错误
		if currentNum == "" {
			return 0, fmt.Errorf("无效的时间格式: %s", d)
		}

		// 解析数字部分
		num, err := strconv.ParseFloat(currentNum, 64)
		if err != nil {
			return 0, fmt.Errorf("无效的数字: %s", currentNum)
		}

		// 根据单位计算时长
		var multiplier time.Duration
		switch c {
		case 'y':
			multiplier = time.Hour * 24 * 365
			i++ // 跳过单位字符
		case 'M':
			multiplier = time.Hour * 24 * 30
			i++ // 跳过单位字符
		case 'w':
			multiplier = time.Hour * 24 * 7
			i++ // 跳过单位字符
		case 'd':
			multiplier = time.Hour * 24
			i++ // 跳过单位字符
		case 'h':
			multiplier = time.Hour
			i++ // 跳过单位字符
		case 'm':
			if i+1 < len(d) && d[i+1] == 's' {
				multiplier = time.Millisecond
				i++ // 额外跳过's'
			} else {
				multiplier = time.Minute
			}
			i++ // 跳过单位字符
		case 's':
			multiplier = time.Second
			i++ // 跳过单位字符
		default:
			return 0, fmt.Errorf("无效的时间单位: %c", c)
		}

		total += time.Duration(float64(multiplier) * num)
		currentNum = "" // 重置数字部分
		i--             // 因为for循环会自增，这里减一以抵消
	}

	// 处理最后一个数字（如果有的话）
	if currentNum != "" {
		return 0, fmt.Errorf("数字 %s 后缺少单位", currentNum)
	}

	return total, nil
}
