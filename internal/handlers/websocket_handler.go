package handlers

import (
	"net/http"

	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gladiator/internal/services"
	ws "gladiator/internal/websocket"
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler handles GET /ws/notebooks/:id?token=...
func WebSocketHandler(authSvc *services.AuthService, nbSvc *services.NotebookService, hub *ws.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		notebookID := c.Param("id")
		token := c.QueryParam("token")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token required"})
		}
		// Validate JWT and get user (simplified: use authService to validate and get user_id from token)
		userID, userName, err := authSvc.UserFromToken(c.Request().Context(), token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
		uid, _ := uuid.Parse(userID)
		_, err = nbSvc.CheckNotebookAccess(c.Request().Context(), notebookID, uid)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		client := &ws.Client{
			ID:         uuid.New().String(),
			UserID:     userID,
			UserName:   userName,
			NotebookID: notebookID,
			Send:       make(chan []byte, 256),
			Hub:        hub,
			Conn:       conn,
		}
		go client.WritePump()
		client.ReadPump()
		return nil
	}
}

