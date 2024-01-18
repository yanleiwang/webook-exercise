package ratelimit

import (
	_ "embed"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"time"
)

var (
	//go:embed lua/slide_window.lua
	luaRedisSlideWindow string
)

type RedisSlideWindowLimiter struct {
	cmd        redis.Cmdable
	windowSize int64 // 窗口大小 毫秒计数
	threshold  int64
}

func NewRedisSlideWindowLimiter(cmd redis.Cmdable, windowSize time.Duration, threshold int64) *RedisSlideWindowLimiter {
	return &RedisSlideWindowLimiter{cmd: cmd, windowSize: windowSize.Milliseconds(), threshold: threshold}
}

func (r *RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaRedisSlideWindow, []string{key}, r.windowSize, r.threshold, time.Now().UnixMilli()).Bool()
}
