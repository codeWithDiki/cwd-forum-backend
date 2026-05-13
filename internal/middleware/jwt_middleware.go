package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	jwtLib "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwtLib.Parse(tokenString, func(token *jwtLib.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			c.Abort()
			return
		}

		exists, err := redis.Exists(c, tokenString).Result()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "internal server error",
			})
			c.Abort()
			return
		}

		if exists > 0 {
			var userID int

			user_id, err := redis.Get(c, tokenString).Result()

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "internal server error",
				})
				c.Abort()
				return
			}

			userID, err = strconv.Atoi(user_id)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "internal server error",
				})
				c.Abort()
				return
			}

			c.Set("user_id", uint(userID))
			c.Set("token", tokenString)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "token expired",
		})
		c.Abort()
	}
}
