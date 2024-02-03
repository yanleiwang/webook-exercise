package web

import (
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web/jwt"
	"gitee.com/geekbang/basic-go/webook/internal/web/middlewares"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const biz = "login"

var _ handler = (*UserHandler)(nil)

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	jwt.Handler
	log logger.Logger
}

func NewUserHandler(svc service.UserService, codeService service.CodeService, jwtHdl jwt.Handler, log logger.Logger) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)

	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)

	return &UserHandler{
		svc:         svc,
		codeSvc:     codeService,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		Handler:     jwtHdl,
		log:         log,
	}
}

func (u *UserHandler) RegisterHandlers(engine *gin.Engine) {
	ug := engine.Group("/users")

	ug.POST("/signup", u.SignUp)
	//ug.POST("/login", u.Login)
	//ug.GET("/profile", u.Profile)

	ug.POST("/login", u.LoginJWT)
	ug.GET("/profile", u.ProfileJWT)
	//ug.POST("/edit", u.Edit)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)
	ug.POST("/logout", u.LogoutJWT)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type Req struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		ctx.String(http.StatusOK, "邮箱格式不对")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		u.log.Error("密码 正则匹配 失败", logger.Error(err))
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case service.ErrUserDuplicate:
		u.log.Info("重复注册")
		ctx.String(http.StatusOK, "手机/邮箱已注册")
		return
	case nil:
		ctx.String(http.StatusOK, "注册成功")
		return
	default:
		u.log.Error("用户注册失败", logger.Error(err))
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (u *UserHandler) Edit(context *gin.Context) {
	panic("implement me")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	user, err := u.svc.Login(ctx, domain.User{Email: req.Email, Password: req.Password})

	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "邮箱或者密码错误")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	err = middlewares.SetUserId(ctx, user.Id, 60)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功")

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	type Resp struct {
		Email string `json:"email"`
	}

	id, ok := middlewares.GetUserId(ctx).(int64)

	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	user, err := u.svc.Profile(ctx, id)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Resp{Email: user.Email})

}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	user, err := u.svc.Login(ctx, domain.User{Email: req.Email, Password: req.Password})

	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "邮箱或者密码错误")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Resp struct {
		Email string `json:"email"`
	}

	claims, ok := ctx.MustGet(jwt.KeyUserClaims).(*jwt.AccessClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	user, err := u.svc.Profile(ctx, claims.Uid)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Resp{Email: user.Email})

}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
			Data: nil,
		})
	}

	// 应该用 正则表达式  判断是不是合法的手机
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}

	err := u.codeSvc.Send(ctx, biz, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}

}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 这边，可以加上各种校验
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码有误",
		})
		return
	}

	// 我这个手机号，会不会是一个新用户呢？
	// 这样子
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 这边要怎么办呢？
	// 从哪来？
	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg: "验证码校验通过",
	})
}

func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	claims, err := u.ExtractRefreshClaims(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	expirationTime, err := claims.GetExpirationTime()
	if err != nil || expirationTime.Before(time.Now()) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = u.CheckSession(ctx, claims.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 搞个新的 access_token
	err = u.SetAccessToken(ctx, claims.Uid, claims.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录OK",
	})
}
