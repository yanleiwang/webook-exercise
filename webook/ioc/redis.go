package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type Config struct {
		addr     string `yaml:"addr"`
		password string `yaml:"password"`
	}
	var c Config
	err := viper.UnmarshalKey("redis", &c)
	if err != nil {
		panic(err)
	}

	return redis.NewClient(&redis.Options{
		Addr:     c.addr,
		Password: c.password, // 没有密码，默认值
	})
}
