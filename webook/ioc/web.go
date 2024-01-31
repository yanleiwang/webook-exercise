package ioc

import (
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/internal/web/middlewares"
	"gitee.com/geekbang/basic-go/webook/pkg/ginx/middlewares/ratelimit"
	ratelimit2 "gitee.com/geekbang/basic-go/webook/pkg/utils/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {

	// session + cookie 登录校验
	//store, _ := redis.NewStore(10, "tcp", "localhost:16379", "",
	//	[]byte("pY8tX3vY7aT8nK2nD6lO9jR4pE5aN4gI"), []byte("rM8eL5rB7pC1fZ4tZ3eT1fM8cS5kK7lD"))
	//engine.Use(sessions.Sessions("mysession", store))
	//
	//engine.Use(middlewares.NewSessionLoginBuilder(time.Minute, time.Second*10).
	//	IgnorePath("/users/signup").
	//	IgnorePath("/users/login").Build())

	// session
	//store := memstore.NewStore([]byte("pY8tX3vY7aT8nK2nD6lO9jR4pE5aN4gI"), []byte("rM8eL5rB7pC1fZ4tZ3eT1fM8cS5kK7lD"))
	//engine.Use(sessions.Sessions("mysession", store))

	return []gin.HandlerFunc{
		corsHdl(),
		// jwt 登录校验
		middlewares.NewJWTLoginMiddlewareBuilder().
			IgnorePath("/users/signup").
			IgnorePath("/users/login").
			IgnorePath("/users/login_sms/code/send").
			IgnorePath("/users/login_sms").
			IgnorePath("/users/login").
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
		AllowHeaders:     []string{"Content-Type", "Authorization"}, // 跨域请求能带上哪些header
		AllowCredentials: true,                                      // 是否允许带cookie
		ExposeHeaders:    []string{"x-jwt-token"},                   // 前端除了 normal header 还能拿到哪些响应header
		MaxAge:           12 * time.Hour,                            // preflight响应 过期时间
	})
}
