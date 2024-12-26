package pkg

import (
	"campus2/pkg/global"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"context"

	"github.com/sirupsen/logrus"
)

func NewLogrus(ctx context.Context) *logrus.Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(global.GVA_CONFIG.Logrus.Level)
	if err != nil {
		fmt.Printf("无效的日志级别: %v，使用默认级别info\n", global.GVA_CONFIG.Logrus.Level)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	switch global.GVA_CONFIG.Logrus.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: false,
			PrettyPrint:      true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", f.File, f.Line)
			},
		})
	case "console":
		fallthrough
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			ForceColors:            true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			DisableQuote:           true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", f.File, f.Line)
			},
		})
	}

	// 设置是否显示堆栈信息
	if global.GVA_CONFIG.Logrus.ShowStacktrace {
		logger.AddHook(&stackHook{})
	}

	// 设置日志输出
	writers := []io.Writer{}
	if global.GVA_CONFIG.Logrus.LogInConsole {
		writers = append(writers, os.Stdout)
	}

	// 创建当天的日志目录
	today := time.Now().Format("2006-01-02")
	logDir := filepath.Join(global.GVA_CONFIG.Logrus.Directory, today)

	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
		return logger
	}

	// 为每个可能的日志级别创建一个写入器
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}

	for _, lvl := range levels {
		if lvl <= level { // 只创建小于等于设置级别的日志写入器
			writers = append(writers, &LazyFileWriter{
				LogDir: logDir,
				Level:  lvl,
			})
		}
	}

	// 使用MultiWriter将日志同时写入所有输出
	logger.SetOutput(io.MultiWriter(writers...))

	// 显示调用方法
	logger.SetReportCaller(global.GVA_CONFIG.Logrus.ShowLine)

	// 如果设置了日志保留天数，启动一个清理协程
	if global.GVA_CONFIG.Logrus.RetentionDay > 0 {
		logger.Infof("启动日志清理任务，保留天数: %d", global.GVA_CONFIG.Logrus.RetentionDay)
		go func() {
			for {
				// 每天凌晨执行清理
				now := time.Now()
				next := now.Add(24 * time.Hour)
				next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
				waitDuration := next.Sub(now)

				logger.Infof("下次日志清理时间: %v (等待 %v)", next.Format("2006-01-02 15:04:05"), waitDuration)
				timer := time.NewTimer(waitDuration)

				select {
				case <-ctx.Done():
					timer.Stop()
					logger.Info("收到退出信号，停止日志清理任务")
					return
				case <-timer.C:
					logger.Info("开始执行日志清理任务...")
					startTime := time.Now()

					// 清理过期日志
					filesCount, err := cleanExpiredLogs(global.GVA_CONFIG.Logrus.Directory, global.GVA_CONFIG.Logrus.RetentionDay)
					if err != nil {
						logger.Errorf("日志清理失败: %v", err)
					} else {
						logger.Infof("日志清理完成，共清理 %d 个文件，耗时: %v", filesCount, time.Since(startTime))
					}
				}
			}
		}()
	} else {
		logger.Info("日志保留天数未设置或无效，不启动清理任务")
	}

	return logger
}

// LazyFileWriter 用于延迟创建日志文件，只在实际写入时创建
type LazyFileWriter struct {
	LogDir string
	Level  logrus.Level
	file   *os.File
}

func (w *LazyFileWriter) Write(p []byte) (n int, err error) {
	logText := string(p)

	// 查找实际的日志级别文本
	var found bool

	// 通过前缀解析（跳过ANSI颜色代码）
	textBytes := []byte(logText)
	for i := 0; i < len(textBytes)-5; i++ {
		if textBytes[i] == 'I' && textBytes[i+1] == 'N' && textBytes[i+2] == 'F' && textBytes[i+3] == 'O' {
			if w.Level != logrus.InfoLevel {
				return len(p), nil
			}
			found = true
			break
		} else if textBytes[i] == 'W' && textBytes[i+1] == 'A' && textBytes[i+2] == 'R' && textBytes[i+3] == 'N' {
			if w.Level != logrus.WarnLevel {
				return len(p), nil
			}
			found = true
			break
		} else if textBytes[i] == 'E' && textBytes[i+1] == 'R' && textBytes[i+2] == 'R' && textBytes[i+3] == 'O' {
			if w.Level != logrus.ErrorLevel {
				return len(p), nil
			}
			found = true
			break
		} else if textBytes[i] == 'D' && textBytes[i+1] == 'E' && textBytes[i+2] == 'B' && textBytes[i+3] == 'U' {
			if w.Level != logrus.DebugLevel {
				return len(p), nil
			}
			found = true
			break
		} else if textBytes[i] == 'T' && textBytes[i+1] == 'R' && textBytes[i+2] == 'A' && textBytes[i+3] == 'C' {
			if w.Level != logrus.TraceLevel {
				return len(p), nil
			}
			found = true
			break
		} else if textBytes[i] == 'F' && textBytes[i+1] == 'A' && textBytes[i+2] == 'T' && textBytes[i+3] == 'A' {
			if w.Level != logrus.FatalLevel {
				return len(p), nil
			}
			found = true
			break
		} else if textBytes[i] == 'P' && textBytes[i+1] == 'A' && textBytes[i+2] == 'N' && textBytes[i+3] == 'I' {
			if w.Level != logrus.PanicLevel {
				return len(p), nil
			}
			found = true
			break
		}
	}

	if !found {
		return len(p), nil
	}

	// 延迟创建文件，只在第一次写入对应级别的日志时创建
	if w.file == nil {
		fileName := filepath.Join(w.LogDir, w.Level.String()+".log")
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, fmt.Errorf("无法创建日志文件: %v", err)
		}
		w.file = file
	}

	return w.file.Write(p)
}

// 清理过期日志文件
func cleanExpiredLogs(logDir string, retentionDays int) (int, error) {
	// 获取截止日期
	deadline := time.Now().AddDate(0, 0, -retentionDays)

	// 遍历日志目录
	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 如果文件修改时间早于截止日期，删除文件
		if info.ModTime().Before(deadline) {
			if err := os.Remove(path); err != nil {
				fmt.Printf("删除过期日志文件失败: %v\n", err)
			} else {
				fmt.Printf("已删除过期日志文件: %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("清理过期日志文件时出错: %v\n", err)
	}

	return 0, err
}

// stackHook 用于添加堆栈信息的钩子
type stackHook struct{}

func (h *stackHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}
}

func (h *stackHook) Fire(entry *logrus.Entry) error {
	// 只为错误级别及以上的日志添加堆栈信息
	if entry.Level <= logrus.ErrorLevel {
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)

		// 格式化堆栈信息
		stackStr := string(stack[:length])
		lines := strings.Split(stackStr, "\n")

		// 直接将堆栈信息添加到消息中，而不是作为字段
		entry.Message += "\nStack Trace:\n" + strings.Join(lines, "\n")
	}
	return nil
}
