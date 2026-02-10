package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gladiator/ent"
)

type UserHandler struct {
	ent *ent.Client
}

func NewUserHandler(entClient *ent.Client) *UserHandler {
	return &UserHandler{ent: entClient}
}

func (h *UserHandler) Me(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	u, err := h.ent.User.Get(c.Request().Context(), uid)
	if err != nil {
		if ent.IsNotFound(err) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":         u.ID.String(),
		"email":      u.Email,
		"name":       u.Name,
		"avatar_url": u.AvatarURL,
		"created_at": u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	var req struct {
		Name      *string `json:"name"`
		AvatarURL *string `json:"avatar_url"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	upd := h.ent.User.UpdateOneID(uid)
	if req.Name != nil {
		upd = upd.SetName(*req.Name)
	}
	if req.AvatarURL != nil {
		upd = upd.SetAvatarURL(*req.AvatarURL)
	}
	u, err := upd.Save(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update failed"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":         u.ID.String(),
		"email":      u.Email,
		"name":       u.Name,
		"avatar_url": u.AvatarURL,
		"created_at": u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
