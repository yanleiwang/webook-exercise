package main

import (
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/pkg/ginx/middlewares/checkLogin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	engine := initWebServer()

	db := initDB()
	initUser(db, engine)
	err := engine.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}
}

func initUser(db *gorm.DB, engine *gin.Engine) {
	userDao := dao.NewUserDaoGorm(db)
	userRepo := repository.NewUserRepoImpl(userDao)
	userSvc := service.NewUserServiceImpl(userRepo)
	userHdl := web.NewUserHandler(userSvc)
	userHdl.RegisterHandlers(engine)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic(err)
	}

	dao.InitDB(db)
	return db
}

func initWebServer() *gin.Engine {
	engine := gin.Default()
	engine.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool { //  哪些来源的url是被允许的
			return strings.HasPrefix(origin, "http://localhost")
		},
		AllowHeaders:     []string{"Content-Type", "Authorization"}, // 跨域请求能带上哪些header
		AllowCredentials: true,                                      // 是否允许带cookie
		ExposeHeaders:    []string{"x-jwt-token"},                   // 前端除了 normal header 还能拿到哪些响应header
		MaxAge:           12 * time.Hour,                            // preflight响应 过期时间
	}))

	store, _ := redis.NewStore(10, "tcp", "localhost:6379", "",
		[]byte("pY8tX3vY7aT8nK2nD6lO9jR4pE5aN4gI"), []byte("rM8eL5rB7pC1fZ4tZ3eT1fM8cS5kK7lD"))
	engine.Use(sessions.Sessions("mysession", store))

	engine.Use(checkLogin.NewBuilder(time.Minute, time.Second*10).
		IgnorePath("/users/signup").
		IgnorePath("/users/login").Build())
	return engine
}
