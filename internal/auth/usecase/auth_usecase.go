package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/dto"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/entity"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/repository"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/utils"
	"golang.org/x/oauth2"
)

type AuthUseCase struct {
	userRepo    repository.UserRepository
	redisClient *redis.Client
}

func NewAuthUseCase(userRepo repository.UserRepository, redisClient *redis.Client) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo, redisClient: redisClient}
}

func (uc *AuthUseCase) FindOrCreateUserByProvider(provider, email, providerID, name string) (*entity.User, error) {
	user, err := uc.userRepo.FindByProvider(provider, providerID)
	if err == nil {
		return user, nil
	}

	user = &entity.User{
		Email:      email,
		Name:       name,
		Provider:   provider,
		ProviderID: providerID,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *AuthUseCase) Register(req dto.RegisterRequest) (string, string, error) {
	existingUser, err := uc.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return "", "", errors.New("user already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}
	if err := uc.userRepo.Create(user); err != nil {
		return "", "", fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// cache to redis
	key := fmt.Sprintf("user:%d:refresh_token", user.ID)
	if err := uc.redisClient.Set(key, refreshToken, utils.JWTExpiration()).Err(); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (uc *AuthUseCase) Login(email, password string) (string, string, error) {
	user, err := uc.userRepo.FindByEmail(email)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store the refresh token in Redis
	key := fmt.Sprintf("user:%d:refresh_token", user.ID)
	if err := uc.redisClient.Set(key, refreshToken, utils.JWTExpiration()).Err(); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (uc *AuthUseCase) Logout(userID uint, accessToken string) error {
	// Blacklist the access token
	expiration := time.Until(time.Now().Add(utils.JWTExpiration()))
	if err := uc.redisClient.Set("blacklist:"+accessToken, userID, expiration).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Remove the refresh token
	if err := uc.redisClient.Del(fmt.Sprintf("user:%d:refresh_token", userID)).Err(); err != nil {
		return fmt.Errorf("failed to remove refresh token: %w", err)
	}

	return nil
}

func (uc *AuthUseCase) RefreshToken(userID uint, refreshToken string) (string, error) {
	storedRefreshToken, err := uc.redisClient.Get(fmt.Sprintf("user:%d:refresh_token", userID)).Result()
	if err != nil {
		return "", fmt.Errorf("invalid or expired refresh token: %w", err)
	}
	if storedRefreshToken != refreshToken {
		return "", errors.New("invalid refresh token")
	}

	newAccessToken, err := utils.GenerateJWT(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return newAccessToken, nil
}

func (uc *AuthUseCase) HandleGoogleCallback(code string, googleOauthConfig *oauth2.Config) (string, string, *dto.GoogleCallbackResponse, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	userInfo, err := fetchUserInfo(token.AccessToken)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	user, err := uc.userRepo.FindOrCreateUserByProvider("google", userInfo.Email, userInfo.GoogleID, userInfo.Name)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create/update user: %w", err)
	}

	accessToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store the refresh token in Redis
	key := fmt.Sprintf("user:%d:refresh_token", user.ID)
	if err := uc.redisClient.Set(key, refreshToken, utils.JWTExpiration()).Err(); err != nil {
		return "", "", nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	userData := &dto.GoogleCallbackResponse{
		Email: user.Email,
		Name:  user.Name,
	}

	return accessToken, refreshToken, userData, nil
}

func fetchUserInfo(accessToken string) (*entity.GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status code %d", resp.StatusCode)
	}

	var userInfo entity.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}
