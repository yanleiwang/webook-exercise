package main

import (
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/internal/web/middlewares"
	ratelimit2 "gitee.com/geekbang/basic-go/webook/pkg/ginx/middlewares/ratelimit"
	"gitee.com/geekbang/basic-go/webook/pkg/utils/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	redisCmd := initRedis()
	engine := initWebServer(redisCmd)

	db := initDB()
	initUser(db, engine)
	err := engine.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}
}

func initRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "", // 没有密码，默认值
	})
}

func initUser(db *gorm.DB, engine *gin.Engine) {
	userDao := dao.NewUserDaoGorm(db)
	userRepo := repository.NewUserRepoImpl(userDao)
	userSvc := service.NewUserServiceImpl(userRepo)
	userHdl := web.NewUserHandler(userSvc)
	userHdl.RegisterHandlers(engine)
}

func initDB() *gorm.DB {
	/*
		&gorm.Config{
				TranslateError: true,
			}) 的目的是
		为了把 mysql 唯一索引冲突 错误转换为 gorm.ErrDuplicatedKey

		也可以 用 MySQL GO 驱动的 error 定义，找到准确的错误

			err := ud.db.WithContext(ctx).Create(&u).Error
			if me, ok := err.(*mysql.MySQLError); ok {
				const uniqueIndexErrNo uint16 = 1062
				if me.Number == uniqueIndexErrNo {
					return ErrUserDuplicateEmail
				}
			}
	*/
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/webook"), &gorm.Config{
		TranslateError: true,
	})

	if err != nil {
		panic(err)
	}

	dao.InitDB(db)
	return db
}

func initWebServer(cmd redis.Cmdable) *gin.Engine {
	engine := gin.Default()
	// cors 跨域
	engine.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool { //  哪些来源的url是被允许的
			return strings.HasPrefix(origin, "http://localhost")
		},
		AllowHeaders:     []string{"Content-Type", "Authorization"}, // 跨域请求能带上哪些header
		AllowCredentials: true,                                      // 是否允许带cookie
		ExposeHeaders:    []string{"x-jwt-token"},                   // 前端除了 normal header 还能拿到哪些响应header
		MaxAge:           12 * time.Hour,                            // preflight响应 过期时间
	}))

	// session + cookie 登录校验
	//store, _ := redis.NewStore(10, "tcp", "localhost:16379", "",
	//	[]byte("pY8tX3vY7aT8nK2nD6lO9jR4pE5aN4gI"), []byte("rM8eL5rB7pC1fZ4tZ3eT1fM8cS5kK7lD"))
	//engine.Use(sessions.Sessions("mysession", store))
	//
	//engine.Use(middlewares.NewSessionLoginBuilder(time.Minute, time.Second*10).
	//	IgnorePath("/users/signup").
	//	IgnorePath("/users/login").Build())

	// session
	store := memstore.NewStore([]byte("pY8tX3vY7aT8nK2nD6lO9jR4pE5aN4gI"), []byte("rM8eL5rB7pC1fZ4tZ3eT1fM8cS5kK7lD"))
	engine.Use(sessions.Sessions("mysession", store))

	// jwt 登录校验
	engine.Use(middlewares.NewJWTLoginMiddlewareBuilder().
		IgnorePath("/users/signup").
		IgnorePath("/users/login").Build())

	// 限流  一分钟 20个 测试用
	rateLimit := ratelimit.NewRedisSlideWindowLimiter(cmd, time.Minute, 20)
	engine.Use(ratelimit2.NewBuilder(rateLimit).Build())
	return engine
}
