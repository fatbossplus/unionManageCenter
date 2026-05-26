package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"unionManageCenter/pkg/auth"
	"unionManageCenter/pkg/response"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roleCode", claims.RoleCode)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	v, _ := c.Get("userID")
	id, _ := v.(uint64)
	return id
}

func GetUsername(c *gin.Context) string {
	v, _ := c.Get("username")
	s, _ := v.(string)
	return s
}

func GetRoleCode(c *gin.Context) string {
	v, _ := c.Get("roleCode")
	s, _ := v.(string)
	return s
}
