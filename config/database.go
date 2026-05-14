package config

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	driver := os.Getenv("DB_DRIVER")
	source := os.Getenv("DB_SOURCE")

	var db *gorm.DB
	var err error

	switch driver {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(source), &gorm.Config{})
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}
