package middleware

import (
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func IsCanUpdateThread(db *gorm.DB, s *service.ThreadService) gin.HandlerFunc {
	return func(c *gin.Context) {
		param := c.Param("id")

		threadID, err := strconv.ParseUint(param, 10, 64)

		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "invalid thread ID",
			})
			c.Abort()
			return
		}

		thread, err := s.GetThreadByID(threadID, c)

		if err != nil {
			c.JSON(404, gin.H{
				"success": false,
				"error":   "Thread not found",
			})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			c.Abort()
			return
		}

		userModel := &model.User{}

		db.Where("id = ?", userID).Model(&userModel).First(&userModel)

		if userModel.Role == enum.RoleAdmin.String() {
			c.Next()
			return
		}

		if thread.AuthorID != userID.(uint) {
			c.JSON(403, gin.H{
				"success": false,
				"error":   "Forbidden",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
