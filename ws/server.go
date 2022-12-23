package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request, h *hub, roomId string) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	// c := &connection{send: make(chan []byte, 256), ws: ws}
	c := NewConnection(ws)          //個人的連線
	s := NewSubscription(c, roomId) //個人欲訂閱的房間
	h.register <- *s                //註冊hub，有人加入了連線
	go s.writePump()                //server 寫進websocket的訊息
	go s.readPump(h)                //server 讀取前端發過來的訊息
}
