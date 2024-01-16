package dao

import "gorm.io/gorm"

func InitDB(db *gorm.DB) {
	err := db.AutoMigrate(&User{})
	if err != nil {
		panic(err)
	}
}
