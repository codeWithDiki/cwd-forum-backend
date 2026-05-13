package middleware

import (
	"gin-quickstart/config"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func IsAdminLogged(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			c.Abort()
			return
		}
		db, err := config.InitDB()

		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			c.Abort()
			return
		}

		userRepo := repository.NewUserRepository(db, redis)

		user, err := userRepo.GetUserByID(uint64(userID.(uint)))
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			c.Abort()
			return
		}

		if user.Role != enum.RoleAdmin.String() {
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
