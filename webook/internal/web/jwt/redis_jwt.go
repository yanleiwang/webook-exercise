package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

// JWTKey 因为 JWT Key 不太可能变，所以可以直接写成常量
// 也可以考虑做成依赖注入
var (
	JWTKey = []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm")
)

type JWTHandler struct {
	cmd redis.Cmdable
}

func NewJWTHandler(cmd redis.Cmdable) Handler {
	return &JWTHandler{
		cmd: cmd,
	}
}

func (j *JWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	claims := ctx.MustGet(KeyUserClaims).(*AccessClaims)
	return j.cmd.Set(ctx, j.getRedisKey(claims.Ssid), "", time.Hour*24*7).Err()
}

func (j *JWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	val, err := j.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil:
		return nil
	case nil:
		if val == 0 {
			return nil
		}
		return errors.New("session 已经无效了")
	default:
		return err
	}
}

func (j *JWTHandler) getRedisKey(ssid string) string {
	return fmt.Sprintf("users:ssid:%s", ssid)
}

func (j *JWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := j.SetAccessToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = j.SetRefreshToken(ctx, uid, ssid)
	return err
}

func (j *JWTHandler) SetAccessToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := AccessClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			// 演示目的设置为一分钟过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (j *JWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (j *JWTHandler) ExtractAccessClaims(ctx *gin.Context) (AccessClaims, error) {
	uc := AccessClaims{}
	tokenStr := j.extractTokenStr(ctx)
	if tokenStr == "" {
		return uc, errors.New("AccessToken 不存在")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil {
		return uc, err
	}
	if !token.Valid {
		return uc, errors.New("AccessToken 无效")
	}

	return uc, nil
}

func (j *JWTHandler) ExtractRefreshClaims(ctx *gin.Context) (RefreshClaims, error) {
	uc := RefreshClaims{}
	tokenStr := j.extractTokenStr(ctx)
	if tokenStr == "" {
		return uc, errors.New("AccessToken 不存在")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil {
		return uc, err
	}
	if !token.Valid {
		return uc, errors.New("RefreshToken 无效")
	}

	return uc, nil
}

func (j *JWTHandler) extractTokenStr(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	authSegments := strings.Split(tokenHeader, " ")
	if len(authSegments) != 2 {
		return ""
	}
	return authSegments[1]
}

const (
	KeyUserClaims = "userClaims"
)
