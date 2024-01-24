package dao

import "gorm.io/gorm"

func InitTables(db *gorm.DB) error {
	err := db.AutoMigrate(&User{})
	return err
}
