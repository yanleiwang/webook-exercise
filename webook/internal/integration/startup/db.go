package startup

import (
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() *gorm.DB {
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

	if db != nil {
		return db
	}

	type Config struct {
		DSN string `yaml:"dsn"`
	}

	var c Config
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(err)
	}

	//var c Config = Config{DSN: "root:root@tcp(localhost:13306)/webook"}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: c.DSN,
		//DefaultStringSize: 256,
	}), &gorm.Config{
		TranslateError: true,
	})

	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	db = db.Debug()
	return db
}
