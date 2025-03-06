package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/config"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/utils"
)

func AuthMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendResponse(c, http.StatusUnauthorized, "Authorization header is required", nil, true)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.SendResponse(c, http.StatusUnauthorized, "Invalid token format", nil, true)
			c.Abort()
			return
		}

		// Check if the token is blacklisted
		_, err := redisClient.Get("blacklist:" + tokenString).Result()
		if err == nil {
			utils.SendResponse(c, http.StatusUnauthorized, "Token is blacklisted", nil, true)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetEnv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			utils.SendResponse(c, http.StatusUnauthorized, "Invalid token", nil, true)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.SendResponse(c, http.StatusUnauthorized, "Invalid token claims", nil, true)
			c.Abort()
			return
		}
		userID := uint(claims["user_id"].(float64))

		// Set the user ID in the Gin context
		c.Set("userID", userID)
		c.Next()
	}
}
