package startup

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type Config struct {
		addr     string `yaml:"addr"`
		password string `yaml:"password"`
	}

	var cfg Config
	viper.UnmarshalKey("redis", &cfg)
	//var cfg Config = Config{
	//	addr:     "localhost:6379",
	//	password: "",
	//}
	return redis.NewClient(&redis.Options{
		Addr:     cfg.addr,
		Password: cfg.password,
	})
}
