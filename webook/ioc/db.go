package ioc

import (
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"time"
)

func InitDB(log logger.Logger) *gorm.DB {
	/*
		&gorm.Config{
				TranslateError: true,
			}) 的目的是
		为了把 mysql 唯一索引冲突 错误转换为 gorm.ErrDuplicatedKey

		也可以 用 MySQL GO 驱动的 error 定义，找到准确的错误

			err := ud.db.WithContext(ctx).Create(&u).Error
			if me, ok := err.(*mysql.MySQLError); ok {
				const uniqueIndexErrNo uint16 = 1062
				if me.Number == uniqueIndexErrNo {
					return ErrUserDuplicate
				}
			}
	*/

	type Config struct {
		DSN string `yaml:"dsn"`
	}

	var c Config
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{
		TranslateError: true,
		Logger: glogger.New(gormLoggerFunc(log.Debug), glogger.Config{
			// 慢查询阈值，只有执行时间超过这个阈值，才会使用
			// 50ms， 100ms
			// SQL 查询必然要求命中索引，最好就是走一次磁盘 IO
			// 一次磁盘 IO 是不到 10ms
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  glogger.Info,
		}),
	})

	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
