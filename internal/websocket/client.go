package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// ReadPump reads messages and forwards to the hub (broadcast to others in same notebook).
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	c.Hub.Register(c)
	// Broadcast user_joined
	msg := Message{
		Type: TypeUserJoined, NotebookID: c.NotebookID, UserID: c.UserID, UserName: c.UserName,
		Payload: map[string]string{"user_id": c.UserID, "user_name": c.UserName},
		Timestamp: time.Now().UnixMilli(),
	}
	b, _ := json.Marshal(msg)
	c.Hub.Broadcast(c.NotebookID, b, nil)

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		c.Hub.Broadcast(c.NotebookID, raw, c)
	}

	// Broadcast user_left
	leave := Message{
		Type: TypeUserLeft, NotebookID: c.NotebookID, UserID: c.UserID, UserName: c.UserName,
		Payload: map[string]string{"user_id": c.UserID, "user_name": c.UserName},
		Timestamp: time.Now().UnixMilli(),
	}
	lb, _ := json.Marshal(leave)
	c.Hub.Broadcast(c.NotebookID, lb, nil)
}

// WritePump writes messages from the hub to the client.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
