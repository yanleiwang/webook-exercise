package dao

import (
	"database/sql"
	"errors"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("邮箱冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDao interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
}

type userDaoGorm struct {
	db *gorm.DB
}

func (u *userDaoGorm) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return user, err
}

func (u *userDaoGorm) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (u *userDaoGorm) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (u *userDaoGorm) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := u.db.WithContext(ctx).Create(&user).Error
	switch err {
	case gorm.ErrDuplicatedKey:

		//if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		//	const uniqueConflictsErrNo uint16 = 1062
		//	if mysqlErr.Number == uniqueConflictsErrNo {
		//		// 邮箱冲突 or 手机号码冲突
		//		return ErrUserDuplicate
		//	}
		//}
		return ErrUserDuplicate
	default:
		return err
	}

}

func NewUserDaoGorm(db *gorm.DB) UserDao {
	return &userDaoGorm{db: db}
}

type User struct {
	Id       int64          `gorm:"primaryKey;autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string
	Ctime    int64
	Utime    int64
}
