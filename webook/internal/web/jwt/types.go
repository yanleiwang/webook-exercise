package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetAccessToken(ctx *gin.Context, uid int64, ssid string) error
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) error
	ExtractAccessClaims(ctx *gin.Context) (AccessClaims, error)
	ExtractRefreshClaims(ctx *gin.Context) (RefreshClaims, error)
}

type AccessClaims struct {
	// 我们只需要放一个 user id 就可以了
	Uid       int64
	Ssid      string
	UserAgent string
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Uid       int64
	Ssid      string
	UserAgent string
	jwt.RegisteredClaims
}
