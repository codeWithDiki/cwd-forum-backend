package middleware

import (
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func S3Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s3Config := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""),
			Endpoint:         aws.String(os.Getenv("S3_ENDPOINT")),
			Region:           aws.String(os.Getenv("S3_REGION")),
			DisableSSL:       aws.Bool(false),
			S3ForcePathStyle: aws.Bool(true),
		}

		newSession, err := session.NewSession(s3Config)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to create AWS session: " + err.Error(),
			})
			c.Abort()
			return
		}

		s3Client := s3.New(newSession)

		c.Set("s3Client", s3Client)

	}
}
