package pkg

import (
	"campus2/pkg/global"
	"flag"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func NewViper(path ...string) *viper.Viper {
	var config string

	if len(path) > 0 {
		config = path[0]
		fmt.Printf("您正在使用func NewViper()传递的参数值,config的路径为%s\n", config)
		goto START
	}

	// 检查是否已经解析过标志
	if !flag.Parsed() {
		// 从命令行参数中获取config配置文件的路径
		flag.StringVar(&config, "c", "", "choose config file.")
		flag.Parse()
	}

	if config == "" {
		// 能运行到这里说明运行程序的命令行中没传递-c参数
		configEnv := os.Getenv("GVA_CONFIG") // 获取Docker容器中的环境变量
		if configEnv == "" {
			switch gin.Mode() {
			case gin.DebugMode:
				config = "configs/config.yaml"
			case gin.ReleaseMode:
				config = "configs/config.yaml"
			case gin.TestMode:
				config = "configs/config.yaml"
			}
			fmt.Printf("您正在使用gin模式的%s环境名称,config的路径为%s\n", gin.Mode(), config)
		}
	} else {
		fmt.Printf("您正在使用命令行的-c参数传递的值,config的路径为%s\n", config)
	}
START:
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("读取配置文件时失败: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config配置文件内容发生了改变:", e.Name)
		if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
			panic(err)
		}
	})

	if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
		panic(err)
	}

	// 结构化打印出global.GVA_CONFIG内的所有数据
	fmt.Printf("配置文件内容: %+v \n", global.GVA_CONFIG)

	return v
}
