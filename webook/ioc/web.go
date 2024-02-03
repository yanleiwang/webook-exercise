package ioc

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/internal/web/jwt"
	"gitee.com/geekbang/basic-go/webook/internal/web/middlewares"
	logger2 "gitee.com/geekbang/basic-go/webook/pkg/ginx/middlewares/logger"
	"gitee.com/geekbang/basic-go/webook/pkg/ginx/middlewares/ratelimit"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	ratelimit2 "gitee.com/geekbang/basic-go/webook/pkg/utils/ratelimit"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler) *gin.Engine {

	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterHandlers(server)
	oauth2WechatHdl.RegisterHandlers(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHdl jwt.Handler, log logger.Logger) []gin.HandlerFunc {

	bd := logger2.NewBuilder(func(ctx context.Context, al *logger2.AccessLog) {
		log.Debug("HTTP请求", logger.Field{Key: "al", Value: al})
	}).AllowReqBody(true).AllowRespBody()
	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logreq")
		bd.AllowReqBody(ok)
	})

	return []gin.HandlerFunc{
		corsHdl(),
		bd.Build(),
		// jwt 登录校验
		middlewares.NewJWTLoginMiddlewareBuilder(jwtHdl).
			IgnorePath("/users/signup").
			IgnorePath("/users/login").
			IgnorePath("/users/login_sms/code/send").
			IgnorePath("/users/login_sms").
			IgnorePath("/users/login").
			IgnorePath("users/refresh_token").
			IgnorePath("/oauth2/wechat/authurl").
			IgnorePath("/oauth2/wechat/callback").Build(),
		ratelimit.NewBuilder(ratelimit2.NewRedisSlideWindowLimiter(redisClient, time.Second, 100)).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool { //  哪些来源的url是被允许的
			return strings.HasPrefix(origin, "http://localhost")
		},
		AllowHeaders:     []string{"Content-Type", "Authorization"},  // 跨域请求能带上哪些header
		AllowCredentials: true,                                       // 是否允许带cookie
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"}, // 前端除了 normal header 还能拿到哪些响应header
		MaxAge:           12 * time.Hour,                             // preflight响应 过期时间
	})
}
