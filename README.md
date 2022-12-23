# Golang-realtime-chat-rooms

### Start the server
```bash
go run main.go
```
 
http://localhost:8080/room/1

### how to use ?
- step1 : New your hub. see  .\ws\hub_list.go
```go
var DemoHub = NewHub("demo_hub")
```
- step2 : listen your hub. see .\main.go
```go
go ws.DemoHub.Run()
```
- step3 : create your ws router
```go
router.GET("/ws/:roomId", func(c *gin.Context) {
    roomId := c.Param("roomId")
    ws.ServeWs(c.Writer, c.Request, ws.DemoHub, roomId)
})
```

### usage example
```go
// push message to room id 2
ws.DemoHub.PushRoom([]byte("hello word"), "2")
//push message to all connect user
ws.DemoHub.PushAll([]byte("hello every body"))
```