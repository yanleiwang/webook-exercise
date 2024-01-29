package service

import (
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	mock_repository "gitee.com/geekbang/basic-go/webook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/net/context"
	"testing"
)

func Test_userServiceImpl_Login(t *testing.T) {

	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepo

		user domain.User

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功", // 用户名和密码是对的
			mock: func(ctrl *gomock.Controller) repository.UserRepo {
				repo := mock_repository.NewMockUserRepo(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$MN9ZKKIbjLZDyEpCYW19auY7mvOG9pcpiIcUUoZZI6pA6OmKZKOVi",
						Phone:    "15212345678",
					}, nil)
				return repo
			},
			user: domain.User{
				Email:    "123@qq.com",
				Password: "hello#world123",
			},
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$MN9ZKKIbjLZDyEpCYW19auY7mvOG9pcpiIcUUoZZI6pA6OmKZKOVi",
				Phone:    "15212345678",
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepo {
				repo := mock_repository.NewMockUserRepo(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			user: domain.User{
				Email:    "123@qq.com",
				Password: "hello#world123",
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userRepo := tc.mock(ctrl)
			userSvc := NewUserServiceImpl(userRepo)
			user, err := userSvc.Login(context.Background(), tc.user)
			assert.Equal(t, tc.wantErr, err)
			if err == nil {
				return
			}
			assert.Equal(t, tc.wantUser, user)

		})
	}

}
