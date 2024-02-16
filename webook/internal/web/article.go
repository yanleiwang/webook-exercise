package web

import (
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web/jwt"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc service.ArticleService
	log logger.Logger
}

func NewArticleHandler(svc service.ArticleService, log logger.Logger) *ArticleHandler {
	return &ArticleHandler{svc: svc, log: log}
}

func (h *ArticleHandler) RegisterHandlers(engine *gin.Engine) {
	g := engine.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	g.POST("/list", h.List)
	g.GET("/detail/:id", h.Detail)
	pub := g.Group("/pub")
	//pub.GET("/pub", a.PubList)
	pub.GET("/:id", h.PubDetail)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		h.log.Error("反序列化请求失败", logger.Error(err))
		return
	}

	uc, ok := ctx.MustGet(jwt.KeyAccessClaims).(*jwt.AccessClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("获得用户会话信息失败")
		return
	}

	article := h.toDomain(req, uc.Uid)

	id, err := h.svc.Save(ctx, article)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("保存数据失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "保存成功",
		Data: id,
	})
}

func (h *ArticleHandler) toDomain(req ArticleReq, uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		h.log.Error("反序列化请求失败", logger.Error(err))
		return
	}

	uc, ok := ctx.MustGet(jwt.KeyAccessClaims).(*jwt.AccessClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("获得用户会话信息失败")
		return
	}

	article := h.toDomain(req, uc.Uid)

	id, err := h.svc.Publish(ctx, article)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("发表帖子失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "发表帖子成功",
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		h.log.Error("反序列化请求失败", logger.Error(err))
		return
	}

	uc, ok := ctx.MustGet(jwt.KeyAccessClaims).(*jwt.AccessClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("获得用户会话信息失败")
		return
	}

	article := h.toDomain(req, uc.Uid)

	id, err := h.svc.Withdraw(ctx, article)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "成功",
		Data: id,
	})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	type Req struct {
		Limit  int `json:"limit,omitempty"`
		Offset int `json:"offset,omitempty"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		h.log.Error("反序列化请求失败", logger.Error(err))
		return
	}

	if req.Limit > 100 || req.Limit < 0 || req.Offset < 0 {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			// 我会倾向于不告诉前端批次太大
			// 因为一般你和前端一起完成任务的时候
			// 你们是协商好了的，所以会进来这个分支
			// 就表明是有人跟你过不去
			Msg: "请求有误",
		})
		h.log.Error("参数有误", logger.Field{Key: "req", Value: req})
		return
	}

	uc, ok := ctx.MustGet(jwt.KeyAccessClaims).(*jwt.AccessClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("获得用户会话信息失败")
		return
	}

	arts, err := h.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("失败", logger.Error(err))
		return
	}

	artVos := make([]ArticleVo, 0, len(arts))
	for _, art := range arts {
		artVos = append(artVos, ArticleVo{
			Id:       art.Id,
			Title:    art.Title,
			Abstract: art.Abstract(),
			Status:   art.Status.ToUint8(),
			// 这个列表请求，不需要返回内容
			//Content: src.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: src.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		})
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "成功",
		Data: artVos,
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.log.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}

	uc, ok := ctx.MustGet(jwt.KeyAccessClaims).(*jwt.AccessClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("获得用户会话信息失败")
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("获得文章信息失败", logger.Error(err))
		return
	}
	// 这是不借助数据库查询来判定的方法
	if art.Author.Id != uc.Uid {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		})
		// 如果公司有风控系统，这个时候就要上报这种非法访问的用户了。
		h.log.Error("非法访问文章，创作者 ID 不匹配",
			logger.Int64("uid", uc.Uid))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.log.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}

	art, err := h.svc.GetPubById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("获得文章信息失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			Author: art.Author.Name,
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
		},
	})

}
