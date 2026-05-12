package seeders

import "gorm.io/gorm"

func Run(db *gorm.DB) {
	SeedAdminUser(db)
}
