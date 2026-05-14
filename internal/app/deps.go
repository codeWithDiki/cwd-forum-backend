package app

import (
	"gin-quickstart/pkg/logger"
	"gin-quickstart/pkg/worker"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Worker *worker.WorkerPool
	Logger *logger.Logger
}
