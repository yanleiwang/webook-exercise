package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	server := InitWebServer()

	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func initViper() {
	cfile := pflag.String("config",
		"config/config.yaml", "指定配置文件路径")

	type Config struct {
		name     string `yaml:"name"`
		addr     string `yaml:"addr"`
		password string `yaml:"password"`
	}

	pflag.Parse()
	viper.SetConfigFile(*cfile)

	// 只能告诉你文件变了，不能告诉你，文件的哪些内容变了
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("db.dsn"))
	})
	// 实时监听配置变更
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3",
		// 通过 webook 和其他使用 etcd 的区别出来
		"http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println(in.Name, in.Op)
	})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}
