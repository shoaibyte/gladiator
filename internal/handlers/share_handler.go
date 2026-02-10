package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gladiator/ent/notebookshare"
	"gladiator/internal/services"
)

type ShareHandler struct {
	share *services.ShareService
	nb    *services.NotebookService
}

func NewShareHandler(share *services.ShareService, nb *services.NotebookService) *ShareHandler {
	return &ShareHandler{share: share, nb: nb}
}

func (h *ShareHandler) Share(c echo.Context) error {
	notebookID := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, _ := uuid.Parse(userID)
	var req struct {
		Email      string `json:"email"`
		Permission string `json:"permission"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	perm := notebookshare.PermissionView
	switch req.Permission {
	case "edit":
		perm = notebookshare.PermissionEdit
	case "admin":
		perm = notebookshare.PermissionAdmin
	}
	if err := h.share.Share(c.Request().Context(), notebookID, uid, req.Email, perm); err != nil {
		if err.Error() == "not found" || err.Error() == "user not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err.Error() == "access denied" || err.Error() == "cannot share with yourself" || err.Error() == "already shared" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ShareHandler) ListShares(c echo.Context) error {
	notebookID := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)
	list, err := h.share.ListShares(c.Request().Context(), notebookID, uid)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	}
	type item struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Permission string `json:"permission"`
	}
	out := make([]item, 0, len(list))
	for _, s := range list {
		email := ""
		if s.Edges.SharedWith != nil {
			email = s.Edges.SharedWith.Email
		}
		out = append(out, item{ID: s.ID.String(), Email: email, Permission: string(s.Permission)})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"shares": out})
}

func (h *ShareHandler) UpdateShare(c echo.Context) error {
	notebookID := c.Param("id")
	shareID := c.Param("sid")
	userID, _ := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)
	var req struct {
		Permission string `json:"permission"`
	}
	c.Bind(&req)
	perm := notebookshare.PermissionView
	if req.Permission == "edit" {
		perm = notebookshare.PermissionEdit
	} else if req.Permission == "admin" {
		perm = notebookshare.PermissionAdmin
	}
	if err := h.share.UpdatePermission(c.Request().Context(), shareID, notebookID, uid, perm); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ShareHandler) RevokeShare(c echo.Context) error {
	notebookID := c.Param("id")
	shareID := c.Param("sid")
	userID, _ := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)
	if err := h.share.Revoke(c.Request().Context(), shareID, notebookID, uid); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ShareHandler) ListPublic(c echo.Context) error {
	page, limit := 1, 20
	if p := c.QueryParam("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if v, err := parseInt(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	search := c.QueryParam("search")
	result, err := h.nb.ListPublic(c.Request().Context(), page, limit, search)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *ShareHandler) Fork(c echo.Context) error {
	notebookID := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)
	forked, err := h.nb.Fork(c.Request().Context(), notebookID, uid)
	if err != nil {
		if err.Error() == "not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, notebookToJSON(forked))
}
