package middlewares

import (
	ijwt "gitee.com/geekbang/basic-go/webook/internal/web/jwt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type JWTLoginMiddlewareBuilder struct {
	ignorePath mapset.Set[string]
	ijwt.Handler
}

func NewJWTLoginMiddlewareBuilder(jwtHdl ijwt.Handler) *JWTLoginMiddlewareBuilder {
	return &JWTLoginMiddlewareBuilder{
		ignorePath: mapset.NewSet[string](),
		Handler:    jwtHdl,
	}
}

func (j *JWTLoginMiddlewareBuilder) IgnorePath(path string) *JWTLoginMiddlewareBuilder {
	j.ignorePath.Add(path)
	return j
}

func (j *JWTLoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if j.ignorePath.ContainsOne(ctx.FullPath()) {
			return
		}

		uc, err := j.ExtractAccessClaims(ctx)
		if err != nil || uc.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime, err := uc.GetExpirationTime()
		if err != nil {
			// 拿不到过期时间
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if expireTime.Before(time.Now()) {
			// 已经过期
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if uc.UserAgent != ctx.Request.UserAgent() {
			// 换 了一个 浏览器  可能是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = j.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// 要么 redis 有问题，要么已经退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 说明 token 是合法的
		// 我们把这个 token 里面的数据放到 ctx 里面，后面用的时候就不用再次 Parse 了
		ctx.Set(ijwt.KeyAccessClaims, &uc)

	}
}
