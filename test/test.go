package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	chatRooms = make(map[string][]*websocket.Conn)
	mutex     = &sync.Mutex{}
)

func main() {
	router := gin.Default()
	router.GET("/ws/:room", handleWebSocket)
	router.Run(":8080")
}

func handleWebSocket(c *gin.Context) {
	room := c.Param("room")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	mutex.Lock()
	chatRooms[room] = append(chatRooms[room], conn)
	mutex.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			mutex.Lock()
			chatRooms[room] = removeConnection(chatRooms[room], conn)
			mutex.Unlock()
			return
		}

		mutex.Lock()
		connections := chatRooms[room]
		mutex.Unlock()

		for _, conn := range connections {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println(err)
				mutex.Lock()
				chatRooms[room] = removeConnection(chatRooms[room], conn)
				mutex.Unlock()
			}
		}
	}
}

func removeConnection(conns []*websocket.Conn, conn *websocket.Conn) []*websocket.Conn {
	for i, c := range conns {
		if c == conn {
			return append(conns[:i], conns[i+1:]...)
		}
	}
	return conns
}
