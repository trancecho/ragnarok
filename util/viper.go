package util

import "flag"
import "github.com/spf13/viper"

func InitViper() {
	var mode string
	flag.StringVar(&mode, "mode", "dev", "运行模式")
	flag.Parse()
	if mode == "dev" {
		// 开发模式
		viper.SetConfigName("config.dev")
	} else if mode == "prod" {
		// 生产模式
		viper.SetConfigName("config.prod")
	} else {
	}

}
