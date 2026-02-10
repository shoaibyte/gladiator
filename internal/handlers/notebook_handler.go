package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gladiator/ent"
	"gladiator/internal/services"
)

type NotebookHandler struct {
	svc *services.NotebookService
}

func NewNotebookHandler(svc *services.NotebookService) *NotebookHandler {
	return &NotebookHandler{svc: svc}
}

func (h *NotebookHandler) Create(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user"})
	}
	var req services.CreateNotebookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	nb, err := h.svc.Create(c.Request().Context(), uid, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, notebookToJSON(nb))
}

func (h *NotebookHandler) List(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user"})
	}
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
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "updated_at"
	}
	order := c.QueryParam("order")
	if order != "asc" {
		order = "desc"
	}
	search := c.QueryParam("search")
	result, err := h.svc.List(c.Request().Context(), uid, page, limit, sort, order, search)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *NotebookHandler) Get(c echo.Context) error {
	id := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user"})
	}
	nb, err := h.svc.Get(c.Request().Context(), id, uid)
	if err != nil {
		if err.Error() == "not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		if err.Error() == "access denied" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, notebookToJSONWithOwner(nb))
}

func (h *NotebookHandler) Update(c echo.Context) error {
	id := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user"})
	}
	var req struct {
		Title       *string                `json:"title"`
		Description *string                `json:"description"`
		Content     map[string]interface{} `json:"content"`
		IsPublic    *bool                  `json:"is_public"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	nb, err := h.svc.Update(c.Request().Context(), id, uid, req.Title, req.Description, req.Content, req.IsPublic)
	if err != nil {
		if err.Error() == "not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		if err.Error() == "access denied" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, notebookToJSON(nb))
}

func (h *NotebookHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user"})
	}
	if err := h.svc.Delete(c.Request().Context(), id, uid); err != nil {
		if err.Error() == "not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		if err.Error() == "access denied" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func notebookToJSON(nb *ent.Notebook) map[string]interface{} {
	m := map[string]interface{}{
		"id":               nb.ID.String(),
		"owner_id":         nb.OwnerID.String(),
		"title":            nb.Title,
		"content":          nb.Content,
		"is_public":        nb.IsPublic,
		"created_at":       nb.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":       nb.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"execution_count":  nb.ExecutionCount,
	}
	if nb.Description != nil {
		m["description"] = *nb.Description
	} else {
		m["description"] = nil
	}
	if nb.LastExecutedAt != nil {
		m["last_executed_at"] = nb.LastExecutedAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		m["last_executed_at"] = nil
	}
	return m
}

func notebookToJSONWithOwner(nb *ent.Notebook) map[string]interface{} {
	m := notebookToJSON(nb)
	if nb.Edges.Owner != nil {
		m["owner"] = map[string]interface{}{
			"id":    nb.Edges.Owner.ID.String(),
			"email": nb.Edges.Owner.Email,
			"name":  nb.Edges.Owner.Name,
		}
	}
	return m
}
