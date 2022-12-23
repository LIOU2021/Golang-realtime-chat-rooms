package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type subscription struct {
	conn *connection
	room string
}

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump(h *hub) {
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
