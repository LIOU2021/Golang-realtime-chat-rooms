package ws

import (
	"log"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// hub name
	name string

	// Registered connections.
	rooms map[string]map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}

// get hub name
func (h *hub) GetName() string {
	return h.name
}

func NewHub(name string) *hub {
	return &hub{
		name:       name,
		broadcast:  make(chan message),
		register:   make(chan subscription),
		unregister: make(chan subscription),
		rooms:      make(map[string]map[*connection]bool), //存放所有房間
	}
}

// push message to specific room id
func (h *hub) PushRoom(msg []byte, roomId string) {
	m := message{msg, roomId}
	h.broadcast <- m //訊息推播
}

// push message to all user
func (h *hub) PushAll(msg []byte) {
	for i := range h.rooms {
		m := message{msg, i}
		h.broadcast <- m
	}
}

// listen register, unregister, broadcast
func (h *hub) Run() {
	for {
		select {
		case s := <-h.register: //加入聊天室
			log.Printf("hub : %s, event : register, room : %s, connectionId: %d", h.GetName(), s.room, s.conn.GetId())
			connections := h.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				h.rooms[s.room] = connections
			}
			h.rooms[s.room][s.conn] = true
		case s := <-h.unregister: //離開聊天室
			log.Printf("hub : %s, event : unregister, room : %s, connectionId: %d", h.GetName(), s.room, s.conn.GetId())
			connections := h.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(h.rooms, s.room)
					}
					Ai.PutID(s.conn.GetId())
				}
			}
		case m := <-h.broadcast: //對聊天室發送訊息
			log.Printf("hub : %s, event : broadcast, room : %s", h.GetName(), m.room)
			connections := h.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.rooms, m.room)
					}
				}
			}
		}
	}
}
