package middleware

import (
	"net/http"
	"os"
	"strings"

	jwtLib "github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {
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

		claims, ok := token.Claims.(jwtLib.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid claims",
			})
			c.Abort()
			return
		}

		// simpan user_id ke gin context
		c.Set("user_id", uint(claims["user_id"].(float64)))

		c.Next()
	}
}
