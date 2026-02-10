package websocket

// Message types for real-time collaboration.
const (
	TypeUserJoined   = "user_joined"
	TypeUserLeft     = "user_left"
	TypeCellUpdate   = "cell_update"
	TypeCellExecute  = "cell_execute"
	TypeCellOutput   = "cell_output"
	TypeCellLock     = "cell_lock"
	TypeCellUnlock   = "cell_unlock"
	TypeCursorMove   = "cursor_move"
)

// Message is the wire format for WebSocket messages.
type Message struct {
	Type       string      `json:"type"`
	NotebookID string      `json:"notebook_id"`
	UserID     string      `json:"user_id"`
	UserName   string      `json:"user_name"`
	Payload    interface{} `json:"payload"`
	Timestamp  int64       `json:"timestamp"`
}
