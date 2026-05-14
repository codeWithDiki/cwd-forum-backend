package main

import (
	"gin-quickstart/config"
	"gin-quickstart/internal/app"
	"gin-quickstart/pkg/logger"
	"gin-quickstart/pkg/worker"
	"gin-quickstart/routes"
	"os"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		panic("failed to load env: " + err.Error())
	}

	db, err := config.InitDB()
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	redis, err := config.InitRedis()
	if err != nil {
		panic("failed to connect to Redis: " + err.Error())
	}
	defer redis.Close()

	isProduction := os.Getenv("LOG_IS_PRODUCTION") == "true"
	logLevel := os.Getenv("LOG_LEVEL")
	isDisableStackTrace := os.Getenv("DISABLE_STACK_TRACE") == "true"
	log := logger.NewLogger(isProduction, logLevel, isDisableStackTrace)

	worker := worker.NewWorker(20)

	deps := app.Dependencies{
		DB:     db,
		Redis:  redis,
		Worker: worker,
		Logger: log,
	}

	r := routes.SetupRouter(deps)
	r.Run(":8080")
}
