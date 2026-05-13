package middleware

import (
	"gin-quickstart/pkg/logger"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("logger", logger)

		c.Next()
	}
}
