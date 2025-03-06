package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/dto"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/auth/usecase"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/config"
	"github.com/satya-nurhutama/go-oauth-boilerplate/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

var (
	googleOauthConfig *oauth2.Config
	once              sync.Once
	randomState       = "random"
)

func initGoogleOAuthConfig() {
	once.Do(func() {
		googleOauthConfig = &oauth2.Config{
			ClientID:     config.GetEnv("GOOGLE_CLIENT_ID"),
			ClientSecret: config.GetEnv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  config.GetEnv("GOOGLE_REDIRECT_URL"),
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		}
	})
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	initGoogleOAuthConfig()
	url := googleOauthConfig.AuthCodeURL(randomState)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	initGoogleOAuthConfig()

	state := c.Query("state")
	if state != randomState {
		utils.SendResponse(c, http.StatusBadRequest, "Invalid state", nil, true)
		return
	}

	code := c.Query("code")

	accessToken, refreshToken, user, err := h.authUseCase.HandleGoogleCallback(code, googleOauthConfig)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, err.Error(), nil, true)
		return
	}

	utils.SendResponse(c, http.StatusOK, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	}, false)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, err.Error(), nil, true)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Login(req.Email, req.Password)
	if err != nil {
		utils.SendResponse(c, http.StatusUnauthorized, err.Error(), nil, true)
		return
	}

	utils.SendResponse(c, http.StatusOK, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, false)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, err.Error(), nil, true)
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Register(req)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, err.Error(), nil, true)
		return
	}

	utils.SendResponse(c, http.StatusOK, "Registration successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, false)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil, true)
		return
	}

	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		utils.SendResponse(c, http.StatusBadRequest, "Authorization header is required", nil, true)
		return
	}

	if err := h.authUseCase.Logout(userID.(uint), accessToken); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, err.Error(), nil, true)
		return
	}

	utils.SendResponse(c, http.StatusOK, "Logout successful", nil, false)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil, true)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, err.Error(), nil, true)
		return
	}

	newAccessToken, err := h.authUseCase.RefreshToken(userID.(uint), req.RefreshToken)
	if err != nil {
		utils.SendResponse(c, http.StatusUnauthorized, err.Error(), nil, true)
		return
	}

	utils.SendResponse(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"token": newAccessToken,
	}, false)
}
