package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a connected user in a notebook.
type Client struct {
	ID         string
	UserID     string
	UserName   string
	NotebookID string
	Send       chan []byte
	Hub        *Hub
	Conn       *websocket.Conn
}

// Hub holds clients per notebook and broadcasts messages.
type Hub struct {
	mu        sync.RWMutex
	notebooks map[string]map[*Client]struct{} // notebookID -> clients
	broadcast chan *broadcastMsg
	register   chan *Client
	unregister chan *Client
}

type broadcastMsg struct {
	notebookID string
	message    []byte
	exclude    *Client
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		notebooks: make(map[string]map[*Client]struct{}),
		broadcast:  make(chan *broadcastMsg, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run runs the hub loop.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.notebooks[c.NotebookID] == nil {
				h.notebooks[c.NotebookID] = make(map[*Client]struct{})
			}
			h.notebooks[c.NotebookID][c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if m, ok := h.notebooks[c.NotebookID]; ok {
				delete(m, c)
				if len(m) == 0 {
					delete(h.notebooks, c.NotebookID)
				}
			}
			h.mu.Unlock()
			close(c.Send)

		case msg := <-h.broadcast:
			h.mu.RLock()
			m := h.notebooks[msg.notebookID]
			for c := range m {
				if c == msg.exclude {
					continue
				}
				select {
				case c.Send <- msg.message:
				default:
					// skip slow client
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all clients in the notebook (optionally excluding one).
func (h *Hub) Broadcast(notebookID string, message []byte, exclude *Client) {
	h.broadcast <- &broadcastMsg{notebookID: notebookID, message: message, exclude: exclude}
}

// Register adds a client.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// GetActiveUsers returns user IDs in the notebook (for presence).
func (h *Hub) GetActiveUsers(notebookID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	m := h.notebooks[notebookID]
	out := make([]string, 0, len(m))
	for c := range m {
		out = append(out, c.UserID)
	}
	return out
}

const writeWait = 10 * time.Second
const pongWait = 60 * time.Second
const pingPeriod = 30 * time.Second
const maxMessageSize = 512 * 1024
