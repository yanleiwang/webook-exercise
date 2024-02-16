package startup

import (
	"flag"
	"github.com/spf13/viper"
)

var cfile = flag.String("config",
	"config/dev.yaml", "指定配置文件路径")

func InitViper() {
	flag.Parse()
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
