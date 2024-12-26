package main

import (
	"campus2/init"
	"campus2/pkg"
	"campus2/pkg/global"
	"fmt"
	"os"
)

func main() {
	// 加载配置
	global.GVA_DB = pkg.GetDB(global.GVA_CONFIG.Mysql)
	if global.GVA_DB != nil {
		// 程序结束前关闭数据库链接
		db, _ := global.GVA_DB.DB()
		defer db.Close()

		if err := init.RegisterTables(); err != nil {
			global.GVA_LOG.Error(err)
			os.Exit(0)
		}
	}
	routers := init.Routers()
	fmt.Println(routers)
}
