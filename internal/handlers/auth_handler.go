package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gladiator/internal/services"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req services.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation failed"})
	}
	u, err := h.auth.Register(c.Request().Context(), req)
	if err != nil {
		if err == services.ErrEmailExists {
			return c.JSON(http.StatusConflict, map[string]string{"error": "email already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "registration failed"})
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":         u.ID.String(),
			"email":      u.Email,
			"name":       u.Name,
			"avatar_url": u.AvatarURL,
			"created_at": u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req services.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation failed"})
	}
	resp, err := h.auth.Login(c.Request().Context(), req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "login failed"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil || req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "refresh_token required"})
	}
	resp, err := h.auth.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired refresh token"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "refresh failed"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	tokenID, _ := c.Get("token_id").(string)
	if userID != "" && tokenID != "" {
		_ = h.auth.Logout(c.Request().Context(), userID, tokenID)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) LogoutAll(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	if userID != "" {
		_ = h.auth.LogoutAll(c.Request().Context(), userID)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
