package ratelimit

import (
	"fmt"
	"gitee.com/geekbang/basic-go/webook/pkg/utils"
	"gitee.com/geekbang/basic-go/webook/pkg/utils/ratelimit"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type Builder struct {
	limiter  ratelimit.Limiter
	genKeyFn func(ctx *gin.Context) string
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limiter.Limit(ctx, b.genKeyFn(ctx))
		if err != nil {
			slog.Error("限流器错误", "err", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		slog.Debug("", "limited", limited)
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

	}
}

func NewBuilder(limiter ratelimit.Limiter, opts ...utils.Option[Builder]) *Builder {
	ret := &Builder{
		limiter: limiter,
		genKeyFn: func(ctx *gin.Context) string {
			return fmt.Sprintf("limiter:ip::%s", ctx.ClientIP())
		},
	}

	utils.Apply[Builder](ret, opts...)
	return ret
}

func WithGenKeyFn(fn func(ctx *gin.Context) string) utils.Option[Builder] {
	return func(t *Builder) {
		t.genKeyFn = fn
	}
}
