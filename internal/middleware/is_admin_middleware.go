package middleware

import (
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func IsAdminLogged(userRepo repository.UserRepository, redis *redis.Client) gin.HandlerFunc {
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

		user, err := userRepo.GetUserByID(c, uint64(userID.(uint)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			c.Abort()
			return
		}

		if user.Role != enum.RoleAdmin.String() {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
