package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gladiator/internal/services"
)

type ExecutionHandler struct {
	execSvc *services.ExecutorService
}

func NewExecutionHandler(execSvc *services.ExecutorService) *ExecutionHandler {
	return &ExecutionHandler{execSvc: execSvc}
}

func (h *ExecutionHandler) Execute(c echo.Context) error {
	notebookID := c.Param("id")
	if notebookID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "notebook id required"})
	}
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	var req struct {
		CellID string `json:"cell_id"`
		Code   string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.CellID == "" || req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cell_id and code required"})
	}
	result, err := h.execSvc.ExecuteCell(c.Request().Context(), notebookID, userID, req.CellID, req.Code)
	if err != nil {
		if err.Error() == "notebook not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		if err.Error() == "access denied" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (h *ExecutionHandler) GetSession(c echo.Context) error {
	notebookID := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	info, err := h.execSvc.GetSessionInfo(c.Request().Context(), notebookID, userID)
	if err != nil {
		if err.Error() == "not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		if err.Error() == "access denied" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"session": info})
}

func (h *ExecutionHandler) ClearSession(c echo.Context) error {
	notebookID := c.Param("id")
	userID, _ := c.Get("user_id").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.execSvc.ClearSession(c.Request().Context(), notebookID, userID); err != nil {
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
