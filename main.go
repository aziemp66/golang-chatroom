package main

import (
	"github.com/gin-gonic/gin"

	internal "github.com/aziemp66/golang-chatroom/internal"
)

func main() {
	go internal.H.Run()

	router := gin.New()
	router.LoadHTMLFiles("index.html")

	router.GET("/room/:roomId", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	router.GET("/ws/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		internal.ServeWs(c.Writer, c.Request, roomId)
	})

	router.Run(":3000")
}
