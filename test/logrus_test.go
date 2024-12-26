package test

import (
	"campus2/pkg/global"
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	fmt.Println("testRun")
	global.GVA_LOG.Info("测试 info 级别")
	global.GVA_LOG.Error("测试 error 级别")
	global.GVA_LOG.Warn("测试 warn 级别")
}
