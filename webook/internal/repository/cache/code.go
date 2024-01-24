package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/set_code.lua
	luaSendCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string
)

var (
	ErrSystemError            = errors.New("系统错误")
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnknownForCode         = errors.New("我也不知发生什么了，反正是跟 code 有关")
)

type CodeCache interface {
	Set(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, code string) (bool, error)
}

type CodeCacheImpl struct {
	cmd redis.Cmdable
}

func (c *CodeCacheImpl) Verify(ctx context.Context, biz string, phone string, code string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{c.genKey(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		// 正常来说，如果频繁出现这个错误，你就要告警，因为有人搞你
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
		//default:
		//	return false, ErrUnknownForCode
	}
	return false, ErrUnknownForCode
}

func NewCodeCacheImpl(cmd redis.Cmdable) CodeCache {
	return &CodeCacheImpl{cmd: cmd}
}

func (c *CodeCacheImpl) Set(ctx context.Context, biz string, phone string, code string) error {
	ret, err := c.cmd.Eval(ctx, luaSendCode, []string{c.genKey(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch ret {
	case -1:
		return ErrCodeSendTooMany
	case 0:
		return nil
	default:
		return ErrSystemError
	}
}

func (c *CodeCacheImpl) genKey(biz string, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
