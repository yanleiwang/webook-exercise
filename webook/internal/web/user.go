package web

import (
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web/middlewares"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type UserHandler struct {
	svc         service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)

	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)

	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
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
		// TODO 日志
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
	case service.ErrUserDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱已注册")
		return
	case nil:
		ctx.String(http.StatusOK, "注册成功")
		return
	default:
		//TODO 日志
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, middlewares.UserClaims{
		Id:        user.Id,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			// 演示目的设置为一分钟过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	})
	tokenStr, err := token.SignedString(middlewares.JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Resp struct {
		Email string `json:"email"`
	}

	value, exist := ctx.Get(middlewares.KeyUserClaims)
	token, ok := value.(middlewares.UserClaims)
	if !exist || !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	user, err := u.svc.Profile(ctx, token.Id)
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Resp{Email: user.Email})

}
