package dao

import (
	"errors"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDao interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
}

type UserDaoGorm struct {
	db *gorm.DB
}

func (u *UserDaoGorm) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (u *UserDaoGorm) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (u *UserDaoGorm) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := u.db.WithContext(ctx).Create(&user).Error
	switch err {
	case gorm.ErrDuplicatedKey:
		return ErrUserDuplicateEmail
	default:
		return err
	}

}

func NewUserDaoGorm(db *gorm.DB) UserDao {
	return &UserDaoGorm{db: db}
}

type User struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	Email    string `gorm:"uniqueIndex;type:varchar(255)"`
	Password string
	Ctime    int64
	Utime    int64
}
