package cache

import (
	"encoding/json"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"time"
)

//go:generate mockgen  -destination=redismocks/mock_redis_cmdable.gen.go -package=redismocks github.com/redis/go-redis/v9 Cmdable

//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE --package=$GOPACKAGEmocks
type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, user domain.User) error
}

var ErrKeyNotExist = redis.Nil

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (r *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	result, err := r.cmd.Get(ctx, r.genKey(id)).Result()
	if err != nil {
		return domain.User{}, err
	}
	var user domain.User
	err = json.Unmarshal([]byte(result), &user)
	return user, err
}

func (r *RedisUserCache) Set(ctx context.Context, user domain.User) error {

	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return r.cmd.Set(ctx, r.genKey(user.Id), userJson, r.expiration).Err()

}

func (r *RedisUserCache) genKey(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
