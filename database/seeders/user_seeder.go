package seeders

import (
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"log"

	"gorm.io/gorm"
)

func SeedAdminUser(db *gorm.DB) {
	adminUser := db.Where("role = ?", enum.RoleAdmin.String()).First(&model.User{})

	if adminUser.RowsAffected == 0 {

		createError := db.Create(&model.User{
			Name:     "admin",
			Email:    "me@dikiakbarasyidiq.dev",
			Username: "admin",
			Password: "password",
			Role:     enum.RoleAdmin.String(),
		}).Error

		if createError != nil {
			panic("Failed to create admin user: " + createError.Error())
		}

	}

	log.Print("Admid user seeded.")

}
