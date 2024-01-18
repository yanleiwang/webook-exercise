package repository

import (
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"golang.org/x/net/context"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepo interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
}

type userRepoImpl struct {
	dao dao.UserDao
}

func (u *userRepoImpl) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := u.dao.FindById(ctx, id)
	return u.daoToDomain(user), err
}

func (u *userRepoImpl) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return u.daoToDomain(user), nil
}

func (u *userRepoImpl) Create(ctx context.Context, user domain.User) error {
	return u.dao.Insert(ctx, u.domainToDao(user))
}

func (u *userRepoImpl) domainToDao(user domain.User) dao.User {
	return dao.User{
		Email:    user.Email,
		Password: user.Password,
	}
}

func (u *userRepoImpl) daoToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
	}
}

func NewUserRepoImpl(dao dao.UserDao) UserRepo {
	return &userRepoImpl{dao: dao}
}
