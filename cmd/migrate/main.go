package main

import (
	"gin-quickstart/config"
	"gin-quickstart/database/seeders"
	"gin-quickstart/internal/model"
	"log"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatal("failed to load env: ", err)
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	log.Println("Running AutoMigrate...")
	err = db.AutoMigrate(
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
		&model.UserUser{},
	)

	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	log.Println("Database migration completed.")

	log.Println("Running seeders...")
	seeders.Run(db)
	log.Println("Seeding completed.")
}
