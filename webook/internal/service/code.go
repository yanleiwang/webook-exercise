package service

import (
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms"
	"golang.org/x/net/context"
	"math/rand"
)

const codeTplId = "1877556"

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type CodeServiceImpl struct {
	smsSvc sms.Service
	repo   repository.CodeRepo
}

func NewCodeServiceImpl(smsSvc sms.Service, repo repository.CodeRepo) CodeService {
	return &CodeServiceImpl{smsSvc: smsSvc, repo: repo}
}

func (c *CodeServiceImpl) Send(ctx context.Context, biz string, phone string) error {

	// 生成验证码
	code := c.generateCode()
	err := c.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	err = c.smsSvc.Send(ctx, codeTplId, []string{code}, phone)

	//if err != nil {
	// 这个地方怎么办？
	// 这意味着，Redis 有这个验证码，但是不好意思，
	// 我能不能删掉这个验证码？
	// 你这个 err 可能是超时的 err，你都不知道，发出了没
	// 在这里重试
	// 要重试的话，初始化的时候，传入一个自己就会重试的 smsSvc
	//}
	return err

}

func (c *CodeServiceImpl) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return c.repo.Verify(ctx, biz, phone, inputCode)
}

func (c *CodeServiceImpl) generateCode() string {
	// 6位随机数， 不够0的 加上前导0
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
