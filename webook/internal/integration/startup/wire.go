//go:build wireinject

package startup

import (
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao/article"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/internal/web/jwt"
	"gitee.com/geekbang/basic-go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(InitDB, InitRedis, ioc.InitLogger, jwt.NewJWTHandler)
var userSvcProvider = wire.NewSet(
	dao.NewUserDaoGorm,
	cache.NewRedisUserCache,
	repository.NewUserRepoImpl,
	service.NewUserServiceImpl)

var articleSvcProvider = wire.NewSet(
	service.NewArticleService,
	repository.NewArticleRepository,
	article.NewArticleDaoGORM,
	cache.NewRedisArticleCache,
)

var codeSvcProvider = wire.NewSet(
	ioc.InitSMSService, cache.NewCodeCacheImpl,
	repository.NewCodeRepoImpl,
	service.NewCodeServiceImpl,
)

var weChatProvider = wire.NewSet(
	ioc.InitWechatService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		web.NewUserHandler,
		// code
		codeSvcProvider,

		// wechat
		weChatProvider,
		web.NewOAuth2WechatHandler,

		// article
		articleSvcProvider,
		web.NewArticleHandler,

		// web
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return &gin.Engine{}
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		articleSvcProvider,
		userSvcProvider,
		web.NewArticleHandler,
	)

	return new(web.ArticleHandler)
}
