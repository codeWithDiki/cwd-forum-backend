package middleware

import (
	"gin-quickstart/pkg/worker"

	"github.com/gin-gonic/gin"
)

func WorkerPoolMiddleware(wp *worker.WorkerPool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("workerPool", wp)
		c.Next()
	}
}
