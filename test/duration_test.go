package test

import (
	"campus2/pkg/utils"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "测试年",
			input:    "2y",
			expected: time.Hour * 24 * 365 * 2,
		},
		{
			name:     "测试月",
			input:    "3M",
			expected: time.Hour * 24 * 30 * 3,
		},
		{
			name:     "测试周",
			input:    "2w",
			expected: time.Hour * 24 * 7 * 2,
		},
		{
			name:     "测试天",
			input:    "7d",
			expected: time.Hour * 24 * 7,
		},
		{
			name:     "测试小时",
			input:    "68h",
			expected: time.Hour * 68,
		},
		{
			name:     "测试分钟",
			input:    "33m",
			expected: time.Minute * 33,
		},
		{
			name:     "测试秒",
			input:    "19s",
			expected: time.Second * 19,
		},
		{
			name:     "测试毫秒",
			input:    "2900ms",
			expected: time.Millisecond * 2900,
		},
		{
			name:     "测试组合-天时",
			input:    "3d2h",
			expected: time.Hour*24*3 + time.Hour*2,
		},
		{
			name:     "测试组合-年月日",
			input:    "1y2M3d",
			expected: time.Hour * 24 * (365 + 60 + 3),
		},
		{
			name:     "测试组合-时分",
			input:    "1h30m",
			expected: time.Hour + time.Minute*30,
		},
		{
			name:     "测试组合-时分秒",
			input:    "2h30m45s",
			expected: time.Hour*2 + time.Minute*30 + time.Second*45,
		},
		{
			name:     "测试小数点",
			input:    "1.5d",
			expected: time.Hour * 24 * 3 / 2,
		},
		{
			name:     "测试复杂组合",
			input:    "1y2M3w4d5h6m7s8ms",
			expected: time.Hour*24*365 + time.Hour*24*60 + time.Hour*24*21 + time.Hour*24*4 + time.Hour*5 + time.Minute*6 + time.Second*7 + time.Millisecond*8,
		},
		{
			name:    "测试错误格式-无单位",
			input:   "123",
			wantErr: true,
		},
		{
			name:    "测试错误格式-无数字",
			input:   "y",
			wantErr: true,
		},
		{
			name:    "测试错误格式-非法单位",
			input:   "1x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseDurationExamples(t *testing.T) {
	// 打印一些示例结果，方便直观查看
	examples := []string{
		"2y",
		"3M",
		"2w",
		"7d",
		"68h",
		"33m",
		"19s",
		"2900ms",
		"3d2h",
		"1y2M3d",
		"1h30m",
		"2h30m45s",
		"1.5d",
	}

	for _, example := range examples {
		duration, err := utils.ParseDuration(example)
		if err != nil {
			t.Logf("%s = 错误: %v", example, err)
			continue
		}
		t.Logf("%s = %v", example, duration)
	}
}
