package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// connection Id
	id int

	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *connection) GetId() int {
	return c.id
}

func NewConnection(ws *websocket.Conn) *connection {
	return &connection{
		send: make(chan []byte, 256),
		ws:   ws,
		id:   Ai.ID(),
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}
