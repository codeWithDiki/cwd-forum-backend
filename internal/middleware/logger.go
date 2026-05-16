package middleware

import (
	"gin-quickstart/pkg/logger"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := logger.SetTraceID(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
