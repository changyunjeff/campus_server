package test

import (
	"campus2/pkg"
	"campus2/pkg/global"
	"campus2/pkg/redis"
	"context"
	"fmt"
)

func init() {
	fmt.Println("初始化 test 模块")

	// 加载配置文件，运行后就可以使用global.GVA_CONFIG来访问配置文件的内容了
	global.GVA_VIPER = pkg.NewViper("../configs/config-test.yaml")
	global.GVA_LOG = pkg.NewLogrus(context.Background())
	global.GVA_DB = pkg.GetDB(global.GVA_CONFIG.Mysql)
	global.GVA_REDIS = redis.GetRedis(global.GVA_CONFIG.Redis)
	fmt.Println("Logrus 初始化完成")

}
