package service

import (
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
)

type UserService interface {
	SignUp(ctx context.Context, user domain.User) error
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
}

type userServiceImpl struct {
	repo repository.UserRepo
}

func (u *userServiceImpl) Profile(ctx context.Context, id int64) (domain.User, error) {
	return u.repo.FindById(ctx, id)
}

func (u *userServiceImpl) Login(ctx context.Context, user domain.User) (domain.User, error) {
	found, err := u.repo.FindByEmail(ctx, user.Email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	// 其他错误
	if err != nil {
		return domain.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(found.Password), []byte(user.Password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return found, nil
}

func (u *userServiceImpl) SignUp(ctx context.Context, user domain.User) error {
	// 加密
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)
	return u.repo.Create(ctx, user)
}

func NewUserServiceImpl(repo repository.UserRepo) UserService {
	return &userServiceImpl{
		repo: repo,
	}
}
