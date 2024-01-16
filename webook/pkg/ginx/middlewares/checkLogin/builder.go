package checkLogin

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Builder struct {
	ignorePaths    mapset.Set[string]
	expiration     int
	updateInterval int64
}

// NewBuilder
// expiration: session 过期时间
// updateInterval: 每次刷新session 的时间间隔
func NewBuilder(expiration time.Duration, updateInterval time.Duration) *Builder {
	return &Builder{
		ignorePaths:    mapset.NewSet[string](),
		expiration:     int(expiration.Seconds()),
		updateInterval: updateInterval.Milliseconds(),
	}

}

func (b *Builder) IgnorePath(path string) *Builder {
	b.ignorePaths.Add(path)
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 不需要登录验证
		if b.ignorePaths.ContainsOne(ctx.FullPath()) {
			return
		}

		// 拿到session
		sess := sessions.Default(ctx)
		id := GetUserId(ctx)
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		const timeKey = "updateTime"
		val := sess.Get(timeKey)
		now := time.Now().UnixMilli()
		updateTime, ok := val.(int64)
		// 处于演示效果，整个 session 的过期时间是 1 分钟，所以我这里十秒钟刷新一次。
		// val == nil 是说明刚登录成功
		// 我们不在登录里面初始化这个 update_time，是因为它属于"刷新"机制，而不属于登录机制
		if val == nil || (ok && now-updateTime > b.updateInterval) {
			sess.Options(sessions.Options{
				MaxAge: b.expiration,
			})
			sess.Set(timeKey, now)
			err := sess.Save()
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if val != nil && !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

	}
}

const userIdKey = "userId"

func GetUserId(ctx *gin.Context) interface{} {
	sess := sessions.Default(ctx)

	id := sess.Get(userIdKey)
	return id
}

func SetUserId(ctx *gin.Context, id any, maxAge int) error {
	sess := sessions.Default(ctx)
	sess.Set(userIdKey, id)
	sess.Options(sessions.Options{MaxAge: maxAge})
	return sess.Save()
}
