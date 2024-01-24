package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepo interface {
	Set(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, code string) (bool, error)
}

type CodeRepoImpl struct {
	cache cache.CodeCache
}

func (c *CodeRepoImpl) Verify(ctx context.Context, biz string, phone string, code string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, code)
}

func NewCodeRepoImpl(cache cache.CodeCache) CodeRepo {
	return &CodeRepoImpl{cache: cache}
}

func (c *CodeRepoImpl) Set(ctx context.Context, biz string, phone string, code string) error {
	return c.cache.Set(ctx, biz, phone, code)
}
