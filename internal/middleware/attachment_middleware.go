package middleware

import (
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
)

func AttachmentMiddleware(wp *workerpool.WorkerPool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("fileUploadWorkerPool", wp)
		c.Next()
	}
}
