package middleware

import (
	"gin-quickstart/auth"

	"github.com/gin-gonic/gin"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err == nil {
			claims, err := auth.ValidateAccessToken(accessToken)
			if err == nil {
				c.Set("userID", claims.UserID)
				c.Set("role", claims.Role)
				c.Next()
				return
			}
		}

		uid, role, err := auth.PerformTokenRefresh(c)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Not logged in"})
			return
		}

		c.Set("userID", uid)
		c.Set("role", role)
		c.Next()
	}

}
