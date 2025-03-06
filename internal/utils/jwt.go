package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/config"
)

func GenerateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     JWTExpiration(),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := config.GetEnv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}

func ParseJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetEnv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func JWTExpiration() time.Duration {
	expirationStr := config.GetEnv("JWT_EXPIRATION")
	if expirationStr == "" {
		// Default to 24 hours if not set
		return 24 * time.Hour
	}

	expiration, err := time.ParseDuration(expirationStr)
	if err != nil {
		// Fallback to 24 hours if parsing fails
		return 24 * time.Hour
	}

	return expiration
}

func GenerateRefreshToken() (string, error) {
	token := make([]byte, 32) // 32 bytes = 256 bits

	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshToken := base64.URLEncoding.EncodeToString(token)
	return refreshToken, nil
}
