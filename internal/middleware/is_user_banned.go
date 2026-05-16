package middleware

import (
	"gin-quickstart/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func IsUserBanned(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			c.Abort()
			return
		}

		userModel := &model.User{}

		db.Where("id = ?", userID).Model(&userModel).First(&userModel)

		if userModel.IsBanned {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Your account has been banned",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
