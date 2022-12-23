package main

import (
	"LIOU2021/Golang-realtime-chat-rooms/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	go ws.DemoHub.Run()

	router := gin.New()
	router.LoadHTMLFiles("index.html")

	router.GET("/room/:roomId", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	router.GET("/", func(c *gin.Context) {
		ws.DemoHub.PushRoom([]byte("hello word"), "2")
		c.String(200, "success")
	})

	router.GET("/push-all", func(c *gin.Context) {
		ws.DemoHub.PushAll([]byte("hello every body"))
		c.String(200, "to all")
	})

	router.GET("/ws/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		ws.ServeWs(c.Writer, c.Request, ws.DemoHub, roomId)
	})

	router.Run("0.0.0.0:8080")
}
