package main

import (
	"fmt"
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

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump() {
	c := s.conn

	defer func() {
		h.unregister <- s //通知hub有人離開連線了，並且將該user的連線資訊傳送過去
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage() //獲取訊息。這裡會堵塞。後面的code要等message接收到。
		if err != nil {                   //如果user斷開連線就會觸發break離開此迴圈，間接觸發h.unregister
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		m := message{msg, s.room}
		h.broadcast <- m //訊息推播
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send: //從readPump > broadcast 會塞資料到 c.send 這個channel裡面
			if !ok { //收到c.send這個channel已經關閉了，所以要關閉連線
				c.write(websocket.CloseMessage, []byte{}) //通知ws 關閉連線
				return
			}

			if err := c.write(websocket.TextMessage, message); err != nil { //前端發來訊息，寫入ws > 觸發readPump > 觸發broadcast(推播)
				return
			}
		case <-ticker.C: //心跳機制。每X秒去ping前端在線用戶
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request, roomId string) {
	fmt.Println("room : ", roomId)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws} //個人的連線
	s := subscription{c, roomId}                           //個人欲訂閱的房間
	h.register <- s                                        //註冊hub，有人加入了連線
	go s.writePump()                                       //server 寫進websocket的訊息
	go s.readPump()                                        //server 讀取前端發過來的訊息
}
