package repository

import (
	"database/sql"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"golang.org/x/net/context"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepo interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
}

type userRepoImpl struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func (u *userRepoImpl) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := u.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return u.daoToDomain(user), err
}

func (u *userRepoImpl) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := u.cache.Get(ctx, id)
	if err == nil {
		return user, err
	}
	ue, err := u.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	user = u.daoToDomain(ue)
	go func() {
		er := u.cache.Set(ctx, user)
		if er != nil {
			// 打日志 做监控
		}
	}()

	return user, nil
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
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
	}
}

func (u *userRepoImpl) daoToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Phone:    user.Phone.String,
		Password: user.Password,
	}
}

func NewUserRepoImpl(dao dao.UserDao, cache cache.UserCache) UserRepo {
	return &userRepoImpl{
		dao:   dao,
		cache: cache,
	}
}
