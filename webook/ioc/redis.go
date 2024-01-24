package ioc

import (
	"gitee.com/geekbang/basic-go/webook/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: "", // 没有密码，默认值
	})
}
