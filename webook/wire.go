//go:build wireinject
// +build wireinject

package main

import (
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRedis,

		//user
		dao.NewUserDaoGorm, cache.NewRedisUserCache,
		repository.NewUserRepoImpl,
		service.NewUserServiceImpl,

		// code
		ioc.InitSMSService, cache.NewCodeCacheImpl,
		repository.NewCodeRepoImpl,
		service.NewCodeServiceImpl,

		// web
		web.NewUserHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return &gin.Engine{}
}
