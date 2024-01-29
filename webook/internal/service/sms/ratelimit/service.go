package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms"
	"gitee.com/geekbang/basic-go/webook/pkg/utils/ratelimit"
)

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

var ErrLimited = errors.New("短信服务触发了限流")

func (r *RatelimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limit, err := r.limiter.Limit(ctx, "sms:tencet")
	if err != nil {
		return fmt.Errorf("短信限流服务 判断是否出现限流出现问题, %w", err)
	}

	if !limit {
		return ErrLimited
	}

	err = r.svc.Send(ctx, tpl, args, numbers...)
	return err
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) *RatelimitSMSService {
	return &RatelimitSMSService{svc: svc, limiter: limiter}
}
