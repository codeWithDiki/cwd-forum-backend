package main

import (
	"gin-quickstart/config"
	"gin-quickstart/database/seeders"
	"gin-quickstart/internal/model"
	"gin-quickstart/routes"
)

func main() {
	db, err := config.InitDB()

	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Thread{},
		&model.Post{},
		&model.Vote{},
		&model.Reaction{},
		&model.Tag{},
		&model.Notification{},
		&model.Badge{},
		&model.ModerationLog{},
		&model.Attachment{},
	)

	seeders.Run(db)

	r := routes.SetupRouter()

	r.Run(":8080")
}
